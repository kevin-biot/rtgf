# ADR-002: Sanctions Dataset Hashing and Evidence References

## Status
Accepted

## Context
Sanctions lists change frequently and are sourced from multiple authorities (EU Sanctions Map, UN, OFAC). RTGF must prove which dataset versions were used during compilation without embedding large payloads in RMT/IMT tokens.

## Decision
- During compilation, fetch the latest sanctions datasets (public APIs or signed exports) and compute SHA-256 digests.
- Record digests and dataset metadata in the policy snapshot (`evidence_sources`).
- Embed the same digests inside generated RMTs/IMTs so verifiers can confirm the compiler used the correct dataset version.
- Store dataset provenance (URI, retrieval timestamp, hash) in the transparency log.

## Consequences
- Tokens stay small and deterministic while preserving linkability to external data.
- Auditors can cross-check dataset hashes independently.
- Requires infrastructure to schedule dataset downloads and cache them during the compilation run.

## Links
- `docs/architecture/policy-source-matrix.md`
- `docs/architecture/rtgf-pipeline.md`
