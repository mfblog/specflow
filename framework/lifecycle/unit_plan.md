# Unit Plan Context Card

`unit_plan:{unit}` creates the package-bounded delta implementation handoff from checked candidate truth.

`unit_plan` must plan from the full unit package passed by `unit_check`, but the active plan output is only for this round's delta. Unchanged appendices, rules, stable unit dependencies, and acceptance truth do not need implementation tasks, but they must remain visible as package constraints for the delta.

## Required Context

Read only:

1. `framework/core/context_card.md`
2. `framework/core/lifecycle_authority.md`
3. `framework/core/independent_evaluation.md`
4. `docs/specs/_status.md` for the target unit row.
5. `docs/specs/_check_result/unit/{unit}.md`
6. `docs/specs/units/candidate/c_unit_{unit}.md`
7. candidate appendices, stable truth, and rule files named by the check result.
8. `docs/specs/repository_mapping.md` and implementation surfaces needed to identify current entry points, main paths, rendering/API/generation paths, and gaps.
9. existing `docs/specs/_plans/active/{unit}.md` or `docs/specs/_plans/draft/{unit}.md` only when updating prior planning work.

The check result, not `_check_work`, is the planning gate.

Before treating `docs/specs/_check_result/unit/{unit}.md` as usable planning input, or before writing draft or active plan files, run:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> command preflight --repo-root <repo-root> --command unit_plan --object-type unit --object <unit>
```

If command preflight is unavailable, run `snapshot validate-process --repo-root <repo-root> --object-type unit --object <unit> --process check` explicitly before any planning write.

## Allowed Writes

Allowed writes are:

1. `docs/specs/_plans/active/{unit}.md` when the handoff is ready and independently reviewed.
2. `docs/specs/_plans/draft/{unit}.md` for non-consumable planning notes when blocked.

The active plan must bind to current candidate truth by normalized content fingerprint, cover every accepted acceptance item, and include `stable_candidate_diff_refs`, `implementation_gap_refs`, `planned_change_scope`, `package_constraint_review`, `package_constraint_refs`, `package_constraint_summary`, and `retirement_targets`.
It must also record the current `unit_appendix_snapshot`, `unit_snapshot`, and `rule_snapshot` so downstream implementation and verification can prove they are using the same package basis.
When stable truth exists, `stable_candidate_diff_refs` must cite both the current stable main Spec and the current candidate main Spec.
`implementation_gap_refs` must cite the implementation and mapping refs inspected for the plan, or be literal `none` only when there is no current implementation surface to inspect.
`planned_change_scope` lists only the round's delta scopes. Each item must use an `id` of `pcs.<slug>` and record `basis_refs`, `acceptance_item_ids`, `implementation_refs`, and `verification_action`.
`basis_refs` must cite package refs that constrain the delta.
`package_constraint_review` must be `pass`.
`package_constraint_refs` must cite refs from the current package snapshot that were considered as constraints for the delta.
`package_constraint_summary` must state how the delta remains bounded by the package.
`retirement_targets` must be literal `none`, or list concrete retired paths, helpers, wrappers, compatibility layers, dependencies, or equivalent targets with retirement method, acceptance item ids, and verification action.
For `candidate_intent: change` with `source_basis: replacement`, `retirement_targets` must not be `none`.
Replacement plans must identify old primary paths, new primary paths, cutover slices, and the retirement target ids that prove old primary paths no longer carry mainline behavior.
Planning must not claim a retained compatibility path is retired; if it remains required, it must stay visible in the plan or be deferred by explicit later planning.
Planning must not expand into a whole-package implementation plan merely because the whole package is used as the constraint surface.

## Forbidden Writes

Do not write:

1. candidate or stable truth.
2. implementation files.
3. lifecycle status.
4. repository mapping.
5. check, verify, or stable-verify process evidence.
6. active plan evidence when the independent reviewer result is not `pass`.

If behavior truth is missing or wrong, return to `unit_check` instead of inventing plan content.

## On-Demand Expansions

Enter only when the trigger appears:

1. `framework/operations/entry_routing.md` when the request is not an exact `unit_plan:{unit}` command or the target object is unclear.
2. `framework/governance/rule_system.md` when plan creation exposes missing or conflicting rule ownership; use `framework/governance/rules/rule_escape.md` when current truth is insufficient to choose or finish the rule flow safely.
3. `framework/lifecycle/recovery.md` when required check evidence is missing, stale, or invalid.
4. `framework/operations/migration.md` when existing plan or process evidence uses an older shape that blocks validation.
5. `framework/core/freshness.md` when validation reports `freshness_layer` or `text_drift`.

## Independent Evaluation

Advancing outcome `plan_ready` requires independent evaluation.

The executor may write the active plan, but the handoff gate must be reviewed by an isolated reviewer using reviewer pack `unit_plan_plan_ready` from `framework/core/independent_evaluation.md`.

Before requesting review, generate the handoff request:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> evaluation request --repo-root <repo-root> --object-type unit --object <unit> --pack unit_plan_plan_ready
```

If the current agent runtime explicitly exposes an independent executor capability, send only the generated request file to that executor.
If no such capability is explicitly available, stop and give the user the generated request file path and trigger instruction.
The reviewer returns the result to the executor and must not modify repository files.

`docs/specs/_plans/active/{unit}.md` must contain the independent evaluation receipt defined in `framework/core/independent_evaluation.md`.

## Close Requirements

Outcomes:

| Outcome | Status Result |
|---|---|
| `plan_ready` | Active plan validates; next command is `unit_impl` |
| `truth_fallback` | Candidate truth is incomplete or stale; fallback to `unit_check` |
| `blocked` | Planning is blocked but truth still stands; stay at `unit_plan` |
| `decision_checkpoint` | Implementation direction needs a decision; stay at `unit_plan` unless truth writeback is required |

Before closing `plan_ready`, run:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> snapshot validate-process --repo-root <repo-root> --object-type unit --object <unit> --process plan
```

Do not advance to `unit_impl` until validation succeeds and this close command accepts the current evidence:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> command close --repo-root <repo-root> --command unit_plan --object-type unit --object <unit> --outcome plan_ready --apply
```

Accepted `text_drift` evidence is valid current evidence; unaccepted freshness drift must stop for independent freshness review or evidence recreation.
A prior `blocked` or `decision_checkpoint` result does not prevent later `plan_ready` when current active plan evidence validates again and carries the required independent reviewer receipt.
