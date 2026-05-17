# Unit Stable Verify Command

## 1. Purpose

This command checks whether current code still aligns with a unit's `stable` Spec.

Goals:

1. decide whether current implementation still satisfies the formal truth
2. if drift exists, decide whether the next legal step is direct small repair, a controlled repair candidate, or a controlled change candidate

## 2. Scope

By default it handles:

1. regression checks after changes, fixes, or refactors on units with `Active Layer=stable`
2. alignment checks between `stable` and current code
3. deciding formal-layer deviations and next action
4. confirming current stable acceptance items are specific enough to be verified

It does not:

1. open a new `candidate`
2. design future behavior
3. directly modify code
4. replace structured verification evidence with a subjective summary

### 2.1 Lifecycle-State Advance Inheritance

Lifecycle-state advancement follows `specflow/framework/command_policy.md` Sections 8.5 and 8.8.
This file states only `unit_stable_verify`-local entry, output, and stop rules.

Stable binding and fingerprint comparisons must use `specflow/framework/process_snapshot_contract.md` normalization rules or deterministic specFlow tooling when available. Manual hash output, shell checksum output, editor display, conversation-derived values, and temporary script results are diagnostic only; they must not support a stable-alignment pass, a drift conclusion, or `_status.md` writeback.

## 3. Preconditions

1. complete required pre-checks
2. the unit's current `Active Layer=stable`
3. `_status.md` says `Next Command=unit_stable_verify`
4. the target unit is explicit
5. the unit has valid `stable`
6. there is actual implementation context that must be checked
7. read any explicitly referenced stable appendix files or bound stable Rule files

## 4. Procedure

1. read `docs/specs/units/stable/s_unit_{unit}.md` and any required appendix or Rule files
2. verify that the stable `Testability / Acceptance Criteria` section contains explicit acceptance items according to `spec_writing_guide.md` Section 6
   - historical stable Specs that still use prose-only acceptance text must not be treated as automatically passing
   - if the stable truth lacks structured acceptance items, report the gap and keep the object at `unit_stable_verify` or route through the smallest legal truth-update path before claiming stable alignment
3. verify stable truth, appendix, Rule, and global-baseline binding fingerprints through the process snapshot contract or deterministic tooling; if authoritative comparison is unavailable, report that stable alignment cannot be confirmed
4. verify current code against key protocols, main flow, error handling, and acceptance criteria in `stable`
4. build a structured verification evidence matrix around acceptance item `id` values:
   - each row must name `acceptance_item_id`, `target`, `verification_surface`, `implementation_surface`, `verification_method`, `evidence`, and `status`
   - `status` must be exactly one of `pass`, `fail`, `partial`, `not_checked`, or `not_runnable_yet`
   - `pass` requires current-round evidence that directly proves the item's `pass_condition`
   - `not_runnable_yet` may be used only when the stable item itself explicitly records `not_runnable_yet` and the current run confirms the same missing runnable surface still exists
   - passing tests that do not prove the item's `pass_condition` must be reported as insufficient evidence, not as `pass`
5. ensure all current-gate acceptance items are covered
6. output `Coverage Summary` with at least:
   - `Total`
   - `Pass`
   - `Fail`
   - `Partial`
   - `Not Checked`
   - `Not Runnable Yet`
7. add risk notes to every `partial`, `not_checked`, and `not_runnable_yet` item
8. classify deviations with the shared severity meanings defined by `specflow/framework/severity_policy.md`
9. conclude:
   - if explicitly referenced stable appendix truth changed enough that the current stable-alignment claim must be re-judged, the result can only be "stable truth drift exists; rerun stable verification against the current stable truth"
   - if any `fail` exists, stable alignment cannot be claimed; continue to Step 10 and classify the smallest legal next action instead of defaulting to a same-command rerun
   - `partial`, `not_checked`, and `not_runnable_yet` are non-blocking only when `specflow/framework/downgrade_policy.md` allows downgrade for the current evidence state
   - if tests pass but do not prove the stable acceptance item, report the evidence gap instead of treating the test result as stable-alignment evidence
   - if key deviations are cleared and evidence is complete, the result is "still aligned with stable"
10. if code, acceptance-item structure, or formal global baseline has drifted from the currently claimed stable state, classify the smallest legal next action:
   - small implementation repair: keep the object at `unit_stable_verify`, return code to `stable` semantics, and rerun `unit_stable_verify:{unit}` after the repair
   - non-trivial stable repair: set `Next Command=unit_fork` and state that the next fork must create `candidate_intent=repair`
   - behavior truth, protocol, boundary, or acceptance change: set `Next Command=unit_fork` and state that the next fork must create `candidate_intent=change`
   - missing structured stable acceptance section: route through the smallest legal truth-update path before claiming stable alignment
   - stable-truth-side drift: rerun stable-layer verification against the current stable truth
   - baseline-side drift: refresh the stable-layer verification conclusion against the current formal global baseline
   - rule-side drift: route through the owning rule-governance path before claiming stable alignment
11. update `_status.md`:
   - if still aligned -> `Next Command=unit_fork`
   - if small repair, evidence, baseline, rule, or stable-truth re-judgment is still required -> keep `Next Command=unit_stable_verify`
   - if controlled candidate work is required -> `Next Command=unit_fork`
   - the deterministic command closure may be executed with `specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_stable_verify --object-type unit --object {unit} --outcome <aligned|small_repair_required|evidence_incomplete|truth_rejudge_required|controlled_repair_required|controlled_change_required> --notes <status-note> --apply`
   - `controlled_repair_required` requires `--candidate-intent repair`
   - `controlled_change_required` requires `--candidate-intent change`

## 5. Stop Conditions

1. alignment with `stable` is clear
2. every explicit acceptance item has one allowed status, or missing acceptance-item structure is reported as the blocker
3. the next action is clear
4. `_status.md` is updated

## 6. Output Contract

1. verification conclusion
2. structured verification evidence matrix by acceptance item `id`
3. `Coverage Summary`
4. downgrade decision when `partial`, `not_checked`, or `not_runnable_yet` exists
5. deviation list
6. `fallback_reason_code` when stable alignment cannot be claimed safely
7. next-step recommendation
   - if the next step remains `unit_stable_verify`, explain which repair, evidence, baseline, rule, or re-judgment work must happen before stable alignment can be claimed
   - if the next step is `unit_fork:{unit}`, state whether the fork must create `candidate_intent=repair` or `candidate_intent=change`
   - a `unit_fork` after failed stable verification is a controlled repair/change candidate entry, not a clean upgrade entry
8. `_status.md` update result
9. the `user-facing close-out block` required by Section 8.6 of `specflow/framework/command_policy.md`
   - `current state` must explicitly confirm `Active Layer=stable` and the written `Next Command`
   - if `Next Command=unit_stable_verify`, `why this next step` must explicitly state the unresolved blocker rather than implying a no-op rerun
   - if `Next Command=unit_fork`, `why this next step` must explicitly state the selected candidate intent

Allowed `fallback_reason_code` values:

1. `truth_drift`
2. `implementation_deviation`
3. `evidence_incomplete`
4. `rule_drift`
5. `baseline_drift`

## 7. Non-Goals

1. creating `candidate`
2. replacing upgrade or repair design with stable-layer verification
3. directly declaring future behavior as valid

## 8. Example

```md
unit_stable_verify:ai
```
