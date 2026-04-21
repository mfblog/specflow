# Flow New Command

## 1. Purpose

`flow_new:{flow}` creates the first candidate truth for a brand-new formal flow object.

## 2. Preconditions

1. the flow name is clear and non-conflicting
2. no current row for that flow exists in `_status.md`

## 3. Procedure

1. create `docs/specs/flows/candidate/c_flow_{name}.md`
2. initialize:
   - `project_ref`
   - `module_refs`
   - `shared_contract_refs`
   - `system_constraints_stable_ref`
3. write or upsert `_status.md` row:
   - `Object Type=flow`
   - `Object=flow_{name}`
   - `Stable=no`
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=flow_check`

## 4. Non-Goals

1. creating stable flow truth
2. editing module code
