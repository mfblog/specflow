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
3. the requested work may affect one or more formal units, bound Shared Contract consumers, or implementation constrained by `system_constraints`

This policy does not replace:

1. unit command files
2. shared-governance routing
3. `unit_stable_verify`, `unit_check`, `unit_impl`, or any other lifecycle gate

Repository mode rule:

1. this repository uses forced diversion only
2. `truth_writeback_required` and `boundary_unclear` must not continue into code modification first
3. reminder-only handling is not allowed here

---

## 3. Required Read Surface Before Classification

Before classification:

1. read `specflow/framework/docs/agent_guidelines/spec_policy.md`
2. read `specflow/framework/docs/agent_guidelines/command_policy.md`
3. if the request names an existing formal unit, read `docs/specs/_status.md` and resolve the unit's current `Active Layer` and `Next Command`
4. read the current-layer main Spec and any explicitly referenced appendix truth needed to judge whether formal behavior truth changes
5. read bound Shared Contract files when the relevant behavior depends on them
6. read `docs/specs/system_constraints.md` when the request may affect shared mechanisms, global default rules, or explicit global exceptions
7. if the request is for a brand-new unit, confirm only that the unit name is clear and non-conflicting before routing to `unit_new:{unit}`

The executor must not classify from code shape alone when repository truth already exists.

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
5. Shared Contract body text or Shared Contract binding relations
6. `system_constraints` default rules or explicit exceptions

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
   - repair an implementation deviation where current Spec already defines the correct behavior

### 6.2 `truth_writeback_required`

Use `truth_writeback_required` when current repository truth already shows that the request would change formal behavior truth, including at least:

1. external behavior changes
2. field set, field meaning, default value, validation rule, or error-return changes
3. state machine or main-flow changes
4. unit responsibility or ownership-boundary changes
5. adding or modifying a Shared Contract
6. adding or modifying a project-wide default rule

### 6.3 `boundary_unclear`

Use `boundary_unclear` when current repository truth is not sufficient to support one implementation result safely, including at least:

1. current Spec does not say enough to decide a protocol, state transition, boundary, or acceptance condition
2. it is unclear whether the requested code change is an implementation repair or a behavior change
3. more than one truth writeback target is plausible, such as candidate main text, appendix truth, Shared Contract text, or `system_constraints_change_proposal`
4. the executor would have to make a new behavior decision in code and explain it later

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
| existing `stable` unit, and the requested change would alter formal behavior truth | `unit_fork:{unit}` first, then write the new candidate truth before implementation |
| existing `candidate` unit, and the requested change would alter current candidate truth | write back into the current candidate main file, required appendix truth, or required Shared Contract truth first, then rerun `unit_check:{unit}` |
| request touches cross-unit shared truth | natural-language routing into the shared-governance branch defined by `shared_ops.md` |
| `implementation_only`, target unit has `Active Layer=stable` | implementation may continue only within current stable truth; after code changes, the unit must return to `unit_stable_verify:{unit}` before stable alignment may be claimed again |
| `implementation_only`, target unit has `Active Layer=candidate` and `_status.md` says `Next Command=unit_impl` | implementation may continue, but only under `unit_impl` semantics |
| `implementation_only`, target unit has `Active Layer=candidate` and `_status.md` says any `Next Command` other than `unit_impl` | do not modify code; return to the currently recorded smallest legal next step first |

Additional routing rules:

1. `implementation_only` does not create permission to skip `Next Command`
2. if the request touches both unit-local truth and cross-unit shared truth, route through natural-language shared governance rather than guessing a local-only shortcut
3. if classification would require guessing whether the target is unit-local truth, Shared Contract truth, or global default-rule truth, the result must stay `boundary_unclear`

---

## 8. Non-Goals

This policy does not:

1. create a new user-facing command
2. let the executor keep truth only in chat
3. weaken the existing `specFlow` lifecycle gates
4. authorize reminder-only handling for truth-changing code requests
