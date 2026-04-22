# Project Promote Command

## 1. Purpose

`project_promote` promotes the current candidate `ProjectSpec` into the new stable `ProjectSpec`.

## 2. Lifecycle-State Advance Inheritance

当本命令推进 `_status.md` 时，这个推进继承 `specflow/framework/docs/agent_guidelines/command_policy.md` 第 8.5 节定义的 authoritative / non-authoritative 中心契约。
Only a new independent full-scope run of `project_promote` may produce that advancing result; later local confirmation or scoped follow-up review must not advance lifecycle state.

## 3. Preconditions

1. `_status.md` says `Object Type=project`, `Active Layer=candidate`, `Next Command=project_promote`
2. current valid `_verify_result/project.md` exists

## 4. Procedure

1. revalidate current candidate truth and current project verification coverage
2. write `docs/specs/project/stable/s_project.md`
3. delete `docs/specs/project/candidate/c_project.md`
4. delete current-round project `_check_result` and `_verify_result`
5. write `_status.md`:
   - `Stable=yes`
   - `Candidate=no`
   - `Active Layer=stable`
   - `Next Command=project_fork`

## 5. Output Contract

The output must report:

1. stable truth file write result
2. candidate truth file delete result
3. `_check_result/project.md` and `_verify_result/project.md` cleanup result
4. lifecycle-state transition result
5. `_status.md` update result
6. `round conclusion`
7. `current state`
8. `next step`
9. `why this next step`
10. `next-stage entry gap`
11. the `user-facing close-out block` required by `specflow/framework/docs/agent_guidelines/command_policy.md` Section 8.6
12. if a future extension introduces a checkpoint stop, the same close-out block must also report `resume signal`

## 6. Non-Goals

1. promoting modules or flows implicitly
2. absorbing `system_constraints` independently
