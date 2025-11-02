# RTGF Coverage Baseline

Baseline captured for commit `c34c02cdca78ac338c9ed8cbc0318b9cd2799a06` on 2025-11-02T09:45:10Z.

## Go Modules

### `rtgf-verify-lib`

| File | Coverage |
|------|----------|
| `static.go:NewStaticVerifier` | 81.0% |
| `static.go:VerifyRRMT` | 100.0% |
| `static.go:VerifyCORT` | 100.0% |
| `static.go:VerifyPSRT` | 100.0% |
| `static.go:Token` | 75.0% |
| `static.go:Metadata` | 100.0% |
| `static.go:detectType` | 0.0% |
| `static.go:verifyType` | 72.7% |

Module statement coverage: **68.8%**

### `rtgf-registry`

| File | Coverage |
|------|----------|
| `cmd/registryd/main.go:main` | 0.0% |
| `internal/api/api.go:NewServer` | 81.8% |
| `internal/api/api.go:ServeHTTP` | 100.0% |
| `internal/api/api.go:routes` | 100.0% |
| `internal/api/api.go:handleHealth` | 60.0% |
| `internal/api/api.go:handleTokenByURI` | 66.7% |
| `internal/api/api.go:handleTokenByType` | 47.6% |
| `internal/api/api.go:handleCatalog` | 70.0% |
| `internal/api/api.go:handleJWKS` | 0.0% |
| `internal/api/api.go:serveStaticJSON` | 66.7% |
| `internal/api/api.go:lookupBySlug` | 75.0% |
| `internal/verify/handler.go:NewService` | 100.0% |
| `internal/verify/handler.go:HandleVerify` | 63.6% |
| `internal/verify/handler.go:HandleRevocationsGet` | 0.0% |
| `internal/verify/handler.go:HandleRevocationsBump` | 60.0% |
| `internal/verify/handler.go:validateTokens` | 78.6% |
| `internal/verify/handler.go:respondJSON` | 100.0% |
| `internal/verify/handler.go:validateWindows` | 72.0% |
| `internal/verify/handler.go:currentTime` | 50.0% |

Module statement coverage: **59.2%**

### Untested Packages

`internal/admin`, `internal/crypto`, `internal/revocation`, `internal/storage`, and `internal/transparency` currently have no automated tests.

## TypeScript Packages

TypeScript packages (`rtgf-compiler/ppe-compiler`, `aarp-core/ppe-evaluator`) are not yet instrumented; coverage tooling will be added in Phase 0.

## Determinism Manifest

Determinism report not yet generated. Placeholder to be populated once the deterministic harness is implemented.
