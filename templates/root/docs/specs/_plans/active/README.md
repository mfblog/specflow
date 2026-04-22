# Candidate Active Plans

This directory stores the consumable implementation plans for candidate progression.

Rules:

1. Each module normally has one `active/{module}.md`.
2. These files are not formal Specs and are not behavior sources of truth.
3. Each active plan file carries:
   - `Implementation Tasks`
   - current-round progress, blockers, and verification focus
4. Each active plan file must also record:
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `module_appendix_snapshot`
   - `system_constraints_stable_file_ref`
   - `system_constraints_stable_version_ref`
   - `system_constraints_stable_fingerprint`
   - `shared_contract_snapshot`
5. `active/{module}.md` does not carry the gate decision of `cand_check` or the promotion decision of `cand_verify`.
6. `cand_plan` writes or updates `active/{module}.md` only when the round is `plan-ready`.
7. `cand_impl` may keep writing progress, blockers, and verification notes, but must not rewrite binding fields.
8. `cand_verify` still requires a current valid `active/{module}.md`.
9. If candidate truth changes, the active plan becomes outdated.
10. It also becomes outdated when formal global baseline bindings or Shared Contract bindings drift.
11. Once outdated, the flow must go back to `cand_check` first and then re-run `cand_plan`.
12. `spec_fork` must delete the previous round's `active/{module}.md`.
13. `cand_promote` must delete the corresponding `active/{module}.md`.
14. When `Candidate=no`, `active/{module}.md` should not remain.
15. This README is also constrained by `specflow/framework/docs/agent_guidelines/candidate_handoff_contract.md`.
16. Snapshot fields in this file must use the fixed definitions from `specflow/framework/docs/agent_guidelines/process_snapshot_contract.md`.
