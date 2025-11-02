# RTGF Test Coverage Uplift Plan

This plan operationalises the RTGF Test Strategy and lays out the work required to reach the target coverage thresholds during the next three iterations (~6 weeks).

## Timeline Overview

| Phase | Duration | Focus |
|-------|----------|-------|
| Phase 0 | Week 0 | Baseline metrics, infra setup. |
| Phase 1 | Weeks 1-2 | Unit test expansion in Go and TypeScript packages. |
| Phase 2 | Weeks 3-4 | Integration and determinism suites, schema validation. |
| Phase 3 | Weeks 5-6 | End-to-end workflows, resiliency, coverage gating. |

## Phase 0 – Baseline and Foundations
1. **Capture Current Coverage**
   - Run `go test ./... -coverprofile=coverage-go.out` within `rtgf-registry` and `rtgf-verify-lib`.
   - Instrument TypeScript modules using `vitest --coverage` (placeholder until package.json added).
   - Store baseline numbers (per module + per file) in `docs/testing/coverage-baseline.md` alongside the git commit hash.
2. **Establish Tooling**
   - Add `package.json` + `tsconfig.json` for `rtgf-compiler/ppe-compiler` and `aarp-core/ppe-evaluator`.
   - Introduce `Makefile` targets (`make test`, `make cover`) to orchestrate Go + TS runs.
   - Configure CI workflow skeleton (lint, unit, integration, coverage upload).
   - Pin toolchains via `.tool-versions`/`.nvmrc` and document supported Go/Node versions.
3. **Governance**
   - Enable coverage badge tracking (Codecov or self-hosted) – optional but recommended.
   - Define pull request template section for test evidence.
   - Publish initial determinism report (sha256s for existing fixtures).

## Phase 1 – Unit Test Expansion
1. **Go: `rtgf-verify-lib`**
   - Add negative tests covering malformed JSON, revoked tokens, nil verifier handling.
   - Introduce table-driven tests for `detectType` and metadata fallbacks.
   - Target ≥90% coverage for package.
2. **Go: `rtgf-registry/internal`**
   - Expand handler tests to cover method-not-allowed, bad slugs, JWKS caching, catalog environment variable `RTGF_URL`.
   - Add table-driven tests for `validateWindows` error branches (invalid nbf/exp formats, revoked flags, clock skew injection).
   - Cover `HandleRevocationsGet`, concurrency on bump, and deterministic error codes.
3. **TypeScript: PPE Compiler**
   - Port markdown scenarios in `rtgf-compiler/ppe-compiler/tests/roundtrip.spec.md` to executable tests (Vitest).
   - Add unit tests for CLI arg parsing, default order, plan sequencing, and schema validation via `ajv`.
   - Cover hash stability across object key permutations; introduce property tests (`fast-check`) for unordered predicates.
4. **TypeScript: PPE Evaluator**
   - Convert `aarp-core/ppe-evaluator/tests/evaluator.spec.md` into Vitest cases.
   - Mock resolvers and confirm decision matrix (PERMIT, DENY, PERMIT_WITH_CONTROLS) with branch coverage thresholds.
   - Snapshot trace output to golden file and assert deterministic digest values.

## Phase 2 – Integration and Determinism
1. **Round-Trip Integration**
   - Use the deterministic harness in `tests/ppe-roundtrip/harness.ts` (run via `npm run run`) with fixed env + seeded RNG; execute twice and fail on digest diffs.
   - Assert JSON outputs against schemas via `ajv` (predicate, eval plan, token envelope) and record manifest to `out/ppe/digests.json`.
2. **Registry <> Verifier Contract**
   - Spin up in-memory registry server (`httptest`) and call `/verify` with real fixtures.
   - Validate error codes for missing tokens, expired tokens, revoked tokens, malformed JWKS, and unknown `kid`.
3. **Schema Validation Pipeline**
   - Add automated validation for `shared/ppe-schemas/*.json` against metaschema.
   - Validate generated tokens against predicate and eval plan schemas.
4. **Coverage Tracking**
   - Merge Go coverage reports using `gocovmerge`, convert to lcov.
   - Ensure Vitest coverage outputs lcov (`coverage/ts/lcov.info`).
   - Publish combined coverage dashboard artifact.

## Phase 3 – End-to-End and Resilience
1. **Full Workflow Smoke**
   - Script end-to-end test: compile snapshot → publish to temp dir → start registry server (with JWKS rotation mid-run) → run `/verify` against static verifier.
   - Gate release branch with this workflow (runs nightly + on PRs touching critical paths).
2. **Fault Injection**
   - Add evaluator tests simulating resolver timeouts, malformed predicate definitions, and resolver retries.
   - Exercise registry with missing JWKS file to confirm failure modes and cache TTL handling.
3. **Performance Baseline**
   - Introduce lightweight benchmarks (`go test -bench . ./rtgf-registry/internal/verify`) to monitor regression readiness.
4. **Coverage Gates**
   - Enforce minimum coverage in CI (`go test -coverpkg ./...`, `vitest --coverage --run`).
   - Fail PR if coverage drops >1% from baseline unless override label applied.
5. **Determinism Guard**
   - Add CI job to run full pipeline twice and assert clean git tree (`git status --porcelain`) plus matching digest manifest.

## Deliverables and Acceptance Criteria
- Coverage dashboards updated weekly with component-level metrics.
- All markdown spec scenarios replaced by executable tests.
- Determinism tests demonstrate byte-identical outputs for canonical fixtures.
- CI pipelines running unit, integration, and end-to-end suites with coverage gates enabled.
- Documentation updated (`docs/testing/TEST_STRATEGY.md`, `docs/testing/coverage-baseline.md`, QA changelog entries).
- Determinism manifest (`artifacts/digests.json`) published per run and compared against goldens.
- JWKS rotation and revocation cache tests automated with clearly defined pass/fail signals.

## CI and Reporting Enhancements
- Split workflows into `lint`, `unit-go`, `unit-ts`, `integration`, `e2e`, and `coverage-merge` jobs with ≤10 minute combined runtime.
- Upload artifacts: `coverage-go.out`, `coverage/ts/lcov.info`, `artifacts/digests.json`, and golden diffs (if any).
- Emit flaky-test summary (flakes ≤0.5% weekly) and alert on coverage or determinism regressions.
- Nightly job updates coverage trend chart and determinism report badge.

## Ownership and Coordination
- **Quality Lead:** orchestrates phases, reports progress, owns CI pipelines.
- **Go Owners:** `rtgf-registry`, `rtgf-verify-lib` maintainers.
- **TypeScript Owners:** PPE compiler and evaluator maintainers.
- **Documentation:** QA team maintains testing docs and coverage history.
- Weekly sync to review blockers, rotate code reviews focused on test debt.

## Risks and Mitigations
- **Tooling Drift:** Ensure consistent Node version via `.nvmrc` or `.tool-versions`.
- **Time Constraints:** Prioritise high-risk modules first; deprioritise low-impact performance work if needed.
- **Flaky Integration Tests:** Use deterministic fixtures, avoid real network IO, mock time via `FIXED_TIME`.
- **CI Duration:** Parallelise Go and TS jobs, cache dependencies.

## Acceptance Checklist
- OpenAPI operations each have at least one 2xx and one 4xx/5xx test alongside schemathesis run evidence.
- Go packages meet ≥90% coverage (with red-lines ≥95% for critical packages and per-file ≥80% line coverage); TypeScript packages meet ≥80% statements/≥70% branches and per-file ≥75% statements.
- Determinism harness leaves working tree clean and output hashes equal goldens on repeat runs.
- JWKS rotation, revocation bump, and cache TTL scenarios pass.
- CI suite runtime ≤10 minutes and weekly flake rate ≤0.5% recorded.

By following this plan we expect to raise confidence in token determinism, registry behaviour, and verification flows while meeting the coverage targets defined in the test strategy.
