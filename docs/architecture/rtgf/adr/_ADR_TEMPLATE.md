# ADR-XXX: <Title> — Rev for Acceptance

**Status:** Proposed → *Ready for Acceptance review*  
**Date:** YYYY-MM-DD  
**Decision Makers:** <Names / Group>  
**Owner:** <Team / Role>  
**Target Acceptance:** YYYY-MM-DD  
**Related ADRs:** <List other ADRs / specs>

---

## 1. Purpose & Scope
<What problem do we solve, for whom, and why now? Link to Context.>

## 2. Architecture Overview
<Describe the essential architecture, key components, and data flow. Include diagrams where useful.>

## 3. Determinism & Provenance
<How is determinism achieved (JCS canonicalization, fixed seeds, build graph pinning)? What is recorded in revEpoch? How are proofs anchored?>

## 4. Security & Trust
<mTLS, JWKS, DANE/TLSA, signature scheme, consent/policy gates, fail-closed behaviour.>

## 5. Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| XYZ_… | … | … |

## 6. Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|-------|
| … | … | … |

Prometheus metrics:  
`component_metric_a`, `component_metric_b`, …

## 7. Interfaces & Integration
| Dependency | Direction | Purpose |
|------------|-----------|---------|
| …          | …         | …       |

## 8. Metrics & Observability
<Call out key dashboards, tracing, and logging requirements.>

## 9. Acceptance Tests
| Test ID | Scenario | Expected outcome |
|---------|----------|------------------|
| RTGF-CT-XX | … | … |

## 10. Acceptance Criteria
1️⃣ …  
2️⃣ …  
3️⃣ …  

## Consequences
✅ … benefits…  
⚠️ … trade-offs…
