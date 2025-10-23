# Reference Token Generation Framework (RTGF)

RTGF is the reference implementation toolkit for issuing, distributing, and verifying Regulatory Matrix Tokens (RMTs) and International Mandate Tokens (IMTs). It complements the aARP routing fabric by providing deterministic compilation pipelines, registry services, CLI tooling, and SDKs for compliance token workflows.

## Repository Layout

```
rtgf-compiler/      # deterministic build engine for RMT/IMT
rtgf-snapshots/     # policy snapshot schemas and example inputs
rtgf-registry/      # HTTP distribution service exposing /.well-known/rtgf, /rmt, /imt, /revocations, /transparency
rtgf-cli/           # administrative CLI for signing snapshots, compiling tokens, publishing and revoking
rtgf-verify-lib/    # verification SDKs (Go / TypeScript / Python adapters)
examples/           # sample policies and generated tokens (e.g., EU↔SG corridor)
docs/               # architecture notes, ADRs, OpenAPI specs, policy source matrix
```

## Status

- Draft Internet-Draft: `draft-lane2-rtgf-00`
- Companion specifications: `draft-aarp-00`, `draft-lane2-imt-rmt-00`

This repository will evolve alongside the Lane2 deterministic compliance stack. Contributions welcome.
