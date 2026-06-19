# Unit Implementation (Trigger)

`unit_impl:{unit}` is a trigger command that enters the implementation phase.
It provides implementation context and boundaries without changing lifecycle state.

## Input

Before triggering, confirm from `docs/specs/_status.md` that `Next Command` is `unit_verify`.

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
5. After `unit_check` passes (`Next Command` is still `unit_verify`), resume with
   `unit_impl:{unit}`

## On-Demand References

Agent may read these as needed during implementation:

- `docs/specs/units/candidate/c_unit_{unit}.md` — acceptance items
- `docs/specs/repository_mapping.md` — file path ownership
- Bound shared rules — constraints
- Dependent unit implementations — interface alignment

## Not Allowed

- Modify lifecycle state (`_status.md`)
- Implement behavior beyond the unit's acceptance items
- Modify candidate spec (`docs/specs/units/candidate/c_unit_{unit}.md`) or appendix files without running `unit_check:{unit}` for re-validation

## Allowed Writes

- `src/**` — implementation files
- `tests/**` — test files
- Configs, fixtures, prompts, and other implementation-side files required by the unit's acceptance items
- `docs/specs/repository_mapping.md` — implementation path registration and `registration_state=landed` update
- `docs/specs/units/candidate/**` — candidate spec and appendix files; only for fixes discovered during implementation; must run `unit_check:{unit}` after any modification

## How to End

`unit_impl:{unit}` is a trigger command that does not produce process evidence or change lifecycle state, so there is no `command close` or outcome table. The terminal condition is that implementation is complete and the candidate truth has been satisfied.

**Terminal outcome:** When implementation is complete, run `unit_verify:{unit}`.
