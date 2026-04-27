# Scenario Check Command

## 1. Purpose

`scenario_check:{scenario}` checks whether the current candidate scenario truth is sufficiently closed to constrain later end-to-end verification.

## 2. Lifecycle-State Advance Inheritance

When this command advances `_status.md`, that advancement inherits the authoritative / non-authoritative central contract defined in Section 8.5 of `specflow/framework/command_policy.md`.
Only a new independent full-scope run of `scenario_check` may produce that advancing result; later repair confirmation or scoped follow-up review must not advance lifecycle state.

## 3. Preconditions

1. `_status.md` says `Object Type=scenario`, `Active Layer=candidate`, `Next Command=scenario_check`
2. current candidate scenario file exists
3. read `specflow/framework/candidate_handoff_contract.md`
4. read `specflow/framework/onboarding_decision_policy.md`
5. if `_check_result/scenario/{scenario}.md`, `_status.md`, candidate truth, or other commit-triggering governance files may change, read the git policy first

## 4. Procedure

1. read current candidate scenario truth and `docs/specs/repository_mapping.md`
2. verify required bindings are explicit:
   - `source_basis`
   - `evidence_appendix_ref`
   - `repository_mapping_ref`
   - `unit_refs`
   - `shared_contract_refs`
   - `system_constraints_ref`
3. process candidate source fields using `onboarding_decision_policy.md`:
   - if `source_basis=existing_implementation` or `source_basis=mixed`, `evidence_appendix_ref` must point to an existing scenario evidence appendix and that appendix must be read
   - if `source_basis=new_design` or `source_basis=replacement`, `evidence_appendix_ref` must be `none`
   - evidence appendix conflicts or unknowns that still affect selected scenario behavior block pass unless the candidate scenario main Spec explicitly makes a bounded selected rule that no longer depends on them
4. verify `repository_mapping_ref` matches the current repository mapping
5. verify entry, path, exit, and failure absorption are explicit enough to verify
6. if pass, write `_check_result/scenario/{scenario}.md` so it satisfies the `scenario_check -> scenario_verify` handoff in `specflow/framework/candidate_handoff_contract.md`, then advance `Next Command=scenario_verify`
7. if not pass, keep `Next Command=scenario_check`
8. perform git close-out if required

## 5. Output Contract

The output must report:

1. `check gate result`
2. candidate source and evidence appendix result
3. `_check_result/scenario/{scenario}.md` write, delete, or keep result
4. `_status.md` update result
5. `round conclusion`
6. `current state`
7. `next step`
8. `why this next step`
9. `next-stage entry gap`
10. git close-out result
11. the `user-facing close-out block` required by `specflow/framework/command_policy.md` Section 8.6
12. if a future extension introduces a checkpoint stop, the same close-out block must also report `resume signal`

## 6. Non-Goals

1. implementation planning
2. direct code editing
