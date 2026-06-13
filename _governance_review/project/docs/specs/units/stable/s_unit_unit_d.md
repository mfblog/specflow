---
id: unit_d
layer: stable
version: 1.0.0
rule_refs:   - s_b_rule_policy@1.0.0
unit_refs: none
evidence_appendix_ref: none
---

# unit_d

## Testability / Acceptance Criteria

acceptance_item_set:
  - id: unit_d.core
    target: Core behavior of unit_d.
    verification_surface: integration
    implementation_surface: src/unit_d
    verification_method: Integration test
    pass_condition: unit_d works as expected
    not_runnable_yet: no
