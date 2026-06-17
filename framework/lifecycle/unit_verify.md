# Unit Verify

`unit_verify:{unit}` verifies whether the implementation satisfies each acceptance item in the candidate truth.

## Input

> **Reading guidance:** Unit truth, process files, and implementation files (listed first) provide the data this command evaluates. Framework and contract files provide format and rule context. Procedural instructions are inline in "What This Step Does" and "How to End" below.

- `docs/specs/_status.md`
- `docs/specs/units/candidate/c_unit_{unit}.md`
- Current unit's candidate-layer appendix files
- Stable-layer truth and rule files referenced by the current unit
- The unit's implementation and test files
- `docs/specs/_check_result/unit/{unit}.md` — present in standard flow (unit_check → unit_impl → unit_verify); may be absent or stale in re-validation flow (unit_check re-validation path; see `unit_check.md` Pre-Execution Self-Check for the full precondition: `Next Command=unit_verify` with `Notes=pending_impl` after spec modification)
- `framework/process_snapshot_contract.md` (for verify result file format and validation rules)
- `framework/spec_writing_guide.md` (for unit Spec format and appendix format)
- `framework/core/independent_evaluation.md` (for independent evaluation procedures, review result recording, and tooling-unavailable fallback)
- `framework/core/status.md` (for lifecycle state validation and Constraints Derivation during command close)
- `docs/specs/repository_mapping.md` (for implementation file ownership discovery)

## Pre-Execution Self-Check (MANDATORY)

Before executing this step, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm the target unit's `Next Command` is `unit_verify`.
        If `docs/specs/_verify_result/unit/{unit}.md` already exists WITHOUT
        receipt fields (`evaluation_mode` absent), check the review result file at
        `docs/specs/_independent_evaluation/results/unit/{unit}/unit_verify_ready_to_promote.md`:
        * If the result file exists and `reviewer_result` is `pass`: update
          `_verify_result` with the receipt fields (`evaluation_mode`,
          `reviewer_result`, `reviewer_context`, `review_input_refs`,
          `review_findings`, `human_decision_refs`) and run `command close` —
          do NOT re-run verification or generate a new evaluation request.
        * If the result file exists and `reviewer_result` is `blocked` or
          `needs_human_decision`: route to the appropriate handling defined in
          the `ready_to_promote` outcome's review stop instructions.
        * If the result file does NOT exist: check your conversation history or
          ask the user whether the prior independent review has returned. If the
          review has NOT yet returned: STOP and report the pending review — do
          NOT re-run verification or generate a new evaluation request.
==ATOM_BEGIN:shared_guards==
2. [ ] If `_status.md` is empty (no units registered): STOP, report that no units are registered, and suggest `unit_new` as the first step.
==ATOM_END:shared_guards==
3. [ ] Read `docs/specs/units/candidate/c_unit_{unit}.md` — confirm candidate truth and acceptance items are available.
4. [ ] Confirm the unit's implementation and test files exist and are accessible.
5. [ ] Compare the current candidate spec fingerprint against `truth_fingerprint` stored in `docs/specs/_check_result/unit/{unit}.md` (see `framework/process_snapshot_contract.md` Section 6 for fingerprint calculation). If the fingerprint changed after the last `unit_check` pass, STOP: the spec was modified without re-validation. Close with outcome `spec_issue` to set `Next Command=unit_check` (see the `spec_issue` outcome row in How to End below), then route through `unit_check:{unit}` for re-validation. If `_check_result/unit/{unit}.md` does not exist (absent in re-validation flow), treat as spec-modified — the candidate truth has never passed check in this session; close with outcome `spec_issue` and route through `unit_check:{unit}` first.
6. [ ] If any check fails: STOP, report what is missing, and do not proceed.

If all checks pass: proceed to "What This Step Does" below.

## What This Step Does

1. **Functional verification**: Verify each acceptance item is satisfied with inspectable evidence
2. **Scope verification**: Verify the `affects` declarations (files, appendices, rules, dependencies) are correctly implemented
3. **Retirement verification** (replacement scenario): Verify old code paths are fully removed with no remaining references
4. **Code quality check**: No dead code, no over-engineering, reasonable change volume

## Verification Evidence Requirements

- Every executable acceptance item must have a corresponding entry in `acceptance_item_evidence_matrix`
- Changes to primary protocols, APIs, UI, or generated artifacts must be verified against real output (screenshots, API return values, CLI output, etc.), not just "tests pass"
- Verification must not automatically delete code or infer business compatibility safety

## Note

- This step requires independent review — **self-approval is not allowed**. An independent reviewer must give `pass` for `ready_to_promote`. Use the `unit_verify_ready_to_promote` reviewer pack from `framework/core/independent_evaluation.md`. When reporting a review stop, document: (1) the generated evaluation request file path, (2) the trigger instruction from `specflowctl evaluation request`, (3) that the reviewer must not modify repository files, (4) that execution resumes after the reviewer returns `pass`, `blocked`, or `needs_human_decision`.
- If implementation issues are found during verification, they may be fixed and re-verified
- If the candidate Spec itself is problematic, return to `unit_check` to fix the Spec

## Not Allowed

- Modify candidate or stable-layer truth (except as required by the `truth_fallback` outcome recovery procedure in the How to End table below)
- Modify lifecycle state
- Modify rule truth

## Allowed Writes

- `docs/specs/_verify_result/unit/{unit}.md` — verify result (including independent evaluation receipt after review)
- `docs/specs/_check_work/unit/{unit}.md` — allowed only for non-standard failure cleanup (see How to End non-standard failure table)
- `docs/specs/_check_result/unit/{unit}.md` — allowed only for non-standard failure cleanup (see How to End non-standard failure table)

## How to End

| Result | Meaning | Next Step | Command Close Writeback |
|--------|---------|-----------|------------------------|
| `ready_to_promote` | Verification passed, review passed | **Step 1 — Write verify result.** Write `_verify_result` at `docs/specs/_verify_result/unit/{unit}.md` with `acceptance_item_evidence_matrix` (each item: `id`, `status`, `evidence_refs`; optional `scope_verification` for items with `affects`). Do NOT include the independent evaluation receipt. Format per `framework/process_snapshot_contract.md` Section 2.<br><br>**Step 2 — Generate evaluation request.** Run `./specflow/tooling/bin/specflowctl-<os>-<arch> evaluation request --repo-root <root> --object-type unit --object {unit} --pack unit_verify_ready_to_promote --process verify`. If unavailable: create `docs/specs/_independent_evaluation/requests/unit/{unit}/unit_verify_ready_to_promote.md` with fields: `reviewer_pack`, `review_standard_refs`, `review_file_refs` (candidate spec, verify result), `review_evidence_refs`, `durable_input_refs`.<br><br>**Step 3 — Request independent review.** Use the `unit_verify_ready_to_promote` reviewer pack. Reviewer returns `pass`, `blocked`, or `needs_human_decision`.<br><br>**Step 4 — Record review result.** Write result file at `docs/specs/_independent_evaluation/results/unit/{unit}/unit_verify_ready_to_promote.md` with `reviewer_result`, `reviewer_context`, `review_input_refs`, `recorded_at`. Then update `_verify_result` with receipt: `evaluation_mode: independent`, `reviewer_result`, `reviewer_context: minimal_context`, `review_input_refs`, `review_findings: none`, `human_decision_refs: none`.<br><br>**Step 5 — Command close.** `./specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_verify --object-type unit --object {unit} --outcome ready_to_promote --apply` (sets `Next Command=unit_promote`). If unavailable, follow the Manual Command Close procedure below. Then proceed to `unit_promote:{unit}` (requires explicit user decision).
  - **blocked:** Write result file, show findings, delete `_verify_result`, `command close` with outcome `blocked` (keeps `Next Command=unit_verify`). Re-run after fix.
  - **needs_human_decision:** Show findings, ask user. Proceed: write result + receipt with `human_decision_refs`, `command close`. Fix: delete `_verify_result`, `command close` with `blocked`.
  - **Not returned:** STOP, report pending review. | command close sets `Next Command=unit_promote`. `unit_promote:{unit}` requires explicit user decision. |
| `truth_fallback` | Candidate truth has drifted from stable baseline | Correct candidate truth per recovery.md truth_drift Recovery Procedure step 1 (restore alignment with stable Spec), then clean evidence per recovery.md truth_drift Recovery Procedure and return to `unit_check:{unit}` | command close sets `Next Command=unit_check` and clears `Notes`. |
| `spec_issue` | Candidate Spec needs repair | Return to `unit_check:{unit}`, fix the Spec, and re-check | command close sets `Next Command=unit_check` and clears `Notes`. |
| `evidence_incomplete` | Evidence insufficient for verification | Supplement evidence (`acceptance_item_evidence_matrix`) and rerun `unit_verify:{unit}` | command close keeps `Next Command=unit_verify`. |
| `human_verify` | Human decision required before promotion | Ask the user and rerun `unit_verify:{unit}` after input | command close keeps `Next Command=unit_verify`. |
| `impl_issue` | Implementation needs repair | Fix code and rerun `unit_verify:{unit}` | command close keeps `Next Command=unit_verify`. |
| `blocked` | Unresolvable condition — use when evidence_incomplete, human_verify, or impl_issue persist without a path to resolution | Report the blocking condition and stop. command close keeps `Next Command=unit_verify`. This outcome is an escalation path when non-advancing outcomes (evidence_incomplete, human_verify, impl_issue) cannot be resolved after multiple retries — it records the persistent block and keeps the unit at `unit_verify` without looping. | command close keeps `Next Command=unit_verify`. |

> **Note:** The `ready_to_promote` flow writes the verify result without the independent evaluation receipt (step 1), then updates it with receipt fields after review returns `pass` (step 4). See `framework/core/independent_evaluation.md` "Handoff Requests" for the two-phase validation rules and `framework/process_snapshot_contract.md` Section 11 item 15 for pre-receipt validation.

For non-standard failures (process validation failure, tooling error, corrupted state), apply fallback cleanup per the failure layer:

| Layer | Failure Types | Cleanup | Next Command |
|-------|--------------|---------|-------------|
| truth_layer | truth_drift, binding_drift, baseline_drift, rule_drift, truth_incomplete | Delete check_work, check_result, verify_result | `unit_check` |
| gate_layer | gate_missing, spec_issue | Delete check_work, check_result | `unit_check` |
| evidence_layer (candidate) | evidence_incomplete | Delete verify_result | `unit_verify` |
| evidence_layer (stable) | stable_verify_invalid | See `framework/lifecycle/recovery.md` evidence_layer row | Per recovery.md |

See `framework/lifecycle/recovery.md` for the full procedure and `framework/process_snapshot_contract.md` Section 4 for layer classification rules.

Tooling invocation: `specflowctl command close --command unit_verify --object-type unit --object <unit> --outcome <outcome> [--notes <notes>] [--apply]`
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
