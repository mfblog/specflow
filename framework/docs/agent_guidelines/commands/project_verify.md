# Project Verify Command

## 1. Purpose

`project_verify` verifies whether the current candidate `ProjectSpec` still matches the current bound object set and whether promotion is allowed.

## 2. Lifecycle-State Advance Inheritance

当本命令推进 `_status.md` 时，这个推进继承 `specflow/framework/docs/agent_guidelines/command_policy.md` 第 8.5 节定义的 authoritative / non-authoritative 中心契约。
Only a new independent full-scope run of `project_verify` may produce that advancing result; later repair confirmation or scoped follow-up review must not advance lifecycle state.

## 3. Preconditions

1. `_status.md` says `Object Type=project`, `Active Layer=candidate`, `Next Command=project_verify`
2. current valid `_check_result/project.md` exists

## 4. Procedure

1. read current candidate `ProjectSpec`
2. revalidate current bound `flow`, `module`, `shared_contract`, and `system_constraints` snapshots
3. if current bindings and current project topology still align, write `_verify_result/project.md`
4. if ready, advance `Next Command=project_promote`
5. if bindings drifted, fall back to `project_check`

## 5. Output Contract

The output must report:

1. verification gate result
2. `_verify_result/project.md` write, delete, or keep result
3. `_status.md` update result
4. `round conclusion`
5. `current state`
6. `next step`
7. `why this next step`
8. `next-stage entry gap`
9. the `user-facing close-out block` required by `specflow/framework/docs/agent_guidelines/command_policy.md` Section 8.6
10. if a future extension introduces a checkpoint stop, the same close-out block must also report `resume signal`

## 6. Non-Goals

1. repairing downstream objects
2. replacing `flow_verify`
