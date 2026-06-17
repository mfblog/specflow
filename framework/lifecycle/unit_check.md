# Unit Check

`unit_check:{unit}` is a pre-verify quality gate that checks whether candidate truth is sufficiently clear and complete. It does not itself advance lifecycle state — however, a `pass` outcome's `command close` sets `Next Command` to `unit_verify` with `Notes=pending_impl` (supplied by the caller alongside the close). This is a side effect of the close operation, not a progression behavior of `unit_check` as a check step.

## Input

> **Reading guidance:** Unit and rule truth files (listed first) provide the data this command evaluates. Framework and contract files provide format and rule context. Procedural instructions are inline in "What This Step Does" and "How to End" below.

- `docs/specs/_status.md`
- `docs/specs/units/candidate/c_unit_{unit}.md`
- Current unit's candidate-layer appendix files
- Stable-layer truth and rule files referenced by the current unit
- `framework/process_snapshot_contract.md` (for check result file format and validation rules)
- `framework/spec_writing_guide.md` (for unit Spec format and source field format)
- `framework/candidate_intent.md` (for candidate_intent field rules, source_basis consistency, and repair candidate requirements)
- `framework/core/independent_evaluation.md` (for independent evaluation procedures, review result recording, and tooling-unavailable fallback)
- `framework/core/status.md` (for lifecycle state validation and Constraints Derivation during command close)
- `docs/specs/repository_mapping.md` (for path ownership and constraint derivation during command close)
- `docs/specs/_check_result/unit/{unit}.md` — present in re-validation flow; absent in standard flow (first check)
- `docs/specs/_check_work/unit/{unit}.md` — optional command-local checklist file for progress tracking (see `framework/process_snapshot_contract.md` Section 10)

## Pre-Execution Self-Check (MANDATORY)

Before executing this step, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm one of:
        - Standard entry: the target unit's `Next Command` is `unit_check` and
          `docs/specs/_check_result/unit/{unit}.md` does not exist or already has
          complete receipt fields (`evaluation_mode` present).
        - Re-validation: the target unit's `Next Command` is `unit_verify`, `Notes` contains
          `pending_impl`, and the candidate spec was modified after the last `unit_check` pass
          (re-validation during the implementation phase).
           To detect spec modification, compare the current spec fingerprint against `truth_fingerprint`
           stored in `docs/specs/_check_result/unit/{unit}.md` (see `framework/process_snapshot_contract.md` Section 6 for fingerprint calculation).
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
| `pass` | Spec meets conditions | **Step 1 — Write check result.** Write `_check_result` at `docs/specs/_check_result/unit/{unit}.md` with these fields (process evidence format per `framework/process_snapshot_contract.md` Section 2):<br>- `acceptance_behavior_fingerprint`: sha256 of normalized acceptance items<br>- `rule_snapshot`: each bound rule as `ref: fingerprint`<br>- `truth_fingerprint`: sha256 of normalized candidate spec<br>- `unit_snapshot`: each dependent unit as `ref: fingerprint`<br>- `unit_appendix_snapshot`: each candidate appendix as `ref: fingerprint`<br>Do NOT include the independent evaluation receipt in this first write.<br><br>**Step 2 — Generate evaluation request.** Run `specflowctl evaluation request --repo-root <root> --object-type unit --object {unit} --pack unit_check_pass --process check`. If specflowctl is unavailable, create `docs/specs/_independent_evaluation/requests/unit/{unit}/unit_check_pass.md` with: `reviewer_pack: unit_check_pass`, `review_standard_refs` listing `framework/core/independent_evaluation.md` and `framework/lifecycle/unit_check.md`, `review_file_refs` listing `docs/specs/units/candidate/c_unit_{unit}.md`, `review_evidence_refs` listing `docs/specs/_check_result/unit/{unit}.md`, and `durable_input_refs` combining these refs.<br><br>**Step 3 — Request independent review.** Use the `unit_check_pass` reviewer pack. The reviewer reads only the request file and returns `pass`, `blocked`, or `needs_human_decision`.<br><br>**Step 4 — Record review result.** First write the result file at `docs/specs/_independent_evaluation/results/unit/{unit}/unit_check_pass.md` with: `reviewer_result`, `reviewer_context`, `review_input_refs`, `recorded_at` (UTC ISO 8601). Then update `_check_result` with receipt fields:<br>- `evaluation_mode: independent`<br>- `reviewer_result: pass`<br>- `reviewer_context: minimal_context`<br>- `review_input_refs: unit_check_pass;docs/specs/_independent_evaluation/requests/unit/{unit}/unit_check_pass.md;...`<br>- `review_findings: none`<br>- `human_decision_refs: none` (or user decision reference)<br><br>**Step 5 — Command close.** Run `specflowctl command close --command unit_check --object-type unit --object {unit} --outcome pass --notes "constraints:phase=pending_impl deny=docs/specs/** deny=framework/** allow=<implementation_paths_from_repository_mapping> allow=docs/specs/repository_mapping.md" --apply`. This sets `Next Command=unit_verify` with `Notes=pending_impl`. Derive constraints from `docs/specs/repository_mapping.md` Object Registry: read the unit's `implementation_paths`, build `allow=` globs for each path, include `allow=docs/specs/repository_mapping.md`. See `framework/core/status.md` Constraints Derivation for the complete constraint rules. If specflowctl is unavailable, use the Tooling-Unavailable Fallback in `framework/lifecycle/overview.md`. After close, run `unit_impl:{unit}`.
  - **If review returned `blocked`:** Write the review result file, show findings to user, delete `_check_result`, run `command close` with outcome `blocked` (sets `Next Command=unit_check`, clears `Notes`). After spec fix, re-run `unit_check:{unit}`.
  - **If review returned `needs_human_decision`:** Show findings, ask user. If proceed: write result file, update `_check_result` with receipt and `human_decision_refs`, run `command close` (sets `Next Command=unit_verify`). If fix: delete `_check_result`, `command close` with outcome `blocked`.
  - **If review not yet returned:** STOP, report pending review. |
| `checkpoint` | Progress saved, review not yet complete | Resume `unit_check:{unit}`. command close sets `Next Command=unit_check`. |
| `fix_required` | Spec needs repair | Fix the candidate Spec and re-check. command close sets `Next Command=unit_check` and clears `Notes`. Delete any existing `_check_work/unit/{unit}.md` — the next `unit_check` round creates a fresh checklist. |
| `blocked` | Missing critical input | Ask the user to resolve the missing input. command close sets `Next Command=unit_check` and clears `Notes`. Delete any existing `_check_work/unit/{unit}.md` — the next `unit_check` round creates a fresh checklist. |

Tooling invocation: `specflowctl command close --command unit_check --object-type unit --object <unit> --outcome <outcome> [--notes <notes>]`
==ATOM_BEGIN:close_fallback==
### Manual Command Close (when `specflowctl` is unavailable)

When `specflowctl command close` is unavailable (tooling not installed, broken, or inaccessible), perform a manual close following these deterministic rules. This is the **only** exception to the rule that `command close` is the sole mechanism for advancing lifecycle state.

**Manual close is scoped to the current lifecycle command only.** It must not be used to skip lifecycle phases, jump ahead in the lifecycle sequence, or perform close operations that involve automatic file mutations that manual file editing cannot reliably reproduce.

**Pre-conditions (mandatory — all must pass):**

1. All required writes from the "How to End" outcome above are complete and correct.
2. All process evidence files are written with the correct schema (see `framework/process_snapshot_contract.md` for file format).
3. For advancing outcomes: the independent evaluation receipt is present in the process evidence, satisfying gate rule requirements from `framework/core/independent_evaluation.md` Section Gate Rules.
4. The `docs/specs/_status.md` file is readable and the target unit's `Next Command` matches the command being closed.

If any pre-condition fails: STOP, report what is missing, and do not perform the manual close.

**Procedure:**

1. From the "How to End" outcome table above, identify your outcome and its Next Step column.
2. Update `docs/specs/_status.md` for the target unit:
   - Set `Next Command` to the value specified in the outcome's Next Step.
   - Set or clear `Notes` per the outcome's Next Step description.
   - For `unit_fork` with outcome `candidate_created`: set `Active Layer` to `candidate`.
   - For `unit_promote` with outcome `promoted`: set `Active Layer` to `stable`, `Stable` to `yes`, `Candidate` to `no`.
   - For `unit_init` with outcome `stable_created`: set `Stable=yes`, `Candidate=no`, `Active Layer=stable`.
   - For `unit_new` with outcome `candidate_created`: set `Stable=no`, `Candidate=yes`, `Active Layer=candidate`.
   - For all other commands and outcomes: do **not** change `Active Layer`, `Stable`, or `Candidate`.
3. If the target unit has **no row** in `_status.md` (applies to `unit_init` and `unit_new`), add a new row with the columns `| unit | {unit} | ... |` and fill values from the mapping above.
4. Perform the cleanup described in the outcome's Next Step column (delete specified evidence files, preserve others).
5. Write the updated `docs/specs/_status.md`.

**Recording the fallback:**

Add the following to the command's process evidence file (if one exists):

```yaml
command_close_fallback: manual
command_close_fallback_recorded_at: <UTC ISO 8601 timestamp>
```

This annotation documents that manual intervention occurred and is consumed by subsequent executors only as advisory context — it is not a lifecycle gate validation input.

For the reference per-outcome state transition mapping across all lifecycle commands, see `framework/lifecycle/overview.md:114-145`.
==ATOM_END:close_fallback==
