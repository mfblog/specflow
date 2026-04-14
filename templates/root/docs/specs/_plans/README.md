# Candidate Plans

This directory stores implementation plan files for candidate progression.

Rules:

1. Each module normally has one `_plans/{module}.md`.
2. These files are not formal Specs and are not behavior sources of truth.
3. Each plan file carries:
   - `Implementation Tasks`
   - current-round progress, blockers, and verification focus
4. Each plan file must also record:
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `system_constraints_stable_file_ref`
   - `system_constraints_stable_version_ref`
   - `system_constraints_stable_fingerprint`
   - `shared_appendix_snapshot`
5. `_plans/{module}.md` does not carry the gate decision of `cand_check` or the promotion decision of `cand_verify`.
6. `cand_plan` creates the file if it does not exist and writes current truth bindings.
7. `cand_impl` may keep writing progress, blockers, and verification notes, but must not rewrite binding fields.
8. `cand_verify` still requires a current valid `_plans/{module}.md`.
9. If candidate truth changes, the plan becomes outdated.
10. It also becomes outdated when formal global baseline bindings or Shared Appendix bindings drift.
11. Once outdated, the flow must go back to `cand_check` first and then re-run `cand_plan`.
12. `spec_fork` must delete the previous round's `_plans/{module}.md`.
13. `cand_promote` must delete the corresponding `_plans/{module}.md`.
14. When `Candidate=no`, `_plans/{module}.md` should not remain.
15. This README is also constrained by `specflow/framework/docs/agent_guidelines/candidate_handoff_contract.md`.
16. Snapshot fields in this file must use the fixed definitions from `specflow/framework/docs/agent_guidelines/process_snapshot_contract.md`.
17. The fixed snapshot fields above do not expand in this round.
18. When plan progress, blockers, or verification focus need to express a fallback, invalidation, or resume reason, they should use the standardized `fallback_reason_code` taxonomy first and then add natural-language explanation.
