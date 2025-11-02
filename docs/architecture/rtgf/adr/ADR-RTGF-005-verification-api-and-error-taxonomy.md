# ADR-RTGF-005: Verification API & Error Taxonomy — Rev for Acceptance

**Status:** Proposed → *Ready for Acceptance review*  
**Date:** 2025-11-02  
**Decision Makers:** RTGF Working Group  
**Owner:** RTGF Working Group  
**Target Acceptance:** 2026-01-31  
**Related ADRs:** ADR-RTGF-001, ADR-RTGF-002

---

## 1. Purpose & Scope
Define the RTGF verification service contract, including REST interfaces, error taxonomy, determinism requirements, and operational SLOs. The service validates tokens (RMT/IMT/CORT/PSRT) against revocation data, consents, and cryptographic proofs for routers, auditors, and corridor partners.

## 2. Architecture Overview
```
Verifier API (HTTP) → Request Validator → Token Resolver → Revocation Checker
                                            ↓
                                   Policy Evaluator (PPE)
                                            ↓
                                       Decision Engine → Response Serializer
```
- **Request Validator:** enforces schema, required token URIs, and auth headers.  
- **Token Resolver:** pulls canonical token JSON from Registry or cache.  
- **Revocation Checker:** validates revEpoch vs current state; consults transparency log when needed.  
- **Policy Evaluator:** optional inline PPE evaluation for runtime decisions (future).  
- **Decision Engine:** maps results to `Valid`, `Reason`, and control requirements; logs audit trail.

## 3. Determinism & Provenance
- Responses must be deterministic for identical inputs: `Valid`, `Reason`, `RevEpoch`.  
- All token payloads validated using canonical hash; mismatches flagged.  
- `RevEpoch` returned in responses is the current monotonic counter.  
- Audit log entries reference request ID, token URIs, hashes, and resolver source.  
- Idempotent POST `/verify` with same payload yields identical JSON (ordering, formatting).

## 4. Security & Trust
- API served over mTLS/TLS 1.3 with client certs for corridor partners; alt: OAuth2 for read-only endpoints.  
- Tokens retrieved from registry over mTLS; caches validate signature before storing.  
- Rate limiting per client; fail-closed on token fetch failures or revocation service timeouts.  
- Input sanitisation prevents injection; error responses exclude sensitive data.

## 5. Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| RTGF_VERIFY_MISSING_TOKEN | Required token URI absent | return 400 |
| RTGF_VERIFY_TOKEN_NOT_FOUND | Registry missing token | return 404 |
| RTGF_VERIFY_SIGNATURE_INVALID | Token signature fail | return 422 |
| RTGF_VERIFY_REVOKED | Token revoked/stale revEpoch | return 200 with `Valid=false` |
| RTGF_VERIFY_INTERNAL | Unexpected failure | return 500, alert SRE |
| RTGF_VERIFY_RATE_LIMIT | Throttle triggered | return 429 |

## 6. Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|------|
| Verify latency | ≤ 100 ms P95 | assuming cached tokens |
| Availability | ≥ 99.95 % | measured monthly |
| Rate-limit accuracy | ≤ 1 % false positives | track per client |
| Error budget | ≤ 0.1 % 5xx | across rolling 30 days |

## 7. Interfaces & Integration
| Dependency | Direction | Purpose |
|------------|-----------|---------|
| RTGF Registry | inbound | Fetch tokens, JWKS |
| Revocation service | inbound | Current revEpoch, revocation proofs |
| PPE Evaluator | outbound | Execute predicate decisions (future integration) |
| Observability (Prometheus, OTEL) | outbound | Metrics, traces |
| Transparency log | inbound/outbound | Validate token hash vs log entries |

## 8. Metrics & Observability
- Prometheus: `rtgf_verify_latency_ms_bucket`, `rtgf_verify_requests_total{result,code}`, `rtgf_verify_cache_hits_total`.  
- Logs: request ID, client ID, token URIs, decision, reasons.  
- OpenTelemetry span `rtgf.verify.request` with child spans for token resolve, revocation check.

## 9. Acceptance Tests
| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| RTGF-CT-40 | Valid token set | 200, `Valid=true`, revEpoch current |
| RTGF-CT-41 | Missing RMT URI | 400 with `RTGF_VERIFY_MISSING_TOKEN` |
| RTGF-CT-42 | Revoked token | 200, `Valid=false`, reason contains `token_revoked` |
| RTGF-CT-43 | Signature invalid | 422 with `RTGF_VERIFY_SIGNATURE_INVALID` |
| RTGF-CT-44 | Rate-limit scenario | 429 returned, metrics incremented |

## 10. Acceptance Criteria
1️⃣ `/verify` responses deterministic and conform to schema; CT-40..44 pass.  
2️⃣ Error taxonomy enforced in API responses and logs; 5xx budget maintained.  
3️⃣ Revocation and signature checks fail closed; tokens fetched via secure channels.  
4️⃣ Metrics, tracing, and audit logging implemented for observability.

## Consequences
- ✅ Unified verification API enables corridor partners and auditors to rely on consistent semantics.  
- ✅ Clear error taxonomy accelerates troubleshooting and compliance validation.  
- ⚠️ Strong security posture (mTLS, rate limits) increases integration effort for third parties.  
- ⚠️ Deterministic responses require careful cache management and version control.
