# ADR-RTGF-002: Sanctions Dataset Hashing & Evidence References

**Status:** Accepted  
**Date:** 2025-11-01  
**Decision Makers:** Kevin Brown, Corridor Governance Group  
**Owner:** RTGF Working Group  
**Related Artifacts:** `docs/architecture/policy-source-matrix.md`, `docs/architecture/rtgf-pipeline.md`

---

## 1. Purpose & Scope
Sanctions lists change frequently and are sourced from multiple authorities (EU Sanctions Map, UN, OFAC). RTGF must prove which dataset versions were used during compilation without embedding large payloads in RMT/IMT tokens.

## 2. Architecture Overview
- During compilation, fetch the latest sanctions datasets (public APIs or signed exports) and compute **SHA-256 digests**.  
- Record digests and dataset metadata in the **policy snapshot** (`evidence_sources`).  
- Embed the same digests inside generated **RMTs/IMTs** so verifiers can confirm the compiler used the correct dataset version.  
- Store dataset provenance (source URI, retrieval timestamp, digest) in the **transparency log** (future ADR-RTGF-014).

## 3. Determinism & Provenance
- Canonical hash format: `sha256:<hex>` using sorted dataset payload.  
- Snapshot includes retrieval timestamp and dataset version; compiler ties digests to token IDs.  
- Transparency log entries anchor the digest with Merkle inclusion proofs.

## 4. Security & Trust
- Dataset downloads over authenticated HTTPS; prefer signed exports.  
- Hash mismatches or unsigned payloads trigger compilation failure.  
- Sanctions providers' public keys pinned where available.

## 5. Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| RTGF_SANCTIONS_HASH_MISMATCH | Computed hash differs from snapshot | halt compilation |
| RTGF_SANCTIONS_FETCH_FAILED | Dataset retrieval failure | retry with backoff / fail |
| RTGF_SANCTIONS_UNSIGNED | Missing signature where required | reject dataset |

## 6. Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|-------|
| Dataset fetch latency | ≤ 5 min per provider | includes retries |
| Hash mismatch incidents | 0 tolerated | triggers incident |

## 7. Interfaces & Integration
| Dependency | Direction | Purpose |
|------------|-----------|---------|
| Sanctions providers | inbound | Retrieve datasets |
| Policy Source Matrix | outbound | Store digests/provenance |
| RTGF Compiler | outbound | Embed digests in tokens |
| Transparency log | outbound | Append dataset proof |

## 8. Metrics & Observability
- Prometheus: `rtgf_sanctions_fetch_seconds_bucket{provider}`, `rtgf_sanctions_hash_mismatch_total`.  
- Structured logs: provider, dataset version, hash, retrieval timestamp, signature metadata.

## 9. Acceptance Tests
| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| RTGF-CT-10 | Valid dataset fetch & hash | hash recorded in snapshot & token |
| RTGF-CT-11 | Hash mismatch | compilation fails with `RTGF_SANCTIONS_HASH_MISMATCH` |
| RTGF-CT-12 | Provider timeout | retries then fail-fast |

## 10. Acceptance Criteria
1️⃣ Sanctions datasets hashed deterministically and linked to snapshots/tokens.  
2️⃣ Transparency log entries include dataset provenance.  
3️⃣ Pipeline fails closed on hash mismatch or signature failure.

## Consequences
- ✅ Tokens remain compact and deterministic while preserving linkability to external datasets.  
- ✅ Auditors can independently cross-check dataset hashes.  
- ⚠️ Requires scheduling and caching infrastructure for dataset acquisition during compilation.

## Links
- `docs/architecture/policy-source-matrix.md`  
- `docs/architecture/rtgf-pipeline.md`
