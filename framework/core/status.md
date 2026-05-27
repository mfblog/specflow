# Status Tracking

`docs/specs/_status.md` records the current lifecycle state for formal units.

## Table

The status table uses this header:

```text
| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |
```

- `Object Type` must be `unit`.
- `Active Layer` is `stable` or `candidate`.
- `Next Command` is the only legal lifecycle command to run next.

## Valid Next Commands

The active unit lifecycle commands are:

1. `unit_init`
2. `unit_new`
3. `unit_fork`
4. `unit_check`
5. `unit_plan`
6. `unit_impl`
7. `unit_verify`
8. `unit_promote`
9. `unit_stable_verify`

## Update Rules

Lifecycle advancement is valid only when `specflowctl command close` succeeds. Manual edits to `_status.md` are not substitutes for command close.

Truth and gate fallback return to `unit_check`. Plan fallback returns to `unit_plan`. Implementation fallback returns to `unit_impl`. Candidate verify evidence fallback returns to `unit_verify`. Stable verify evidence fallback returns to `unit_stable_verify`.
