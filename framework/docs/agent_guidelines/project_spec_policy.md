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

`ProjectSpec` is the formal project-topology truth object.

It answers:

1. what the project is
2. which formal `module` objects currently belong to the project
3. which formal `flow` objects currently belong to the project
4. which formal `shared_contract` objects are formally reused by the project surface
5. which stable `system_constraints` version currently constrains the project
6. how those objects connect at the project-topology level

It does not answer:

1. one module's internal behavior
2. one flow's step-by-step business semantics
3. one shared object's local protocol text
4. implementation planning or implementation ownership

## 3. Files

`ProjectSpec` uses two version layers:

1. `docs/specs/project/stable/s_project.md`
2. `docs/specs/project/candidate/c_project.md`

Additional rules:

1. `ProjectSpec` is a command-target object, but it is not a module
2. it enters `docs/specs/_status.md` using `Object Type=project`
3. there is exactly one current `ProjectSpec` per repository
4. `project` is the stable command prefix for this object family

## 4. Required Bindings

`ProjectSpec` must record at minimum:

1. `flow_refs`
2. `module_refs`
3. `shared_contract_refs`
4. `system_constraints_stable_ref`

Binding rules:

1. `ProjectSpec` may bind stable or candidate `flow` objects only at the matching project layer
2. `ProjectSpec stable` must not bind candidate-layer `flow` or candidate-layer `shared_contract` truth
3. `ProjectSpec candidate` may bind candidate-layer `flow` or candidate-layer `shared_contract` truth, but the bound layer must be explicit
4. `ProjectSpec` is downstream of `flow`, `module`, `shared_contract`, and `system_constraints`

## 5. Lifecycle Responsibility

`ProjectSpec` owns:

1. project-topology closure
2. project-topology verification against current bound objects
3. promotion of candidate project topology into stable project topology

It does not own:

1. code implementation
2. module implementation planning
3. module implementation verification

Therefore:

1. `project` command family has `check`, `verify`, and `promote`
2. `project` command family does not have `plan` or `impl`

## 6. Invalidation Rules

`ProjectSpec` process files become invalid when any current required binding changes, including:

1. current `ProjectSpec` truth changes
2. any bound `flow` truth, layer, version, or snapshot changes
3. any bound `module` current truth identity set changes
4. any bound `shared_contract` truth, layer, version, or snapshot changes
5. `system_constraints_stable_ref` no longer matches the current stable global baseline

Fallback rules:

1. invalid candidate `ProjectSpec` falls back to `project_check`
2. invalid stable `ProjectSpec` falls back to `project_stable_verify`

## 7. Non-Goals

This file does not:

1. create a project-side implementation chain
2. replace `flow` truth
3. replace `module` truth
4. create an independent lifecycle for `shared_contract` or `system_constraints`
