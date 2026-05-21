# Slice Work-State Protocol

## 1. Purpose

This file defines the common rules for slice-based work state.

It is command-independent.
It does not decide which command or governance flow must use slices.
It does not define lifecycle progression, review scope, pass criteria, verification criteria, or output conclusions.

An adopting command or governance flow owns all of the following:

1. whether it uses this protocol
2. which state carrier records slice state
3. which slice fields are required
4. which baseline slices must exist
5. whether dynamic slices are allowed
6. whether cross-convergence slices are required
7. what closes a slice or slice set
8. whether a closed slice set can support lifecycle progression

This protocol defines only the shared standards that apply after an adopting owner chooses to use slice-based work state.

## 2. Core Terms

A `slice` is one bounded unit of work, review, planning, implementation progress, or evidence coverage.
The adopting owner defines the business meaning of each slice.

A `slice set` is the collection of slices that one adopting owner requires for one round.
The adopting owner defines whether a slice set is fixed, extensible, or both.

A `state carrier` is the file or table that records slice progress for one round.
It may be a dedicated work-state file, a review run-state file, a plan file, a verification result file, or another process file explicitly named by the adopting owner.

A `baseline slice` is a slice required by the adopting owner before execution starts.
Baseline slices are the owner's minimum coverage outline.

A `dynamic slice` is a slice added during execution because the existing slice set does not fully cover a discovered risk, dependency, work surface, evidence surface, or convergence path.
Dynamic slices may only increase coverage.
They must not replace required baseline slices.

A `cross-convergence slice` is a slice that checks whether multiple local slices, work surfaces, evidence surfaces, or rule areas compose into one coherent result.
The adopting owner defines which local slices or surfaces it depends on.

An `input_fingerprint` is a deterministic fingerprint of the declared inputs for a slice.
It exists to make hidden input drift visible.

`stale` means the previous slice result can no longer be reused because the slice's declared inputs changed, disappeared, or inherited stale status from a dependency.

`resume_next_step` is the smallest next action that lets the adopting owner continue the same round or restart through the legal owner.
It is not a pass result.

## 3. State Carrier Rules

A state carrier is process state.
It is not:

1. a Spec
2. behavior truth
3. rule truth
4. repository ownership truth
5. implementation truth
6. a final conclusion unless the adopting owner explicitly defines that conclusion elsewhere

A state carrier must not create a second source of truth for behavior, acceptance, rule meaning, ownership, implementation requirements, or review judgment.

The adopting owner must state:

1. the carrier path or carrier section
2. whether the carrier is downstream-consumable or only a resume aid
3. the required run-level fields
4. the required slice-level fields
5. the legal status values
6. the freshness and stale rules
7. the cleanup or invalidation owner

If the adopting owner does not define one of those items, an executor must not infer it from another command, another flow, a similarly named file, or a previous conversation.

## 4. Slice Field Rules

When a state carrier records generic slice rows, each slice row should identify at least:

1. `slice_id`
2. `slice_origin`
3. `slice_type`
4. `status`
5. `review_question`, `work_goal`, or another adopting-owner-defined purpose field
6. `why_added`
7. `parent_slice_id` when dynamic slices are allowed
8. `input_files` or another adopting-owner-defined input reference field
9. `input_fingerprint` when the carrier uses freshness checks
10. `depends_on` when stale status can propagate from dependencies
11. `finding_refs`, `evidence_refs`, `progress_refs`, or another adopting-owner-defined result reference field
12. `result_summary`
13. `exit_condition`
14. `resume_next_step`

The adopting owner may use domain-specific field names when the carrier is a business process file rather than a generic slice table.
Those domain-specific fields must still make the same facts visible: purpose, coverage target, status, evidence or progress, blocker, and next action.

## 5. Status Rules

The adopting owner must define the legal slice status values.

When the owner uses the generic review-style status set, the meanings are:

1. `pending`
   - the slice has not been judged or completed
2. `passed`
   - the slice's required judgment or work result has been completed under the adopting owner's criteria
3. `blocked`
   - the slice cannot close without a named repair, decision, prerequisite, or reroute
4. `stale`
   - the slice's prior result cannot be reused because its inputs or dependencies changed
5. `skipped_not_applicable`
   - the adopting owner explicitly declares the slice out of scope for this round and records why
6. `skipped_not_in_scope`
   - a review owner explicitly declares the slice outside a narrowed review scope and records why

Business process files may use domain-specific status values instead.
Those values must not imply a formal review pass, verification pass, or lifecycle advance unless the adopting command explicitly defines that implication.

## 6. Freshness Rules

If the adopting owner uses input freshness, it must define:

1. which inputs are fingerprinted
2. which normalization rules are used
3. when fingerprints are created
4. when fingerprints are refreshed
5. which status changes are allowed when inputs change
6. whether stale status propagates through `depends_on`
7. what must happen before a stale slice can be reused

Manual hash output, shell checksum output, editor display, temporary scripts, and conversation-derived values are diagnostic only unless the adopting owner explicitly makes a deterministic tooling entry authoritative.

If one slice depends on another slice, and the dependency becomes stale, the dependent cross-convergence slice must not remain closed unless the adopting owner explicitly defines a narrower safe reuse rule.

## 7. Dynamic Slice Rules

Dynamic slices are allowed only when the adopting owner explicitly allows them.

When allowed, a dynamic slice must:

1. state why the existing slice set did not fully cover the discovered work
2. name a `parent_slice_id` that already exists in the same carrier
3. state whether it is local or cross-convergence when the carrier distinguishes slice types
4. declare its inputs
5. declare its exit condition
6. close before the adopting owner claims a complete slice-set result

A dynamic slice must not:

1. replace a required baseline slice
2. weaken a baseline slice's exit condition
3. hide a cross-area risk inside a local slice when the adopting owner requires cross-convergence
4. let a command continue after discovering that another owner must first change durable truth

## 8. Cross-Convergence Rules

Cross-convergence is required only when the adopting owner says it is required.

When required, a cross-convergence slice must:

1. name the local slices, work surfaces, evidence surfaces, or rule areas it depends on
2. check composition, not only local completion
3. fail or block when locally plausible results do not combine into the adopting owner's required whole
4. become stale when a required dependency becomes stale, unless the adopting owner defines a narrower safe reuse rule

A cross-convergence slice must not be treated as complete only because all dependency slices were visited.
It must answer the adopting owner's convergence question.

## 9. Tooling Boundary

Deterministic tooling may maintain slice work state only when an adopting owner defines the exact carrier, fields, statuses, freshness rules, and writeback target.

Allowed mechanical tooling actions are limited to:

1. creating carrier skeletons
2. creating baseline skeleton rows
3. validating field presence and legal values
4. validating parent links and dependency references
5. writing UTC timestamps
6. computing input fingerprints
7. marking stale status caused by changed or missing inputs
8. propagating stale status through declared dependencies

Tooling must not:

1. decide that a semantic slice passed
2. write finding content
3. choose finding severity
4. decide review scores
5. decide whether verification evidence is sufficient
6. decide final conclusions
7. choose lifecycle outcomes
8. create a new durable carrier that the adopting owner did not define

## 10. Adoption Contract

Any command or governance flow that adopts this protocol must state its adoption rules in its own owner file.

The adoption rules must answer:

1. where slice state is recorded
2. which required fields or domain-specific equivalent fields are used
3. which baseline slices or business slices must exist
4. whether dynamic slices are allowed
5. whether cross-convergence is required
6. how input freshness and stale handling work
7. what closes a slice
8. what closes the whole slice set
9. whether closure can support a lifecycle advance, pass claim, verification claim, or only local progress
10. what happens when slice work discovers missing durable truth, unclear ownership, or a required upstream decision

The protocol does not supply defaults for those decisions.
If an adopting owner omits one of them, the missing decision remains undefined and must not be guessed.

## 11. Non-Goals

This protocol does not:

1. define a command list
2. classify commands or flows into modes
3. route user requests
4. define lifecycle transitions
5. define pass, blocked, fix-required, score, promotion, or verification criteria
6. require any command to use a dedicated work-state file
7. require any business process file to use generic slice table names
8. let tooling replace command judgment or review judgment
