# Flow New Command

## 1. Purpose

`scenario_new:{flow}` creates the first candidate truth for a brand-new formal flow object.

## 2. Lifecycle-State Advance Inheritance

When this command advances `_status.md`, that advancement inherits the authoritative / non-authoritative central contract defined in Section 8.5 of `specflow/framework/docs/agent_guidelines/command_policy.md`.
Only a new independent full-scope run of `scenario_new` may produce that advancing result; later local confirmation or scoped follow-up review must not advance lifecycle state.

## 3. Preconditions

1. the flow name is clear and non-conflicting
2. no current row for that flow exists in `_status.md`

## 4. Procedure

1. create `docs/specs/scenarios/candidate/c_scenario_{name}.md`
2. initialize:
   - `project_ref`
   - `unit_refs`
   - `shared_contract_refs`
   - `system_constraints_stable_ref`
3. write or upsert `_status.md` row:
   - `Object Type=scenario`
   - `Object=flow_{name}`
   - `Stable=no`
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=scenario_check`

## 5. Output Contract

The output must report:

1. candidate truth file write result
2. `_status.md` registration result
3. lifecycle-state transition result
4. `round conclusion`
5. `current state`
6. `next step`
7. `why this next step`
8. `next-stage entry gap`
9. the `user-facing close-out block` required by `specflow/framework/docs/agent_guidelines/command_policy.md` Section 8.6
10. if a future extension introduces a checkpoint stop, the same close-out block must also report `resume signal`

## 6. Non-Goals

1. creating stable flow truth
2. editing module code
