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

This repository has three command-target truth object families:

1. `unit`
2. `scenario`
3. `project`

Shared rules:

1. all three families support `stable` and `candidate`
2. all three families enter `docs/specs/_status.md`
3. only these three families are standard command targets

Family differences:

1. `unit` is the minimal governed unit and is the only family that owns implementation planning and implementation work
2. `scenario` owns trigger-to-outcome chain truth and end-to-end verification, but not implementation planning
3. `project` owns governed-unit definition, support-surface rules, topology mapping, and the current formal object graph, but not implementation planning

### 2.2 `system_constraints`

`system_constraints` is the unique global system-constraint object.

It answers:

1. what the current repository-wide engineering baseline is
2. which global default rules are formally active
3. which global prohibitions or explicit exceptions exist
4. which shared mechanisms have already been absorbed into the global baseline

It does not answer:

1. one unit's local behavior truth
2. one scenario's trigger-to-outcome chain
3. one project's topology mapping detail
4. one shared contract's local reusable rule text

It has one effective file only:

1. `docs/specs/system_constraints/stable/s_system_constraints.md`

It is not a command target.

### 2.3 `shared_contract`

`shared_contract` is an independent shared local-truth object reused by multiple downstream formal objects.

It answers:

1. which local reusable rule is currently shared
2. which exact layer and file carry that shared rule now
3. which formal objects are currently bound to it through declarative metadata

It does not answer:

1. a whole unit
2. a whole scenario
3. the whole project model
4. the whole global baseline

It is not a standard command target.
Users enter shared work through `shared_ops:{natural-language request}`.

## 3. Identity And Files

### 3.1 Object Identity

Formal object identity uses the following rules:

1. `_status.md` records bare object IDs
   - `agent`
   - `ai`
   - `task_execution`
   - `project`
2. file names still carry object family prefixes
   - `c_unit_agent.md`
   - `s_unit_ai.md`
   - `c_scenario_task_execution.md`
3. `bound_objects` in Shared Contract files must use typed refs
   - `unit:ai`
   - `scenario:task_execution`
   - `project:project`

### 3.2 `unit`

1. `stable` -> `docs/specs/units/stable/s_unit_{unit}.md`
2. `candidate` -> `docs/specs/units/candidate/c_unit_{unit}.md`

### 3.3 `scenario`

1. `stable` -> `docs/specs/scenarios/stable/s_scenario_{scenario}.md`
2. `candidate` -> `docs/specs/scenarios/candidate/c_scenario_{scenario}.md`

### 3.4 `project`

1. `stable` -> `docs/specs/project/stable/s_project.md`
2. `candidate` -> `docs/specs/project/candidate/c_project.md`

### 3.5 `shared_contract`

1. `stable` -> `docs/specs/shared_contracts/stable/*.md`
2. `candidate` -> `docs/specs/shared_contracts/candidate/*.md`

### 3.6 `_status.md`

`docs/specs/_status.md` is the formal object-state index file.

It records rows for:

1. `unit`
2. `scenario`
3. `project`

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
3. project mapping rules
4. direct implementation editing

### 4.3 `project`

`project` is the project governance coordinate-system object.

It answers five mandatory sections:

1. `Governed Unit Definition`
   - what qualifies as a formal `unit`
   - what must become `shared_contract`
   - what stays outside command-target truth
2. `Support Surface Rules`
   - which paths are governed support surfaces rather than command-target objects
3. `Topology Mapping`
   - which roots are governed
   - which paths are ignored
   - which paths map to which current formal object
   - how conflicts are decided
4. `Current Formal Object Graph`
   - current `unit_refs`
   - current `scenario_refs`
   - current `shared_contract_refs`
   - current relations among them
5. `Global Constraint Alignment`
   - which stable `system_constraints` version currently constrains the project

It does not own:

1. unit-local behavior truth
2. shared-contract body text
3. scenario-local chain detail
4. implementation planning or implementation editing

### 4.4 `shared_contract`

`shared_contract` answers:

1. one shared reusable local rule
2. not the whole unit
3. not the whole scenario
4. not the whole project
5. not the whole global baseline

## 5. Required Binding Fields

### 5.1 `unit`

Each current-layer unit truth must record:

1. `system_constraints_stable_ref`
2. `shared_contract_refs`

`unit` does not formally record `scenario_refs`.

### 5.2 `scenario`

Each current-layer scenario truth must record:

1. `project_ref`
2. `unit_refs`
3. `shared_contract_refs`
4. `system_constraints_stable_ref`

### 5.3 `project`

Each current-layer project truth must record:

1. `scenario_refs`
2. `unit_refs`
3. `shared_contract_refs`
4. `system_constraints_stable_ref`

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
   - `project:project`
7. when `shared_contract_refs` is written as a markdown list, executors must normalize the ref order by exact shared ref string in ascending lexical order

### 6.2 Dependency Direction Contract

Formal dependency direction is fixed:

1. `system_constraints -> shared_contract/unit/scenario/project`
2. `shared_contract -> unit/scenario/project`
3. `unit -> scenario/project`
4. `scenario -> project`

Downstream invalidation rule:

1. upstream change may invalidate downstream process files or stable-alignment claims
2. downstream change does not automatically invalidate upstream truth

## 7. Reading Rules

Before any governance action:

1. read the target object's current-layer main file
2. read any explicitly required appendix truth for that object family
3. read bound shared files when `shared_contract_refs` is not empty
4. read `s_system_constraints.md` when `system_constraints_stable_ref` is part of current truth or when the task requires baseline judgment

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
3. `project`
   - `_check_result`
   - `_verify_result`

Object-owned snapshot extensions are fixed by object type:

1. `unit`
   - `unit_appendix_snapshot`
   - `shared_contract_snapshot`
2. `scenario`
   - `unit_snapshot`
   - `shared_contract_snapshot`
3. `project`
   - `scenario_snapshot`
   - `unit_snapshot`
   - `shared_contract_snapshot`

Process files become invalid when their required current truth or required current bindings no longer match.

The exact snapshot field definitions come from `process_snapshot_contract.md`.

## 9. Invalidation And Reconciliation

When upstream truth or binding changes:

1. invalidate downstream process files deterministically
2. fall back the downstream object to the minimum legal next step defined by its command family
3. keep Shared Contract topology reconciliation and generic impact reconciliation separate

Formal routing remains:

1. `shared_sync` for shared-governance downstream discovery
2. `impact_sync` for generic fallback and cleanup once the affected downstream object set is fixed
