# Policy Source Matrix (PSM)

The Policy Source Matrix (PSM) is the canonical dataset that feeds the RTGF compiler. It enumerates signed, machine-readable policy snapshots for each jurisdiction and domain combination, allowing the compiler to emit deterministic Regulatory Matrix Tokens (RMTs) and International Mandate Tokens (IMTs).

## 1. Conceptual Model

| Dimension | Meaning |
|-----------|---------|
| Jurisdiction (rows) | Sovereign or supranational authority (e.g., EU, UK, SG, US, AU). |
| Domain (columns) | Regulatory domain (e.g., payments_aml, ai_act, data_protection, sanctions, medical, energy). |
| Cell value | A signed JSON-LD policy snapshot describing controls, prohibitions, duties, evidence requirements, normative references, and validity. |

Each non-empty cell becomes a `policy.jsonld` artefact that satisfies Section 3.1 of `draft-lane2-rtgf-00`.

## 2. Snapshot Structure

Policy snapshots **MUST** include:
- `@type`: `policy:Snapshot`
- `jurisdiction`: ISO 3166-1 alpha-2 or recognised regional code
- `domain`: consistent with the RTGF domain registry (Section 9.5)
- `effective_date`, `expires_at`: RFC 3339 UTC timestamps
- `normative_references`: list of legal sources and guidance
- `controls`, `duties`, `prohibitions`, `data_residency`, `assurance_level`
- `evidence_sources`: hashes of external datasets (sanctions lists, rulebooks)
- Detached Ed25519 signature by the issuing authority DID/JWKS

Example (EU payments/AML):
```json
{
  "@type": "policy:Snapshot",
  "jurisdiction": "EU",
  "domain": "payments_aml",
  "effective_date": "2025-10-01T00:00:00Z",
  "expires_at": "2026-10-01T00:00:00Z",
  "normative_references": [
    "Directive (EU) 2018/1673",
    "Regulation (EU) 2015/847",
    "Council Regulation (EU) 269/2014",
    "Council Regulation (EU) 833/2014"
  ],
  "controls": {
    "kyc_verification": "obligatory_before_onboarding",
    "beneficial_owner_reporting": "required",
    "transaction_monitoring": "real_time",
    "sanctions_screening": {
      "public_list": "EU_Sanctions_Map_API",
      "private_list_check": "MPC_private_set_intersection"
    }
  },
  "duties": {
    "report_suspicious_activity": "within_24h_to_FIU",
    "retain_records_years": 5
  },
  "prohibitions": [
    "transactions_with_designated_entities",
    "anonymous_payment_instruments_above_threshold"
  ],
  "data_residency": "EU_or_adequate_third_country",
  "assurance_level": "high",
  "evidence_sources": {
    "sanctions_dataset_hash": "sha256:…",
    "aml_rulebook_hash": "sha256:…"
  },
  "signature": "jws:eddsa:…"
}
```

## 3. EU Payments/AML Example

1. **Regulatory inputs**: PSD2/PSD3 drafts, AMLD 6, Wire Transfer Regulation, EU sanctions regulations, EBA ML/TF guidelines.
2. **Sanctions feeds**: EU Sanctions Map API, UN Consolidated List, OFAC SDN. The compiler captures hash values to link evidence sources.
3. **Snapshot signing**: European Commission key signs base policy; European Banking Authority may countersign AML-specific sections.
4. **Compiler output**: `RMT-EU-payments_aml-2025-10-01` including evidence hashes and policy snapshot hash.

## 4. Policy Matrix Metadata Registry

Maintain `/policy-matrix/index.jsonld` enumerating available snapshots:
```json
[
  {
    "jurisdiction": "EU",
    "domain": "payments_aml",
    "source_uri": "https://reg.europa.eu/policy/payments_aml/policy.jsonld",
    "issuing_authority_did": "did:web:eba.europa.eu",
    "last_updated": "2025-09-28T12:00:00Z"
  }
]
```

## 5. Update Workflow

1. Regulator publishes new snapshot when legislation or sanctions data changes.
2. Snapshot is signed and added to the registry index.
3. RTGF compiler ingests snapshots nightly, verifies signatures, and regenerates RMT/IMT artefacts.
4. Transparency log records each issuance with Merkle inclusion proofs for auditability.

## 6. Provenance Chain Example

1. EU Commission key signs base policy snapshot.
2. EBA key countersigns AML-specific controls.
3. Lane² RTGF compiler verifies signatures and emits RMT.
4. RTGF registry publishes token, status list, and transparency entry.
5. aARP caches tokens and enforces compliance via PDP/PEP components.

This provenance ensures regulators, operators, and auditors can validate the compliance artefacts end-to-end.
