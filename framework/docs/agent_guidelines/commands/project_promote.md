# Project Promote Command

## 1. Purpose

`project_promote` promotes the current candidate `ProjectSpec` into the new stable `ProjectSpec`.

## 2. Preconditions

1. `_status.md` says `Object Type=project`, `Active Layer=candidate`, `Next Command=project_promote`
2. current valid `_verify_result/project.md` exists

## 3. Procedure

1. revalidate current candidate truth and current project verification coverage
2. write `docs/specs/project/stable/s_project.md`
3. delete `docs/specs/project/candidate/c_project.md`
4. delete current-round project `_check_result` and `_verify_result`
5. write `_status.md`:
   - `Stable=yes`
   - `Candidate=no`
   - `Active Layer=stable`
   - `Next Command=project_fork`

## 4. Non-Goals

1. promoting modules or flows implicitly
2. absorbing `system_constraints` independently
