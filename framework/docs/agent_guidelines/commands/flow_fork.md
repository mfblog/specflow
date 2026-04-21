# Flow Fork Command

## 1. Purpose

`flow_fork:{flow}` opens a new candidate flow round from the current stable flow.

## 2. Preconditions

1. `_status.md` says `Object Type=flow`, `Active Layer=stable`, `Next Command=flow_fork`
2. stable flow truth exists

## 3. Procedure

1. read stable flow truth
2. create or overwrite `docs/specs/flows/candidate/c_flow_{name}.md`
3. carry forward stable bindings
4. delete outdated candidate-side flow process files if they exist
5. write `_status.md`:
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=flow_check`

## 4. Non-Goals

1. stable verification
2. flow promotion
