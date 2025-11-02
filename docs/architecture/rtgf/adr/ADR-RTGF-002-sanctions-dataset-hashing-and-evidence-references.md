# ADR-RTGF-002: Sanctions Dataset Hashing & Evidence References

**Status:** Accepted  
**Date:** 2025-11-01  
**Decision Makers:** Kevin Brown, Corridor Governance Group  
**Owner:** RTGF Working Group  
**Related Artifacts:** `docs/architecture/policy-source-matrix.md`, `docs/architecture/rtgf-pipeline.md`

---

## Context
Sanctions lists change frequently and are sourced from multiple authorities (EU Sanctions Map, UN, OFAC). RTGF must prove which dataset versions were used during compilation without embedding large payloads in RMT/IMT tokens.

## Decision
- During compilation, fetch the latest sanctions datasets (public APIs or signed exports) and compute **SHA-256 digests**.  
- Record digests and dataset metadata in the **policy snapshot** (`evidence_sources`).  
- Embed the same digests inside generated **RMTs/IMTs** so verifiers can confirm the compiler used the correct dataset version.  
- Store dataset provenance (source URI, retrieval timestamp, digest) in the **transparency log** (future ADR-RTGF-014).

## Consequences
- Tokens remain compact and deterministic while preserving linkability to external datasets.  
- Auditors can independently cross-check dataset hashes.  
- Requires scheduling and caching infrastructure for dataset acquisition during compilation.

## Links
- `docs/architecture/policy-source-matrix.md`  
- `docs/architecture/rtgf-pipeline.md`

