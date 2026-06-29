# Rule System

Rules are reusable constraints that apply across units.

This file is the governance entry for rule work. Exact rule-governance entries route to the flow files under `framework/governance/rules/`.

## Rule Scopes

1. Stable global rules (`g_rule_`) apply to every current-layer unit. Candidate global rules are not enforced against current-layer units until promoted to stable.
2. Bound rules (`b_rule_`) apply only to units that reference them through `rule_refs`.

Consumer lists are derived from unit frontmatter. Rule files must not store consumer lists.

## Governance Flows

1. `rule_new` -> `framework/governance/rules/rule_new.md`
2. `rule_extract` -> `framework/governance/rules/rule_extract.md`
3. `rule_bind` -> `framework/governance/rules/rule_bind.md`
4. `rule_topology` -> `framework/governance/rules/rule_topology.md`
5. `rule_sync` -> `framework/governance/rules/rule_sync.md`
6. `rule_escape` -> `framework/governance/rules/rule_escape.md`

Rule-governance flows may delegate to `framework/governance/impact_sync.md` (through `rule_sync`).

## Routing

For an exact rule-governance entry, read this file and the matching flow file above.

For natural-language rule requests, select the first applicable flow from the table below. Read the matching flow file. If repository truth is insufficient to pick or finish a rule flow, route to `rule_escape`.

| Natural-language intent | Flow file |
|---|---|
| Create a new rule, define a new shared constraint, add a reusable prohibition | `framework/governance/rules/rule_new.md` |
| Extract a rule from existing unit truth, formalize a locally-repeated constraint | `framework/governance/rules/rule_extract.md` |
| Change rule bindings, add or remove a unit's rule reference, rebind consumers | `framework/governance/rules/rule_bind.md` |
| Rename, merge, split, replace, or retire a rule | `framework/governance/rules/rule_topology.md` |
| Release a rule version, sync rule changes to affected units, validate downstream impact | `framework/governance/rules/rule_sync.md` |
| Unclear which rule flow to use, multi-step rule changes, rule-change recovery | `framework/governance/rules/rule_escape.md` |
