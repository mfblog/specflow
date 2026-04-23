# Flow Promote Command

## 1. Purpose

`scenario_promote:{flow}` promotes the current candidate flow into the new stable flow truth.

## 2. Lifecycle-State Advance Inheritance

When this command advances `_status.md`, that advancement inherits the authoritative / non-authoritative central contract defined in Section 8.5 of `specflow/framework/docs/agent_guidelines/command_policy.md`.
Only a new independent full-scope run of `scenario_promote` may produce that advancing result; later local confirmation or scoped follow-up review must not advance lifecycle state.

## 3. Preconditions

1. `_status.md` says `Object Type=scenario`, `Active Layer=candidate`, `Next Command=scenario_promote`
2. current valid `_verify_result/{flow}.md` exists

## 4. Procedure

1. revalidate current candidate flow truth and current verification coverage
2. write `docs/specs/scenarios/stable/s_scenario_{name}.md`
3. delete `docs/specs/scenarios/candidate/c_scenario_{name}.md`
4. delete current-round flow `_check_result` and `_verify_result`
5. write `_status.md`:
   - `Stable=yes`
   - `Candidate=no`
   - `Active Layer=stable`
   - `Next Command=scenario_fork`

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
