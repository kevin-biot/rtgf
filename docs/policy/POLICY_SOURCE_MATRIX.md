# Policy Source Matrix (PSM)

Purpose: map jurisdictions × domains → authoritative inputs and their normalization pipeline.

| Jurisdiction | Domain        | Primary Sources                                   | Live Datasets                | Normalization Output                   |
|--------------|---------------|----------------------------------------------------|------------------------------|----------------------------------------|
| EU           | payments_aml  | AMLD6, Reg 2015/847                                | EU Sanctions Map (API)       | `EU.payments_aml.snapshot.signed.json` |
| EU           | ai            | EU AI Act 2024/1689                                | —                            | `EU.ai.snapshot.signed.json`           |
| INTL         | corridors     | BIS Mandala Core + Technical                       | —                            | Informative references                 |

**Workflow**
1. **Discover & Fetch**: `make -C policy-sources sync`
2. **Normalize**: `make -C policy-sources normalize-eu-aml`
3. **Compile** (RTGF): `rtgf compile --snapshot rtgf-snapshots/eu/... --out out/`
4. **Publish** (Registry): `rtgf publish --registry https://reg.example.com --path out/`

**Mandala Alignment**
- Mandala tech reports are informative; we *reuse* the concepts of CCID and proof receipts.
- Where privacy proofs are required, the IMT `evidence_requirements` SHOULD include `@mandala` proof types.
