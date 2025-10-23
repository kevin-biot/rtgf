# PPE-Compiler Determinism Test

1. Given policy snapshot `A`
2. Run `rtgf-ppe compile` twice
3. Expect identical bytes for `predicate_set.json` and `eval_plan.json`
4. Validate outputs against shared schemas
5. Recompute hash matches embedded `hash` field
