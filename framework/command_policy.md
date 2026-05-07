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

Commands are not the user's required vocabulary.
Natural-language routing may translate an ordinary user goal into one or more command chains internally, but each command still owns only its own lifecycle boundary and may advance only by its own command rules.

## 3. Command-Target Object Families

This repository has two command-target object families:

1. `unit`
2. `scenario`

Rule notes:

1. both families write state into `docs/specs/_status.md`
2. both families may use `stable` and `candidate`
3. only `unit` owns direct implementation responsibility
4. `scenario` is a command target, but it is not a unit

Non-command objects:

1. `rule` is not a standard command target
2. stable `g_` rule is not a standard command target
3. `repository_mapping` is not a standard command target
4. `impact_sync` is an internal governance flow, not a user-facing standard command
5. `spec_flow_migrate` is a project-instance migration governance entry, not a standard command target

## 4. Command Forms

This repository uses two user-facing command shapes:

1. `unit` command form:

```text
{command}:{unit}
```

1. `scenario` command form:

```text
{command}:{scenario}
```

Additional rules:

1. stable `g_` rule is not a legal command target
2. `rule` is not a legal standard command target
3. `repository_mapping` is not a legal standard command target
4. natural-language routing is the default user-facing entry for requests that do not use explicit command syntax
5. `rule_new`, `rule_extract`, `rule_bind`, `rule_topology`, `rule_sync`, `rule_escape`, and `impact_sync` are internal governance flows, not direct user-facing standard commands
6. `spec_flow_migrate` is entered by its exact name and is governed by `specflow/framework/spec_flow_migrate.md`; it must not be written as `{command}:{unit}` or `{command}:{scenario}`

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

Natural-language entry is a user-goal governance entry, not a command-alias system.
It diagnoses the user's goal, reads the current repository truth needed for routing, chooses the legal specFlow route internally, and reports the current state and next action through user-goal language, project-structure language, and plain engineering action language.
Internal routing names are trace details, not the user's required decision language.

Natural-language requests must follow:

1. `specflow/framework/natural_language_routing.md`
2. the routed command or governance-flow file

Rules:

1. a natural-language request must first be diagnosed as a user goal before command ownership is chosen
2. the executor must classify the work shape and resolve formal ownership from current repository truth
3. the executor must read the current repository truth needed to prove the route
4. if the request can be safely decomposed, only the first smallest legal step may be entered in the current handling round
5. natural-language routing may assemble an internal chain across multiple existing command families or governance flows, but that chain is not permission to skip a command gate or continue after the first step without rerouting from current truth
6. if the request is missing target, scope, success meaning, acceptance meaning, or boundary truth, the executor must stop through the checkpoint protocol instead of guessing
7. checkpoint questions and ordinary user-facing route reports must not require the user to choose internal object-family names, command names, lifecycle state names, or internal rule-governance flow names
8. if the request touches cross-unit rule truth, route into the rule-governance branch defined by `natural_language_routing.md`
9. direct shared command shapes are not user-facing command forms
10. natural-language requests to update old project-instance files to current `specFlow` framework contracts route to `specflow/framework/spec_flow_migrate.md`

### 5.4 Project-Instance Migration Entry

`spec_flow_migrate` owns project-instance format migration after a framework update.

Rules:

1. exact input `spec_flow_migrate` routes directly to `specflow/framework/spec_flow_migrate.md`
2. `spec_flow_migrate` is not a `unit` or `scenario` lifecycle command
3. `spec_flow_migrate` may not advance a `unit` or `scenario` lifecycle gate
4. `spec_flow_migrate` may invalidate stale process state only under the migration policy and the shared process-state rules it links
5. `spec_flow_migrate` must stop instead of choosing business meaning, object ownership, acceptance meaning, rule-truth ownership, or global-rule meaning

### 5.5 Rule Governance Internal Routing

Rule governance is a branch of natural-language routing.

Rules:

1. users enter rule work by stating their rule intent in natural language
2. natural-language routing decides whether rule governance owns the request
3. the rule-governance branch routes directly into `rule_new`, `rule_extract`, `rule_bind`, `rule_topology`, `rule_sync`, or `rule_escape`
4. executors must not ask users to choose among `rule_new`, `rule_extract`, `rule_bind`, `rule_topology`, `rule_sync`, or `rule_escape`

## 6. Responsibilities By Family

### 6.1 Unit

`unit` commands own:

1. unit truth authoring
2. implementation planning
3. implementation work
4. implementation verification
5. promotion into stable unit truth

`unit` commands may be one part of a larger natural-language development chain, but they do not own end-to-end user-flow closure unless that closure is already represented as unit-local acceptance truth.

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

When a scenario route discovers that implementation work is still required in affected units, those units must return to their own legal `unit` command chains.
Scenario commands must not repair or advance unit implementation on behalf of those units.

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

Command-target truth file resolution is not stored as a current concrete path in `repository_mapping`.
Commands must resolve the current main Spec file by combining:

1. the object row in `docs/specs/_status.md`
2. the stable or candidate path template defined in `specflow/framework/spec_policy.md`
3. the object's `truth_surface_rule` in `docs/specs/repository_mapping.md`

Changing only `Active Layer` through `unit_fork`, `unit_promote`, `scenario_fork`, or `scenario_promote` does not require a repository mapping update.
Repository mapping changes are required only when the object map, truth-surface rule, implementation surface, rule path, support surface, governed root, ignore rule, or conflict rule changes.

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

## 8. Rule Gate Rules

These rules apply by default to every command family:

1. do not execute a command if its prerequisite self-checks have not passed
2. process files are not valid just because they exist; their bound truth refs, fingerprints, and command-required fields must also match
3. a formal pass gate, formal verification pass, or lifecycle-state advance may be produced only by a new independent full-scope run of the corresponding command
4. after a command ends with any non-pass result other than a resumable checkpoint explicitly allowed by that command file, later repair or scoped recheck is non-authoritative for lifecycle progression
5. checkpoints are structured stops inside a command, not second lifecycles
6. `rule`, stable `g_` rule, and `repository_mapping` are always upstream inputs, never the primary output of `scenario` commands
7. commands that rely on repository path ownership must consume `docs/specs/repository_mapping.md`

### 8.1 Binding Drift

Candidate-side process files become invalid when any current required binding changes.

At minimum:

1. truth or binding drift in `unit` candidate process files falls back to `unit_check`
2. truth or binding drift in `scenario` candidate process files falls back to `scenario_check`
3. process-shape, plan, evidence, implementation, and dependency-readiness failures follow the layered recovery targets in `specflow/framework/recovery_policy.md`

### 8.2 Stable Drift

Stable-layer alignment claims become invalid when any current required binding changes.

At minimum:

1. `unit` stable alignment falls back to `unit_stable_verify`
2. `scenario` stable alignment falls back to `scenario_stable_verify`

### 8.3 Rule And Global Rule Inputs

1. if a command depends on bound `rule` truth, it must read the exact currently bound rule files
2. if a command depends on the formal global baseline, it must read `docs/specs/rules/stable/s_g_rule_repository_baseline.md`
3. if a command depends on repository path ownership, it must read `docs/specs/repository_mapping.md`
4. `bound_objects`-only metadata drift does not by itself invalidate downstream process files

### 8.4 Impact Reconciliation

1. when one object family's truth or binding change may invalidate downstream objects, the handling round must complete deterministic downstream reconciliation before claiming closure
2. `rule_sync` remains the rule-governance impact-discovery flow for rule changes
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

Formal command close-out output is part of the shared `specflow_response` / `user_facing_response_clarity` output surface defined by `specflow/framework/project_standards_policy.md`.
Registered project-local standards selected by that surface may tighten or clarify only command close-out wording, ordering, and execution-note separation.
They must not affect command result types, lifecycle advancement, `_status.md`, `_check_result` writeback, fallback selection, checkpoint semantics, or command-local required fields.
Command files inherit this shared output surface through this section and must not restate the registry shape in each command file.

This block must report at least:

1. `round conclusion`
2. `current state`
3. `next step`
4. `why this next step`
5. `next-stage entry gap`
6. when the command enters a checkpoint or another explicit resumable stop, it must also report `resume signal`
7. individual command files may add stricter fields or wording requirements, but they must not delete the fixed fields defined here

User-facing close-out language rules:

1. the block must be understandable without internal specFlow knowledge
2. it must use user-goal language first, project-structure language second, and plain engineering action language third
3. project-structure language means the current repository's capability areas, delivery surfaces, entry points, and responsibility areas as proven by current repository truth or named by the user
4. project-structure language must not become a raw directory listing when a responsibility phrase is available
5. if current repository truth does not clearly identify the relevant project structure, the block must say that structure ownership is unclear instead of inventing a friendly label
6. internal command names, lifecycle state names, object-family names, policy-file names, and formal route names must not appear as the recommended action in the user-facing block unless the user explicitly asked for those internal details
7. traceability details may appear only in a short execution note after the user-facing close-out block
8. the execution note may record internal state, command names, file paths, and policy inputs, but it must not be required for the user to understand the conclusion, next step, reason, or remaining gap

### 8.7 Command-File Economy Contract

Command files inherit the shared command rules in this section by default.

Rules:

1. command files must not restate the full text of Section 8.5 when a short inheritance sentence is enough
2. command files must state only command-local additions when they tighten rerun entry, checkpoint handling, fallback handling, output fields, or stop conditions
3. command files must not create a second definition for the user-facing close-out block; they may only add command-local fields or stricter wording requirements on top of Section 8.6
4. command files may include a short read summary before dense procedure text; that summary is navigation only and must not override the command's preconditions, procedure, stop conditions, or output contract

### 8.8 Lifecycle-Advance Inheritance

Every standard command inherits the authoritative and non-authoritative result contract from Section 8.5.

Rules:

1. when a command advances `_status.md`, writes a formal pass gate, or writes a formal verification pass, that advancement is valid only from a new independent full-scope run of that command
2. command-local follow-up checks after a non-pass result remain non-authoritative unless the command file explicitly defines a resumable checkpoint
3. command files may tighten how a fresh full-scope rerun is recognized, but they must not allow a repair-only or scoped follow-up to advance lifecycle state

## 9. Non-Goals

This file does not:

1. redefine object truth content in place of `spec_policy.md`
2. create a separate lifecycle for `rule`
3. create a separate lifecycle for stable `g_` rule
4. replace project-local standards registration
