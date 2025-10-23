# PPE-Evaluator Conformance Cases

- **Test 01**: Allow path → expect `PERMIT`, empty reasons, stable digest.
- **Test 02**: Sanctions hit → expect `DENY`, reasons:`["SANCTIONS_HIT"]`.
- **Test 03**: Resolver timeout → expect `DENY`, reasons:`["RESOLVER_ERROR:SANCTIONS_API"]`.
- **Test 04**: Missing input → expect `DENY`, reasons:`["INPUT_MISSING:<field>"]`.

Each case MUST produce deterministic EvalTrace digest across runs.
