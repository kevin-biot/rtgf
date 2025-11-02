import { mkdtempSync, readFileSync, writeFileSync } from "fs";
import { tmpdir } from "os";
import { join } from "path";
import { describe, expect, it } from "vitest";
import { compile, run } from "../cmd/rtgf-ppe";

function expectPredicateSchema(predicate: any) {
  expect(typeof predicate.id).toBe("string");
  expect(typeof predicate.domain).toBe("string");
  expect(Array.isArray(predicate.inputs)).toBe(true);
  predicate.inputs.forEach((input: any) => {
    expect(typeof input.name).toBe("string");
    expect(typeof input.type).toBe("string");
    expect(typeof input.required).toBe("boolean");
  });
  expect(typeof predicate.logic).toBe("object");
  if (predicate.on_fail) {
    expect(["DENY", "PERMIT_WITH_CONTROLS"]).toContain(predicate.on_fail.decision);
    if (predicate.on_fail.reason) {
      expect(typeof predicate.on_fail.reason).toBe("string");
    }
  }
}

function expectEvalPlanSchema(plan: any) {
  expect(typeof plan.eval_plan_id).toBe("string");
  expect(Array.isArray(plan.sequence)).toBe(true);
  plan.sequence.forEach((step: any) => {
    expect(typeof step.stage).toBe("string");
    if (step.predicate) {
      expect(typeof step.predicate).toBe("string");
    }
    if (step.ops) {
      expect(Array.isArray(step.ops)).toBe(true);
    }
  });
  expect(typeof plan.hash).toBe("string");
}

const snapshotFixture = {
  jurisdiction: "EU",
  domain: "PSD3",
  required_predicates: ["pred.checkKYC", "pred.checkSanctions"],
  predicates: [
    {
      id: "pred.checkKYC",
      domain: "kyc",
      inputs: [
        { name: "customer.id", type: "string", required: true },
        { name: "customer.country", type: "string", required: true }
      ],
      logic: { op: "AND", operands: [] },
      on_fail: { decision: "DENY", reason: "KYC_MISSING" }
    },
    {
      id: "pred.checkSanctions",
      domain: "sanctions",
      inputs: [{ name: "customer.id", type: "string", required: true }],
      logic: { op: "SANCTIONS_SCREEN", operands: [] },
      on_fail: { decision: "DENY", reason: "SANCTIONS_HIT" }
    }
  ]
};

function writeSnapshot(): { snapshotPath: string; outDir: string } {
  const outDir = mkdtempSync(join(tmpdir(), "rtgf-ppe-"));
  const snapshotPath = join(outDir, "snapshot.json");
  writeFileSync(snapshotPath, JSON.stringify(snapshotFixture, null, 2), "utf-8");
  return { snapshotPath, outDir };
}

describe("rtgf-ppe CLI", () => {
  it("produces deterministic predicate set and evaluation plan", () => {
    const { snapshotPath, outDir } = writeSnapshot();
    const predPath = join(outDir, "predicate_set.json");
    const planPath = join(outDir, "eval_plan.json");
    const first = compile({ snapshotPath, outPred: predPath, outPlan: planPath });
    const predJSON = JSON.parse(readFileSync(predPath, "utf-8"));
    const planJSON = JSON.parse(readFileSync(planPath, "utf-8"));
    predJSON.predicates.forEach((predicate: any) => expectPredicateSchema(predicate));
    expectEvalPlanSchema(planJSON);
    expect(first.predicateSet).toEqual(predJSON);
    expect(first.plan).toEqual(planJSON);
    expect(predJSON.order).toEqual(snapshotFixture.required_predicates);
    expect(planJSON.sequence.length).toBe(1 + snapshotFixture.required_predicates.length);
    expect(planJSON.sequence[1].predicate).toBe("pred.checkKYC");

    // Re-run compile into a separate directory and ensure byte equality.
    const secondPaths = writeSnapshot();
    const secondPred = join(secondPaths.outDir, "predicate_set.json");
    const secondPlan = join(secondPaths.outDir, "eval_plan.json");
    compile({ snapshotPath: secondPaths.snapshotPath, outPred: secondPred, outPlan: secondPlan });
    const predAgain = readFileSync(secondPred, "utf-8");
    const planAgain = readFileSync(secondPlan, "utf-8");
    expect(readFileSync(predPath, "utf-8")).toBe(predAgain);
    expect(readFileSync(planPath, "utf-8")).toBe(planAgain);
  });

  it("throws on missing arguments", () => {
    expect(() => run([])).toThrow(/Usage:/);
  });
});
