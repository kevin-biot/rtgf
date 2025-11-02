# ADR-RTGF-010: Operational SLOs & Observability for RTGF — Rev for Acceptance

**Status:** Proposed → *Ready for Acceptance review*  
**Date:** 2025-11-02  
**Decision Makers:** RTGF Working Group  
**Owner:** RTGF Working Group  
**Target Acceptance:** 2026-01-31  
**Related ADRs:** ADR-RTGF-001, ADR-RTGF-002

---

## 1. Purpose & Scope
Define operational service levels, observability, and incident response for RTGF components (compiler, registry, verifier, transparency, revocation). This ADR ensures corridors meet regulatory uptime commitments and provides shared dashboards, alerting, and runbooks.

## 2. Architecture Overview
- **Service Inventory:** `rtgf-compiler`, `rtgf-registry`, `rtgf-verifier`, `transparency-service`, `revocation-service`.  
- **SLO Management:** central SLO controller aggregates metrics, calculates error budgets.  
- **Observability Stack:** Prometheus, Grafana, OpenTelemetry collector, Loki/ELK for logs.  
- **Incident Response:** PagerDuty/Matrix integrations, with documented severity matrix.  
- **Configuration:** SLO policies declared via Infrastructure-as-Code repository.

## 3. Determinism & Provenance
- Operational metrics tied to specific artefact versions (hash, git commit) to correlate behaviour.  
- Every alert includes the build ID and token versions served to enable provenance.  
- SLO histograms use deterministic bucket boundaries to ensure comparable metrics across clusters.  
- Observability pipelines versioned and logged in transparency service for audit.

## 4. Security & Trust
- Observability endpoints secured via mTLS; read-only dashboards require SSO.  
- Metrics and logs scrub PII; only token metadata and aggregates stored.  
- Incident reports signed and archived; access controlled.  
- Alert channels require authenticated webhooks; noise reduction via rate limiting.

## 5. Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| RTGF_OP_SLO_BREACH | Error budget exhaustion | trigger incident review |
| RTGF_OP_DATA_GAP | Metrics pipeline failure | escalate to SRE |
| RTGF_OP_ALERT_SUPPRESSED | Alert not delivered | failover channel |
| RTGF_OP_LOG_DROP | Log ingestion failure | raise priority incident |

## 6. Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|------|
| Registry availability | ≥ 99.95 % | per month |
| Verifier availability | ≥ 99.95 % | per month |
| Transparency publish latency | ≤ 5 min P95 | from event to root |
| Revocation propagation | ≤ 60 s P95 | revEpoch update |
| Compiler MTTR | ≤ 4 h | from failure to rerun |

## 7. Interfaces & Integration
| Dependency | Direction | Purpose |
|------------|-----------|---------|
| Prometheus | inbound | metrics scraping |
| Grafana | outbound | SLO dashboards |
| Alertmanager/PagerDuty | outbound | incident notifications |
| Transparency log | inbound | correlate operational events |
| Runbook repo | inbound | link SOPs |

## 8. Metrics & Observability
- Standard dashboards per service: latency, errors, saturation, revocation state.  
- Tracing: distributed traces across registry ⇄ verifier ⇄ transparency.  
- Logs: structured JSON containing request ID, token ID, decision, latencies.  
- Incident timeline: automatically populated from alerts/events.

## 9. Acceptance Tests
| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| RTGF-CT-90 | SLO budget burn simulation | triggers alert & incident workflow |
| RTGF-CT-91 | Metrics pipeline outage | detects `RTGF_OP_DATA_GAP`, fails over |
| RTGF-CT-92 | Alert delivery failure | backup channel engaged |
| RTGF-CT-93 | Revocation propagation latency test | meets ≤ 60 s requirement |
| RTGF-CT-94 | Post-incident report | generated with signed summary |

## 10. Acceptance Criteria
1️⃣ Documented SLOs, dashboards, and alerting policies for all RTGF services; CT-90..94 green.  
2️⃣ Metrics pipelines resilient; data gaps detected and mitigated within 5 minutes.  
3️⃣ Incident response playbooks maintained; post-incident reviews signed and archived.  
4️⃣ Operational provenance ties incidents to artefact versions for auditability.

## Consequences
- ✅ Shared operational model keeps corridor partners aligned on availability targets.  
- ✅ Observability instrumentation accelerates troubleshooting and compliance reporting.  
- ⚠️ Maintaining SLO stack requires dedicated SRE investment.  
- ⚠️ Strict noise reduction may hide early warning signals if thresholds misconfigured.
