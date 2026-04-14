# Candidate Verify Results

This directory stores candidate-implementation verification result files.

Rules:

1. Each module in the candidate upgrade chain normally has one `_verify_result/{module}.md`.
2. These files are not formal Specs and are not behavior sources of truth.
3. Each file carries the latest `cand_verify` result for the current candidate.
4. `Verify Result Snapshot` must use fixed fields:
   - `gate`
   - `decision`
   - `allow_next`
   - `next_command`
   - `blocking_summary`
   - `coverage_summary`
   - `spec_layer_ref`
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `verification_scope_ref`
   - `system_constraints_stable_file_ref`
   - `system_constraints_stable_version_ref`
   - `system_constraints_stable_fingerprint`
   - `shared_appendix_snapshot`
5. `gate` is always `cand_verify`.
6. `next_command` may only be `cand_promote`, `cand_verify`, `cand_impl`, or `cand_check`.
7. `cand_verify` creates the file if it does not exist and later overwrites it instead of appending review noise.
8. If candidate truth changes, or implementation changes after verification, the file becomes outdated.
9. It also becomes outdated when formal global baseline bindings or Shared Appendix bindings drift.
10. `spec_fork` must delete the previous round's `_verify_result/{module}.md`.
11. `cand_promote` must delete the corresponding `_verify_result/{module}.md`.
12. Consumers must validate bindings, not just existence.
13. This README is also constrained by `specflow/framework/docs/agent_guidelines/candidate_handoff_contract.md`.
14. The fixed snapshot fields above do not expand in this round.
15. When commands explain why verification must continue, fall back to implementation, or fall back to closure, they must use the standardized `fallback_reason_code` taxonomy first and then add natural-language explanation.
