# Reference Token Generation Framework (RTGF)

**RTGF** is the reference implementation toolkit for issuing, distributing, and verifying **Regulatory Matrix Tokens (RMTs)** and **International Mandate Tokens (IMTs)** — the cryptographically signed artefacts that express jurisdictional and cross-border regulatory rules.

RTGF complements the **Autonomous Agent Routing Protocol (aARP)** and **Deterministic Orchestration Pipeline (DOP)** by providing deterministic compilation pipelines, registry services, CLI tooling, and SDKs for fully auditable compliance token workflows.

---

## 🔍 Overview

RTGF transforms *policy snapshots* published by regulators into signed, machine-verifiable artefacts that autonomous systems can consume at runtime.

**Build-time:**
- Verify and canonicalize policy snapshots.
- Generate jurisdictional **RMTs** and bilateral **IMTs**.
- Embed deterministic predicate sets and evaluation plans (PPE-Compiler output).
- Sign and log all issuances into transparency logs.

**Run-time:**
- Serve tokens and revocation data over authenticated HTTPS registries.
- Provide SDKs and verification libraries for routers, PDPs, and external auditors.
- Support fail-closed caching and cross-registry synchronization.

---

## 📂 Repository Layout

| Path | Purpose |
|------|---------|
| `rtgf-compiler/` | Deterministic build engine for RMT/IMT, incl. the **Policy Predicate Engine (PPE-Compiler)**. |
| `rtgf-snapshots/` | Policy snapshot schemas and example jurisdictional inputs. |
| `rtgf-registry/` | Authoritative HTTP service exposing `/.well-known/rtgf`, `/rmt`, `/imt`, `/revocations`, `/transparency`. |
| `rtgf-cli/` | Administrative CLI for signing snapshots, compiling tokens, publishing, and revoking. |
| `rtgf-verify-lib/` | Verification SDKs (Go / TypeScript / Python adapters) for runtime validation. |
| `shared/ppe-schemas/` | Shared JSON-Schema definitions (predicates, evaluation plans, operators). |
| `examples/` | Sample policies and generated tokens (e.g., EU↔SG corridor). |
| `docs/` | Architecture notes, ADRs, OpenAPI specs, policy source matrix, and Internet-Drafts. |
| `docs/policy/` | Policy source ingestion matrix, big-picture workflow, developer guide. |
| `docs/corridors/` | Corridor process, playbooks, templates, metrics, and workflow changelog. |

---

## 🧠 Core Concepts

**Regulatory Matrix Token (RMT)**  
Represents a jurisdiction’s regulatory rule set for a specific domain (e.g., *EU / AI Act*).

**International Mandate Token (IMT)**  
Encodes the deterministic intersection between two RMTs, defining a lawful “corridor” for cross-border operations.

**Policy Predicate Engine (PPE)**  
Build-time compiler and run-time evaluator that translate textual legal requirements into executable, deterministic predicates.
- *PPE-Compiler* (in `rtgf-compiler/`) produces predicate sets and evaluation plans.
- *PPE-Evaluator* (in `aarp-core/`) executes them deterministically inside routing or policy-decision components.

**Transparency & Revocation**  
All tokens and revocations are logged in append-only Merkle trees with public inclusion proofs.

---

## 🚀 Quick Start

```bash
# compile a policy snapshot into tokens
rtgf compile --snapshot ./examples/policy.eu.json --out ./out/

# publish to a local registry
rtgf publish --registry http://localhost:8080 --path ./out/

# verify a token
rtgf verify ./out/RMT-EU-AI-2025-10-22.json
```

---

## 📜 Specifications
- Draft Internet-Draft: `draft-lane2-rtgf-00`
- Companion specifications:
  - `draft-aarp-00` — Autonomous Agent Routing Protocol
  - `draft-lane2-imt-rmt-00` — RMT/IMT Token Format
  - `draft-lane2-corridor-registry-00` — Corridor & Domain Naming Registry

## 🧭 Additional References
- `docs/policy/BIG_PICTURE.md` — end-to-end ingestion summary.
- `docs/corridors/process/PROCESS_AND_MAPPING.md` — corridor S0–S7 workflow.
- `docs/mandala/alignment.md` — Mandala proof interoperability guide.
- `docs/MVP_SCOPE.md` — unified deterministic automation MVP scope across CaaS, DOP, aARP/SAPP, and RTGF.
- `docs/RTGF_INTEGRATION_GUIDE.md` — practical integration steps for posting evidence bundles and consuming replay APIs.

---

## 📌 Status
- Stage: Draft / Early Implementation
- License: Apache-2.0 (reference code); deterministic evaluation core under Lane² IP.
- Maintainer: Lane² Architecture — info@lane2.ai
- Contributions: Pull requests and regulator feedback welcome.
- Version alignment: Works with Lane² Deterministic Compliance Stack v0.4+

© 2025 Lane² Architecture. Patent applications pending (GB 2517464.0 and related). Draft specifications submitted to the IETF under BCP 78/79.
