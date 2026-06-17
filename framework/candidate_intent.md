# Candidate Intent

`candidate_intent` explains why a unit candidate layer exists. Used only by the unit candidate layer.

## Allowed Values

| Intent | Purpose |
|--------|---------|
| `change` | The candidate layer intends to change the stable layer's behavior, dependencies, rule binding, acceptance, or implementation expectations |
| `repair` | The candidate layer preserves the stable layer's expected behavior, fixing missing/outdated/malformed/insufficient truth |

`unit_fork` must write `candidate_intent`. When `evidence_appendix_ref` is not `none`, the evidence appendix file referenced by that field must be created before or during the `unit_fork` writeback (see `framework/spec_writing_guide.md` Section 7 for appendix format and ownership). The agent executing `unit_fork` MUST observe the actual current implementation — inspecting interfaces, data formats, behaviors, and side effects — and record those observations in the evidence appendix file. The evidence appendix records observed implementation behavior as traceability evidence, not as durable behavior truth. It must not be generated from spec intent or second-hand description alone.

`unit_new` does not write `candidate_intent` — it creates the first candidate truth with no stable-layer parent to relate to. Write `source_basis` per the Onboarding Source Decision in `framework/operations/entry_routing.md`; `candidate_intent` is not required for `unit_new`. When `source_basis` is `existing_implementation` or `mixed`, the candidate Spec MUST include `evidence_appendix_ref` pointing to a valid evidence appendix file recording observed implementation behavior (same rule as change candidate). The agent executing `unit_new` MUST inspect the actual implementation files and record observed behavior in the evidence appendix; it must not fabricate or infer the appendix content from spec documentation.

## Change Candidate

### Field Requirements

```yaml
candidate_intent: change
source_basis: see Onboarding Source Decision in framework/operations/entry_routing.md
evidence_appendix_ref: none | <candidate appendix path>
```

- `repair_basis` is not allowed
- If the behavior depends on current implementation/tests/runtime behavior, `existing_implementation` or `mixed` must be used with an evidence appendix
- If replacing existing behavior without using it as selected truth, use `replacement` + `evidence_appendix_ref=none`
- For `replacement`, at least one acceptance item with `verification_type: inspectable` must have `evidence_requirements` including `old_code_deleted` and `no_remaining_refs`

### Command Behavior

- **unit_fork**: Derive from current stable-layer main Spec, write `candidate_intent=change`, record behavior differences from stable layer
- **unit_check**: Verify behavior differences are explicit, boundaries are clear, acceptance items are directly verifiable, and source fields are consistent. Verify that `evidence_appendix_ref` references exist and their content is semantically consistent with the declared `source_basis`.
- **unit_verify**: Verify the implementation satisfies the candidate truth
- **unit_promote**: `candidate_intent` metadata is not written to the stable layer after promotion

## Repair Candidate

### Field Requirements

```yaml
candidate_intent: repair
repair_basis: s_unit_{unit}@<version>
source_basis: new_design
evidence_appendix_ref: none
```

- `repair_basis` must name the stable-layer version to restore
- Must include a `Repair Scope` section specifying: acceptance item IDs being restored, observed deviations, expected implementation-side changes, and verification evidence required
- `Repair Scope` must not redefine behavior truth
- On promotion, `Repair Scope` and `candidate_intent` are not written to the stable layer

### Command Behavior

- **unit_fork**: Derive from stable-layer main Spec, version uses the next PATCH
- **unit_check**: Must verify the repair candidate does not change stable behavior truth. Violations (modifying protocol/fields/ownership/state machine semantics) must require `fix_required` and recommend switching to `change`. Verify that `Repair Scope` fields match the repair basis and that `evidence_appendix_ref=none` is consistent with `source_basis=new_design`.
- **unit_verify**: Must prove the implementation satisfies the repair basis and acceptance items. New behavior or relaxed pass conditions must not be treated as repair success
- **unit_promote**: Stable version is the PATCH version of the repair basis; candidate-specific fields are not written to the stable layer

## Not Allowed

- Chat-only behavior decisions becoming truth
- Bypassing `source_basis` and `evidence_appendix_ref` rules
- Bypassing the standard unit candidate command chain
