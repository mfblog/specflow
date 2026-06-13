<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: rule/{rule_id}

## STATUS
- Scope: bound | Layer: candidate
- Stable baseline: s_{rule_id}@{version}
- Candidate version: {candidate_version}

## GUIDANCE
This rule has an active candidate round. The candidate version proposes a change to the stable rule.
After the candidate is promoted or the change is abandoned, consumer units must be coordinated.

## WRITES
- docs/specs/rules/candidate/c_{rule_id}.md

## READS
- docs/specs/rules/stable/s_{rule_id}.md (baseline)
- docs/specs/rules/candidate/c_{rule_id}.md
- All consuming unit specs (verify compatibility)
- framework/governance/rules/rule_new.md

## IMPACTS (consuming units)
| Unit | Layer | Next Command |
|------|-------|-------------|
{affected_units_rows}

## BLOCKED
- Directly editing consuming units' specs (their lifecycle owns them)
- Promoting without rule_sync
- Directly editing _status.md

## CLOSE
Rule governance flow closes through its own procedures (rule_new, rule_extract, etc.), not unit command close.

## Next Steps
After the rule change is finalized → run rule_sync to coordinate consumers
Re-run: specflowctl context card --object-type rule --object {rule_id}
