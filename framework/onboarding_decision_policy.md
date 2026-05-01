# Onboarding Decision Policy

## 1. Purpose

This file defines how `specFlow` decides the candidate source for an unowned or not-yet-stable target scope.

It exists to prevent one unsafe shortcut:

1. code currently behaves in one way
2. therefore the formal truth should say the system must behave that way

That shortcut is not allowed.
Existing implementation may be evidence for a candidate, but it does not become formal behavior truth until the candidate main Spec states the selected rule and the normal command gates close.

This file answers six questions:

1. whether the target scope already has formal truth
2. whether existing implementation is present for the target scope
3. whether the current candidate depends on that existing implementation
4. which `source_basis` value the candidate must record
5. whether an evidence appendix is required
6. which command or checkpoint owns the next step

This policy does not create a new lifecycle state.
The formal lifecycle remains `candidate` and `stable` as defined by `spec_policy.md`.

---

## 2. Scope

This policy applies when a route, command, or implementation gate needs to decide candidate source for a `unit` or `scenario`.

Use this policy when at least one of these is true:

1. a target scope has no current `candidate` or `stable`
2. a request touches implementation under a scope whose formal truth is missing, incomplete, or unmapped
3. a first candidate is being created for a scope that already has implementation, tests, runtime behavior, or historical material
4. an existing candidate is missing `source_basis` or `evidence_appendix_ref`
5. a command needs to decide whether a candidate that depends on existing implementation has a required evidence appendix
6. a historical first-stable onboarding path is being considered

This policy covers:

1. `unit` candidates
2. `scenario` candidates

This policy does not independently govern `shared_contract`.
Shared truth remains governed by the shared-governance branch and the internal shared flows.

---

## 3. Required Read Surface

Before applying this policy, read only the current truth needed for the target scope.

Required reads:

1. `docs/specs/repository_mapping.md`
   - to resolve path ownership, target scope, declared truth surface, and declared implementation surface
2. `docs/specs/_status.md`
   - when the target scope is already registered as a `unit` or `scenario`
3. the current-layer candidate or stable Spec when it exists
4. the candidate evidence appendix when `evidence_appendix_ref` is not `none`
5. directly relevant implementation or test files only when repository mapping shows that they are inside the target scope or when the user named them directly
6. `docs/specs/system_constraints.md` when the candidate source decision depends on a global default, shared mechanism, prohibition, or exception

Rules:

1. do not classify source from user wording alone
2. do not classify source from directory shape alone
3. do not scan the whole repository by default
4. when repository mapping cannot identify the target scope or implementation surface, the result is `boundary_unclear`
5. when an existing candidate is present, the recorded `source_basis` and `evidence_appendix_ref` must be consumed before deciding whether implementation may proceed

---

## 4. Candidate Source Fields

Every `unit` and `scenario` candidate main Spec must record these frontmatter fields:

```yaml
source_basis: new_design | existing_implementation | mixed | replacement
evidence_appendix_ref: none | <candidate appendix path>
```

Allowed `source_basis` values:

1. `new_design`
   - the candidate does not use existing implementation as the source of selected behavior truth
   - existing implementation is absent or irrelevant to this candidate's rules
   - `evidence_appendix_ref` must be `none`
2. `existing_implementation`
   - the candidate mainly captures behavior that already exists in implementation, tests, runtime behavior, or historical material
   - `evidence_appendix_ref` must point to a current candidate evidence appendix
3. `mixed`
   - the candidate combines retained existing behavior with new or changed design
   - `evidence_appendix_ref` must point to a current candidate evidence appendix
4. `replacement`
   - existing implementation may exist, but this round explicitly does not use it as the source of selected candidate behavior
   - the candidate is replacing or discarding the old behavior as a new design
   - `evidence_appendix_ref` must be `none`

Rules:

1. candidate main Spec text is the selected rule surface
2. evidence appendix text is the observed-fact surface
3. implementation and verification must use the candidate main Spec and formal bindings as truth, not the evidence appendix
4. if `source_basis` is missing, unsupported, or internally inconsistent with `evidence_appendix_ref`, the candidate is not closed enough for `unit_check` or `scenario_check` to pass
5. if the candidate's selected behavior depends on existing implementation but the candidate records `new_design` or `replacement`, the candidate is internally inconsistent and must return to candidate repair

---

## 5. Evidence Appendix

An evidence appendix is a current-round candidate appendix.

Recommended paths:

1. `docs/specs/units/candidate/appendix/c_unit_{unit}_evidence.md`
2. `docs/specs/scenarios/candidate/appendix/c_scenario_{scenario}_evidence.md`

It must not enter `docs/specs/_status.md`.
It must not create an `evidence` lifecycle state.
It must not define formal behavior truth.

An evidence appendix must include:

1. target scope
2. inspected sources
   - code paths, tests, runtime observations, historical documents, or other named material
3. observed behavior
4. mapping from observed behavior to candidate main-Spec rules
5. conflicts
6. unknowns
7. behavior decisions that require human confirmation before they can be treated as selected candidate truth

Rules:

1. evidence may say what the system currently does
2. evidence must not say what the system should do unless the candidate main Spec also states that selected rule
3. observed behavior that is not selected in the candidate main Spec remains evidence only
4. a material conflict or material unknown blocks `unit_check` or `scenario_check` unless the candidate main Spec explicitly makes a bounded design decision that no longer depends on that unknown
5. evidence appendix participates in candidate appendix snapshots only to prove which evidence was reviewed by a gate; snapshot inclusion does not make the appendix an implementation truth source

---

## 6. Decision Procedure

Apply this procedure to the target scope.

1. Resolve target scope.
   - Use `repository_mapping.md` and `_status.md`.
   - If target scope cannot be resolved, stop with `boundary_unclear`.
2. Check formal truth state.
   - If current `stable` exists and no candidate exists, the scope is stable-governed.
   - If current `candidate` exists, the scope is candidate-governed.
   - If neither exists, the scope is unowned by formal behavior truth.
3. For an existing candidate, validate source fields.
   - Missing or unsupported `source_basis` is a candidate repair blocker.
   - Missing or invalid `evidence_appendix_ref` is a candidate repair blocker when `source_basis` is `existing_implementation` or `mixed`.
4. For a new candidate, decide whether existing implementation is a source for selected behavior.
   - Existing implementation may include code, tests, runtime behavior, historical documentation, or manually verified production behavior.
   - If selected candidate rules depend on that material, use `existing_implementation` or `mixed`.
   - If selected candidate rules do not depend on that material, use `new_design` or `replacement`.
5. Decide evidence appendix requirement.
   - `existing_implementation` requires evidence appendix.
   - `mixed` requires evidence appendix.
   - `new_design` requires `evidence_appendix_ref=none`.
   - `replacement` requires `evidence_appendix_ref=none`.
6. Decide next route.
   - missing formal truth plus `new_design` or `replacement` routes to ordinary candidate creation
   - missing formal truth plus `existing_implementation` or `mixed` routes to candidate creation with evidence appendix
   - existing candidate with valid source fields routes to the current recorded next command
   - existing candidate with invalid source fields routes to candidate repair and then `unit_check` or `scenario_check`
   - stable-governed scope routes through stable command rules; behavior changes must open a candidate round first

### 6.1 Stable-Fork Candidate Source

This rule applies when `unit_fork` or `scenario_fork` creates a new candidate from an existing stable Spec.

The stable Spec is already accepted formal truth.
It is not existing-implementation evidence, and it does not require a candidate evidence appendix merely because the same behavior already exists in code.

When the forked candidate is generated only from stable formal truth plus the current round's selected design changes, write these candidate source fields during the same candidate write:

```yaml
source_basis: new_design
evidence_appendix_ref: none
```

In this stable-fork case, `source_basis=new_design` means the candidate is not using implementation, tests, runtime behavior, or historical material as the source of selected behavior truth.
It does not mean every carried-forward stable rule was newly invented in this round.

If the fork command selects behavior from implementation, tests, runtime behavior, historical material, or other non-stable evidence, it must use the normal source decision rules in Section 6:

1. write `source_basis=existing_implementation` or `source_basis=mixed` and create the required candidate evidence appendix in the same round
2. stop before candidate writeback when the evidence appendix or source decision is not ready

`unit_fork` and `scenario_fork` must not create a candidate main Spec without both `source_basis` and `evidence_appendix_ref`.

---

## 7. Human Judgment Boundary

Executors may inspect repository truth and implementation facts.
Executors must not decide by themselves that historical behavior is business-correct when more than one business interpretation remains plausible.

Ask a user-facing question only when the route depends on a business judgment that cannot be derived from repository truth.
The question must be phrased in ordinary language.

Allowed question shape:

```text
Should the current behavior of this area be preserved as the starting rule, or should this round replace it with a new design?
```

Disallowed question shape:

```text
Should source_basis be mixed?
```

When the user confirms that existing behavior should be preserved only in part, use `mixed` and record the retained behavior and the replaced behavior explicitly in the candidate main Spec and evidence appendix.

---

## 8. Historical First-Stable Onboarding

`unit_init` may create the first `stable` only when the target already has a fully reviewable accepted behavior baseline.

`unit_init` must not create the first `stable` directly from raw implementation inspection alone.

Direct first-stable onboarding is allowed only when all of these hold:

1. the target behavior is already accepted as current business truth outside `specFlow`
2. the evidence is complete enough to cover the stable Spec's behavior, boundary, protocols, error semantics, and acceptance criteria
3. material conflicts are closed
4. material unknowns are closed or explicitly proven irrelevant to the stable scope
5. shared truth and global constraints are already resolved or bound

If any condition does not hold, the next legal route is candidate creation using this policy, not first-stable creation.

---

## 9. Non-Goals

This policy does not:

1. create an `evidence` status
2. create `_onboarding`
3. authorize implementation before formal truth writeback
4. replace `unit_check`, `scenario_check`, `unit_plan`, `unit_impl`, or verification gates
5. make evidence appendix a behavior truth source
6. decide business correctness of historical behavior without user or durable truth confirmation
