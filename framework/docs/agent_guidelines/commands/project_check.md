# Project Check Command

## 1. Purpose

`project_check` checks whether the current candidate `ProjectSpec` is sufficiently closed to constrain later project verification and promotion.

## 2. Preconditions

1. `_status.md` says `Object Type=project`, `Active Layer=candidate`, `Next Command=project_check`
2. current candidate `ProjectSpec` exists

## 3. Procedure

1. read current candidate `ProjectSpec`
2. verify required bindings are explicit:
   - `flow_refs`
   - `module_refs`
   - `shared_contract_refs`
   - `system_constraints_stable_ref`
3. verify all referenced objects exist at the declared layer
4. if pass, write `_check_result/project.md` and advance `Next Command=project_verify`
5. if not pass, keep `Next Command=project_check`

## 4. Non-Goals

1. implementation planning
2. code implementation
