# ADR-RTGF-007: Transparency & Proof Bundles for RTGF — Rev for Acceptance

**Status:** Proposed → *Ready for Acceptance review*  
**Date:** 2025-11-02  
**Decision Makers:** RTGF Working Group  
**Owner:** RTGF Working Group  
**Target Acceptance:** 2026-01-31  
**Related ADRs:** ADR-RTGF-001, ADR-RTGF-002

---

## 1. Purpose & Scope
Define the RTGF transparency and proof system that anchors policy snapshots, compiler artefacts, tokens, and revocation events into append-only logs. Proof bundles allow external auditors and partners to verify integrity and inclusion without trusted intermediaries.

## 2. Architecture Overview
```
Event Producers (compiler, revocation, jwks) → Transparency Ingest → Merkle Tree Builder
                                                             ↓
                                                       Proof Generator → Bundles API
                                                             ↓
                                                  External Auditors / Verifiers
```
- **Transparency Ingest:** accepts signed events (`policy_snapshot`, `token_issue`, `revocation`, `jwks_publish`).  
- **Merkle Tree Builder:** batches events, updates Merkle root, signs tree heads.  
- **Proof Generator:** provides inclusion proofs (RFC 6962 style) and consistency proofs.  
- **Bundles API:** exposes `/proofs/<event_id>` and `/roots/latest` via registry.

## 3. Determinism & Provenance
- All events serialized via JCS before hashing (`sha256`).  
- Event IDs follow `evt:<type>:<timestamp>:<uuid>` with deterministic timestamp (UTC).  
- Signed tree heads include `tree_size`, `root_hash`, `signature`.  
- Proof bundles packaged as deterministic JSON:
```json
{
  "eventId": "evt:token_issue:2025-11-02T10:00:00Z:1234",
  "merkleRoot": "sha256:...",
  "leafHash": "sha256:...",
  "path": ["sha256:...", "..."],
  "consistency": ["sha256:..."],
  "issuedAt": "2025-11-02T10:00:05Z",
  "signatures": [...]
}
```

## 4. Security & Trust
- Transparency service authenticated via mTLS; ingestion requires signed payloads.  
- Tree head signatures use dedicated transparency key pair anchored in ADR-RTGF-006 trust model.  
- Access control on proof API (read-mostly) but public auditors allowed.  
- Monitoring for append-only property; detection of inconsistent roots triggers incident.

## 5. Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| RTGF_TXP_EVENT_INVALID | Ingest payload fails schema/signature | reject |
| RTGF_TXP_APPEND_VIOLATION | Attempted append breaks Merkle invariants | halt, alert |
| RTGF_TXP_PROOF_NOT_FOUND | Event ID absent | return 404 |
| RTGF_TXP_ROOT_SIGNATURE_INVALID | Tree head signature fail | mark root invalid, alert |

## 6. Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|------|
| Proof generation latency | ≤ 500 ms P95 | inclusion proof |
| Root publication interval | ≤ 5 min | running tree head exposure |
| Append symmetry checks | 100 % | each append validated |
| Availability | ≥ 99.9 % | Bundles API uptime |

## 7. Interfaces & Integration
| Dependency | Direction | Purpose |
|------------|-----------|---------|
| Compiler | inbound | Record token issuance events |
| Revocation service | inbound | Append revocation updates |
| JWKS generator | inbound | Anchor key rotations |
| Auditors/verifiers | outbound | Provide proof bundles |
| External logs (optional cross-log) | outbound | Gossip / consistency proofs |

## 8. Metrics & Observability
- Prometheus: `rtgf_transparency_ingest_total{result,type}`, `rtgf_transparency_proof_latency_ms_bucket`, `rtgf_transparency_root_publish_total`.  
- Audit logs for each event with leaf hash, tree size.  
- Alerting on append violations, proof failures, or root signature issues.

## 9. Acceptance Tests
| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| RTGF-CT-60 | Issue token & retrieve proof | proof verifies against root |
| RTGF-CT-61 | Simulate tampered payload | ingestion rejects with `RTGF_TXP_EVENT_INVALID` |
| RTGF-CT-62 | Request unknown event | returns 404 `RTGF_TXP_PROOF_NOT_FOUND` |
| RTGF-CT-63 | Consistency proof check | verifies across sequential roots |
| RTGF-CT-64 | Append violation test | service halts and raises alert |

## 10. Acceptance Criteria
1️⃣ All issuance, revocation, and key events recorded with deterministic hashes and retrievable proofs (CT-60..64 green).  
2️⃣ Transparency log enforces append-only property with signed roots and consistency checks.  
3️⃣ Proof bundle format documented; verifiers can validate tokens using bundles without trusted access.  
4️⃣ Monitoring detects anomalies in root publication or append operations.

## Consequences
- ✅ External auditors can verify the integrity of RTGF artefacts without trusting the issuer.  
- ✅ Cross-log gossip enables detection of equivocation.  
- ⚠️ Operating a transparency log requires reliable storage and signature infrastructure.  
- ⚠️ Clients must implement proof verification to gain full benefits.
