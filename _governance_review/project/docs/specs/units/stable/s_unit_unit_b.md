---
id: unit_b
layer: stable
version: 1.0.0
rule_refs: none
unit_refs: none
evidence_appendix_ref: none
---

# unit_b

## Testability / Acceptance Criteria

acceptance_item_set:
  - id: unit_b.core
    target: Core behavior of unit_b.
    verification_surface: integration
    implementation_surface: src/unit_b
    verification_method: Integration test
    pass_condition: unit_b works as expected
    not_runnable_yet: no
