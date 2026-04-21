# Flow Stable Verify Command

## 1. Purpose

`flow_stable_verify:{flow}` checks whether current repository truth still aligns with the stable flow truth.

## 2. Preconditions

1. `_status.md` says `Object Type=flow`, `Active Layer=stable`, `Next Command=flow_stable_verify`
2. current stable flow file exists

## 3. Procedure

1. read stable flow truth
2. revalidate current bound modules, shared contracts, and stable baseline
3. if still aligned, advance `Next Command=flow_fork`
4. if drift exists, keep `Next Command=flow_stable_verify`

## 4. Non-Goals

1. flow candidate authoring
2. module implementation repair
