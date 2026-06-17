# Entry Routing

> **Agent note:** This file is the internal routing source for `specflowctl next`.
> In normal operation, locate and run the specflowctl binary with the full path
> (`./specflow/tooling/bin/specflowctl-<os>-<arch> next --unit <name>`) to receive
> your current directive. Read this file only when `specflowctl next --explain`
> directs you here, or when `specflowctl` is unavailable.

---

## ⚠️ Fallback Entry

**When `specflowctl next` is insufficient:**
1. Run `./specflow/tooling/bin/specflowctl-<os>-<arch> next --unit <name> --explain` for full lifecycle context.
2. If still unclear, this file can help resolve routing for:
   - Natural-language unit or rule requests where no unit name is known
   - Onboarding new units (source decision)
   - Implementation classification
   - Independent review format
   - Governance review entries

Use the directive from `specflowctl next` (see fallback entry above) as the primary directive source.

**If `specflowctl` is unavailable**, read this file and the matching lifecycle Context Card at `framework/lifecycle/unit_*.md` to determine the correct action. The Context Card provides state-specific guidance.

---

This file provides routing rules used by `specflowctl next` and also directly by agents in fallback situations. Its Exact Commands section maps exact command forms to their Context Cards; its Natural Language Routes section resolves requests that do not match an exact command form.
`framework/...` refs are framework-root relative. Installed project entry files define the physical framework root; the specFlow source repository resolves them under local `framework/...`.

Use this file when:
- `specflowctl next` is unavailable or its directive is not sufficient for the current request
- The request does not name a known unit or rule
- Implementation classification is needed before proceeding
Requests limited to implementation-side work must satisfy the Implementation Classification section of this file before proposing or editing implementation-side files when no exact lifecycle Context Card is already active.
Requests must route through this file before implementation-change classification when any of the following is true:
- The request asks for formal truth creation or change
- The request affects behavior, protocol, boundary, acceptance, rule, or ownership
- The request affects lifecycle state, Next Command, stable/candidate state, or unit phase
- The request affects repository mapping, guidance, or skips `_status.md` or owner checks
- The request is a custom reconciliation, audit, alignment, or gap-review
- The request may change field meaning, schema fields, output fields, fixture fields, contract-like log fields, or downstream compatibility (unless the user explicitly limits the work to internal non-semantic implementation support)
This file consumes implementation classification results when the next legal owner is the Onboarding Source Decision section of this file, a lifecycle Context Card, rule governance, repository mapping, framework governance, or guidance.

## Sections

- [Exact Commands](#exact-commands) — exact command forms for lifecycle, rule governance, framework governance
- [Routing Inputs](#routing-inputs) — required durable truth reads before natural-language routing
- [Natural Language Routes](#natural-language-routes) — lifecycle, rule governance, repository mapping, guidance
- [Framework Governance](#framework-governance) — review and migration entries
- [Entry File Registration](#entry-file-registration) — managed block consistency across AGENTS.md, CLAUDE.md, GEMINI.md
- [Implementation Classification](#implementation-classification) — formal classification of implementation-side requests
- [Onboarding Source Decision](#onboarding-source-decision) — candidate unit source_basis rules
- [Hard Stops](#hard-stops) — conditions that require stop/ask/reroute
- [User-Facing Output](#user-facing-output) — report format, human stop, independent review stop, command close-out

## Exact Commands

If the request exactly matches one of these forms, read `framework/lifecycle/overview.md` and the matching lifecycle Context Card:

```text
unit_init:{unit}
unit_new:{unit}
unit_fork:{unit}
unit_check:{unit}
unit_impl:{unit}
unit_verify:{unit}
unit_promote:{unit}
unit_stable_verify:{unit}
```

`unit_init:{unit}`, `unit_new:{unit}`, and `unit_fork:{unit}` share one Context Card at `framework/lifecycle/unit_init_new_fork.md`.
`unit_impl:{unit}` is a trigger command. It does not change lifecycle state or advance `_status.md`. Valid only when `Next Command=unit_verify`. Routes to `framework/lifecycle/unit_impl.md`.

`unit_plan:{unit}` is a removed command. If the user explicitly requests `unit_plan:{unit}`, report that it is no longer a SpecFlow-governed command and that the agent handles planning internally. Route to `unit_impl:{unit}` trigger when `Next Command=unit_verify`. For all other `Next Command` values, route to the Context Card matching the current `Next Command`, reporting that planning is handled internally during the implementation phase.

After a lifecycle Context Card is selected, read the following sections in order:

- **Input** — always read first to gather required truth files
- **Pre-Execution Self-Check** — always perform before proceeding (present in `unit_check.md`, `unit_verify.md`, `unit_stable_verify.md`, `unit_init_new_fork.md`, `unit_promote.md`)
- **What This Step Does** — always read for procedural context
- **Not Allowed** — always observe as hard boundaries
- **Allowed Writes** — always observe as permission authority
- **How to End** — always read for outcome tables and termination conditions
- **On-Demand References** (present in `unit_impl.md`) — enter only when their trigger condition appears

If the request exactly matches a rule governance entry, read `framework/governance/rule_system.md` first, then the matching rule file:

1. `rule_new` -> `framework/governance/rules/rule_new.md`
2. `rule_extract` -> `framework/governance/rules/rule_extract.md`
3. `rule_bind` -> `framework/governance/rules/rule_bind.md`
4. `rule_topology` -> `framework/governance/rules/rule_topology.md`
5. `rule_sync` -> `framework/governance/rules/rule_sync.md`
6. `rule_escape` -> `framework/governance/rules/rule_escape.md`

If the request explicitly invokes one of these framework governance entries, with or without a narrowing phrase, read the matching framework file:

1. `spec_flow_review` -> `framework/governance/review.md`
2. `spec_flow_review:full` -> `framework/governance/review.md`
3. `spec_flow_design_review` -> `framework/governance/review.md`
4. `spec_flow_migrate` -> `framework/operations/migration.md`

Explicit invocation means the request contains the literal entry name as the operation being requested.
Ordinary-language descriptions such as "review governance", "review design", or "migrate specs" do not route to those entries by implication.

If the request uses `scenario_*`, `scenario_advance:{id}`, or `object-type=scenario`, stop and report that scenario lifecycle support has been removed.

## Routing Inputs

Before choosing a natural-language route, read the smallest necessary durable truth:

1. Read `docs/specs/_status.md` when a unit is named.
2. Read `docs/specs/repository_mapping.md` when path ownership, object registration, implementation path registration, or support-surface ownership matters.
3. Read current-layer unit truth when a named unit is involved.
4. Read current-layer rule truth when rule governance is involved.
5. Read `framework/core/adoption_modes.md` when the request asks how small a specFlow adoption can be.

Do not guess ownership from directory shape alone.

## Natural Language Routes

For unit lifecycle requests, select one existing command from the table below based on the current `_status.md` state. Detailed rules follow the table.

| Current State | Selected Command | Notes |
|---|---|---|
| No unit row exists | `unit_init:{unit}` or `unit_new:{unit}` | Read Onboarding Source Decision first |
| Formal row exists | Command matching `Next Command` | `Next Command` is the only legal command |
| `Next Command=unit_verify`, first entry | `unit_impl:{unit}` → then `unit_verify:{unit}` | Trigger command before verify |
| Stable layer, `Next Command=unit_fork`, changes formal unit truth | `unit_fork:{unit}` | Determine `candidate_intent` via `candidate_intent.md` |
| Stable layer, check alignment | `unit_stable_verify:{unit}` | `Next Command` must not be `unit_promote` |
| Candidate truth repair | `unit_check:{unit}` | Standard: `Next Command=unit_check`. Re-validation: `Next Command=unit_verify` with `Notes=pending_impl` and spec changed |
| Implementation-only | Read Implementation Classification section | Routes based on classification result |

> **Procedural surface note:** The natural-language route selection is deterministic when based on the state table above. The table row matching the current `_status.md` state selects the command; the rules below define how the state table rows map to routing decisions. When `specflowctl` is available, run `./specflow/tooling/bin/specflowctl-<os>-<arch> next --unit <name>` instead — it produces a deterministic directive from current `_status.md` state. The rules below are the fallback path and the policy source; they are not the preferred execution entry.

**Priority rule — guidance before lifecycle:** If the request qualifies as pre-formal-truth guidance work (guidance scenarios 1-5 described below), route to guidance **before** applying the lifecycle routing rules. A request about shaping a design before formal truth is clear must not enter Onboarding Source Decision before the guidance session concludes.

Route to unit lifecycle when the request creates, changes, validates, plans, implements, verifies, promotes, or checks one independently governed engineering responsibility.
The responsibility may be local or end-to-end. If the user describes a complete workflow result, model it as a unit whose responsibility is that complete result.

For a natural-language unit lifecycle request, start from the state table row whose "Current State" matches the `_status.md` condition for the target unit, then apply the rules below. Each rule corresponds to one table row. Rules are evaluated top to bottom; when a later rule's specific conditions are met, it overrides the general principle in rule 2.

1. **State-table row: "No unit row exists."** Verified by `_status.md` having no entry for the named unit. Read the Onboarding Source Decision section of this file. Select `unit_init:{unit}` only when the Onboarding Source Decision criteria for direct first-stable onboarding are all met (the status table must already be non-empty, and the capability must be already accepted with no business intent/evidence/ownership unknowns to resolve during writeback). Select `unit_new:{unit}` when `unit_init` conditions are not met or the request creates new candidate truth. If neither selection can be proven from current truth and the request, stop and report the missing decision or prerequisite.

   After selecting `unit_new`, determine `source_basis` from the Onboarding Source Decision table. If the user's request does not specify the source of behavior truth, stop and ask the user which `source_basis` applies (`new_design`, `existing_implementation`, `mixed`, or `replacement`). Do not proceed without a confirmed `source_basis`.

2. **State-table row: "Formal row exists, Command matching `Next Command`."** Verified by `_status.md` having a row for the target unit. The cell value in column `Next Command` is the only legal lifecycle command. Select the matching existing command form and Context Card only when the requested work can legally be performed at that recorded next step. If the user asks for a later lifecycle result, report the recorded prerequisite step instead of skipping it.
   - **unit_impl trigger note:** When `Next Command=unit_verify` and the executor is entering this phase for the first time (no implementation files exist), run `unit_impl:{unit}` as a trigger command before proceeding to `unit_verify:{unit}`.
3. **State-table row: "Stable layer, `Next Command=unit_fork`, changes formal unit truth."** Applicable when `_status.md` shows `Active Layer=stable` AND `Next Command=unit_fork` AND the request touches behavior/boundary/acceptance/rule/ownership truth. Select `unit_fork:{unit}`; determine `candidate_intent` through `framework/candidate_intent.md` and any current valid stable-verify constraint required by the entry Context Card.
4. **State-table row: "Stable layer, check alignment."** Applicable when `_status.md` shows `Active Layer=stable` AND the request asks to verify implementation alignment with stable truth AND `Next Command` is not `unit_promote`. Select `unit_stable_verify:{unit}` (see `framework/core/status.md` "Valid Next Commands" for `unit_stable_verify` check-command semantics).
5. **State-table row: "Candidate truth repair."** Standard entry: `Next Command=unit_check` in `_status.md`. Re-validation entry: `Next Command=unit_verify` with `Notes=pending_impl` in `_status.md` AND candidate spec was modified after the last check pass (per `unit_check.md` Pre-Execution Self-Check item 1). Select `unit_check:{unit}`; the repair must stay inside the card's allowed writes. Other candidate progression selects only the Context Card matching the recorded `Next Command`.
6. **State-table row: "Implementation-only."** A request limited to implementation-side edits must first satisfy the implementation-only route below. It enters a lifecycle Context Card only when the Implementation Classification section of this file requires an existing lifecycle command as the next legal step.

After selecting a lifecycle command from natural language, read `framework/lifecycle/overview.md` and that command's matching Context Card. Do not invent a command alias, enter a generic unit lifecycle without an active Context Card, or ask the user to choose an internal command name.

If a natural-language request contains both rule-governance intent and unit-lifecycle intent, resolve the rule-governance part first through `rule_escape` before entering any unit lifecycle Context Card. Rule changes affect unit constraints and must be settled before lifecycle routing.

Route to rule governance when the request changes shared constraints, reusable prohibitions, mandatory process behavior, rule binding, or rule topology.
For non-exact rule-governance requests, read `framework/governance/rule_system.md` first. Read `framework/governance/rules/rule_escape.md` when selecting a safe first rule flow or rerouting is required.
Bound shared rule consumer discovery must use current-layer unit `rule_refs`.

Route to repository mapping when the request changes path ownership, object registration, implementation path registration, or support-surface boundaries.
Read `framework/core/repository_mapping.md`. Repository mapping does not change unit behavior or rule meaning by itself.

Route to guidance when the request asks to shape a design before formal truth is clear.
Guidance applies when the request is about one of these work shapes before candidate truth, rule truth, global rule truth, or repository mapping truth is ready to write:

1. framing a vague project or feature idea
2. cutting scope for a first useful version
3. choosing between materially different solution directions
4. reviewing a discussion-stage design before writing it into candidate truth
5. turning an approved discussion conclusion into formal truth

When guidance applies, read `framework/guidance/using-specflow-guidance/SKILL.md`.
Guidance must not replace an exact command, advance lifecycle state, authorize implementation-side edits, or treat chat-only agreement as durable truth.
If a guidance conclusion affects behavior truth, boundary truth, acceptance truth, rule truth, global rule truth, or repository ownership, re-enter `framework/operations/entry_routing.md` with the clarified request to determine the correct lifecycle command, rule-governance flow, or framework governance entry. Follow that determined path before any implementation work.

Route requests that are limited to implementation-side proposals or edits through the Implementation Classification section of this file when the requested work touches repo-tracked code, tests, configs, prompts, fixtures, integration scripts, or other implementation-side files for a formal unit and no exact lifecycle Context Card is already active.
That operation owns the implementation-only, truth-writeback-required, and boundary-unclear classification.
It may route to the Onboarding Source Decision section of this file, a lifecycle Context Card, rule governance, repository mapping, framework governance, or guidance.
Implementation permission must be proven before proposing or editing implementation-side files.
Testing, debugging, review, and exploration may inspect or verify. They do not authorize mutation. If they discover behavior, protocol, boundary, acceptance, rule, ownership, lifecycle, or implementation-permission impact, stop before proposing a repair path and route to the owning lifecycle, rule-governance, repository-mapping, guidance, onboarding, or implementation-change owner.

## Framework Governance

Requests that explicitly invoke `spec_flow_review`, `spec_flow_review:full`, `spec_flow_design_review`, or match the keyword table in `framework/governance/review.md` (Entries section, rule 0) route to `framework/governance/review.md`.
If the expression does not match an exact form or any keyword entry, stop per `framework/governance/review.md` "Unrecognized Entry".
`framework/governance/review.md` decides the default path for each review entry.
For `spec_flow_review`, the default is `scoped_review`; it delegates to `framework/spec_flow_review.md` only for exact `spec_flow_review:full`.
For `spec_flow_design_review`, there is no scoped mode; `framework/governance/review.md` delegates to `framework/spec_flow_design_review.md` for the default full-scope design-baseline review.

Requests that explicitly invoke `spec_flow_migrate` route to `framework/operations/migration.md`.

## Entry File Registration

Registered entry index files: `AGENTS.md`, `GEMINI.md`, `CLAUDE.md`.

All registered entry index files must contain exactly one managed block (`==SPECFLOW:BEGIN==` to `==SPECFLOW:END==`). Content outside that block belongs to the host repository. All registered entry files must keep their managed blocks consistent.

The managed block's governance rules take precedence over host content in the same entry file. If host content contradicts the managed block, the executor must follow the managed block and, if the contradiction is material, report it as a governance concern.

**Exception — source_repo layout:** The specFlow source repository (`source_repo` layout) develops the framework itself and does not use specFlow governance for its own development. Its root entry files operate as host-content-only without a managed block. For `source_repo` layout only, the managed block requirement is waived. See `framework/lifecycle/overview.md` for `source_repo` / `installed_project` layout distinction.

If managed blocks differ after edits, choose one file as source and run:
```text
./specflow/tooling/bin/specflowctl-<os>-<arch> entry sync --source <registered-entry-file>
```

When a project entry file's managed block contradicts framework governance rules (including routing, review scope, or lifecycle rules defined in this file or `framework/governance/review.md`), the framework governance rule takes precedence. The executor must report the contradiction as a governance concern.

## Implementation Classification

When a user request is limited to implementation-side work (code, tests, configs, prompts, fixtures) and no exact lifecycle command is active, classify before proceeding.

**Risk-level reference** (guides which classification to pick):

| Level | Scope | Typical Lifecycle |
|-------|-------|-------------------|
| **L1 — micro** | ≤3 files, 1 unit, no API/Schema change | `unit_check → unit_impl → unit_verify` |
| **L2 — repair** | Repair intent (`candidate_intent=repair`), 1 unit | `fork → check → impl → verify → promote` |
| **L3 — feature** | New behavior, API/Schema change, cross-unit | Full lifecycle: `fork → check → impl → verify → promote` |

Risk level is a pre-classification guide only. The formal classification below determines the actual routing.

### Formal Classification

1. **implementation_only** — Fits already-written formal truth. No behavior/boundary/acceptance/rule/ownership truth changes.
2. **truth_writeback_required** — Changes behavior, boundary, acceptance, rule, or ownership truth.
3. **boundary_unclear** — Current truth is insufficient to decide; treat as truth_writeback_required.

Use `implementation_only` only when all: no formal behavior truth changes, repository truth is explicit enough to constrain one implementation result, and the request is pure refactor/test/observability/performance change with unchanged semantics.

A request touches formal behavior truth when it changes: unit goal/boundary, external protocols/field meanings/defaults/validation/error semantics, main flow/state transitions, acceptance criteria, rule body text/binding, or stable global rules.

### Next Steps After Classification

1. Brand-new unit + code request → confirm `source_basis` through the Onboarding Source Decision section (`candidate_intent` is not required for `unit_new` per `framework/candidate_intent.md`) before routing to `unit_new:{unit}`
2. No formal truth + undecided source → resolve onboarding source first through Onboarding Source Decision
3. Existing stable + behavior change → when `Next Command=unit_fork`, route to `unit_fork:{unit}` with `candidate_intent=change`; otherwise route to the Context Card matching the current `Next Command`. Determine `candidate_intent` through `framework/candidate_intent.md` before invoking the fork.
4. Existing stable + large repair → when `Next Command=unit_fork`, route to `unit_fork:{unit}` with `candidate_intent=repair`; otherwise route to the Context Card matching the current `Next Command`. Determine `candidate_intent` through `framework/candidate_intent.md` before invoking the fork.
5. Existing candidate + truth change → write candidate truth first, then route to `unit_check:{unit}`. This enters the re-validation path — confirm the unit's `Next Command` is `unit_check` (standard entry) or `unit_verify` with `Notes=pending_impl` and spec fingerprint changed (re-validation per `unit_check.md` Pre-Execution Self-Check item 1).
6. Small `implementation_only` on stable → continue within stable truth. Do not automatically route to `unit_stable_verify:{unit}` — that requires explicit user intent or evidence misalignment first. When evidence misalignment is detected, route to `unit_stable_verify:{unit}` only after confirming the `Active Layer` is `stable`.
7. `implementation_only` on candidate → route to `unit_impl:{unit}` when `Next Command=unit_verify` allows

## Onboarding Source Decision

Candidate units must record where selected behavior truth comes from:

| source_basis | Meaning | evidence_appendix_ref |
|---|---|---|
| `new_design` | Not using existing implementation | `none` |
| `existing_implementation` | Capturing existing behavior | Must point to evidence appendix |
| `mixed` | Combining retained + new design | Must point to evidence appendix |
| `replacement` | Replacing, not using existing | `none` |

`unit_init` (direct first-stable onboarding) is allowed only when the selected behavior baseline is already accepted and fully reviewable before stable writeback — meaning no business intent, evidence conflicts, material unknowns, or ownership boundaries must be resolved during the writeback itself. `unit_init` cannot be used for the very first unit in a project: the status table must already be non-empty (at least one unit must be registered), because onboarding requires existing repository context to define the capability boundary.

## Hard Stops

Stop and ask or reroute when:

1. the target unit is unclear
2. path ownership is unclear
3. behavior or rule truth exists only in chat and has not been written to durable truth
4. implementation permission is not proven
5. a rule or repository mapping change is required first
6. the request tries to use scenario lifecycle concepts
7. a natural-language unit request cannot be resolved to one legal existing lifecycle command and active Context Card from current durable truth

The agent runtime entry file (`templates/CLAUDE.md` for `source_repo` layout; project-root registered entry file for `installed_project` layout) may impose additional startup-level stop conditions (including empty `_status.md`, unresolvable framework file paths, and multi-unit ambiguity). See the entry file's HARD RULE 4 for the complete startup-level stop list.

## User-Facing Output

User-facing reports must use ordinary project language before internal file names.
When applicable, report:

1. current state
2. completed action
3. next action
4. reason for the next action
5. remaining gap

Use ordinary project language before internal command names.

### Human Stop

When work cannot continue without user input, stop with a plain-language report stating:
1. what is blocking progress
2. the one answer or action needed from the user
3. why the work cannot close correctly without it
4. where execution resumes after the user responds
5. what still cannot be claimed complete

### Independent Review Stop

When the only blocker is independent evaluation and no independent executor capability is available, state:
1. the generated evaluation request file path
2. the trigger instruction from `specflowctl evaluation request`
3. that the reviewer must not modify repository files
4. that execution resumes after the reviewer returns `pass`, `blocked`, or `needs_human_decision`

### Command Close-Out

When a lifecycle command or governance flow mutates durable truth or process state, state:
1. the durable files changed
2. the validation or review evidence used
3. the resulting next legal command or route
4. any affected downstream unit or rule sync work
5. any claim that cannot yet be made

Do not claim lifecycle advancement, rule closure, migration completion, or stable alignment unless the owning gate has accepted the evidence.
