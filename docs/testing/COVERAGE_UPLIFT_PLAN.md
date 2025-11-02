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
   - Store baseline numbers in `docs/testing/coverage-baseline.md`.
2. **Establish Tooling**
   - Add `package.json` + `tsconfig.json` for `rtgf-compiler/ppe-compiler` and `aarp-core/ppe-evaluator`.
   - Introduce `Makefile` targets (`make test`, `make cover`) to orchestrate Go + TS runs.
   - Configure CI workflow skeleton (lint, unit, integration, coverage upload).
3. **Governance**
   - Enable coverage badge tracking (Codecov or self-hosted) – optional but recommended.
   - Define pull request template section for test evidence.

## Phase 1 – Unit Test Expansion
1. **Go: `rtgf-verify-lib`**
   - Add negative tests covering malformed JSON, revoked tokens, nil verifier handling.
   - Introduce table-driven tests for `detectType` and metadata fallbacks.
   - Target ≥90% coverage for package.
2. **Go: `rtgf-registry/internal`**
   - Expand handler tests to cover method-not-allowed, bad slugs, JWKS caching, catalog environment variable `RTGF_URL`.
   - Add tests for `validateWindows` error branches (invalid nbf/exp formats, revoked flags).
   - Cover `HandleRevocationsGet` and concurrency on bump.
3. **TypeScript: PPE Compiler**
   - Port markdown scenarios in `rtgf-compiler/ppe-compiler/tests/roundtrip.spec.md` to executable tests (Vitest).
   - Add unit tests for CLI arg parsing, default order, and plan sequencing.
   - Stub hash calculation test (mark TODO but assert placeholder).
4. **TypeScript: PPE Evaluator**
   - Convert `aarp-core/ppe-evaluator/tests/evaluator.spec.md` into Vitest cases.
   - Mock resolvers and confirm decision matrix (PERMIT, DENY, PERMIT_WITH_CONTROLS).
   - Snapshot trace output to golden file.

## Phase 2 – Integration and Determinism
1. **Round-Trip Integration**
   - Wrap `tests/ppe-roundtrip/run.sh` in a Node script or Go test to execute determinism checks programmatically.
   - Assert JSON outputs against schemas via `ajv`.
2. **Registry <> Verifier Contract**
   - Spin up in-memory registry server (`httptest`) and call `/verify` with real fixtures.
   - Validate error codes for missing tokens, expired tokens, revoked tokens.
3. **Schema Validation Pipeline**
   - Add automated validation for `shared/ppe-schemas/*.json` against metaschema.
   - Validate generated tokens against predicate and eval plan schemas.
4. **Coverage Tracking**
   - Merge Go coverage reports using `gocovmerge`, convert to lcov.
   - Ensure Vitest coverage outputs lcov (`coverage/ts/lcov.info`).

## Phase 3 – End-to-End and Resilience
1. **Full Workflow Smoke**
   - Script end-to-end test: compile snapshot → publish to temp dir → start registry server → run `/verify` against static verifier.
   - Gate release branch with this workflow (runs nightly + on PRs touching critical paths).
2. **Fault Injection**
   - Add evaluator tests simulating resolver timeouts and malformed predicate definitions.
   - Exercise registry with missing JWKS file to confirm failure modes.
3. **Performance Baseline**
   - Introduce lightweight benchmarks (`go test -bench . ./rtgf-registry/internal/verify`) to monitor regression readiness.
4. **Coverage Gates**
   - Enforce minimum coverage in CI (`go test -coverpkg ./...`, `vitest --coverage --run`).
   - Fail PR if coverage drops >1% from baseline unless override label applied.

## Deliverables and Acceptance Criteria
- Coverage dashboards updated weekly with component-level metrics.
- All markdown spec scenarios replaced by executable tests.
- Determinism tests demonstrate byte-identical outputs for canonical fixtures.
- CI pipelines running unit, integration, and end-to-end suites with coverage gates enabled.
- Documentation updated (`docs/testing/TEST_STRATEGY.md`, `docs/testing/coverage-baseline.md`, QA changelog entries).

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

By following this plan we expect to raise confidence in token determinism, registry behaviour, and verification flows while meeting the coverage targets defined in the test strategy.
