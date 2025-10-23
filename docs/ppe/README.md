# Policy Predicate Engine (PPE) — Scope & Design

## Objective
Deterministically decide `PERMIT | PERMIT_WITH_CONTROLS | DENY` for a concrete request by evaluating a **signed predicate set** embedded in RMT/IMT tokens, using a **fixed evaluation plan**, producing a **canonical evaluation trace** for audit.

## Two Components
1. **PPE-Compiler (Build-time)** — lives in `rtgf-compiler/ppe-compiler/`
   - Input: Signed *policy snapshots* (jurisdiction/domain), or derived RMT/IMT.
   - Output: `predicate_set` + `eval_plan` (deterministic order) embedded into RMT/IMT.
   - Guarantees: deterministic compilation, JSON canonicalization (RFC 8785), signed artefacts.

2. **PPE-Evaluator (Run-time)** — lives in `aarp-core/ppe-evaluator/`
   - Input: IMT/RMT’s `predicate_set` + `eval_plan`, request **context** (amount, parties, corridor), and injected **resolvers** (e.g., sanctions check, proof verification).
   - Output: decision, reasons[], controls[], and a canonical **EvalTrace** (hash → Merkle leaf for DOP).
   - Where used: aARP PDP (pre-STA issuance), optionally PEP (last-mile enforcement).

## Determinism Rules
- Fixed operator set; no loops/recursion; no implicit type coercion.
- Static evaluation plan defines **total order** of steps.
- All inputs/outputs canonicalized with **RFC 8785 (JCS)** before hashing/signing.

## Minimal Predicate DSL (subset)
- Comparators: EQUALS, NOT_EQUALS, GT/GTE/LT/LTE, IN, MATCHES_REGEX
- Boolean: AND/OR/NOT
- Numeric: ADD/SUB/MUL/DIV (bounded, currency-safe decimals)
- Set/Map: HAS_KEY, HAS_VALUE, LENGTH, SUBSET_OF
- Special: LOOKUP (resolver), NOW (UTC), AGE_LT_DAYS

## Failure Semantics
- Fail-closed: missing input, unknown operator, resolver error/timeout ⇒ `DENY` with reason.
- Priority: hard constraints override later predicates.

## Evidence
- EvalTrace (token id, plan hash, ordered steps, results, timestamp) → sha256(JCS(trace)) → DOP Merkle leaf.
