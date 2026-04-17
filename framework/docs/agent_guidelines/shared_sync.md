# Shared Sync Flow

## 1. Purpose

`shared_sync` is the internal flow that closes impact after `shared_contract` truth changes.

It answers three questions:

1. which modules still hold outdated shared bindings or snapshots
2. what the smallest actionable fallback step is for each affected module
3. which candidate-side process files must be deleted so invalid gates cannot be reused

This is not the user-facing command entry.
The user reaches it through `shared_ops:{natural-language request}`.

---

## 2. Scope

By default it handles:

1. version, body, layer, or binding changes under `docs/specs/shared_contracts/candidate/` and `docs/specs/shared_contracts/stable/`
2. whether each module's current-layer `Global Constraint Alignment.shared_contract_refs` and process-file snapshots still match the current shared truth
3. unified fallback and cleanup for `_status.md`, `_check_result`, `_plans`, and `_verify_result` when modules still hold old snapshots
4. reporting `bound_modules` drift that must be fixed in the command responsible for the binding change

It does not:

1. rewrite module body content
2. replace `spec_flow_review`
3. replace `cand_check`, `stable_verify`, or `cand_promote`
4. decide whether a `shared_contract` should become `system_constraints`

---

## 3. Preconditions

Before execution:

1. read `specflow/framework/docs/agent_guidelines/spec_policy.md`
2. read `specflow/framework/docs/agent_guidelines/command_policy.md`
3. read `docs/specs/_status.md`
4. read the current `shared_contract` files under `docs/specs/shared_contracts/candidate/` and `docs/specs/shared_contracts/stable/`
5. identify the affected-object source:
   - if the current task changed `docs/specs/shared_contracts/**`, use those changed files to resolve the changed `shared_contract_id` set first
   - if the current task changed any module's `shared_contract_refs`, include those modules directly in the affected-module set
   - if `shared_ops` routed in from a named shared target, resolve that target first
6. scan `_status.md` and build the repository-wide current-layer module set
7. if a shared file changed or a shared target was named, read every current-layer module main file needed to derive the real binding set from `shared_contract_refs`
8. read each affected module's current-layer main file according to `_status.md`
9. if the task may modify `_status.md`, process files, or other commit-triggering governance files, read the Git closure rules first
10. if the current task changed which layer now carries a directly affected module's formal binding source, `_status.md` must already be updated before this flow builds the current-layer module set

---

## 4. Procedure

1. build the current `shared_contract` view:
   - read the files that currently exist under `docs/specs/shared_contracts/candidate/` and `docs/specs/shared_contracts/stable/`
   - record each object's `shared_contract_id`, `layer`, `shared_version`, current body fingerprint, and `bound_modules`
2. build the repository-wide module current-layer binding index from the already-updated `_status.md`:
   - enumerate formal modules from `_status.md`
   - read `Global Constraint Alignment.shared_contract_refs` from each module's current-layer main file needed for binding resolution
   - treat module `shared_contract_refs` as the only formal source of which modules currently bind which shared truth
   - treat `bound_modules` only as declarative metadata
3. derive the affected module set:
   - include modules whose `shared_contract_refs` point to any changed or named `shared_contract_id`
   - include modules whose own `shared_contract_refs` changed in the current task
   - do not use `bound_modules` as the sole source for deciding which modules are affected
4. build the module current snapshot view:
   - if the module is at `candidate`, read any existing `_check_result/{module}.md`, `_plans/{module}.md`, and `_verify_result/{module}.md`
   - extract `shared_contract_snapshot` from those files when present
   - rebuild the normalized snapshot according to `process_snapshot_contract.md`
5. for each affected module, judge whether its shared binding is still valid:
   - if `shared_contract_refs=none` and the module is not in a changed-binding case, leave it unchanged
   - treat the binding as invalid if the referenced file is missing, the layer mismatches, the version reference mismatches, or the module-to-shared relation changed
   - for `candidate` modules, also treat it as invalid if any existing process file's `shared_contract_snapshot` differs from the rebuilt snapshot, except when the delta comes only from `bound_modules`
   - for `stable` modules, also treat it as invalid if the stable shared truth changed enough that "still aligned with stable" can no longer be claimed safely
6. for invalid `candidate` modules:
   - delete `_check_result/{module}.md`
   - delete `_plans/{module}.md`
   - delete `_verify_result/{module}.md`
   - set `Next Command=cand_check` in `_status.md`
   - keep `Candidate=yes` and `Active Layer=candidate`
7. for invalid `stable` modules:
   - do not generate candidate-side file deletions
   - set `Next Command=stable_verify` in `_status.md`
   - keep `Stable=yes` and `Active Layer=stable`
8. if `bound_modules` differs from the real binding set implied by the repository-wide binding index:
   - report governance drift
   - name the command or change owner that must fix it
   - do not invalidate modules on a `bound_modules`-only delta
9. if `_status.md` points to a step later than the real smallest actionable step, correct it
10. finish git close-out if required by policy

---

## 5. Stop Conditions

Stop when one of the following is true:

1. all affected modules have been judged and required fallback or cleanup has been completed
2. no affected modules exist and repository shared state has still been reconciled against current truth
3. repository truth is insufficient to determine the real binding set and the active flow must stop instead of guessing

---

## 6. Output Contract

The output must include at least:

1. a summary of shared-truth changes
2. the list of affected modules
3. the fallback result for each affected module
4. the list of deleted process files
5. any mismatch between `bound_modules` and the real binding set
6. the standardized `fallback_reason_code` for each affected module
7. the git close-out result

Allowed `fallback_reason_code` values:

1. `shared_contract_drift`
2. `binding_drift`
3. `truth_drift`

---

## 7. Non-Goals

`shared_sync` does not:

1. create an independent lifecycle for `shared_contract`
2. directly modify module body text or `shared_contract` body text "just to fix the binding"
3. replace `cand_check` to re-pass candidate closure
4. replace `stable_verify` to re-check stable alignment
5. absorb `shared_contract` conclusions into `system_constraints`
