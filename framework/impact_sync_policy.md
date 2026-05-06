# Impact Sync Policy

## 1. Purpose

This file defines the internal governance flow `impact_sync`.

It answers six questions:

1. what `impact_sync` is responsible for
2. what it consumes as input
3. what it must write back
4. which object families it may reconcile
5. how it differs from `rule_sync`
6. which centralized snapshot and recovery contracts it must use

`impact_sync` is an internal governance flow.
It is not a user-facing command entry and it is not a second lifecycle.

## 2. Scope

`impact_sync` handles only deterministic downstream invalidation and fallback after an upstream truth or binding change is already known.

It may reconcile:

1. `unit`
2. `scenario`

It does not:

1. decide whether a change belongs to rule governance
2. decide whether a boundary is still unit-local or shared
3. rewrite object truth files
4. judge semantic sufficiency of candidate closure

## 3. Inputs

Before `impact_sync` runs, the caller must already know:

1. which upstream object changed
2. which downstream object set is in scope
3. which exception-resolved generic invalidation input applies, if any
4. the current binding and snapshot state of those downstream objects

`impact_sync` may consume:

1. `_status.md`
2. current object truth files
3. current process files
4. caller-resolved generic inputs such as:
   - `invalidating_rule_refs`
   - `explicit_fallback_scope`
   - `allowed_shared_snapshot_mismatch_file_refs`
   - same-round stable landing retarget fallback after `rule_sync` has already validated the retargeted candidate units and converted them into explicit fallback scope

`impact_sync` must not interpret raw rule-specific exception inputs such as stable landing owners, stable landing refs, or retargeted unit lists.

Before `impact_sync` revalidates any process file or writes any fallback result, the executor must read and apply:

1. `specflow/framework/process_snapshot_contract.md`
2. `specflow/framework/recovery_policy.md`

`impact_sync` must not invent local snapshot shapes, local fingerprint rules, local ordering rules, local cleanup maps, or local fallback targets when those rules are defined by those contracts.

## 4. Writeback Contract

`impact_sync` may write only:

1. `_status.md`
2. candidate-side process file cleanup for invalid downstream objects

Candidate fallback rules:

1. invalid `unit` -> `unit_check`
2. invalid `scenario` -> `scenario_check`

Stable fallback rules:

1. invalid `unit` -> `unit_stable_verify`
2. invalid `scenario` -> `scenario_stable_verify`

Snapshot revalidation rules:

1. rebuild current process snapshots according to `process_snapshot_contract.md`
2. apply the fingerprint, ordering, and exact-comparison rules from that contract
3. apply only caller-resolved generic exceptions that the contract allows, such as `allowed_shared_snapshot_mismatch_file_refs`
4. treat the process file as invalid for downstream use when any required stored field differs from the rebuilt value after allowed exceptions

Candidate-side cleanup rules:

1. when an invalid downstream `unit` falls back to `unit_check`, update `_status.md` to that next step and delete exactly the candidate-side files listed for `unit -> unit_check` in `recovery_policy.md`
2. when an invalid downstream `scenario` falls back to `scenario_check`, update `_status.md` to that next step and delete exactly the candidate-side files listed for `scenario -> scenario_check` in `recovery_policy.md`
3. a cleanup target that is already absent is recorded as an absent cleanup target; it does not create a different fallback state

Stable-side fallback rules:

1. when an invalid stable `unit` falls back to `unit_stable_verify`, update `_status.md` to that next step
2. when an invalid stable `scenario` falls back to `scenario_stable_verify`, update `_status.md` to that next step
3. stable-side fallback must not delete candidate-side process files solely because stable alignment became stale

`impact_sync` output must report:

1. the upstream change it consumed
2. the affected downstream objects
3. the fallback reason code for each invalid downstream object
4. each `_status.md` update
5. candidate cleanup files deleted
6. candidate cleanup targets already absent
7. any object that could not be reconciled and the legal resume owner

## 5. Relationship To `rule_sync`

`rule_sync` remains the rule-governance internal flow.

Responsibility split:

1. `rule_sync` owns rule-specific impact discovery, rule-specific exception handling, exception-to-generic-input conversion, and rule-governance stop conditions
2. `impact_sync` owns the generic downstream invalidation and fallback execution once the affected object set is already fixed

Therefore:

1. `rule_sync` may call `impact_sync`
2. `impact_sync` must not replace `rule_sync` as the rule-governance intent or boundary flow

## 6. Non-Goals

`impact_sync` does not:

1. create a new user-facing command
2. replace the rule-governance branch
3. replace `unit_check` or `scenario_check`
4. replace `unit_stable_verify` or `scenario_stable_verify`
