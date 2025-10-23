# ADR-001: Policy Source Matrix Format

## Status
Accepted

## Context
RTGF requires deterministic, versioned policy inputs per jurisdiction and domain. Regulators publish heterogeneous legal texts, sanctions feeds, and guidance; the compiler needs a uniform schema to ingest them.

## Decision
Create a Policy Source Matrix comprised of signed JSON-LD snapshots (`policy:Snapshot`) keyed by jurisdiction and domain. Each snapshot MUST include:
- Jurisdiction code (ISO 3166-1 alpha-2 or regional)
- Domain identifier aligned with the RTGF domain registry
- Effective/expiry timestamps (RFC 3339 UTC)
- Controls, duties, prohibitions, assurance levels
- Evidence source hashes for external datasets (e.g., sanctions)
- Normative references (legal instruments, guidance)
- Detached Ed25519 signature by the issuing authority DID/JWKS

Snapshots are listed in `/policy-matrix/index.jsonld` with metadata for automated retrieval.

## Consequences
- Enables deterministic compiler outputs (RMT/IMT) once signatures and hashes are validated.
- Provides clear provenance chain from legal source to enforcement token.
- Requires regulators to maintain signing keys and update snapshots when laws or datasets change.

## Links
- `draft-lane2-rtgf-00` Section 3.1
- `docs/architecture/policy-source-matrix.md`
