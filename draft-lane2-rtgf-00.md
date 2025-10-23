Internet-Draft: draft-lane2-rtgf-00
Intended status: Experimental
Expires: 13 April 2026
Individual Submission – Lane2 Architecture

Lane2 Architecture                                            October 13, 2025

## Status of This Memo
This Internet-Draft is submitted in full conformance with the provisions of BCP 78 and BCP 79. Internet-Drafts are working documents of the Internet Engineering Task Force (IETF). Note that other groups may also distribute working documents as Internet-Drafts. The list of current Internet-Drafts is at https://datatracker.ietf.org/drafts/current/.

Internet-Drafts are draft documents valid for a maximum of six months and may be updated, replaced, or obsoleted by other documents at any time. It is inappropriate to use Internet-Drafts as reference material or to cite them other than as “work in progress.”

Copyright (c) 2025 IETF Trust and the persons identified as the document authors. All rights reserved.

This document is subject to BCP 78 and the IETF Trust's Legal Provisions Relating to IETF Documents (https://trustee.ietf.org/license-info) in effect on the date of publication of this document. Please review these documents carefully, as they describe your rights and restrictions with respect to this document.

# draft-lane2-rtgf-00 — Reference Token Generation Framework (RTGF) for RMT/IMT

*Status:* Working Draft 0.1  
*Editors:* K. Brown et al. (Lane2 Architecture)  
*Date:* 2025-10-13

## Abstract
The Reference Token Generation Framework (RTGF) defines the authoritative pipeline, distribution interfaces, and verification profile for Regulatory Matrix Tokens (RMTs) and International Mandate Tokens (IMTs). RTGF standardises how regulators publish policy snapshots, how compilers deterministically derive tokens, how registries expose them over HTTPS, and how verifiers consume them within the Autonomous Agent Routing Protocol (aARP) ecosystem. This document pairs with `draft-aarp-00` and `draft-lane2-imt-rmt-00` to provide a complete compliance-by-design architecture for autonomous agents.

## 1. Terminology
This specification uses the keywords **MUST**, **MUST NOT**, **SHALL**, **SHALL NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** as described in RFC 2119 and RFC 8174 when, and only when, they appear in all capitals.

| Term | Definition |
|------|------------|
| RMT (Regulatory Matrix Token) | Jurisdiction- and domain-specific regulatory bundle produced by an RTGF compiler. |
| IMT (International Mandate Token) | Bilateral corridor token derived from two RMTs for a given domain (ordered `JUR_A->JUR_B`). |
| Policy Snapshot | Signed, versioned JSON-LD artefact describing the authoritative regulatory rules for a jurisdiction/domain. |
| Compiler | Deterministic RTGF engine that ingests policy snapshots and emits RMT/IMT tokens. |
| Transparency Log | Append-only Merkle log recording issuance and revocation events. |
| Registry | HTTPS service exposing RTGF distribution endpoints compatible with aARP discovery. |

Requirement identifiers follow the pattern `RTGF-REQ-<NNN>`.

## 2. Threat Model and Goals
- **Adversaries**: attempt token forgery, corridor manipulation, stale-policy replay, selective disclosure, or revocation suppression.
- **Goals**: deterministic builds, short-lived tokens, verifiable provenance, easy verification, strict fail-closed behaviour, and compatibility with aARP caching.

RTGF assumes regulator-controlled signing keys protected by HSM/KMS and that transport occurs over mutually authenticated TLS 1.3 sessions.

## 3. Token Generation Pipeline

### 3.1 Inputs
RTGF-REQ-001: Inputs **MUST** be signed policy snapshots (`policy.jsonld`) containing:
- `jurisdiction` (ISO 3166-1 alpha-2 or regional code)
- `domain` identifier
- `effective_date`, `expires_at`
- `normative_references` (list of legal sources)
- `policy_snapshot_hash` (SHA-256 over canonicalised policy body)
- Detached Ed25519 JWS signature by the regulatory authority DID/JWKS (or an authorised delegate)

Snapshots **MUST** be canonicalised using RFC 8785 (JCS) before hashing and signing. All timestamps in snapshots and derived tokens **MUST** use RFC 3339 format with UTC designator `Z`.

### 3.2 Compiler Requirements
RTGF-REQ-002: Compilers **MUST** perform the following steps deterministically:
1. Verify the snapshot signature and trust chain.
2. Canonicalise snapshot payload using RFC 8785.
3. Emit RMT tokens per jurisdiction/domain.
4. For each ordered corridor `(A,B)`, compute IMT = INTERSECT(RMT_A, RMT_B) using the algorithm in Section 3.3.
5. Attach `scope_hash = sha256(sorted(policy_hashes))` and `build_id` to each token.

RTGF-REQ-003: Output tokens **MUST** be signed with Ed25519 JWS over canonicalised JSON and include JOSE headers with `alg=EdDSA` and a `kid` resolvable via the issuer JWKS. Tokens **MUST NOT** carry JOSE `crit` parameters unless this specification defines them. Each token **MUST** contain the following fields:
- `rmt_id` or `imt_id`
- `jurisdiction` or `corridor`
- `domain`
- `effective_date`, `expires_at`
- `policy_snapshot_hash` (and `derived_from` for IMTs)
- `controls`, `prohibitions`, `duties`, `assurance_level`, `data_residency` (as applicable)
- `ttl_sec` (≤ 86400 recommended)
- `revocation_info` (URI or status list pointer)
- `kid` (JWKS key id)
- `signature`
- `compiler_version`, `build_manifest`

RTGF-REQ-004: Tokens MAY include `signatures[]` for multi-signature endorsement. Verifiers **MUST** enforce the threshold defined in the token metadata or associated policy snapshot.

RTGF-REQ-005: Compilation **MUST** be deterministic—identical inputs yield identical token bytes. Compilers **MUST** pin schema versions, sort all JSON object keys, include a `policy_max_ttl` value consistent with Section 5, and provide reproducible build manifests.

`policy_max_ttl` denotes the maximum token lifetime advertised by the issuer and **MUST** align with freshness policies described in Section 5.

### 3.3 IMT Intersection Algorithm
```
INTERSECT(RMT_A, RMT_B):
  effective_window   = overlap(A.effective_date..A.expires_at, B.effective_date..B.expires_at)
  prohibitions       = union(A.prohibitions, B.prohibitions)
  controls           = union(A.controls, B.controls)
  duties             = most_restrictive(A.duties, B.duties)
  data_residency     = most_restrictive(A.data_residency, B.data_residency)
  assurance_level    = min(A.assurance_level, B.assurance_level)
  policy_hash        = sha256(sort([A.policy_snapshot_hash, B.policy_snapshot_hash]))
  ttl_sec            = min(A.ttl_sec, B.ttl_sec, policy_max_ttl)
  conflicts          = detect_conflicts(A, B)
```
Conflicts **MUST** be recorded in `imt.effective_conflicts` for manual resolution or policy updates.

## 4. Distribution and Caching
RTGF registries expose TLS 1.3 endpoints compatible with aARP discovery and **MUST** follow the guidance in RFC 9325 when configuring TLS parameters, cipher suites, and certificate validation.

RTGF-REQ-010: Servers **MUST** publish a well-known resource at `/.well-known/rtgf` containing registry roots, trust anchors, supported domains, and transparency log endpoints.

RTGF-REQ-011: The following endpoints **MUST** be implemented:
- `GET /rmt/{jurisdiction}/{domain}`
- `GET /imt/{corridor}/{domain}` (corridor format `SRC-DST`)
- `GET /revocations` (status list or delta via `?since=`)
- `GET /transparency?since=` (Merkle inclusion/consistency proofs)

Responses **MUST** supply `Content-Type: application/imt-rmt+json`, `ETag`, and `Cache-Control`. TTL SHOULD be ≤ 24 hours. Registry error responses **MUST** use `application/problem+json` (RFC 9457) with stable `type` identifiers.

RTGF-REQ-012: Clients **MUST** revalidate cached tokens before expiry and **MUST** fail closed if a fresh token cannot be obtained.

## 5. Verification Profile
RTGF-REQ-020: Verifiers (routers, PDPs, PEPs, SDKs) **MUST**:
1. Validate Ed25519 JWS signature against trusted JWKS.
2. Enforce `nbf`/`expires_at` with ±120 s clock skew tolerance.
3. Confirm `policy_snapshot_hash` matches the policy snapshot referenced by the STA or local catalog.
4. Check revocation status via status list or transparency proof.
5. Ensure `corridor` / `domain` match the requested execution context.
6. Apply `controls`, `prohibitions`, and `duties` to the planned operation.
7. Deny execution (`imt_verification_failed`) on any failure.

RTGF-REQ-021: Verifiers **MUST** maintain a bounded cache of observed `jti` values per issuer for at least the greater of `ttl_sec` or 24 hours to prevent replay of revoked or expired tokens.

## 6. Revocation and Transparency
RTGF-REQ-030: Registries **MUST** provide signed revocation lists mapping `jti` to status values; lists MAY be compressed bitsets.

RTGF-REQ-031: Transparency logs **MUST** record every issuance and revocation with Merkle proofs accessible via the registry. Compilers **MUST** submit proofs to the aARP evidence plane for auditability.

## 7. Security Considerations
- Issuer keys SHOULD reside in HSM/KMS with dual control. Rotations MUST overlap and be published via JWKS.
- Multi-signature endorsements mitigate single-authority compromise.
- Short TTLs and mandatory revocation checks limit stale-token replay.
- Algorithm agility: Ed25519 signatures are mandatory-to-implement. Deployments **SHOULD** plan for hybrid Ed25519 + PQ signatures (e.g., ML-DSA) once standardised by the IETF or CFRG, and issuers **MUST** announce algorithm changes via transparency events.
- Transport security follows TLS 1.3 with ALPN `aarp/1` and mutual authentication.

## 8. Privacy Considerations
RTGF artefacts contain regulatory metadata only; personal data MUST NOT appear. Sanction lists SHOULD use pseudonymous identifiers with regulated lookup procedures. Transparency artefacts MUST avoid exposing sensitive operational details beyond token metadata.

## 9. IANA Considerations
This section follows the policies defined in RFC 8126.

### 9.1 Media Type Registration
- Type name: application
- Subtype name: imt-rmt+json
- Required parameters: `charset` (default UTF-8)
- Optional parameters: none
- Encoding considerations: 8bit (UTF-8 JSON text)
- Security considerations: Tokens are detached JWS-protected JSON-LD envelopes; see Sections 3, 5, and 7 of this document.
- Interoperability considerations: none
- Published specification: this document and `draft-lane2-imt-rmt-00`
- Applications that use this media type: RTGF registries, compilers, verifiers within aARP deployments
- Additional information: file extensions: none; fragment identifier considerations: N/A
- Person & email address for further information: Lane2 Architecture (draft authors)
- Intended usage: COMMON
- Restrictions on usage: none
- Author: Lane2 Architecture
- Change controller: IETF

### 9.2 Well-Known URI Registration
- URI suffix: `rtgf`
- Change controller: IETF
- Specification document(s): this document (Section 4)

### 9.3 RMT Jurisdiction Registry
Create a registry mapping jurisdiction codes to issuing authority metadata.

Registration policy: Expert Review. Designated experts **MUST** confirm that the submitter represents a recognised regulatory authority, that a DID/JWKS is published over HTTPS, that the jurisdiction code is ISO 3166-1 alpha-2 or recognised regional code, and that transparency endpoints are reachable.

Registry fields: `jurisdiction_code`, `authority_name`, `authority_did`, `jwks`, `transparency_log_endpoint`, `status`, `updated_at`.

### 9.4 IMT Corridor Registry
Create a registry mapping corridor identifiers to metadata.

ABNF:
```
corridor = jur "-" jur
jur      = 2ALPHA / "EU" / "UK" / "EA"
```

Registration policy: Expert Review with the same criteria as Section 9.3 plus verification that the proposed corridor identifier is unique. Registry fields: `corridor_id`, `source_jurisdiction`, `destination_jurisdiction`, `domain_set`, `issuing_authority`, `jwks`, `status`, `updated_at`.

### 9.5 Domain Code Registry
Create a registry for domain identifiers used in tokens and policy snapshots.

Syntax: `domain = 1*( ALPHA / DIGIT / "-" / "_" )`

Registration policy: Expert Review ensuring descriptive, versioned identifiers accompanied by normative references.

### 9.6 IMT/RMT Error Code Registry
Create a registry for canonical error codes surfaced via `application/problem+json`. Registration policy: Specification Required. Initial values include `imt_verification_failed` (HTTP 403) and `snapshot_signature_invalid` (HTTP 400).

## 10. Normative References
- RFC 2119 — Key words for use in RFCs to Indicate Requirement Levels
- RFC 8174 — Ambiguity of Uppercase vs Lowercase in RFC 2119 Key Words
- RFC 3339 — Date and Time on the Internet: Timestamps
- RFC 7515 — JSON Web Signature (JWS)
- RFC 7517 — JSON Web Key (JWK)
- RFC 8446 — The Transport Layer Security (TLS) Protocol Version 1.3
- RFC 9325 — Recommendations for Secure Use of Transport Layer Security (TLS) and Datagram Transport Layer Security (DTLS)
- RFC 8615 — Well-Known Uniform Resource Identifiers (URIs)
- RFC 8785 — JSON Canonicalization Scheme (JCS)
- RFC 9110 — HTTP Semantics
- RFC 9114 — HTTP/3
- RFC 9457 — Problem Details for HTTP APIs
- draft-aarp-00 — Autonomous Agent Routing Protocol (aARP)
- draft-lane2-imt-rmt-00 — Regulatory Matrix Tokens and International Mandate Tokens

## 11. Informative References
- Lane2 Regulatory Matrix White Paper (2025)
- EU Artificial Intelligence Act (2024/1689)
- Project Mandala Technical Report (BIS, 2024)

## 12. Change Control
Future revisions will refine compiler profiles, multi-signature governance, and PQ migration guidance. Changes will remain aligned with updates to `draft-aarp-00` and `draft-lane2-imt-rmt-00`. All contributions fall under the IETF Note Well, and the Lane² reference implementation is released under the Apache License 2.0.

## 13. Revision History
- v0.1 (2025-10-13): Initial draft defining RTGF pipeline, distribution, verification, and governance requirements.

## 14. Acknowledgements
The authors thank early reviewers and contributors to the Lane² programme for comments on determinism, transparency, and multi-jurisdiction token governance. A permissive reference implementation is maintained at https://github.com/lane2-rtgf under the Apache License 2.0; Lane² Architecture asserts no patent rights over the RTGF specification to encourage adoption as public regulatory infrastructure.
