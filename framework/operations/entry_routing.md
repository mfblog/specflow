# Entry Routing

This file is the only natural-language entry route into active SpecFlow owners.
`framework/...` refs are framework-root relative. Installed project entry files define the physical framework root; the specFlow source repository resolves them under local `framework/...`.

## Exact Commands

If the request exactly matches `unit_advance:{unit}`, read `framework/advance_policy.md`.
`unit_advance:{unit}` is not a Context Card route.

If the request exactly matches one of these forms, read `framework/lifecycle/overview.md` and the matching lifecycle Context Card:

```text
unit_init:{unit}
unit_new:{unit}
unit_fork:{unit}
unit_check:{unit}
unit_plan:{unit}
unit_impl:{unit}
unit_verify:{unit}
unit_promote:{unit}
unit_stable_verify:{unit}
```

After a lifecycle Context Card is selected, read only its Required Context. Enter On-Demand Expansions only when their trigger appears.

If the request exactly matches a rule governance entry, read the matching rule file:

1. `rule_new` -> `framework/governance/rules/rule_new.md`
2. `rule_extract` -> `framework/governance/rules/rule_extract.md`
3. `rule_bind` -> `framework/governance/rules/rule_bind.md`
4. `rule_topology` -> `framework/governance/rules/rule_topology.md`
5. `rule_sync` -> `framework/governance/rules/rule_sync.md`
6. `rule_escape` -> `framework/governance/rules/rule_escape.md`

If the request explicitly invokes one of these framework governance entries, with or without a narrowing phrase, read the matching framework file:

1. `spec_flow_review` -> `framework/governance/review.md`
2. `spec_flow_design_review` -> `framework/governance/review.md`
3. `spec_flow_migrate` -> `framework/operations/migration.md`

Explicit invocation means the request contains the literal entry name as the operation being requested.
Ordinary-language descriptions such as "review governance", "review design", or "migrate specs" do not route to those entries by implication.

If the request uses `scenario_*`, `scenario_advance:{id}`, or `object-type=scenario`, stop and report that scenario lifecycle support has been removed.

## Routing Inputs

Before choosing a natural-language route, read the smallest necessary durable truth:

1. Read `docs/specs/_status.md` when a unit is named.
2. Read `docs/specs/repository_mapping.md` when path ownership, object registration, implementation path registration, or support-surface ownership matters.
3. Read current-layer unit truth when a named unit is involved.
4. Read current-layer rule truth when rule governance is involved.
5. Read `framework/core/adoption_modes.md` when the request asks how small a specFlow adoption can be.

Do not guess ownership from directory shape alone.

## Natural Language Routes

Route to unit lifecycle when the request creates, changes, validates, plans, implements, verifies, promotes, or checks one independently governed engineering responsibility.
The responsibility may be local or end-to-end. If the user describes a complete workflow result, model it as a unit whose responsibility is that complete result.

For a natural-language unit lifecycle request, select one existing lifecycle command and its Context Card before any lifecycle write:

1. If no formal unit row exists, read `framework/onboarding_decision_policy.md`. Select `unit_init:{unit}` only when an existing accepted capability already satisfies every direct first-stable onboarding condition. Select `unit_new:{unit}` when the request creates new candidate truth or when a historical capability cannot qualify for direct first-stable onboarding. If neither selection can be proven from current truth and the request, stop and report the missing decision or prerequisite.
2. If a formal unit row exists, its recorded `Next Command` is the only legal lifecycle command. Select the matching existing command form and Context Card only when the requested work can legally be performed at that recorded next step. If the user asks for a later lifecycle result, report the recorded prerequisite step instead of skipping it.
3. A stable unit request that changes formal unit truth selects `unit_fork:{unit}` only when the recorded `Next Command` is `unit_fork`; determine `candidate_intent` through `framework/candidate_intent_policy.md` and any current valid stable-verify constraint required by the entry Context Card.
4. A candidate truth repair selects `unit_check:{unit}` only when the recorded `Next Command` is `unit_check` and the repair stays inside that card's allowed writes. Other candidate progression selects only the Context Card matching the recorded `Next Command`.
5. A request limited to implementation-side edits must first satisfy the implementation-only route below. It enters a lifecycle Context Card only when `framework/operations/implementation_change.md` requires an existing lifecycle command as the next legal step.

After selecting a lifecycle command from natural language, read `framework/lifecycle/overview.md` and that command's matching Context Card. Do not invent a command alias, enter a generic unit lifecycle without an active Context Card, or ask the user to choose an internal command name.

Route to rule governance when the request changes shared constraints, reusable prohibitions, mandatory process behavior, rule binding, or rule topology.
For non-exact rule-governance requests, read `framework/governance/rule_system.md` first. Read `framework/governance/rules/rule_escape.md` when selecting a safe first rule flow or rerouting is required.
Bound shared rule consumer discovery must use current-layer unit `rule_refs`.

Route to repository mapping when the request changes path ownership, object registration, implementation path registration, or support-surface boundaries.
Read `framework/core/repository_mapping.md`. Repository mapping does not change unit behavior or rule meaning by itself.

Route to guidance when the request asks to shape a design before formal truth is clear.
Guidance applies when the request is about one of these work shapes before candidate truth, rule truth, global rule truth, or repository mapping truth is ready to write:

1. framing a vague project or feature idea
2. cutting scope for a first useful version
3. choosing between materially different solution directions
4. reviewing a discussion-stage design before writing it into candidate truth
5. turning an approved discussion conclusion into formal truth

When guidance applies, read `framework/skills/using-specflow-guidance/SKILL.md`.
Guidance must not replace an exact command, advance lifecycle state, authorize implementation-side edits, or treat chat-only agreement as durable truth.
If a guidance conclusion affects behavior truth, boundary truth, acceptance truth, rule truth, global rule truth, or repository ownership, route that conclusion into the proper formal truth writeback path before implementation.

Route implementation-only requests to `framework/operations/implementation_change.md` only when the request asks only for implementation-side edits and does not require truth, boundary, shared rule, system rule, migration, governance, or guidance work.
Implementation permission must be proven before editing implementation files.

## Framework Governance

Requests that explicitly invoke `spec_flow_review` or `spec_flow_design_review` route to `framework/governance/review.md`.
`framework/governance/review.md` decides the default path for each review entry.
For `spec_flow_review`, the default is `scoped_review`; it delegates to `framework/spec_flow_review.md` only for exact `spec_flow_review:full`.
For `spec_flow_design_review`, there is no scoped mode; `framework/governance/review.md` delegates to `framework/spec_flow_design_review.md` for the default full-scope design-baseline review.

Requests that explicitly invoke `spec_flow_migrate` route to `framework/operations/migration.md`.

## Hard Stops

Stop and ask or reroute when:

1. the target unit is unclear
2. path ownership is unclear
3. behavior or rule truth exists only in chat and has not been written to durable truth
4. implementation permission is not proven
5. a rule or repository mapping change is required first
6. the request tries to use scenario lifecycle concepts
7. a natural-language unit request cannot be resolved to one legal existing lifecycle command and active Context Card from current durable truth

## User-Facing Output

User-facing reports must use ordinary project language before internal file names.
When applicable, report current state, next action, why that action is legal, expected result, and remaining gap.

This entry inherits `framework/operations/output_standard.md`.
