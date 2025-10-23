# PPE-Compiler (Build-time)

Converts signed policy snapshots → normalized `predicate_set` + `eval_plan` and embeds them into RMT/IMT during RTGF compilation.

## CLI
```
rtgf-ppe compile \
  --snapshot ./policy.jsonld \
  --out ./predicate_set.json \
  --plan-out ./eval_plan.json
```

## Guarantees
- Deterministic output (same inputs → same bytes).
- Validates against shared schemas; signs outputs; computes plan hash.
