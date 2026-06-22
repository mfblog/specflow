# Status Tracking

`docs/specs/_status.md` records the current lifecycle state for formal units.

## Table

The status table uses this header:

```text
| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |
```

- `Object Type` must be `unit`.
- `Stable` and `Candidate` columns accept `yes` or `no`.
- `Active Layer` is `stable` or `candidate`.
- `Next Command` is the legal lifecycle command(s) to run next. A single value for most lifecycle states; a comma-separated set (`unit_check, unit_impl, unit_verify`) during the implementation phase (`[implementation]`).

Legal (Stable, Candidate, Active Layer) combinations:
| Stable | Candidate | Active Layer | Meaning |
|--------|-----------|-------------|---------|
| `yes` | `no` | `stable` | Pure stable unit (no active candidate round) |
| `no` | `yes` | `candidate` | New candidate unit (not derived from a stable unit) |
| `yes` | `yes` | `candidate` | Candidate derived from a stable unit (via `unit_fork`) |

All other combinations are illegal and must be rejected.

## Valid Next Commands

The active unit lifecycle commands are:

1. `unit_init`
2. `unit_new`
3. `unit_fork`
4. `unit_check`
5. `unit_impl`
6. `unit_verify`
7. `unit_promote`
8. `unit_stable_verify`

For units with Active Layer=stable, `unit_stable_verify` may also be selected as a check command regardless of the recorded Next Command, provided that Next Command is not `unit_promote`. This is the "allows" semantics referenced by the unit_stable_verify Context Card precondition.

## Notes Field — Write Constraints

The `Notes` field may carry a `constraints:` prefix to define write-permission boundaries for the current lifecycle phase.

When no `constraints:` prefix exists, no tool-enforced write boundary is active.

The deterministic tooling entry is `specflowctl validate write --path <path> --phase <phase>`.

### Compact Inline Format

```text
constraints:phase=<phase> deny=<glob> [allow=<glob>];phase=<phase> deny=<glob> [allow=<glob>]
```

- `phase`: current lifecycle phase name (e.g. `implementation`, `unit_verify`)
- `deny`: file glob pattern that the executor must not write in this phase
- `allow`: file glob pattern that the executor may write in this phase (optional)

Multiple constraint groups may be separated by `;`.
When both `deny` and `allow` are specified within the same group, `allow` patterns define exceptions to `deny` patterns. If a path matches both a `deny` pattern and an `allow` pattern, the path is permitted — `allow` carves an exception from the `deny` scope.

Example (single-line Notes value):

```text
constraints:phase=implementation deny=docs/specs/** allow=src/my_feature/**
```

### YAML-like Block Format

When the Notes field contains multiple lines, the constraints may use a YAML-like block structure:

```text
constraints:allowed_writes:
  - pattern: "src/my_feature/**"
    phases: [implementation, unit_verify]
  - pattern: "tests/my_feature/**"
forbidden_writes:
  - pattern: "docs/specs/units/stable/**"
  - pattern: "docs/specs/_status.md"
```

- `allowed_writes:` defines patterns that the executor may write
- `forbidden_writes:` defines patterns the executor must not write (takes precedence)
- Each `- pattern:` specifies a glob pattern
- `phases:` is an optional list of lifecycle phases the rule applies to; when absent, the rule applies to all phases

## Notes Field — Implementation Phase

The implementation phase does not use a Notes keyword. Instead, it is indicated by the `Next Command` field: during the implementation phase, `Next Command` contains `unit_impl` as one of its comma-separated values (e.g. `unit_check, unit_impl, unit_verify`). The `[implementation]` label in the lifecycle flow diagram marks this phase. See `framework/lifecycle/overview.md` for the lifecycle flow and `framework/lifecycle/unit_impl.md` for the trigger command and close outcomes.

**Constraints requirement during the implementation phase:** When `Next Command` contains `unit_impl`, a `constraints:` prefix MUST be set on `Notes` to define write-permission boundaries. Without constraints, implementation-phase agents have unbounded write access, which risks unintended modifications to spec files, status, or other governed content. The tooling enforces this requirement mechanically: when no constraints are defined during the implementation phase, `specflowctl validate write` denies writes to the following paths by default:
- `docs/specs/units/stable/**`
- `docs/specs/_check_result/**`
- `docs/specs/_check_work/**`
- `docs/specs/_verify_result/**`
- `docs/specs/_stable_verify_result/**`
- `docs/specs/_independent_evaluation/**`
- `docs/specs/_plans/**`
- `docs/specs/_status.md`
- `framework/**`

Paths NOT denied by default include `docs/specs/units/candidate/**` (candidate spec and appendix files) and `docs/specs/repository_mapping.md` — these are intentionally writable during implementation to support path registration and candidate appendix maintenance.

Lifecycle reviews and governance audits MUST flag the absence of constraints during the implementation phase as a governance gap.

### Constraints Derivation

When a command close sets `Next Command` to include `unit_impl` (entering the implementation phase), it MUST also set a `constraints:` prefix. The constraints values are derived from the unit's `implementation_paths` in `docs/specs/repository_mapping.md` Object Registry.

The command close caller MUST:
1. Read the target unit's `implementation_paths` from the repository_mapping.md Object Registry.
2. Build the constraint string using this template:
   ```
   constraints:phase=implementation deny=docs/specs/units/stable/** deny=docs/specs/_check_result/** deny=docs/specs/_check_work/** deny=docs/specs/_verify_result/** deny=docs/specs/_stable_verify_result/** deny=docs/specs/_independent_evaluation/** deny=docs/specs/_plans/** deny=docs/specs/_status.md deny=framework/** allow=<implementation_paths> allow=docs/specs/repository_mapping.md allow=docs/specs/units/candidate/**
   ```
3. The tooling appends the constraint string to `Notes` automatically during `command close`. For tooling-unavailable manual close, include it in the Notes update.

When `implementation_paths` is empty, use the default tooling behavior (`deny=docs/specs/units/stable/** deny=docs/specs/_check_result/** deny=docs/specs/_check_work/** deny=docs/specs/_verify_result/** deny=docs/specs/_stable_verify_result/** deny=docs/specs/_independent_evaluation/** deny=docs/specs/_plans/** deny=docs/specs/_status.md deny=framework/** allow=docs/specs/repository_mapping.md allow=docs/specs/units/candidate/**`) — `allow=docs/specs/repository_mapping.md` and `allow=docs/specs/units/candidate/**` are always included to support implementation path registration per `framework/lifecycle/unit_impl.md` Allowed Writes and candidate-file fixes during implementation, even when no implementation paths are registered yet. With no implementation-path allow patterns, the executor may write to implementation directories (`src/**`, `tests/**`) per the Context Card's Allowed Writes; registering implementation paths in `repository_mapping.md` is the agent's first implementation action.

After recovery (truth_layer or gate_layer fallback) that transitions the unit to `unit_check`, the tooling removes the `constraints:` prefix from `Notes` while preserving other content such as `appendix_exc:`. The next `unit_check` pass outcome writer MUST re-derive constraints from current `repository_mapping.md` when entering the implementation phase again.

The multi-value Next Command is used by tooling for (a) `ContainsNextCommand(NextCommand, "unit_impl")` to differentiate `StateCandidatePending` from `StateCandidateVerify` and (b) `unit_check` re-validation gate enforcement during the implementation phase.

### Appendix Coverage Exclusions

The `Notes` field may carry an `appendix_exc:` prefix to declare stable appendix file references for which the corresponding candidate appendix is intentionally absent. The tooling uses this to avoid false appendix coverage validation failures when a stable appendix is not relevant to the current candidate round.

Alternatively, a stable appendix file may declare `status: exempt` in its frontmatter (see `framework/spec_writing_guide.md` §Appendix Files). When present, the tooling skips that appendix during coverage checks without requiring a `_status.md` entry. The frontmatter approach is preferred for new exclusions because the intent is stored alongside the artifact itself and is respected in all coverage validation paths, including `unit_fork`. The `appendix_exc:` mechanism is retained for compatibility with existing exclusions.

Format (single Notes value, `|`-separated list):

```text
appendix_exc:docs/specs/units/stable/appendix/s_unit_x_a.md|s_unit_x_b.md
```

When combined with other Notes values, separate groups with `;`. Combined example:

```text
appendix_exc:docs/specs/units/stable/appendix/s_unit_x_a.md
```

The tooling loads exclusions during snapshot validation and appendix coverage rendering. `specflowctl snapshot --fix` auto-detects missing candidate appendices and adds the corresponding stable references as exclusions to the unit's Notes (see `framework/process_snapshot_contract.md` §Snapshot Maintenance).

After recovery (truth_layer or gate_layer fallback) that transitions the unit to `unit_check`, the tooling preserves `appendix_exc:` entries — they remain valid for the current candidate round even when the unit returns to `unit_check`. The next `unit_check` pass outcome writer MUST verify that each excluded stable appendix is still irrelevant to the current round; if a stable appendix has become relevant, the corresponding exclusion MUST be removed from `Notes`.

## Update Rules

Lifecycle advancement is valid only when `specflowctl command close` succeeds. Manual edits to `_status.md` are not substitutes for command close.

Truth and gate fallback return to `unit_check`. Candidate verify evidence fallback returns to `unit_verify`. Stable verify evidence fallback returns to `unit_stable_verify`.
