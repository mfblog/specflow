# Flow Fork Command

## 1. Purpose

`flow_fork:{flow}` opens a new candidate flow round from the current stable flow.

## 2. Lifecycle-State Advance Inheritance

当本命令推进 `_status.md` 时，这个推进继承 `specflow/framework/docs/agent_guidelines/command_policy.md` 第 8.5 节定义的 authoritative / non-authoritative 中心契约。
Only a new independent full-scope run of `flow_fork` may produce that advancing result; later local confirmation or scoped follow-up review must not advance lifecycle state.

## 3. Preconditions

1. `_status.md` says `Object Type=flow`, `Active Layer=stable`, `Next Command=flow_fork`
2. stable flow truth exists

## 4. Procedure

1. read stable flow truth
2. create or overwrite `docs/specs/flows/candidate/c_flow_{name}.md`
3. carry forward stable bindings
4. delete outdated candidate-side flow process files if they exist
5. write `_status.md`:
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=flow_check`

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
2. flow promotion
