# Command Policy

## 1. Purpose

This file defines how formal commands work in this repository.

It answers six questions:

1. what a command is
2. which object families commands operate on
3. which commands are standard lifecycle commands
4. which objects are not command targets
5. which shared gate rules every command must follow
6. how natural-language requests enter the command and governance system

## 2. What A Command Is

A command is the standard workflow entry for one formal command-target object family.

In plain words:

1. `Spec` is the truth
2. `Command` is the action

## 3. Command-Target Object Families

This repository has two command-target object families:

1. `unit`
2. `scenario`

Shared notes:

1. both families write state into `docs/specs/_status.md`
2. both families may use `stable` and `candidate`
3. only `unit` owns direct implementation responsibility
4. `scenario` is a command target, but it is not a unit

Non-command objects:

1. `shared_contract` is not a standard command target
2. `system_constraints` is not a standard command target
3. `repository_mapping` is not a standard command target
4. `impact_sync` is an internal governance flow, not a user-facing standard command

## 4. Command Forms

This repository uses two user-facing command shapes:

1. `unit` command form:

```text
{command}:{unit}
```

2. `scenario` command form:

```text
{command}:{scenario}
```

Additional rules:

1. `system_constraints` is not a legal command target
2. `shared_contract` is not a legal standard command target
3. `repository_mapping` is not a legal standard command target
4. natural-language routing is the default user-facing entry for requests that do not use explicit command syntax
5. `shared_new`, `shared_extract`, `shared_bind`, `shared_topology`, `shared_sync`, `shared_escape`, and `impact_sync` are internal governance flows, not direct user-facing standard commands

## 5. Standard Commands

### 5.1 Unit Commands

1. `unit_init:{unit}`
2. `unit_stable_verify:{unit}`
3. `unit_new:{unit}`
4. `unit_fork:{unit}`
5. `unit_check:{unit}`
6. `unit_plan:{unit}`
7. `unit_impl:{unit}`
8. `unit_verify:{unit}`
9. `unit_promote:{unit}`

### 5.2 Scenario Commands

1. `scenario_new:{scenario}`
2. `scenario_stable_verify:{scenario}`
3. `scenario_fork:{scenario}`
4. `scenario_check:{scenario}`
5. `scenario_verify:{scenario}`
6. `scenario_promote:{scenario}`

### 5.3 Natural-Language Entry

The default user-facing entry is natural language.

Natural-language requests must follow:

1. `specflow/framework/natural_language_routing.md`
2. the routed command or governance-flow file

Rules:

1. a natural-language request must first be resolved into intent fragments
2. the executor must read the current repository truth needed to prove the route
3. if the request can be safely decomposed, only the first smallest legal step may be entered in the current handling round
4. if the request is missing target, scope, success meaning, acceptance meaning, or boundary truth, the executor must stop through the checkpoint protocol instead of guessing
5. if the request touches cross-unit shared truth, route into the shared-governance branch defined by `natural_language_routing.md`
6. direct shared command shapes are not user-facing command forms

### 5.4 Shared Governance Internal Routing

Shared governance is a branch of natural-language routing.

Rules:

1. users enter shared work by stating their shared intent in natural language
2. natural-language routing decides whether shared governance owns the request
3. the shared-governance branch routes directly into `shared_new`, `shared_extract`, `shared_bind`, `shared_topology`, `shared_sync`, or `shared_escape`
4. executors must not ask users to choose among `shared_new`, `shared_extract`, `shared_bind`, `shared_topology`, `shared_sync`, or `shared_escape`

## 6. Responsibilities By Family

### 6.1 Unit

`unit` commands own:

1. unit truth authoring
2. implementation planning
3. implementation work
4. implementation verification
5. promotion into stable unit truth

### 6.2 Scenario

`scenario` commands own:

1. trigger-to-outcome chain truth authoring
2. chain closure
3. end-to-end verification
4. promotion into stable scenario truth

`scenario` commands do not own:

1. implementation planning
2. implementation editing
3. unit-local repair

### 6.3 Repository Mapping

`repository_mapping` is consumed by commands, but it is not a command family.

It owns the current repository-structure truth:

1. governed-unit definition
2. support-surface rules
3. topology mapping
4. current formal object map
5. repository-level global constraint alignment

It does not own:

1. command lifecycle state
2. implementation planning
3. implementation editing
4. unit-local behavior authoring
5. scenario verification

## 7. Default Lifecycle Order

### 7.1 Unit

1. `unit_init`
2. `unit_stable_verify`
3. `unit_fork`
4. `unit_new`
5. `unit_check`
6. `unit_plan`
7. `unit_impl`
8. `unit_verify`
9. `unit_promote`

### 7.2 Scenario

1. `scenario_new`
2. `scenario_stable_verify`
3. `scenario_fork`
4. `scenario_check`
5. `scenario_verify`
6. `scenario_promote`

## 8. Shared Gate Rules

These rules apply by default to every command family:

1. do not execute a command if its prerequisite self-checks have not passed
2. process files are not valid just because they exist; their bound truth refs, fingerprints, and command-required fields must also match
3. a formal pass gate, formal verification pass, or lifecycle-state advance may be produced only by a new independent full-scope run of the corresponding command
4. after a command ends with any non-pass result other than a resumable checkpoint explicitly allowed by that command file, later repair or scoped recheck is non-authoritative for lifecycle progression
5. checkpoints are structured stops inside a command, not second lifecycles
6. `shared_contract`, `system_constraints`, and `repository_mapping` are always upstream inputs, never the primary output of `scenario` commands
7. commands that rely on repository path ownership must consume `docs/specs/repository_mapping.md`

### 8.1 Binding Drift

Candidate-side process files become invalid when any current required binding changes.

At minimum:

1. `unit` candidate process files fall back to `unit_check`
2. `scenario` candidate process files fall back to `scenario_check`

### 8.2 Stable Drift

Stable-layer alignment claims become invalid when any current required binding changes.

At minimum:

1. `unit` stable alignment falls back to `unit_stable_verify`
2. `scenario` stable alignment falls back to `scenario_stable_verify`

### 8.3 Shared And Baseline Inputs

1. if a command depends on bound `shared_contract` truth, it must read the exact currently bound shared files
2. if a command depends on the formal global baseline, it must read `docs/specs/system_constraints.md`
3. if a command depends on repository path ownership, it must read `docs/specs/repository_mapping.md`
4. `bound_objects`-only metadata drift does not by itself invalidate downstream process files

### 8.4 Impact Reconciliation

1. when one object family's truth or binding change may invalidate downstream objects, the handling round must complete deterministic downstream reconciliation before claiming closure
2. `shared_sync` remains the shared-governance impact-discovery flow for shared changes
3. `impact_sync` is the generic internal fallback-and-cleanup flow once the affected downstream object set is already fixed

### 8.5 Authoritative And Non-Authoritative Result Contract

Lifecycle progression may only come from one new, independent, full-scope command run.

Rules:

1. only one new full-scope run of the current command may produce a formal pass gate, a formal verification pass, or an advancing `_status.md` result
2. once a command has ended with a non-pass result, every later repair, local confirmation, scoped recheck, or follow-up assessment is non-authoritative unless that command file explicitly allows a checkpoint as a resumable stop
3. a non-authoritative follow-up may report that local repair is complete, but it must not claim new lifecycle progression, write advancing `_status.md` updates, or repackage a local recheck as a new formal pass
4. individual command files may tighten rerun conditions within their own boundary, but they must not weaken the authoritative / non-authoritative distinction defined here

### 8.6 User-Facing Close-Out Block Contract

Every formal command output must include a `user-facing close-out block`.

This block must report at least:

1. `round conclusion`
2. `current state`
3. `next step`
4. `why this next step`
5. `next-stage entry gap`
6. when the command enters a checkpoint or another explicit resumable stop, it must also report `resume signal`
7. individual command files may add stricter fields or wording requirements, but they must not delete the fixed fields defined here

## 9. Non-Goals

This file does not:

1. redefine object truth content in place of `spec_policy.md`
2. create a separate lifecycle for `shared_contract`
3. create a separate lifecycle for `system_constraints`
4. replace project-local standards registration
