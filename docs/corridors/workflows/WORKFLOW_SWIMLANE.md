# Corridor Workflow — Swimlane (S0–S7)

```
Policy Eng. | S0 Discover → S1 Ingest → S2 Normalize ┐
Predicate   |                               → S3 Map Predicates → (compile plan)
RTGF        |                                             S4 Compile Tokens → S5 Verify
RegistryOps |                                                                 S6 Publish → S7 Monitor
Auditor     |                                                   ^ PPE traces ^  transparency proofs
```

- **Policy Engineering**: maintains the source registry, fetches external documents/datasets, normalizes into Policy Snapshots.
- **Predicate Engineering**: maps snapshots to predicate definitions and deterministic evaluation plans (PPE-Compiler).
- **RTGF Build**: compiles RMT/IMT tokens embedding predicate sets + eval plans; signs and logs artefacts.
- **Registry Operations**: publishes tokens, revocations, transparency updates, `/.well-known/rtgf` metadata.
- **Auditors**: verify transparency logs, check EvalTrace digests, confirm corridor validity.

Key checkpoints:
1. **CP0** – Sources recorded with hashes (`provenance.json`).
2. **CP1** – Policy Snapshots validated (schema + signature).
3. **CP2** – Predicate sets & eval plans schema/hashes approved.
4. **CP3** – RMT/IMT tokens signed, transparency entries published.
5. **CP4** – PPE dry-run traces archived with expected outcomes.
6. **CP5** – Registry TTL/ETag monitored; alerts for drift.
