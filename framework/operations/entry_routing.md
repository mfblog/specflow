# Entry Routing

This file is the only natural-language entry route into active SpecFlow owners.
`framework/...` refs are framework-root relative. Installed project entry files define the physical framework root; the specFlow source repository resolves them under local `framework/...`.

Use this file after the installed entry addendum has identified the request as specFlow work and no exact command directly owns the whole request.
Requests limited to implementation-side work must satisfy the Implementation Classification section of this file before proposing or editing implementation-side files when no exact lifecycle Context Card is already active.
Requests that already ask for formal truth creation or change, no formal truth, behavior, protocol, boundary, acceptance, rule, ownership, lifecycle, lifecycle state, Next Command, stable/candidate state, unit phase, repository mapping, guidance, skipping `_status.md` or owner checks, or a custom reconciliation, audit, alignment, or gap-review route through this file before implementation-change classification.
Requests that may change field meaning, schema fields, output fields, fixture fields, contract-like log fields, or downstream compatibility route through this file unless the user explicitly limits the work to internal non-semantic implementation support.
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

If the request exactly matches `unit_advance:{unit}`, read `framework/advance_policy.md`.
`unit_advance:{unit}` is not a Context Card route.

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

`unit_impl:{unit}` is an auto-advance state, not a user command. It is recorded as `Next Command` by `unit_check pass` close. If a user request arrives during the implementation phase (status shows `Next Command=unit_impl`), the agent should continue implementation work within the governance boundaries defined by `framework/lifecycle/unit_impl.md`.

`unit_plan:{unit}` is a removed command. If the user explicitly requests `unit_plan:{unit}`, report that it is no longer a SpecFlow-governed command and that the agent handles planning internally. Route to `unit_impl` state or `unit_verify:{unit}` depending on current lifecycle state.

After a lifecycle Context Card is selected, read only its Required Context. Enter On-Demand Expansions only when their trigger appears.

If the request exactly matches a rule governance entry, read the matching rule file:

1. `rule_new` -> `framework/governance/rules/rule_new.md`
2. `rule_extract` -> `framework/governance/rules/rule_extract.md`
3. `rule_bind` -> `framework/governance/rules/rule_bind.md`
4. `rule_topology` -> `framework/governance/rules/rule_topology.md`
5. `rule_sync` -> `framework/governance/rules/rule_sync.md`
6. `rule_escape` -> `framework/governance/rules/rule_escape.md`

If the request explicitly invokes one of these framework governance entries, with or without a narrowing phrase, read the matching framework file:

1. `spec_flow_review` -> `framework/governance/review.md`
2. `spec_flow_design_review` -> `framework/governance/review.md`
3. `spec_flow_migrate` -> `framework/operations/migration.md`

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

Route to unit lifecycle when the request creates, changes, validates, plans, implements, verifies, promotes, or checks one independently governed engineering responsibility.
The responsibility may be local or end-to-end. If the user describes a complete workflow result, model it as a unit whose responsibility is that complete result.

For a natural-language unit lifecycle request, select one existing lifecycle command and its Context Card before any lifecycle write:

1. If no formal unit row exists, read the Onboarding Source Decision section of this file. Select `unit_init:{unit}` only when an existing accepted capability already satisfies every direct first-stable onboarding condition. Select `unit_new:{unit}` when the request creates new candidate truth or when a historical capability cannot qualify for direct first-stable onboarding. If neither selection can be proven from current truth and the request, stop and report the missing decision or prerequisite.
2. If a formal unit row exists, its recorded `Next Command` is the only legal lifecycle command. Select the matching existing command form and Context Card only when the requested work can legally be performed at that recorded next step. If the user asks for a later lifecycle result, report the recorded prerequisite step instead of skipping it.
3. A stable unit request that changes formal unit truth selects `unit_fork:{unit}` only when the recorded `Next Command` is `unit_fork`; determine `candidate_intent` through `framework/candidate_intent.md` and any current valid stable-verify constraint required by the entry Context Card.
4. A candidate truth repair selects `unit_check:{unit}` only when the recorded `Next Command` is `unit_check` and the repair stays inside that card's allowed writes. Other candidate progression selects only the Context Card matching the recorded `Next Command`.
5. A request limited to implementation-side edits must first satisfy the implementation-only route below. It enters a lifecycle Context Card only when the Implementation Classification section of this file requires an existing lifecycle command as the next legal step.

After selecting a lifecycle command from natural language, read `framework/lifecycle/overview.md` and that command's matching Context Card. Do not invent a command alias, enter a generic unit lifecycle without an active Context Card, or ask the user to choose an internal command name.

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
If a guidance conclusion affects behavior truth, boundary truth, acceptance truth, rule truth, global rule truth, or repository ownership, route that conclusion into the proper formal truth writeback path before implementation.

Route requests that are limited to implementation-side proposals or edits through the Implementation Classification section of this file when the requested work touches repo-tracked code, tests, configs, prompts, fixtures, integration scripts, or other implementation-side files for a formal unit and no exact lifecycle Context Card is already active.
That operation owns the implementation-only, truth-writeback-required, and boundary-unclear classification.
It may route to the Onboarding Source Decision section of this file, a lifecycle Context Card, rule governance, repository mapping, framework governance, or guidance.
Implementation permission must be proven before proposing or editing implementation-side files.
Testing, debugging, review, and exploration may inspect or verify. They do not authorize mutation. If they discover behavior, protocol, boundary, acceptance, rule, ownership, lifecycle, or implementation-permission impact, stop before proposing a repair path and route to the owning lifecycle, rule-governance, repository-mapping, guidance, onboarding, or implementation-change owner.

## Framework Governance

Requests that explicitly invoke `spec_flow_review` or `spec_flow_design_review` route to `framework/governance/review.md`.
`framework/governance/review.md` decides the default path for each review entry.
For `spec_flow_review`, the default is `scoped_review`; it delegates to `framework/spec_flow_review.md` only for exact `spec_flow_review:full`.
For `spec_flow_design_review`, there is no scoped mode; `framework/governance/review.md` delegates to `framework/spec_flow_design_review.md` for the default full-scope design-baseline review.

Requests that explicitly invoke `spec_flow_migrate` route to `framework/operations/migration.md`.

## Entry File Registration

Registered entry index files: `AGENTS.md`, `GEMINI.md`, `CLAUDE.md`.

All registered entry index files must contain exactly one managed block (`<!-- SPECFLOW:BEGIN -->` to `<!-- SPECFLOW:END -->`). Content outside that block belongs to the host repository. All registered entry files must keep their managed blocks consistent.

If managed blocks differ after edits, choose one file as source and run:
```text
specflowctl entry sync --source <registered-entry-file>
```

## Implementation Classification

When a user request is limited to implementation-side work (code, tests, configs, prompts, fixtures) and no exact lifecycle command is active, classify before proceeding.

**Risk-level reference** (guides which classification to pick):

| Level | Scope | Typical Lifecycle |
|-------|-------|-------------------|
| **L1 — micro** | ≤3 files, 1 unit, no API/Schema change | Direct `unit_verify`, `unit_check` optional |
| **L2 — repair** | Repair intent (`candidate_intent=repair`), 1 unit | `fork → verify → promote`, `unit_check` optional |
| **L3 — feature** | New behavior, API/Schema change, cross-unit | Full lifecycle: `fork → check → impl → verify → promote` |

Risk level is a pre-classification guide only. The formal classification below determines the actual routing.

### Formal Classification

1. **implementation_only** — Fits already-written formal truth. No behavior/boundary/acceptance/rule/ownership truth changes.
2. **truth_writeback_required** — Changes behavior, boundary, acceptance, rule, or ownership truth.
3. **boundary_unclear** — Current truth is insufficient to decide; treat as truth_writeback_required.

Use `implementation_only` only when all: no formal behavior truth changes, repository truth is explicit enough to constrain one implementation result, and the request is pure refactor/test/observability/performance change with unchanged semantics.

A request touches formal behavior truth when it changes: unit goal/boundary, external protocols/field meanings/defaults/validation/error semantics, main flow/state transitions, acceptance criteria, rule body text/binding, or stable global rules.

### Next Steps After Classification

1. Brand-new unit + code request → `unit_new:{unit}`
2. No formal truth + undecided source → resolve onboarding source first
3. Existing stable + behavior change → `unit_fork:{unit}` with `candidate_intent=change`
4. Existing stable + large repair → `unit_fork:{unit}` with `candidate_intent=repair`
5. Existing candidate + truth change → write candidate truth first, then `unit_verify:{unit}`
6. Small `implementation_only` on stable → continue within stable truth, then `unit_stable_verify:{unit}`
7. `implementation_only` on candidate → continue only when `Next Command=unit_verify` allows

## Onboarding Source Decision

Candidate units must record where selected behavior truth comes from:

| source_basis | Meaning | evidence_appendix_ref |
|---|---|---|
| `new_design` | Not using existing implementation | `none` |
| `existing_implementation` | Capturing existing behavior | Must point to evidence appendix |
| `mixed` | Combining retained + new design | Must point to evidence appendix |
| `replacement` | Replacing, not using existing | `none` |

`unit_init` (direct first-stable onboarding) is allowed only when the selected behavior baseline is already accepted and fully reviewable before stable writeback — meaning no business intent, evidence conflicts, material unknowns, or ownership boundaries must be resolved during the writeback itself.

## Hard Stops

Stop and ask or reroute when:

1. the target unit is unclear
2. path ownership is unclear
3. behavior or rule truth exists only in chat and has not been written to durable truth
4. implementation permission is not proven
5. a rule or repository mapping change is required first
6. the request tries to use scenario lifecycle concepts
7. a natural-language unit request cannot be resolved to one legal existing lifecycle command and active Context Card from current durable truth

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
