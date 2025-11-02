# ADR-RTGF-008: Policy Rule Algebra & Token Generation Semantics — Rev for Acceptance

**Status:** Proposed → *Ready for Acceptance review*  
**Date:** 2025-11-02  
**Decision Makers:** RTGF Working Group  
**Owner:** RTGF Working Group  
**Target Acceptance:** 2026-01-31  
**Related ADRs:** ADR-RTGF-001, ADR-RTGF-002

---

## 1. Purpose & Scope
Define the semantics for translating policy source matrix entries into executable predicates and controls within RMT/IMT/CORT tokens. This ADR covers rule algebra, precedence, and mapping of legal requirements to deterministic evaluation logic.

## 2. Architecture Overview
```
Policy Snapshot (obligations, controls) → Policy Mapper → Predicate Library Resolver
                                                        ↓
                                               Predicate Assembly → Control Catalogue Linker
                                                        ↓
                                                Token Generator (ADR-RTGF-003)
```
- **Policy Mapper:** interprets snapshot clauses (duty/prohibition/control) and maps them to predicate templates.  
- **Predicate Library Resolver:** selects canonical predicate definitions (`predicate.schema.json`).  
- **Predicate Assembly:** populates templates with jurisdiction-specific thresholds, references.  
- **Control Catalogue Linker:** aligns controls with RTGF catalogue (human review, logging).

## 3. Determinism & Provenance
- Policy clauses grouped using deterministic ordering: obligations > prohibitions > controls; within group sorted by clause ID.  
- Each mapping produces a `predicate.id` derived from clause hash: `pred.<jurisdiction>.<domain>.<sequence>`.  
- Mapping table stored in `policy_token_map.json` per build with clause reference, predicate ID, control ID.  
- Provenance includes citations (legal article, paragraph) attached to predicate metadata.

## 4. Security & Trust
- Mapping rules versioned and reviewed by policy engineers; changes require approval.  
- Predicate library signatures ensure templates are untampered.  
- Controls catalogue enforced; unknown control IDs rejected.  
- Mapping engine runs within compiler trust boundary (ADR-RTGF-003).

## 5. Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| RTGF_MAP_UNSUPPORTED_CLAUSE | Clause type unsupported | raise manual review |
| RTGF_MAP_PREDICATE_TEMPLATE_MISSING | Template not found | fail build |
| RTGF_MAP_CONTROL_UNKNOWN | Control ID missing in catalogue | fail build |
| RTGF_MAP_RULE_CONFLICT | Conflicting clause precedence | flag and halt |

## 6. Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|------|
| Mapping duration | ≤ 2 min per snapshot | determinism check included |
| Manual review rate | ≤ 5 % clauses | indicates template coverage |
| Control alignment | 100 % | every control maps to catalogue entry |

## 7. Interfaces & Integration
| Dependency | Direction | Purpose |
|------------|-----------|---------|
| Policy Source Matrix | inbound | Provide clauses & metadata |
| Predicate template library | inbound | Canonical predicate definitions |
| Control catalogue | inbound | Standard control identifiers/semantics |
| Compiler pipeline | outbound | Predicate set, controls for token assembly |
| QA tooling | outbound | Review unresolved clauses |

## 8. Metrics & Observability
- Prometheus: `rtgf_policy_clause_mapped_total{status}`, `rtgf_policy_manual_review_total`, `rtgf_policy_mapping_duration_seconds`.  
- Mapping logs include clause ID, predicate ID, decision path, provenance references.  
- Dashboard showing mapping coverage per jurisdiction/domain.

## 9. Acceptance Tests
| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| RTGF-CT-70 | Standard obligation clause | maps to predicate template with deterministic ID |
| RTGF-CT-71 | Unsupported clause | build halts with `RTGF_MAP_UNSUPPORTED_CLAUSE` |
| RTGF-CT-72 | Control alignment test | unknown control triggers `RTGF_MAP_CONTROL_UNKNOWN` |
| RTGF-CT-73 | Clause order invariance | mapping output identical regardless of input order |
| RTGF-CT-74 | Provenance capture | predicate metadata includes legal citation |

## 10. Acceptance Criteria
1️⃣ All supported policy clauses deterministically mapped to predicate/control definitions; CT-70..74 green.  
2️⃣ Mapping provenance recorded and reviewable; manual review path for unsupported clauses.  
3️⃣ Predicate and control IDs stable across builds; mapping manifest stored with artefacts.

## Consequences
- ✅ Consistent mapping ensures compiled tokens faithfully represent legal obligations.  
- ✅ Provenance data simplifies audits and regulator feedback loops.  
- ⚠️ Maintaining template coverage requires ongoing policy expertise.  
- ⚠️ Manual review backlog may slow onboarding of new jurisdictions without automation.
