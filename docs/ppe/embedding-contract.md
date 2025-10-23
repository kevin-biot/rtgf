# Embedding PPE Artefacts in IMT/RMT

Tokens MUST embed:
- `predicate_set`: `{ predicate_set_id, version, order:[ids], predicates:[…] }`
- `eval_plan`: `{ eval_plan_id, sequence:[…], hash }`

At run-time, the aARP PDP hands both objects and the execution context to PPE-Evaluator. The evaluator returns the decision, reasons, controls, and EvalTrace digest. If the decision is `PERMIT` or `PERMIT_WITH_CONTROLS`, an STA MAY be issued; otherwise the request is denied. The EvalTrace digest MUST be included in the DOP Evidence Bundle and transparency log.
