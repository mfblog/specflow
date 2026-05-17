# Unit Init Command

## 1. Purpose

This command creates the first `stable` Spec for a historical unit.

Goals:

1. capture the unit's already-effective formal behavior
2. create the unit's first formal truth file
3. register the unit in `docs/specs/repository_mapping.md`
4. register the unit in `docs/specs/_status.md`

## 2. Scope

By default this command handles:

1. first-time governance onboarding of a historical unit
2. units that already have implementation and stable behavior but are not yet inside the Spec system
3. creation of the first `stable` only when a fully reviewable accepted behavior baseline already exists
4. registration of the historical unit in `docs/specs/repository_mapping.md`

It does not handle:

1. creating a new unit
2. forking a new candidate from existing `stable`
3. creating `candidate` directly
4. onboarding a historical unit by first writing duplicated unit-local formal truth when the real task is unresolved cross-unit rule-truth governance
5. creating the first `stable` directly from raw implementation inspection when evidence is incomplete, conflicting, or still needs business confirmation

### 2.1 Lifecycle-State Advance Inheritance

Lifecycle-state advancement follows `specflow/framework/command_policy.md` Sections 8.5 and 8.8.
This file states only `unit_init`-local entry, output, and stop rules.

## 3. Preconditions

Before execution:

1. complete the required pre-checks from `spec_policy.md`; if the unit is not registered yet, at minimum confirm there are no conflicting old status or leftover process files
2. the target unit name is explicit
3. the unit is not yet in `docs/specs/_status.md`
4. the goal is to capture current truth, not define future design
5. read `specflow/framework/onboarding_decision_policy.md`
6. read `specflow/framework/repository_mapping_policy.md`
7. read `docs/specs/repository_mapping.md`
8. confirm the target unit is not already present in `Object Registry` and does not conflict with any current `unit`, `rule`, support-surface, or ignore rule
9. direct first-stable onboarding is allowed only when `onboarding_decision_policy.md` proves that the accepted behavior baseline is complete, conflicts are closed, material unknowns are closed or irrelevant, and shared/global truth is resolved
10. if the target only has raw implementation evidence, incomplete evidence, unresolved conflicts, or retained behavior that still needs business confirmation, do not start `unit_init`; route to candidate creation with the required `source_basis` and evidence appendix
11. if onboarding current truth would create duplicated formal truth across units, or if the shared/unit boundary is still unstable, do not start `unit_init`; resolve that rule-truth boundary through natural-language rule governance first
12. if the first `stable` reuses already-existing rule truth, read the relevant `rule` files before writing `rule_refs`
13. if the task also touches global baseline, shared mechanisms, or exceptions, read `docs/specs/rules/stable/s_g_rule_repository_baseline.md`
14. if the unit involves technical choices, shared infrastructure, cross-unit reuse, global exceptions, or system-level constraint relationships, the first `stable` must include `Rule Alignment` or an equivalent section
15. if the round creates, updates, or deletes any unit `rule_refs` value or any file under `docs/specs/rules/**`, read `specflow/framework/rule_sync.md` first
16. if the round may remove intentional-unbound retention fields from a touched Rule file, read every current-layer unit main file needed to derive the real repository-wide binding set of each touched Rule from `rule_refs`

## 4. Procedure

1. summarize the unit's already-effective behavior baseline
2. if needed, read `s_g_rule_repository_baseline.md` as an upstream input
3. confirm that first-stable onboarding is allowed by `onboarding_decision_policy.md`; if not, stop before writing stable truth and route to candidate creation
4. if onboarding current truth shows that one or more existing formal units already depend on the same formal truth and that truth is not yet formalized as one stable rule object, stop and reroute through natural-language rule governance from current repository truth instead of writing duplicated unit-local `stable` truth
5. prepare the `docs/specs/repository_mapping.md` writeback for the historical unit before stable truth or `_status.md` mutation:
   - add or update one `Object Registry` row for the target unit
   - set `kind=unit`, `id={unit}`, and the one-line `responsibility`
   - do not write `scope`; `scope` is not an Object Registry column
   - set `spec_files=docs/specs/units/stable/s_unit_{unit}.md` after the stable file is created in this same round
   - set `registration_state=landed` only when concrete implementation paths are declared
   - if no implementation path is declared yet, set `registration_state=planned` and `implementation_paths=none`
   - record any implementation surface, support surface, governed root, ignore rule, or conflict rule that this first stable onboarding round already needs
   - if current repository truth is insufficient to write the exact mapping update without guessing, stop before stable truth and `_status.md` writeback
6. create `docs/specs/units/stable/s_unit_{unit}.md`
7. ensure the file covers:
   - `Context & Motivation`
   - `Terminology`
   - `Data Structures / Protocols`
   - `State Machine / Business Flow`
   - `Edge Cases & Error Handling`
   - `Testability / Acceptance Criteria` with explicit acceptance items that satisfy `spec_writing_guide.md` Section 6
8. if needed, add `Rule Alignment` with at least:
   - `rule_refs` written according to the Rule References contract in `specflow/framework/spec_writing_guide.md` Section 4
   - `rule_reuse_summary`
   - `rule_exceptions`
9. write the prepared `docs/specs/repository_mapping.md` update in the same round as the stable truth writeback and before `_status.md` mutation
10. if the round changed Rule bindings or touched Rule files:
   - derive the real repository-wide binding set of each touched Rule from current-layer unit `rule_refs` plus this round's prepared target-unit stable writeback
   - if current repository truth is insufficient to derive that touched real binding set safely, stop and reroute through natural-language rule governance from current repository truth instead of guessing
   - do not write consumer metadata into touched Rule files; every touched Rule file must omit `bound_objects` after this writeback
   - if a touched Rule file now has one or more formal bound units after this round, remove or stop carrying any `unbound_retention`, `unbound_retention_reason`, and `unbound_retention_owner` fields from that resulting bound file state in the same round
11. update `docs/specs/_status.md`:
   - `Stable=yes`
   - `Candidate=no`
   - `Active Layer=stable`
   - `Next Command=unit_fork`
   - the deterministic command closure may be executed with `specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_init --object-type unit --object {unit} --outcome stable_created --notes <status-note> --apply`
12. if the round changed any unit `rule_refs` value or any file under `docs/specs/rules/**`, run `rule_sync` after `_status.md` has been updated, even when no additional affected object is known yet
   - pass execution-local `current_stable_landing_unit={unit}` into that `rule_sync` run because this same round just wrote the unit's first stable truth together with its current stable Rule binding
   - pass execution-local `stable_landing_rule_refs=<exact-shared-ref-list-written-by-this-landing>` into that same `rule_sync` run; `current_stable_landing_unit` alone is not sufficient
   - the deterministic reconciliation part may be executed with `specflow/tooling/bin/specflowctl-<os>-<arch> rule sync-impact --rule-refs <rule-ref> --units {unit} --stable-landing-unit {unit} --stable-landing-rule-refs <exact-stable-landing-rule-ref-list>` or the corresponding `--rule-ids` form, and at least one rule trigger input must already be known before this deterministic execution starts

## 5. Stop Conditions

1. the first `stable` exists
2. `docs/specs/repository_mapping.md` includes the unit in `Object Registry` with its implementation registration state, the created stable Spec file, and any path-ownership entries required by this first stable onboarding round
3. `_status.md` registration is complete
4. Rule side effects, if any, are closed
5. if onboarding discovered unresolved cross-unit rule truth, the command stopped and rerouted through natural-language rule governance instead of writing duplicated unit-local `stable` truth
6. if onboarding evidence was insufficient for first-stable landing, the command stopped before stable writeback and routed to candidate creation with evidence handling
7. if repository truth was insufficient to write the required repository mapping update safely, the command stopped before stable truth and `_status.md` writeback instead of guessing
8. the command does not automatically open a candidate round

## 6. Output Contract

1. onboarding judgment
2. created file path
3. first-stable eligibility result from `onboarding_decision_policy.md`
4. acceptance-item structure result for the first stable Spec
5. whether `Rule Alignment` was required and why
6. whether the command had to stop and reroute through natural-language rule governance because rule-truth boundary closure was required before onboarding could continue
7. whether the command had to stop and route to candidate creation because evidence was not sufficient for direct stable onboarding
8. `docs/specs/repository_mapping.md` writeback result, including the new `Object Registry` row and any path-ownership entries written in this round
9. `_status.md` update result
10. Rule reconciliation result when the round changed rule truth or bindings
11. next-step suggestion
12. the `user-facing close-out block` required by Section 8.6 of `specflow/framework/command_policy.md`
   - `current state` must explicitly confirm the stable-layer landing written to `_status.md`
   - if the round stopped and rerouted through natural-language rule governance, `next step` must name that reroute directly instead of implying that onboarding closed

## 7. Non-Goals

1. creating the first `candidate`
2. jumping directly into implementation
3. redesigning the unit
4. using first-time historical onboarding to bypass required rule-truth boundary closure

## 8. Example

```md
unit_init:ai
```
