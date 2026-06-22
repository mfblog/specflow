# Unit Check

`unit_check:{unit}` is a pre-verify quality gate that checks whether candidate truth is sufficiently clear and complete. It does not itself advance lifecycle state — however, a `pass` outcome's `command close` sets `Next Command` to `unit_check, unit_impl, unit_verify` (the implementation-phase multi-value set). This is a side effect of the close operation, not a progression behavior of `unit_check` as a check step.

## Input

> **Reading guidance:** Must Read files are the truth and process data this command evaluates. May Reference files hold the format and policy contracts referenced by the checks — read them when a specific check question needs the exact rule text. Procedural instructions are inline in "What This Step Does" and "How to End" below.

### Must Read

- `docs/specs/_status.md`
- `docs/specs/units/candidate/c_unit_{unit}.md`
- Current unit's candidate-layer appendix files
- Stable-layer truth and rule files referenced by the current unit
- `docs/specs/repository_mapping.md` (for path ownership and constraint derivation during command close)
- `docs/specs/_check_result/unit/{unit}.md` — present in re-validation flow; absent in standard flow (first check)
- `docs/specs/_check_work/unit/{unit}.md` — optional command-local checklist file for progress tracking (see `framework/process_snapshot_contract.md` Section 11)

### May Reference

- `framework/process_snapshot_contract.md` (check result file format and validation rules)
- `framework/spec_writing_guide.md` (unit Spec format and source field format)
- `framework/candidate_intent.md` (candidate_intent field rules, source_basis consistency, repair candidate requirements)
- `framework/core/independent_evaluation.md` (independent evaluation procedures, review result recording, tooling-unavailable fallback)
- `framework/core/status.md` (lifecycle state validation and Constraints Derivation during command close)

## Pre-Execution Self-Check (MANDATORY)

Before executing this step, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm one of:
        - Standard entry: the target unit's `Next Command` is `unit_check` and
          `docs/specs/_check_result/unit/{unit}.md` does not exist or already has
          complete receipt fields (`evaluation_mode` present).
        - Re-validation: the target unit's `Next Command` contains `unit_check` (meaning the unit is in the implementation phase), and the candidate spec was modified after the last `unit_check` pass (re-validation during the implementation phase).
           To detect spec modification, compare the current spec fingerprint against `truth_fingerprint`
           stored in `docs/specs/_check_result/unit/{unit}.md` (see `framework/process_snapshot_contract.md` Section 7 for fingerprint calculation).
           If `docs/specs/_check_result/unit/{unit}.md` is absent during re-validation (e.g., the file was not yet created for this implementation session), proceed with re-validation — the absence itself indicates the spec has not been validated for the current implementation session.
        - Checkpoint resume: the target unit's `Next Command` is `unit_check` AND
          `docs/specs/_check_result/unit/{unit}.md` already exists WITHOUT receipt
          fields (`evaluation_mode` absent). Check the review result file at
          `docs/specs/_independent_evaluation/results/unit/{unit}/unit_check_pass.md`:
          * If the result file exists and `reviewer_result` is `pass`: update
            `_check_result` with the receipt fields (`evaluation_mode`,
            `reviewer_result`, `reviewer_context`, `review_input_refs`,
            `review_findings`, `human_decision_refs`) and run `command close` —
            do NOT re-run checks or generate a new evaluation request.
          * If the result file exists and `reviewer_result` is `blocked` or
            `needs_human_decision`: route to the appropriate handling defined in
            the `pass` outcome's review stop instructions.
          * If the result file does NOT exist: check your conversation history or
            ask the user whether the prior independent review has returned. If the
            review has NOT yet returned: STOP and report the pending review — do
            NOT re-run checks or generate a new evaluation request.
==ATOM_BEGIN:shared_guards==
2. [ ] If `_status.md` is empty (no units registered): STOP, report that no units are registered, and suggest `unit_new` as the first step.
==ATOM_END:shared_guards==
3. [ ] Read `docs/specs/units/candidate/c_unit_{unit}.md` — confirm it exists and has acceptance items.
4. [ ] Confirm candidate-layer appendix files exist (if required).
5. [ ] If any check fails: STOP, report what is missing, and do not proceed.

If all checks pass: proceed to "What This Step Does" below.

## What This Step Does

Create or update the `_check_work` checklist at `docs/specs/_check_work/unit/{unit}.md` for progress tracking. Then check the following questions. All must pass for a `pass` result:

1. Is the unit's goal and responsibility scope clear?
2. Are dependencies, rule bindings, and ownership boundaries explicit?
3. Are the main flow, data, protocol, states, and error paths complete enough for verification?
4. Can verification proceed without guessing behavior, boundaries, or acceptance?
5. Do all acceptance items have the correct format (`verification_type`, `evidence_requirements`, `affects`)?
6. If `candidate_intent: change` + `source_basis: replacement`, is there at least one `verification_type: inspectable` item with `evidence_requirements` including `old_code_deleted` and `no_remaining_refs`?
7. Are all `affects` scopes correct (must not be empty without reason)?
8. If `evidence_appendix_ref` is not `none`, does the referenced appendix file exist with valid frontmatter and correct `unit`/`layer` values?
9. If `source_basis` is `existing_implementation` or `mixed`: (a) `evidence_appendix_ref` must be present and not `none`; (b) the referenced evidence appendix file must exist with valid frontmatter and correct unit/layer values. (This closes the `source_basis`–to–evidence-appendix consistency check: a claim of existing-implementation source status must be backed by a real, valid evidence appendix.)
10. If `candidate_intent: repair`, does the Spec include a `Repair Scope` section with the required sub-fields (acceptance item IDs being restored, observed deviations, expected implementation-side changes, verification evidence required)?
11. For `unit_fork`-derived candidates: do all Markdown document references within the candidate main Spec body and candidate appendix files use candidate-layer paths (`c_unit_*`)? Are there any remaining stable-layer paths (`s_unit_*`) that were not rewritten during the fork?
12. If `candidate_intent: repair`, is `repair_basis` present and correctly formatted (`s_unit_{unit}@<version>`), and are `source_basis=new_design` and `evidence_appendix_ref=none`?
13. If `candidate_intent: change`, is `repair_basis` absent (not allowed for change candidates)?
14. Does the repair candidate preserve stable behavior truth? If it modifies protocol, fields, ownership, or state machine semantics, it must require `fix_required` and recommend switching to `change`.
15. If `evidence_appendix_ref` is present and not `none`, is its content semantically consistent with the declared `source_basis`?
16. If the target unit's `Notes` in `docs/specs/_status.md` contains `appendix_exc:` entries and the unit is in a post-recovery state (returned to `unit_check` from a higher phase), verify each excluded stable appendix is still irrelevant to the current candidate round. Remove any exclusion for stable appendices that have become relevant.

## Not Allowed

- Modify implementation files
- Modify candidate or stable-layer truth (except fix_required and blocked outcomes may modify candidate-layer truth for repair per the How to End table)
- Modify lifecycle state
- Modify rule truth

## Allowed Writes

- `docs/specs/_check_work/unit/{unit}.md` — progress checklist
- `docs/specs/_check_result/unit/{unit}.md` — check result (including independent evaluation receipt after review)

## How to End

| Result | Meaning | Next Step |
|--------|---------|-----------|
| `pass` | Spec meets conditions | **Step 1 — Write check result.** Write `_check_result` at `docs/specs/_check_result/unit/{unit}.md` with these fields (process evidence format per `framework/process_snapshot_contract.md` Section 2):<br>- `acceptance_behavior_fingerprint`: sha256 of normalized acceptance items<br>- `rule_snapshot`: each bound rule as `ref: fingerprint`<br>- `truth_fingerprint`: sha256 of normalized candidate spec<br>- `unit_snapshot`: each dependent unit as `ref: fingerprint`<br>- `unit_appendix_snapshot`: each candidate appendix as `ref: fingerprint`<br><br>Do NOT include the independent evaluation receipt in this first write.<br><br>**Step 2 — Generate evaluation request.** Run `./specflow/tooling/bin/specflowctl-<os>-<arch> evaluation request --repo-root <root> --object-type unit --object {unit} --pack unit_check_pass --process check`. If specflowctl is unavailable, create `docs/specs/_independent_evaluation/requests/unit/{unit}/unit_check_pass.md` with: `reviewer_pack: unit_check_pass`, `review_standard_refs` listing `framework/core/independent_evaluation.md` and `framework/lifecycle/unit_check.md`, `review_file_refs` listing `docs/specs/units/candidate/c_unit_{unit}.md`, `review_evidence_refs` listing `docs/specs/_check_result/unit/{unit}.md`, and `durable_input_refs` combining these refs.<br><br>**Step 3 — Request independent review.** Use the `unit_check_pass` reviewer pack. The reviewer reads only the request file and returns `pass`, `blocked`, or `needs_human_decision`.<br><br>**Step 4 — Record review result.** First write the result file at `docs/specs/_independent_evaluation/results/unit/{unit}/unit_check_pass.md` with: `reviewer_result`, `reviewer_context`, `review_input_refs`, `recorded_at` (UTC ISO 8601). Then update `_check_result` with receipt fields:<br>- `evaluation_mode: independent`<br>- `reviewer_result: pass`<br>- `reviewer_context: minimal_context`<br>- `review_input_refs: unit_check_pass;docs/specs/_independent_evaluation/requests/unit/{unit}/unit_check_pass.md;...`<br>- `review_findings: none`<br>- `human_decision_refs: none` (or user decision reference)<br><br>**Step 5 — Command close.** Run `./specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_check --object-type unit --object {unit} --outcome pass --apply`. This sets `Next Command` to `unit_check, unit_impl, unit_verify` (the implementation-phase multi-value set). The tool automatically derives write constraints from `docs/specs/repository_mapping.md` Object Registry. If specflowctl is unavailable, use the Tooling-Unavailable Fallback in `framework/lifecycle/overview.md`. After close, the unit enters the `[implementation]` phase; run `unit_impl:{unit}` to begin implementing.
  - **If review returned `blocked`:** Write the review result file, show findings to user, delete `_check_result`, run `command close` with outcome `blocked` (sets `Next Command=unit_check`, clears `Notes`). After spec fix, re-run `unit_check:{unit}`.
  - **If review returned `needs_human_decision`:** Show findings, ask user. If proceed: write result file, update `_check_result` with receipt and `human_decision_refs`, run `command close` (sets `Next Command=unit_check, unit_impl, unit_verify`). If fix: delete `_check_result`, `command close` with outcome `blocked`.
  - **If review not yet returned:** STOP, report pending review. |
| `checkpoint` | Progress saved, review not yet complete | Resume `unit_check:{unit}`. command close sets `Next Command=unit_check`. |
| `fix_required` | Spec needs repair | Fix the candidate Spec and re-check. command close sets `Next Command=unit_check`. Notes: `constraints:` prefix removed; `appendix_exc:` entries preserved for re-evaluation per check item 16. Delete any existing `_check_work/unit/{unit}.md` — the next `unit_check` round creates a fresh checklist. |
| `blocked` | Missing critical input | Ask the user to resolve the missing input. command close sets `Next Command=unit_check`. Notes: `constraints:` prefix removed; `appendix_exc:` entries preserved for re-evaluation per check item 16. Delete any existing `_check_work/unit/{unit}.md` — the next `unit_check` round creates a fresh checklist. |

Tooling invocation: `specflowctl command close --command unit_check --object-type unit --object <unit> --outcome <outcome> [--notes <notes>]`
==ATOM_BEGIN:close_fallback==
### Manual Command Close (when `specflowctl` is unavailable)

When `specflowctl command close` is unavailable (tooling not installed, broken, or
inaccessible), read `framework/lifecycle/command_close_fallback.md` for the complete
manual command close procedure.
==ATOM_END:close_fallback==
