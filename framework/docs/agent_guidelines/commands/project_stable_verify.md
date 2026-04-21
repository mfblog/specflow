# Project Stable Verify Command

## 1. Purpose

`project_stable_verify` checks whether current repository truth still aligns with the stable `ProjectSpec`.

## 2. Preconditions

1. `_status.md` says `Object Type=project`, `Active Layer=stable`, `Next Command=project_stable_verify`
2. current stable `ProjectSpec` exists

## 3. Procedure

1. read stable `ProjectSpec`
2. revalidate current `flow`, `module`, `shared_contract`, and `system_constraints` bindings required by that project truth
3. if current bindings still align, keep or advance `Next Command=project_fork`
4. if drift exists, keep `Next Command=project_stable_verify`

## 4. Non-Goals

1. creating candidate project truth
2. mutating downstream object truth
