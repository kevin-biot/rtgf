# Corridor Policy — Process & Mapping (Living Document)

## Objective
Establish a reproducible, auditable path from **real-world law & guidance** → **Policy Snapshots** → **RMT/IMT tokens** → **deterministic enforcement**. This doc is the source of truth for how we *ingest, map, validate, and publish* corridor policies.

## Scope (MVP → V1)
- **MVP (this month)**: 4 pilot corridors with minimal-but-real controls (sanctions, AML threshold, reporting window).
- **V1 (next)**: add data-protection (GDPR/transfer), AI-Act transparency predicates, and Mandala-proof hooks.

## Roles & RACI
- **Policy Engineer (PE)**: Extract rules, thresholds, approvals → snapshots. (R/A)
- **Predicate Engineer (PrE)**: Map rules → predicates & eval plans. (R)
- **Registry Ops (RO)**: Publish tokens, revocations, transparency. (R)
- **Reviewer (REV)**: Cross-check citations & numbers. (A)
- **Stakeholders (STK)**: Regulators/partners; provide source pointers. (C/I)

## Pipeline Stages (S0–S7)
- **S0 Discover**: Identify authoritative sources (laws, circulars, APIs, BIS Mandala refs); log in Source Registry.
- **S1 Ingest**: Fetch/Archive (hash, timestamp, provenance JSON).
- **S2 Normalize**: Convert to **Policy Snapshot** JSON-LD (jurisdiction/domain, controls/duties/prohibitions, references).
- **S3 Map Predicates**: Translate snapshots → predicate_set; define thresholds & resolvers; compile eval_plan (deterministic).
- **S4 Compile Tokens**: RTGF builds **RMTs** and **IMTs** (corridor intersections).
- **S5 Verify**: Schema + signature + deterministic roundtrip; dry-run evaluation (PPE) with sample contexts.
- **S6 Publish**: Push to **Registry** `/rmt`, `/imt`, `/revocations`, `/transparency`; set TTL; announce.
- **S7 Monitor**: Track freshness (TTL, ETag), revocations, policy changes; nightly sync & dashboards.

## Artefact Mapping
| Artefact | Producer | Location | Verification |
|----------|----------|----------|--------------|
| Source Registry | PE | `policy-sources/registry/sources.yaml` | provenance hash |
| Policy Snapshot | PE | `rtgf-snapshots/<jur>/...json` | schema + signature |
| Predicate Set | PrE (compiler) | embedded in token | schema + plan-hash |
| Eval Plan | PrE (compiler) | embedded in token | RFC8785 + sha256 |
| RMT/IMT | RTGF | `out/tokens/*.json` | JOSE sig + transparency |
| Evidence Bundle | DOP | chain storage | Merkle proof |

## Acceptance Criteria
- Every token references at least one **normative** source (CELEX, MAS/BNM, OFAC/EU API).
- Repeat builds with same inputs → byte-identical outputs (plan hash stable).
- Dry-run PPE → consistent decisions for **allow**, **deny**, **controls** scenarios.

## Risk & Controls
- **Stale law**: TTL ≤ 24h + nightly sync.
- **Ambiguity**: Flag conflicts in IMT `effective_conflicts`; require Reviewer sign-off.
- **Resolver outage**: Fail-closed; reason code `RESOLVER_ERROR:*`.

This document is living; major updates recorded in `docs/corridors/workflows/CHANGELOG.md`.
