# Unit Check Context Card

`unit_check:{unit}` decides whether candidate truth is clear enough to become planning input.

## Required Context

Read only:

1. `framework/core/context_card.md`
2. `framework/core/lifecycle_authority.md`
3. `framework/core/independent_evaluation.md`
4. `docs/specs/_status.md` for the target unit row.
5. `docs/specs/units/candidate/c_unit_{unit}.md`
6. candidate appendices explicitly referenced by the target unit.
7. stable unit truth explicitly referenced by the target unit.
8. rule files explicitly bound by the target unit or its referenced truth.
9. `docs/specs/repository_mapping.md` only when ownership or boundary mapping is part of the check.

`unit_check` validates candidate truth, not implementation strategy. It answers:

1. Is the unit goal and responsibility clear?
2. Are dependencies, rule bindings, and ownership boundaries explicit?
3. Are main flow, data, protocol, state, errors, and acceptance criteria complete enough for planning?
4. Can `unit_plan` proceed without inventing missing behavior, boundary, or acceptance truth?

## Allowed Writes

Allowed writes are:

1. `docs/specs/_check_work/unit/{unit}.md` as an optional resume aid.
2. `docs/specs/_check_result/unit/{unit}.md` only for an advancing `pass` result with valid independent evaluation receipt.
3. `docs/specs/units/candidate/c_unit_{unit}.md` and explicitly referenced candidate appendices only when the outcome is `fix_required` and the repair stays inside the current unit truth.

## Forbidden Writes

Do not write:

1. implementation files.
2. stable truth.
3. repository mapping unless ownership repair has been routed by an on-demand expansion.
4. lifecycle status.
5. rule truth or global rules unless rule governance is explicitly triggered.
6. `_check_result` as pass evidence when the reviewer result is not `pass`.

`_check_work` is not downstream gate evidence and must not be consumed by `unit_plan`.

## On-Demand Expansions

Enter only when the trigger appears:

1. `framework/operations/entry_routing.md` when the request is not an exact `unit_check:{unit}` command or the target object is unclear.
2. `framework/governance/rule_system.md` when rule ownership, rule truth, or repository-wide defaults must change; use `framework/governance/rules/rule_escape.md` when current truth is insufficient to choose or finish the rule flow safely.
3. `framework/lifecycle/recovery.md` when required evidence is missing, stale, or internally inconsistent.
4. `framework/operations/migration.md` when existing files use an older process shape that blocks deterministic validation.
5. `framework/core/freshness.md` when validation reports `freshness_layer` or `text_drift`.

## Independent Evaluation

Advancing outcome `pass` requires independent evaluation.

The executor may draft or update `_check_result`, but the pass gate must be reviewed by an isolated reviewer using reviewer pack `unit_check_pass` from `framework/core/independent_evaluation.md`.

`docs/specs/_check_result/unit/{unit}.md` must contain the independent evaluation receipt defined in `framework/core/independent_evaluation.md`.

## Close Requirements

Outcomes:

| Outcome | Status Result |
|---|---|
| `pass` | Valid `_check_result` exists; next command is `unit_plan` |
| `blocked` | Stay at `unit_check`; user, rule, ownership, or prerequisite input is missing |
| `fix_required` | Stay at `unit_check`; candidate truth can be repaired before rerun |
| `checkpoint` | Stay at `unit_check`; ask for the smallest missing decision in plain language |

Before closing `pass`, run:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> snapshot validate-process --repo-root <repo-root> --object-type unit --object <unit> --process check
```

Do not advance to `unit_plan` until validation succeeds and this close command accepts the current evidence:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> command close --repo-root <repo-root> --command unit_check --object-type unit --object <unit> --outcome pass --apply
```

Accepted `text_drift` evidence is valid current evidence; unaccepted freshness drift must stop for independent freshness review or evidence recreation.
A prior `blocked`, `fix_required`, or `checkpoint` result does not prevent later `pass` when current evidence validates again and carries the required independent reviewer receipt.
