# PPE Actions (Build-time & Run-time)

## rtgf-compiler/ppe-compiler
- [ ] Implement snapshot → predicate transformation (normalize inputs, map legal references to predicate IDs).
- [ ] Generate deterministic `order` plus `eval_plan.sequence`; compute plan hash (RFC 8785 → SHA-256).
- [ ] Validate outputs against shared schemas; add detached Ed25519 JWS over canonical bytes.
- [ ] Embed `{predicate_set, eval_plan}` into RMT/IMT payloads during token build.
- [ ] Add unit tests for repeatable bytes, schema validation, and hash stability.

## aarp-core/ppe-evaluator
- [ ] Implement operator evaluation with strict typing & bounded arithmetic.
- [ ] Implement DERIVE and LOOKUP stages with pluggable resolvers (e.g., sanctions, Mandala proof verifier).
- [ ] Build canonical EvalTrace and digest; integrate with DOP evidence pipeline.
- [ ] Define error taxonomy: `INPUT_MISSING:x`, `RESOLVER_ERROR:name`, `PREDICATE_FAIL`, `SANCTIONS_HIT`.
- [ ] Conformance tests: allow path, sanctions hit, resolver timeout, missing input.

## Integration
- [ ] aARP PDP invokes PPE-Evaluator before STA issuance; PEP may re-check at execution time.
- [ ] Registries/RTGF publish tokens with embedded PPE artefacts; routers cache with TTL ≤ 24h.
- [ ] Optional Mandala hook: predicates may require external proof; evaluator verifies and logs proof hash.
