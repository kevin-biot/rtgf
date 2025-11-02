# ADR-RTGF-009: RTGF Compilation Pipeline & Reproducibility — Rev for Acceptance

**Status:** Proposed → *Ready for Acceptance review*  
**Date:** 2025-11-02  
**Decision Makers:** RTGF Working Group  
**Owner:** RTGF Working Group  
**Target Acceptance:** 2026-01-31  
**Related ADRs:** ADR-RTGF-001, ADR-RTGF-002

---

## 1. Purpose & Scope
Document the end-to-end RTGF pipeline orchestration covering snapshot ingestion, compilation, verification, transparency logging, and deployment to registry. Defines deterministic build graph, scheduling, and change management.

## 2. Architecture Overview
```
Snapshot Watcher → Build Orchestrator → Compiler (ADR-RTGF-003) → Artefact QA
        ↓                                           ↓                     ↓
Dataset Fetcher (ADR-RTGF-002) → Determinism Harness → Transparency Log → Registry Publish
                                                   ↓
                                          Notification Hub (Ops/Partners)
```
- **Snapshot Watcher:** detects new/updated policy snapshots, validates signatures.  
- **Build Orchestrator:** runs deterministic builds via hermetic containers, captures logs.  
- **Artefact QA:** verifies tokens using deterministic harness (tests/ppe-roundtrip), schema checks.  
- **Transparency Log:** records all activities.  
- **Registry Publish:** pushes signed artefacts, JWKS updates, revocation baseline.

## 3. Determinism & Provenance
- Build manifests capture toolchain versions (`compiler`, `node`, `go`, `npm`).  
- Build orchestrator runs twice; digests compared before promotion (Phase 2 harness).  
- Randomness seeded (`RTGF_SEED`); environment variables pinned (TZ, LANG).  
- Pipeline stores provenance bundle: snapshots, dataset hashes, compiler logs, deterministic manifest.  
- Git commit hash recorded for policy source & compiler code used per run.

## 4. Security & Trust
- Pipeline triggers require signed approvals; RBAC enforced.  
- Build steps run in isolated runners with no outbound network except whitelisted dataset providers.  
- Secrets stored in vault; access tokens rotated.  
- Publishing to registry uses mTLS and signed manifests; fail closed if transparency logging fails.  
- Notifications (email/Matrix) send signed messages summarizing build.

## 5. Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| RTGF_PIPELINE_DETERMINISM_FAIL | Double-run digests mismatch | halt promotion |
| RTGF_PIPELINE_SNAPSHOT_FAIL | Snapshot validation failure | abort build |
| RTGF_PIPELINE_QA_FAIL | Schema/verification tests fail | block publish |
| RTGF_PIPELINE_PUBLISH_FAIL | Registry publish error | retry + alert |

## 6. Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|------|
| Build pipeline duration | ≤ 15 min | per jurisdiction/domain |
| Determinism validation | 100 % | no mismatches |
| Failed builds resolved | ≤ 24 h P90 | from detection to fix |
| Registry publish success | ≥ 99.9 % | tracked monthly |

## 7. Interfaces & Integration
| Dependency | Direction | Purpose |
|------------|-----------|---------|
| Snapshot repo | inbound | Source inputs |
| Dataset fetcher | inbound | Provide evidence materials |
| Transparency log | outbound | Append build events |
| Registry | outbound | Deploy artefacts |
| Notification hub | outbound | Inform stakeholders |
| CI/CD system | inbound/outbound | Orchestrate builds |

## 8. Metrics & Observability
- Prometheus: `rtgf_pipeline_runs_total{result}`, `rtgf_pipeline_duration_seconds`, `rtgf_pipeline_determinism_failures_total`.  
- Build logs stored in immutable object storage with retention policy; referenced in transparency records.  
- Dashboard showing pipeline status by jurisdiction, commit hash, dataset version.

## 9. Acceptance Tests
| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| RTGF-CT-80 | Full build success | digests match, artefacts published |
| RTGF-CT-81 | Determinism mismatch injection | pipeline halts (no publish) |
| RTGF-CT-82 | Snapshot validation failure | pipeline stops with `RTGF_PIPELINE_SNAPSHOT_FAIL` |
| RTGF-CT-83 | Registry publish failure simulation | retries then alerts ops |
| RTGF-CT-84 | Transparency logging unreachable | pipeline fails closed |

## 10. Acceptance Criteria
1️⃣ Pipeline guarantees deterministic builds with double-run verification; CT-80..84 green.  
2️⃣ Transparency logging and registry publish steps are mandatory; failure halts release.  
3️⃣ Metrics and alerting in place for determinism failures, publish errors, and build duration regressions.  
4️⃣ Provenance bundles stored with each release for audit.

## Consequences
- ✅ Provides auditable, deterministic release process trusted by regulators.  
- ✅ Automation reduces manual errors and ensures consistent artefact quality.  
- ⚠️ Strict controls may delay urgent fixes if pipeline fails; requires incident runbooks.  
- ⚠️ Running duplicate builds increases resource usage; plan capacity accordingly.
