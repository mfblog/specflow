# Lifecycle Overview

## Lifecycle Sequence

When formal unit governance is needed, the standard lifecycle is:

```
unit_new / unit_fork → unit_check → unit_impl → unit_verify → unit_promote
```

- `unit_check` is a quality gate that validates whether candidate truth is clear enough. It can be
  re-run as a spec re-validation when candidate truth changes during the implementation phase
  (`Next Command=unit_verify`, `Notes=pending_impl`) — see `unit_check.md` precondition exception.
- `unit_impl` is a non-command phase between check and verify — the agent implements the candidate truth
- `unit_verify` verifies whether the implementation satisfies the candidate truth
- `unit_promote` promotes the verified candidate truth to stable truth

## Entry Method

`entry_routing.md` decides which lifecycle path a natural-language request should follow.
Both exact command matching (`command:{unit}`) and natural language are supported.

## Entry Commands

| Command | Purpose |
|---------|---------|
| `unit_init:{unit}` | Existing capability → first stable truth |
| `unit_new:{unit}` | Brand new → first candidate truth |
| `unit_fork:{unit}` | Stable truth → candidate change round |
| `unit_check:{unit}` | Candidate truth quality check (re-runnable during implementation for spec re-validation) |
| `unit_verify:{unit}` | Verify implementation vs candidate truth |
| `unit_promote:{unit}` | Candidate truth → stable truth |
| `unit_stable_verify:{unit}` | Check implementation vs stable truth |

`unit_impl:{unit}` is a trigger command — it provides implementation context to the agent without changing lifecycle state. It is valid when `Next Command=unit_verify`. After implementation, run `unit_verify:{unit}`. There is no `command close` for `unit_impl`. If spec issues are found during implementation, fix the spec and re-run `unit_check:{unit}` for re-validation (see `unit_check.md`).

## Command Execution Rules

- `command close` is the only operation that can advance lifecycle state
- Advancing evidence (`unit_check pass`, `unit_verify ready_to_promote`, `unit_stable_verify advancing`) requires an independent review receipt
- `unit_promote` consumes already-verified evidence and does not need a new independent review
- Non-advancing results (blocked, fix_required, evidence_incomplete) do not block subsequent correct evidence from advancing

## Dependency Management

- A candidate unit may depend on current stable-layer unit versions or current candidate truth (when the Context Card allows it)
- Stable-layer promotion must not silently change the stable versions consumed by other units
- When a stable version changes, `governance/impact_sync.md` must be run to check the impact

## Rule Consumption

- Global rules automatically apply to all current-layer units
- Bound rules apply only when the unit explicitly lists them in `rule_refs`
- Rule changes are managed through `framework/governance/rule_system.md`

## Lifecycle State

`docs/specs/_status.md` records each unit's current state (layer, Next Command).
Only `command close` may modify this file.

## Context Card Layout (Framework Designer Reference)

Each lifecycle Context Card contains the following sections:

1. **Input** — files that must be read for this step
2. **What This Step Does** — the goal and execution content of the current command
3. **Not Allowed** — hard boundaries
4. **How to End** — pass/fail/blocked results and the next step

**Exception — `unit_impl`:** As a trigger command that does not change lifecycle state, produce process evidence, or execute a `command close`, `unit_impl` may adjust this layout: its Input section may be preceded by a Condition paragraph (pre-execution prerequisite), and its How to End section uses prose instead of an outcome table (there is only one terminal outcome with no command close). The four standard section labels remain the same for discoverability.

`framework/...` paths in a Context Card are relative to the framework root:
- Installed project: `framework/...` → `specflow/framework/...`
- Source repository: `framework/...` → `framework/...`

## Lifecycle Permission Rules

`command close` is the only way to advance lifecycle state.

Advancing evidence is valid only when:
1. The current Context Card allows that evidence to be written
2. The current process file passes the corresponding `snapshot validate-process` check
3. The process file contains a valid independent review receipt (when required by the Context Card)
4. `command close` accepts the result and evidence

Valid input evidence is consumptive — only files that pass deterministic validation may be consumed.

Non-advancing results (blocked, fix_required, checkpoint) do not permanently disqualify subsequent work. After repair, the current evidence may still advance as long as it passes validation and carries an independent review receipt.
