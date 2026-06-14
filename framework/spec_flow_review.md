# Spec Flow Review

## 1. Purpose

`spec_flow_review` reviews the governance mechanism itself.
This file owns explicit `deep_audit` review for mechanism correctness.
Ordinary or plain exact `spec_flow_review` entry routes through `framework/governance/review.md` first and stays `scoped_review`.
The only full-scope mechanism review entry is exact `spec_flow_review:full`.

It answers five questions:

1. whether the governance rule set still closes the full Spec Flow
2. whether the tooling layer still matches the rule layer
3. whether rule-governance and impact-reconciliation semantics still converge with the command core
4. whether governance documents can make an executor operational without prior `specFlow` knowledge or avoidable reading cost
5. whether the repository may still claim one coherent governance baseline

Deep audit must use exact `spec_flow_review:full`. Plain exact entry must not automatically start full-scope run-state review.

This flow does not review business truth by default.
It reviews the mechanism that governs business truth.
It does not prove that the current governance design is sensible, humane, or worth using as designed.
That judgment belongs to `spec_flow_design_review`.

## 2. Review Standard

`spec_flow_review` judges whether the in-scope governance rules are correct, closed, coherent, executable, and handoff-safe.

It does not pass a review only because the required files were read or the required slices were visited.
Each in-scope rule, file, slice, and cross-convergence path must satisfy the standards in this section.

The fixed standards are `content validity`, `logical closure`, `chain closure`, `state-space closure`, `governance closure and ownership`, `contract drift`, `cross-convergence`, `supporting truth lifecycle closure`, `agent operability`, `tooling boundary`, `project-instance compatibility`, and `project-instance migration closure`.

### 2.1 Content Validity

Governance content must be valid rule information.
It must not use wrong statements, unsupported claims, empty explanation, or text with no execution effect to create only apparent closure.

Content is valid only when:

1. each rule claim is supported by its owner file, template contract, tooling source, or another current in-scope governance rule
2. names, terms, states, command forms, field names, paths, inputs, and outputs really exist where the document says they exist
3. producer and consumer files accept the same names, states, field shapes, inputs, outputs, and result meanings
4. examples and explanatory text do not direct an executor toward an action that conflicts with the formal rule
5. each local rule changes at least one allowed action, forbidden action, stop condition, output, dependency order, writeback target, or resume path
6. each rule can execute against the current repository structure or explicitly stops when the required repository object is absent

If a statement is logically closed only because it rests on a wrong premise, wrong interface, wrong owner, wrong path, missing consumer, or impossible action, the affected slice must not be marked `passed`.

This review may judge whether governance content is wrong, unsupported, unowned, unconsumed, inconsistent with an actual interface, or not executable.
It must not judge whether the whole governance design is too heavy, humane enough, or worth maintaining as designed.
Those design-value judgments belong to `spec_flow_design_review`.

If a discovered concern is a design-value concern rather than a governance-correctness concern, report that boundary and point to `spec_flow_design_review`.
Do not use a design-value concern by itself as a `spec_flow_review` finding.

### 2.2 Logical Closure

Each in-scope governance file must be internally closed.

A file is internally closed only when a capable executor can determine from that file and its explicit links:

1. the entry condition for the rule or flow
2. the governing owner
3. what action is allowed
4. what action is forbidden
5. what must be read before action
6. where durable writeback or process-state writeback may happen
7. when execution must stop
8. what output or stop report is required
9. how execution resumes after a stop, checkpoint, repair, or downstream handoff

If one of those items is intentionally owned elsewhere, the file must link or name the owner clearly enough that the executor does not guess.

### 2.3 Chain Closure

When an in-scope file is one step in a governance chain, the review must include the in-scope files it calls, hands off to, consumes, or depends on.

A governance chain is closed only when:

1. the calling file and owner file describe the same governed object
2. outputs from one step are accepted by the next step with the same meaning
3. status values and result values are fixed and shared by every producer and consumer that uses them
4. permission boundaries match across the handoff
5. stop conditions have a legal resume owner or next action
6. downstream side effects are either closed in the same chain or explicitly handed to the correct owner
7. no step requires the executor to fill a missing interface, owner, state transition, or branch condition from memory or ordinary term meaning

If a chain is missing one required file, contains a conflicting interface, or only works by executor inference, the related local or cross-convergence slice must not be marked `passed`.

### 2.4 State-Space Closure

The review must prove that the governance mechanism has legal progress paths for important states.
Reading every relevant file is not enough.
A locally consistent rule is still invalid when one of its important results leaves the executor with no legal state-changing next action.

For this review, a `state` is any recorded or implied governance condition that can affect:

1. `Active Layer`
2. `Next Command`
3. truth writeback permission or truth writeback target
4. process-file presence, validity, freshness, cleanup, or reuse
5. command, review, migration, checkpoint, or tooling gate result
6. checkpoint position or resume permission
7. implementation permission
8. recovery owner

An `important state type` is any class of state that changes at least one of:

1. the next legal action
2. the allowed writeback location
3. the verification responsibility
4. the recovery path
5. the implementation permission boundary

A `transition` is the complete relation:

```text
current state + command/result/condition -> allowed writeback -> next state -> next legal owner/action
```

State-space closure requires all of the following:

1. every important command result, review result, migration result, checkpoint result, tooling freshness result, fallback result, blocked result, drift result, repair result, and impact-sync result in scope has a legal transition
2. the transition names the allowed writeback target, or explicitly states that no writeback is allowed
3. the transition names the next legal owner or action
4. the transition can change governance state, implementation state, or a documented blocker state before the same command, gate, or review is run again
5. a stop is closed only when the blocking condition, owner, allowed writeback, and resume action are explicit
6. process invalidation, stale fingerprints, changed truth, migration, impact sync, and recovery states identify the owner that may clear, rebuild, or replace the affected state

If `Next Command` remains the same command after a failed, blocked, stale, or fallback result, the review must prove that an intermediate legal action can change the state consumed by that command before rerun.
If no such action exists, or if the intermediate action is only implied by executor judgment, the review must report a dead-loop finding.

The review must not use "the files were read", "the local rule is self-consistent", or "an executor can decide what to do" as a substitute for transition proof.
Missing transitions, ownerless transitions, contradictory transitions, and same-command reruns without a legal state-changing source are findings.

### 2.5 Governance Closure And Ownership

Each in-scope owner area must close from a legal entry to one legal next action, final result, or required stop.

The review must find a real finding when an in-scope rule can cause:

1. ambiguous entry selection
2. missing truth, process-state, tooling, recovery, or close-out ownership
3. bypass of a required command gate, truth writeback gate, rule-governance gate, impact-reconciliation gate, recovery gate, or close-out gate
4. a side effect with no downstream owner
5. a branch that never rejoins a legal command, review, repair, or stop path
6. chat agreement, repository history, directory shape, or ordinary term meaning to substitute for durable governance truth

### 2.6 Contract Drift

Governance contracts must not drift across rule documents, templates, run-state files, tooling contracts, and tooling source.

Contract drift exists when two in-scope surfaces define or consume the same governance object differently, including:

1. different command names or entry forms
2. different state values or result values
3. different required fields or writeback containers
4. different path ownership rules
5. different lifecycle advancement or fallback meanings
6. different tooling responsibilities
7. different freshness, fingerprint, cleanup, sync, or validation rules

Any contract drift that can change execution, stop behavior, review judgment, or downstream ownership is a finding.

### 2.7 Cross-Convergence

Locally correct rules must still compose into one coherent governance baseline.

The review must test cross-convergence wherever one rule area depends on another rule area.
At minimum, cross-convergence covers routing, commands, project-instance migration, truth writeback, implementation gates, onboarding source decision, rule governance, impact reconciliation, process state, entry files, tooling, and recovery when those areas are in scope.

When onboarding source decision is in scope, the review must verify that candidate source fields, candidate main Spec text, evidence appendix handling, implementation permission, and lifecycle gates converge without allowing observed implementation behavior to become implementation truth outside the candidate main Spec.

If a narrowed review crosses a boundary whose owner slice is not included, the narrowed review must stop or explicitly remain non-baseline.
It must not claim default governance-baseline `pass`.

### 2.7.1 Supporting Truth Lifecycle Closure

Durable supporting truth must remain layer-correct across fork, promote, cleanup, rule release, and impact-reconciliation paths.

For this review, supporting truth includes:

1. stable and candidate main Spec files
2. stable and candidate appendix files
3. evidence appendix files
4. Rule refs and Rule files consumed by current-layer Specs
5. process snapshots and cleanup targets that preserve or delete those files
6. deterministic tooling that creates, retargets, deletes, validates, or reports those files

Supporting truth lifecycle closure requires all of the following:

1. a fork path that creates candidate truth from stable truth must prove whether every stable-layer supporting file referenced by the stable main Spec is copied, retargeted, intentionally omitted, or remains only historical evidence
2. a promote path that lands candidate truth as stable truth must prove whether every candidate-layer supporting file is migrated, absorbed, deleted as evidence, or intentionally left out of stable behavior truth
3. success cleanup must preserve current-round supporting truth until the owning command has migrated, absorbed, or intentionally deleted it
4. deterministic auto-fork tooling must implement the same supporting-file writeback, retarget, and cleanup contract as the command rule it mechanizes
5. current-layer main Specs must not depend on previous-layer appendix files as current behavior truth
6. stable main Specs must not point to candidate-layer appendix files as stable behavior truth
7. evidence appendix files must stay evidence only; they must not become stable behavior truth unless an in-scope command rule explicitly makes that migration legal

A review must not mark a supporting-truth lifecycle path `passed` only because files are present, readable, or shape-valid.
It must prove the current layer, link targets, cleanup ownership, and tool/source agreement for the path being reviewed.

If a fork, promote, auto-fork, cleanup, or impact-reconciliation path can leave a current-layer Spec dependent on the wrong layer's supporting truth, can delete current-layer supporting truth before it is absorbed or migrated, or can make tooling behavior diverge from command rules, the review must report a finding.

### 2.8 Agent Operability

Governance files must be operable by a capable executor without prior `specFlow` memory.

A narrowed review must include entry behavior, routing, commands, project-instance migration, checkpoints, rule governance, process state, entry files, and tooling contracts when the narrowed scope covers those areas.
Governance rules are enforced by tooling (`specflowctl`) rather than by agent-operability documents.

Agent-operability review must cover execution clarity, content economy, formal rule voice, and self-containment under Section 2.12 — whether Agent-facing instruction files deliver essential phase instructions inline rather than through chain-linked reading.
A pass claim for an in-scope governance file must not ignore an applicable agent-operability failure.

### 2.8.1 Card Output Verification

> **Scope:** This section applies to `spec_flow_review:full` (deep audit) only. Scoped review does not perform full card generation and verification.

The `specflowctl context card` command generates a per-object context card that serves as the primary agent-facing execution guide. A governance mechanism that depends on card content for agent operability must verify that the generated cards are correct and self-contained for every reachable state.

The reviewer must set up a representative test project (see `framework/governance/card_review_setup.md`) under `_governance_review/` (gitignored), rebuild `specflowctl` from the current source (reusing a stale binary invalidates the review), then generate and verify cards for every reachable state.

For unit cards, every state must be verified: `stable_idle`, `stable_verify`, `candidate_check`, `candidate_pending_impl`, `candidate_verify`, `candidate_promote`, `unregistered`. For rule cards: `stable_bound`, `stable_global`, `candidate_bound`, `candidate_global`, `candidate_new_bound`, `candidate_new_global`, `rule_unregistered`.

The following properties must hold on every generated card:

1. **Content self-containment** — the GUIDANCE section must contain the complete lifecycle procedure from the corresponding `framework/lifecycle/unit_*.md` file (with `{unit}` replaced). The card must not require the agent to chain-read additional files to understand the current state and next action. The GUIDANCE text must never be hardcoded in Go — it must be read from the lifecycle file at runtime.
2. **Heading hierarchy** — headings from the lifecycle file must be demoted by at least one level (`#` → `##`, `##` → `###`) so they nest under `## GUIDANCE` without creating competing top-level sections. Inlined file content in Core Truth must use code blocks so internal headings do not affect the document structure.
3. **Inline vs reference balance** — only `_status.md` and `repository_mapping.md` may be inlined in Core Truth as full content. All other files (specs, rules, appendices, framework documents) must be listed as paths in the References section. The card must not exceed ~200 lines for a typical unit.
4. **Placeholder resolution** — no raw `{unit}` or other template markers may remain unresolved in the output. Every `{unit}` reference in the lifecycle file must be replaced with the actual unit name.
5. **No noise entries** — the card must not show `(missing)` for files that are optional for the current state (e.g., stable spec for a new candidate unit without a stable baseline). Optional files that do not exist must be silently omitted.
6. **Section completeness** — all of STATUS, GUIDANCE, Core Truth, References, WRITES, READS, BLOCKED, CLOSE must be present.
7. **State classification** — the STATUS section must match the object's actual `_status.md` row. `Stable=yes Candidate=no Active=stable Next=unit_fork` must produce `stable_idle`.
8. **Guidance correctness** — the GUIDANCE section must use the lifecycle file that matches the current state (stable_idle → `unit_init_new_fork.md`, candidate_check → `unit_check.md`, etc.).

If any of these properties fails for any in-scope object, the mechanism must not pass the review until the defect is fixed. The reviewer must not substitute a sample-based check for per-object verification.

This standard applies to both unit cards and rule cards. For rule cards, additionally verify:
9. **Impact completeness** — the IMPACTS section must list every unit that references the rule in its `rule_refs` (for bound rules) or every current-layer unit (for global rules). Missing or extra entries are defects.
10. **Consumer count correctness** — the correct number of affected units must be shown, matching the IMPACTS table row count.

### 2.8.2 Evaluation Request Verification

> **Scope:** This section applies to `spec_flow_review:full` (deep audit) only.

The `specflowctl evaluation request` command generates a reviewer handoff file. A governance mechanism that depends on evaluation requests must verify that every reviewer pack produces a correct, minimal request.

The reviewer must generate requests for every pack (`unit_check_pass`, `unit_verify_ready_to_promote`, `unit_stable_verify_advancing`, `freshness_text_drift_reuse`) and verify:

1. **Content sourcing** — `Evaluation Questions`, `Allowed Inputs`, and `Forbidden Inputs` must be parsed from `framework/core/independent_evaluation.md` at runtime, not hardcoded in Go.
2. **Subject as path only** — the Review Subject section must list artifact paths only; full file contents must not be inlined in the request.
3. **Standards as path only** — framework standard files must be listed as paths, not inlined; Evaluation Questions carry the actionable criteria.
4. **No stale placeholders** — `{unit}` must be replaced with the actual unit name in Allowed/Forbidden Inputs.
5. **Section completeness** — all of Request, Reviewer Role, Review Goal, Allowed Inputs, Forbidden Inputs, Review Subject, Review Evidence Refs, Evaluation Questions, Reviewer Output, Executor Receipt, and Trigger Instruction must be present.
6. **No Review Standard Refs section** — the Review Standard Refs section must not appear; standard files are listed in Review Subject instead.

### 2.8.3 Agent Runtime Entry Path Review

The agent runtime entry point is `templates/*.md` for `source_repo` layout and the project-root registered entry file (`AGENTS.md`, `CLAUDE.md`, or `GEMINI.md`) for `installed_project` layout. This file is the first governance content an executor reads at startup. Its navigation logic determines whether the executor reaches the correct route, command, or stop condition. A review that does not trace the entry path from this root file cannot claim that agent operability has been verified.

The review must verify the agent runtime entry path for every in-scope governance surface. Scoped review must apply these checks to the extent that the narrowed scope includes entry files and routing policy.

Required checks:

1. **Entry-to-routing path alignment** — the navigation steps in the entry file (typically "Step 1 → Step 2 → Step 3" or equivalent) must produce the same first-owner selection, route, and command as `operations/entry_routing.md` would produce for the same request type. The entry file must not describe a navigation sequence that diverges from `operations/entry_routing.md` without an explicit override rule. If the entry file and `operations/entry_routing.md` assign different first owners for the same request category, that is a contract-drift finding under Section 2.6.

2. **Fallback path determinism** — when the entry file provides fallback instructions for unavailable tooling, missing context cards, or unresolved state (e.g., "If `specflowctl` is unavailable, fall back to Step 3"), the fallback path must produce a deterministic next owner, explicit route, or clear stop condition. A fallback that routes the executor to a file whose entry instructions then route back to the original fallback trigger is a dead-loop finding under Section 2.4.

3. **Step-defined responsibility boundary** — each step in the entry file's navigation sequence must name a concrete next action, target file, or stop condition. Steps that rely on executor intuition, general term meaning, or prior conversation to decide the next action are agent-operability failures under this section.

4. **Context Card Priority non-circularity** — when `operations/entry_routing.md` declares that the context card takes priority over that file, and the entry file's first step generates a context card, the review must verify that every card state produces either actionable GUIDANCE or an explicit instruction to read `operations/entry_routing.md`. The path "generate card → card says `unregistered` → read `operations/entry_routing.md` → `operations/entry_routing.md` says card takes priority" must not create a circular dependency where no file provides the next legal action. The review must walk at least the `unregistered` card state to verify non-circularity.

5. **Natural-language routing reachability** — when the entry file's fallback step routes to `operations/entry_routing.md` for natural-language requests, the review must verify that a request arriving at `operations/entry_routing.md` through that entry-file path can still reach every legal lifecycle command, rule-governance flow, framework governance entry, and guidance skill that is in scope. The entry file must not filter or narrow the routing surface exposed by `operations/entry_routing.md` without documenting that narrowing as an intentional design decision.

6. **Unavailable-tooling path independence** — when `specflowctl` is unavailable and the entry file directs the executor to read the matching lifecycle Context Card directly, the review must verify that the executor can select the correct Context Card without a context card. This requires at minimum that `_status.md` is readable and the route (exact command or natural-language-derived command) can resolve to one specific lifecycle Context Card under Section 2.2 (Logical Closure). If the path requires a context card that is not available, the review must report an agent-operability finding.

7. **Stop-condition transport** — any stop condition defined in `operations/entry_routing.md` (Hard Stops section) that is restated, implied, or weakened in the entry file must have the same meaning, trigger boundary, and effect in both locations. A stop condition in the entry file that allows the executor to proceed past a condition that `operations/entry_routing.md` requires to stop is a contract-drift finding under Section 2.6. A stop condition in `operations/entry_routing.md` that is missing from the entry file and affects the executor's startup behavior must be reported as an agent-operability finding unless the entry file explicitly delegates to `operations/entry_routing.md` for that stop class.

The reviewer must report, for every in-scope governance surface, which entry file was used as the runtime root, which checks from this subsection were applied, and whether each check passed or produced a finding. A pass claim for agent operability must not ignore an unreviewed or unresolved runtime entry path.

### 2.9 Tooling Boundary

Governance tooling may execute only mechanical work already decided by governance rules, prior human judgment, or explicit caller parameters.
Tooling must not become a second semantic source of truth.

Default full-scope `spec_flow_review` must read and consume `<framework-root>/tooling_execution_policy.md`.
A narrowed review must read and consume that policy whenever the narrowed scope includes governance tooling, tooling contracts, run-state tooling, tooling source, or document/source agreement for tooling.

The tooling review must verify tooling necessity, allowed mechanical action surface, forbidden semantic judgment, freshness rules, and agreement between tooling source and tooling-governing documents.

### 2.10 Project-Instance Compatibility

Default full-scope `spec_flow_review` must perform a narrow project-instance compatibility check for the layout-selected project-instance surface.

For `installed_project`, that surface is real project-instance files under `docs/specs/`.
For `source_repo`, that surface is template bootstrap files under `<template-root>/docs/specs/**` and does not require real project-instance `docs/specs/` files.

This check verifies only whether the current project's SpecFlow instance files can still be read and consumed by the current framework contracts, templates, commands, and tooling.
It does not review business truth correctness.

The compatibility check may judge only:

1. required file presence for current project-instance entry points
2. required section, table, field, frontmatter, status value, command name, reference, and binding shape
3. agreement between project-instance process files and the template-side process contracts
4. agreement between project-instance object references and the layout-selected status file, repository mapping file, and current framework path rules
5. whether existing project-instance files use names, states, command forms, and reference formats that the current framework can consume
6. candidate metadata shape for current candidates, including unit `candidate_intent`, unit `repair_basis` when required, `source_basis`, `evidence_appendix_ref`, required evidence appendix reference presence, and evidence appendix file shape when the current framework requires one
7. current-layer supporting-truth reference shape, including whether candidate main Specs avoid stable appendix dependencies and stable main Specs avoid candidate appendix dependencies
8. appendix frontmatter and path agreement for owner, layer, and file-prefix shape, without judging the appendix's business content

The compatibility check must not judge:

1. whether a unit or rule describes the right business behavior
2. whether acceptance criteria are sufficient for the product
3. whether a candidate or stable Spec should make different design decisions
4. whether implementation actually satisfies a unit or rule
5. whether the current governance design is worth using
6. whether an evidence appendix's observed behavior is business-correct or should be retained

If the project-instance compatibility check finds old file shape, unsupported status values, missing required references, invalid binding format, missing candidate source fields, invalid evidence appendix references, missing required evidence appendix files, or unreadable process-state shape, it is a `spec_flow_review` finding because the framework cannot safely operate on the current project instance.
If the compatibility check finds a current-layer main Spec whose supporting-truth references point to the wrong layer, or an appendix whose owner, layer, or path prefix disagrees with the current framework path rules, it is a `spec_flow_review` finding because current framework commands cannot safely consume that project instance.
If the discovered concern is only about the truth content being wrong, incomplete, or undesirable as business truth, report that it is outside this check and route it to the owning command, rule-governance flow, repository-mapping flow, or design review.

### 2.11 Project-Instance Migration Closure

Default full-scope `spec_flow_review` must review `spec_flow_migrate` as the owner of project-instance format migration after framework rule updates.

The migration closure check verifies only whether the migration flow can safely update old project-instance files to the current framework shape.
It does not review business truth correctness.

The migration closure check must judge:

1. exact entry routing for `spec_flow_migrate`
2. rejection of migration write authority for requests that do not explicitly invoke `spec_flow_migrate`
3. migration read surface and target surface
4. mechanical writeback boundaries
5. forbidden compatibility aliases, fallback logic, and business-truth rewriting
6. process-state invalidation after migrated truth or support files change
7. registered entry managed-block handling
8. blocked-stop and output contracts for blocked migration
9. agreement with tooling boundaries when existing tooling is used

If migration can rewrite project files without a current rule-derived target, preserve stale process pass claims, choose business meaning, or leave invalidated downstream state without a legal next action, it is a `spec_flow_review` finding.

### 2.12 Self-Containment

When a governance file is an Agent-facing instruction file (Context Card, entry file, operation policy that an executor reads directly to decide the next governed action), the file must be self-contained for its essential instructions.

A file fails self-containment when:

1. the file contains an instruction that the Agent must follow to complete the current phase, but the instruction body is only available by reading a linked file
2. the file requires the Agent to read N sequential linked files to obtain the set of essential phase instructions (chain reading)
3. the file uses a link as the primary delivery mechanism for a required action, allowed write, forbidden write, close condition, or gate requirement

Cross-file links are acceptable only for:

1. non-essential background context or design rationale
2. optional skill files that the Agent may choose to load
3. data references (file paths to specs, truth, evidence) that the Agent needs to read as input — these are not instructions about what to do

A review must find a self-containment finding when a Context Card or entry file requires the Agent to follow a chain of two or more links to obtain essential phase instructions that should have been stated directly.

### 2.13 Tool-Enforcement Boundary

When a governance rule describes a hard constraint — an allowed write, forbidden write, required gate, lifecycle state advancement rule, or permission requirement that the executor must not violate — the review must judge whether the rule could be enforced by deterministic tooling (`specflowctl`).

Review rules:

1. if a hard constraint can be validated by a deterministic check (pattern match, state comparison, file existence, fingerprint comparison, phase check), the governance file must not rely on Agent self-enforcement alone; the rule must be implemented in tooling or a documented finding must explain why tooling enforcement is not feasible
2. if a hard constraint requires semantic judgment that cannot be mechanically validated, the governance file may state it as an Agent-facing rule, but the review must note this limitation
3. a governance file that lists multiple hard constraints without tooling enforcement for any of them is a finding — the design is relying entirely on Agent self-discipline

This standard does not require every rule to have tooling enforcement. It requires the review to distinguish between rules that could be enforced (and should be) and rules that inherently require judgment (and must rely on Agent capability).

### 2.14 Relationship To The Slice Catalog

`spec_flow_review` adopts `<framework-root>/slice_work_state_protocol.md` when it uses a review run-state file.
This review file owns the adoption details, the review standard, the slice catalog, and the final conclusion rules.

The baseline slice catalog is an execution organization for this review.
It is not the review standard by itself.

Every local slice, cross-convergence slice, and dynamic slice must be judged against this section.
Coverage without the standards in this section is not sufficient for `pass`.

Command-specific adoption rules:

1. the state carrier for exact `spec_flow_review:full` is `docs/specs/_governance_review/spec_flow_review.md`
2. ordinary scoped `spec_flow_review` does not use that carrier
3. required run fields and slice fields are defined in Section 8
4. baseline local and cross-convergence slices are defined in Section 4
5. dynamic slices are allowed only under Section 5
6. freshness and stale handling are defined in Section 6.5
7. slice-set closure supports a final review conclusion only when every in-scope baseline and dynamic slice closes under this review standard
8. missing governance truth, unclear ownership, or unsupported state transition must become a finding, blocker, or narrowed-scope stop; it must not be hidden by adding implementation work

## 3. Default Scope

This section applies only to explicit `deep_audit`.

Default scope is layout-normalized.

Supported review layouts:

1. `installed_project`
   - framework root: `specflow/framework/`
   - template root: `specflow/templates/`
   - tooling root: `specflow/tooling/`
   - project-instance compatibility mode: real project `docs/specs/`
2. `source_repo`
   - framework root: `framework/`
   - template root: `templates/`
   - tooling root: `tooling/`
   - project-instance compatibility mode: template bootstrap compatibility under `templates/docs/specs/`

`specflowctl review ... --layout auto` detects the layout. `--layout installed` and `--layout source` are explicit overrides.
When auto detection finds both layouts, the review must stop and require an explicit layout.

The default scope includes:

1. framework governance rules
   - `<framework-root>/*.md`
   - `<framework-root>/core/*.md`
   - `<framework-root>/governance/**/*.md` (recursive, includes subdirectories such as `rules/`)
   - `<framework-root>/operations/*.md`
2. command rules
   - active command contracts: `<framework-root>/lifecycle/*.md`
   - lifecycle Context Cards under `<framework-root>/lifecycle/*.md` are the active command contract
3. candidate intent standard rules
   - `<framework-root>/candidate_intent.md`
4. guidance skill rules
   - `<framework-root>/guidance/*/SKILL.md`
5. template-side process and state contracts
   - `<template-root>/docs/specs/_status.md`
   - `<template-root>/docs/specs/_check_work/README.md`
   - `<template-root>/docs/specs/_check_result/README.md`
   - `<template-root>/docs/specs/_plans/README.md`
   - `<template-root>/docs/specs/_plans/draft/README.md`
   - `<template-root>/docs/specs/_plans/active/README.md`
   - `<template-root>/docs/specs/_verify_result/README.md`
   - `<template-root>/docs/specs/_stable_verify_result/README.md`
   - `<template-root>/docs/specs/_governance_review/README.md`
   - `<template-root>/docs/specs/_independent_evaluation/README.md`
6. template-side project-instance bootstrap contracts
   - `<template-root>/docs/specs/repository_mapping.md`
   - `<template-root>/docs/specs/rules/stable/s_g_rule_repository_baseline.md`
7. template entry files
   - `<template-root>/AGENTS.md`
   - `<template-root>/GEMINI.md`
   - `<template-root>/CLAUDE.md`
8. project entry files for `installed_project`
   - `AGENTS.md`
   - `GEMINI.md`
   - `CLAUDE.md`
9. source repository local-entry example for `source_repo`
   - `example.md`
10. entry registry and project-level agent rule files
   - `<framework-root>/operations/entry_routing.md` (Entry File Registration section)
11. tooling contract, tooling source input, and reader runtime input
   - `<framework-root>/tooling_execution_policy.md`
   - `<framework-root>/slice_work_state_protocol.md`
   - `<tooling-root>/README.md`
   - `<tooling-root>/cmd/**/*.go`
   - `<tooling-root>/internal/**/*.go`
   - `<tooling-root>/go.mod`
   - `<tooling-root>/manifest.tsv`
   - `<tooling-root>/go.sum` when it exists
   - `<tooling-root>/scripts/**`
   - `<tooling-root>/reader/web/**`

Default scope excludes project-instance truth and project-instance state files under `docs/specs/` from business-truth review.

Files excluded from business-truth review include:

1. `docs/specs/repository_mapping.md`
2. `docs/specs/_status.md`
3. `docs/specs/rules/stable/s_g_rule_repository_baseline.md`
4. `docs/specs/units/**`
6. `docs/specs/rules/**`
7. `docs/specs/_check_result/**`
8. `docs/specs/_check_work/**`
9. `docs/specs/_plans/**`
10. `docs/specs/_verify_result/**`
11. `docs/specs/_stable_verify_result/**`
12. `docs/specs/_governance_review/**`
13. `docs/specs/_independent_evaluation/**`

Those files may be reviewed for business-truth correctness only when the user explicitly narrows `spec_flow_review` to project-instance state, or when a command, repository-mapping flow, rule-governance flow, or verification flow consumes them under its own policy.

Default full-scope `spec_flow_review` must still perform the compatibility check from Section 2.10.
This check is narrow and does not turn `docs/specs/` into default business-truth review scope.

For `installed_project`, the compatibility input surface includes:

1. `docs/specs/_status.md`
2. `docs/specs/repository_mapping.md`
3. `docs/specs/rules/stable/s_g_rule_repository_baseline.md`
4. existing project process files under `docs/specs/_check_work/**`, `docs/specs/_check_result/**`, `docs/specs/_plans/**`, `docs/specs/_verify_result/**`, and `docs/specs/_stable_verify_result/**`
5. existing independent evaluation request files under `docs/specs/_independent_evaluation/**`, only for request-file path, field, reviewer-pack, and reference shape
6. existing project truth files under `docs/specs/units/**` and `docs/specs/rules/**`, only for file shape, required fields, references, and binding format

`docs/specs/_governance_review/**` is not part of the compatibility input fingerprint.
The active full-scope run-state file is governed by the run-state procedure in Section 6, because including that file in its own slice fingerprint would create self-referential stale state.

For `source_repo`, compatibility input is template bootstrap compatibility under `<template-root>/docs/specs/**`.
It must not require real project-instance `docs/specs/_status.md`, `docs/specs/repository_mapping.md`, or project truth files.

Default scope must explicitly include:

1. the onboarding source decision rule set
   - at minimum `operations/entry_routing.md` where it enters onboarding source decision or advance routing, `advance_policy.md`, `candidate_intent.md`, `spec_writing_guide.md`, `lifecycle/unit_init_new_fork.md` for `unit_init`, `unit_new`, and `unit_fork`, `lifecycle/unit_check.md`, `lifecycle/unit_promote.md`
2. the rule-governance rule set
   - at minimum `operations/entry_routing.md` and `governance/rule_system.md` where they define the rule-governance branch, plus `governance/rules/rule_new.md`, `governance/rules/rule_extract.md`, `governance/rules/rule_bind.md`, `governance/rules/rule_topology.md`, `governance/rules/rule_sync.md`, and `governance/rules/rule_escape.md`
3. the guidance-skill rule set
   - at minimum `using-specflow-guidance/SKILL.md`, `project-framing/SKILL.md`, `scope-cutting/SKILL.md`, `solution-design/SKILL.md`, `design-quality-review/SKILL.md`, and `spec-writeback-guidance/SKILL.md`
4. the impact-reconciliation rule set
   - at minimum `governance/impact_sync.md`, `process_snapshot_contract.md`, `slice_work_state_protocol.md`, `lifecycle/recovery.md`, template `_status.md`, and the template-side process, review-state, and independent-evaluation README files
5. the tooling execution contract set
   - at minimum `tooling_execution_policy.md`, `slice_work_state_protocol.md`, `<tooling-root>/README.md`, the in-scope tooling source files, and the runtime reader web files
6. the agent-operability standard
   - at minimum entry files, routing policy files, `advance_policy.md`, `core/independent_evaluation.md`, `core/freshness.md`, `lifecycle/overview.md`, lifecycle Context Cards, `candidate_intent.md`, rule-governance files, guidance skill files, review policy files, Spec writing policy files, and process-state contract files in the current review scope
7. the state-space closure check
   - at minimum routing policy, advance policy, lifecycle overview, lifecycle Context Cards, candidate intent policy and standards, implementation permission rules, process-state contracts, recovery rules, impact-sync rules, migration rules, and project-instance compatibility inputs needed to prove important non-success transitions
8. the project-instance compatibility check
   - at minimum the layout-selected status, repository mapping, global rule, process-file, and formal truth compatibility inputs, limited by Section 2.10
9. the project-instance migration flow
   - at minimum `operations/migration.md`, `operations/entry_routing.md` where it routes project-instance migration, `lifecycle/overview.md` where it defines command boundaries, `process_snapshot_contract.md`, `lifecycle/recovery.md`, and the template-side process, independent-evaluation, and entry files that migration consumes
10. the supporting-truth lifecycle closure check
   - at minimum fork commands, promote commands, process cleanup and recovery rules, Rule sync and release-version paths, project-instance compatibility inputs, tooling contracts, and tooling source that creates, retargets, preserves, or deletes supporting truth

If any one of those ten coverage sets is missing from a default-scope review, that review is not complete and must not issue `pass`.

## 4. Baseline Slice Catalog

For the default governance-baseline review, the executor must use the baseline slice catalog below.

Baseline slices are the minimum review outline.
They do not limit what the review may discover.
If the review discovers a material risk that is not fully covered by a baseline slice, the executor must add a dynamic slice under Section 5.

The baseline slice catalog organizes review execution.
It does not replace the Review Standard in Section 2.

### 4.1 Local Baseline Slices

Local slices review one owner area for internal closure, side effects, contract drift, missing ownership, and local agent operability.

1. `scope_inventory`
   - verifies default-scope collection, excluded project-instance truth, installed project entry files, source repository entry example files, and unassigned file handling
   - includes the deterministic scope produced by `review collect-default-scope --flow spec_flow_review`
2. `review_entry_policy`
   - reviews `spec_flow_review.md`, `spec_flow_design_review.md`, `governance/review.md`, `governance/review_scope.md`, `severity_policy.md`, and `operations/entry_routing.md` (User-Facing Output section)
   - verifies review entry meaning, output contracts, finding contracts, and stop behavior
3. `routing_and_lifecycle_policy`
   - reviews `operations/entry_routing.md`, `advance_policy.md`, `core/adoption_modes.md`, `core/independent_evaluation.md`, `core/freshness.md`, `lifecycle/overview.md`, `operations/migration.md`, `candidate_intent.md`, `lifecycle/*.md`, and `guidance/*/SKILL.md`
   - verifies exact command routing, exact advance routing, exact project-instance migration routing, natural-language routing, onboarding source routing, unit command progression, and guidance entry behavior
4. `truth_and_implementation_gates`
   - reviews `spec_writing_guide.md`, `core/status.md`, `core/repository_mapping.md`, `candidate_intent.md`, `lifecycle/recovery.md`, and `lifecycle/overview.md`
   - verifies truth ownership, candidate source fields, evidence appendix ownership, implementation diversion, handoff, fallback, and recovery rules
5. `shared_governance`
   - reviews `operations/entry_routing.md` and `governance/rule_system.md` where they define the rule-governance branch
   - reviews `governance/rules/rule_new.md`, `governance/rules/rule_extract.md`, `governance/rules/rule_bind.md`, `governance/rules/rule_topology.md`, `governance/rules/rule_sync.md`, and `governance/rules/rule_escape.md`
6. `process_and_impact_state`
   - reviews `core/independent_evaluation.md`, `core/freshness.md`, `governance/impact_sync.md`, `process_snapshot_contract.md`, `slice_work_state_protocol.md`, `lifecycle/recovery.md`, template `_status.md`, template `_check_work`, template `_check_result`, template `_plans`, template `_verify_result`, template `_stable_verify_result`, template `_governance_review`, and template `_independent_evaluation`
   - verifies process-state contracts, independent-evaluation request contracts, snapshot invalidation, impact handling, and governance-review run-state boundaries
7. `project_instance_contract_compatibility`
   - reviews the current project-instance files under `docs/specs/` only for format and contract compatibility with current framework rules
   - reviews `core/status.md`, `core/repository_mapping.md`, and `spec_writing_guide.md` as the owner contracts for object family, object state, registry shape, formal Spec shape, reference format, and rule binding format
   - reviews `operations/migration.md` as the migration owner for project-instance shape drift discovered by this slice
   - verifies status shape, repository mapping shape, global rules shape, process-file shape, formal object file shape, candidate source metadata shape, candidate intent standard shape, evidence appendix reference shape, evidence appendix file shape, current-layer supporting-truth reference shape, appendix owner/layer/path agreement, reference format, status values, command names, rule binding format, migration writeback boundary, migration state invalidation, migration blocked-stop handling, and migration output closure
   - must not judge unit, rule, or evidence-appendix business truth correctness
8. `entry_and_project_extension`
   - reviews `operations/entry_routing.md` (Entry File Registration section), registered entry files, and template entry files
   - verifies entry file navigation logic correctness under Section 2.8.3, including Step-sequence alignment with routing policy and fallback path completeness
9. `tooling_execution`
   - reviews `tooling_execution_policy.md`, `slice_work_state_protocol.md`, `<tooling-root>/README.md`, in-scope tooling source files, and runtime reader web files
   - verifies tooling necessity, allowed mechanical action surface, forbidden semantic judgment, freshness, reader runtime coverage, and document/source/runtime agreement
10. `agent_operability_local`
   - reviews entry files, routing policy files, `advance_policy.md`, `core/independent_evaluation.md`, `core/freshness.md`, `lifecycle/overview.md`, lifecycle Context Cards, `candidate_intent.md`, rule-governance files, guidance skill files, review policy files, Spec writing policy files, and process-state contract files in the current review scope
   - verifies entry files as the root agent execution entry under Section 2.8.3, including entry-to-routing navigation alignment, fallback path determinism, stop-condition transport between entry files and routing policy, and that entry file routing instructions do not contradict `operations/entry_routing.md`
   - verifies that local slice conclusions, including candidate intent policy, entry-file consumption, tooling-root command refs, and review entry behavior, did not rely on prior conversation, ordinary term meanings, hidden layout assumptions, or avoidable repeated reading

### 4.2 Cross-Convergence Baseline Slices

Cross-convergence slices review whether locally correct rules still compose into one coherent governance baseline.

1. `routing_to_command_convergence`
   - verifies natural-language routing, exact command routing, exact advance routing, exact project-instance migration routing, guidance entry, and review entry behavior converge without ambiguous owner selection
2. `command_to_process_state_convergence`
   - verifies command pass, fallback, cleanup, snapshot, and process-file consumption rules converge
3. `truth_to_implementation_convergence`
   - verifies truth writeback, onboarding source decision, repository mapping, implementation gates, evidence appendix non-truth handling, handoff, and recovery converge
4. `state_space_closure`
   - depends on `routing_and_lifecycle_policy`, `truth_and_implementation_gates`, `process_and_impact_state`, and `project_instance_contract_compatibility`
   - verifies important advance loops, command results, fallback states, checkpoint states, drift states, blocked states, repair states, migration states, and impact-sync states have legal progress transitions
   - verifies same-command reruns have a legal state-changing source before the rerun
   - must not use file-read coverage or local rule consistency as a substitute for transition proof
5. `shared_to_impact_convergence`
   - verifies rule-governance changes correctly converge with impact reconciliation and downstream process-state invalidation
6. `entry_extension_to_review_convergence`
   - verifies entry files and project-level agent rules cannot bypass the framework baseline, narrow default scope silently, or change review meaning without owner rules
7. `tooling_to_rule_convergence`
   - verifies tooling executes only rule-decided mechanical work, does not become a second semantic source of truth, and does not introduce a migration command unless a rule owner defines its mechanical surface
8. `supporting_truth_lifecycle_convergence`
   - depends on `routing_and_lifecycle_policy`, `truth_and_implementation_gates`, `process_and_impact_state`, `project_instance_contract_compatibility`, and `tooling_execution`
   - verifies fork, promote, cleanup, rule `release-version`, rule sync, project-instance compatibility, and tooling-to-rule agreement for stable and candidate main Specs, appendices, evidence appendices, and Rule refs
   - must explicitly report the supporting-truth lifecycle paths walked; a generic cross-convergence `passed` statement is not sufficient
9. `project_instance_to_framework_convergence`
   - verifies the project-instance compatibility check and `spec_flow_migrate` compose with routing, lifecycle, process-state, repository-mapping, shared-binding, entry-file, and tooling rules without judging business truth content
10. `agent_operability_path_walk`
   - walks representative execution paths starting from the agent runtime entry file (`templates/*.md` for `source_repo` or project-root registered entry file for `installed_project`), through routing, advance, command, shared, process-state, entry, and tooling rules
   - verifies a new executor can proceed from the entry file's first step to the correct first owner, route, command, and next legal action without hidden context, prior `specFlow` memory, or circular navigation between the entry file and `operations/entry_routing.md`
   - entry-to-routing alignment under Section 2.8.3, fallback path completeness, and stop-condition transport must be explicitly reported for every walked path

The final result must not issue `pass` until every required local baseline slice, every required cross-convergence baseline slice, and every dynamic slice is closed as `passed` or `skipped_not_in_scope`.

## 5. Dynamic Slices

Dynamic slices extend the baseline catalog during execution.
They are required when a discovered risk is not fully covered by an existing baseline slice.

Rules:

1. a dynamic slice may be local or cross-convergence
2. a cross-area risk must become a cross-convergence dynamic slice instead of being hidden inside one local slice
3. a dynamic slice may only increase review coverage; it must not weaken or replace a baseline slice
4. a dynamic slice must be added before final conclusion when the executor discovers:
   - a new dependency boundary
   - a new owner conflict
   - a new process-state or tooling interaction
   - a new agent-operability risk
   - a finding that needs a separate repairability check
5. every dynamic slice must record:
   - `slice_id`
   - `parent_slice_id`
   - `slice_type`
   - `review_question`
   - `why_added`
   - `input_files`
   - `depends_on`
   - `exit_condition`
   - `status`

## 6. Full-Scope Review Run State

Default full-scope `spec_flow_review` uses a run-state process file.

The process file is not a Spec, not durable behavior truth, and not a substitute for the review output.
It records review progress, slice inputs, stale status, findings, and resume position for one full-scope review run.

The run-state path is:

```text
docs/specs/_governance_review/spec_flow_review.md
```

`review_run_id` is a field inside the run-state file.
It must use this shape:

```text
YYYYMMDD-HHMMSS-{scope_label}
```

There must be at most one `spec_flow_review` run-state file in the repository at any time.
Starting a new full-scope default review must delete the previous `spec_flow_review` run-state file before writing the new run state.
The file name must not contain the run ID, because the run ID identifies the review round inside the file rather than creating a history archive.

### 6.1 When To Use Run State

Rules:

1. exact `spec_flow_review:full` must use the run-state file procedure in this section
2. ordinary scoped `spec_flow_review` must use `framework/governance/review_scope.md` and must not use full-scope run state
3. project-instance truth under `docs/specs/` remains outside default governance-baseline review even though the run-state file itself is read for full-scope resume handling

### 6.1.1 Run-State Tooling Boundary

Run-state files contain both mechanical fields and review judgment fields.

Mechanical fields must be written by deterministic tooling when the tooling is available.
If the tooling is unavailable, the executor must obtain UTC time from the runtime environment before writing timestamp fields.
The executor must not invent timestamps, input fingerprints, or stale-refresh results from conversation context.

The mechanical fields are the fields allowed by `slice_work_state_protocol.md` and this review's run-state contract:

1. `created_at`
2. `last_updated_at`
3. `review_layout`
4. baseline slice skeleton rows
5. `input_fingerprint`
6. stale status changes caused only by changed or missing `input_files`

The deterministic tooling entry is `specflowctl review run-* --flow spec_flow_review --layout auto|installed|source`.

Rules:

1. `review run-init --flow spec_flow_review --layout auto|installed|source` creates, reuses, deletes, or recreates the fixed full-scope run-state file
2. `review run-validate --flow spec_flow_review --layout auto|installed|source` checks the run-state file shape and all fixed status values, including closed statuses; it is not a reuse decision
3. `review run-refresh --flow spec_flow_review --layout auto|installed|source` recomputes slice fingerprints and marks affected `passed` slices as `stale` only for an open run-state file
4. `review run-touch --flow spec_flow_review --layout auto|installed|source` updates only `last_updated_at` on a structurally valid run-state file
5. tooling must not decide whether a slice has passed review
6. tooling must not write finding content
7. tooling must not decide final `pass` or `blocked`
8. an explicit layout that conflicts with an existing open run-state file's `review_layout` must fail instead of rewriting that file

### 6.2 Startup Procedure

At the start of a full-scope review:

1. inspect `docs/specs/_governance_review/spec_flow_review.md`
2. if no unclosed run-state file exists, create a new run-state file and start at `scope_inventory`
3. if one unclosed run-state file exists, run the basic validity check from Section 6.3
4. if the basic validity check fails, delete the old run-state file, report the deletion reason, create a new run-state file, and start at `scope_inventory`
5. if the basic validity check passes, apply the timestamp rules from Section 6.4
6. if the existing file is in `closed_pass` or `closed_blocked`, delete it, report the deletion reason, create a new run-state file, and start at `scope_inventory`
7. the startup procedure must not scan a per-flow subdirectory or preserve old closed run-state files as review history

### 6.3 Basic Validity Check

The basic validity check verifies only that the run-state file can be used as an open progress file.
It does not judge whether old review conclusions are still semantically correct.
It is different from `review run-validate`, which validates file shape and fixed status values without deciding reuse.

For startup reuse, the file is open-valid only when:

1. the file can be read
2. `review_flow` is `spec_flow_review`
3. `scope_label` is `default_governance_baseline`
4. `status` is one of:
   - `in_progress`
   - `blocked_on_finding`
   - `ready_for_final`
5. all required run fields from Section 8.1 exist
6. `created_at` and `last_updated_at` use the timestamp format from Section 6.4
7. the baseline and dynamic slice tables can be parsed
8. every slice status is one of the fixed slice status values from Section 6.6
9. every baseline slice has `parent_slice_id` set to `none`
10. every dynamic slice has `parent_slice_id` set to an existing baseline or dynamic slice in the same run-state file

Rules:

1. `closed_pass` and `closed_blocked` are closed states and must not be reused
2. `review run-validate` may still report a closed run-state file as structurally valid when all required fields, tables, timestamps, and fixed status values are valid
3. any other run status value is invalid and fails the basic validity check
4. if `last_updated_at` cannot be parsed, the file fails the basic validity check

The basic validity check must not decide:

1. whether the old review plan still covers every current risk
2. whether old slice conclusions remain semantically trustworthy after framework changes
3. whether current `specFlow` design is still worthwhile

### 6.4 Timestamp Reuse Rules

The run-state file must update `last_updated_at` whenever a slice status, active slice, finding list, blocked reason, or resume step changes.

Timestamp format rules:

1. `created_at` and `last_updated_at` must use UTC ISO 8601 in this exact shape:
   - `YYYY-MM-DDTHH:MM:SSZ`
2. examples:
   - `2026-04-26T10:30:00Z`
3. timezone offsets other than `Z` are invalid
4. timestamps with missing seconds are invalid
5. invalid timestamps fail the basic validity check
6. `last_updated_at` later than the current UTC time fails the basic validity check

Reuse rules:

1. compute run-state age as current UTC time minus `last_updated_at`
2. if the run-state age is within 2 hours, automatically reuse the run-state file
3. if the run-state age is older than 2 hours and no older than 24 hours, ask the user whether to reuse the run-state file or delete it and start a new run
4. if the run-state age is older than 24 hours and no older than 7 days, ask the user, but recommend deleting the old run-state file and starting a new run
5. if the run-state age is older than 7 days, delete the old run-state file and create a new run unless the user explicitly requests continuing that exact run

When a user chooses to reuse an old run-state file, that choice accepts the old progress record as the continuation basis.
The executor still must refresh file fingerprints and stale statuses under Section 6.5.

### 6.5 Stale Slice Handling

Every slice must record `input_files` and `input_fingerprint`.

On reuse:

1. recompute each slice input fingerprint from the current files
2. change any `passed` slice with changed input fingerprint to `stale`
3. change any cross-convergence slice that depends on a stale slice to `stale`
4. keep unaffected slices in their current status
5. add dynamic slices for newly discovered risks

The authoritative refresh entry is `specflowctl review run-refresh --flow spec_flow_review`.
If that tooling entry is available, executors must use it to update `input_fingerprint` and stale slice state before resuming review work.
Manual fingerprint calculations may be used only as diagnostics; they must not be written into the run-state file or used to decide that a slice remains fresh.

### 6.5.1 Slice Input Fingerprint Contract

Slice input fingerprints use the same text normalization rules as `<framework-root>/process_snapshot_contract.md`.

For each file in `input_files`:

1. file paths must be repository-relative paths rendered with `/`
2. file paths must be sorted lexicographically before hashing
3. read the full file text
4. normalize the text using `process_snapshot_contract.md` Section 7
5. compute `sha256` of the normalized UTF-8 bytes
6. render the file hash as lowercase hexadecimal

The slice fingerprint is computed from the ordered file records.

For each sorted file, append these exact lines to the fingerprint payload:

```text
file_ref: <path>
file_sha256: <hex>

```

Then compute `sha256` of the full payload encoded as UTF-8 and render it as lowercase hexadecimal.

Rules:

1. an empty `input_files` list is invalid unless the slice status is `skipped_not_in_scope`
2. if any input file is missing during fingerprint refresh, the slice becomes `stale`
3. if a missing file prevents the slice from being reviewed, the slice must become `blocked`
4. executors must not use filesystem timestamps, file size, git metadata, or conversation history as fingerprint input
5. dynamic slices use the same fingerprint contract as baseline slices
6. executors must not use shell checksum output, editor display, temporary scripts, or conversation-derived hashes as the authoritative `input_fingerprint`

### 6.6 Run And Slice Status Values

Run status values are fixed:

1. `in_progress`
2. `blocked_on_finding`
3. `ready_for_final`
4. `closed_pass`
5. `closed_blocked`

Slice status values are fixed:

1. `pending`
2. `passed`
3. `blocked`
4. `stale`
5. `skipped_not_in_scope`

Blocked rules:

1. a blocking finding changes the run status to `blocked_on_finding`
2. while blocked, the next review action must be the repair path or re-review of affected slices
3. a blocked run must not advance to final conclusion until all blocking findings are resolved and affected slices are re-reviewed

## 7. Procedure

For full-scope review:

1. collect the default in-scope governance files
2. execute the run-state startup procedure from Section 6.2
3. build or refresh the baseline slice table
4. review local baseline slices
5. add required dynamic slices when new risks are discovered
6. review cross-convergence baseline slices
7. review any cross-convergence dynamic slices
8. refresh stale statuses whenever an input file changes during the run
9. produce findings ordered by governance risk
   - every real finding must use the fixed finding contract from Section 8.2
   - do not collapse a real finding into a one-line conclusion with no repair guidance
10. issue the final result only after all required baseline and dynamic slices are closed

For ordinary scoped review, use `framework/governance/review_scope.md` instead of this full-scope slice procedure.
Ordinary scoped review must not use the full-scope run-state file, baseline slice table, dynamic slice table, or this final `pass | blocked` conclusion contract.

## 8. Output Contract

This output contract applies to explicit `deep_audit`.

The output must report at least:

1. the review scope
2. the review layout
3. the framework root, template root, tooling root, and project-instance compatibility mode
4. whether full-scope run state was created, reused, deleted and recreated, or not used
5. the run-state file path when full-scope run state is used
6. the baseline slice table and slice statuses
7. the dynamic slice table and slice statuses, or explicit `none`
8. the stale slice result
9. the rule-governance coverage result
10. the guidance-skill coverage result
11. the impact-reconciliation coverage result
12. the tooling coverage result, including reader runtime coverage
13. the project-instance compatibility and migration-flow result
14. the agent-operability result, including local slice results and path-walk result
15. the supporting-truth lifecycle result:
   - fork paths reviewed for stable-to-candidate supporting truth retargeting
   - promote paths reviewed for candidate-to-stable supporting truth migration, absorption, or deletion
   - cleanup paths reviewed for preserving current-round supporting truth until owned handling completes
   - deterministic tooling paths reviewed for agreement with command rules
   - wrong-layer current reference cases found, or explicit `none`
16. the cross-convergence results
17. the state-space coverage result:
   - covered state carriers
   - covered commands and governance flows
   - key non-success transitions reviewed
   - same-command rerun cases and their legal state-changing source, or explicit `none`
   - uncovered important state types, if any, reported as findings
18. the findings result:
   - explicit `none` when no real finding exists
   - otherwise every finding must satisfy Section 8.2
   - when real findings exist, the final or stop report shown to the user must include every required information item from Section 8.2 for each finding
   - a run-state file may store the same finding information, but pointing to that file does not satisfy the user-facing report requirement
   - do not summarize a real finding only as a problem statement, impact statement, or blocked reason
19. the final conclusion:
   - `pass`
   - `blocked`

If the output does not explicitly report Items 9 through 18, the review is not complete.

### 8.1 Run-State File Shape

The run-state file must contain these run fields:

1. `review_flow`
2. `review_layout`
3. `review_run_id`
4. `scope_label`
5. `status`
6. `created_at`
7. `last_updated_at`
8. `active_slice`
9. `baseline_slice_table`
10. `dynamic_slice_table`
11. `finding_refs`
12. `blocked_reason`
13. `resume_next_step`

Each slice entry must contain:

1. `slice_id`
2. `slice_origin`
   - `baseline` or `dynamic`
3. `slice_type`
   - `local` or `cross_convergence`
4. `status`
5. `review_question`
6. `why_added`
   - use `baseline_catalog` for baseline slices
7. `parent_slice_id`
   - use `none` for baseline slices
   - dynamic slices must reference an existing baseline or dynamic slice
8. `input_files`
9. `input_fingerprint`
10. `depends_on`
11. `finding_refs`
12. `result_summary`
13. `exit_condition`
14. `resume_next_step`

### 8.2 Narrative Finding Contract

When `spec_flow_review` reports a real finding, that finding must be written as one self-contained repairable story.
The first paragraph must help a new maintainer understand the execution path that breaks.

The first paragraph must use plain language and must be 4 to 6 sentences.
It must answer all of these questions before detailed evidence or trace data appears:

1. who is executing the flow
2. what the executor is trying to complete
3. what the governing rule should make clear
4. where the actual rule, handoff, or state path loses direction
5. how the executor can take the wrong next step
6. what the smallest correct repair point is

Do not present a raw field dump as the user-facing finding.
Lists such as `background`, `what happened`, `impact`, and `recommended fix` may be used only as private drafting aids or trace details; they do not satisfy the user-facing finding requirement by themselves.

The first use of an internal term must explain the term in place.
For example, if a finding mentions a `Context Card`, it must state that this is the command file that tells the executor what to read, what it may write, and when it must stop.

Every real finding must still contain these information items:

1. a title
   - one short problem label
2. severity
   - required for every real finding and must be one of `P0`, `P1`, `P2`, or `P3`
3. background
   - the minimum repository or rule context needed to understand why this finding matters
4. what happened
   - the concrete mismatch, drift, omission, or conflict that was observed
5. impact
   - what governance risk, flow break, or downstream instability this creates
6. recommended fix
   - the concrete repair direction that should be executed next
7. why this fix is the minimal correct fix
   - why the recommendation closes the problem without inventing a wider redesign
8. blocking
   - explicit `yes` or `no`
9. evidence
   - the file refs, block boundary, or tool/runtime result that directly supports the finding

Recommended user-facing shape:

```text
Finding F-006: Natural-language unit requests can send the executor into the wrong lifecycle path.

A user may ask to create, modify, or continue a unit in ordinary language instead of typing an exact command. In that situation, the executor first reads entry_routing.md to decide which lifecycle command file to use. That command file, also called a Context Card, tells the executor what files must be read, what files may be written, and when the command must stop. The current routing text only points to a broad unit lifecycle area, so the executor still has to guess which specific Context Card applies. That can make the executor choose the wrong flow or skip the required read/write boundary. The minimal repair is to make entry_routing.md select a concrete existing Context Card, or stop when it cannot safely choose one.

Fix:
Make entry_routing.md choose the first concrete existing Context Card for natural-language unit lifecycle requests after reading the minimum required status and truth files. If the correct card cannot be determined, it must stop instead of routing to a broad lifecycle area.

Evidence:
...

Status:
P1; blocking for the current review slice.

Trace:
...
```

Additional rules:

1. `severity` must satisfy the shared explanation baseline from `<framework-root>/severity_policy.md`
2. severity describes harm level; it does not replace explicit `blocking`
3. `P0` and `P1` are normally blocking; `P2` and `P3` are normally non-blocking unless the finding explains why the current review must stop
4. `recommended fix` must be specific enough that a later user instruction such as "go fix it" can clearly refer back to that proposed repair without requiring a second clarification round
5. do not replace `recommended fix` with a vague statement such as "should be aligned" or "needs cleanup"
6. if more than one plausible repair exists and the review cannot justify one minimal correct fix, the finding must say that the repair path is still unresolved and the review must not present a guessed fix as settled
7. when no real finding exists, the output must say so explicitly instead of omitting the finding section

## 9. Non-Goals

This flow does not:

1. replace the rule-governance branch
2. replace `impact_sync`
3. review business truth by default
4. treat recently touched governance files as the whole scope unless the user explicitly narrows it that way
5. prove design adequacy, human operability, or design worthiness
6. replace the default `scoped_review` front door in `framework/governance/review.md`
