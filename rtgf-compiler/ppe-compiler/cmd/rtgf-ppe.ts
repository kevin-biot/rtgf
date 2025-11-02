#!/usr/bin/env node
// Minimal CLI skeleton. TODO: implement deterministic snapshot → predicate compilation.
import { readFileSync, writeFileSync } from "fs";
import { pathToFileURL } from "url";

export interface CompileOptions {
  snapshotPath: string;
  outPred: string;
  outPlan: string;
}

export function compile(options: CompileOptions) {
  const policy = JSON.parse(readFileSync(options.snapshotPath, "utf-8"));
  const predicateSet = {
    predicate_set_id: `ps.${policy.jurisdiction}.${policy.domain}.v1`,
    version: "1.0.0",
    order: policy.required_predicates || [],
    predicates: policy.predicates || []
  };
  const sequence = [{ stage: "P0", op: "VALIDATE_INPUTS" }];
  for (const [idx, id] of (predicateSet.order || []).entries()) {
    sequence.push({ stage: `P${idx + 1}`, predicate: id, ops: ["D", "L"] });
  }
  const plan = {
    eval_plan_id: `plan.${policy.jurisdiction}.${policy.domain}.v1`,
    sequence,
    hash: "sha256:TODO-plan-hash"
  };
  writeFileSync(options.outPred, JSON.stringify(predicateSet, null, 2));
  writeFileSync(options.outPlan, JSON.stringify(plan, null, 2));
  return { predicateSet, plan };
}

function parseArgs(argv: string[]) {
  const get = (flag: string) => {
    const idx = argv.indexOf(flag);
    return idx >= 0 ? argv[idx + 1] : "";
  };
  const snapshotPath = get("--snapshot");
  const outPred = get("--out");
  const outPlan = get("--plan-out");
  if (!snapshotPath || !outPred || !outPlan) {
    throw new Error("Usage: rtgf-ppe compile --snapshot policy.jsonld --out predicate_set.json --plan-out eval_plan.json");
  }
  return { snapshotPath, outPred, outPlan };
}

export function run(argv = process.argv.slice(2)) {
  const opts = parseArgs(argv);
  const { outPred, outPlan } = opts;
  compile(opts);
  console.log("PPE-Compiler outputs written:", outPred, outPlan);
}

const invokedPath = process.argv[1] ? pathToFileURL(process.argv[1]).href : "";
if (import.meta.url === invokedPath) {
  try {
    run();
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    console.error(message);
    process.exit(2);
  }
}
