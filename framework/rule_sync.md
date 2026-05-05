# Rule Sync Flow

## 1. Purpose

`rule_sync` is the internal rule-governance flow that resolves one question only:

1. after a `rule` change, which downstream formal objects are truly affected

It is not the generic cleanup engine.
It is the rule-specific impact-discovery and exception-adaptation layer that runs before generic fallback execution.

This file answers five questions:

1. what `rule_sync` owns
2. what it must read before it can decide scope
3. when it must stop and return control to `rule_escape`
4. what it passes to `impact_sync`
5. which rule-specific exceptions it may interpret

`rule_sync` is not a user-facing command.
Users still enter rule work through natural-language routing.

## 2. Scope

`rule_sync` handles only rule-governance questions created by Rule change.

It may:

1. resolve which `rule` files changed or are explicitly in scope
2. rebuild the repository-wide real binding set from downstream truth files
3. detect `bound_objects` metadata drift
4. determine the affected downstream command-object set for:
   - `unit`
   - `scenario`
5. interpret rule-specific execution-local exceptions such as:
   - `current_stable_landing_unit`
   - `stable_landing_rule_refs`
   - `bound_objects_only_rule_file_refs`
6. convert those exceptions into exception-resolved downstream impact input
7. pass the final affected object set to `impact_sync`

It does not:

1. rewrite downstream truth files
2. delete downstream process files directly
3. update `_status.md` directly
4. replace `impact_sync`
5. replace `rule_escape`
6. replace `docs/specs/repository_mapping.md`

## 3. Preconditions

Before `rule_sync` runs:

1. read `specflow/framework/spec_policy.md`
2. read `specflow/framework/command_policy.md`
3. read `specflow/framework/impact_sync_policy.md`
4. read `docs/specs/repository_mapping.md`
5. read `docs/specs/_status.md`
6. read the current in-scope `rule` files under:
   - `docs/specs/rules/stable/`
   - `docs/specs/rules/candidate/`
7. if the current task changed any current-layer truth file or binding source used to derive downstream scope, that writeback must already be present before `rule_sync` computes impact
8. if the current task created, removed, renamed, split, merged, replaced, retired, or otherwise changed the current rule object map, the required `docs/specs/repository_mapping.md` writeback must already be present before `rule_sync` computes impact

Execution-local caller inputs may include:

1. `current_stable_landing_unit`
   - use only when the same round has just landed stable truth and must not invalidate that same landing merely because the new stable rule file now exists
   - this input declares only which stable unit may use the landing exception
2. `stable_landing_rule_refs`
   - use only when the caller can name the exact rule refs written by that same landing round
   - this input is required whenever `current_stable_landing_unit` is present
3. `bound_objects_only_rule_file_refs`
   - use only when the caller has already proven that the current-round delta for those exact Rule files is limited to `bound_objects` metadata

`rule_sync` must not invent either input when the caller did not provide it.

## 4. Procedure

### 4.1 Build current rule view

1. load the current `rule` files from repository truth
2. record for each current file:
   - `rule_id`
   - `layer`
   - `file_ref`
   - `version_ref`
   - current fingerprint
   - declared `bound_objects`
3. verify `docs/specs/repository_mapping.md` against the in-scope rule files when the current task changed the rule object map or rule truth-path rules:
   - every new or remaining current `rule_id` required by the current shared scope must be present in the mapping
   - every retired rule ID resolved by the current round must no longer be listed as a current rule
   - rule truth-path rules must still point to the resulting rule truth locations
   - if the mapping is missing or conflicting, stop `rule_sync`, do not update the mapping here, and return control to `rule_escape`

### 4.2 Rebuild Real Binding Set

1. read current object rows from `docs/specs/_status.md`
2. rebuild the real binding set from downstream formal truth:
   - `unit.rule_refs`
   - `scenario.rule_refs`
3. treat downstream `rule_refs` as the only formal source of which rule files are currently bound
4. treat `bound_objects` only as declarative metadata

### 4.3 Derive Affected Object Set

1. include any downstream object whose current formal `rule_refs` bind:
   - a changed rule file
   - a changed layer of a named `rule_id`
   - a changed rule binding introduced by the current task
2. do not include a sibling layer only because it shares the same `rule_id`
3. apply the dependency direction from `spec_policy.md`:
   - `rule -> unit/scenario`
4. if a rule change also changes the repository object map or path-ownership truth, verify that `docs/specs/repository_mapping.md` already contains the required current truth before handing off to `impact_sync`
5. if repository mapping truth is missing, conflicts with the current shared scope, or is insufficient to determine the real binding set safely:
   - stop `rule_sync`
   - do not perform local fallback cleanup
   - do not write `docs/specs/repository_mapping.md` from `rule_sync`
   - return control to `rule_escape`

### 4.4 Interpret Rule-Specific Exceptions

`rule_sync` may apply only these rule-specific exception rules:

1. `bound_objects`-only exception
   - if a changed Rule file is explicitly declared in `bound_objects_only_rule_file_refs`, do not treat that metadata-only delta as downstream invalidation by itself
2. current stable landing exception
   - if `current_stable_landing_unit` is present, apply the exception only to the exact rule refs explicitly listed in `stable_landing_rule_refs`
   - if any other selected shared ref still invalidates that same unit, do not suppress that invalidation

`rule_sync` must not infer either exception from fingerprint difference alone.

### 4.5 Hand Off To `impact_sync`

After the affected downstream object set and exception set are fixed:

1. convert rule-specific exceptions into already-resolved downstream impact input:
   - final `invalidating_rule_refs`
   - final `explicit_fallback_scope`
   - final `allowed_shared_snapshot_mismatch_file_refs`
2. pass that exception-resolved downstream object set to `impact_sync`
2. let `impact_sync` perform:
   - candidate-side process cleanup
   - `_status.md` fallback writeback
   - stable-side reroute to the correct verify command

`rule_sync` remains responsible for the scope and exception judgment.
`impact_sync` remains responsible for the generic fallback execution.
`impact_sync` must not receive raw `current_stable_landing_unit`, raw `stable_landing_rule_refs`, or raw `bound_objects_only_rule_file_refs`.

## 5. Output Contract

The output must report at least:

1. which `rule` files or ids were treated as changed or in scope
2. the affected downstream object set grouped by:
   - `unit`
   - `scenario`
3. whether `docs/specs/repository_mapping.md` was current for the shared scope before impact handoff, or whether missing mapping truth caused return to `rule_escape`
4. any `bound_objects` metadata drift
5. which execution-local shared exceptions were applied
6. whether control was passed to `impact_sync`
7. whether control was returned to `rule_escape`

## 6. Stop Conditions

`rule_sync` stops only when one of these is true:

1. the affected downstream object set is fully determined and passed to `impact_sync`
2. no affected downstream object exists and the current shared scope has been reconciled
3. repository truth is insufficient and control has been returned to `rule_escape`

## 7. Non-Goals

`rule_sync` does not:

1. create an independent lifecycle for `rule`
2. replace the rule-governance branch
3. replace `impact_sync`
4. write downstream truth files
5. perform generic fallback execution by itself
