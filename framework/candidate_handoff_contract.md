# Candidate Handoff Contract

Candidate handoff defines which unit process evidence may be consumed by the next unit command.

Only unit handoffs are supported.

## 1. Check To Plan

`unit_plan` may consume `docs/specs/_check_result/unit/{unit}.md` only when the check result validates against current candidate unit truth.

If the check result is missing, malformed, or stale, the next legal step remains `unit_check`.

`unit_plan` must not consume `docs/specs/_check_work/unit/{unit}.md`.
That file is only a resumable `unit_check` work-state file.
It is not a pass gate and cannot prove that the candidate is ready for planning.

## 2. Plan To Implementation

`unit_impl` may consume the active plan only when the plan validates against current candidate unit truth.

If the plan is missing, malformed, or stale, the next legal step remains `unit_plan`.

## 3. Plan And Check To Verify

`unit_verify` may consume check and plan evidence only when both validate against current candidate unit truth.

If implementation work no longer matches the current plan, route to `unit_impl`.

If truth or rule bindings drifted, route to `unit_check`.

## 4. Verify To Promote

`unit_promote` may consume `docs/specs/_verify_result/unit/{unit}.md` only when it validates against current candidate unit truth.

Before stable writeback, `unit_promote` must resolve:

1. `unit_refs`
2. `rule_refs`
3. global baseline rules

`unit_refs` must reference stable unit versions.

## 5. Rejection

No scenario handoff exists.

Scenario process files must not be used as unit evidence.
