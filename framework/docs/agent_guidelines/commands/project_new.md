# Project New Command

## 1. Purpose

`project_new` creates the first candidate `ProjectSpec` for a brand-new project-topology round when no historical stable `ProjectSpec` exists yet.

## 2. Scope

It handles:

1. creating `docs/specs/project/candidate/c_project.md`
2. initializing current candidate bindings
3. registering project candidate state in `_status.md`

## 3. Preconditions

1. no current project row exists in `_status.md`
2. the repository can safely state one initial candidate project topology
3. read `project_spec_policy.md`, `spec_policy.md`, and `command_policy.md`

## 4. Procedure

1. create `c_project.md`
2. initialize `flow_refs`, `module_refs`, `shared_contract_refs`, and `system_constraints_stable_ref`
3. write or upsert `_status.md` row:
   - `Object Type=project`
   - `Object=project`
   - `Stable=no`
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=project_check`

## 5. Non-Goals

1. creating stable project truth
2. replacing `project_init`
