# ADR-RTGF-004: Token Types & Canonical Encodings (RMT/IMT/CORT/PSRT) — Rev for Acceptance

**Status:** Proposed → *Ready for Acceptance review*  
**Date:** 2025-11-02  
**Decision Makers:** RTGF Working Group  
**Owner:** RTGF Working Group  
**Target Acceptance:** 2026-01-31  
**Related ADRs:** ADR-RTGF-001, ADR-RTGF-002

---

## 1. Purpose & Scope
Define the canonical data model and encoding rules for all RTGF-issued artefacts: Regulatory Matrix Tokens (RMT), International Mandate Tokens (IMT), Corridor Tokens (CORT), and Payment Service Route Tokens (PSRT). These encodings must be deterministic, self-describing, and verifiable across jurisdictions, enabling regulators, verifiers, and operators to rely on a stable contract.

## 2. Architecture Overview
Token structure builds on a shared envelope:
```
{
  "token_id": "RMT:EU:PSD3:2025-10",
  "type": "RMT",
  "version": "2025.10",
  "issuer": "did:org:rtgf.eu",
  "jurisdiction": "EU",
  "domain": "PSD3",
  "effective": {
    "nbf": "2025-10-01T00:00:00Z",
    "exp": "2026-10-01T00:00:00Z"
  },
  "revocation": {
    "revEpoch": 42,
    "merkleRoot": "sha256:...",
    "transparencyLog": "https://log.rtfg.example/..."
  },
  "predicate_set": {...},
  "eval_plan": {...},
  "evidence": {
    "policy_snapshot": "snapshot:eu:psd3:2025-10",
    "sanctions_hash": "sha256:...",
    "inputs": [...]
  },
  "signatures": [
    {
      "alg": "EdDSA",
      "kid": "ed25519:2025-10:k1",
      "jws": "..."
    }
  ]
}
```
Differences per token type:
- **RMT:** jurisdiction-specific policy obligations.  
- **IMT:** intersection between two RMTs; includes `corridor_id` and bilateral controls.  
- **CORT:** corridor operational metadata (route operators, controls).  
- **PSRT:** payment service routing obligations (acquirer-specific).

## 3. Determinism & Provenance
- Canonical encoding uses RFC 8785 JSON Canonicalization Scheme (JCS); keys sorted lexicographically, numbers fixed precision.  
- Token digests: `sha512` of canonical JSON stored as `canonical_hash`.  
- All tokens embed `build_manifest_id` linking back to ADR-RTGF-003 compiler run.  
- `revEpoch` originates from revocation service; tokens carry baseline value to detect stale states.  
- Evidence hashes reference ADR-RTGF-001/002 assets; transparency log entries include the same digests.

## 4. Security & Trust
- Tokens signed with Ed25519 keys managed per issuer; multiple signatures supported (issuer, regulator).  
- Signature block includes algorithm, key id, and detached JWS (payload canonical).  
- JWKS endpoints published via RTGF registry (ADR-RTGF-006) with kid rotation schedule.  
- Verification rules: must validate canonical hash, signature(s), and revocation status before accepting token.

## 5. Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| RTGF_TOKEN_CANONICAL_MISMATCH | Recomputed digest differs | reject token |
| RTGF_TOKEN_SCHEMA_INVALID | Envelope fails schema validation | reject token |
| RTGF_TOKEN_SIGNATURE_INVALID | Signature/JWS verification fails | reject token |
| RTGF_TOKEN_TYPE_UNSUPPORTED | Unknown `type` field | reject and alert |

## 6. Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|------|
| Canonicalisation latency | ≤ 50 ms P95 per token | measured in compiler & verifier |
| Signature validations | ≥ 99.99 % success | per verification request |
| Schema validation coverage | 100 % | enforce via CI tests |

## 7. Interfaces & Integration
| Dependency | Direction | Purpose |
|------------|-----------|---------|
| PPE Compiler | inbound | Provides predicate set & eval plan fragments |
| Transparency Log | outbound | Stores canonical hash & metadata |
| RTGF Registry API | outbound | Serves tokens under `/tokens` routes |
| Revocation Service | inbound/outbound | Supplies revEpoch baseline & updates |
| External verifiers | outbound | Standardised token schema for runtime validation |

## 8. Metrics & Observability
- Prometheus: `rtgf_token_canonical_duration_seconds_bucket`, `rtgf_token_signature_failures_total{type}`, `rtgf_token_schema_invalid_total`.  
- Structured audit logs capturing token ID, hash, signatures, and revEpoch.  
- Schema validation CI pipeline outputs coverage reports.

## 9. Acceptance Tests
| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| RTGF-CT-30 | Canonicalise token twice | identical `canonical_hash`, byte-equal JCS output |
| RTGF-CT-31 | Missing signature | verification fails with `RTGF_TOKEN_SIGNATURE_INVALID` |
| RTGF-CT-32 | Unsupported type | validation rejects with `RTGF_TOKEN_TYPE_UNSUPPORTED` |
| RTGF-CT-33 | Schema mutation (extra field) | schema validator rejects token |
| RTGF-CT-34 | Verify revEpoch mismatch | runtime verifier flags stale token |

## 10. Acceptance Criteria
1️⃣ Token schema definitions published with versioning; canonical JSON generation produces consistent digests across builds.  
2️⃣ Verification library validates schema, signatures, and revocation using shared code paths; CT-30..34 green.  
3️⃣ Registry serves canonical tokens; clients can recompute hashes that match transparency log entries.  
4️⃣ Error taxonomy enforced in logging and API responses.

## Consequences
- ✅ Deterministic encodings enable cross-language verifiers and auditability.  
- ✅ Clear contract supports corridor interoperability and revocation handling.  
- ⚠️ Strict canonicalisation increases sensitivity to schema changes—requires coordinated versioning.  
- ⚠️ Additional signatures per token increase build and verification latency; need capacity planning.
