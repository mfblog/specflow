# Spec-Driven Development Policy

## 1. Purpose

This file defines the formal truth objects used by `specFlow` in this repository.

It answers five questions:

1. which formal objects exist
2. which files carry those objects
3. how object identity is recorded
4. how bindings, snapshots, and invalidation are anchored
5. how executors must read truth before governance actions

## 2. Core Object Families

### 2.1 Command-Target Objects

This repository has two command-target truth object families:

1. `unit`
2. `scenario`

Common rules:

1. both families support `stable` and `candidate`
2. both families enter `docs/specs/_status.md`
3. only these two families are standard command targets
4. every candidate main Spec for these families records `source_basis` and `evidence_appendix_ref` as defined by `onboarding_decision_policy.md`

Family differences:

1. `unit` is the minimal governed unit and is the only family that owns implementation planning and implementation work
2. `scenario` owns trigger-to-outcome chain truth and end-to-end verification, but not implementation planning

### 2.2 `repository_mapping`

`repository_mapping` is the current repository-structure truth.

It answers:

1. what this repository is for
2. which formal objects currently exist
3. which paths belong to which objects or support surfaces
4. which paths are ignored
5. which boundary rules humans and agents must use when judging new paths

It does not answer:

1. one unit's local behavior truth
2. one scenario's trigger-to-outcome chain
3. one reusable rule's body text
4. lifecycle state progression
5. implementation planning or implementation editing

It has one current file:

1. `docs/specs/repository_mapping.md`

It is not a command target.
It does not enter `docs/specs/_status.md`.

### 2.3 `rule`

`rule` is the formal reusable-rule object family.

It answers:

1. which repository-wide default rules are formally active
2. which local reusable rules may be bound by downstream formal objects
3. which exact layer and file carry each rule now
4. which formal objects are currently bound to a bound rule through `rule_refs`

It does not answer:

1. one unit's local behavior truth
2. one scenario's trigger-to-outcome chain
3. repository-structure mapping detail
4. lifecycle state progression
5. implementation planning or implementation editing

Rule files use two formal scopes:

1. `g_` rules are global-scope rules
   - stable `g_` rules are read automatically by every `unit`
   - candidate `g_` rules are proposals and do not apply automatically
2. `b_` rules are bound-scope rules
   - they apply only when a `unit` or `scenario` truth file lists them in `rule_refs`
   - `bound_objects` is metadata and does not replace `rule_refs`

It is not a command target.
Users enter rule work through the rule-governance branch defined by `natural_language_routing.md`.

## 3. Identity And Files

### 3.1 Object Identity

Formal object identity uses the following rules:

1. `_status.md` records bare object IDs
   - `agent`
   - `ai`
   - `task_execution`
2. file names still carry object family prefixes
   - `c_unit_agent.md`
   - `s_unit_ai.md`
   - `c_scenario_task_execution.md`
3. `bound_objects` in Rule files must use typed refs
   - `unit:ai`
   - `scenario:task_execution`

### 3.2 `unit`

1. `stable` -> `docs/specs/units/stable/s_unit_{unit}.md`
2. `candidate` -> `docs/specs/units/candidate/c_unit_{unit}.md`

### 3.3 `scenario`

1. `stable` -> `docs/specs/scenarios/stable/s_scenario_{scenario}.md`
2. `candidate` -> `docs/specs/scenarios/candidate/c_scenario_{scenario}.md`

### 3.4 `repository_mapping`

1. current -> `docs/specs/repository_mapping.md`

### 3.5 `rule`

1. `stable` -> `docs/specs/rules/stable/*.md`
2. `candidate` -> `docs/specs/rules/candidate/*.md`

### 3.6 `_status.md`

`docs/specs/_status.md` is the formal object-state index file.

It records rows for:

1. `unit`
2. `scenario`

Required columns are:

1. `Object Type`
2. `Object`
3. `Stable`
4. `Candidate`
5. `Active Layer`
6. `Next Command`
7. `Notes`

`_status.md` is not behavior truth.
It is the state index that commands must keep aligned with current governance state.

### 3.7 Command-Target Truth Path Resolution

Command-target truth paths are resolved from two inputs:

1. `docs/specs/_status.md`
   - selects the current `Active Layer` for each `unit` or `scenario`
2. the fixed file templates in Sections 3.2 and 3.3
   - define the stable and candidate main Spec file path for that object family

Rules:

1. `_status.md` is the only source for the current layer.
2. `repository_mapping.md` records the truth-surface rule name for each command-target object, not the current active file path.
3. A command must resolve the current main Spec path by applying the object's `Active Layer` to the fixed template for that object family.
4. `unit_promote`, `unit_fork`, `scenario_promote`, and `scenario_fork` change `_status.md` and the relevant truth files; they must not update `repository_mapping.md` only because the active layer changed.
5. `repository_mapping.md` changes only when the object map, path template rule, implementation surface, rule truth path, support surface, governed root, ignore rule, or conflict rule changes.
6. Process snapshots may keep historical `truth_file_ref` values because they describe the truth file used at the time the process snapshot was created.

## 4. Object Boundaries

### 4.1 `unit`

`unit` is the minimal governed unit.

It answers:

1. what responsibility the unit owns
2. which truth surface and implementation surface belong to it
3. which local behavior, protocol, state, and acceptance rules define it
4. how that unit is validated
5. where repairs must land first when the unit is wrong

It must not silently mean:

1. one directory
2. one package
3. one service
4. one bounded context

### 4.2 `scenario`

`scenario` is the formal trigger-to-outcome chain object.

It answers:

1. what triggers the chain
2. which units it traverses
3. which rules it reuses
4. what success means
5. where failure is absorbed, surfaced, or rolled back
6. how the chain is verified end to end

It does not own:

1. unit-local implementation detail
2. rule body text
3. repository mapping rules
4. direct implementation editing

### 4.3 `repository_mapping`

`repository_mapping` is the repository structure truth file.

It answers these mandatory sections:

1. `Project Overview`
   - what this repository is for
   - the main delivery surface
   - the shortest useful reading path
2. `Governed Object Map`
   - current `unit` IDs and one-line responsibilities
   - current `scenario` IDs and one-line responsibilities, or `none`
   - current `rule` IDs and one-line responsibilities
3. `Boundary Rules`
   - what qualifies as a formal `unit`
   - what must become `rule`
   - what stays outside command-target truth
4. `Path Ownership`
   - which roots are governed
   - which paths are ignored
   - which path rules and implementation surfaces map to which current formal object
   - how conflicts are decided
5. `Rule Alignment`
   - which stable `g_` rules currently constrain the repository mapping
6. `Drift Handling`
   - what counts as mapping drift
   - how consumers must stop when drift is found

It does not own:

1. unit-local behavior truth
2. rule body text
3. scenario-local chain detail
4. implementation planning or implementation editing
5. command lifecycle state

### 4.4 `rule`

`rule` answers:

1. one global or bound reusable rule
2. not the whole unit
3. not the whole scenario
4. not the whole repository mapping
5. not the whole global baseline

When a candidate-layer rule file already has a stable-layer sibling for the same `rule_id`, that candidate file also owns the explicit next-landing owner for the reopened rule round.

## 5. Required Binding Fields

### 5.1 `unit`

Each current-layer unit truth must record:

2. `rule_refs`

Each candidate-layer unit main file must additionally record these frontmatter fields:

1. `source_basis`
2. `evidence_appendix_ref`

`unit` does not formally record `scenario_refs`.

### 5.2 `scenario`

Each current-layer scenario truth must record:

1. `repository_mapping_ref`
2. `unit_refs`
3. `rule_refs`

Each candidate-layer scenario main file must additionally record these frontmatter fields:

1. `source_basis`
2. `evidence_appendix_ref`

### 5.3 `repository_mapping`

`repository_mapping` must record:

1. current `unit` IDs
2. current `scenario` IDs, or `none`
3. current `rule` IDs

This is repository-structure truth, not lifecycle binding metadata for a command-target object.

### 5.4 `rule`

Each current-layer rule file must record:

1. `rule_id`
2. `rule_scope`
   - `global`
   - `bound`
3. `layer`
4. `rule_version`
5. `bound_objects`
   - `all_units` for stable global rules
   - typed refs or `none` for bound rules and candidate global rules

Conditional field:

1. when a candidate-layer rule file already has a stable-layer sibling for the same `rule_id`, that candidate file must also record exactly one `promotion_owner_unit`
   - it must be a bare unit id
   - it must name one formal unit from current repository truth
   - it is the only unit round allowed to land that candidate rule file as the next stable-layer Rule file
2. when a candidate-layer rule file does not have a stable-layer sibling, `promotion_owner_unit` must not be recorded
3. stable-layer rule files must not record `promotion_owner_unit`
4. when a command or rule-governance flow explicitly keeps a touched bound rule file with no current formal bindings as independently authored rule truth, that same file must record exactly these intentional-unbound retention fields:
   - `unbound_retention: intentional`
   - `unbound_retention_reason: <non-empty reason>`
   - `unbound_retention_owner: <owning command or rule-governance flow>`
5. `unbound_retention_owner` must name the command or internal rule-governance flow that owns the terminal-state decision in the current round, for example `unit_fork`, `unit_promote`, or `rule_topology`
6. the intentional-unbound retention fields may be recorded only when `bound_objects=none`
7. when the resulting rule file has one or more formal bound objects, the intentional-unbound retention fields must not be recorded
8. when a file that previously carried intentional-unbound retention becomes formally bound again, the same round that restores the binding must remove `unbound_retention`, `unbound_retention_reason`, and `unbound_retention_owner`
9. intentional-unbound retention fields are terminal-state truth for the rule file only; they do not replace `rule_refs`, do not create a formal binding, and do not skip required `rule_sync` or `impact_sync` reconciliation

### 5.5 Testability / Acceptance Criteria Contract

Each current-layer `unit` and `scenario` main Spec must include a `Testability / Acceptance Criteria` section, or an explicitly equivalent acceptance section title.

This section is not a prose-only result description.
It is the formal list of verifiable acceptance items that downstream `check`, `plan`, `verify`, `stable_verify`, and `promote` commands must consume.

Each acceptance item must record these fields:

1. `id`
   - a stable, object-local identifier
   - examples: `ai.model_provider_public_port`, `runtime.task_dispatch_integration`
2. `target`
   - the exact behavior, protocol, boundary, event, storage effect, or external outcome being accepted
3. `verification_surface`
   - exactly one value from the fixed list below
4. `implementation_surface`
   - the concrete package, path set, entrypoint, storage surface, event surface, or manual effect surface that must satisfy the item
5. `verification_method`
   - the command, test, inspection, fixture, external-consumer stub, or manual observation that can prove the item
6. `pass_condition`
   - the concrete observed condition required for this item to pass

The first-version fixed `verification_surface` values are:

1. `public_api`
2. `internal_flow`
3. `error_handling`
4. `eventing`
5. `storage`
6. `integration`
7. `manual_effect`

Runnable-state rules:

1. If an acceptance item cannot be verified in the current repository state, the item must explicitly record `not_runnable_yet` and a non-empty reason.
2. `not_runnable_yet` never counts as `pass`.
3. A command must not silently treat missing test harnesses, missing runtime entrypoints, or unavailable external effects as passed acceptance.
4. `not_runnable_yet` may be used only to avoid making a false pass claim. It does not allow implementation or verification to claim the underlying behavior is complete.

Surface-specific rules:

1. For `verification_surface=public_api`:
   - `implementation_surface` must name the public package, file, or exported contract surface
   - `verification_method` must describe an external-consumer style check
   - `pass_condition` must state that the consumer can satisfy the contract without importing `internal` packages
2. For `verification_surface=integration`:
   - the item must name the runnable integration entrypoint or chain
   - if no such entrypoint exists yet, the item must be marked `not_runnable_yet` with the missing entrypoint reason
   - a broad integration claim must not be counted as accepted only because unit-local pieces pass
3. For `verification_surface=manual_effect`:
   - `verification_method` must name the exact human-observable effect and the observation procedure
   - commands may use `human_verify` only when the remaining uncertainty is truly effect judgment rather than missing executable evidence

Acceptance item writing rules:

1. Do not infer acceptance items from words such as "must", "only", "external", or "replaceable".
2. If a requirement is important enough to block planning, implementation, verification, or promotion, it must appear as an explicit acceptance item.
3. A vague item such as "works", "aligns with design", "is replaceable", or "supports integration" is not sufficient unless the required fields make it directly verifiable.
4. By default, every acceptance item is a current gate item that downstream commands must close.
5. An item is outside the current pass claim only when it explicitly records `not_runnable_yet`, gives the reason, and its `pass_condition` states that it is not a current pass claim.
6. Commands must not infer "key" or "non-key" status from wording, position, length, or apparent importance.
7. Historical stable Specs are not required to be rewritten immediately only because this contract was introduced. They must be brought into this format the next time the object enters `unit_stable_verify`, `scenario` verification, or a fork that touches the acceptance section.

Example item shapes:

```md
- id: ai.model_provider_public_port
  target: External model adapters can implement the AI model provider contract without importing internal packages.
  verification_surface: public_api
  implementation_surface: AgentCore/contracts/model*.go; AgentCore/ports/model_provider.go
  verification_method: External-consumer style compile test with a stub provider that imports only contracts and ports.
  pass_condition: The stub implements ModelProvider and StreamingModelProvider using only public contracts/ports types and no internal import path.
```

```md
- id: runtime.task_dispatch_integration
  target: Runtime dispatch reaches the task execution scenario entrypoint.
  verification_surface: integration
  implementation_surface: Runtime trigger-to-outcome entrypoint
  verification_method: not_runnable_yet
  not_runnable_yet_reason: The repository does not yet expose a complete runtime entrypoint for this chain.
  pass_condition: Not a current pass claim; it becomes runnable only after the runtime entrypoint exists.
```

## 6. Binding Contracts

### 6.1 Rule Binding Contract

When current-layer truth records `rule_refs`, executors must treat that field as the only formal source of which rule files are currently bound.

Rules:

1. `rule_refs` must name the exact layer and file currently bound
2. stable-layer command-target objects may bind only stable-layer rule truth
3. candidate-layer command-target objects may bind stable-layer or candidate-layer rule truth, but the bound layer must be explicit
4. `bound_objects` is declarative metadata only; it does not replace the command-target object's formal binding source
5. a `bound_objects`-only delta does not by itself invalidate downstream process files
6. `bound_objects` must use typed refs only
   - `unit:<id>`
   - `scenario:<id>`
7. when `rule_refs` is written as a markdown list, executors must normalize the ref order by exact rule ref string in ascending lexical order
8. during a same-round stable landing by `unit_promote`, another candidate-layer unit may be retargeted from a candidate-layer Rule ref to the stable-layer Rule ref created by that same landing only after the stable Rule file has been written in the same repository state
9. same-round stable landing retargeting is legal only when the candidate and stable Rule refs share the same `rule_id` and `rule_version`, and the change is a layer target change from `candidate` to `stable`
10. same-round stable landing retargeting must not change Rule body truth, unit-local behavior truth, or acceptance meaning beyond the exact `rule_refs` layer target and the directly required body-level reference wording
11. any unit retargeted by same-round stable landing must be at `candidate`; stable units must fork before their truth can be retargeted
12. a retargeted candidate unit's process files are no longer reusable and must fall back to `unit_check` through rule impact reconciliation

### 6.2 Dependency Direction Contract

Formal dependency direction is fixed:

1. `repository_mapping -> unit/scenario/rule`
2. `stable g_ rule -> rule/unit/scenario/repository_mapping`
3. `rule -> unit/scenario`
4. `unit -> scenario`

Downstream invalidation rule:

1. upstream change may invalidate downstream process files or stable-alignment claims
2. downstream change does not automatically invalidate upstream truth

### 6.3 Version Contract

Formal version values use `MAJOR.MINOR.PATCH`.

Unit `stable` version meaning:

1. `MAJOR`
   - incompatible formal contract change
2. `MINOR`
   - new capability or compatible behavior change in the formal contract
3. `PATCH`
   - implementation-only fix or alignment-only fix against the current aligned layer

`s_g_rule_repository_baseline.md` version meaning:

1. `MAJOR`
   - incompatible global constraint change
2. `MINOR`
   - new global default rule, reusable mechanism, or compatible extension
3. `PATCH`
   - wording-only clarification that does not change the meaning of formal constraints

Rule version rules:

1. `rule_version` uses `MAJOR.MINOR.PATCH`.
2. The first candidate-layer file for a brand-new rule object starts at `0.1.0`.
3. When a current round opens the next candidate-layer file for a rule object that already has a stable-layer file, that candidate file must carry the intended next stable `rule_version`.
4. When Rule 3 applies, that candidate file must also record the exact `promotion_owner_unit` required by Section 5.4.
5. `MAJOR`
   - incompatible change to the formally bound rule semantics, required consumer interpretation, or cross-unit contract shape
6. `MINOR`
   - compatible rule-truth extension, additional reusable capability, or compatible topology evolution that requires consumer awareness but not contract breakage
7. `PATCH`
   - wording-only clarification, compatible tightening, or alignment-only update that does not change the required consumer interpretation

Candidate content may change frequently.
It enters formal version semantics only when promoted into a new `stable`.

## 7. Reading Rules

Before any governance action:

1. read the target object's current-layer main file
2. read any explicitly required appendix truth for that object family
3. read bound rule files when `rule_refs` is not empty
5. read `docs/specs/repository_mapping.md` when object boundary, path ownership, support surface, or current object map matters

Additional rules:

1. do not guess bindings by scanning unrelated files first
2. do not treat natural-language mentions as formal bindings
3. do not skip explicitly bound current-layer truth

## 8. Process Files And Snapshots

Process files are not behavior truth.
They are current-round derived artifacts.

Process containers by object family:

1. `unit`
   - `_check_result`
   - `_plans`
   - `_verify_result`
2. `scenario`
   - `_check_result`
   - `_verify_result`

Object-owned snapshot extensions are fixed by object type:

1. `unit`
   - `unit_appendix_snapshot`
   - `rule_snapshot`
2. `scenario`
   - `repository_mapping_snapshot`
   - `unit_snapshot`
   - `rule_snapshot`

Process files become invalid when their required current truth or required current bindings no longer match.

Candidate evidence appendix files are candidate appendix files for snapshot and invalidation purposes.
Their inclusion in `unit_appendix_snapshot` or the scenario candidate snapshot proves which evidence was reviewed by the gate, but it does not make the evidence appendix an implementation truth source.
Implementation and verification commands must use the candidate main Spec, retained behavior rules, bound Rule files, and `s_g_rule_repository_baseline.md` as truth.

The exact snapshot field definitions come from `process_snapshot_contract.md`.

## 9. Invalidation And Reconciliation

When upstream truth or binding changes:

1. invalidate downstream process files deterministically
2. fall back the downstream object to the minimum legal next step defined by its command family
3. keep Rule topology reconciliation and generic impact reconciliation separate

Formal routing remains:

1. `rule_sync` for rule-governance downstream discovery
2. `impact_sync` for generic fallback and cleanup once the affected downstream object set is fixed
