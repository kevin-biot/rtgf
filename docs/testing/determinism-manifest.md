# Determinism Manifest

The deterministic harness under `tests/ppe-roundtrip/` produces a digest manifest for the PPE compile + evaluate round-trip.

- Expected location for generated manifest: `out/ppe/digests.json`
- Harness requirements:
  - Run workflows twice with fixed environment (`TZ=UTC`, `LANG=C`, `FIXED_TIME`, `RTGF_SEED` set by the script).
  - Validate outputs against shared schemas via AJV.
  - Record `sha256` for predicate sets, evaluation plans, tokens, and evaluation traces.
  - Fail when digests or artifacts diverge between runs.
- Command (after installing dev dependencies in `tests/ppe-roundtrip`):

```bash
cd tests/ppe-roundtrip
npm install
npm run run -- --snapshot ../../examples/policy.snapshot.json --out ../../out/ppe --manifest ../../out/ppe/digests.json
```

The manifest must be committed when artifacts change and will be compared in CI to guarantee deterministic builds.
