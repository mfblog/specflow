# Shared Appendix State Reconciliation Flow

## 1. Purpose

This flow reconciles module states and invalid process files after Shared Appendix truth changes.

It answers three questions:

1. which modules still hold old binding snapshots that no longer match current Shared Appendix truth
2. what the smallest actionable fallback step is for each affected module
3. which candidate-side process files must be deleted so invalid gates cannot keep being reused

This flow is not a normal module command, is not in `{command}:{module}` form, and does not enter `docs/specs/_status.md`.
It is an internal governance flow that the agent runs after Shared Appendix changes when needed.

## 2. Scope

By default it handles:

1. version, body, layer, or binding changes under `docs/specs/shared/candidate/` and `docs/specs/shared/stable/`
2. whether each module's current-layer `Global Constraint Alignment.shared_appendix_refs` and process-file snapshots still match the current Shared Appendix state
3. unified fallback and cleanup for `_status.md`, `_check_result`, `_plans`, and `_verify_result` when modules still hold old snapshots
4. reporting Shared Appendix `bound_modules` drift that must be fixed in the command responsible for the binding change

It does not:

1. rewrite module body content
2. replace `spec_flow_review`
3. replace `cand_check`, `stable_verify`, or `cand_promote`

## 3. Preconditions

Before execution:

1. read `specflow/framework/docs/agent_guidelines/spec_policy.md`, `specflow/framework/docs/agent_guidelines/command_policy.md`, and `docs/specs/_status.md`
2. read the current Shared Appendix files under `docs/specs/shared/candidate/` and `docs/specs/shared/stable/`
3. identify the affected-object source:
   - if the current task changed `docs/specs/shared/**`, use those changed Shared Appendix files as the review set
   - if the task did not change `docs/specs/shared/**` but changed any module's `shared_appendix_refs`, build the binding-review set from those modules
   - if the user names a specific Shared Appendix, use that named set
4. read each formal module's current-layer main file according to `_status.md`; if `shared_appendix_refs` is not empty, read the corresponding Shared Appendix files too
5. if the task will modify `_status.md`, `docs/specs/_check_result/*.md`, `docs/specs/_plans/*.md`, `docs/specs/_verify_result/*.md`, or any other commit-triggering files, read the Git closure rules first
6. if the current standard command already recalculated and rewrote fresh Shared Appendix bindings for a target module, or already closed that target module directly to new stable truth, mark that module as "already closed directly in this round" so this flow does not mechanically fall it back again

## 4. Procedure

1. Build the current Shared Appendix view:
   - use the files that currently exist under `docs/specs/shared/candidate/` and `docs/specs/shared/stable/`
   - record each shared object's `shared_id`, `layer`, `shared_version`, current body fingerprint, and `bound_modules`
2. Build the module current-layer binding view:
   - read every formal module from `_status.md`
   - read `Global Constraint Alignment.shared_appendix_refs` from each module's current-layer main file
   - treat module `shared_appendix_refs` as the formal binding source; treat `bound_modules` only as declarative help
3. Build the module current-snapshot view:
   - if the module is at `candidate`, read any existing `_check_result/{module}.md`, `_plans/{module}.md`, and `_verify_result/{module}.md`
   - extract `shared_appendix_snapshot` from those files when present
   - regenerate the normalized snapshot from current truth according to `process_snapshot_contract.md`
4. For each module, judge whether its Shared Appendix binding is still valid:
   - skip modules already marked as directly closed in this round
   - if `shared_appendix_refs=none` and the module is not in a changed-binding case, leave it unchanged
   - treat the binding as invalid if the referenced Shared Appendix file is missing, the layer mismatches, the version reference mismatches, or the module-to-shared binding relation changed
   - for `candidate` modules, also treat it as invalid if any existing process file's `shared_appendix_snapshot` differs from the freshly normalized snapshot
   - for `stable` modules, also treat it as invalid if the stable Shared Appendix truth changed enough that "still aligned with stable" can no longer be claimed safely
5. For invalid `candidate` modules:
   - delete `docs/specs/_check_result/{module}.md`
   - delete `docs/specs/_plans/{module}.md`
   - delete `docs/specs/_verify_result/{module}.md`
   - set `Next Command=cand_check` in `_status.md`
   - keep `Candidate=yes` and `Active Layer=candidate`
6. For invalid `stable` modules:
   - do not generate candidate-side file deletions
   - set `Next Command=stable_verify` in `_status.md`
   - keep `Stable=yes` and `Active Layer=stable`
7. If `bound_modules` differs from the real module binding set:
   - report governance drift
   - state which binding-change command should fix it
   - do not directly rewrite Shared Appendix body content here
   - do not change module state based only on this drift
8. If `_status.md` currently points to a step later than the real smallest actionable step, correct it.
9. If the task hits git-closure trigger conditions, finish git close-out according to the policy.

## 5. Stop Conditions

1. the Shared Appendix view and module binding view have been fully compared
2. every invalid module has been fallen back to its smallest actionable step
3. every invalid candidate-side process file has been cleaned up
4. every `bound_modules` drift case has been clearly reported together with its repair owner

## 6. Output Contract

The output must include at least:

1. a summary of Shared Appendix changes
2. the list of target modules already closed directly in this round
3. the list of affected modules
4. the fallback result for each affected module
5. the list of deleted process files
6. any mismatch between `bound_modules` and the real binding set
7. the standardized `fallback_reason_code` for each affected module
8. the git close-out result

Allowed `fallback_reason_code` values:

1. `shared_appendix_drift`
2. `binding_drift`
3. `truth_drift`

## 7. Non-Goals

This flow does not:

1. create an independent state machine for Shared Appendix files
2. directly modify module body text or Shared Appendix body text "just to fix the binding"
3. replace `cand_check` to re-pass candidate closure
4. replace `stable_verify` to re-check stable alignment

## 8. Example

```md
The user says: "I changed shared files. Help me check which modules need state fallback."
```
