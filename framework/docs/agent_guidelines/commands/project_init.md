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

### 2.1 Lifecycle-State Advance Inheritance

当本命令推进 `_status.md` 时，这个推进继承 `specflow/framework/docs/agent_guidelines/command_policy.md` 第 8.5 节定义的 authoritative / non-authoritative 中心契约。
Only a new independent full-scope run of `project_init` may produce that advancing result; later local confirmation or scoped follow-up review must not advance lifecycle state.

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

## 6. Output Contract

The output must report:

1. stable truth file write result
2. `_status.md` registration result
3. lifecycle-state transition result
4. `round conclusion`
5. `current state`
6. `next step`
7. `why this next step`
8. `next-stage entry gap`
9. the `user-facing close-out block` required by `specflow/framework/docs/agent_guidelines/command_policy.md` Section 8.6
10. if a future extension introduces a checkpoint stop, the same close-out block must also report `resume signal`

## 7. Non-Goals

1. creating a candidate project round
2. implementing code changes
