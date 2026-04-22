# Flow Check Command

## 1. Purpose

`flow_check:{flow}` checks whether the current candidate flow truth is sufficiently closed to constrain later end-to-end verification.

## 2. Lifecycle-State Advance Inheritance

当本命令推进 `_status.md` 时，这个推进继承 `specflow/framework/docs/agent_guidelines/command_policy.md` 第 8.5 节定义的 authoritative / non-authoritative 中心契约。
Only a new independent full-scope run of `flow_check` may produce that advancing result; later repair confirmation or scoped follow-up review must not advance lifecycle state.

## 3. Preconditions

1. `_status.md` says `Object Type=flow`, `Active Layer=candidate`, `Next Command=flow_check`
2. current candidate flow file exists

## 4. Procedure

1. read current candidate flow truth
2. verify required bindings are explicit:
   - `project_ref`
   - `module_refs`
   - `shared_contract_refs`
   - `system_constraints_stable_ref`
3. verify entry, path, exit, and failure absorption are explicit enough to verify
4. if pass, write `_check_result/{flow}.md` and advance `Next Command=flow_verify`
5. if not pass, keep `Next Command=flow_check`

## 5. Output Contract

The output must report:

1. `check gate result`
2. `_check_result/{flow}.md` write, delete, or keep result
3. `_status.md` update result
4. `round conclusion`
5. `current state`
6. `next step`
7. `why this next step`
8. `next-stage entry gap`
9. the `user-facing close-out block` required by `specflow/framework/docs/agent_guidelines/command_policy.md` Section 8.6
10. if a future extension introduces a checkpoint stop, the same close-out block must also report `resume signal`

## 6. Non-Goals

1. implementation planning
2. direct code editing
