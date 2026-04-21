# Flow Verify Command

## 1. Purpose

`flow_verify:{flow}` verifies whether the current candidate flow is actually wired from entry to outcome under the current bound object set.

## 2. Preconditions

1. `_status.md` says `Object Type=flow`, `Active Layer=candidate`, `Next Command=flow_verify`
2. current valid `_check_result/{flow}.md` exists

## 3. Procedure

1. read current candidate flow truth
2. revalidate current bound module, shared, and baseline snapshots
3. verify the business path from entry to claimed outcome
4. report `affected_modules` when implementation work is still required downstream
5. if pass, write `_verify_result/{flow}.md` and advance `Next Command=flow_promote`
6. if bindings drifted, fall back to `flow_check`

## 4. Non-Goals

1. replacing `cand_impl:{module}`
2. implicitly repairing affected modules
