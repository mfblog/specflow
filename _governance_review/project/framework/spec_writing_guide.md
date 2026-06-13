# Spec Writing Guide

This guide defines the writing contract for specFlow delivery documents.

Files under `specflow/` are framework and delivery documents and are written in English.

Files under `docs/` are project communication documents and are written in Chinese unless a specific delivery artifact requires otherwise.

This file defines formal Spec shape and reference rules, including the semantic authoring baseline in Section 9.

Format compliance does not by itself prove handoff completeness.

## 1. Formal Spec Paths

Unit main Specs:

| kind | layer | path |
|---|---|---|
| unit | stable | `docs/specs/units/stable/s_unit_{id}.md` |
| unit | candidate | `docs/specs/units/candidate/c_unit_{id}.md` |

Rule Specs:

| kind | layer | path |
|---|---|---|
| rule | stable | `docs/specs/rules/stable/s_{g_or_b}_rule_{id}.md` |
| rule | candidate | `docs/specs/rules/candidate/c_{g_or_b}_rule_{id}.md` |

`docs/specs/scenarios/**` is not a supported formal Spec path.

## 2. Unit Frontmatter

Each unit main Spec must include these fields:

```yaml
id: {unit}
layer: stable|candidate
version: x.y.z
unit_refs: none
rule_refs: none
```

`unit_refs` may also be a YAML list of stable unit refs:

```yaml
unit_refs:
  - s_unit_agent@0.6.0
```

`rule_refs` may also be a YAML list of exact rule refs:

```yaml
rule_refs:
  - s_b_rule_example@1.0.0
```

Candidate unit Specs must also record the candidate source fields required by the active unit command, such as `candidate_intent`, `source_basis`, `repair_basis`, and `evidence_appendix_ref` when that command requires them.

## 3. Unit Dependencies

`unit_refs` means the current unit depends on the referenced stable unit's formal behavior.

It does not mean:

1. the current unit may edit the referenced unit
2. the referenced unit is part of the current unit's ownership
3. the dependency can point to a candidate unit

If the body says the unit relies on another unit's official behavior, `unit_refs` must list that stable unit ref.

## 4. Rule References

Stable global rules are default inputs for every current-layer unit and are not repeated in each unit's `rule_refs`.

`rule_refs` is the only formal consumer binding for bound shared rules.

Rules:

1. `rule_refs` must be in frontmatter
2. no formal `rule_refs` list should be duplicated in the body
3. refs must use exact layer and version
4. refs must be sorted lexically when written as a list
5. `rule_refs: none` means the unit binds no bound shared rule

Rule files must not record consumer truth through `bound_objects`.

## 5. Rule Frontmatter

Each rule Spec must include:

```yaml
rule_id: {rule}
rule_scope: global|bound
layer: stable|candidate
rule_version: x.y.z
```

`promotion_owner_unit` may be present when one unit owns the promotion decision.

## 6. Acceptance Criteria

Each current-layer unit main Spec must include a `Testability / Acceptance Criteria` section or an explicitly equivalent acceptance section.

The section must include structured acceptance items:

```yaml
acceptance_item_set:
  - id: demo.core
    description: Demo behavior is accepted.
    verification_type: testable          # testable | inspectable | reviewable
    verification_surface: internal_flow
    implementation_surface: AgentCore/internal/demo
    verification_method: Go test for demo behavior.
    pass_condition: Demo behavior passes the declared checks.
    not_runnable_yet: no
    evidence_requirements:               # minimum evidence required for this item
      - automated_test_pass
    affects:                             # scope that verify must check globally
      files:
        - internal/demo/handler.go
      appendices: []
      rules: []
      dependencies: []
```

### Acceptance Item Fields

| Field | Required | Description |
|---|---|---|
| `id` | yes | Unique identifier within the item set; used as primary key in process evidence |
| `description` | yes | Plain-language description of this acceptance item |
| `verification_type` | yes | How this item is verified: `testable` (automated test), `inspectable` (file/artifact inspection), `reviewable` (human review) |
| `verification_surface` | yes | Where verification is targeted (e.g. `internal_flow`, `api`, `ui`) |
| `implementation_surface` | yes | Implementation code surface path |
| `verification_method` | yes | How to verify (e.g. "Go test for demo behavior") |
| `pass_condition` | yes | What constitutes a pass |
| `not_runnable_yet` | yes | `yes` or `no` |
| `not_runnable_yet_reason` | recommended | Reason the item is not yet runnable; required when `not_runnable_yet: yes` |
| `target` | recommended | The behavior subject or protocol this item targets (e.g. API endpoint, module boundary, protocol name); used in `acceptance_behavior_fingerprint` calculation |
| `evidence_requirements` | recommended | List of minimum evidence types needed (e.g. `automated_test_pass`, `integration_test_pass`, `old_code_deleted`, `no_remaining_refs`) |
| `affects.files` | recommended | Implementation files that must be verified as part of this item's scope |
| `affects.appendices` | recommended | Appendix names that must be checked |
| `affects.rules` | recommended | Rule names that must be respected |
| `affects.dependencies` | recommended | Stable unit dependency names that must be maintained |

When `verification_type` is `inspectable`, the `evidence_requirements` should specify what inspection evidence is needed (e.g. `old_code_deleted`, `no_remaining_refs`).
When `verification_type` is `reviewable`, human review is the primary verification method; `evidence_requirements` may include `human_review_pass`.

The acceptance item ids are used by process evidence. Changing ids invalidates existing process files.

For `candidate_intent: change` with `source_basis: replacement`, at least one acceptance item must have `verification_type: inspectable` and `evidence_requirements` that include `old_code_deleted` and `no_remaining_refs`. This declares the retirement scope for the replaced code paths.

## 7. Appendix Files

Appendix files are support truth for one unit.

They do not replace the unit main Spec.

Appendix ownership is derived from the appendix path and appendix frontmatter, not from a main-Spec index.

Each unit appendix must:

1. use the current path shape for its layer and unit id
2. declare `unit: {unit}` in frontmatter
3. declare `layer: stable|candidate` in frontmatter

When a stable unit with appendix files is forked to candidate, every stable appendix `s_unit_{unit}_{name}.md` must have a corresponding candidate appendix `c_unit_{unit}_{name}.md`.
The candidate may have additional candidate appendices.

**Evidence appendix promotion restriction:** Evidence appendix files referenced by `evidence_appendix_ref` record observed behavior (traceability data) and are not durable behavior truth. They must not be promoted to stable truth as behavior-correctness claims during `unit_promote` (tooling removes all candidate appendix files during promotion cleanup, structurally preventing evidence appendix survival into the stable layer). The `evidence_appendix_ref` field is a candidate-only concept; stable units must not carry `evidence_appendix_ref` frontmatter. See `framework/lifecycle/unit_promote.md` for promotion write rules and `framework/candidate_intent.md` for evidence appendix semantics.

## 8. Process Snapshots

Candidate check and verify process files must include:

1. `unit_appendix_snapshot`
2. `unit_snapshot`
3. `rule_snapshot`

`unit_appendix_snapshot` records the current-layer appendix files owned by the unit.

`unit_snapshot` records resolved stable unit dependencies from `unit_refs`.

`rule_snapshot` records resolved stable global rules and resolved bound shared rule dependencies from `rule_refs`.

Snapshots prove what one command reviewed and preserve package constraints across handoff. They do not create or replace formal truth.

## 9. Authoring Baseline

A formal Spec must make the following clear for the next lifecycle step:

1. the intended user, actor, or caller
2. the unit responsibility and why the unit owns it
3. the entry point or trigger
4. the normal path from input to result
5. the boundaries crossed on that path
6. the data, state, or durable truth each step reads or writes
7. the owner of each read/write responsibility
8. the output artifact or observable result
9. the way failures or unavailable dependencies are exposed
10. the verification surface and success condition

The Spec must close implementation-affecting decisions. The downstream executor must not be forced to choose:
- which object owns a responsibility
- which entry point starts the behavior
- where state or durable truth lives
- how ordered steps connect
- how boundary failures are reported
- what the result shape means
- how acceptance proves the stated responsibility

If a decision is intentionally not made, the Spec must state that boundary and explain why.

### Appendix Handoff

Appendix files may carry detailed truth for one unit but do not weaken the handoff baseline. An appendix used as implementation truth must not contain only background, motivation, principles, or patch notes — it must state the current rule or design as directly readable truth.

## 10. Handoff Contract

### Verify to Promote

`unit_promote` consumes `docs/specs/_verify_result/unit/{unit}.md` only when it validates against current candidate unit truth. The verify result must prove every executable acceptance item through the `acceptance_item_evidence_matrix` with `status: pass` and durable `evidence_refs`.

Before stable writeback, `unit_promote` must resolve:
1. `unit_refs` (must reference stable unit versions)
2. `rule_refs`
3. global baseline rules

### Stable Verify to Fork

`unit_stable_verify` advancing outcomes consume `docs/specs/_stable_verify_result/unit/{unit}.md` only when it validates against current stable unit truth. If the stable verify result is missing, malformed, stale, or records a different decision, the unit must remain at `unit_stable_verify`.

## 11. Candidate Relation Graph

Candidate advancement order is a computed relation, not a manually maintained field.

The relation graph reads only:
1. `docs/specs/_status.md`
2. current-layer candidate unit main Specs
3. same-layer candidate appendix files owned by current-layer candidates
4. `unit_refs`, `rule_refs`, Markdown `.md` links, explicit version refs

The graph builder must not infer candidate order from prose alone.

Edge meanings:
- **stable dependency edge**: current-layer unit depends on stable unit or Rule truth
- **candidate progression edge**: candidate explicitly references another current candidate — the referencing candidate waits for the referenced one
- **reference-only edge**: evidence appendix references — traceability only, does not block

Cycle rules:
1. stable-dependency-only cycles: diagnostic only, does not block
2. any cycle containing candidate progression edges: blocks all candidates in that cycle
3. blocked candidates must not receive `unit_check pass` or be entered by `unit_advance`

## 12. Rule Scope Resolution

Rule scope is resolved from rule truth:
- `rule_scope: global` or id beginning with `g_rule_` → repository-wide rule, applies to every current-layer unit
- `rule_scope: bound` or id beginning with `b_rule_` → bound shared rule, applies only to units listing it in `rule_refs`

Rule files must not store consumer lists. `bound_objects` is not the source of rule consumers. The bound shared rule consumer graph is reconstructed from current-layer unit frontmatter `rule_refs`.

## 13. Dependency Order

```text
repository_mapping → unit → rule
stable global rule → unit and rule
rule → unit
unit → unit through stable-only unit_refs
```

1. repository mapping decides path ownership
2. unit truth owns behavior responsibility
3. rule truth owns reusable constraints
4. unit-to-unit dependency is explicit and stable-only
