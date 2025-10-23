# Corridor Metrics & Dashboards

| Metric | Target | Source | Notes |
|--------|--------|--------|-------|
| Snapshot Freshness | < 24h | policy-sources/registry/provenance.json | Difference between latest fetch timestamp and now |
| Token TTL Compliance | 100% within TTL | registry transparency logs | Alert if router caches stale token |
| PPE Determinism Failures | 0 per corridor | tests/ppe-roundtrip | Re-run after predicate updates |
| Mandala Proof Alignment | 100% for corridors requiring proofs | registry/static/mandala/proof-types.json | Ensure IMT evidence_requirements satisfied |
| Transparency Gaps | 0 | registry/transparency | Check monotonic sequence |
