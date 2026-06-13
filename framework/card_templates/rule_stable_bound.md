<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: rule/{rule_id}

## STATUS
- Scope: bound | Layer: stable
- Rule version: {version}
- Consumers: {n} unit(s)

## GUIDANCE
This is a stable bound shared rule. All units that list this rule in their rule_refs must comply with its constraints.

To modify this rule:
  1. Create a candidate version: follow the rule governance process
  2. Changing a bound rule may require consuming units to re-check/verify
  3. After changes, use rule_sync to coordinate affected units

## WRITES
- Rule governance operates through framework/governance/rules/ (do not edit directly)

## READS
- docs/specs/rules/stable/s_{rule_id}.md
- docs/specs/repository_mapping.md
- framework/governance/rule_system.md
- framework/governance/rules/rule_new.md (if modification needed)

## IMPACTS (consuming units)
| Unit | Layer | Next Command |
|------|-------|-------------|
{affected_units_rows}

These units depend on this rule. Changing the rule may require them to re-check or re-verify.

## BLOCKED
- Editing rule files outside the governance process
- Editing consuming units' specs (their lifecycle owns them)
- Editing _status.md

## CLOSE
Rule governance flow closes through its own procedures (rule_new, rule_extract, rule_bind, etc.), not unit command close.

## Related Flows
- Modify rule → rule_new / rule_extract / rule_bind
- Coordinate consumers → rule_sync
- Exit when stuck → rule_escape
Re-run: specflowctl context card --object-type rule --object {rule_id}
