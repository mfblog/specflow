# Project Fork Command

## 1. Purpose

`project_fork` opens a new candidate `ProjectSpec` from the current stable `ProjectSpec`.

## 2. Lifecycle-State Advance Inheritance

When this command advances `_status.md`, that advancement inherits the authoritative / non-authoritative central contract defined in Section 8.5 of `specflow/framework/docs/agent_guidelines/command_policy.md`.
Only a new independent full-scope run of `project_fork` may produce that advancing result; later local confirmation or scoped follow-up review must not advance lifecycle state.

## 3. Preconditions

1. `_status.md` says `Object Type=project`, `Active Layer=stable`, `Next Command=project_fork`
2. stable `ProjectSpec` exists

## 4. Procedure

1. read stable `ProjectSpec`
2. create or overwrite `docs/specs/project/candidate/c_project.md`
3. carry forward stable bindings as the new candidate starting point
4. delete outdated candidate-side project process files if they exist
5. write `_status.md`:
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=project_check`

## 5. Output Contract

The output must report:

1. candidate truth file write result
2. candidate-side process file cleanup result
3. lifecycle-state transition result
4. `_status.md` update result
5. `round conclusion`
6. `current state`
7. `next step`
8. `why this next step`
9. `next-stage entry gap`
10. the `user-facing close-out block` required by `specflow/framework/docs/agent_guidelines/command_policy.md` Section 8.6
11. if a future extension introduces a checkpoint stop, the same close-out block must also report `resume signal`

## 6. Non-Goals

1. stable verification
2. project promotion
