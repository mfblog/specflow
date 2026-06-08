<!-- SPECFLOW:BEGIN -->
## specFlow Governance

### 1. First Read

When no exact command matches, first read `docs/specs/_status.md`. This tells you the current lifecycle state for every unit. A recorded Next Command means an active lifecycle is in progress for that unit.

Then, based on both the request and the current state, route through the first matching owner in `framework/operations/entry_routing.md`.

For pure code work (no behavior truth change), see "Implementation Classification" in `framework/operations/entry_routing.md` for the smaller legal path.

| If the request is... | Then the shortest path is... |
|---|---|
| An exact command (`unit_check:{unit}`, `unit_verify:{unit}`, etc.) | Read the matching Context Card directly (`framework/lifecycle/`). Skip `entry_routing.md`. |
| Pure code work with no truth change | Read Implementation Classification in `framework/operations/entry_routing.md`. |
| A truth change, new unit, rule change, or unclear | Read `docs/specs/_status.md`, then route through `framework/operations/entry_routing.md`. |

### 2. What specFlow Is

This repository uses specFlow to manage development work.
specFlow maintains project documents that record accepted design, behavior, boundaries, acceptance criteria, shared rules, and code ownership.

This does not mean every code edit must change a spec document. A request must enter the specFlow flow only when it changes documented project truth, or when the current documents are not clear enough to choose one correct implementation result.

### 3. Spec Types

| Type | Description |
|------|-------------|
| **unit** | One independently governed engineering responsibility. May be a feature, module, service, or end-to-end result. |
| **rule** | A reusable shared constraint that multiple units may need to follow. |

### 4. Layers

| Layer | Meaning |
|-------|---------|
| **stable** | Accepted current project truth. Implementation should conform to stable documents. |
| **candidate** | Proposed next project truth. Must be checked, implemented, and verified before promotion to stable. |

### 5. State Files

- `docs/specs/_status.md` — Records each unit's current layer and the only legal next lifecycle command.
- `docs/specs/repository_mapping.md` — Records ownership between units, spec files, and implementation paths.

### 6. Development Flow

```text
unit_new / unit_fork → unit_check → unit_impl → unit_verify → unit_promote
```

`unit_check` is a required pre-verify quality gate. `unit_impl` is set automatically by `unit_check pass`. Agents handle implementation internally.

For stable implementation alignment checks: `unit_stable_verify`.

Lifecycle state may advance only through legal `command close`. Do not manually edit `_status.md`.

### 7. Command Index

When the user request exactly matches a command, read the matching Context Card first.
For `unit_advance:{unit}`, read `framework/advance_policy.md` first.

### 8. Rule Locations

Detailed routing, lifecycle, implementation-change, migration, governance review, rule-governance, repository mapping, guidance, and sync rules live under `framework/`.

Project truth inputs live under `docs/specs/`.

Framework-root relative paths use `framework/...` as the logical framework root. In installed projects, resolve them under `specflow/framework/...`; project refs such as `docs/specs/...` remain repository-root relative.
<!-- SPECFLOW:END -->

## Host Instructions

Content outside the managed block below belongs to the host repository.

Keep repository-specific rules outside the managed block. `specFlow` tooling may update only the managed block.
