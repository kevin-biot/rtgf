export type Decision = "PERMIT" | "PERMIT_WITH_CONTROLS" | "DENY";
export interface EvalOpts {
  token: any;
  context: any;
  resolvers?: Record<string, Function>;
  now?: string;
}
export interface EvalResult {
  decision: Decision;
  reasons: string[];
  controls: string[];
  trace: any;
  digest: string;
}

// NOTE: Placeholder implementation. TODO: operator evaluation, RFC 8785 canonicalization, hashing, resolvers.
export async function evaluate(opts: EvalOpts): Promise<EvalResult> {
  const { token, context, resolvers = {}, now } = opts;
  const plan = token.eval_plan || token.evalPlan || {};
  const predicateSet = token.predicate_set || token.predicateSet || { predicates: [] };
  const trace: any = {
    token_id: token.imt_id || token.rmt_id,
    plan_hash: plan.hash,
    steps: [],
    ts: now || new Date().toISOString()
  };
  const reasons: string[] = [];
  const controlsPermit: string[] = token.controls_on_permit || [];
  const controlsPwC: string[] = token.controls_on_permit_with_controls || [];

  // Stage 0: input validation (stub)
  trace.steps.push({ stage: "P0", ok: true });

  for (const step of (plan.sequence || []).filter((s: any) => s.stage !== "P0")) {
    const predicateId = step.predicate;
    const predicate = (predicateSet.predicates || []).find((p: any) => p.id === predicateId) || {};
    // TODO: evaluate predicate logic deterministically
    const ok = true; // placeholder result
    trace.steps.push({ stage: step.stage, predicate: predicateId, result: ok });

    if (!ok) {
      reasons.push(predicate.on_fail?.reason || "PREDICATE_FAIL");
      if (predicate.on_fail?.decision === "PERMIT_WITH_CONTROLS") {
        return {
          decision: "PERMIT_WITH_CONTROLS",
          reasons,
          controls: controlsPwC,
          trace,
          digest: "sha256:TODO"
        };
      }
      return {
        decision: "DENY",
        reasons,
        controls: [],
        trace,
        digest: "sha256:TODO"
      };
    }
  }

  return {
    decision: "PERMIT",
    reasons,
    controls: controlsPermit,
    trace,
    digest: "sha256:TODO"
  };
}
