# ADR-RTGF-010: Operational SLOs & Observability for RTGF

**Status:** Accepted  
**Date:** 2025-11-02  
**Decision Makers:** RTGF Working Group  
**Owner:** Site Reliability Engineering & Operations Team  
**Related ADRs:** ADR-RTGF-003, ADR-RTGF-004, ADR-RTGF-005, ADR-RTGF-007, ADR-RTGF-009

**Planned Tests:** RTGF-CT-90, RTGF-CT-91, RTGF-CT-92, RTGF-CT-93, RTGF-CT-94

---

## 1. Purpose & Scope
Establish operational SLOs, observability standards, and incident response for RTGF services (compiler, registry, verifier, transparency, revocation) to satisfy corridor and regulatory uptime requirements.

## 2. Decision
Implement a unified observability stack:

- **Service Inventory:** `rtgf-compiler`, `rtgf-registry`, `rtgf-verifier`, `transparency-service`, `revocation-service`.  
- **SLO Controller:** central job calculates error budgets, enforces SLO alerts.  
- **Observability Stack:** Prometheus + Alertmanager, Grafana dashboards, OpenTelemetry collector, Loki/ELK for logs.  
- **Incident Response:** PagerDuty/Matrix integrations with severity matrix and runbooks.  
- **Configuration:** SLO policies declared via IaC repository (version-controlled).

## 3. Determinism & Provenance
- Metrics tagged with artefact versions (git commit, manifest ID) for correlation.  
- Alerts include build ID/token versions to accelerate replay.  
- Histogram buckets fixed across clusters for deterministic dashboards.  
- Observability configs versioned, transparency logged.

## 4. Security & Trust
- Observability endpoints protected via mTLS; dashboards require SSO/RBAC.  
- Logs/metrics scrub PII; token metadata aggregated.  
- Incident reports signed, archived with access controls.  
- Alert webhooks authenticated; rate limiting prevents alert storms.

## 5. Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| `RTGF_OP_SLO_BREACH` | Error budget exhausted | Trigger incident review |
| `RTGF_OP_DATA_GAP` | Metrics pipeline failure | Escalate to SRE |
| `RTGF_OP_ALERT_SUPPRESSED` | Alert not delivered | Failover channel |
| `RTGF_OP_LOG_DROP` | Log ingestion failure | Raise incident |

## 6. Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|-------|
| Registry availability | ≥ 99.95% per month | HTTPS endpoints |
| Verifier availability | ≥ 99.95% per month | `/verify` latency & uptime |
| Transparency publish latency | ≤ 5 min P95 | event → root publication |
| Revocation propagation | ≤ 60 s P95 | revEpoch updates |
| Compiler MTTR | ≤ 4 h | detect → rerun |

## 7. Interfaces & Integration
| Dependency | Direction | Purpose |
|------------|-----------|---------|
| Prometheus | inbound | Scrape metrics |
| Grafana | outbound | Dashboards |
| Alertmanager/PagerDuty | outbound | Incident notifications |
| Transparency log | inbound | Correlate operational events |
| Runbook repo | inbound | Link SOPs |

## 8. Observability
- Standard dashboards per service: latency, errors, saturation, revocation state.  
- Distributed tracing across registry ⇄ verifier ⇄ transparency.  
- Structured logs (request ID, token ID, decision, latencies).  
- Incident timeline auto-populated from alerts/events.

## 9. Planned Tests
| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| RTGF-CT-90 | SLO budget burn simulation | Triggers alert & incident workflow |
| RTGF-CT-91 | Metrics pipeline outage | Detects `RTGF_OP_DATA_GAP`, fails over |
| RTGF-CT-92 | Alert delivery failure | Backup channel engaged |
| RTGF-CT-93 | Revocation propagation latency | Meets ≤ 60 s requirement |
| RTGF-CT-94 | Post-incident report | Generated with signed summary |

## 10. Acceptance Criteria
1. Documented SLOs, dashboards, and alerting policies for all RTGF services; CT-90..94 pass.  
2. Metrics pipeline resilient; gaps detected/mitigated within 5 minutes.  
3. Incident response playbooks maintained; post-incident reviews signed and archived.  
4. Operational provenance ties incidents to artefact versions for auditability.

## 11. Consequences
- ✅ Shared operational model keeps corridor partners aligned on availability targets.  
- ✅ Observability instrumentation accelerates troubleshooting and compliance reporting.  
- ⚠️ SLO stack requires continuous SRE investment.  
- ⚠️ Aggressive noise reduction may miss warning signals; tuning required.
