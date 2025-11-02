# RTGF Test Strategy

## 1. Purpose
- Define a shared approach for validating the Reference Token Generation Framework (RTGF) across compiler, evaluator, registry, and verifier components.
- Align contributors on quality goals, tooling, environments, and release gates.
- Provide the baseline for the coverage uplift plan and future QA automation.

## 2. Product and Scope

| Component | Language | Repository Path | Primary Responsibilities | Current Test Notes |
|-----------|----------|-----------------|---------------------------|--------------------|
| PPE Compiler | TypeScript (Node) | `rtgf-compiler/ppe-compiler` | Transform policy snapshots into predicate sets and evaluation plans. | CLI is stubbed, tests live in markdown specs only. |
| PPE Evaluator | TypeScript (Node) | `aarp-core/ppe-evaluator/src` | Execute compiled plans deterministically at runtime. | Only doc specs; no automated assertions. |
| Registry Service | Go | `rtgf-registry` | Serve tokens, catalog, JWKS, and verification endpoints. | Foundational handler tests exist; no integration coverage. |
| Verification Library | Go | `rtgf-verify-lib` | Load fixtures, expose token metadata, and type gate checkers. | Basic happy-path and error-path unit tests. |
| CLI + Tooling | TBD (Go/TS) | `rtgf-cli`, `tests/*.sh` | Operational workflows (compile, publish, verify). | Shell scripts only; no portable harness. |
| Schemas & Snapshots | JSON/JSON Schema | `shared/ppe-schemas`, `rtgf-snapshots` | Canonical definitions and inputs. | Schema validation not automated. |

**In scope**
- All runtime and build-time code paths that influence issued tokens, registry APIs, verification, and snapshot compilation.
- Deterministic behaviour (hashing, ordering, serialization) and cryptographic boundaries exposed by the reference implementation.
- Tooling used to run workflows (`rtgf` CLI, round-trip scripts).

**Out of scope (until future milestone)**
- Proprietary or external services (Lane² SaaS) beyond documented mocks.
- Mobile SDKs or language ports not present in this repository.
- Performance benchmarking beyond smoke checks.

## 3. Quality Objectives and Metrics
- Deterministic outputs: same inputs must yield byte-identical predicate sets, evaluation plans, and registry artefacts.
- Backwards compatibility: existing token fixtures must remain valid unless versioned.
- API reliability: HTTP surface obeys OpenAPI contract (`docs/openapi/registry-openapi.yaml`).
- Security posture: signed artefacts are validated, revocation windows enforced, JWKS served consistently.
- Observability: tests should fail fast with actionable diagnostics (structured diffs, traces).

**Targets**
- Go packages: ≥90% line coverage per package with red-lines of ≥95% for `rtgf-registry/internal/verify` and `rtgf-registry/internal/jwks`; no file may drop below 80% line coverage without an explicit waiver.
- TypeScript packages (`rtgf-compiler/ppe-compiler`, `aarp-core/ppe-evaluator`): ≥80% statement and ≥70% branch coverage per package, ≥75% statements per file; CI fails if coverage delta falls more than 1.0% below the recorded baseline.
- Contract coverage: every OpenAPI path exercised with at least one 2xx and one 4xx/5xx case; schemathesis fuzzing covers ≥100 examples per route.
- Determinism checks: byte-for-byte assertion for every generated token artefact with stable digests recorded alongside golden fixtures.
- Reliability guardrails: total CI runtime ≤10 minutes across parallel jobs; weekly flake rate ≤0.5% (captured via flaky-test tracker).

## 4. Test Levels and Approach

### 4.1 Unit Tests
- **Goals:** Validate pure functions, struct methods, and predicate helpers without external IO.
- **Tooling:** `go test`, `testing/fstest`, `httptest` for Go; `vitest` (Node 18 runtime), `ts-node` for TypeScript.
- **Focus Areas:**
  - `rtgf-verify-lib`: error surfaces (revoked tokens, invalid JSON, type detection).
  - `rtgf-registry/internal/*`: handler behaviour, catalog composition, revocation state machine.
  - `ppe-compiler`: snapshot parsing, plan sequencing, hash calculation.
  - `ppe-evaluator`: predicate evaluation, resolver orchestration, trace generation.

### 4.2 Component and Integration Tests
- **Goals:** Exercise module boundaries with real fixtures and file IO.
- **Approach:**
- Spin up in-memory registry server via `httptest.Server`.
- Run CLI commands (`rtgf-ppe.ts`, future `rtgf` binary) against sample snapshots.
- Validate that outputs match schemas in `shared/ppe-schemas` via `ajv`.
- Reuse `tests/ppe-roundtrip/run.sh` as an automated integration target executed through `npm test` or `go test` wrappers.
- Exercise JWKS lifecycle: rotate keys mid-run, ensure cached JWKS invalidation, confirm fail-closed responses (`401 kid_unknown`) and acceptance of tokens signed by both old and new keys until expiry.
- **Artifacts:** JSON diffs, log traces, generated tokens stored under `out/` with cleanup.

### 4.3 End-to-End Workflows
- **Goals:** Cover regulator snapshot ingestion to registry publication.
- **Scenarios:**
  1. Compile sample policy (`examples/policy.snapshot.json`) → produce predicate set and plan → assemble IMT.
  2. Publish outputs to local registry FS → serve through HTTP → verify via `verify` endpoint.
  3. Simulate revocation window bump and confirm caching behaviour.
- **Execution:** Containerised smoke test (Docker Compose) or ephemeral process orchestrated in CI.

### 4.4 Contract and Schema Tests
- Validate registry routes against `docs/openapi/registry-openapi.yaml` using `schemathesis` (HTTP) with ≥100 generated examples per operation and `spectral` (lint).
- Ensure every path/operationId has both positive (2xx) and negative (4xx/5xx) coverage and that request/response payloads match documented examples.
- Enforce JSON Schema validation for `predicate.schema.json`, `eval-plan.schema.json`, generated tokens, and ensure metaschema version drift is detected.
- Add contract tests ensuring JWKS response matches expected key set and that cache headers/ETags align with rotation scenarios.

### 4.5 Determinism and Property Tests
- Property-based tests (Go: `testing/quick`, TS: `fast-check`) for ordering, hash stability, and resolver outputs.
- Replay determinism: compile same snapshot twice, compare `sha256` digests.
- Time window fuzzing: randomised `nbf`/`exp` ranges to assert revocation responses.
- Determinism controls: all suites run with `TZ=UTC`, `LANG=C`, injected fixed clock via `FIXED_TIME`, and seeded RNG (`RTGF_SEED=1337`); JSON serialization uses stable key ordering and outputs are written as content-addressed blobs (`out/<sha256>.json`) with a human-readable index. Deterministic harness lives at `tests/ppe-roundtrip/harness.ts`.

### 4.6 Non-Functional Tests
- **Performance:** Benchmark evaluator throughput (`go test -bench` placeholder) and Node evaluator microbenchmarks once implemented.
- **Security:** Static analysis (`gosec`, `npm audit`) and signature validation mocks.
- **Resilience:** Fault injection for resolver timeouts, malformed inputs, partial file reads.

## 5. Test Data Management
- Source snapshots live in `rtgf-snapshots` and `examples/`; store canonical fixtures under `registry/static/tokens`.
- Maintain deterministic fixtures with hashed filenames; update `DefaultTokens` table when versions change.
- Use builders/factories in tests to avoid fixture drift.
- Adopt golden files for evaluator traces and registry catalog payloads; pin them via `testdata/`.
- Treat goldens as single source of truth shared across Go and TypeScript suites—hashes, payloads, and error codes must match exactly; update via `make golden-update` with reviewer approval.

## 6. Tooling and Infrastructure
- **Languages:** Go 1.22+, Node 18 LTS.
- **Package Managers:** `go` modules; `pnpm` or `npm` for TypeScript packages (introduce `package.json` per module).
- **CI:** GitHub Actions or CircleCI with jobs:
  - `lint` (gofmt, golangci-lint, eslint/tsc).
  - `test-unit` (Go, TS).
  - `test-integration` (round-trip workflows).
  - `coverage` (merge Go and TS reports via `gocovmerge`, `lcov`).
- **Artifacts:** Upload coverage reports, JSON diffs, and golden outputs.

## 7. Release Gates
- Pull requests must run unit + integration tests and meet coverage thresholds.
- Main branch requires deterministic run (no uncommitted generated artefacts); CI enforces `git status --porcelain` clean check post-test.
- Release candidate tag triggers extended smoke suite (end-to-end workflow, schema validation).

## 8. Roles and Responsibilities
- **Module owners:** maintain respective unit tests, review coverage regressions.
- **QA/Release owner:** enforces gates, curates integration suite.
- **Infra owner:** maintains CI pipelines, fixtures, secrets (for JWKS test keys).

## 9. Risk Register and Mitigations
- **Placeholder implementations** (compiler/evaluator) risk under-tested behaviour. Mitigation: implement contract tests early; track TODOs.
- **Fixture drift** between docs and code can break determinism. Mitigation: automated schema validation, golden tests.
- **Cross-language divergence** (Go vs TS). Mitigation: enforce single golden fixture set consumed by both toolchains, compare digests in CI, and align error taxonomies.
- **Time-critical revocation logic** may rely on system clock. Mitigation: injectable clock, deterministic env vars (`FIXED_TIME`).

## 10. Reporting
- Nightly job publishes trend graphs (coverage %, failed tests, flake rate).
- Failures triaged with root cause template (component, scenario, fix, regression tests).
- Maintain changelog in `docs/testing/QA_CHANGELOG.md` (future) recording major test additions.

## 11. Continuous Improvement
- Quarterly review of test suite effectiveness (flake audit, duration metrics).
- Explore mutation testing once baseline coverage is stable (Go: `mutagen`; TS: `stryker`).
- Align with broader Lane² compliance test harness for shared corridors.

## 12. Component Execution Notes
- **PPE Compiler (TS):** cover plan ordering, hash stability under key permutation, and property tests ensuring unordered predicate inputs yield identical digests.
- **PPE Evaluator (TS):** achieve branch coverage across `PERMIT` / `DENY` / `PERMIT_WITH_CONTROLS` decisions; inject resolver timeout faults and assert deterministic traces.
- **Registry (Go):** test JWKS cache TTL and rotation, `/verify` behaviour for revoked/expired/malformed tokens, and ensure catalog responses expose stable ETags and cache headers aligned with OpenAPI examples.
- **Verify Library (Go):** expand table-driven tests for type detection, malformed JSON, revoked metadata, and clock skew handling with deterministic error codes.
- **CLI & Tooling:** replace ad-hoc shell scripts with portable harnesses returning non-zero on golden mismatches and include `--update-golden` flow guarded by review.

## 13. Acceptance Checklist (apply to PR reviews)
- All OpenAPI paths exercised with at least one positive and one negative scenario; schemathesis runs recorded.
- Go coverage meets per-package ≥90% (with `internal/verify` and `internal/jwks` ≥95%) and per-file ≥80%; TypeScript compiler/evaluator meet ≥80% statements and ≥70% branches with ≥75% statements per file.
- Determinism guard: repeat test run leaves working tree clean and output hashes match goldens.
- JWKS rotation and fail-closed (`kid_unknown`) scenarios pass along with revocation window and cache TTL assertions.
- CI wall-clock runtime ≤10 minutes and flake rate tracked ≤0.5% weekly.

This strategy guides current efforts and will evolve alongside new corridors, token formats, and runtime features.
