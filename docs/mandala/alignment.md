# Mandala Alignment Guide

Lane² RMT/IMT/RTGF integrates with BIS Project Mandala by referencing its proof receipts.

- IMT/RMT tokens MAY list `evidence_requirements` referencing Mandala `proof_type` identifiers.
- PDP/PEP components verify Mandala proof receipts (CCID + proof_hash) before execution.
- Transparency logs anchor Mandala proof hashes; evidence bundles include `mandala_proofs` arrays.
- Registry servers publish `/.well-known/mandala.json` to announce supported proof types and endpoints.

Example IMT fragment:
```json
{
  "evidence_requirements": [
    "E-AML_ATTESTATION@mandala:ccid",
    "E-SANCTIONS_PROOF@mandala:zkp"
  ],
  "mandala_proofs": [
    {
      "proof_type": "zkp_sanctions",
      "provider": "https://mandala.bis.org/evidence/v1",
      "ccid": "CCID-123456",
      "proof_hash": "sha256:abc",
      "version": "1.0",
      "verified": true
    }
  ]
}
```
