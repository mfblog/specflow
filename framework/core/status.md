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
5. `unit_verify`
6. `unit_promote`
7. `unit_stable_verify`

## Notes Field — Write Constraints

The `Notes` field may carry a `constraints:` prefix to define write-permission boundaries for the current lifecycle phase.

When no `constraints:` prefix exists, no tool-enforced write boundary is active.

The deterministic tooling entry is `specflowctl validate write --path <path> --phase <phase>`.

### Compact Inline Format

```text
constraints:phase=<phase> deny=<glob> [allow=<glob>];phase=<phase> deny=<glob> [allow=<glob>]
```

- `phase`: current lifecycle phase name (e.g. `pending_impl`, `unit_verify`)
- `deny`: file glob pattern that the executor must not write in this phase
- `allow`: file glob pattern that the executor may write in this phase (optional)

Multiple constraint groups may be separated by `;`.
When both `deny` and `allow` are specified within the same group, `deny` takes precedence.

Example (single-line Notes value):

```text
constraints:phase=pending_impl deny=docs/specs/** allow=src/my_feature/**
```

### YAML-like Block Format

When the Notes field contains multiple lines, the constraints may use a YAML-like block structure:

```text
constraints:allowed_writes:
  - pattern: "src/my_feature/**"
    phases: [pending_impl, unit_verify]
  - pattern: "tests/my_feature/**"
forbidden_writes:
  - pattern: "docs/specs/units/stable/**"
  - pattern: "docs/specs/_status.md"
```

- `allowed_writes:` defines patterns that the executor may write
- `forbidden_writes:` defines patterns the executor must not write (takes precedence)
- Each `- pattern:` specifies a glob pattern
- `phases:` is an optional list of lifecycle phases the rule applies to; when absent, the rule applies to all phases

## Notes Field — Lifecycle Phase

The `Notes` field may carry a lifecycle phase value to indicate the unit's current activity within a `Next Command`:

- `pending_impl` — unit_check has passed; implementation has not started or is in progress. `Next Command` is `unit_verify`.

This value is informational only. It does not affect routing or `command close` validation.

## Update Rules

Lifecycle advancement is valid only when `specflowctl command close` succeeds. Manual edits to `_status.md` are not substitutes for command close.

Truth and gate fallback return to candidate truth repair. Candidate verify evidence fallback returns to `unit_verify`. Stable verify evidence fallback returns to `unit_stable_verify`.
