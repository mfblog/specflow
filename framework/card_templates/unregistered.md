<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: unit/{unit}

## STATUS
- This unit is not yet registered in `_status.md`

## GUIDANCE
There is no row for `{unit}` in `docs/specs/_status.md`. Registration is required before any lifecycle work.

> **Pre-check:** Verify that the unit name `{unit}` is correct and matches the user's intent. If unsure, ask the user to confirm.

**Execution:**
1. Read `framework/operations/entry_routing.md` → Onboarding Source Decision section to determine the correct registration command.
2. Choose the registration method:

   - **unit_init** — use when the capability already exists and is acceptable. Records directly as stable truth.
     Write: `docs/specs/units/stable/s_unit_{unit}.md`
     Follow: `framework/lifecycle/unit_init_new_fork.md` for the init procedure.
     
   - **unit_new** — use when creating a brand-new capability from scratch.
     Write: `docs/specs/units/candidate/c_unit_{unit}.md`
     Follow: `framework/lifecycle/unit_init_new_fork.md` for the new procedure.

3. Write the spec file with correct frontmatter (see `framework/spec_writing_guide.md`).
4. Register the unit via command close.

**Close:** `specflowctl command close --command unit_init --object-type unit --object {unit} --outcome registered`
or: `specflowctl command close --command unit_new --object-type unit --object {unit} --outcome registered`

> If the current request does not involve spec changes: registering a new unit necessarily involves writing a spec file. Register first before any implementation work.

## WRITES
- To be determined after registration

## READS
- docs/specs/_status.md
- docs/specs/repository_mapping.md
- framework/operations/entry_routing.md
- framework/lifecycle/unit_init_new_fork.md

## BLOCKED
- Modifying implementation files (not yet registered)
- Advancing lifecycle state
- Modifying any spec files

## CLOSE
Unit is unregistered, cannot execute command close. Register first by running unit_init or unit_new.

## Next Steps
Confirm the registration method with the user, then run unit_init:{unit} or unit_new:{unit}
Re-run: specflowctl context card
