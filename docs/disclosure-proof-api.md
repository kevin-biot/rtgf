# Disclosure Proof API (Draft)

This note captures the interfaces RTGF will expose to support Regulatory Access Gateway (RAG) disclosure logging. The goal is to provide append-only, Merkle-linked proofs keyed by `trace_id` / `sta_id` / `bundle_root` so that downstream systems (RAG, DOP, regulators) can correlate disclosure events with original evidence bundles.

## Append Endpoint
```
POST /v1/disclosures
{
  "trace_id": "tx-123",
  "sta_id": "sta:abc",
  "bundle_root": "0x9af2...",
  "payload_ref": "QmZ4E2aA...",
  "policy_id": "POL-RA-0023",
  "proof_hash": "0x94c3...",
  "timestamp": "2025-11-09T12:00:00Z"
}
```
- RTGF appends the entry to the disclosure log (Merkle tree per day/region) and returns `disclosure_root` + `log_index`.

## Proof Endpoint
```
GET /v1/disclosures/proof?trace_id=tx-123&bundle_root=0x9af2...
```
Response:
```
{
  "trace_id": "tx-123",
  "bundle_root": "0x9af2...",
  "log_index": 512,
  "tree_size": 1024,
  "merkle_path": ["0xa1...","0xb2..."],
  "root": "0xde45..."
}
```
- Consumers verify the Merkle proof to confirm the disclosure event exists without requiring global consensus.

## Query by STA / Time Range
```
GET /v1/disclosures?sta_id=sta:abc&since=2025-11-09T00:00:00Z
```
Returns a list of disclosure metadata (no payloads) and corresponding proof hashes for further verification.

## Storage Model
- Each disclosure log is an append-only Merkle tree scoped by `{region}/{day}`.
- Roots are anchored in the RTGF transparency registry alongside corridor roots.
- Proofs require only the path nodes; payloads remain under the control of the RAG/evidence fabric.

## Next Steps
- Implement schemas mirroring `docs/schemas/disclosure-event.schema.json` from dop-static.
- Expose CLI tooling under `rtgf-cli` for verifying disclosure proofs.
