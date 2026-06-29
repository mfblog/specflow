# SpecFlow Concepts

This project uses SpecFlow to manage design documents. SpecFlow maintains spec documents that record accepted design, behavior, boundaries, and shared rules. These documents serve as the consensus protocol between the user and the agent — the user reviews spec documents to confirm intent, and the agent reads spec documents to understand design intent across sessions.

## Core Principle

**File existence is state.** No state machine, no status table, no lifecycle phases. Candidate file exists = being edited. No candidate file = not being edited.

| Directory | Meaning |
|-----------|---------|
| `docs/specs/units/stable/` | Accepted, promoted design truth |
| `docs/specs/units/candidate/` | Design currently being edited |
| `docs/specs/rules/stable/` | Accepted shared rules |
| `docs/specs/rules/candidate/` | Rules being edited |

## Key Terms

- **unit** — One independently governed engineering responsibility
- **rule** — A reusable shared constraint that multiple units may follow
- **stable** — Accepted current project truth
- **candidate** — Proposed next project truth

## specflowctl Location

The specflowctl binary is at:
  specflow/tooling/bin/specflowctl-<os>-<arch>
Replace `<os>` and `<arch>` with your platform (e.g. `linux-amd64`, `darwin-arm64`, `windows-amd64.exe`). Use the full path when running specflowctl commands.

## Workflow

### 1. Discover

Run `specflowctl next --unit <name>` to discover the unit's candidate and stable spec files, appendices, rules, and related units.

### 2. Edit and implement (default mode)

Update the candidate spec and code. No gate before this step. Read first, then write.

### 3. Validate, verify, promote (triggered by user)

At natural transition points the agent suggests the next action. The user confirms with any affirmative response:

| Agent says | Meaning |
|-----------|---------|
| "Shall I run **validate** to check the spec design?" | Read-only subagent per §3a |
| "Shall I run **verify** to check the implementation?" | Read-only subagent per §3b |
| "Ready to **promote** to stable?" | Finalize and archive |

The user can also use explicit triggers at any time:

| Trigger | What agent does |
|---------|-----------------|
| `spec_validate {unit}` | Read-only subagent with the validate checklist below. See §3a. |
| `spec_verify {unit}` | Read-only subagent with the verify checklist below. See §3b. |
| `spec_promote {unit}` | Runs validate then verify. If both pass: `specflowctl promote --unit {unit}`. If fails: stop, report. |

If the user's language is vague ("check this", "see if it's right"), clarify:
"Did you mean **spec_validate** (check design) or **spec_verify** (verify implementation)?"

If the user declines a suggestion, continue editing. Do not insist.

### 3a. Validate

When running `spec_validate {unit}`, read this section, then open a read-only subagent with the checklist below embedded in the task prompt.

**Subagent permissions:**
- ALLOWED: Read, Grep, Glob
- FORBIDDEN: Write, Edit, Bash, Task — do not modify files, execute commands, or spawn sub-agents

**Checklist:**

1. **Frontmatter completeness** — Read `docs/specs/units/candidate/c_unit_{unit}.md`. Verify `id`, `layer` (must be "candidate"), `version`, `unit_refs`, `rule_refs` are all present.
2. **Acceptance items** — Verify `acceptance_item_set` exists with at least one item. Each item must have: `id`, `description`, `verification_type`, `verification_surface`, `implementation_surface`, `verification_method`, `pass_condition`, `not_runnable_yet`.
3. **Reference integrity** — Check that all `unit_refs` point to existing stable spec files. Check that all `rule_refs` point to existing rule files. Check that any referenced appendix files exist.
4. **Cross-unit consistency** — Read related unit candidate specs (from `unit_refs`). Check for contradicting statements about shared protocols, data formats, or behavior.

**Output:**
```
Validate result: PASS | FAIL
1. Frontmatter: PASS | FAIL — reason
2. Acceptance items: PASS | FAIL — reason
3. References: PASS | FAIL — reason
4. Cross-unit: PASS | FAIL — reason
Summary: ...
```

### 3b. Verify

When running `spec_verify {unit}`, read this section, then open a read-only subagent with the checklist below embedded in the task prompt.

**Subagent permissions:**
- ALLOWED: Read, Grep, Glob
- FORBIDDEN: Write, Edit, Bash, Task — do not modify files, execute commands, or spawn sub-agents

**Checklist:**

1. **Per-item verification** — For each acceptance item in the spec:
   - Read the implementation files listed in `affects.files` (or files matching `implementation_surface` / `verification_surface`)
   - Does the implementation satisfy the `pass_condition`?
   - Report per item: PASS / FAIL / CANNOT_DETERMINE with evidence (file paths, line ranges, observations)

2. **Scope check** — Scan affected files for behavioral changes not declared in acceptance items. Verify `affects.files` declarations match actual changes.

3. **Implementation quality** — Check for dead code, over-engineering, or disproportionate change volume.

**Output:**
```
Verify result: PASS | FAIL
Items:
  - {item.id}: PASS | FAIL | CANNOT_DETERMINE — evidence
Scope: PASS | FAIL — findings
Quality: PASS | FAIL — findings
Summary: ...
```

### 4. Promote (only gate)

`specflowctl promote --unit <name>` runs deterministic validation (format checks, required fields, reference integrity) then copies candidate files to stable. This is called only after the agent's internal validate and verify both pass. Only `promote` writes files — everything else is done by the agent directly.

## HARD RULES

These override default helpful-assistant behavior. They are not suggestions.

**HARD RULE 1: Read Specs Before Implementation**
Before modifying code, read the unit's stable and/or candidate spec. Create or update spec when design changes. If no spec exists, create one. Read `framework/spec_writing_guide.md` or reference existing specs for format.

**HARD RULE 2: Promote Is the Only Gate**
Never call `specflowctl promote` without user confirmation. Before promote, always run validate then verify. If either fails, stop and report. The agent does not decide when to validate, verify, or promote — it suggests, the user confirms.

**HARD RULE 3: No Command Is a Gate (Except promote)**
Commands like `next`, `rule`, `validate`, `doctor`, `init`, `migrate` are for discovery and maintenance. They do not check quality or advance state. Only `promote` is a gate.

**HARD RULE 4: Stop When Unclear**
Stop and ask when the target unit is unclear, the required spec or framework file cannot be found, or the next workflow step cannot be determined. Do not guess or proceed with incomplete information.

## Commands Reference

| Command | What it does | Who calls it |
|---------|-------------|-------------|
| `specflowctl next --unit <name>` | Discover unit files and dependencies | Agent |
| `specflowctl promote --unit <name>` | Validate format + copy candidate→stable | Agent (after user confirmation, after internal validate+verify) |
| `spec_validate {name}` (agent trigger) | Read-only subagent with validate checklist (§3a) | User says "spec_validate" or confirms agent suggestion |
| `spec_verify {name}` (agent trigger) | Read-only subagent with verify checklist (§3b) | User says "spec_verify" or confirms agent suggestion |
| `spec_promote {name}` (agent trigger) | Runs validate→verify→promote | User says "spec_promote" or confirms agent suggestion |
| `specflowctl init` | Initialize specFlow project | Human |
| `specflowctl doctor` | Diagnose project setup | Human |
| `specflowctl migrate` | Update hook files and check tooling version | Agent or human (fallback) |
| `spec_flow_migrate` (agent trigger) | Full migration: run tool then check document format | User says "spec_flow_migrate" |
| `specflowctl rule *` | Rule governance | Human maintainer |
| `specflowctl validate` | Validate file write permissions | Human maintainer |

Project truth inputs: `docs/specs/`.
