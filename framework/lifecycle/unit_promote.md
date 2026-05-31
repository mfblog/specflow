# Unit Promote Context Card

`unit_promote:{unit}` lands verified candidate truth as stable truth.

## Required Context

Read only:

1. `framework/core/context_card.md`
2. `framework/core/lifecycle_authority.md`
3. `docs/specs/_status.md` for the target unit row.
4. `docs/specs/_plans/active/{unit}.md`
5. `docs/specs/_verify_result/unit/{unit}.md`
6. `docs/specs/units/candidate/c_unit_{unit}.md`
7. candidate appendices, stable truth, unit refs, and rule files named by the active plan or verify result.
8. `docs/specs/repository_mapping.md` only when promotion retargeting or cleanup needs ownership confirmation.

Before promotion judgment, stable writeback, unit or rule ref retargeting, cleanup, or command close, run:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> command preflight --repo-root <repo-root> --command unit_promote --object-type unit --object <unit>
```

If command preflight is unavailable, run both `snapshot validate-process --repo-root <repo-root> --object-type unit --object <unit> --process plan` and `snapshot validate-process --repo-root <repo-root> --object-type unit --object <unit> --process verify` explicitly before any promotion write.

## Allowed Writes

Allowed writes are:

1. `docs/specs/units/stable/s_unit_{unit}.md` and explicitly verified stable appendices produced from the verified candidate.
2. deterministic unit refs and rule refs required to retarget the promoted stable truth.
3. lifecycle status, stable promotion summary, and candidate/process cleanup only through successful `command close --command unit_promote --object-type unit --object <unit> --outcome promoted --apply`.

The close command writes `docs/specs/_verify_result/stable/unit/{unit}.md` as the stable promotion summary before it deletes candidate verify evidence or candidate truth.
The stable promotion summary is written before candidate verify evidence is cleaned up.
Candidate cleanup depends on stable promotion summary writeback.
Keep current candidate truth, candidate appendices, and `unit_verify` evidence in place until the close command starts.

## Forbidden Writes

Do not write:

1. implementation files.
2. unverified stable truth or appendices.
3. new behavior, acceptance, ownership, or rule meaning not present in the verified candidate.
4. check, plan, verify, or stable-verify process evidence.
5. lifecycle status by hand.
6. rule truth or global rules unless the change is deterministic retargeting of already verified refs.

Truth or gate invalidation falls back to `unit_check`.
Plan invalidation falls back to `unit_plan`.
Evidence invalidation falls back to `unit_verify`.

## On-Demand Expansions

Enter only when the trigger appears:

1. `framework/governance/rule_system.md` when promotion exposes rule ownership or global-rule conflicts; use `framework/governance/rules/rule_escape.md` when current truth is insufficient to choose or finish the rule flow safely.
2. `framework/lifecycle/recovery.md` when verify evidence, candidate truth, retargeting inputs, or cleanup state are stale, missing, or internally inconsistent.
3. `framework/operations/migration.md` when stable writeback or process evidence uses an older shape that blocks validation.
4. `framework/core/freshness.md` when preflight or verify validation reports `freshness_layer` or `text_drift`.

## Independent Evaluation

`unit_promote` does not require a new independent reviewer receipt.

Promotion consumes the current active plan plus verified evidence from `unit_verify`, including the independent evaluation receipt in `docs/specs/_verify_result/unit/{unit}.md`.
The verify evidence must bind to the current active plan and complete every active plan retirement target with `result: pass` and `mainline_dependency: not_required`.

## Close Requirements

Successful close uses outcome `promoted`, sets `Active Layer=stable`, clears candidate state, and sets `Next Command=unit_fork`.

Before closing `promoted`, ensure the command preflight above has succeeded or both plan and verify evidence have been explicitly validated.
Do not close until stable writeback and deterministic ref retargeting are complete.
Then this close command validates current verify evidence, writes the stable promotion summary, advances lifecycle status, and runs candidate/process cleanup:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> command close --repo-root <repo-root> --command unit_promote --object-type unit --object <unit> --outcome promoted --apply
```

Do not delete candidate verify evidence, candidate truth, candidate appendices, or promotion process files before this close command finishes.

If close fails before `status_updated: true`, keep the unit on `unit_promote` and resolve the failed input through `framework/lifecycle/recovery.md` before retrying close.
If close reports `status_updated: true` and a success-cleanup failure, the promotion state is already stable; fix the filesystem blocker and rerun only:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> process cleanup-success --repo-root <repo-root> --object-type unit --object <unit> --mode unit_promote
```

Do not rerun `unit_promote` close after `status_updated: true`, because `_status.md` already points to `unit_fork`.

Accepted `text_drift` verify evidence is current valid evidence; unaccepted freshness drift must stop before promotion.
