<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: rule/{rule_id}

## STATUS
- Scope: global | Layer: candidate
- Stable baseline: s_{rule_id}@{version}
- Candidate version: {candidate_version}

## GUIDANCE
This global rule has an active candidate round. After the change, all current-layer units will be affected.
Must use rule_sync to coordinate all units.

## WRITES
- docs/specs/rules/candidate/c_{rule_id}.md

## READS
- docs/specs/rules/stable/s_{rule_id}.md (baseline)
- docs/specs/rules/candidate/c_{rule_id}.md
- All current-layer unit specs (verify compatibility)
- framework/governance/rules/rule_new.md
- docs/specs/rules/stable/s_g_rule_repository_baseline.md

## IMPACTS
All current-layer units:

| Unit | Layer | Next Command |
|------|-------|-------------|
{affected_units_rows}

## BLOCKED
- Promoting without rule_sync
- Directly editing unit specs
- Directly editing _status.md

## CLOSE
Rule governance flow closes through its own procedures.

## Next Steps
After the rule change is finalized → run rule_sync to coordinate all units
Re-run: specflowctl context card --object-type rule --object {rule_id}
