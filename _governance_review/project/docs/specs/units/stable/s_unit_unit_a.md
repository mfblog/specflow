---
id: unit_a
layer: stable
version: 1.0.0
rule_refs:   - s_b_rule_policy@1.0.0
unit_refs: none
evidence_appendix_ref: none
---

# unit_a

## Testability / Acceptance Criteria

acceptance_item_set:
  - id: unit_a.core
    target: Core behavior of unit_a.
    verification_surface: integration
    implementation_surface: src/unit_a
    verification_method: Integration test
    pass_condition: unit_a works as expected
    not_runnable_yet: no
