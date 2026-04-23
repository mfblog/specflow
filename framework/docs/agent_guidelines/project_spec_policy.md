# Project Spec Policy

## 1. Purpose

This file defines what `ProjectSpec` is in this repository and how it relates to other formal objects.

It answers five questions:

1. what `ProjectSpec` formally owns
2. what `ProjectSpec` must not own
3. which files carry `ProjectSpec`
4. which bindings `ProjectSpec` must record
5. how `ProjectSpec` participates in lifecycle and invalidation

## 2. Object Definition

`ProjectSpec` is the formal project governance coordinate-system object.

It must answer these five mandatory sections:

1. `Governed Unit Definition`
   - what qualifies as a formal `unit`
   - what must be promoted into `shared_contract`
   - what remains outside command-target truth
2. `Support Surface Rules`
   - which paths are governed support surfaces rather than command-target objects
3. `Topology Mapping`
   - governed roots
   - ignore rules
   - unit/shared/support ownership rules
   - conflict resolution order
4. `Current Formal Object Graph`
   - current `unit_refs`
   - current `scenario_refs`
   - current `shared_contract_refs`
   - the currently active relation graph among them
5. `Global Constraint Alignment`
   - which stable `system_constraints` version constrains the project now

It does not answer:

1. one unit's local behavior
2. one scenario's local chain semantics
3. one shared-contract body's field-level rule text
4. implementation planning or implementation ownership

## 3. Files

`ProjectSpec` uses two version layers:

1. `docs/specs/project/stable/s_project.md`
2. `docs/specs/project/candidate/c_project.md`

Additional rules:

1. `ProjectSpec` is a command-target object, but it is not a unit
2. it enters `docs/specs/_status.md` using `Object Type=project`
3. there is exactly one current `ProjectSpec` per repository
4. `project` remains the stable command prefix for this object family

## 4. Required Bindings

`ProjectSpec` must record at minimum:

1. `scenario_refs`
2. `unit_refs`
3. `shared_contract_refs`
4. `system_constraints_stable_ref`

Binding rules:

1. `ProjectSpec stable` must not bind candidate-layer `scenario` or candidate-layer `shared_contract` truth
2. `ProjectSpec candidate` may bind candidate-layer `scenario` or candidate-layer `shared_contract` truth, but the bound layer must be explicit
3. `ProjectSpec` is downstream of `scenario`, `unit`, `shared_contract`, and `system_constraints`
4. `ProjectSpec` is the only formal object that may define support-surface ownership rules for the repository

## 5. Lifecycle Responsibility

`ProjectSpec` owns:

1. governed-unit-definition closure
2. support-surface-rule closure
3. topology-mapping closure
4. current formal object-graph verification
5. promotion of candidate project truth into stable project truth

It does not own:

1. code implementation
2. unit implementation planning
3. unit implementation verification
4. scenario implementation repair

Therefore:

1. `project` command family has `check`, `verify`, and `promote`
2. `project` command family does not have `plan` or `impl`

## 6. Invalidation Rules

`ProjectSpec` process files become invalid when any current required binding changes, including:

1. current `ProjectSpec` truth changes
2. any bound `scenario` truth, layer, version, or snapshot changes
3. any bound `unit` identity set, truth, or snapshot changes
4. any bound `shared_contract` truth, layer, version, or snapshot changes
5. `system_constraints_stable_ref` no longer matches the current stable global baseline
6. the project's own topology mapping now resolves a governed path differently

Fallback rules:

1. invalid candidate `ProjectSpec` falls back to `project_check`
2. invalid stable `ProjectSpec` falls back to `project_stable_verify`

## 7. Non-Goals

This file does not:

1. create a project-side implementation chain
2. replace `scenario` truth
3. replace `unit` truth
4. create an independent lifecycle for `shared_contract` or `system_constraints`
