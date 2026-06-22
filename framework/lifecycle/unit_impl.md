# Unit Implementation

`unit_impl:{unit}` enters the implementation phase.
It provides implementation context and boundaries without changing lifecycle state.
The `Next Command` field contains `unit_check, unit_impl, unit_verify` during this phase.

## Input

> **Reading guidance:** Must Read files are the truth and process data this command evaluates. May Reference files hold the format and policy contracts referenced by the checks — read them when a specific question needs the exact rule text. Procedural instructions are inline in "What This Step Does" and "How to End" below.

### Must Read

- `docs/specs/_status.md` — confirm `Next Command` contains `unit_impl`
- `docs/specs/units/candidate/c_unit_{unit}.md` — acceptance items
- `docs/specs/repository_mapping.md` — implementation path ownership

### May Reference

- `framework/process_snapshot_contract.md` (constraints and phase rules)
- `framework/spec_writing_guide.md` (unit Spec format and appendix format)
- `framework/core/status.md` (constraints derivation, phase write boundaries)

## Pre-Execution Self-Check (MANDATORY)

Before executing this step, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm the target unit's `Next Command` contains `unit_impl`.
2. [ ] Read `docs/specs/units/candidate/c_unit_{unit}.md` — confirm it exists with valid acceptance items.
3. [ ] If any check fails: STOP, report what is missing, and do not proceed.

If all checks pass: proceed to "What This Step Does" below.

## What This Step Does

- Provide the agent with unit identity and acceptance item location
- Agent implements independently based on candidate truth
- Conversational iteration during implementation is normal
- The agent is free to determine how to implement
- After creating or modifying implementation files, update `docs/specs/repository_mapping.md` with the implementation paths and set `registration_state=landed`

## If Spec Issues Are Found During Implementation

If acceptance items are incomplete, incorrect, or unclear:

1. Stop implementation
2. Report the issue to the user
3. Fix the candidate spec (`docs/specs/units/candidate/c_unit_{unit}.md`)
4. Run `unit_check:{unit}` to re-validate the modified spec — this is accepted as a
   re-validation during the implementation phase (see `unit_check.md` precondition
   exception). `unit_check` re-runs the quality checks defined in `unit_check.md` against the modified spec.
5. After `unit_check` passes, resume with `unit_impl:{unit}`

## On-Demand References

Agent may read these as needed during implementation:

- `docs/specs/units/candidate/c_unit_{unit}.md` — acceptance items
- `docs/specs/repository_mapping.md` — file path ownership
- Bound shared rules — constraints
- Dependent unit implementations — interface alignment

## Not Allowed

- Modify lifecycle state (`_status.md`) except through `command close`
- Implement behavior beyond the unit's acceptance items
- Modify candidate spec (`docs/specs/units/candidate/c_unit_{unit}.md`) or appendix files without running `unit_check:{unit}` for re-validation

## Allowed Writes

- `src/**` — implementation files
- `tests/**` — test files
- Configs, fixtures, prompts, and other implementation-side files required by the unit's acceptance items
- `docs/specs/repository_mapping.md` — implementation path registration and `registration_state=landed` update
- `docs/specs/units/candidate/**` — candidate spec and appendix files; only for fixes discovered during implementation; must run `unit_check:{unit}` after any modification

## How to End

| Result | Meaning | Next Step |
|--------|---------|-----------|
| `impl_complete` | Implementation finished, candidate truth satisfied | Run `command close` with outcome `impl_complete` → `Next Command` becomes `unit_verify`. Then run `unit_verify:{unit}`. |
| `spec_issue` | Spec issues discovered during implementation | Run `command close` with outcome `spec_issue` → `Next Command` becomes `unit_check`. Fix the candidate spec, then run `unit_check:{unit}` for re-validation. |
| `checkpoint` | Progress saved, continue later | Run `command close` with outcome `checkpoint` → `Next Command` stays as `unit_check, unit_impl, unit_verify`. Resume later with `unit_impl:{unit}`. |

Tooling invocation: `specflowctl command close --command unit_impl --object-type unit --object <unit> --outcome <outcome>`

==ATOM_BEGIN:close_fallback==
### Manual Command Close (when `specflowctl` is unavailable)

When `specflowctl command close` is unavailable (tooling not installed, broken, or
inaccessible), read `framework/lifecycle/command_close_fallback.md` for the complete
manual command close procedure.
==ATOM_END:close_fallback==
