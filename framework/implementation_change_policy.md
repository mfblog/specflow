# Direct Implementation Change Policy

## 1. Purpose

This file defines the mandatory gate for user requests that ask the executor to modify repo-tracked code directly instead of first entering a standard `specFlow` command.

It answers five questions:

1. when a direct implementation request may continue without truth writeback
2. when the request must write formal truth first
3. what the fixed classification results mean
4. how uncertainty must be handled
5. what the smallest legal next step is after classification

This is a governance gate, not an independent command.

---

## 2. Scope

By default this policy applies when all of the following are true:

1. the user asks to modify repo-tracked code, tests, or other implementation-side files
2. the request is not already entered as a standard unit command
3. the requested work may affect one or more formal units, bound Rule consumers, or implementation constrained by stable `g_` rule

This policy does not replace:

1. unit command files
2. rule-governance routing
3. `unit_stable_verify`, `unit_check`, `unit_impl`, or any other lifecycle gate

Repository mode rule:

1. this repository uses forced diversion only
2. `truth_writeback_required` and `boundary_unclear` must not continue into code modification first
3. reminder-only handling is not allowed here

### 2.1 Direct Implementation Lightweight Entry

This policy may be the first policy file for a natural-language request only when the request asks only to edit implementation-side files and does not explicitly ask for any of these changes:

1. behavior truth
2. acceptance truth
3. object boundary or path ownership
4. Rule truth or binding
5. global rule truth
6. end-to-end user-result truth that belongs in a unit Spec
7. governance rules, project standards, or migration behavior
8. guidance before formal truth writeback

When this entry applies, the executor must classify the request through this policy before reading the full natural-language routing file.

Rules:

1. if classification is `implementation_only`, this policy owns the first legal implementation-side action, subject to the current target's recorded `Next Command` and the post-action impact check in Section 3.4
2. if classification is `truth_writeback_required` or `boundary_unclear`, implementation must stop and the executor must read `specflow/framework/natural_language_routing.md` using the classification result as routing evidence
3. if the request contains one of the excluded truth, boundary, shared, system, governance, migration, or guidance fragments, this lightweight entry does not apply; route through `specflow/framework/natural_language_routing.md`
4. this entry is a smaller read path for one natural-language work shape; it does not create a new user-facing command or weaken any formal truth writeback gate

---

## 3. Layered Read Surface Before Classification

Use the smallest read surface that can prove the classification.
Do not read the full framework rule set merely because a request touches implementation files.
Pre-action reading is mandatory only when late discovery could allow an unsafe implementation write, wrong owner, skipped lifecycle state, shared or system drift, or false alignment claim.

### 3.1 Lightweight Pre-Action Prohibitions

Before any implementation-side edit, the executor must already know and follow these prohibitions:

1. do not classify from code shape alone when repository truth already exists
2. do not modify implementation files when the target object, current layer, current `Next Command`, or current truth surface is unknown
3. do not modify implementation files when current truth is not explicit enough to constrain one implementation result
4. do not use code experimentation to discover whether the request is a behavior change
5. do not skip the current `Next Command`
6. do not continue into implementation when the classification is `truth_writeback_required` or `boundary_unclear`

These prohibitions are the required pre-action hard rules for direct implementation requests.
They do not require reading all owner policy files when the needed current truth is already clear from the minimum classification reads below.

### 3.2 Minimum Classification Reads

Before classification, read only the current truth needed to prove the target, state, boundary, and requested change.

Required minimum reads:

1. read the user request and any user-named paths, objects, commands, or implementation surfaces
2. read `docs/specs/_status.md` when the request names an existing formal `unit`, or when the target object must be resolved before implementation permission can be judged
3. read `docs/specs/repository_mapping.md` only when path ownership, object boundary, support-surface ownership, or target-object resolution cannot be proven from the user request and `_status.md`
4. read the current-layer main Spec sections needed to decide whether the requested change alters formal behavior truth
5. read explicitly referenced appendix truth only when the current-layer main Spec makes that appendix relevant to the requested change
6. read bound Rule files only when the current-layer main Spec shows that the relevant behavior depends on those rules
7. read directly relevant implementation or test files only to understand the requested implementation surface or verify whether the requested change is limited to an already-defined result

The minimum read result must be enough to answer:

1. which formal target or support surface the request touches
2. which current layer and `Next Command` govern that target, when the target is a command-target object
3. which current truth constrains the requested implementation result
4. whether the request changes a formal behavior truth item from Section 5
5. whether any shared, system, onboarding, or mapping uncertainty remains

If the minimum reads cannot answer those questions, the result is `boundary_unclear` unless an on-demand owner lookup below identifies a required truth writeback route.

### 3.3 On-Demand Owner Lookups

Read owner policy files only when the current classification depends on the rule that file owns.

On-demand reads:

1. read `specflow/framework/spec_policy.md` when classification depends on formal object definitions, current-layer truth path resolution, binding contracts, acceptance criteria ownership, or process-file invalidation meaning
2. read `specflow/framework/command_policy.md` when classification depends on command-family responsibility, standard command shape, lifecycle-advance authority, or whether the current request is already inside a standard command
3. read `specflow/framework/onboarding_decision_policy.md` when the target has no formal truth, implementation ownership is unmapped, candidate source fields are missing or invalid, or selected candidate truth may depend on existing implementation
4. read `docs/specs/rules/stable/s_g_rule_repository_baseline.md` when the request may affect a reusable mechanism, global default rule, prohibition, or explicit exception
5. read bound Rule files when the current-layer main Spec shows that the relevant behavior depends on those rules

If an on-demand owner lookup is required and cannot safely resolve the uncertainty, classify the request as `boundary_unclear`.
Do not replace the missing owner decision with an implementation-side guess.

### 3.4 Required Post-Action Impact Check

When classification is `implementation_only` and implementation work is allowed, the executor must still complete the post-action impact check before claiming alignment or closure.

Required checks:

1. confirm that the completed change did not alter any formal behavior truth item from Section 5
2. confirm that no newly discovered boundary, rule-truth, global-rule, or onboarding uncertainty invalidates the original classification
3. when the target unit has `Active Layer=stable`, return the unit to `unit_stable_verify` before stable alignment may be claimed again
4. when the target unit has `Active Layer=candidate`, continue only under `unit_impl` semantics when `_status.md` already allowed `Next Command=unit_impl`
5. if the post-action check discovers a truth, boundary, shared, system, or onboarding issue, stop and route through the required truth or boundary step before any pass, alignment, or closure claim

The post-action impact check does not create permission to edit implementation files before pre-action classification is proven.

---

## 4. Fixed Classification Results

Only these classification results are allowed:

1. `implementation_only`
   - the requested code change can be completed within already-written formal truth
   - no truth-side writeback is required before implementation starts
2. `truth_writeback_required`
   - the requested work would change formal behavior truth that the repository must acknowledge first
   - implementation must not start from code
3. `boundary_unclear`
   - current repository truth is not sufficient to safely decide whether the request is only implementation work or a behavior change
   - treat this exactly like `truth_writeback_required`

---

## 5. What Counts As Formal Behavior Truth

For this policy, a request touches formal behavior truth when it would create, remove, or change any formally acknowledged answer about:

1. unit goal and unit boundary
2. external protocols, field meanings, default values, validation rules, and error semantics
3. main flow, state transitions, or branch convergence semantics
4. acceptance criteria or testable success conditions
5. Rule body text or Rule binding relations
6. stable `g_` rule default rules or explicit exceptions

If a request touches any item above, it is not implementation-only.

---

## 6. Classification Rules

### 6.1 `implementation_only`

Use `implementation_only` only when all of the following hold:

1. no formal behavior truth item from Section 5 changes
2. current repository truth is already explicit enough to constrain one implementation result without inventing a new behavior decision
3. the request does only one or more of the following:
   - pure refactor
   - add or adjust tests
   - add logging, tracing, or other observability
   - performance optimization with unchanged semantics
   - small repair of an implementation deviation where current Spec already defines the correct behavior and the repair can be completed and checked in the same handling round

### 6.2 `truth_writeback_required`

Use `truth_writeback_required` when current repository truth already shows that the request would change formal behavior truth, including at least:

1. external behavior changes
2. field set, field meaning, default value, validation rule, or error-return changes
3. state machine or main-flow changes
4. unit responsibility or ownership-boundary changes
5. adding or modifying a Rule
6. adding or modifying a project-wide default rule

### 6.3 `boundary_unclear`

Use `boundary_unclear` when current repository truth is not sufficient to support one implementation result safely, including at least:

1. current Spec does not say enough to decide a protocol, state transition, boundary, or acceptance condition
2. it is unclear whether the requested code change is an implementation repair or a behavior change
4. the executor would have to make a new behavior decision in code and explain it later
5. the target scope has no current formal truth and onboarding source decision has not selected ordinary candidate creation, candidate with evidence appendix, or a stable-governed route
6. the current unit candidate is missing `candidate_intent`
7. the current candidate is missing `source_basis` or `evidence_appendix_ref`
8. the current candidate records `source_basis=existing_implementation` or `source_basis=mixed`, but the referenced evidence appendix is missing or cannot be read

Rules:

1. `boundary_unclear` is not a softer version of `truth_writeback_required`
2. `boundary_unclear` must be routed exactly like `truth_writeback_required`
3. executors must not use code experimentation to discover the truth boundary

---

## 7. Routing And Smallest Legal Next Step

The smallest legal next step after classification is fixed as follows:

| Current situation | Smallest legal next step |
|---|---|
| brand-new unit, user directly asks to write code | `unit_new:{unit}` |
| no formal truth exists and candidate source is not yet decided | route through `onboarding_decision_policy.md` before creating candidate truth or editing code |
| existing `stable` unit, and the requested change would alter formal behavior truth | `unit_fork:{unit}` first, creating `candidate_intent=change`, then write the new candidate truth before implementation |
| existing `stable` unit, current stable truth is correct, and the repair is too large to safely complete and verify as a direct small implementation repair | `unit_fork:{unit}` first, creating `candidate_intent=repair` |
| existing `candidate` unit, and the requested change would alter current candidate truth | write back into the current candidate main file, required appendix truth, or required Rule truth first, then rerun `unit_check:{unit}` |
| existing `candidate` unit is missing required candidate source fields or required evidence appendix | repair the candidate source fields or evidence appendix first, then rerun `unit_check:{unit}` |
| request touches cross-unit rule truth | natural-language routing into the rule-governance branch defined by `natural_language_routing.md` |
| small `implementation_only`, target unit has `Active Layer=stable` | implementation may continue only within current stable truth; after code changes, the unit must return to `unit_stable_verify:{unit}` before stable alignment may be claimed again |
| `implementation_only`, target unit has `Active Layer=candidate` and `_status.md` says `Next Command=unit_impl` | implementation may continue, but only under `unit_impl` semantics |
| `implementation_only`, target unit has `Active Layer=candidate` and `_status.md` says any `Next Command` other than `unit_impl` | do not modify code; return to the currently recorded smallest legal next step first |

Additional routing rules:

1. `implementation_only` does not create permission to skip `Next Command`
2. if the request touches both unit-local truth and cross-unit rule truth, route through natural-language rule governance rather than guessing a local-only shortcut
3. if classification would require guessing whether the target is unit-local truth, Rule truth, or global default-rule truth, the result must stay `boundary_unclear`

---

## 8. Non-Goals

This policy does not:

1. create a new user-facing command
2. let the executor keep truth only in chat
3. weaken the existing `specFlow` lifecycle gates
4. authorize reminder-only handling for truth-changing code requests
