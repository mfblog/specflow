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
5. `unit_impl` (auto state, not a user command; set by `unit_check pass` close)
6. `unit_verify`
7. `unit_promote`
8. `unit_stable_verify`

## Notes Field — Write Constraints

The `Notes` field may carry a `constraints:` prefix to define write-permission boundaries for the current lifecycle phase. The optional format is:

```text
constraints:phase=<phase> [deny=<glob>] [allow=<glob>]
```

- `phase`: current lifecycle phase name (e.g. `unit_impl`, `unit_verify`)
- `deny`: file glob pattern that the executor must not write in this phase
- `allow`: file glob pattern that the executor may write in this phase

Multiple constraints may be separated by `;`. When no `constraints:` prefix exists, no tool-enforced write boundary is active.

The deterministic tooling entry is `specflowctl validate write --path <path> --phase <phase>`.

## Update Rules

Lifecycle advancement is valid only when `specflowctl command close` succeeds. Manual edits to `_status.md` are not substitutes for command close.

Truth and gate fallback return to candidate truth repair. Candidate verify evidence fallback returns to `unit_verify`. Stable verify evidence fallback returns to `unit_stable_verify`.
