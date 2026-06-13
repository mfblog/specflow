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
3. `framework/lifecycle/overview.md`
4. the lifecycle Context Card for the unit's current `Next Command`

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
2. `unit_verify:{unit}`

It must stop when the next command is:

1. `unit_new`
2. `unit_init`
3. `unit_promote`
4. `unit_fork`
5. `unit_stable_verify`

Those entries require an explicit user decision.
`unit_promote:{unit}` remains the explicit manual entry for landing a verified candidate as stable truth.

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

## 5. Recovery Path

After any stop in Section 4, the executor must:
1. Report the blocking condition to the user in plain language, including the target unit and the specific stop condition that triggered.
2. If process evidence was invalidated, apply `framework/lifecycle/recovery.md` before any reroute.
3. Return to `framework/operations/entry_routing.md` unless the stop condition names an explicit resume owner or the user provides a direct next action.

The stop-report must include:
- the target unit and stop condition reason
- what the user must resolve before advance can continue
- where the executor resumes after resolution

## 6. Completion

Automatic advance is complete at the promotion-ready stop when the unit row records:

1. `Candidate=yes`
2. `Active Layer=candidate`
3. `Next Command=unit_promote`

The executor must report the promotion-ready status and state that `unit_promote:{unit}` requires an explicit user decision.
Stable completion still requires the separate `unit_promote:{unit}` command to close successfully and then record `Stable=yes`, `Candidate=no`, `Active Layer=stable`, and `Next Command=unit_fork`.
