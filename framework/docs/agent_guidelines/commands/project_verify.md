# Project Verify Command

## 1. Purpose

`project_verify` verifies whether the current candidate `ProjectSpec` still matches the current bound object set and whether promotion is allowed.

## 2. Preconditions

1. `_status.md` says `Object Type=project`, `Active Layer=candidate`, `Next Command=project_verify`
2. current valid `_check_result/project.md` exists

## 3. Procedure

1. read current candidate `ProjectSpec`
2. revalidate current bound `flow`, `module`, `shared_contract`, and `system_constraints` snapshots
3. if current bindings and current project topology still align, write `_verify_result/project.md`
4. if ready, advance `Next Command=project_promote`
5. if bindings drifted, fall back to `project_check`

## 4. Non-Goals

1. repairing downstream objects
2. replacing `flow_verify`
