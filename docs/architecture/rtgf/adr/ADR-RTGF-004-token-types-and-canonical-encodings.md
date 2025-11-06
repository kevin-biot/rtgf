# ADR-RTGF-004: Token Types & Canonical Encodings

**Status:** Accepted  
**Date:** 2025-11-02  
**Decision Makers:** RTGF Working Group  
**Owner:** Token & Registry Subsystem Team  
**Related ADRs:** ADR-RTGF-001 (Policy Source Matrix), ADR-RTGF-002 (Sanctions Hashing), ADR-RTGF-003 (Compiler Pipeline)

**Planned Tests:** RTGF-CT-30, RTGF-CT-31, RTGF-CT-32, RTGF-CT-33, RTGF-CT-34

---

## 1. Purpose & Scope
Define the canonical envelope and encoding rules for RTGF-issued artefacts—Regulatory Matrix Tokens (RMT), International Mandate Tokens (IMT), Corridor Tokens (CORT), and Payment Service Route Tokens (PSRT). The specification must guarantee deterministic, self-describing tokens that regulators, verifiers, and routers can validate across jurisdictions.

## 2. Decision
Adopt a shared token envelope with deterministic serialization:

```json
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
    "merkleRoot": "sha256:…",
    "transparencyLog": "https://log.rtfg.example/…"
  },
  "predicate_set": { /* canonical PPE output */ },
  "eval_plan": { /* canonical plan output */ },
  "evidence": {
    "policy_snapshot": "snapshot:eu:psd3:2025-10",
    "sanctions_hash": "sha256:…",
    "inputs": [ /* evidence hashes */ ]
  },
  "signatures": [
    {
      "alg": "EdDSA",
      "kid": "ed25519:2025-10:k1",
      "jws": "…"
    }
  ]
}
```

Specialisation per token type:
- **RMT:** Contains jurisdictional obligations/controls.  
- **IMT:** Adds `corridor_id`, bilateral controls, combined evidence from the two RMTs.  
- **CORT:** Captures corridor operational metadata (operators, hand-off evidence).  
- **PSRT:** Encodes payment service routing requirements (acquirer, limits, compliance flags).

Producers MUST reuse this envelope; type-specific fields live inside `evidence` or namespaced sub-objects.

## 3. Determinism & Provenance
- Canonical JSON (RFC 8785) for the entire token envelope; keys sorted lexicographically, numbers fixed precision.  
- Canonical hash stored as `canonical_hash = "sha512:" + hex_digest`.  
- `build_manifest_id` embeds the compiler run (ADR-RTGF-003).  
- `revEpoch` baseline recorded at issuance; verifiers compare against revocation service.  
- Evidence hashes reference ontology artefacts and sanctions digests defined in ADR-RTGF-001/002.  
- Example canonical signature input:  
  ```json
  {"canonical_hash":"sha512:…","domain":"PSD3","effective":{"exp":"2026-10-01T00:00:00Z","nbf":"2025-10-01T00:00:00Z"},"evidence":{…},"issuer":"did:org:rtgf.eu","jurisdiction":"EU","revocation":{"merkleRoot":"sha256:…","revEpoch":42,"transparencyLog":"https://log.rtfg.example/…"},"token_id":"RMT:EU:PSD3:2025-10","type":"RMT","version":"2025.10"}
  ```

## 4. Security & Trust
- Tokens signed with Ed25519 keys per issuer/regulator; multiple signatures supported (`signatures[]`).  
- JWKS served at `/.well-known/jwks.json`; verifiers refresh on kid rotation (ADR-RTGF-006).  
- Consumers must validate: schema, canonical hash, signatures, revocation status, transparency proof (if supplied).  
- Signatures cover canonical JSON excluding the `signatures` array.

## 5. Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| `RTGF_TOKEN_CANONICAL_MISMATCH` | Recomputed canonical hash differs | Reject token, log integrity alert |
| `RTGF_TOKEN_SCHEMA_INVALID` | Schema validation fails | Reject, raise telemetry |
| `RTGF_TOKEN_SIGNATURE_INVALID` | Signature verification fails | Reject, trigger JWKS refresh |
| `RTGF_TOKEN_TYPE_UNSUPPORTED` | Unknown `type` | Reject, alert operators |
| `RTGF_TOKEN_REVOCATION_STALE` | revEpoch older than registry state | Reject pending refresh |

## 6. Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|-------|
| `rtgf_token_canonical_duration_seconds_bucket` | ≤ 50 ms P95 | compiler/runtime check |
| Signature validation success | ≥ 99.99% | per verification request |
| Schema validation coverage | 100% | enforced in CI |
| Revocation freshness | revEpoch drift ≤ 1 | monitors stale tokens |

## 7. Interfaces & Integration
| Dependency | Direction | Purpose |
|------------|-----------|---------|
| PPE Compiler | inbound | Provide predicate/eval plan JSON |
| Transparency Log | outbound | Persist token metadata + hashes |
| Revocation Service | inbound/outbound | Baseline revEpoch + updates |
| RTGF Registry API | outbound | Serve tokens under `/tokens` routes |
| Verification SDKs | outbound | Consume schema + error taxonomy |

## 8. Observability
- Prometheus: `rtgf_token_canonical_duration_seconds_bucket`, `rtgf_token_signature_failures_total{type}`, `rtgf_token_schema_invalid_total`, `rtgf_token_revocation_stale_total`.  
- Logs record token ID, canonical hash, signature KIDs, schema version.  
- CI pipeline publishes schema validation coverage and canonical diff snapshots.

## 9. Planned Tests
| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| RTGF-CT-30 | Canonicalise token twice | Identical `canonical_hash`, byte-identical canonical JSON |
| RTGF-CT-31 | Missing signature | Verifier returns `RTGF_TOKEN_SIGNATURE_INVALID` |
| RTGF-CT-32 | Unsupported type | Verifier returns `RTGF_TOKEN_TYPE_UNSUPPORTED` |
| RTGF-CT-33 | Schema mutation (extra field) | Schema validator rejects token |
| RTGF-CT-34 | RevEpoch mismatch | Runtime verifier flags `RTGF_TOKEN_REVOCATION_STALE` |

## 10. Acceptance Criteria
1. Schema + canonicalisation rules published and enforced; tokens hashed/signatured deterministically.  
2. Verification library implements schema/canonical hash/signature/revocation checks consistent with this ADR.  
3. Registry serves canonical tokens; clients can recompute hashes matching transparency entries.  
4. Error taxonomy surfaces in logs/metrics and is consumed by downstream SDKs; CT-30..34 passing in CI.

## 11. Consequences
- ✅ Deterministic encodings enable cross-language verifiers and auditability.  
- ✅ Clear contract simplifies corridor interoperability and replay analysis.  
- ⚠️ Strict canonicalisation requires coordinated schema versioning across repos.  
- ⚠️ Multiple signatures and revocation metadata add latency; capacity planning required.
