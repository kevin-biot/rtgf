# Corridor Playbook — {{CORRIDOR_ID}}

## 0. Executive Summary
- Jurisdictions: {{SRC}} → {{DST}}
- Domain(s): {{DOMAIN_LIST}}
- Current Status: {{STATUS}}
- Decision Summary: pre-validation = {{PREVALIDATION}}; sanctions screening = {{SANCTIONS}}

## 1. Sources (S0/S1)
| id | type | url | hash | citation |
|---|---|---|---|---|
|  | law |  |  |  |
|  | circular |  |  |  |
|  | dataset |  |  |  |

## 2. Policy Snapshot(s) (S2)
- File(s): `rtgf-snapshots/{{SRC_LOWER}}/...json`, `rtgf-snapshots/{{DST_LOWER}}/...json`
- Controls: …
- Duties: …
- Prohibitions: …
- Normative References: …

## 3. Predicate Mapping (S3)
| predicate_id | applies_when | inputs | logic | on_fail |
|---|---|---|---|---|
| pred.sanctions.public.v1 | {{SRC}}/cross_border | counterparty_country, sanctions_dataset_hash | EQUALS(sanctions_result,true) | DENY:SANCTIONS_HIT |
| pred.aml.threshold.v1 | {{DST}} only | amount.value | LTE(amount, THRESHOLD) | PERMIT_WITH_CONTROLS:AMOUNT_REVIEW |
| pred.transfer.mechanism.v1 | {{SRC}} specific | transfer_mechanism | IN(SCC,DPF,ADEQUACY) | DENY:TRANSFER_MECHANISM_UNSUPPORTED |

## 4. IMT Intersection (S4)
- IMT id: IMT-{{SRC}}.{{DST}}-{{DOMAIN}}-{{DATE}}
- Effective Requirements:
  - approval_threshold: …
  - sanctions_screening: mandatory
  - pre_validation: required
- Interaction Matrix:
  - loan_drawdown: deny_if_threshold_exceeded_without_approval
  - cross_border_lending: permit_if_controls_met

## 5. Verification (S5)
- Schema: ✓
- Signature (JWS Ed25519): ✓
- Plan hash (RFC8785 + sha256): `sha256:…`
- Dry-run PPE results (context fixtures):
  - allow.json → PERMIT
  - deny_sanctions.json → DENY (SANCTIONS_HIT)
  - review_threshold.json → PERMIT_WITH_CONTROLS (AMOUNT_REVIEW)

## 6. Publication (S6)
- Registry URLs:
  - `/imt/{{CORRIDOR_ID}}/{{DOMAIN}}`
  - `/rmt/{{SRC}}/{{DOMAIN}}`, `/rmt/{{DST}}/{{DOMAIN}}`
- TTL: 86400s
- Revocation status list: `/revocations`

## 7. Monitoring (S7)
- Freshness SLO: ETag delta ≤ 24h
- Alerts: resolver errors, transparency divergence
- Next review: {{NEXT_REVIEW_DATE}}
