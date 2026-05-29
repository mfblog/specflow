# Direct Implementation Change

This gate applies when a user asks for implementation-side proposals or repo-tracked code, test, or implementation changes outside an exact lifecycle command.
It is the operation owner for the `implementation-only` adoption mode described in `framework/core/adoption_modes.md`.

Implementation-only is a narrow entry path, not a shortcut around truth. It may propose implementation work or change code and tests only when the request fits already-written formal truth.

This is a governance gate, not an independent command.

## Scope

This gate applies when all of the following are true:

1. the user asks for an implementation-side proposal or asks to modify repo-tracked code, tests, or other implementation-side files
2. the request is not already entered as a standard unit command
3. the requested work may affect one or more formal units, bound rule consumers, or implementation constrained by a stable global rule

This gate does not replace lifecycle Context Cards, rule-governance routing, `unit_stable_verify`, `unit_check`, `unit_impl`, or any other lifecycle gate.

Repository mode rule:

1. this repository uses forced diversion only
2. `truth_writeback_required` and `boundary_unclear` must not continue into code modification first
3. reminder-only handling is not allowed here

## Classification

Classify the request before proposing or editing implementation-side files:

1. `implementation_only` - fits already-written formal truth.
2. `truth_writeback_required` - changes behavior, boundary, acceptance, rule, or ownership truth.
3. `boundary_unclear` - current truth is insufficient to decide; treat as truth writeback required.

Use `implementation_only` only when all of the following hold:

1. no formal behavior truth item changes
2. current repository truth is already explicit enough to constrain one implementation result without inventing a new behavior decision
3. the request is a pure refactor, test change, observability change, performance optimization with unchanged semantics, or small repair of an implementation deviation where the current Spec already defines the correct behavior

Use `truth_writeback_required` when current repository truth shows that the request would change external behavior, field sets, field meaning, defaults, validation rules, error semantics, state transitions, unit responsibility, rule text, rule binding, or project-wide default rules.

Use `boundary_unclear` when current repository truth is not sufficient to support one implementation result safely. Treat it exactly like `truth_writeback_required`.

## Minimum Reads

Use the smallest read surface that can prove target, state, boundary, and requested change.

Required minimum reads:

1. read the user request and any user-named paths, objects, commands, or implementation surfaces
2. read `docs/specs/_status.md` when the request names an existing formal unit, or when the target object must be resolved before implementation permission can be judged
3. read `docs/specs/repository_mapping.md` only when path ownership, object boundary, support-surface ownership, or target-object resolution cannot be proven from the user request and `_status.md`
4. read the current-layer main Spec sections needed to decide whether the requested change alters formal behavior truth
5. read explicitly referenced appendix truth only when the current-layer main Spec makes that appendix relevant
6. read bound rule files only when the current-layer main Spec shows that relevant behavior depends on those rules
7. read directly relevant implementation or test files only to understand the requested implementation surface or verify that the requested change is limited to an already-defined result

Before any implementation-side proposal or edit, the executor must already know that:

1. repository truth proves the target object, current layer, current `Next Command`, and implementation permission
2. current truth is explicit enough to constrain one implementation result
3. the request does not require behavior, boundary, acceptance, rule, ownership, governance, migration, or guidance work first

If these facts are not proven, `implementation_only` must not authorize an implementation proposal or implementation-side edit.

Do not classify from code shape alone when repository truth exists.
Do not use code experimentation to discover whether the request is a behavior change.

## On-Demand Owner Lookups

Read owner policy files only when classification depends on the rule that file owns:

1. read `framework/spec_policy.md` when classification depends on formal object definitions, current-layer truth path resolution, binding contracts, acceptance criteria ownership, or process-file invalidation meaning
2. read `framework/lifecycle/overview.md` and `framework/core/lifecycle_authority.md` when classification depends on command-family responsibility, standard command shape, lifecycle-advance authority, or whether the current request is already inside a standard command
3. read `framework/onboarding_decision_policy.md` when the target has no formal truth, implementation ownership is unmapped, candidate source fields are missing or invalid, or selected candidate truth may depend on existing implementation
4. read `docs/specs/rules/stable/s_g_rule_repository_baseline.md` when the request may affect a reusable mechanism, global default rule, prohibition, or explicit exception
5. read bound rule files when current-layer unit truth shows that relevant behavior depends on those rules

If an owner lookup cannot safely resolve uncertainty, classify the request as `boundary_unclear`.

## Next Steps

The smallest legal next step after classification is fixed:

1. Brand-new unit plus a direct code request enters `unit_new:{unit}`.
2. No formal truth plus undecided candidate source enters `framework/onboarding_decision_policy.md` before candidate truth or code edits.
3. Existing stable unit plus behavior truth change starts `unit_fork:{unit}` with `candidate_intent=change`, then writes candidate truth before implementation.
4. Existing stable unit plus a repair that is too large for direct implementation starts `unit_fork:{unit}` with `candidate_intent=repair`.
5. Existing candidate unit plus candidate truth change writes current candidate truth, appendix truth, or rule truth first, then reruns `unit_check:{unit}`.
6. Existing candidate unit with missing candidate source fields or evidence appendix repairs that truth first, then reruns `unit_check:{unit}`.
7. Cross-unit rule truth routes through `framework/operations/entry_routing.md` into rule governance.
8. Small `implementation_only` work on a stable unit may continue only within current stable truth; after code changes, return to `unit_stable_verify:{unit}` before stable alignment may be claimed.
9. `implementation_only` work on a candidate unit may continue only when `_status.md` already allows `Next Command=unit_impl`; otherwise return to the recorded smallest legal next step first.
10. If the user selected implementation-only and the request exceeds it, stop at the smallest legal truth step and explain that mode boundary in the close-out.

## Post-Action Impact Check

When classification is `implementation_only` and implementation work is allowed, complete this check before claiming alignment or closure:

1. confirm the completed change did not alter formal behavior truth
2. confirm no newly discovered boundary, rule-truth, global-rule, onboarding, or repository mapping uncertainty invalidates the original classification
3. return stable targets to `unit_stable_verify` before stable alignment is claimed
4. keep candidate targets under `unit_impl` semantics only when lifecycle state allows it
5. stop and reroute through the required truth or boundary step if the post-action check finds truth, boundary, shared, system, or onboarding issues

## Formal Behavior Truth

A request touches formal behavior truth when it would create, remove, or change any formally acknowledged answer about:

1. unit goal or unit boundary
2. external protocols, field meanings, default values, validation rules, or error semantics
3. main flow, state transitions, or branch convergence semantics
4. acceptance criteria or testable success conditions
5. rule body text or rule binding relations
6. stable global rule default rules or explicit exceptions

If a request touches any item above, it is not implementation-only.
