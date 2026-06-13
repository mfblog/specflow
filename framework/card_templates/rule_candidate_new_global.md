<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: rule/{rule_id}

## STATUS
- Scope: global | Layer: candidate
- Version: 0.1.0 (new rule)

## GUIDANCE
This is a brand-new candidate global rule. No stable baseline exists.
On promote, all current-layer units automatically inherit this constraint.

## WRITES
- docs/specs/rules/candidate/c_{rule_id}.md
- docs/specs/repository_mapping.md (if registering a new rule object)

## READS
- framework/spec_writing_guide.md
- framework/candidate_intent.md
- framework/lifecycle/overview.md
- framework/governance/rules/rule_new.md
- docs/specs/repository_mapping.md
- docs/specs/rules/stable/s_g_rule_repository_baseline.md

## IMPACTS
Not yet in effect. On promote, automatically affects all current-layer units.

## BLOCKED
- Directly editing unit specs
- Promoting without rule_sync

## CLOSE
Rule governance flow closes through its own procedures (rule_new, rule_sync, etc.), not unit command close.

## Related Flows
- Create rule → rule_new
- Coordinate consumers → rule_sync
Re-run: specflowctl context card --object-type rule --object {rule_id}
