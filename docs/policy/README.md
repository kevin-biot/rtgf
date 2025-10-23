## Policy Source Ingestion (Developer Guide)

- Edit `policy-sources/registry/sources.yaml` to add or update authoritative URLs.
- Run `make -C policy-sources sync` to fetch and record provenance.
- Create or update predicate maps under `policy-sources/maps/predicates/`.
- Normalize a snapshot: `make -C policy-sources normalize-eu-aml` (produces signed JSON-LD).
- Feed the snapshot into RTGF: `rtgf compile --snapshot <file> --out out/`.
- Publish via RTGF Registry once validated.

**Provenance**
- Every fetch records `{source_id, url, sha256, ts}` in `policy-sources/registry/provenance.json`.
- PDFs, HTML, and datasets are archived under `policy-sources/raw/` for audit.

**Mandala Notes**
- BIS Mandala files are stored as *informative* alignment.
- If a predicate requires ZK/MPC proof, include a `@mandala` `proof_type` in `evidence_requirements`.
