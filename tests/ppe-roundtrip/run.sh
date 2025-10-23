#!/usr/bin/env bash
set -euo pipefail
SNAP=${1:-examples/policy.snapshot.json}
OUTP=out/ppe
mkdir -p "$OUTP"

# Build-time: compile predicates + plan
node rtgf-compiler/ppe-compiler/cmd/rtgf-ppe.ts \
  --snapshot "$SNAP" \
  --out "$OUTP/predicate_set.json" \
  --plan-out "$OUTP/eval_plan.json"

# Assemble token stub
jq -n --argfile ps "$OUTP/predicate_set.json" --argfile pl "$OUTP/eval_plan.json" '
  {
    "imt_id": "IMT-EU.SG-PAYMENTS_AML-2025-10-22",
    "predicate_set": $ps,
    "eval_plan": $pl,
    "controls_on_permit": ["LOG:L2"],
    "controls_on_permit_with_controls": ["HUMAN_REVIEW"]
  }
' > "$OUTP/token.json"

# Run-time: evaluate with sample context
node -e '
  const fs = require("fs");
  const path = require("path");
  const { evaluate } = require(path.resolve("aarp-core/ppe-evaluator/src/engine.ts"));
  (async () => {
    const token = JSON.parse(fs.readFileSync("out/ppe/token.json", "utf-8"));
    const ctx = JSON.parse(fs.readFileSync("aarp-core/ppe-evaluator/examples/context.sample.json", "utf-8"));
    const res = await evaluate({ token, context: ctx });
    console.log(JSON.stringify(res, null, 2));
  })();
'
