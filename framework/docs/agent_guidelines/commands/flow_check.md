# Flow Check Command

## 1. Purpose

`flow_check:{flow}` checks whether the current candidate flow truth is sufficiently closed to constrain later end-to-end verification.

## 2. Preconditions

1. `_status.md` says `Object Type=flow`, `Active Layer=candidate`, `Next Command=flow_check`
2. current candidate flow file exists

## 3. Procedure

1. read current candidate flow truth
2. verify required bindings are explicit:
   - `project_ref`
   - `module_refs`
   - `shared_contract_refs`
   - `system_constraints_stable_ref`
3. verify entry, path, exit, and failure absorption are explicit enough to verify
4. if pass, write `_check_result/{flow}.md` and advance `Next Command=flow_verify`
5. if not pass, keep `Next Command=flow_check`

## 4. Non-Goals

1. implementation planning
2. direct code editing
