import { mkdirSync, readFileSync, rmSync, writeFileSync } from "fs";
import { createHash } from "crypto";
import { join, relative, resolve } from "path";
import { fileURLToPath } from "url";
import Ajv from "ajv";
import addFormats from "ajv-formats";
import { compile } from "../../rtgf-compiler/ppe-compiler/cmd/rtgf-ppe";
import { evaluate } from "../../aarp-core/ppe-evaluator/src/engine";

interface HarnessOptions {
  snapshot: string;
  outDir: string;
  manifest: string;
}

interface DigestMap {
  [artifact: string]: string;
}

const DEFAULT_OPTIONS: HarnessOptions = {
  snapshot: "examples/policy.snapshot.json",
  outDir: "out/ppe",
  manifest: "out/ppe/digests.json"
};

const HERE = fileURLToPath(new URL(".", import.meta.url));
const REPO_ROOT = resolve(HERE, "..", "..");

setDeterministicEnv();

async function main() {
  const opts = parseArgs(process.argv.slice(2));
  const snapshotPath = resolvePath(opts.snapshot);
  const outDir = resolvePath(opts.outDir);
  const manifestPath = resolvePath(opts.manifest);

  rmSync(outDir, { recursive: true, force: true });
  mkdirSync(outDir, { recursive: true });

  const ajv = new Ajv({ strict: false, allErrors: true });
  addFormats(ajv);
  const predicateSchema = loadJSON("shared/ppe-schemas/predicate.schema.json");
  const evalPlanSchema = loadJSON("shared/ppe-schemas/eval-plan.schema.json");
  const validatePredicate = ajv.compile(predicateSchema);
  const validatePlan = ajv.compile(evalPlanSchema);

  const runs = [];
  const digests: DigestMap = {};

  for (const iteration of [1, 2]) {
    const runDir = join(outDir, `run-${iteration}`);
    mkdirSync(runDir, { recursive: true });

    const outPred = join(runDir, "predicate_set.json");
    const outPlan = join(runDir, "eval_plan.json");

    compile({
      snapshotPath,
      outPred,
      outPlan
    });

    const predicateSet = loadJSON(outPred);
    const evalPlan = loadJSON(outPlan);

    predicateSet.predicates.forEach((predicate: any, idx: number) => {
      if (!validatePredicate(predicate)) {
        throw new Error(
          `Predicate schema validation failed (index ${idx}): ${ajv.errorsText(validatePredicate.errors)}`
        );
      }
    });
    if (!validatePlan(evalPlan)) {
      throw new Error(`Eval plan schema validation failed: ${ajv.errorsText(validatePlan.errors)}`);
    }

    const token = {
      imt_id: "IMT-EU.SG-PAYMENTS_AML-2025-10-22",
      predicate_set: predicateSet,
      eval_plan: evalPlan,
      controls_on_permit: ["LOG:L2"],
      controls_on_permit_with_controls: ["HUMAN_REVIEW"]
    };
    const tokenPath = join(runDir, "token.json");
    writeFileSync(tokenPath, JSON.stringify(token, null, 2), "utf-8");

    const contextPath = resolvePath("aarp-core/ppe-evaluator/examples/context.sample.json");
    const context = JSON.parse(readFileSync(contextPath, "utf-8"));
    const evaluation = await evaluate({
      token,
      context,
      now: process.env.FIXED_TIME,
      resolvers: {}
    });
    const evaluationPath = join(runDir, "evaluation.json");
    writeFileSync(evaluationPath, JSON.stringify(evaluation, null, 2), "utf-8");

    const currentDigests: DigestMap = {
      predicate_set: digestFile(outPred),
      eval_plan: digestFile(outPlan),
      token: digestFile(tokenPath),
      evaluation: digestFile(evaluationPath)
    };

    runs.push({ runDir, digests: currentDigests, evaluation });
    if (iteration === 1) {
      Object.assign(digests, currentDigests);
      copyArtifacts(runDir, outDir);
    } else {
      ensureMatch(digests, currentDigests);
    }
  }

  const manifest = {
    snapshot: relative(process.cwd(), snapshotPath),
    generatedAt: new Date().toISOString(),
    environment: {
      TZ: process.env.TZ,
      LANG: process.env.LANG,
      FIXED_TIME: process.env.FIXED_TIME,
      RTGF_SEED: process.env.RTGF_SEED
    },
    artifacts: digests
  };
  writeFileSync(manifestPath, JSON.stringify(manifest, null, 2), "utf-8");
  console.log(`Determinism manifest written to ${relative(process.cwd(), manifestPath)}`);
}

function parseArgs(argv: string[]): HarnessOptions {
  const opts: HarnessOptions = { ...DEFAULT_OPTIONS };
  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];
    switch (arg) {
      case "--snapshot":
        opts.snapshot = argv[++i] ?? opts.snapshot;
        break;
      case "--out":
        opts.outDir = argv[++i] ?? opts.outDir;
        break;
      case "--manifest":
        opts.manifest = argv[++i] ?? opts.manifest;
        break;
      default:
        throw new Error(`Unknown argument ${arg}`);
    }
  }
  return opts;
}

function setDeterministicEnv() {
  process.env.TZ = process.env.TZ || "UTC";
  process.env.LANG = process.env.LANG || "C";
  process.env.FIXED_TIME = process.env.FIXED_TIME || "2025-01-01T00:00:00Z";
  process.env.RTGF_SEED = process.env.RTGF_SEED || "1337";
  seedRandom(Number(process.env.RTGF_SEED));
}

function seedRandom(seed: number) {
  if (!Number.isFinite(seed)) {
    return;
  }
  let state = seed >>> 0;
  const modulus = 0xffffffff;
  const multiplier = 1664525;
  const increment = 1013904223;
  Math.random = () => {
    state = (state * multiplier + increment) % modulus;
    return state / modulus;
  };
}

function loadJSON(pathRelativeToHarness: string) {
  const absolute = resolvePath(pathRelativeToHarness);
  return JSON.parse(readFileSync(absolute, "utf-8"));
}

function digestFile(pathToFile: string): string {
  const data = readFileSync(pathToFile);
  return `sha256:${createHash("sha256").update(data).digest("hex")}`;
}

function ensureMatch(first: DigestMap, second: DigestMap) {
  for (const key of Object.keys(first)) {
    if (first[key] !== second[key]) {
      throw new Error(`Digest mismatch for ${key}: ${first[key]} !== ${second[key]}`);
    }
  }
  for (const key of Object.keys(second)) {
    if (!(key in first)) {
      throw new Error(`Unexpected artifact digest in repeat run: ${key}`);
    }
  }
}

function copyArtifacts(runDir: string, outDir: string) {
  const artifacts = ["predicate_set.json", "eval_plan.json", "token.json", "evaluation.json"];
  for (const name of artifacts) {
    const src = join(runDir, name);
    const dest = join(outDir, name);
    writeFileSync(dest, readFileSync(src));
  }
}

function resolvePath(p: string): string {
  return resolve(REPO_ROOT, p);
}

main().catch((err) => {
  console.error(err instanceof Error ? err.message : err);
  process.exit(1);
});
