# Candidate Handoff Contract

Candidate handoff defines which unit process evidence may be consumed by the next unit command.

Only unit handoffs are supported.

## 1. Check To Plan

`unit_plan` may consume `docs/specs/_check_result/unit/{unit}.md` only when the check result validates against current candidate unit truth.

If the check result is missing, malformed, or stale, the next legal step remains `unit_check`.

`unit_plan` must not consume `docs/specs/_check_work/unit/{unit}.md`.
That file is only a resumable `unit_check` checklist file.
It is not a pass gate and cannot prove that the candidate is ready for planning.

## 2. Plan To Implementation

`unit_impl` may consume the active plan only when the plan validates against current candidate unit truth.

If the plan is missing, malformed, or stale, the next legal step remains `unit_plan`.

## 3. Plan And Check To Verify

`unit_verify` may consume check and plan evidence only when both validate against current candidate unit truth.

If implementation work no longer matches the current plan, route to `unit_impl`.

If truth or rule bindings drifted, route to `unit_check`.

## 4. Verify To Promote

`unit_promote` may consume `docs/specs/_plans/active/{unit}.md` and `docs/specs/_verify_result/unit/{unit}.md` only when both validate against current candidate unit truth and the verify result binds to the current active plan.

If the active plan has retirement targets, the verify result must prove every target with `result: pass` and `mainline_dependency: not_required` before promotion.

Before stable writeback, `unit_promote` must resolve:

1. `unit_refs`
2. `rule_refs`
3. global baseline rules

`unit_refs` must reference stable unit versions.

## 5. Rejection

No scenario handoff exists.

Scenario process files must not be used as unit evidence.

## 6. Stable Verify To Fork

`unit_stable_verify` advancing outcomes may consume `docs/specs/_stable_verify_result/unit/{unit}.md` only when it validates against current stable unit truth.

The close outcome must match the stable verify result `decision`.

If the stable verify result is missing, malformed, stale, or records a different decision, the unit must remain at `unit_stable_verify`.

`docs/specs/_verify_result/stable/unit/{unit}.md` is a stable promotion summary only.
It must not be used as current stable implementation-alignment evidence.
