#!/usr/bin/env bash
# RTGF ADR Bootstrap & Gap Filler
# Usage: scripts/rtgf_adr_bootstrap.sh [-r repo_root] [-d adr_dir] [--commit] [--force]
# Creates/updates the RTGF ADR template and seed ADR documents without overwriting existing
# files unless --force is supplied. Optionally stages & commits the changes.
set -euo pipefail

print_usage() {
  cat <<'USAGE'
Usage: scripts/rtgf_adr_bootstrap.sh [options]

Options:
  -r <repo_root>   Repository root (default: current directory)
  -d <adr_dir>     Target ADR directory (default: <repo_root>/docs/architecture/rtgf/adr)
  --commit         Stage + commit generated files if inside a git repo
  --force          Overwrite existing ADR files instead of skipping
  -h, --help       Show this help message

Environment overrides:
  ADR_ACCEPTED_DATE   Accepted ADR effective date (default: 2025-11-01)
  ADR_TARGET_ACCEPT   Target acceptance date for proposed ADRs (default: 2026-01-31)
USAGE
}

ROOT="."
ADIR=""
DO_COMMIT=false
FORCE=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    -r)
      ROOT="${2:-}"
      shift 2
      ;;
    -d)
      ADIR="${2:-}"
      shift 2
      ;;
    --commit)
      DO_COMMIT=true
      shift
      ;;
    --force)
      FORCE=true
      shift
      ;;
    -h|--help)
      print_usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      print_usage
      exit 1
      ;;
  esac
done

ROOT="${ROOT:-.}"
if [[ ! -d "$ROOT" ]]; then
  echo "Repo root $ROOT does not exist." >&2
  exit 1
fi

if [[ -z "$ADIR" ]]; then
  ADIR="$ROOT/docs/architecture/rtgf/adr"
else
  case "$ADIR" in
    /*) ADIR="$ADIR" ;;
    *) ADIR="$ROOT/$ADIR" ;;
  esac
fi

mkdir -p "$ADIR"

timestamp="$(date +%F)"
accepted_date="${ADR_ACCEPTED_DATE:-2025-11-01}"
target_accept="${ADR_TARGET_ACCEPT:-2026-01-31}"

write_file() {
  local path="$1"
  local marker="$2"
  if [[ -f "$path" && $FORCE == false ]]; then
    echo "• Skipping existing $marker (use --force to overwrite)"
    return 1
  fi
  shift 2
  cat > "$path" <<'EOF'
EOF
  # The here-doc above is intentionally empty; we will append content next.
  : > "$path"
  cat >> "$path" <<EOF
$@
EOF
  echo "✔  Wrote $marker → ${path#$ROOT/}"
  return 0
}

# 0) Template
template_path="$ADIR/_ADR_TEMPLATE.md"
write_file "$template_path" "ADR template" '# ADR-XXX: <Title> — Rev for Acceptance

**Status:** Proposed → *Ready for Acceptance review*  
**Date:** YYYY-MM-DD  
**Decision Makers:** <Names / Group>  
**Owner:** <Team / Role>  
**Target Acceptance:** YYYY-MM-DD  
**Related ADRs:** <List other ADRs / specs>

---

## 1) Purpose & Scope
<What problem do we solve, for whom, and why now? Link to Context.>

## 2) Architecture / Data Model
<Describe the essential architecture, key components, and/or canonical schemas. Diagrams encouraged.>

## 3) Determinism & Provenance
<How is determinism achieved (JCS canonicalization, fixed seeds, build graph pinning)? What is recorded in revEpoch? How are proofs anchored?>

## 4) Security & Trust
<mTLS, JWKS, DANE/TLSA, signature scheme, consent/policy gates, fail-closed behaviour.>

## 5) Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| XYZ_… | … | … |

## 6) Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|-------|
| … | … | … |

Prometheus metrics:  
`component_metric_a`, `component_metric_b`, …

## 7) Test Plan Mapping
| Test ID | Scenario | Expected outcome |
|---------|----------|------------------|
| CTX-NN  | …        | …                |

## 8) Acceptance Criteria
1️⃣ …  
2️⃣ …  
3️⃣ …  

## 9) Consequences
✅ … benefits…  
⚠️ … trade-offs…
' || true

# 1) Accepted ADRs
write_file "$ADIR/ADR-RTGF-001-policy-source-matrix-format.md" "ADR-RTGF-001" "# ADR-RTGF-001: Policy Source Matrix Format

**Status:** Accepted  
**Date:** $accepted_date  
**Decision Makers:** Kevin Brown, Corridor Governance Group  
**Owner:** RTGF Working Group  
**Related Artifacts:** \`docs/architecture/policy-source-matrix.md\`, *draft-lane2-rtgf-00 §3.1*

---

## Context
RTGF requires deterministic, versioned policy inputs per jurisdiction and domain. Regulators publish heterogeneous legal texts, sanctions feeds, and guidance; the compiler needs a uniform schema to ingest them.

## Decision
Create a **Policy Source Matrix** comprised of signed JSON-LD snapshots (\`policy:Snapshot\`) keyed by jurisdiction and domain. Each snapshot **MUST** include:
- Jurisdiction code (ISO 3166-1 alpha-2 or regional)
- Domain identifier aligned with the RTGF domain registry
- Effective/expiry timestamps (RFC 3339 UTC)
- Controls, duties, prohibitions, assurance levels
- Evidence source hashes for external datasets (e.g., sanctions)
- Normative references (legal instruments, guidance)
- **Detached Ed25519** signature by the issuing authority (DID/JWKS)

Snapshots are listed in \`docs/architecture/policy-source-matrix/index.jsonld\` with metadata for automated retrieval.

## Consequences
- Enables deterministic compiler outputs (RMT/IMT) once signatures and hashes are validated.  
- Provides clear provenance chain from legal source to enforcement token.  
- Requires regulators to maintain signing keys and update snapshots when laws or datasets change.

## Links
- *draft-lane2-rtgf-00* Section 3.1  
- \`docs/architecture/policy-source-matrix.md\`
"

write_file "$ADIR/ADR-RTGF-002-sanctions-dataset-hashing-and-evidence-references.md" "ADR-RTGF-002" "# ADR-RTGF-002: Sanctions Dataset Hashing & Evidence References

**Status:** Accepted  
**Date:** $accepted_date  
**Decision Makers:** Kevin Brown, Corridor Governance Group  
**Owner:** RTGF Working Group  
**Related Artifacts:** \`docs/architecture/policy-source-matrix.md\`, \`docs/architecture/rtgf-pipeline.md\`

---

## Context
Sanctions lists change frequently and are sourced from multiple authorities (EU Sanctions Map, UN, OFAC). RTGF must prove which dataset versions were used during compilation without embedding large payloads in RMT/IMT tokens.

## Decision
- During compilation, fetch the latest sanctions datasets (public APIs or signed exports) and compute **SHA-256 digests**.  
- Record digests and dataset metadata in the **policy snapshot** (\`evidence_sources\`).  
- Embed the same digests inside generated **RMTs/IMTs** so verifiers can confirm the compiler used the correct dataset version.  
- Store dataset provenance (source URI, retrieval timestamp, digest) in the **transparency log** (future ADR-RTGF-014).

## Consequences
- Tokens remain compact and deterministic while preserving linkability to external datasets.  
- Auditors can independently cross-check dataset hashes.  
- Requires scheduling and caching infrastructure for dataset acquisition during compilation.

## Links
- \`docs/architecture/policy-source-matrix.md\`  
- \`docs/architecture/rtgf-pipeline.md\`
"

# 2) Skeleton ADRs
create_skeleton() {
  local num="$1"
  local slug="$2"
  local title="$3"
  local file="$ADIR/ADR-RTGF-$num-$slug.md"
  if [[ -f "$file" && $FORCE == false ]]; then
    echo "• Skipping existing ADR-RTGF-$num ($title) — use --force to replace"
    return
  fi
  cat > "$file" <<EOF
# ADR-RTGF-$num: $title — Rev for Acceptance

**Status:** Proposed → *Ready for Acceptance review*  
**Date:** $timestamp  
**Decision Makers:** RTGF Working Group  
**Owner:** RTGF Working Group  
**Target Acceptance:** $target_accept  
**Related ADRs:** ADR-RTGF-001, ADR-RTGF-002

---

## 1) Purpose & Scope
<!-- Describe the problem and scope. Why now? Who benefits? -->

## 2) Architecture / Data Model
<!-- Describe high-level architecture, data flow, canonical schemas. Include diagrams if helpful. -->

## 3) Determinism & Provenance
<!-- JCS canonicalization, build graph pinning, RNG seeds, revEpoch propagation, transparency hooks. -->

## 4) Security & Trust
<!-- mTLS, JWKS/key rotation, signed artefacts, fail-closed behaviour. -->

## 5) Error Taxonomy
| Code | Condition | Action |
|------|----------|--------|
| RTGF_* | … | … |

## 6) Metrics & SLOs
| Metric | Target | Notes |
|--------|--------|------|
|  |  |  |

## 7) Test Plan Mapping
| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| RTGF-CT-XX | … | … |

## 8) Acceptance Criteria
1️⃣ …  
2️⃣ …  
3️⃣ …  

## 9) Consequences
✅ …  
⚠️ …
EOF
  echo "✔  Seeded ADR-RTGF-$num — $title"
}

create_skeleton "003" "compiler-and-deterministic-build-pipeline" "Compiler & Deterministic Build Pipeline"
create_skeleton "004" "token-types-and-canonical-encodings" "Token Types & Canonical Encodings (RMT/IMT/CORT/PSRT)"
create_skeleton "005" "verification-api-and-error-taxonomy" "Verification API & Error Taxonomy"
create_skeleton "006" "jwks-and-trust-model" "JWKS, DANE/TLSA & Trust Model"
create_skeleton "007" "transparency-and-proof-bundles" "Transparency & Proof Bundles for RTGF"
create_skeleton "008" "policy-to-token-mapping-semantics" "Policy Rule Algebra & Token Generation Semantics"
create_skeleton "009" "rtgf-pipeline-and-determinism" "RTGF Compilation Pipeline & Reproducibility"
create_skeleton "010" "operational-slo-and-observability" "Operational SLOs & Observability for RTGF"

# 3) Gap report
echo
echo "=== RTGF ADR Gap Report ==="
expected=(
  "ADR-RTGF-001-policy-source-matrix-format.md"
  "ADR-RTGF-002-sanctions-dataset-hashing-and-evidence-references.md"
  "ADR-RTGF-003-compiler-and-deterministic-build-pipeline.md"
  "ADR-RTGF-004-token-types-and-canonical-encodings.md"
  "ADR-RTGF-005-verification-api-and-error-taxonomy.md"
  "ADR-RTGF-006-jwks-and-trust-model.md"
  "ADR-RTGF-007-transparency-and-proof-bundles.md"
  "ADR-RTGF-008-policy-to-token-mapping-semantics.md"
  "ADR-RTGF-009-rtgf-pipeline-and-determinism.md"
  "ADR-RTGF-010-operational-slo-and-observability.md"
)
missing=()
for name in "${expected[@]}"; do
  if [[ ! -f "$ADIR/$name" ]]; then
    missing+=("$name")
  fi
done

if [[ ${#missing[@]} -eq 0 ]]; then
  echo "No gaps detected. All planned RTGF ADRs present under ${ADIR#$ROOT/}."
else
  echo "Missing ADR files:"
  for name in "${missing[@]}"; do
    echo " - $name"
  done
fi

# 4) Optional git add/commit
if $DO_COMMIT; then
  if git -C "$ROOT" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git -C "$ROOT" add "$ADIR"
    if ! git -C "$ROOT" diff --cached --quiet; then
      git -C "$ROOT" commit -m "chore(rtfg-adr): bootstrap RTGF ADRs and template"
      echo "✔  Committed RTGF ADR bootstrap to git."
    else
      echo "ℹ  No staged changes to commit."
    fi
  else
    echo "ℹ  --commit ignored: $ROOT is not a git repository."
  fi
fi

echo "Done. Review ADR drafts under ${ADIR#$ROOT/} and tailor content as required."
