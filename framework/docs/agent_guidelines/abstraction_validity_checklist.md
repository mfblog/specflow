# Governance Abstraction Validity Checklist

## 1. Purpose

This file defines one compact checklist for judging whether the current `specFlow` core objects are still valid governance abstractions.

It answers one question:

1. whether `module`, `flow`, `shared_contract`, `system_constraints`, and `ProjectSpec` still model real governance work rather than forcing real repositories into an artificial shape

This file is not a command.
It is a design-review aid for framework governance.

---

## 2. How To Use This Checklist

Use this checklist in the following order:

1. review each object's valid-fit boundary
2. review each object's failure boundary
3. if one object repeatedly hits failure signals, stop optimizing downstream mechanisms such as mapping and repair that abstraction first

This checklist does not judge whether current docs are well written.
It judges whether the abstraction itself is still worth keeping.

---

## 3. Object Checklist

### 3.1 `module`

`module` is a valid abstraction only when all of the following mostly hold:

1. it represents one local capability unit rather than a casually cut directory slice
2. it can answer its goal, boundary, local protocol, state transition, and local acceptance criteria on its own
3. it remains the natural landing unit for implementation planning, implementation work, and implementation verification
4. its boundary is driven mainly by capability ownership rather than folder position alone

`module` is drifting or failing as an abstraction when one or more of the following repeatedly hold:

1. outside the folder name, the repository cannot explain what that `module` actually owns
2. one `module` carries several unrelated capabilities with no stable local boundary
3. one capability must stay permanently split across several `module` objects to be described truthfully
4. `module` regularly has to answer questions that belong to `flow`, `shared_contract`, or `system_constraints`
5. real implementation work cannot reliably use `module` as the execution and repair landing point

Passing conclusion:

1. `module` still works as the natural unit for local capability and implementation governance

Failing conclusion:

1. `module` has degraded into a directory label or arbitrary slice

### 3.2 `flow`

`flow` is a valid abstraction only when all of the following mostly hold:

1. the repository really has user-visible or business-significant cross-module paths
2. modeling those paths separately improves end-to-end verification, failure absorption, and cross-module closure
3. `flow` mainly answers how one business path traverses multiple modules
4. without `flow`, cross-module path truth would remain scattered across `module` files and would not close cleanly

`flow` is drifting or failing as an abstraction when one or more of the following repeatedly hold:

1. the repository has little or no path-level work worth modeling separately
2. most `flow` files merely list module names in sequence
3. `flow` keeps repeating module-local truth without adding design control
4. `flow` exists mainly to satisfy framework form rather than repository need
5. `flow` does not materially help verification, review, or impact judgment

Passing conclusion:

1. `flow` still captures real business-path truth with distinct governance value

Failing conclusion:

1. `flow` has become formal stitching with little information gain

### 3.3 `shared_contract`

`shared_contract` is a valid abstraction only when all of the following mostly hold:

1. it carries one local truth reused by more than one formal object
2. that truth is neither a global default rule nor one module's private internal detail
3. extraction into shared truth reduces double-writing, drift, and boundary ambiguity
4. even after reuse grows, the shared body remains local, concrete, and bindable

`shared_contract` is drifting or failing as an abstraction when one or more of the following repeatedly hold:

1. any broadly useful material gets pushed into `shared_contract`, turning it into a general holding area
2. module-private local truth is extracted too early only because it feels reusable
3. material that is already a project-wide default rule still lives as a shared local object
4. shared files repeatedly answer whole-flow or whole-project questions
5. the practical role of shared files becomes "store content whose owner is unclear"

Passing conclusion:

1. `shared_contract` still holds reusable local truth with stable boundaries

Failing conclusion:

1. `shared_contract` has become a public junk drawer or a substitute for global rules

### 3.4 `system_constraints`

`system_constraints` is a valid abstraction only when all of the following mostly hold:

1. it records only project-wide default technical rules, preferred shared mechanisms, prohibitions, and explicit exceptions
2. those rules actually apply across multiple modules in a stable way
3. without it, modules would keep inventing incompatible local defaults
4. it mainly answers global default choice rather than one module's design detail

`system_constraints` is drifting or failing as an abstraction when one or more of the following repeatedly hold:

1. module-internal design detail is written into `system_constraints`
2. directory structure, file ownership, or local protocol text is written into `system_constraints`
3. rules that really apply to only one or two modules are being promoted as global defaults
4. `system_constraints` keeps substituting for `shared_contract`
5. any engineering preference is elevated into a system-wide constraint with no real global need

Passing conclusion:

1. `system_constraints` still carries only true project-wide defaults

Failing conclusion:

1. `system_constraints` has become a global miscellany file

### 3.5 `project`

`project` is a valid abstraction only when all of the following mostly hold:

1. it mainly answers overall topology: which formal objects exist and how they connect
2. it can stably carry the combined surface of module, flow, shared, and system relations
3. without `project`, the current formal surface of the repository is hard to state in one place
4. it does not absorb module-local or flow-local body truth

`project` is drifting or failing as an abstraction when one or more of the following repeatedly hold:

1. `project` begins repeating the internal behavior of each `module`
2. `project` is forced to own implementation planning or ownership assignment
3. `project` cannot stay at topology level and must sink into detail to remain meaningful
4. `project` heavily duplicates `module` or `flow`
5. repository-wide surface changes cannot be expressed stably through `project`

Passing conclusion:

1. `project` still works as the topology object rather than a giant summary spec

Failing conclusion:

1. `project` has become a repeated index of everyone else's body text

---

## 4. Whole-System Checks

### 4.1 Mutual-Exclusion Check

The current abstraction set passes this check only when all of the following mostly hold:

1. `module`, `flow`, `shared_contract`, `system_constraints`, and `project` answer different layers of governance questions
2. when one governance question appears, there is usually one best landing object
3. it is relatively rare that one truth segment could plausibly belong in several of those objects at once

If this check repeatedly fails, boundary hardness is too weak.

### 4.2 Coverage Check

The current abstraction set passes this check only when all of the following mostly hold:

1. local capability boundaries have a stable owner
2. cross-module business paths have a stable owner
3. reusable local truth has a stable owner
4. project-wide default rules have a stable owner
5. overall topology relations have a stable owner

If a high-frequency governance problem repeatedly has nowhere to land, the abstraction set is incomplete.

### 4.3 Non-Template Check

The current abstraction set passes this check only when all of the following mostly hold:

1. these five objects are governance viewpoints, not repository folder templates
2. a user's physical project structure does not need to resemble those five object shapes before governance can begin
3. downstream mapping exists to interpret the repository into `specFlow`, not to force the repository into a required physical form

If this check fails, the framework is misusing governance abstractions as structural mandates.

---

## 5. Decision Rule

Continue improving downstream mechanisms such as mapping only when the current abstraction set still mostly passes:

1. the object-level checks
2. the mutual-exclusion check
3. the coverage check
4. the non-template check

Repair the abstraction first when at least one of the following holds:

1. any one object repeatedly hits several failure signals
2. the repository often cannot tell where one truth segment should land
3. the system increasingly depends on mapping or repair mechanisms to compensate for unstable object boundaries

Default risk notes:

1. `flow` tends to become heavy in repositories with weak business-path shape
2. `shared_contract` and `system_constraints` are the easiest pair to blur
3. `project` is the easiest object to over-expand into a duplicated summary layer

---

## 6. Non-Goals

This checklist does not:

1. prove that current files are well written
2. replace `spec_flow_review`
3. replace `spec_flow_design_review`
4. create a new lifecycle object
5. force every repository to use `flow` with the same density
