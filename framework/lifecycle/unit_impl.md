# Unit Implementation Context Card

`unit_impl:{unit}` changes implementation files according to the active plan.

## Required Context

Read only:

1. `framework/core/context_card.md`
2. `framework/core/lifecycle_authority.md`
3. `docs/specs/_status.md` for the target unit row.
4. `docs/specs/_check_result/unit/{unit}.md`
5. `docs/specs/_plans/active/{unit}.md`
6. `docs/specs/units/candidate/c_unit_{unit}.md`
7. candidate appendices, stable truth, and rule files named by the active plan.
8. `docs/specs/repository_mapping.md` entries for implementation paths touched by the plan.
9. existing implementation files explicitly named by the plan or repository mapping.

Before treating `docs/specs/_check_result/unit/{unit}.md` or `docs/specs/_plans/active/{unit}.md` as usable implementation input, or before writing implementation files, local fixtures, support files, or repository mapping, run:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> command preflight --repo-root <repo-root> --command unit_impl --object-type unit --object <unit>
```

If command preflight is unavailable, run explicit `snapshot validate-process` checks for `check` and `plan` before any implementation write.

## Allowed Writes

Allowed writes are:

1. implementation files named by the active plan and repository mapping.
2. local test fixtures or support files required by the active plan.
3. `docs/specs/repository_mapping.md` only when the active plan requires adding the implementation paths being touched and no ownership decision is invented.

## Forbidden Writes

Do not write:

1. candidate or stable truth.
2. lifecycle status.
3. check, plan, verify, or stable-verify process evidence.
4. rule truth or global rules.
5. implementation behavior not already covered by candidate truth and active plan.

If behavior truth is missing or wrong, stop and fall back to `unit_check`.
If the plan is incomplete but truth stands, fall back to `unit_plan`.

## On-Demand Expansions

Enter only when the trigger appears:

1. `framework/operations/implementation_change.md` when the request includes implementation changes outside an exact `unit_impl:{unit}` route.
2. `framework/governance/rule_system.md` when implementation exposes rule ownership or global-rule conflicts; use `framework/governance/rules/rule_escape.md` when current truth is insufficient to choose or finish the rule flow safely.
3. `framework/lifecycle/recovery.md` when the active plan, check result, or repository mapping is stale or missing.
4. `framework/operations/migration.md` when implementation paths cannot be mapped with the current repository mapping shape.
5. `framework/core/freshness.md` when preflight reports `freshness_layer` or accepted `text_drift` on required input evidence.

## Independent Evaluation

`unit_impl` does not require an independent reviewer receipt for `ready_for_verify`.

This command changes code; it does not approve that code. Verification and promotion readiness are independently evaluated in `unit_verify`.

## Close Requirements

Successful close uses outcome `ready_for_verify` and advances to `unit_verify`.

Do not close `ready_for_verify` until implementation work is complete, local evidence named by the active plan has been run or explicitly bounded, and this close command accepts the lifecycle transition:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> command close --repo-root <repo-root> --command unit_impl --object-type unit --object <unit> --outcome ready_for_verify --apply
```

Accepted `text_drift` input evidence can be consumed by preflight; unaccepted freshness drift must stop before implementation close.
`ready_for_verify` does not require an independent reviewer receipt because verification, not implementation, approves correctness.
