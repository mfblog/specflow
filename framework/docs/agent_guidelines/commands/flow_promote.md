# Flow Promote Command

## 1. Purpose

`flow_promote:{flow}` promotes the current candidate flow into the new stable flow truth.

## 2. Lifecycle-State Advance Inheritance

当本命令推进 `_status.md` 时，这个推进继承 `specflow/framework/docs/agent_guidelines/command_policy.md` 第 8.5 节定义的 authoritative / non-authoritative 中心契约。
Only a new independent full-scope run of `flow_promote` may produce that advancing result; later local confirmation or scoped follow-up review must not advance lifecycle state.

## 3. Preconditions

1. `_status.md` says `Object Type=flow`, `Active Layer=candidate`, `Next Command=flow_promote`
2. current valid `_verify_result/{flow}.md` exists

## 4. Procedure

1. revalidate current candidate flow truth and current verification coverage
2. write `docs/specs/flows/stable/s_flow_{name}.md`
3. delete `docs/specs/flows/candidate/c_flow_{name}.md`
4. delete current-round flow `_check_result` and `_verify_result`
5. write `_status.md`:
   - `Stable=yes`
   - `Candidate=no`
   - `Active Layer=stable`
   - `Next Command=flow_fork`

## 5. Output Contract

The output must report:

1. stable truth file write result
2. candidate truth file delete result
3. `_check_result/{flow}.md` and `_verify_result/{flow}.md` cleanup result
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

1. module promotion
2. project promotion
