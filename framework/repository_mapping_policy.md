# Repository Mapping Policy

## 1. Purpose

This file defines `repository_mapping`.

It answers seven questions:

1. what `repository_mapping` owns
2. what it must not own
3. which file carries it
4. how command objects consume it
5. how repository-structure drift is handled
6. when a flow must read it
7. how a flow checks it against current repository reality

`repository_mapping` is not a command-target object.
It is the current repository-structure truth that lets humans and agents share one explicit view of how real repository paths map to Spec Flow objects.

## 2. Object Definition

`repository_mapping` is the repository structure truth file.

It must answer these sections:

1. `Project Overview`
   - what this repository is for
   - the main delivery surface of the repository
   - the shortest useful reading path for a human or agent entering the repository
2. `Object Registry`
   - one fixed Markdown table that registers every current or planned `unit`, `scenario`, and `rule`
   - the table is the only machine-readable registry that connects those objects to implementation paths
   - the fixed header is `| kind | id | scope | registration_state | implementation_paths | spec_files | responsibility |`
   - `kind` must be `unit`, `scenario`, or `rule`
   - `scope` must be `capability` for `unit`, `flow` for `scenario`, and `bound` or `global` for `rule`
   - `registration_state` must be `planned` or `landed`
   - `planned` rows must use `implementation_paths=none`
   - `landed` rows must list concrete implementation paths
   - `spec_files` records the related Spec or rule documents and does not decide the registration state
3. `Boundary Rules`
   - what qualifies as a formal `unit`
   - what must become `rule`
   - what remains outside command-target truth
4. `Path Ownership`
   - governed roots
   - ignore rules
   - command-target truth path rules and implementation surfaces
   - rule truth paths
   - support-surface paths
   - conflict resolution order
5. `Rule Alignment`
   - which stable `g_` rule version currently constrains the repository mapping
6. `Drift Handling`
   - what counts as mapping drift
   - which consuming command must stop when the mapping no longer matches the repository
   - how the mapping is repaired

It does not answer:

1. one unit's local behavior
2. one scenario's trigger-to-outcome chain semantics
3. one rule body's field-level rule text
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
5. changes to `repository_mapping` are normal repository-truth edits and must be reviewed through the active governance route when the route requires review.

## 4. Consumers

These flows consume `repository_mapping`:

1. `unit` commands
2. `scenario` commands
3. the rule-governance branch and its routed internal rule flows
4. governance reviews
5. repository health checks such as `doctor` or future mapping checks

Consumption rules:

1. a command that relies on path ownership must read `docs/specs/repository_mapping.md` before claiming a boundary-sensitive result
2. a command may use `repository_mapping` to decide whether the target path belongs to a `unit`, `scenario`, `rule`, `support_surface`, or `ignore`
3. a command must not rewrite `repository_mapping` as an incidental side effect of implementation work
4. when the command discovers that the mapping is incomplete or conflicts with the current repository, it must stop and require a `repository_mapping` truth update before continuing
5. a command must not expect `repository_mapping` to name the current active `unit` or `scenario` main Spec file directly; it must resolve that file from `_status.md` and the templates defined by `spec_policy.md`

## 5. Drift Handling

Mapping drift exists when current repository reality no longer matches `docs/specs/repository_mapping.md`.

At minimum, drift includes:

1. a governed path is not mapped to any formal object, support surface, or ignore rule
2. a path maps to more than one command-target object
3. a command-target truth path resolved from `_status.md` and the declared path rule does not exist
4. a declared rule truth path does not exist
5. a declared support surface has moved without the mapping being updated
6. a consuming command's target path is outside the ownership declared for that target object
7. a command-target object still lists a concrete current-layer truth file under its mapping entry instead of naming a truth-surface rule
8. an `Object Registry` row is malformed, uses an unsupported field value, declares a missing landed implementation path, or omits a formal object that must be registered

Handling rules:

1. consuming commands must stop instead of guessing a new mapping
2. the next required action is to update `docs/specs/repository_mapping.md`
3. after the mapping update, rerun the original command from its normal legal entry point
4. if a mapping update also changes unit behavior truth, rule truth, scenario truth, or global rules, route that separate truth change through the corresponding object rules

## 6. Read Trigger Rules

`repository_mapping` must not be read just because a request is inside specFlow.
It is read only when the current task needs repository-structure judgment.

Read `docs/specs/repository_mapping.md` when at least one of these is true:

1. the task needs to decide which `unit`, `scenario`, `rule`, `support_surface`, or `ignore` owns one or more repository paths
2. the task creates, removes, moves, or renames repository paths under a governed root
3. the task creates a new formal object or changes the object map of an existing formal object
4. the task changes a declared truth-surface rule, implementation surface, rule path, support-surface path, governed root, ignore rule, or conflict rule
5. the task is a direct implementation request and the executor must classify whether the requested file changes fit the current formal object boundary
6. the rule-governance branch or an internal rule flow must determine affected downstream objects from current repository structure
7. a governance review, repository health check, or explicit user request asks whether repository structure and mapping still match

Do not read `docs/specs/repository_mapping.md` only for these tasks:

1. explaining behavior already contained in a named current-layer `unit` or `scenario` truth file
2. reading `docs/specs/_status.md` to report `Active Layer` or `Next Command`
3. validating process-file snapshots when the command does not need path ownership or object-boundary judgment
4. reading or updating stable `g_` rule when the change does not alter repository structure
5. editing a current-layer truth file whose target path has already been resolved by the command and whose object boundary is not in question
6. changing only a command-target object's `Active Layer` through a legal fork or promote command

Read scope rules:

1. default to a local read: inspect only the mapping sections needed for the current target object, the current changed paths, and the conflict rules
2. use an expanded read when the task adds, removes, moves, or renames paths under a governed root; include the affected parent path and sibling mapping rules that may conflict
3. use a full mapping read only for new or unfamiliar repositories, object-map changes, support-surface changes, governed-root changes, ignore-rule changes, repository-wide reviews, repository health checks, or explicit user requests
4. after reading, the flow must state whether the mapping was not needed, locally checked, expanded-checked, or fully checked when that distinction matters for the command result

## 7. Consistency Check Procedure

A consistency check compares current repository reality against `docs/specs/repository_mapping.md`.
It is scoped by the read trigger unless the current task explicitly requires a full repository check.

The default procedure is:

1. collect the relevant path set
   - include paths explicitly named by the user
   - include paths the command plans to read, write, create, move, rename, or delete
   - include current truth files resolved from `_status.md` for the target object
   - include parent or sibling paths only when conflict detection needs them
2. classify each relevant path by the mapping's conflict order
   - current command-target truth file path resolved from `_status.md` and the mapping's truth-surface rule
   - declared implementation surface
   - rule truth path
   - support surface
   - ignore
   - unmapped
3. compare the classification with the command target
   - a `unit` command may operate only inside that unit's declared truth or implementation surface unless the command explicitly owns another formal writeback
   - a `scenario` command may operate on scenario truth and declared scenario bindings, but must not claim unit-local implementation ownership
   - a rule-governance flow may operate on declared rule truth and binding metadata, but must not silently rewrite unit behavior truth
   - support-surface edits may continue only when the current task explicitly targets that support surface or a governance flow owns it
4. check existence of declared files that are relevant to the current task
   - landed `Object Registry` implementation paths must exist
   - declared `Object Registry` spec files must exist
   - target object truth files must exist when `_status.md` and the mapping's truth-surface rule resolve to them
   - selected rule truth files must exist when they are part of the current binding or current shared scope
   - selected support-surface files or directories must exist when the task depends on them
5. detect conflicts
   - a relevant path that maps to more than one command-target object is mapping drift
   - a relevant governed path that maps to no formal object, support surface, or ignore rule is mapping drift
   - a command target path outside the target object's declared ownership is mapping drift
   - a command-target object entry that lists concrete active truth files instead of a truth-surface rule is mapping drift
6. decide the result
   - if no drift is found, continue the original flow
   - if drift is found, stop the original flow before boundary-sensitive work continues
   - report the concrete drift path, the observed repository fact, the mapping rule that failed, and the required mapping update
7. repair only the mapping when the mismatch is structural
   - update `docs/specs/repository_mapping.md`
   - rerun the original command from its normal legal entry point
8. do not repair behavior truth through mapping
   - if the mismatch changes unit behavior, scenario behavior, rule rules, or global rules, route that separate truth change through the corresponding object rules

Full repository checks must additionally:

1. enumerate all governed roots declared by the mapping
2. classify all tracked paths under those roots, excluding ignore rules
3. verify that every declared formal object has its declared implementation paths when its registry state is `landed`
4. verify that no tracked governed path is both unmapped and relevant to governance
5. verify that no tracked governed path maps to multiple command-target objects
6. verify that every `_status.md` object has an `Object Registry` row

## 8. Non-Goals

This file does not:

1. create `repository_mapping_*` lifecycle commands
2. create a repository-level command family
3. replace `unit` truth
4. replace `scenario` truth
5. replace `rule` truth
6. create an independent lifecycle for stable `g_` rule
