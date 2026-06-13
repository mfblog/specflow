---
id: unit_d
layer: candidate
version: 0.2.0
candidate_intent: change
source_basis: new_design
rule_refs:
  - s_b_rule_policy@1.0.0
  - c_b_rule_draft@0.1.0
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
