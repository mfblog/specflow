<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: unit/{unit} (limited mode)

## STATUS
- Stage: {phase} | Next: {next_command}
- Layer: {layer}
- Adoption mode: {mode}

## GUIDANCE
This repository is using **{mode}** adoption mode. Your scope of operations is limited by the table below.

| Mode | Allowed | Prohibited |
|------|---------|------------|
| reader-only | Read existing status, truth, and evidence | Lifecycle commands, evidence writes, state changes, implementation edits, promote, verify |
| implementation-only | Modify code/tests within existing formal truth scope | Change behavior/boundaries/acceptance/rules/ownership truth |
| single-unit-trial | Lifecycle steps for the target unit only | Promote, stable verification, rule governance, governance review |
| unit-check-only | Run unit_check | Plan, implement, verify, promote, stable verification, governance review |

**If the user's request exceeds the mode's scope:**
Stop. Explain the mode boundary to the user. Propose the smallest legal next step.

**If the request is within scope, follow the state-specific guidance below for the current state ({phase}):**

{state_guidance}

## WRITES
- Restricted by mode, see table above and state-specific guidance

## READS
- docs/specs/_status.md
- docs/specs/repository_mapping.md
- framework/core/adoption_modes.md

## BLOCKED
- Operations outside the allowed scope of the selected adoption mode
- Lifecycle commands not listed in the mode's Allowed column
- Writing to files not owned by this unit or not permitted by the mode

## CLOSE
Adoption-limited mode does not use command close. Follow the state-specific CLOSE section above for terminal actions.

## Next Steps
If the request exceeds the mode's scope, stop at the smallest legal next step and explain the mode boundary to the user.
Re-run: specflowctl context card