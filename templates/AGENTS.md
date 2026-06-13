==SPECFLOW:BEGIN==
## specFlow Governance

### What specFlow Is

This repository uses specFlow to manage development work. specFlow maintains project documents that record accepted design, behavior, boundaries, acceptance criteria, shared rules, and code ownership.

A request enters the specFlow flow only when it changes documented project truth, or when current documents are unclear. Not every code edit changes a spec document.

### Key Terms

- **Context Card** — The command-specific lifecycle file in `framework/lifecycle/` (e.g. `unit_check.md`) that tells the executor what files to read, what writes are allowed, and when the command must stop. Only one Context Card is active at a time.
- **command close** — A deterministic tooling operation (`specflowctl command close`) that records the result of a completed lifecycle command, advances `Next Command` in `_status.md` according to fixed transition rules, and produces or cleans up process evidence files. The executor does not manually edit `_status.md`.

---

## ⚠️ HARD RULES — You MUST obey these before any action

These rules override your default helpful-assistant behavior. They are not suggestions.

### HARD RULE 1: Read Status Before Code (MANDATORY)

You MUST read `docs/specs/_status.md` BEFORE writing, modifying, or proposing any code. This is your first action on every project-related request.

### HARD RULE 2: No Implementation Without Lifecycle Authority

You MUST NOT modify implementation code unless:
- An active lifecycle command authorizes it (Next Command = `unit_verify`), OR
- The Implementation Classification in `framework/operations/entry_routing.md` classifies the work as `implementation_only`.

If neither condition is met: STOP and report the current status.

### HARD RULE 3: No Truth Drift

You MUST NOT modify spec files, rule truth, lifecycle state, or repository mapping outside the active Context Card's permitted writes.

### HARD RULE 4: Stop When Unclear

You MUST stop and report status when any of these are true:
- `_status.md` is empty (no units registered). Exception: if the user asks to create the first unit, route to `unit_new:{unit}` per `framework/operations/entry_routing.md` Onboarding Source Decision instead of stopping.
- No Next Command is recorded for the target unit.
- The path to a required framework file cannot be resolved.
- The request spans multiple units and the correct lifecycle path is ambiguous.

### HARD RULE 5: Path Resolution

`framework/...` paths:
- **Source repo** (this is the specFlow repository itself): `framework/...` → `./framework/...`
- **Installed project** (specflow is in a subdirectory): `framework/...` → `specflow/framework/...`

`docs/specs/...` paths are ALWAYS repository-root relative (`./docs/specs/...`).

If the resolved file does not exist: STOP and report the missing path.

---

### 1. Spec Types

| Type | Description |
|------|-------------|
| **unit** | One independently governed engineering responsibility. May be a feature, module, service, or end-to-end result. |
| **rule** | A reusable shared constraint that multiple units may need to follow. |

### 2. Layers

| Layer | Meaning |
|-------|---------|
| **stable** | Accepted current project truth. Implementation must conform to stable documents. |
| **candidate** | Proposed next project truth. Must be checked, implemented, and verified before promotion to stable. |

### 3. State Files

- `docs/specs/_status.md` — Each unit's current layer and the only legal next lifecycle command.
- `docs/specs/repository_mapping.md` — Ownership between units, spec files, and implementation paths.

### 4. Development Flow

```text
unit_new / unit_fork → unit_check → unit_impl → unit_verify → unit_promote
```

- `unit_check` is a required pre-verify quality gate.
- `unit_impl:{unit}` is a trigger command — provides implementation context without changing lifecycle state. Implementation proceeds during the `unit_verify` phase.
- Lifecycle state advances only through legal `command close`. Do not manually edit `_status.md`.
- For stable alignment checks: `unit_stable_verify`.

### 5. How to Start

**Step 1 — Generate your context card**
If the request names a unit, run:
```text
specflowctl context card --object-type unit --object <name>
```
If the request involves a rule, use `--object-type rule`.
The card shows the unit's current state and actionable GUIDANCE.

**Step 2 — Follow the card**
Execute the steps in the card's GUIDANCE section. It covers pre-checks, execution, close syntax, and outcome routing. The card is self-contained for standard lifecycle steps.

**Step 3 — When the card is not enough**
If the card shows the unit is **unregistered**, or the GUIDANCE does not cover your situation:
Read `framework/operations/entry_routing.md` for:
- Onboarding new units (source decision)
- Natural-language routing
- Implementation classification
- Independent review format

**If `specflowctl` is unavailable**, fall back to Step 3 directly: read `framework/operations/entry_routing.md` and the matching lifecycle Context Card in `framework/lifecycle/` to determine the correct action.

### Commands Reference

| Command | Purpose |
|---------|---------|
| `unit_init:{unit}` | Existing capability → first stable truth |
| `unit_new:{unit}` | Brand new → first candidate truth |
| `unit_fork:{unit}` | Stable truth → candidate change round |
| `unit_check:{unit}` | Candidate truth quality check |
| `unit_verify:{unit}` | Verify implementation vs candidate truth |
| `unit_promote:{unit}` | Candidate truth → stable truth |
| `unit_stable_verify:{unit}` | Check implementation vs stable truth |
| `unit_advance:{unit}` | Read `framework/advance_policy.md` first |

### 6. Rule Locations

Detailed routing, lifecycle, implementation-change, migration, governance review, rule-governance, repository mapping, guidance, and sync rules: `framework/`.

Project truth inputs: `docs/specs/`.

Framework-root relative paths: `framework/...` → `./framework/...` (source) or `specflow/framework/...` (installed).
==SPECFLOW:END==

## Host Instructions

Content outside the managed block below belongs to the host repository.

Keep repository-specific rules outside the managed block. `specFlow` tooling may update only the managed block.
