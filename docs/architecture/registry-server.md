# RTGF Registry Reference Server

This document captures the design of the authoritative RTGF Registry Server that publishes RMT/IMT tokens to aARP routers, PDPs, and PEPs.

## 1. Roles & Keys
- **Issuer (Compiler)**: runs RTGF compiler, signs submission bundles.
- **Registry**: verifies submissions, publishes tokens, maintains revocations and transparency logs.
- **Auditor**: consumes transparency log and JWKS for independent verification.

Keys: issuer signing key (Ed25519 MTI, HSM-backed), registry signing key, regulator DID/JWKS trust anchor.

## 2. Transport & Security
- TLS 1.3 only, ALPN `aarp/1`, HTTP/2 or HTTP/3.
- Submission API uses mTLS; client cert SAN includes issuer DID.
- Public read APIs require no auth but responses are signed and cacheable.
- JOSE: detached JWS, `alg=EdDSA`, `kid` resolvable via registry JWKS.

## 3. Discovery & Metadata
- `GET /.well-known/rtgf` exposes issuer DID, JWKS, base URLs, freshness policy, supported domains.
- Responses use `application/imt-rmt+json`. Errors use `application/problem+json` with stable type URIs under `https://lane2.ai/ietf/imt-rmt/errors#`.

## 4. Public Endpoints
- `GET /rmt/{jurisdiction}/{domain}`: latest RMT.
- `GET /imt/{corridor}/{domain}`: latest IMT.
- `GET /revocations?type=<rmt|imt>&since=...`: signed status list segments and Merkle checkpoints.
- `GET /transparency`, `GET /transparency/proof`: append-only log queries and proofs.

High-level flow:

```
nightly ─► rtgf-compiler ─► POST /admin/submit (mTLS bundle)
           verifies snapshots │
           builds tokens       ▼
registry ─► validates, logs, publishes
routers  ─► GET /rmt /imt /revocations /transparency
             verify signatures, cache, enforce fail-closed
```

## 5. Submission API (mTLS)
- `POST /admin/submit`: accepts compiler bundle with tokens, manifests, issuer claims.
- Registry verifies schema, signatures, timestamps, corridor format, and optional quorum requirements before publication.
- `POST /admin/revoke`: updates status lists and transparency log for listed JTIs.

## 6. Components
- `internal/api`: public read handlers (ETag, Cache-Control, signed envelopes).
- `internal/admin`: submission validation, publication pipeline.
- `internal/transparency`: RFC 6962-style Merkle tree, checkpoints, proofs.
- `internal/revocation`: compact bitset status list, delta serialization.
- `internal/crypto`: key loading, signing proxy, JWKS rotation.
- `internal/storage`: persistence layer (Badger/BoltDB/Postgres depending on deployment).

## 7. Observability & Ops
- Metrics at `/metrics` (Prometheus) covering latency, revocation operations, submission success/failure.
- Structured logs (trace ID, JTI, result codes) without PII.
- Key rotations logged to transparency; configuration changes tracked in git/ADR.
- High availability via hot/warm replicas and immutable transparency store.

## 8. Deployment Quickstart (Draft)

```bash
# generate signing keys
docker run lane2/rtgf-registry keygen --kid r1

# start registry via Helm (example values)
helm install lane2-registry ./helm \
  --set registry.domain=reg.eu.example \
  --set hsm.kmsProvider=aws

# publish JWKS and metadata
curl https://reg.eu.example/.well-known/rtgf | jq '.'

# routers fetch latest IMT
curl -s https://reg.eu.example/imt/EU-SG/ai | jq '.'
```

Refer to `docs/openapi/registry-openapi.yaml` for the evolving OpenAPI description of the endpoints.
