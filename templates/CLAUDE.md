==SPECFLOW:BEGIN==

### 1. Key Terms and References

**What specFlow Is**

This repository uses specFlow to manage development work. specFlow maintains project documents that record accepted design, behavior, boundaries, acceptance criteria, shared rules, and code ownership. A request enters the specFlow flow only when it changes documented project truth. Not every code edit changes a spec document.

**Key Terms**

- **unit** — One independently governed engineering responsibility. May be a feature, module, service, or end-to-end result.
- **rule** — A reusable shared constraint that multiple units may need to follow.
- **Context Card** — A lifecycle file in `specflow/framework/lifecycle/` (e.g. `unit_check.md`) that serves as reference material for the current governance step. The primary agent-facing instruction comes from `specflowctl next`.
- **command close** — A deterministic tooling operation (`specflowctl command close`) that records the result of a completed lifecycle command, advances `Next Command` in `_status.md` according to fixed transition rules, and produces or cleans up process evidence files. The executor does not manually edit `_status.md`.

**Spec Types**

| Type | Description |
|------|-------------|
| **unit** | One independently governed engineering responsibility. May be a feature, module, service, or end-to-end result. |
| **rule** | A reusable shared constraint that multiple units may need to follow. |

**Layers**

| Layer | Meaning |
|-------|---------|
| **stable** | Accepted current project truth. Implementation must conform to stable documents. |
| **candidate** | Proposed next project truth. Must be checked, implemented, and verified before promotion to stable. |

**State Files**

- `docs/specs/_status.md` — Each unit's current layer and the only legal next lifecycle command.
- `docs/specs/repository_mapping.md` — Ownership between units, spec files, and implementation paths.

**specflowctl Location**
`specflow/tooling/bin/specflowctl-<os>-<arch>` — replace `<os>` and `<arch>` with your platform (e.g. `linux-amd64`, `darwin-arm64`, `windows-amd64.exe`)

---

### 2. Classify Your Entry

All specFlow operations fall into one of these categories.
Determine which applies _before_ proceeding to the unit-based flow.

**Pre-formal-truth guidance — proceed to Section 2a.**
If the request is about shaping a design before formal truth is ready (framing a vague idea, cutting scope, choosing between solution directions, reviewing a discussion-stage design, or turning an approved conclusion into formal truth), read `specflow/framework/guidance/using-specflow-guidance/SKILL.md` first. Guidance must be resolved before entering any lifecycle or governance routing.

**Framework governance operations (not unit operations):**
- `spec_flow_review` — scoped review (changed files only). Route to `specflow/framework/governance/review.md`.
- `spec_flow_review:full` — full-scope governance-baseline deep audit. Route to `specflow/framework/governance/review.md`.
- `spec_flow_design_review` — full-scope design-baseline review (no scoped mode). Route to `specflow/framework/governance/review.md`.
- Rule governance entries: any `rule_*` command — read `specflow/framework/governance/rule_system.md` first, then the matching rule file.
- Migration entry: `spec_flow_migrate` — route to `specflow/framework/operations/migration.md`.

For all framework governance entries, skip `--unit` and route directly to the file listed above. Keyword-table routing (expressions containing "mechanism audit", "design review", etc.) is defined in `specflow/framework/governance/review.md` Entries section.

**Unit lifecycle operations — proceed to Section 3.**

---

### 2a. Guidance (Pre-Formal-Truth Design Work)

If the request qualifies as guidance work — shaping a design before formal truth is ready — route to guidance **before** lifecycle routing. Guidance applies when the request is about:

1. framing a vague project or feature idea
2. cutting scope for a first useful version
3. choosing between materially different solution directions
4. reviewing a discussion-stage design before writing it into candidate truth
5. turning an approved discussion conclusion into formal truth

Read `specflow/framework/guidance/using-specflow-guidance/SKILL.md`. Guidance must not replace an exact command, advance lifecycle state, or authorize implementation-side edits. If a guidance conclusion affects behavior truth, re-enter `specflow/framework/operations/entry_routing.md` with the clarified request.

---

### 3. Get Your Unit Directive

This is your first action on every unit-related project request.

If the user named a unit, use that name. If no unit is named, read `docs/specs/_status.md` first to discover active units. If still ambiguous, stop and ask.

Run `specflowctl next --unit <name>`.

Its output tells you:
- **TASK** — what to do in this step
- **READS** — files you may read for reference
- **WRITES** — files you may modify
- **BLOCKED** — files you must not touch
- **COMPLETION** — how to close when done

Execute the TASK. When done, run the COMPLETION command.

**If the directive is insufficient**, run:
  specflowctl next --unit <name> --explain
for full lifecycle context. If still unclear, read `specflow/framework/operations/entry_routing.md`.

**If `specflowctl` is unavailable**, read `specflow/framework/lifecycle/overview.md`, `specflow/framework/operations/entry_routing.md`, and the matching lifecycle Context Card in `specflow/framework/lifecycle/`. If command close is needed and specflowctl is unavailable, follow the Manual Command Close procedure in the active Context Card's "How to End" section.

---

### 4. ⚠️ HARD RULES

These rules override your default helpful-assistant behavior. They are not suggestions.

**HARD RULE 1: Get Your Directive First (MANDATORY)**
You MUST run `specflowctl next --unit <name>` before writing, modifying, or proposing any code. The tool reads the current lifecycle state so you don't have to read `_status.md` directly.
**Exception:** If the unit does not exist in `_status.md` and `specflowctl next` returns an error or empty directive, skip Hard Rule 1 for this round. Instead, read `specflow/framework/operations/entry_routing.md` Natural Language Routes to determine the correct entry action (Onboarding Source Decision for new units), read `specflow/framework/lifecycle/overview.md`, then proceed to the matching lifecycle Context Card.

**HARD RULE 2: No Implementation Without Directive Authority**
You MUST NOT modify implementation code unless `specflowctl next` output lists the target files in WRITES or the active Context Card's Allowed Writes section authorizes the write. If both show "(none)", you must not write implementation code.

**HARD RULE 3: No Truth Drift**
You MUST NOT modify spec files, rule truth, lifecycle state, or repository mapping outside the files listed in `specflowctl next`'s WRITES section or the active Context Card's Allowed Writes.

**HARD RULE 4: Stop When Unclear**
You MUST stop and report status when any of these are true:
- `_status.md` is empty (no units registered). Read `specflow/framework/operations/entry_routing.md` Natural Language Routes for onboarding new units. **Exception:** If the Hard Rule 1 exception condition is met (new unit creation or non-unit operation), follow that exception path instead of stopping.
- No Next Command is recorded for the target unit.
- The path to a required framework file cannot be resolved.
- The request spans multiple units and the correct lifecycle path is ambiguous.
- A natural-language unit request cannot be resolved to one legal existing lifecycle command and active Context Card from current durable truth.
- The target unit is unclear from the request.
- Path ownership is unclear — the `docs/specs/repository_mapping.md` entry for the target unit does not clearly indicate spec or implementation path ownership.
- Implementation permission is not proven — either `specflowctl next` WRITES is empty, no Context Card is active, or the current Context Card does not authorize the write.
- Behavior or rule truth exists only in chat and has not been written to durable truth.
- A rule or repository mapping change is required first.
- The request uses removed scenario lifecycle concepts (`scenario_*`, `scenario_advance:{id}`, or `object-type=scenario`).
Before stopping, run `specflowctl next --explain` for full context.

---

### 5. Lifecycle and Commands Reference

**Lifecycle Flow**

```text
unit_new / unit_fork → unit_check → unit_impl → unit_verify → unit_promote
```

- `unit_check` is a required pre-verify quality gate.
- `unit_impl:{unit}` is a trigger command — provides implementation context without changing lifecycle state. Implementation proceeds during the `unit_verify` phase.
- Lifecycle state advances only through legal `command close`. Do not manually edit `_status.md`.
- For stable alignment checks: `unit_stable_verify`.

**Commands Reference**

**Unit Lifecycle Commands**

| Command | Purpose |
|---------|---------|
| `unit_init:{unit}` | Existing capability → first stable truth |
| `unit_new:{unit}` | Brand new → first candidate truth |
| `unit_fork:{unit}` | Stable truth → candidate change round |
| `unit_check:{unit}` | Candidate truth quality check |
| `unit_impl:{unit}` | Implementation context trigger command |
| `unit_verify:{unit}` | Verify implementation vs candidate truth |
| `unit_promote:{unit}` | Candidate truth → stable truth |
| `unit_stable_verify:{unit}` | Check implementation vs stable truth |

**Rule Governance Commands**

| Command | Purpose |
|---------|---------|
| `rule_new` | Author independent rule truth |
| `rule_extract` | Move unit-local truth into a shared rule |
| `rule_bind` | Bind/remove/retarget a unit's rule dependency |
| `rule_topology` | Edit target-source mapping of shared rule topology |
| `rule_sync` | Compute downstream unit impact after rule changes |
| `rule_escape` | Stop unsafe rule work and route to smallest legal action |

See `specflow/framework/governance/rule_system.md` for routing.

**Framework Governance Commands**

| Command | Purpose |
|---------|---------|
| `spec_flow_review` | Scoped mechanism review (changed files) |
| `spec_flow_review:full` | Full-scope governance-baseline deep audit |
| `spec_flow_design_review` | Full-scope design-baseline review |
| `spec_flow_migrate` | Project-instance format migration |

See `specflow/framework/governance/review.md` for review routing. See `specflow/framework/operations/migration.md` for migration.

**Rule Locations**

Detailed routing, lifecycle, implementation-change, migration, governance review, rule-governance, repository mapping, guidance, and sync rules: `specflow/framework/`.

Project truth inputs: `docs/specs/`.
==SPECFLOW:END==

## Host Instructions

Content outside the managed block below belongs to the host repository.

Keep repository-specific rules outside the managed block. `specFlow` tooling may update only the managed block.
