# Candidate Plan Drafts

This directory stores non-consumable planning working artifacts for `cand_plan`.

Rules:

1. Each module may have one `draft/{module}.md`.
2. Draft plans are not valid inputs to `cand_impl` or `cand_verify`.
3. A draft plan may record:
   - known implementation findings
   - open implementation unknowns
   - blocking summary
   - resume signal
   - research notes
4. Minimum fields are:
   - `object_ref`
   - `truth_file_ref`
   - `truth_version_ref`
   - `truth_fingerprint`
   - `fallback_reason_code`
   - `blocking_summary`
   - `resume_signal`
   - `known_findings`
   - `open_unknowns`
   - `research_notes`
5. Draft plans do not carry gate-bearing semantics and do not replace the active plan.
6. `truth-fallback`, `spec_fork`, `cand_promote`, recovery, and `Candidate=no` must delete the corresponding draft file.
7. After a round successfully writes `active/{module}.md`, the corresponding draft file should normally be deleted.
