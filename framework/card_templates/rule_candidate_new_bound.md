<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: rule/{rule_id}

## STATUS
- Scope: bound | Layer: candidate
- Version: 0.1.0 (new rule)

## GUIDANCE
This is a brand-new candidate bound rule. No stable baseline exists.
After creation, decide which units bind to this rule via rule_refs.

## WRITES
- docs/specs/rules/candidate/c_{rule_id}.md
- docs/specs/repository_mapping.md (if registering a new rule object)

## READS
- framework/spec_writing_guide.md
- framework/candidate_intent.md
- framework/lifecycle/overview.md
- framework/governance/rules/rule_new.md
- docs/specs/repository_mapping.md

## IMPACTS
No units bound yet. Bindings are established when other units reference this rule in their rule_refs.

## BLOCKED
- Binding units simultaneously (use rule_bind flow)
- Directly editing unit specs

## CLOSE
Rule governance flow closes through its own procedures (rule_new, rule_bind, etc.), not unit command close.

## Related Flows
- Create rule → rule_new
- Bind to unit → rule_bind
Re-run: specflowctl context card --object-type rule --object {rule_id}
