# SAPP Alignment – RTGF Touchpoints

Lane² SAPP now consumes an expanded suite of corridor tokens (RRMT, CORT, PSRT) and expects deterministic verification plus DOP evidence hand-off. This note captures what exists in the SAPP repo today and the RTGF work required to stay in lockstep.

## Token Surface Area
- `RRMT` (Revenue Rule Matrix) – pricing tiers, surcharges, VAT. Struct + JSON Schema live in `sapp/pkg/tokens/rrmt.go` and `sapp/schemas/rrmt.schema.json`.
- `CORT` (Commercial Operating & Revenue Terms) – parties, revenue splits, FX guardrails. See `sapp/pkg/tokens/cort.go` and `sapp/schemas/cort.schema.json`.
- `PSRT` (Payment Settlement Rule Token) – capture cadence, dispute reserves, KYB attestation. Defined in `sapp/pkg/tokens/psrt.go` and `sapp/schemas/psrt.schema.json`.
- `TokenSet` – shared helper that bundles `RMT/IMT` plus the new payment tokens for Merkle receipts (`sapp/pkg/tokens/token_set.go`).

## What RTGF Needs To Mirror
- **Schemas** – copy or reference the RRMT/CORT/PSRT JSON Schemas into `rtgf/schemas/` so registry + compiler can validate the artefacts they emit.
- **Token Builders** – extend the compiler (or a new payment-token module) to deterministically build RRMT, CORT, PSRT payloads from policy + corridor inputs, mirroring the structs above.
- **Registry Endpoints** – expose read/verify endpoints for the new tokens (`GET /rrmt/{id}`, `POST /tokens/verify`, etc.) per `sapp/docs/architecture/rtgf-endpoints.md`.
- **Verification Library** – make sure `rtgf-verify-lib` implements `VerifyRRMT/CORT/PSRT` to satisfy `pkg/tokens/verify.go` in SAPP.
- **Static Fixtures** – publish dummy RRMT/CORT/PSRT objects under `registry/static/` so SAPP’s sandbox runs can retrieve and validate without hitting live compilers.

## DOP Test Client Expectations
- SAPP settlement tests now mint Merkle roots that include all token URIs (card receipt hash + RRMT/CORT/PSRT). A lightweight RTGF “DOP client” should:
  - produce deterministic token URIs for fixtures (`urn:lane2:token:RRMT:...`);
  - emit a dummy evidence bundle ID that the DOP pipeline can accept;
  - return a canonical JWS manifest (see `sapp/pkg/signing/manifest.go`) so SAPP can exercise Ed25519 signing/verification.
- Long term the DOP client should push the generated bundle into the transparency log (`rtgf-registry/internal/transparency`) once implemented.

## Immediate Next Steps
1. Land the RRMT/CORT/PSRT schemas + example payloads in RTGF and wire them into CI validation. ✅
2. Update `rtgf-verify-lib` API surface (and generated SDKs) so SAPP can stop using mocks for token verification. ✅ Static verifier now exposes token metadata for TTL checks.
3. Extend `rtgf-registry` catalog to surface RMT/IMT transparency metadata alongside payment tokens (hashes, nbf/exp, revoked) so SAPP can run corridor audits. ⏳

Track progress here, and update once RTGF emits the new token classes end-to-end.
