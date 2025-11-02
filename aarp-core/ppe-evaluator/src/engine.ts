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
    const stepTrace: any = { stage: step.stage, predicate: predicateId };
    let ok = true;
    let failureReason: string | undefined;

    if (!predicate || !predicate.id) {
      ok = false;
      failureReason = `PREDICATE_MISSING:${predicateId}`;
    }

    if (ok && Array.isArray(predicate.inputs)) {
      for (const input of predicate.inputs) {
        if (input?.required) {
          if (!hasValue(context, input.name)) {
            ok = false;
            failureReason = `INPUT_MISSING:${input.name}`;
            break;
          }
        }
      }
    }

    if (ok && Object.prototype.hasOwnProperty.call(predicate, "mock_result")) {
      ok = Boolean(predicate.mock_result);
    }

    if (ok && predicate.resolver) {
      const resolverFn = resolvers[predicate.resolver];
      if (!resolverFn) {
        throw new Error(`resolver_missing:${predicate.resolver}`);
      }
      try {
        const result = await resolverFn({ predicate, context, token, step });
        if (typeof result === "boolean") {
          ok = result;
        } else if (result && typeof result === "object") {
          if (Object.prototype.hasOwnProperty.call(result, "ok")) {
            ok = Boolean((result as any).ok);
          }
          if (result.reason) {
            failureReason = String(result.reason);
          }
        }
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        const resolverReason = `RESOLVER_ERROR:${predicate.resolver}`;
        reasons.push(resolverReason);
        stepTrace.result = false;
        stepTrace.error = message;
        trace.steps.push(stepTrace);
        return {
          decision: "DENY",
          reasons,
          controls: [],
          trace,
          digest: "sha256:TODO"
        };
      }
    }

    if (!ok) {
      const reason = failureReason || predicate.on_fail?.reason || `PREDICATE_FAIL:${predicateId}`;
      if (reason) {
        reasons.push(reason);
        stepTrace.reason = reason;
      }
      stepTrace.result = false;
      trace.steps.push(stepTrace);
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

    stepTrace.result = true;
    trace.steps.push(stepTrace);
  }

  return {
    decision: "PERMIT",
    reasons,
    controls: controlsPermit,
    trace,
    digest: "sha256:TODO"
  };
}

function hasValue(context: any, path: string): boolean {
  if (!path) {
    return true;
  }
  const segments = path.split(".");
  let cursor = context;
  for (const segment of segments) {
    if (cursor == null || !Object.prototype.hasOwnProperty.call(cursor, segment)) {
      return false;
    }
    cursor = cursor[segment];
  }
  return cursor !== undefined && cursor !== null && cursor !== "";
}
