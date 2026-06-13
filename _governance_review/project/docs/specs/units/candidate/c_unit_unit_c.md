---
id: unit_c
layer: candidate
version: 0.1.0
candidate_intent: change
source_basis: new_design
rule_refs:
  - s_b_rule_policy@1.0.0
unit_refs:
  - s_unit_unit_a@1.0.0
evidence_appendix_ref: none
---

# Unit C

## Testability / Acceptance Criteria

acceptance_item_set:
  - id: unit_c.core
    target: Core behavior of unit_c.
    verification_surface: integration
    implementation_surface: src/unit_c
    verification_method: Integration test
    pass_condition: unit_c works as expected
    not_runnable_yet: no
