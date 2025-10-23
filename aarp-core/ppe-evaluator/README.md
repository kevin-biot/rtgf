# PPE-Evaluator (Run-time)

Evaluates token-embedded `predicate_set` + `eval_plan` against a concrete context with optional resolvers.
Outputs a decision, reasons, controls, and a canonical EvalTrace digest.

## API (conceptual)
```
evaluate({ token, context, resolvers, now? })
  → { decision, reasons[], controls[], trace, digest }
```

## Determinism
- Execute exactly in plan sequence.
- Canonicalize all step records (RFC 8785) before hashing.
- Time inputs are RFC 3339 UTC; PDP applies ±120s skew policy externally.
