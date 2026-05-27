# Unit Stable Verify Context Card

`unit_stable_verify:{unit}` checks whether current implementation still matches stable truth.

## Required Context

Read only:

1. `framework/core/context_card.md`
2. `framework/core/lifecycle_authority.md`
3. `framework/core/independent_evaluation.md`
4. `docs/specs/_status.md` for the target unit row.
5. `docs/specs/units/stable/s_unit_{unit}.md`
6. stable appendices and rule files referenced by the unit.
7. `docs/specs/repository_mapping.md` entries for the unit.
8. current implementation files and tests named by repository mapping or stable truth.
9. existing `docs/specs/_stable_verify_result/unit/{unit}.md` only when updating prior stable verification evidence.

## Allowed Writes

Allowed writes are:

1. `docs/specs/_stable_verify_result/unit/{unit}.md` for current stable implementation-alignment evidence with valid independent evaluation receipt.
2. local test output artifacts required by the verification method.

The stable verify result must include stable truth refs and fingerprint, repository mapping snapshot, unit/rule/appendix snapshots, acceptance item set, acceptance item evidence matrix, implementation surface refs, evidence refs, and the independent evaluation receipt.

## Forbidden Writes

Do not write:

1. stable truth.
2. candidate truth.
3. implementation files.
4. lifecycle status.
5. rule truth or global rules.
6. stable verify evidence that claims `aligned`, `controlled_repair_required`, or `controlled_change_required` when the independent reviewer result is not `pass`.

Stable verification does not create candidate truth by itself.

## On-Demand Expansions

Enter only when the trigger appears:

1. `framework/governance/rule_system.md` when verification exposes rule ownership or global-rule conflict; use `framework/governance/rules/rule_escape.md` when current truth is insufficient to choose or finish the rule flow safely.
2. `framework/lifecycle/recovery.md` when stable truth, mapping, or evidence inputs are stale, missing, or internally inconsistent.
3. `framework/operations/migration.md` when existing stable verify evidence uses an older shape that blocks validation.
4. `framework/lifecycle/unit_init_new_fork.md` when controlled repair or controlled change requires new candidate truth after close.
5. `framework/core/freshness.md` when validation reports `freshness_layer` or `text_drift`.

## Independent Evaluation

Advancing outcomes `aligned`, `controlled_repair_required`, and `controlled_change_required` require independent evaluation.

The executor may write stable verify evidence, but stable alignment or controlled-change readiness must be reviewed by an isolated reviewer using reviewer pack `unit_stable_verify_advancing` from `framework/core/independent_evaluation.md`.

`docs/specs/_stable_verify_result/unit/{unit}.md` must contain the independent evaluation receipt defined in `framework/core/independent_evaluation.md`.

## Close Requirements

Outcomes:

| Outcome | Status Result |
|---|---|
| `aligned` | Valid stable verify evidence proves stable truth and implementation align; next command is `unit_fork` |
| `small_repair_required` | Stay at `unit_stable_verify` |
| `evidence_incomplete` | Stay at `unit_stable_verify` |
| `truth_rejudge_required` | Stay at `unit_stable_verify` |
| `controlled_repair_required` | Valid matching stable verify evidence exists; next command is `unit_fork` with repair intent |
| `controlled_change_required` | Valid matching stable verify evidence exists; next command is `unit_fork` with change intent |

Before claiming `aligned`, `controlled_repair_required`, or `controlled_change_required`, run:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> snapshot validate-process --repo-root <repo-root> --object-type unit --object <unit> --process stable_verify
```

For `aligned`, every executable acceptance item must have evidence status `pass`.
Items marked `not_runnable_yet: yes` in stable truth must use evidence status `not_runnable_yet`.

Do not advance until validation succeeds and the close command for the selected outcome accepts the evidence:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> command close --repo-root <repo-root> --command unit_stable_verify --object-type unit --object <unit> --outcome <outcome> --apply
```

Accepted `text_drift` evidence is valid current evidence; unaccepted freshness drift must stop for independent freshness review or evidence recreation.
A prior non-advancing stable verification result does not prevent later advancement when current stable verify evidence validates again, carries the required independent reviewer receipt, and matches the command close outcome.
