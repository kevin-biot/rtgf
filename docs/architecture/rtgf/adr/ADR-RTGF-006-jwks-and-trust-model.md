# ADR-RTGF-006: JWKS, DANE/TLSA & Trust Model — Rev for Acceptance

**Status:** Proposed → *Ready for Acceptance review*  
**Date:** 2025-11-02  
**Decision Makers:** RTGF Working Group  
**Owner:** RTGF Working Group  
**Target Acceptance:** 2026-01-31  
**Related ADRs:** ADR-RTGF-001, ADR-RTGF-002

---

## 1. Purpose & Scope
Establish the trust and key-management model for RTGF artefacts: publishing JWKS endpoints, handling key rotation, anchoring trust via DANE/TLSA, and enforcing mTLS between components. This ADR ensures verifiers can authenticate issuers and regulators across corridors.

## 2. Architecture Overview
```
HSM/KMS → Key Signing Service → JWKS Generator → Registry / .well-known
                                  ↓
                             Transparency Log
```
- **Key Signing Service (KSS):** interfaces with HSM to generate Ed25519 key pairs and sign artefacts.  
- **JWKS Generator:** produces JWKS documents with `kid`, `alg`, `use`, `createdAt`, `expiresAt`.  
- **Registry JWKS Endpoint:** served at `/.well-known/rtgf/jwks.json` and `/jwks.json`.  
- **Trust Anchors:** optional DANE/TLSA records for registry domains; backup trust via DID documents.  
- **Client Enforcement:** verifiers fetch JWKS, cache per `expiresAt`, and validate tokens.

## 3. Determinism & Provenance
- JWKS documents canonicalised (sorted keys) with `sha256` digest logged in transparency ledger.  
- Key lifecycle recorded with `key_manifest.json` (creation timestamp, rotation schedule).  
- `kid` format: `<issuer>:<alg>:<yyyy-mm>:<sequence>` to ensure deterministic mapping.  
- rotated keys overlap for 7 days to allow signature verification during rollout.

## 4. Security & Trust
- HSM-backed key storage with audit logging; no raw private keys on disk.  
- JWKS served over TLS 1.3 with HSTS; optionally signed JWT for JWKS payload.  
- mTLS between registry, compiler, and verifier when exchanging keys or artefacts.  
- DANE/TLSA records published for registry domain to bind certificates.  
- Key compromise response: revoke kid, publish emergency JWKS with `revoked=true`, notify partners.

## 5. Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| RTGF_JWKS_FETCH_FAILED | Unable to retrieve JWKS | retry with backoff, fail closed |
| RTGF_JWKS_SIGNATURE_INVALID | JWKS signed hash mismatch | reject JWKS |
| RTGF_KID_EXPIRED | Key past `expiresAt` | reject signature, request refresh |
| RTGF_KID_UNKNOWN | Token references unknown kid | fail verification, alert |

## 6. Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|------|
| JWKS fetch latency | ≤ 200 ms P95 | CDN cached |
| Key rotation window | ≤ 7 days | overlapping validity |
| Key compromise detection | ≤ 15 min mean-time-to-detect | via monitoring |
| JWKS availability | ≥ 99.99 % | served via HA registry |

## 7. Interfaces & Integration
| Dependency | Direction | Purpose |
|------------|-----------|---------|
| HSM/KMS | inbound | Generate/store keys |
| Registry API | outbound | Serve JWKS |
| Transparency log | outbound | Anchor JWKS digests |
| Verifiers & partners | outbound | Provide trust anchors |
| DNS (DANE/TLSA) | outbound | Publish TLS bindings |

## 8. Metrics & Observability
- Prometheus: `rtgf_jwks_fetch_total{result}`, `rtgf_key_rotation_events_total`, `rtgf_kid_unknown_total`.  
- Audit log entries on key creation, rotation, revocation (include actor, reason).  
- Alerting on JWKS fetch failures, DANE/TLSA mismatches, expired keys in use.

## 9. Acceptance Tests
| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| RTGF-CT-50 | Fetch JWKS | matches canonical digest, schema valid |
| RTGF-CT-51 | Rotate key (overlap) | old & new signatures accepted within grace window |
| RTGF-CT-52 | Use expired kid | verification fails with `RTGF_KID_EXPIRED` |
| RTGF-CT-53 | DANE/TLSA mismatch | clients detect and fail connection |
| RTGF-CT-54 | JWKS tampered | consumer rejects due to digest mismatch |

## 10. Acceptance Criteria
1️⃣ JWKS publishing automated with transparency logging; CT-50..54 pass.  
2️⃣ Key rotation documented, overlapping validity ensures no verification gaps.  
3️⃣ mTLS and DANE/TLSA integrations operational; invalid or expired keys handled fail-closed.  
4️⃣ Monitoring in place for key lifecycle events and fetch failures.

## Consequences
- ✅ Strong trust foundation for token verification and corridor interoperability.  
- ✅ Formal lifecycle management reduces incident response time.  
- ⚠️ HSM integration and DANE setup add operational overhead.  
- ⚠️ Strict expiration may cause outages if partners fail to refresh keys promptly; need clear communications.
