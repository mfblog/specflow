# Command Policy

This file defines supported specFlow command entry forms.

specFlow recognizes only unit lifecycle commands plus framework governance commands. `scenario_*` commands are unsupported and must be rejected.

## 1. Standard Unit Commands

Standard unit commands use this form:

```text
{command}:{unit}
```

Supported unit commands are:

1. `unit_init:{unit}`
2. `unit_new:{unit}`
3. `unit_stable_verify:{unit}`
4. `unit_fork:{unit}`
5. `unit_check:{unit}`
6. `unit_plan:{unit}`
7. `unit_impl:{unit}`
8. `unit_verify:{unit}`
9. `unit_promote:{unit}`

No `scenario_*` command is supported. No command may treat `object-type=scenario` as a compatibility alias.

## 2. Unit Lifecycle

The standard unit lifecycle is:

```text
new or fork -> unit_check -> unit_plan -> unit_impl -> unit_verify -> unit_promote -> stable
```

Stable units may be checked by:

```text
unit_stable_verify -> unit_fork
```

`unit_check` validates that unit truth is sufficient and internally consistent.

`unit_plan` creates the implementation plan for the current candidate unit.

`unit_impl` changes implementation files only after the required plan and check evidence exists.

`unit_verify` validates the implementation against the current candidate unit truth.

`unit_promote` writes stable unit truth after verification passes.

## 3. Unit Dependency Rules

Each unit may list stable unit dependencies in frontmatter `unit_refs`.

Commands must apply these rules:

1. `unit_check` must reject or route back when the unit body relies on another unit's formal behavior but `unit_refs` does not record that dependency.
2. `unit_check` must reject candidate-layer `unit_refs`.
3. `unit_promote` must resolve `unit_refs` before stable writeback.
4. after a unit is promoted, `unit release-version` tooling must find current-layer units that still reference the promoted unit's previous stable version, retarget those `unit_refs` to the new stable version, and reroute them to the legal revalidation entry.
5. `unit_refs` never grants write permission to the referenced unit.

## 4. Rule Consumption

Stable global rules are repository-wide default inputs for every current-layer unit.

Bound shared rule consumers are derived only from current-layer unit frontmatter `rule_refs`.

Commands and governance flows must not read consumers from rule files. Rule files must not carry `bound_objects` as consumer truth.

## 5. Framework Commands

The following exact entries are not unit lifecycle commands:

1. `spec_flow_review`
2. `spec_flow_design_review`
3. `spec_flow_migrate`
4. `rule_new`
5. `rule_extract`
6. `rule_bind`
7. `rule_topology`
8. `rule_sync`
9. `rule_escape`

Each such entry is governed by its matching framework file.

## 6. Rejection Rules

The executor and tooling must reject:

1. any `scenario_*` command
2. `scenario_advance:{id}`
3. `--object-type scenario`
4. `docs/specs/scenarios/**` as a supported formal Spec path
5. Object Registry rows whose `kind` is not `unit` or `rule`

The rejection must be explicit. It must not silently convert a scenario request into a unit request.

## 7. Shared Command Gate Rules

Standard unit command files own their local preconditions, procedure, stop conditions, and output fields.

This file owns only the shared command contracts that every standard unit command inherits.

Rules:

1. do not execute a command if its prerequisite self-checks have not passed
2. process files are not valid just because they exist; their bound truth refs, fingerprints, and command-required fields must also match
3. a formal pass gate, formal verification pass, or lifecycle-state advance may be produced only by a new independent full-scope run of the corresponding command
4. after a command ends with any non-pass result other than a resumable checkpoint explicitly allowed by that command file, later repair or scoped recheck is non-authoritative for lifecycle progression
5. checkpoints are structured stops inside a command, not second lifecycles
6. `rule`, stable `g_` rule, and `repository_mapping` are upstream governance inputs, not standard lifecycle command targets
7. commands that rely on repository path ownership must consume `docs/specs/repository_mapping.md`
8. when a command uses slice-based work state, the command file must declare its adoption of `specflow/framework/slice_work_state_protocol.md`; the protocol supplies standards only and does not decide command adoption, carrier paths, slice catalogs, closure rules, or lifecycle progression

## 8. Rule Gate Rules

These rules apply by default to every standard unit command.

### 8.1 Binding Drift

Candidate-side process files become invalid when any current required binding changes.

At minimum:

1. truth or binding drift in unit candidate process files falls back to `unit_check`
2. process-shape, plan, evidence, implementation, and dependency-readiness failures follow the layered recovery targets in `specflow/framework/recovery_policy.md`

### 8.2 Stable Drift

Stable-layer alignment claims become invalid when any current required binding changes.

At minimum:

1. unit stable alignment falls back to `unit_stable_verify`

### 8.3 Rule And Global Rule Inputs

Rules:

1. if a command depends on bound `rule` truth, it must read the exact currently bound rule files
2. if a command depends on the formal global baseline, it must read `docs/specs/rules/stable/s_g_rule_repository_baseline.md`
3. if a command depends on repository path ownership, it must read `docs/specs/repository_mapping.md`
4. Stable global rules are repository-wide default inputs for every current-layer unit
5. Bound shared rule consumers are derived only from current-layer unit frontmatter `rule_refs`
6. Rule files must not record `bound_objects` as consumer truth

### 8.4 Impact Reconciliation

Rules:

1. when unit truth or binding changes may invalidate downstream units, the handling round must complete deterministic downstream reconciliation before claiming closure
2. `rule_sync` remains the rule-governance impact-discovery flow for rule changes
3. `impact_sync` is the generic internal fallback-and-cleanup flow once the affected downstream unit set is already fixed

### 8.4.1 Preflight Before Judgment

Commands must not make lifecycle, drift, fallback, cleanup, or promotion judgments from process-file contents before the required mechanical validation has passed.

Rules:

1. when a command consumes `_check_result`, `_plans/active`, or `_verify_result`, the first command-local judgment step must be `specflowctl command preflight` for the current command and unit
2. if `command preflight` is not available, the command must run each required `snapshot validate-process` command explicitly before reading the process file as a usable gate, plan, or verification result
3. a failed preflight may be used only as an entry stop or as input to the command's explicitly defined tool-backed fallback path
4. before authoritative validation succeeds, the command must not delete process files, update `_status.md`, write an active plan, write a verify result, write a stable truth file, or promote stable acceptance coverage
5. manual hash output, shell checksum output, editor display, conversation-derived values, and temporary script results are diagnostic only and must not classify drift, choose a failure layer, select cleanup, or advance lifecycle state
6. stable-layer verification commands that compare truth, Rule, repository mapping, or global-baseline fingerprints must use the fingerprint contract or deterministic tooling named by the active policy; if no authoritative comparison is available, they must report that alignment cannot be confirmed instead of claiming pass or drift from manual hashes
7. `command close` is the final tool-backed guard for standard lifecycle state progression; when a close outcome would continue the current command result or advance to a later command while consuming process-file input, it must run the same mechanical preflight internally before writing `_status.md`, cleaning process files, or reporting the close as successful
8. explicit fallback and recovery close outcomes may run without valid input process files only when the active command file defines that outcome as the legal recovery path
9. low-level status tools such as `status set-object` and manual `_status.md` edits are not valid substitutes for `command close` in ordinary lifecycle progression

### 8.5 Authoritative And Non-Authoritative Result Contract

Lifecycle progression may only come from one new, independent, full-scope command run.

Rules:

1. only one new full-scope run of the current command may produce a formal pass gate, a formal verification pass, or an advancing `_status.md` result
2. once a command has ended with a non-pass result, every later repair, local confirmation, scoped recheck, or follow-up assessment is non-authoritative unless that command file explicitly allows a checkpoint as a resumable stop
3. a non-authoritative follow-up may report that local repair is complete, but it must not claim new lifecycle progression, write advancing `_status.md` updates, or repackage a local recheck as a new formal pass
4. individual command files may tighten rerun conditions within their own boundary, but they must not weaken the authoritative and non-authoritative distinction defined here
5. when a command enters recovery mode because repository mutation started but the command cannot safely close, the command must follow `specflow/framework/recovery_policy.md` for layered recovery before any checkpoint answer can be processed or before the next command may enter

### 8.6 User-Facing Close-Out Block Contract

Every formal command output must include a `user-facing close-out block`.

Formal command close-out output inherits the framework output baseline defined by `specflow/framework/output_baseline.md`.
The fields listed below are the command-specific minimum for close-out blocks on top of the baseline.
Fields not applicable in a given close-out context are covered by the baseline's escape hatch for non-applicable items.
Command files may tighten or clarify close-out wording, ordering, and execution-note separation within their own output contract.
They must not affect command result types, lifecycle advancement, `_status.md`, `_check_result` writeback, fallback selection, checkpoint semantics, or command-local required fields.

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
