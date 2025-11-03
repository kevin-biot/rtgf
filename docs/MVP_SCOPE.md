# Unified Deterministic Automation MVP Scope

This scope document describes how RTGF integrates with the broader deterministic automation MVP spanning four coordinated repositories. The goal is to deliver an end-to-end, replayable demonstration of lawful, context-aware automation that generates cryptographically verifiable evidence.

---

## Overview

| Stack Component | Repository (lead) | Purpose |
|-----------------|-------------------|---------|
| CaaS (Context-as-a-Service) | Network Edge / Telco | Generate and sign real-time contextual evidence from network and IoT signals. |
| Mini-CaaS / DOP-Lib | Device SDK | Local, privacy-preserving context fusion on the user’s device. |
| DOP (Deterministic Orchestration Pipeline) | API Server | Normalize phrases, gate execution on confidence, route capabilities, and generate evidence. |
| aARP / SAPP Mock Compliance Layer | Compliance Simulator | Demonstrate lawful routing and cryptographic compliance for automated payments. |
| **RTGF (Replay & Trace Governance Fabric)** | This repo | Aggregate, verify, and expose deterministic replay manifests across all traces. |

---

## 1. CaaS (Context-as-a-Service – Network Edge / Telco)
- **Purpose:** Generate and sign real-time contextual evidence artifacts from network and IoT signals.  
- **Inputs:** Device location, dwell time, motion, Wi-Fi, temporal context.  
- **Process:** Deterministic enrichment → fused confidence score (Σ wᵢ·cᵢ) → Ed25519 signature over artifact.  
- **Outputs:** JSON artifact with fields `{trace_id, interval, candidate, scores, policy, provenance, signature}`.  
- **API:** `POST /caas/v1/artifact`, `GET /caas/v1/artifact/{id}`.  
- **Performance:** <150 ms enrichment latency, >95% success rate.  
- **Goal:** Provide signed context evidence (e.g., “dwell@Café Milano 18:05–18:45 confidence 0.91”) consumable by DOP and RTGF.

## 2. Mini-CaaS / DOP-Lib (On-Device Context Library)
- **Purpose:** Run locally on phone or edge client as a privacy-preserving “personal CaaS”.  
- **Inputs:** GNSS, motion, app state, optional Telco CaaS artifacts.  
- **Process:** Fuse signals + external artifacts → compute fused confidence and automation mode (auto / ask / deny).  
- **Outputs:** Signed mini-CaaS artifact (same schema as CaaS) delivered via SDK.  
- **Features:** Version-pinned rubrics, fail-closed defaults, deterministic replay, optional consent UI.  
- **Goal:** Supply DOP with device-side context under the same deterministic model.

## 3. DOP (Deterministic Orchestration Pipeline API Server)
- **Purpose:** Normalize user phrases, gate execution on confidence, deterministically route capabilities, and generate hash-linked evidence contracts.  
- **Core Endpoints:**  
  - `POST /dop/v1/normalize` – canonicalize phrase using static dictionary + context artifact; returns confidence & decision.  
  - `POST /dop/v1/route` – choose deterministic mock payment path; hash-link to normalization record.  
  - `POST /dop/v1/evidence/contract` – bind normalization + route + execution provenance into Merkle-rooted evidence.  
  - `POST /dop/v1/identity/verify` – mock eIDAS/EUDI credential verification producing signed identity-attestation artifact.  
  - `GET /dop/v1/replay/{trace_id}` – reconstruct deterministic chain.  
- **Security & Storage:** Version-pinned policy snapshot, idempotency keys, Ed25519 signatures, append-only log; in-memory storage for MVP with hash links between stages.  
- **Goal:** Act as central controller proving phrase normalization, context integration, identity verification, and evidence generation.

## 4. aARP / SAPP Mock Compliance Layer
- **Purpose:** Demonstrate lawful routing and cryptographic compliance for automated payments.  
- **aARP:** Validates jurisdictional corridors via mock IMT/RMT tokens; denies routing when invalid.  
- **SAPP:** Wraps payment execution with mock federation contract, evidence bundle, and liability propagation.  
- **Integration:** DOP `/route` queries aARP for “lawful route ok”; `/evidence/contract` triggers SAPP to attach mock compliance evidence (Merkle root + demo signatures).  
- **Goal:** Show deterministic, regulation-compliant transaction flow without real money movement.

---

## Unified MVP Flow

1. **Context** – CaaS (Telco) and mini-CaaS (Device) emit signed artifacts describing real-world state.
2. **Phrase Normalization** – Client posts “pay John five pounds” + context artifact to DOP `/normalize`; DOP computes confidence and gates execution.
3. **Lawful Routing** – DOP `/route` consults mock aARP for jurisdictional validation and selects deterministic payment path.
4. **Identity Verification** – DOP `/identity/verify` checks demo EU digital ID credential, creating identity artifact.
5. **Evidence Contract** – DOP `/evidence/contract` binds normalization + context + identity + routing + execution into a Merkle-rooted evidence record.
6. **Replay & Audit** – `/replay/{trace_id}` assembles the chain; RTGF consumes the outputs to provide cross-system audit.

---

## MVP Objectives

- **End-to-End Determinism:** identical inputs yield identical outputs with full hash-linked traceability.  
- **Fail-Closed Security:** execution halts if confidence < threshold or corridor tokens invalid.  
- **Cryptographic Evidence:** every stage signed and logged; Merkle root per transaction.  
- **Interoperability Proof:** context (CaaS), phrase (DOP), routing (aARP), compliance (SAPP), identity (EU ID mock), and governance (RTGF) operate as one lawful fabric.  
- **Demonstration Outcome:** verifiable JSON audit trail showing a context-aware, regulation-ready, deterministic “pay” action executed end-to-end without external dependencies or live data.

---

# RTGF — Replay & Trace Governance Fabric (This Repo)

## Purpose
Provide deterministic audit, replay, and cross-system trace governance for all `trace_id` events produced by the other MVP components.

## MVP Objectives
1. Aggregate logs/artifacts from DOP, CaaS, mini-CaaS, aARP, and SAPP into a unified replay record.  
2. Verify hash-links, signatures, and Merkle roots across systems.  
3. Expose `GET /rtgf/v1/replay/{trace_id}` API for cross-repo audit.  
4. Produce integrity metrics (`hash_ok`, `sig_ok`, `linkage_ok`).

## Key Functions

- Ingest JSON evidence bundles via webhook (`POST /rtgf/v1/ingest`) or file drop.  
- Validate cross-links:  
  - `normalization.hash == route.link`.  
  - `evidence.merkle_root` includes normalization + route hashes.  
  - Signature verification passes for each artifact.  
- Build composite replay manifest:
  ```json
  {
    "trace_id": "trace-123",
    "components": {
      "caas": {...},
      "mini_caas": {...},
      "dop": {...},
      "aarp": {...},
      "sapp": {...}
    },
    "integrity": {
      "hash_ok": true,
      "sig_ok": true,
      "linkage_ok": true
    }
  }
  ```
- Expose replay viewer (`GET /rtgf/v1/replay/{trace_id}`) and diff tool for comparing runs.

## Architecture

- Lightweight verifier service (Python or Node) consuming Ontology schemas for validation.  
- Local SQLite store preserving the most recent traces for inspection.  
- CLI + REST surface sharing validation logic.

## Integration

- DOP `/replay` posts deterministic chain to RTGF after evidence contract finalization.  
- RTGF acts as independent audit fabric across all domains, enabling third-party verification.

## Deliverables

- MVP CLI + REST verifier.  
- Example replay manifest for a complete “pay John five pounds” trace.  
- Integrity report and visualization output (JSON + HTML) demonstrating hash-link verification.  
- Documentation describing ingestion process, validation rules, and API usage.

### Clarifications: Evidence Schema, Hashing, Keys, and Ingestion Lifecycle

#### Shared Evidence Schema
All producing components MUST emit artifacts conforming to the versioned schema at `https://ontology.example.com/schema/evidence-bundle.json`:

```json
{
  "$id": "https://ontology.example.com/schema/evidence-bundle.json",
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "UnifiedEvidenceBundle",
  "type": "object",
  "required": ["trace_id", "producer", "artifact_type", "hash", "signature", "schema_version"],
  "properties": {
    "trace_id": { "type": "string", "pattern": "^trc-[A-Za-z0-9]{8,}$" },
    "artifact_type": {
      "type": "string",
      "enum": ["context", "normalization", "route", "identity", "evidence"]
    },
    "producer": { "type": "string", "enum": ["caas", "dop", "aarp", "sapp", "mini-caas"] },
    "schema_version": { "type": "string", "example": "v0.1.0" },
    "payload": { "type": "object" },
    "hash": {
      "type": "string",
      "description": "sha256 over RFC8785 canonical JSON of payload"
    },
    "signature": {
      "type": "object",
      "required": ["alg", "kid", "value"],
      "properties": {
        "alg": { "type": "string", "enum": ["EdDSA"] },
        "kid": { "type": "string" },
        "value": { "type": "string", "contentEncoding": "base64" }
      }
    }
  }
}
```

#### Canonicalisation & Hashing Rules
- Hashes use **SHA-256** over **RFC 8785 canonical JSON** of the `payload`. `hash` MUST equal `"sha256:" + hex_digest`.  
- `trace_id` propagates unchanged across all components.  
- `normalization.hash` == `route.linked_hash`.  
- `evidence.merkle_root` includes hashes from normalization + route (verifiable via Merkle proof).  
- RTGF recomputes these values; mismatches trigger integrity alarms.

#### Signature & Key Distribution

| Component  | JWKS URL                                             | Rotation  | Scope                          |
|------------|------------------------------------------------------|-----------|--------------------------------|
| CaaS       | `https://caas.example.com/.well-known/jwks.json`     | 90 days   | Context artifacts              |
| DOP        | `https://dop.example.com/.well-known/jwks.json`      | 90 days   | Normalization, route, evidence |
| aARP       | `https://aarp.example.com/.well-known/jwks.json`     | 180 days  | Route admission proofs         |
| SAPP       | `https://sapp.example.com/.well-known/jwks.json`     | 180 days  | Compliance bundles             |
| mini-CaaS  | Registered device key (embedded)                     | lifecycle | Local context artifacts        |

RTGF caches JWKS for 24 hours and re-fetches when verification fails.

#### Ingestion Interface & Lifecycle

1. **Webhook (preferred)** – `POST /rtgf/v1/ingest`
   - Headers: `Authorization: Bearer <token>`, `Content-Type: application/json`.  
   - Auth: mutual TLS or OAuth2 client credentials.  
   - Body: one or more `UnifiedEvidenceBundle` objects (≤20 per request).  
   - Response: `202 Accepted` with per-artifact status.

2. **File Drop / Pull** – Daily JSONL batches at `/exports/YYYY/MM/DD/`, collected hourly by RTGF.

Producers call the webhook immediately after artifact creation (or within 15 minutes if offline). RTGF verifies within 60 seconds of receipt.

#### Error Budgets & Observability
- RTGF notifies producers via `POST {component}/v1/integrity/failure` containing `{trace_id, artifact_id, reason, detected_at}`.  
- Metrics: `rtgf_ingest_total{component}`, `rtgf_integrity_failures_total{component,reason}`, `rtgf_verification_latency_seconds`.  
- SLO: 99% of artifacts verified within 60 s; integrity-check false positives <0.1%.

#### Validation Summary

| Artifact Type | Producer          | Key (kid)   | Hash Field            | RTGF Validation |
|----------------|-------------------|-------------|-----------------------|-----------------|
| context        | CaaS / mini-CaaS  | `caas_kid`  | `payload.hash`        | Schema, hash, signature, trace continuity |
| normalization  | DOP               | `dop_kid`   | `normalization.hash`  | Equals `route.linked_hash` |
| route          | DOP + aARP        | `dop_kid`, `aarp_kid` | `route.hash` | Token signature & lawful route result |
| identity       | DOP               | `dop_kid`   | `id_artifact.hash`    | Credential signature validation |
| evidence       | DOP + SAPP        | `dop_kid`, `sapp_kid` | `evidence.merkle_root` | Merkle proof includes normalization + route hashes |

**Success Criterion:** RTGF deterministically recomputes hashes, verifies signatures via JWKS, and builds a unified replay manifest without missing or ambiguous fields.

---

**Success Criterion:** all five repositories deliver their modules so that, when composed, the system demonstrates deterministic orchestration, context fusion, mock identity verification, lawful routing, and audit-grade evidence generation with RTGF providing unified replay governance.
