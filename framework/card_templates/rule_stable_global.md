<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: rule/{rule_id}

## STATUS
- Scope: global | Layer: stable
- Rule version: {version}
- Impact: automatically inherited by all current-layer units
- Consumer count: {n} unit(s)

## GUIDANCE
This is a stable global rule. All current-layer units automatically inherit this constraint.
Modifying a global rule requires special caution — it affects all units in the repository.

To modify this rule:
  1. Create a candidate version: follow the rule governance process
  2. Changing a global rule may affect all current-layer units
  3. After changes, must use rule_sync to coordinate all units

## WRITES
- Rule governance operates through framework/governance/rules/ (do not edit directly)

## READS
- docs/specs/rules/stable/s_{rule_id}.md
- docs/specs/repository_mapping.md
- framework/governance/rule_system.md
- framework/governance/rules/rule_new.md (if modification needed)
- docs/specs/rules/stable/s_g_rule_repository_baseline.md

## IMPACTS
All current-layer units (units with Stable=yes or Candidate=yes)

| Unit | Layer | Next Command |
|------|-------|-------------|
{affected_units_rows}

## BLOCKED
- Editing rule files outside the governance process
- Editing any unit's spec (their lifecycle owns them)
- Editing _status.md

## CLOSE
Rule governance flow closes through its own procedures (rule_new, rule_extract, rule_bind, etc.), not unit command close.

## Related Flows
- Modify rule → rule_new / rule_extract / rule_bind
- Coordinate consumers → rule_sync
- Exit when stuck → rule_escape
Re-run: specflowctl context card --object-type rule --object {rule_id}
