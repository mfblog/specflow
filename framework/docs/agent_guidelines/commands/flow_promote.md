# Flow Promote Command

## 1. Purpose

`flow_promote:{flow}` promotes the current candidate flow into the new stable flow truth.

## 2. Preconditions

1. `_status.md` says `Object Type=flow`, `Active Layer=candidate`, `Next Command=flow_promote`
2. current valid `_verify_result/{flow}.md` exists

## 3. Procedure

1. revalidate current candidate flow truth and current verification coverage
2. write `docs/specs/flows/stable/s_flow_{name}.md`
3. delete `docs/specs/flows/candidate/c_flow_{name}.md`
4. delete current-round flow `_check_result` and `_verify_result`
5. write `_status.md`:
   - `Stable=yes`
   - `Candidate=no`
   - `Active Layer=stable`
   - `Next Command=flow_fork`

## 4. Non-Goals

1. module promotion
2. project promotion
