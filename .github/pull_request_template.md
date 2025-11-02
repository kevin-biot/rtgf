# Summary

<!-- Briefly describe the change. -->

## Testing

- [ ] `go test ./...`
- [ ] `npm test` / `pnpm test` (TypeScript packages)
- [ ] Integration / determinism harness (if applicable)

## Quality Gates

- [ ] Coverage meets or exceeds required thresholds (Go ≥90% packages / ≥95% critical, TS ≥80% statements / ≥70% branches).
- [ ] Determinism check (`git status --porcelain` clean after test run, digest manifest unchanged).
- [ ] JWKS rotation / revocation scenarios touched? If yes, attach results.
- [ ] Schemathesis / contract tests executed when API surface changes.

## Notes

<!-- Include links to coverage reports, determinism manifests, or follow-up tasks. -->
