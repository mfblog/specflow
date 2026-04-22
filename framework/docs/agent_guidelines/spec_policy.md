# Spec-Driven Development Policy

## 1. Purpose

This file defines the formal truth objects used by `specFlow` in this repository.

It answers four questions:

1. which formal objects exist
2. which files carry those objects
3. how executors must read those objects before governance actions
4. how bindings, snapshots, and invalidation are anchored to those objects

## 2. Core Object Families

### 2.1 Command-Target Objects

This repository has three command-target truth object families:

1. `module`
2. `flow`
3. `project`

They all:

1. support `stable` and `candidate` layers
2. enter `docs/specs/_status.md`
3. may produce process files when their command family explicitly allows them

Differences:

1. only `module` owns implementation planning and implementation work
2. `flow` owns business-chain truth and business-chain verification, but not implementation planning
3. `project` owns project-topology truth and project-topology verification, but not implementation planning

### 2.2 `system_constraints`

`system_constraints` is the unique global system-constraint object.

It answers:

1. what the current project-wide technical baseline is
2. which shared mechanisms are formally preferred
3. which default engineering choices are formally active
4. which global prohibitions or explicit exceptions exist

It does not answer:

1. one module's internal state machine
2. one flow's path semantics
3. one project's topology detail

It has only one formal effective file:

1. `docs/specs/system/stable/s_system_constraints.md`

It is not a command target.

### 2.3 `shared_contract`

`shared_contract` files are independent shared truth objects reused by multiple command-target objects.

They answer:

1. which shared local protocol or shared semantics multiple formal objects reuse now
2. which exact layer and file currently carries that shared truth

They do not answer:

1. full business-chain truth
2. project-topology truth
3. global default-rule truth

They are not standard command targets.
Users enter shared work through `shared_ops:{natural-language request}`.

## 3. Files

### 3.1 `module`

1. `stable` -> `docs/specs/modules/stable/s_{module}.md`
2. `candidate` -> `docs/specs/modules/candidate/c_{module}.md`

### 3.2 `flow`

1. `stable` -> `docs/specs/flows/stable/s_flow_{name}.md`
2. `candidate` -> `docs/specs/flows/candidate/c_flow_{name}.md`

### 3.3 `project`

1. `stable` -> `docs/specs/project/stable/s_project.md`
2. `candidate` -> `docs/specs/project/candidate/c_project.md`

### 3.4 `shared_contract`

1. `stable` -> `docs/specs/shared_contracts/stable/*.md`
2. `candidate` -> `docs/specs/shared_contracts/candidate/*.md`

### 3.5 `_status.md`

`docs/specs/_status.md` is the formal object-state index file.

It records rows for:

1. `project`
2. `flow`
3. `module`

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

### 4.1 `module`

`module` truth answers:

1. module goal and boundary
2. module protocols
3. module state transitions
4. module-local acceptance criteria

### 4.2 `flow`

`flow` truth answers:

1. entry
2. path across modules
3. success outcome
4. failure absorption
5. end-to-end verification expectation

### 4.3 `project`

`project` truth answers:

1. what the project is
2. which formal `module` objects belong to the project
3. which formal `flow` objects belong to the project
4. which formal `shared_contract` objects are formally reused by the project surface
5. which stable `system_constraints` version currently constrains the project
6. how those objects connect at topology level

### 4.4 `shared_contract`

`shared_contract` truth answers:

1. one shared reusable local truth
2. not the whole flow
3. not the whole project
4. not the whole global baseline

## 5. Required Binding Fields

### 5.1 `module`

Each current-layer module truth must record:

1. `system_constraints_stable_ref`
2. `shared_contract_refs`

`module` does not formally record `flow_refs`.

### 5.2 `flow`

Each current-layer flow truth must record:

1. `project_ref`
2. `module_refs`
3. `shared_contract_refs`
4. `system_constraints_stable_ref`

### 5.3 `project`

Each current-layer project truth must record:

1. `flow_refs`
2. `module_refs`
3. `shared_contract_refs`
4. `system_constraints_stable_ref`

## 6. Binding Contracts

### 6.1 Shared Contract Binding Contract

When current-layer truth records `shared_contract_refs`, executors must treat that field as the only formal source of which shared files are currently bound.

Rules:

1. `shared_contract_refs` must name the exact layer and file currently bound
2. stable-layer command-target objects may bind only stable-layer shared truth
3. candidate-layer command-target objects may bind stable-layer or candidate-layer shared truth, but the bound layer must be explicit
4. `bound_modules` is declarative metadata only; it does not replace the command-target object's formal binding source
5. a `bound_modules`-only delta does not by itself invalidate downstream process files
6. when `shared_contract_refs` is written as a markdown list, executors must normalize the ref order by exact shared ref string in ascending lexical order
7. the ordering contract applies to the underlying shared ref values, not to whether a list item is rendered with backticks

### 6.2 Dependency Direction Contract

Formal dependency direction is fixed:

1. `system_constraints -> shared_contract/module/flow/project`
2. `shared_contract -> module/flow/project`
3. `module -> flow/project`
4. `flow -> project`

Downstream invalidation rule:

1. upstream change may invalidate downstream process files or stable-alignment claims
2. downstream change does not automatically invalidate upstream truth

## 7. Reading Rules

Before any governance action:

1. read the target object's current-layer main file
2. read any explicitly required appendix truth for that object family
3. read bound shared files when `shared_contract_refs` is not empty
4. read `s_system_constraints.md` when `system_constraints_stable_ref` is part of the current truth or when the task requires baseline judgment

Additional rules:

1. do not guess bindings by scanning unrelated files first
2. do not treat natural-language mentions as formal bindings
3. do not skip explicitly bound current-layer truth

## 8. Process Files And Snapshots

Process files are not behavior truth.
They are current-round derived artifacts.

Process containers by object family:

1. `module`
   - `_check_result`
   - `_plans`
   - `_verify_result`
2. `flow`
   - `_check_result`
   - `_verify_result`
3. `project`
   - `_check_result`
   - `_verify_result`

Process files become invalid when their required current truth or required current bindings no longer match.

The exact snapshot field definitions come from `process_snapshot_contract.md`.

## 9. Invalidation And Reconciliation

Shared rules:

1. invalid candidate-side process files fall back to the smallest legal `check`-side step for that object family
2. invalid stable-layer alignment falls back to the matching stable verification command
3. downstream invalidation and fallback execution may be handled by `impact_sync` once the affected downstream set is already fixed
4. shared-governance uncertainty must still be handled through `shared_ops` and its internal shared flows rather than by `impact_sync`

## 10. Non-Goals

This file does not:

1. define command procedures in place of command files
2. create an independent command chain for `shared_contract`
3. create an independent command chain for `system_constraints`
4. allow executors to bypass current truth by starting from code
