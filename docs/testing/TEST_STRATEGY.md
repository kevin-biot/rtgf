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
- ≥90% line coverage for Go packages (`rtgf-registry`, `rtgf-verify-lib`) with per-package gates.
- ≥80% statement coverage for TypeScript compiler and evaluator packages using an instrumented runner.
- Contract coverage: 100% paths enumerated in OpenAPI examples exercised in integration tests.
- Determinism checks: byte-for-byte assertion for every generated token artefact.

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
- **Artifacts:** JSON diffs, log traces, generated tokens stored under `out/` with cleanup.

### 4.3 End-to-End Workflows
- **Goals:** Cover regulator snapshot ingestion to registry publication.
- **Scenarios:**
  1. Compile sample policy (`examples/policy.snapshot.json`) → produce predicate set and plan → assemble IMT.
  2. Publish outputs to local registry FS → serve through HTTP → verify via `verify` endpoint.
  3. Simulate revocation window bump and confirm caching behaviour.
- **Execution:** Containerised smoke test (Docker Compose) or ephemeral process orchestrated in CI.

### 4.4 Contract and Schema Tests
- Validate registry routes against `docs/openapi/registry-openapi.yaml` using `schemathesis` (HTTP) and `spectral` (lint).
- Enforce JSON Schema validation for `predicate.schema.json`, `eval-plan.schema.json`, and generated tokens.
- Add contract tests ensuring JWKS response matches expected key set.

### 4.5 Determinism and Property Tests
- Property-based tests (Go: `testing/quick`, TS: `fast-check`) for ordering, hash stability, and resolver outputs.
- Replay determinism: compile same snapshot twice, compare `sha256` digests.
- Time window fuzzing: randomised `nbf`/`exp` ranges to assert revocation responses.

### 4.6 Non-Functional Tests
- **Performance:** Benchmark evaluator throughput (`go test -bench` placeholder) and Node evaluator microbenchmarks once implemented.
- **Security:** Static analysis (`gosec`, `npm audit`) and signature validation mocks.
- **Resilience:** Fault injection for resolver timeouts, malformed inputs, partial file reads.

## 5. Test Data Management
- Source snapshots live in `rtgf-snapshots` and `examples/`; store canonical fixtures under `registry/static/tokens`.
- Maintain deterministic fixtures with hashed filenames; update `DefaultTokens` table when versions change.
- Use builders/factories in tests to avoid fixture drift.
- Adopt golden files for evaluator traces and registry catalog payloads; pin them via `testdata/`.

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
- Main branch requires deterministic run (no uncommitted generated artefacts).
- Release candidate tag triggers extended smoke suite (end-to-end workflow, schema validation).

## 8. Roles and Responsibilities
- **Module owners:** maintain respective unit tests, review coverage regressions.
- **QA/Release owner:** enforces gates, curates integration suite.
- **Infra owner:** maintains CI pipelines, fixtures, secrets (for JWKS test keys).

## 9. Risk Register and Mitigations
- **Placeholder implementations** (compiler/evaluator) risk under-tested behaviour. Mitigation: implement contract tests early; track TODOs.
- **Fixture drift** between docs and code can break determinism. Mitigation: automated schema validation, golden tests.
- **Cross-language divergence** (Go vs TS). Mitigation: shared JSON fixtures, contract tests, alignment meetings.
- **Time-critical revocation logic** may rely on system clock. Mitigation: injectable clock, deterministic env vars (`FIXED_TIME`).

## 10. Reporting
- Nightly job publishes trend graphs (coverage %, failed tests, flake rate).
- Failures triaged with root cause template (component, scenario, fix, regression tests).
- Maintain changelog in `docs/testing/QA_CHANGELOG.md` (future) recording major test additions.

## 11. Continuous Improvement
- Quarterly review of test suite effectiveness (flake audit, duration metrics).
- Explore mutation testing once baseline coverage is stable (Go: `mutagen`; TS: `stryker`).
- Align with broader Lane² compliance test harness for shared corridors.

This strategy guides current efforts and will evolve alongside new corridors, token formats, and runtime features.
