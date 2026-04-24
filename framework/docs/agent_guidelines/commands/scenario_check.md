# Scenario Check Command

## 1. Purpose

`scenario_check:{scenario}` checks whether the current candidate scenario truth is sufficiently closed to constrain later end-to-end verification.

## 2. Lifecycle-State Advance Inheritance

When this command advances `_status.md`, that advancement inherits the authoritative / non-authoritative central contract defined in Section 8.5 of `specflow/framework/docs/agent_guidelines/command_policy.md`.
Only a new independent full-scope run of `scenario_check` may produce that advancing result; later repair confirmation or scoped follow-up review must not advance lifecycle state.

## 3. Preconditions

1. `_status.md` says `Object Type=scenario`, `Active Layer=candidate`, `Next Command=scenario_check`
2. current candidate scenario file exists

## 4. Procedure

1. read current candidate scenario truth
2. verify required bindings are explicit:
   - `repository_mapping_ref`
   - `unit_refs`
   - `shared_contract_refs`
   - `system_constraints_stable_ref`
3. verify entry, path, exit, and failure absorption are explicit enough to verify
4. if pass, write `_check_result/{scenario}.md` and advance `Next Command=scenario_verify`
5. if not pass, keep `Next Command=scenario_check`

## 5. Output Contract

The output must report:

1. `check gate result`
2. `_check_result/{scenario}.md` write, delete, or keep result
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
