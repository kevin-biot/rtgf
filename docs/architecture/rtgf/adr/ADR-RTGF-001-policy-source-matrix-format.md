# ADR-RTGF-001: Policy Source Matrix Format

**Status:** Accepted  
**Date:** 2025-11-01  
**Decision Makers:** Kevin Brown, Corridor Governance Group  
**Owner:** RTGF Working Group  
**Related Artifacts:** `docs/architecture/policy-source-matrix.md`, *draft-lane2-rtgf-00 §3.1*

---

## 1. Purpose & Scope
RTGF requires deterministic, versioned policy inputs per jurisdiction and domain. Regulators publish heterogeneous legal texts, sanctions feeds, and guidance; the compiler needs a uniform schema to ingest them.

## 2. Architecture Overview
Create a **Policy Source Matrix** comprised of signed JSON-LD snapshots (`policy:Snapshot`) keyed by jurisdiction and domain. Each snapshot **MUST** include:
- Jurisdiction code (ISO 3166-1 alpha-2 or regional)
- Domain identifier aligned with the RTGF domain registry
- Effective/expiry timestamps (RFC 3339 UTC)
- Controls, duties, prohibitions, assurance levels
- Evidence source hashes for external datasets (e.g., sanctions)
- Normative references (legal instruments, guidance)
- **Detached Ed25519** signature by the issuing authority (DID/JWKS)

Snapshots are listed in `docs/architecture/policy-source-matrix/index.jsonld` with metadata for automated retrieval.

## 3. Determinism & Provenance
- Signatures and evidence hashes ensure snapshots are immutable and auditable.
- Snapshot IDs embed jurisdiction, domain, and semantic version (e.g., `snapshot:eu:psd3:2025-10`).

## 4. Security & Trust
- Snapshots must be signed by regulator-issued keys (DID/JWKS); verification fail-closed.
- Hashes of external datasets recorded in `evidence_sources`.

## 5. Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| RTGF_POL_MATRIX_SIGNATURE_INVALID | Signature verification fails | Reject snapshot |
| RTGF_POL_MATRIX_SCHEMA_INVALID | JSON-LD schema validation fails | Reject snapshot |

## 6. Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|-------|
| Snapshot validation latency | ≤ 1s P95 | including signature & schema checks |
| Snapshot freshness | ≤ 24h drift | regulators publish updates within 24h of change |

## 7. Interfaces & Integration
| Dependency | Direction | Purpose |
|------------|-----------|---------|
| Regulator signing service | inbound | Retrieve signed snapshots |
| RTGF compiler | outbound | Consume validated snapshots |
| Transparency log | outbound | Record snapshot metadata & hashes |

## 8. Metrics & Observability
- Prometheus: `rtgf_snapshot_validations_total{result}`, `rtgf_snapshot_signature_failures_total`.
- Logging: structured entries for snapshot ID, jurisdiction, domain, hash, signature key id.

## 9. Acceptance Tests
| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| RTGF-CT-00 | Valid signed snapshot | Accepted, recorded |
| RTGF-CT-01 | Invalid signature | Rejected with `RTGF_POL_MATRIX_SIGNATURE_INVALID` |
| RTGF-CT-02 | Schema violation | Rejected with `RTGF_POL_MATRIX_SCHEMA_INVALID` |

## 10. Acceptance Criteria
1️⃣ Validation process enforces signatures and schemas deterministically.  
2️⃣ Evidence hashes recorded for all referenced datasets.  
3️⃣ Compiler consumes snapshots via Policy Source Matrix with consistent IDs and provenance.

## Consequences
- ✅ Enables deterministic compiler outputs (RMT/IMT) once signatures and hashes are validated.  
- ✅ Provides clear provenance chain from legal source to enforcement token.  
- ⚠️ Requires regulators to maintain signing keys and update snapshots when laws or datasets change.

## Links
- *draft-lane2-rtgf-00* Section 3.1  
- `docs/architecture/policy-source-matrix.md`
