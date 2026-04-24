# Repository Mapping Policy

## 1. Purpose

This file defines `repository_mapping`.

It answers five questions:

1. what `repository_mapping` owns
2. what it must not own
3. which file carries it
4. how command objects consume it
5. how repository-structure drift is handled

`repository_mapping` is not a command-target object.
It is the current repository-structure truth that lets humans and agents share one explicit view of how real repository paths map to Spec Flow objects.

## 2. Object Definition

`repository_mapping` is the repository structure truth file.

It must answer these sections:

1. `Project Overview`
   - what this repository is for
   - the main delivery surface of the repository
   - the shortest useful reading path for a human or agent entering the repository
2. `Governed Object Map`
   - current `unit` IDs and one-line responsibilities
   - current `scenario` IDs and one-line responsibilities, or `none`
   - current `shared_contract` IDs and one-line responsibilities
3. `Boundary Rules`
   - what qualifies as a formal `unit`
   - what must become `shared_contract`
   - what remains outside command-target truth
4. `Path Ownership`
   - governed roots
   - ignore rules
   - unit truth and implementation surfaces
   - shared-contract truth paths
   - support-surface paths
   - conflict resolution order
5. `Global Constraint Alignment`
   - which `system_constraints` version currently constrains the repository mapping
6. `Drift Handling`
   - what counts as mapping drift
   - which consuming command must stop when the mapping no longer matches the repository
   - how the mapping is repaired

It does not answer:

1. one unit's local behavior
2. one scenario's trigger-to-outcome chain semantics
3. one shared-contract body's field-level rule text
4. implementation planning or implementation ownership
5. lifecycle state progression

## 3. File

`repository_mapping` has one current file:

1. `docs/specs/repository_mapping.md`

Rules:

1. `repository_mapping` does not use `stable` and `candidate` layers.
2. `repository_mapping` does not enter `docs/specs/_status.md`.
3. `repository_mapping` is not promoted.
4. `repository_mapping` is not forked.
5. changes to `repository_mapping` are normal repository-truth edits and must be reviewed or committed according to `git_policy.md`.

## 4. Consumers

These flows consume `repository_mapping`:

1. `unit` commands
2. `scenario` commands
3. `shared_ops` and its routed internal shared-governance flows
4. governance reviews
5. repository health checks such as `doctor` or future mapping checks

Consumption rules:

1. a command that relies on path ownership must read `docs/specs/repository_mapping.md` before claiming a boundary-sensitive result
2. a command may use `repository_mapping` to decide whether the target path belongs to a `unit`, `scenario`, `shared_contract`, `support_surface`, or `ignore`
3. a command must not rewrite `repository_mapping` as an incidental side effect of implementation work
4. when the command discovers that the mapping is incomplete or conflicts with the current repository, it must stop and require a `repository_mapping` truth update before continuing

## 5. Drift Handling

Mapping drift exists when current repository reality no longer matches `docs/specs/repository_mapping.md`.

At minimum, drift includes:

1. a governed path is not mapped to any formal object, support surface, or ignore rule
2. a path maps to more than one command-target object
3. a declared unit truth path does not exist
4. a declared shared-contract truth path does not exist
5. a declared support surface has moved without the mapping being updated
6. a consuming command's target path is outside the ownership declared for that target object

Handling rules:

1. consuming commands must stop instead of guessing a new mapping
2. the next required action is to update `docs/specs/repository_mapping.md`
3. after the mapping update, rerun the original command from its normal legal entry point
4. if a mapping update also changes unit behavior truth, shared-contract truth, scenario truth, or system constraints, route that separate truth change through the corresponding object rules

## 6. Non-Goals

This file does not:

1. create `repository_mapping_*` lifecycle commands
2. create a repository-level command family
3. replace `unit` truth
4. replace `scenario` truth
5. replace `shared_contract` truth
6. create an independent lifecycle for `system_constraints`
