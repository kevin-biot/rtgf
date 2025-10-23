# PPE-Compiler Contracts

- Inputs: verified policy snapshot (JSON-LD, RFC 8785 canonicalizable)
- Outputs: `predicate_set.json`, `eval_plan.json` (must match shared schemas)
- Sorting: alphabetical keys and stable array order
- Hashing: RFC 8785 canonical form → SHA-256 → `"sha256:<hex>"`
- Signing: detached Ed25519 JWS over canonical bytes (future implementation step)
