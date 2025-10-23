# RTGF Reference Registry Server

This service provides the authoritative RTGF registry API described in `draft-lane2-rtgf-00` and the architecture notes. It exposes:

- Public read endpoints for RMT/IMT retrieval, revocation status lists, and transparency proofs.
- An administrative submission API (mTLS) for RTGF compiler bundles (token issuance, revocation).
- Transparency log maintenance and JWKS hosting.

Directory layout:
```
cmd/registryd/        # main entrypoint
internal/api/         # public read handlers (/rmt, /imt)
internal/admin/       # submission + revoke flows
internal/transparency/ # Merkle log + checkpoints
internal/revocation/  # status list generation and storage
internal/crypto/      # signing, JWKS, key rotation helpers
internal/storage/     # persistence (e.g., Badger/Bolt/SQL)
config/               # sample YAML/TOML configs (TLS, mTLS, policy)
```

The service targets Go 1.22+ with chi or net/http. OpenAPI specs live under `docs/openapi/` at the repository root.
