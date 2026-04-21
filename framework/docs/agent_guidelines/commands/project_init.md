# Project Init Command

## 1. Purpose

`project_init` creates the first stable `ProjectSpec` for a historical repository that is entering the project-topology governance model for the first time.

## 2. Scope

It handles:

1. creating `docs/specs/project/stable/s_project.md`
2. registering `project` in `_status.md`
3. recording current stable bindings to `flow`, `module`, `shared_contract`, and `system_constraints`

It does not:

1. create candidate project truth
2. replace module, flow, or shared truth authoring

## 3. Preconditions

1. no current stable `ProjectSpec` exists
2. the repository's current formal object set can be stated safely from current truth
3. read `project_spec_policy.md`, `spec_policy.md`, and `command_policy.md`

## 4. Procedure

1. read current formal `module`, `flow`, `shared_contract`, and `system_constraints` truth
2. write the first stable `ProjectSpec`
3. write or upsert `_status.md` row:
   - `Object Type=project`
   - `Object=project`
   - `Stable=yes`
   - `Candidate=no`
   - `Active Layer=stable`
   - `Next Command=project_stable_verify`

## 5. Stop Conditions

1. the first stable `ProjectSpec` exists
2. project state is registered in `_status.md`

## 6. Non-Goals

1. creating a candidate project round
2. implementing code changes
