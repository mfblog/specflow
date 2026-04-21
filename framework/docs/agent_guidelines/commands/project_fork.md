# Project Fork Command

## 1. Purpose

`project_fork` opens a new candidate `ProjectSpec` from the current stable `ProjectSpec`.

## 2. Preconditions

1. `_status.md` says `Object Type=project`, `Active Layer=stable`, `Next Command=project_fork`
2. stable `ProjectSpec` exists

## 3. Procedure

1. read stable `ProjectSpec`
2. create or overwrite `docs/specs/project/candidate/c_project.md`
3. carry forward stable bindings as the new candidate starting point
4. delete outdated candidate-side project process files if they exist
5. write `_status.md`:
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=project_check`

## 4. Non-Goals

1. stable verification
2. project promotion
