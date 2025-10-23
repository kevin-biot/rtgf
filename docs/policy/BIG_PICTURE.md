# RTGF Policy Ingestion & Corridor Big Picture

```
External Sources ──► Source Registry ──► Raw Archive ──► Policy Snapshots ──► PPE Predicates ──► RTGF Tokens ──► Registry & Routers
       (laws, APIs)      (sources.yaml)     (hash, ts)        (JSON-LD)            (plan hash)        (RMT/IMT)             (HTTPS)
```

1. **Source Discovery (S0)**: Policy engineers register authoritative URLs (legislation, datasets, Mandala documentation) in `policy-sources/registry/sources.yaml`.
2. **Fetch & Provenance (S1)**: `make -C policy-sources sync` archives files under `policy-sources/raw/`, computes SHA-256 hashes, records entries in `registry/provenance.json`.
3. **Normalization (S2)**: `normalize_snapshot.py` converts raw content into signed Policy Snapshots in `rtgf-snapshots/<jur>/...json`.
4. **Predicate Mapping (S3)**: PPE-Compiler transforms snapshots into `predicate_set` + `eval_plan` artefacts using shared schemas (`shared/ppe-schemas/...`).
5. **Compilation (S4)**: RTGF builds RMTs/IMTs, embedding predicate data and referencing Mandala proof types where required.
6. **Verification (S5)**: PPE-Evaluator dry-runs contexts; hashes & signatures checked; transparency entries staged.
7. **Publication (S6)**: Registry publishes tokens (`/.well-known/rtgf`, `/rmt`, `/imt`, `/revocations`, `/transparency`).
8. **Monitoring (S7)**: Nightly sync workflow ensures freshness; metrics dashboards track TTL, transparency, resolver health.

Supporting documents:
- `docs/corridors/process/PROCESS_AND_MAPPING.md` — end-to-end corridor workflow.
- `docs/policy/POLICY_SOURCE_MATRIX.md` — jurisdiction × domain matrix.
- `docs/mandala/alignment.md` — Mandala proof interoperability.
- `docs/ppe/` — Policy Predicate Engine design, embedding, and actions.
