# Corridor Playbook — SG-MY

## 0. Executive Summary
- Jurisdictions: SG → MY
- Domain(s): payments_aml
- Current Status: planning
- Decision Summary: pre-validation = required; sanctions screening = mandatory

## 1. Sources (S0/S1)
| id | type | url | hash | citation |
|---|---|---|---|---|
|  | law |  |  |  |
|  | circular |  |  |  |
|  | dataset |  |  |  |

## 2. Policy Snapshot(s) (S2)
- File(s): `rtgf-snapshots/sg/...json`, `rtgf-snapshots/my/...json`
- Controls: …
- Duties: …
- Prohibitions: …
- Normative References: …

## 3. Predicate Mapping (S3)
| predicate_id | applies_when | inputs | logic | on_fail |
|---|---|---|---|---|
| pred.sanctions.public.v1 | SG/cross_border | counterparty_country, sanctions_dataset_hash | EQUALS(sanctions_result,true) | DENY:SANCTIONS_HIT |
| pred.aml.threshold.v1 | MY only | amount.value | LTE(amount, THRESHOLD) | PERMIT_WITH_CONTROLS:AMOUNT_REVIEW |
| pred.transfer.mechanism.v1 | SG specific | transfer_mechanism | IN(SCC,DPF,ADEQUACY) | DENY:TRANSFER_MECHANISM_UNSUPPORTED |

## 4. IMT Intersection (S4)
- IMT id: IMT-SG.MY-payments_aml-2025-10-23
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
  - `/imt/SG-MY/payments_aml`
  - `/rmt/SG/payments_aml`, `/rmt/MY/payments_aml`
- TTL: 86400s
- Revocation status list: `/revocations`

## 7. Monitoring (S7)
- Freshness SLO: ETag delta ≤ 24h
- Alerts: resolver errors, transparency divergence
- Next review: 2025-11-23
