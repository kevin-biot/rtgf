# ADR-RTGF-003: Compiler & Deterministic Build Pipeline — Rev for Acceptance

**Status:** Proposed → *Ready for Acceptance review*  
**Date:** 2025-11-02  
**Decision Makers:** RTGF Working Group  
**Owner:** RTGF Working Group  
**Target Acceptance:** 2026-01-31  
**Related ADRs:** ADR-RTGF-001, ADR-RTGF-002

---

## 1. Purpose & Scope
Define the RTGF compiler pipeline that transforms policy snapshots into deterministic RMT/IMT/CORT tokens, guaranteeing reproducibility, provenance, and security boundaries across jurisdictions and domains.

## 2. Architecture Overview
```
PolicySnapshotFetcher → SchemaValidator → DeterministicPlanner → PredicateCompiler
      ↓                                            ↓
EvidenceHasher --------------------------> TokenAssembler → ArtefactSigner → TransparencyLogger
```
- **PolicySnapshotFetcher:** retrieves signed snapshots per ADR-RTGF-001.  
- **SchemaValidator:** enforces JSON-LD schema & signature validation.  
- **DeterministicPlanner:** sorts predicates, applies canonical ordering, builds eval plans.  
- **PredicateCompiler:** generates predicate sets (PPE compiler).  
- **EvidenceHasher:** records external dataset hashes (ADR-RTGF-002).  
- **TokenAssembler:** constructs RMT/IMT with canonical fields, references.  
- **ArtefactSigner:** applies Ed25519 signatures using controlled keys.  
- **TransparencyLogger:** appends artefact metadata/hashes for audit.

## 3. Determinism & Provenance
- Inputs pinned by snapshot version and evidence hashes.  
- JSON serialization via RFC 8785 canonical form; stable key sorting.  
- Build graph recorded in `build_manifest.json` with tool versions.  
- Hashes: `sha256` for predicate plans, `sha512` for token body, both logged.  
- revEpoch increments on token revocation; compile step captures snapshot revEpoch baseline.  
- Deterministic environment variables: `TZ=UTC`, `LANG=C`, `RTGF_SEED=1337`.

## 4. Security & Trust
- Compiler runs inside hermetic container; dependencies hashed.  
- Input snapshots verified against regulator JWKS; fail-closed.  
- Signing keys stored in HSM; access via mTLS.  
- ArtefactSigner attaches JWS with key rotation policy (ADR-RTGF-006).  
- Pipeline requires authenticated operator invocation (RBAC).

## 5. Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| RTGF_BUILD_SNAPSHOT_INVALID | Signature/schema failure | halt build |
| RTGF_BUILD_PLAN_DIVERGENCE | Non-deterministic ordering detected | fail |	null |
| RTGF_BUILD_SIGNING_ERROR | HSM signing failure | retry ×3 then fail |
| RTGF_BUILD_HASH_MISMATCH | Post-signature hash mismatch | abort, alert SRE |

## 6. Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|-------|
| Build duration | ≤ 5 min per jurisdiction/domain | assuming cached datasets |
| Plan determinism checks | 100 % pass rate | double-run diff |
| Signing success rate | ≥ 99.99 % | per artefact |

## 7. Interfaces & Integration
| Dependency | Direction | Purpose |
|------------|-----------|---------|
| Policy Source Matrix | inbound | Consume signed snapshots |
| Sanctions dataset fetcher | inbound | Provide evidence hashes |
| Transparency log | outbound | Append build manifest, token metadata |
| Revocation service | outbound | Seed revEpoch baseline |
| PPE evaluator test harness | outbound | Validate tokens via round-trip |

## 8. Metrics & Observability
- Prometheus: `rtgf_build_duration_seconds_bucket`, `rtgf_build_failures_total{code}`, `rtgf_artifact_sign_total`.  
- Structured logs include snapshot IDs, tool versions, digests.  
- OpenTelemetry trace: span `rtgf.compiler.run`, child spans per stage.

## 9. Acceptance Tests
| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| RTGF-CT-20 | Compile snapshot twice | identical digests, no diff |
| RTGF-CT-21 | Invalid snapshot signature | build fails with `RTGF_BUILD_SNAPSHOT_INVALID` |
| RTGF-CT-22 | Evidence hash mismatch | build aborts, transparency log untouched |
| RTGF-CT-23 | Artefact signing failure simulation | retries then surfaces `RTGF_BUILD_SIGNING_ERROR` |

## 10. Acceptance Criteria
1️⃣ Compiler outputs deterministic artefacts (predicate set, eval plan, tokens) with identical digests across runs.  
2️⃣ All artefacts signed and logged with provenance metadata.  
3️⃣ Pipeline fails closed on input validation, hash mismatch, or signing errors.  
4️⃣ Metrics and traces emitted for each stage; CT-20..23 pass.

## Consequences
- ✅ Reproducible tokens build trust with regulators and verifiers.  
- ✅ Transparency hooks facilitate independent audits.  
- ⚠️ Hermetic builds require disciplined dependency management and HSM integration.
