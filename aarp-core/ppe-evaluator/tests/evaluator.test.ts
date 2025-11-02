import { describe, expect, it } from "vitest";
import { evaluate } from "../src/engine";

function buildToken(overrides: Partial<any> = {}) {
  return {
    imt_id: "IMT-EU.SG-PAYMENTS_AML-2025-10-22",
    predicate_set: {
      predicate_set_id: "ps.EU.PSD3.v1",
      version: "1.0.0",
      order: ["pred.allow", "pred.controls", "pred.resolver", "pred.inputs"],
      predicates: [
        {
          id: "pred.allow",
          inputs: [],
          domain: "kyc",
          logic: { op: "PASS" }
        },
        {
          id: "pred.controls",
          inputs: [],
          domain: "controls",
          logic: { op: "PASS" },
          on_fail: { decision: "PERMIT_WITH_CONTROLS", reason: "MANUAL_REVIEW_REQUIRED" },
          mock_result: true
        },
        {
          id: "pred.resolver",
          inputs: [{ name: "customer.id", type: "string", required: true }],
          domain: "sanctions",
          logic: { op: "SANCTIONS" },
          on_fail: { decision: "DENY", reason: "SANCTIONS_HIT" },
          resolver: "sanctions_api"
        },
        {
          id: "pred.inputs",
          inputs: [{ name: "application.field", type: "string", required: true }],
          domain: "inputs",
          logic: { op: "MISSING_INPUT" },
          on_fail: { decision: "DENY", reason: "INPUT_MISSING" }
        }
      ]
    },
    eval_plan: {
      eval_plan_id: "plan.EU.PSD3.v1",
      sequence: [
        { stage: "P0", op: "VALIDATE_INPUTS" },
        { stage: "P1", predicate: "pred.allow", ops: ["D"] },
        { stage: "P2", predicate: "pred.controls", ops: ["D"] },
        { stage: "P3", predicate: "pred.resolver", ops: ["D"] },
        { stage: "P4", predicate: "pred.inputs", ops: ["D"] }
      ],
      hash: "sha256:test"
    },
    controls_on_permit: ["LOG:L2"],
    controls_on_permit_with_controls: ["HUMAN_REVIEW"],
    ...overrides
  };
}

describe("evaluate", () => {
  it("returns PERMIT when all predicates pass", async () => {
    const token = buildToken({
      predicate_set: {
        predicate_set_id: "ps.EU.PSD3.v1",
        version: "1.0.0",
        order: ["pred.allow"],
        predicates: [
          {
            id: "pred.allow",
            inputs: [],
            domain: "kyc",
            logic: { op: "PASS" }
          }
        ]
      },
      eval_plan: {
        eval_plan_id: "plan.EU.PSD3.v1",
        sequence: [
          { stage: "P0", op: "VALIDATE_INPUTS" },
          { stage: "P1", predicate: "pred.allow", ops: ["D"] }
        ],
        hash: "sha256:test"
      }
    });
    const result = await evaluate({ token, context: {} });
    expect(result.decision).toBe("PERMIT");
    expect(result.controls).toEqual(["LOG:L2"]);
    expect(result.reasons).toHaveLength(0);
    expect(result.trace.steps).toHaveLength(2);
  });

  it("denies when predicate fails with explicit reason", async () => {
    const token = buildToken({
      predicate_set: {
        predicate_set_id: "ps.EU.PSD3.v1",
        version: "1.0.0",
        order: ["pred.controls"],
        predicates: [
          {
            id: "pred.controls",
            inputs: [],
            domain: "controls",
            logic: { op: "FAIL" },
            on_fail: { decision: "DENY", reason: "SANCTIONS_HIT" },
            mock_result: false
          }
        ]
      },
      eval_plan: {
        eval_plan_id: "plan.EU.PSD3.v1",
        sequence: [
          { stage: "P0", op: "VALIDATE_INPUTS" },
          { stage: "P1", predicate: "pred.controls", ops: ["D"] }
        ],
        hash: "sha256:test"
      }
    });
    const result = await evaluate({ token, context: {} });
    expect(result.decision).toBe("DENY");
    expect(result.reasons).toEqual(["SANCTIONS_HIT"]);
    expect(result.controls).toHaveLength(0);
  });

  it("permits with controls when fallback decision is PERMIT_WITH_CONTROLS", async () => {
    const token = buildToken({
      predicate_set: {
        predicate_set_id: "ps.EU.PSD3.v1",
        version: "1.0.0",
        order: ["pred.controls"],
        predicates: [
          {
            id: "pred.controls",
            inputs: [],
            domain: "controls",
            logic: { op: "FAIL" },
            mock_result: false,
            on_fail: { decision: "PERMIT_WITH_CONTROLS", reason: "MANUAL_REVIEW_REQUIRED" }
          }
        ]
      },
      eval_plan: {
        eval_plan_id: "plan.EU.PSD3.v1",
        sequence: [
          { stage: "P0", op: "VALIDATE_INPUTS" },
          { stage: "P1", predicate: "pred.controls", ops: ["D"] }
        ],
        hash: "sha256:test"
      }
    });
    const result = await evaluate({ token, context: {} });
    expect(result.decision).toBe("PERMIT_WITH_CONTROLS");
    expect(result.controls).toEqual(["HUMAN_REVIEW"]);
    expect(result.reasons).toEqual(["MANUAL_REVIEW_REQUIRED"]);
  });

  it("denies when resolver throws with resolver error reason", async () => {
    const token = buildToken({
      predicate_set: {
        predicate_set_id: "ps.EU.PSD3.v1",
        version: "1.0.0",
        order: ["pred.resolver"],
        predicates: [
          {
            id: "pred.resolver",
            inputs: [{ name: "customer.id", type: "string", required: true }],
            domain: "sanctions",
            logic: { op: "SANCTIONS" },
            on_fail: { decision: "DENY", reason: "SANCTIONS_HIT" },
            resolver: "sanctions_api"
          }
        ]
      },
      eval_plan: {
        eval_plan_id: "plan.EU.PSD3.v1",
        sequence: [
          { stage: "P0", op: "VALIDATE_INPUTS" },
          { stage: "P1", predicate: "pred.resolver", ops: ["D"] }
        ],
        hash: "sha256:test"
      }
    });
    const result = await evaluate({
      token,
      context: { customer: { id: "123" } },
      resolvers: {
        sanctions_api: () => {
          throw new Error("timeout");
        }
      }
    });
    expect(result.decision).toBe("DENY");
    expect(result.reasons).toEqual(["RESOLVER_ERROR:sanctions_api"]);
  });

  it("denies when required input missing", async () => {
    const token = buildToken({
      predicate_set: {
        predicate_set_id: "ps.EU.PSD3.v1",
        version: "1.0.0",
        order: ["pred.inputs"],
        predicates: [
          {
            id: "pred.inputs",
            inputs: [{ name: "application.field", type: "string", required: true }],
            domain: "inputs",
            logic: { op: "MISSING_INPUT" },
            on_fail: { decision: "DENY", reason: "INPUT_MISSING" }
          }
        ]
      },
      eval_plan: {
        eval_plan_id: "plan.EU.PSD3.v1",
        sequence: [
          { stage: "P0", op: "VALIDATE_INPUTS" },
          { stage: "P1", predicate: "pred.inputs", ops: ["D"] }
        ],
        hash: "sha256:test"
      }
    });
    const result = await evaluate({ token, context: {} });
    expect(result.decision).toBe("DENY");
    expect(result.reasons).toEqual(["INPUT_MISSING:application.field"]);
  });
});
