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

Shared rules:

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
3. one shared contract's local reusable rule text
4. lifecycle state progression
5. implementation planning or implementation editing

It has one current file:

1. `docs/specs/repository_mapping.md`

It is not a command target.
It does not enter `docs/specs/_status.md`.

### 2.3 `system_constraints`

`system_constraints` is the unique global system-constraint object.

It answers:

1. what the current repository-wide engineering baseline is
2. which global default rules are formally active
3. which global prohibitions or explicit exceptions exist
4. which shared mechanisms have already been absorbed into the global baseline

It does not answer:

1. one unit's local behavior truth
2. one scenario's trigger-to-outcome chain
3. repository-structure mapping detail
4. one shared contract's local reusable rule text

It has one effective file only:

1. `docs/specs/system_constraints.md`

It is not a command target.

### 2.4 `shared_contract`

`shared_contract` is an independent shared local-truth object reused by multiple downstream formal objects.

It answers:

1. which local reusable rule is currently shared
2. which exact layer and file carry that shared rule now
3. which formal objects are currently bound to it through declarative metadata

It does not answer:

1. a whole unit
2. a whole scenario
3. the whole repository mapping
4. the whole global baseline

It is not a standard command target.
Users enter shared work through the shared-governance branch defined by `natural_language_routing.md`.

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
3. `bound_objects` in Shared Contract files must use typed refs
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

### 3.5 `shared_contract`

1. `stable` -> `docs/specs/shared_contracts/stable/*.md`
2. `candidate` -> `docs/specs/shared_contracts/candidate/*.md`

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
3. which shared contracts it reuses
4. what success means
5. where failure is absorbed, surfaced, or rolled back
6. how the chain is verified end to end

It does not own:

1. unit-local implementation detail
2. shared-contract body text
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
   - current `shared_contract` IDs and one-line responsibilities
3. `Boundary Rules`
   - what qualifies as a formal `unit`
   - what must become `shared_contract`
   - what stays outside command-target truth
4. `Path Ownership`
   - which roots are governed
   - which paths are ignored
   - which paths map to which current formal object
   - how conflicts are decided
5. `Global Constraint Alignment`
   - which `system_constraints` version currently constrains the repository mapping
6. `Drift Handling`
   - what counts as mapping drift
   - how consumers must stop when drift is found

It does not own:

1. unit-local behavior truth
2. shared-contract body text
3. scenario-local chain detail
4. implementation planning or implementation editing
5. command lifecycle state

### 4.4 `shared_contract`

`shared_contract` answers:

1. one shared reusable local rule
2. not the whole unit
3. not the whole scenario
4. not the whole repository mapping
5. not the whole global baseline

When a candidate-layer shared file already has a stable-layer sibling for the same `shared_contract_id`, that candidate file also owns the explicit next-landing owner for the reopened shared round.

## 5. Required Binding Fields

### 5.1 `unit`

Each current-layer unit truth must record:

1. `system_constraints_ref`
2. `shared_contract_refs`

Each candidate-layer unit main file must additionally record these frontmatter fields:

1. `source_basis`
2. `evidence_appendix_ref`

`unit` does not formally record `scenario_refs`.

### 5.2 `scenario`

Each current-layer scenario truth must record:

1. `repository_mapping_ref`
2. `unit_refs`
3. `shared_contract_refs`
4. `system_constraints_ref`

Each candidate-layer scenario main file must additionally record these frontmatter fields:

1. `source_basis`
2. `evidence_appendix_ref`

### 5.3 `repository_mapping`

`repository_mapping` must record:

1. current `unit` IDs
2. current `scenario` IDs, or `none`
3. current `shared_contract` IDs
4. `system_constraints_ref`

This is repository-structure truth, not lifecycle binding metadata for a command-target object.

### 5.4 `shared_contract`

Each current-layer shared-contract file must record:

1. `shared_contract_id`
2. `layer`
3. `shared_version`
4. `bound_objects`
5. `system_constraints_ref`

Conditional field:

1. when a candidate-layer shared-contract file already has a stable-layer sibling for the same `shared_contract_id`, that candidate file must also record exactly one `promotion_owner_unit`
   - it must be a bare unit id
   - it must name one formal unit from current repository truth
   - it is the only unit round allowed to land that candidate shared file as the next stable-layer Shared Contract file
2. when a candidate-layer shared-contract file does not have a stable-layer sibling, `promotion_owner_unit` must not be recorded
3. stable-layer shared-contract files must not record `promotion_owner_unit`
4. when a command or shared-governance flow explicitly keeps a touched shared-contract file with no current formal bindings as independently authored shared truth, that same file must record exactly these intentional-unbound retention fields:
   - `unbound_retention: intentional`
   - `unbound_retention_reason: <non-empty reason>`
   - `unbound_retention_owner: <owning command or shared-governance flow>`
5. `unbound_retention_owner` must name the command or internal shared-governance flow that owns the terminal-state decision in the current round, for example `unit_fork`, `unit_promote`, or `shared_topology`
6. the intentional-unbound retention fields may be recorded only when `bound_objects=none`
7. when the resulting shared-contract file has one or more formal bound objects, the intentional-unbound retention fields must not be recorded
8. when a file that previously carried intentional-unbound retention becomes formally bound again, the same round that restores the binding must remove `unbound_retention`, `unbound_retention_reason`, and `unbound_retention_owner`
9. intentional-unbound retention fields are terminal-state truth for the shared-contract file only; they do not replace `shared_contract_refs`, do not create a formal binding, and do not skip required `shared_sync` or `impact_sync` reconciliation

## 6. Binding Contracts

### 6.1 Shared Contract Binding Contract

When current-layer truth records `shared_contract_refs`, executors must treat that field as the only formal source of which shared files are currently bound.

Rules:

1. `shared_contract_refs` must name the exact layer and file currently bound
2. stable-layer command-target objects may bind only stable-layer shared truth
3. candidate-layer command-target objects may bind stable-layer or candidate-layer shared truth, but the bound layer must be explicit
4. `bound_objects` is declarative metadata only; it does not replace the command-target object's formal binding source
5. a `bound_objects`-only delta does not by itself invalidate downstream process files
6. `bound_objects` must use typed refs only
   - `unit:<id>`
   - `scenario:<id>`
7. when `shared_contract_refs` is written as a markdown list, executors must normalize the ref order by exact shared ref string in ascending lexical order

### 6.2 Dependency Direction Contract

Formal dependency direction is fixed:

1. `repository_mapping -> unit/scenario/shared_contract`
2. `system_constraints -> shared_contract/unit/scenario/repository_mapping`
3. `shared_contract -> unit/scenario`
4. `unit -> scenario`

Downstream invalidation rule:

1. upstream change may invalidate downstream process files or stable-alignment claims
2. downstream change does not automatically invalidate upstream truth

## 7. Reading Rules

Before any governance action:

1. read the target object's current-layer main file
2. read any explicitly required appendix truth for that object family
3. read bound shared files when `shared_contract_refs` is not empty
4. read `system_constraints.md` when `system_constraints_ref` is part of current truth or when the task requires baseline judgment
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
   - `shared_contract_snapshot`
2. `scenario`
   - `unit_snapshot`
   - `shared_contract_snapshot`

Process files become invalid when their required current truth or required current bindings no longer match.

Candidate evidence appendix files are candidate appendix files for snapshot and invalidation purposes.
Their inclusion in `unit_appendix_snapshot` or the scenario candidate snapshot proves which evidence was reviewed by the gate, but it does not make the evidence appendix an implementation truth source.
Implementation and verification commands must use the candidate main Spec, retained behavior rules, bound Shared Contract files, and `system_constraints.md` as truth.

The exact snapshot field definitions come from `process_snapshot_contract.md`.

## 9. Invalidation And Reconciliation

When upstream truth or binding changes:

1. invalidate downstream process files deterministically
2. fall back the downstream object to the minimum legal next step defined by its command family
3. keep Shared Contract topology reconciliation and generic impact reconciliation separate

Formal routing remains:

1. `shared_sync` for shared-governance downstream discovery
2. `impact_sync` for generic fallback and cleanup once the affected downstream object set is fixed
