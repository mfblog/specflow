# Advance Policy

This file governs automatic progression for one current unit.

The only supported advance entry is:

```text
unit_advance:{unit}
```

`scenario_advance:{id}` is unsupported and must be rejected.

## 1. Entry

`unit_advance:{unit}` may run only when `docs/specs/_status.md` contains a formal row for that unit.

The row must have `Object Type=unit`.

The executor must read:

1. `docs/specs/_status.md`
2. the current-layer unit main Spec
3. `specflow/framework/command_policy.md`
4. the command file for the unit's current `Next Command`

When the target row has `Candidate=yes` and `Active Layer=candidate`, the executor must run the computed candidate relation preflight before entering the next command:

```text
specflowctl relation candidate-preflight --object {unit}
```

If the target is not in the current ready set, the executor must stop.
It must report the target unit, the ready candidates, the blocking candidate units or candidate Rules, any candidate cycle, and the source files reported by the relation calculation.
It must not silently advance a different candidate unit.

## 2. Allowed Progression

`unit_advance:{unit}` may automatically enter only these commands:

1. `unit_check:{unit}`
2. `unit_plan:{unit}`
3. `unit_impl:{unit}`
4. `unit_verify:{unit}`
5. `unit_promote:{unit}`

It must stop when the next command is:

1. `unit_new`
2. `unit_init`
3. `unit_fork`
4. `unit_stable_verify`

Those entries require an explicit user decision.

## 3. Recursion

Automatic advance may follow unit dependencies only when a command result explicitly identifies a blocking unit.

The executor must not infer dependency work from directory shape or from informal prose. It may continue into another unit only when:

1. the blocking unit is listed in `_status.md`
2. the current command result names it as the next required unit
3. entering that unit does not create a cycle

The executor must stop on any cycle.

## 4. Stop Conditions

The executor must stop when:

1. the target unit row is missing
2. the target row is not a unit row
3. current truth is insufficient to choose the next command
4. required process evidence is missing
5. rule, repository mapping, or global baseline truth must be changed first
6. a human decision is required
7. a referenced stable unit is outdated and the dependent unit must be revalidated
8. candidate relation preflight says the target is blocked by another current candidate, a candidate Rule, or a candidate progression cycle

## 5. Completion

Advance is complete when the unit row records:

1. `Stable=yes`
2. `Candidate=no`
3. `Active Layer=stable`
4. `Next Command=unit_fork`

The executor must report the final status and any remaining downstream reroutes caused by `unit_refs` or `rule_refs`.
