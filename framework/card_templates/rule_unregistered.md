<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: rule/{rule_id}

## STATUS
- This rule file does not exist

## GUIDANCE
There is no corresponding rule file for `{rule_id}`. To create this rule:

1. Determine the rule scope:
   - `g_rule_` prefix → global rule (affects all units)
   - `b_rule_` prefix → bound rule (only affects explicitly referencing units)

2. Follow the rule governance process to create:
   ```text
   Rule governance entry → framework/governance/rule_system.md
   Creating a new rule → framework/governance/rules/rule_new.md
   ```

## WRITES
- To be determined after creation

## READS
- framework/governance/rule_system.md
- framework/governance/rules/rule_new.md
- docs/specs/repository_mapping.md

## BLOCKED
- Creating rule files directly (must go through rule governance process)
- Modifying any unit's spec or status

## CLOSE
Rule is unregistered, cannot execute command close. First run rule_create or the relevant rule governance flow to create this rule.

## Next Steps
Run the rule_new flow to create this rule, or check whether the rule ID is correct
Re-run: specflowctl context card --object-type rule --object {rule_id}
