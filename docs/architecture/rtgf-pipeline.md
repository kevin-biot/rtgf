# RTGF Pipeline Overview

This document explains the deterministic flow from policy snapshots to distributed RMT/IMT tokens, building on the Policy Source Matrix and RTGF Internet-Draft.

## 1. Pipeline Stages

1. **Snapshot Ingestion**
   - Fetch signed policy snapshots from `/policy-matrix/index.jsonld` sources.
   - Validate signatures (Ed25519 JWS), canonicalise via RFC 8785, verify context digests.

2. **Compilation**
   - Generate RMTs for each jurisdiction/domain.
   - Fetch external data hashes (sanctions, AML rulebooks, medical device registries) and embed in `evidence_sources`.
   - Produce IMTs for ordered corridor pairs using the intersection algorithm (minimum TTL, union of prohibitions, strictest controls).

3. **Transparency Logging**
   - Append issuance events to the transparency log with Merkle inclusion proof.
   - Publish status list updates when revocations occur.

4. **Registry Publication**
   - Host `/.well-known/rtgf`, `/registry/rmt`, `/registry/imt`, `/revocations`, `/transparency` over HTTPS (TLS 1.3, ALPN `aarp/1`).
   - Emit Problem Details errors for any fetch failures.

5. **Verification & Enforcement**
   - PDPs/PEPs retrieve tokens, validate signatures, check status lists, enforce controls.
   - Evidence bundles reference IMT/RMT IDs and hash values for audit.

## 2. EU → Singapore Corridor Example

1. Snapshots: `policy_EU_payments_aml.jsonld`, `policy_SG_payments_aml.jsonld`.
2. RMTs: `RMT-EU-payments_aml-2025-10-01`, `RMT-SG-payments_aml-2025-10-01`.
3. IMT: `IMT-EU.SG-payments_aml-2025-10-01` with:
   - `retention` = max(5 years EU, 7 years SG);
   - `sanctions_dataset_hash` union of EU & SG hashes;
   - `assurance_level` = higher of both;
   - `data_residency` = stricter of EU adequacy vs SG PDPA obligations;
   - `effective_conflicts` listing unresolved policy differences.
4. Transparency entry provides Merkle root, status list index, prior hash.
5. PEP enforces corridor controls before executing cross-border transactions.

## 3. Scheduler & Automation

- Nightly or hourly jobs pull updated snapshots and regenerate tokens.
- CI/CD should include signature verification, idempotence tests (same input → same output), and coverage of revocation handling.

## 4. Future Extensions

- Support for hybrid PQ signatures in pipeline once CFRG standards settle.
- Hooks for zero-knowledge or MPC proofs where evidence requirements demand Mandala-style attestations.

See the ADR series in `docs/adr/` for design decisions behind snapshot formats, sanctions hashing, multi-signature issuance, and transparency log structure.
