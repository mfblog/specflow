# Unit Verify Context Card

`unit_verify:{unit}` verifies implementation against checked candidate truth and the active plan.

## Required Context

Read only:

1. `framework/core/context_card.md`
2. `framework/core/lifecycle_authority.md`
3. `framework/core/independent_evaluation.md`
4. `docs/specs/_status.md` for the target unit row.
5. `docs/specs/_check_result/unit/{unit}.md`
6. `docs/specs/_plans/active/{unit}.md`
7. `docs/specs/units/candidate/c_unit_{unit}.md`
8. candidate appendices, stable truth, and rule files named by the plan.
9. implementation files and tests named by the active plan, repository mapping, or verification scope.

Before treating check or plan evidence as usable verification input, before making repair writes, or before writing verify evidence, run:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> command preflight --repo-root <repo-root> --command unit_verify --object-type unit --object <unit>
```

If command preflight is unavailable, run explicit `snapshot validate-process` checks for `check` and `plan` before any verification or repair write.

## Allowed Writes

Allowed writes are:

1. `docs/specs/_verify_result/unit/{unit}.md` only for an advancing `ready_to_promote` result with valid independent evaluation receipt.
2. implementation files only when the close outcome is an implementation fallback and the active plan still authorizes repair in the same command session.
3. local test output artifacts when required by the verification method.

Candidate verify evidence must bind to the current active plan through `active_plan_file_ref` and `active_plan_fingerprint`.
It must also record per-item `acceptance_item_evidence_matrix.evidence_refs` and `retirement_evidence_matrix`.
For executable acceptance items, promotion-ready verify evidence requires `status: pass` and durable evidence refs that prove the candidate behavior through the declared verification surface.
If the active plan has `retirement_targets: none`, the matrix must be `none`.
If the active plan lists retirement targets, each target must have `result: pass`, `mainline_dependency: not_required`, and durable `evidence_refs` before `ready_to_promote` may close.
For primary protocol, default page, primary presentation, API, or artifact-generation changes, evidence must inspect the real generated artifact, API return value, DOM/screenshot, rendered text, CLI output, or a test that proves the mainline path uses the candidate protocol.
Generic test success, absence of old strings, presence of new files, or presence of new fields must not be used as the only evidence for semantic replacement.
Verification must not delete code automatically or infer business compatibility safety; it only proves whether planned retirement targets are no longer required by the mainline path.

## Forbidden Writes

Do not write:

1. candidate or stable truth.
2. lifecycle status.
3. check, plan, or stable-verify process evidence.
4. rule truth or global rules.
5. verify evidence that claims promotion readiness when the independent reviewer result is not `pass`.

Truth or gate drift falls back to `unit_check`.
Plan drift falls back to `unit_plan`.
Implementation deviation falls back to `unit_impl`.

## On-Demand Expansions

Enter only when the trigger appears:

1. `framework/governance/rule_system.md` when verification exposes rule ownership or global-rule conflict; use `framework/governance/rules/rule_escape.md` when current truth is insufficient to choose or finish the rule flow safely.
2. `framework/lifecycle/recovery.md` when check, plan, or verification inputs are stale, missing, or internally inconsistent.
3. `framework/operations/migration.md` when existing verify evidence uses an older shape that blocks validation.
4. `framework/operations/implementation_change.md` when a repair request is not already authorized by the exact lifecycle route.
5. `framework/core/freshness.md` when validation reports `freshness_layer` or `text_drift`.

## Independent Evaluation

Advancing outcome `ready_to_promote` requires independent evaluation.

The executor may write verify evidence, but promotion readiness must be reviewed by an isolated reviewer using reviewer pack `unit_verify_ready_to_promote` from `framework/core/independent_evaluation.md`.

Before requesting review, generate the handoff request:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> evaluation request --repo-root <repo-root> --object-type unit --object <unit> --pack unit_verify_ready_to_promote
```

If the current agent runtime explicitly exposes an independent executor capability, send only the generated request file to that executor.
If no such capability is explicitly available, stop and give the user the generated request file path and trigger instruction.
The reviewer returns the result to the executor and must not modify repository files.

`docs/specs/_verify_result/unit/{unit}.md` must contain the independent evaluation receipt defined in `framework/core/independent_evaluation.md`.

## Close Requirements

Before closing `ready_to_promote`, run:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> snapshot validate-process --repo-root <repo-root> --object-type unit --object <unit> --process verify
```

Successful close uses outcome `ready_to_promote` and advances to `unit_promote`.
Do not advance until validation succeeds and this close command accepts the current evidence:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> command close --repo-root <repo-root> --command unit_verify --object-type unit --object <unit> --outcome ready_to_promote --apply
```

Accepted `text_drift` evidence is valid current evidence; unaccepted freshness drift must stop for independent freshness review or evidence recreation.
A prior `evidence_incomplete`, `human_verify`, or fallback result does not prevent later `ready_to_promote` when current verify evidence validates again and carries the required independent reviewer receipt.
