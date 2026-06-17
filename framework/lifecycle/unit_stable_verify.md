# Unit Stable Verify

`unit_stable_verify:{unit}` checks whether the current implementation still conforms to the stable-layer truth.

## Input

> **Reading guidance:** Unit truth, process files, and implementation files (listed first) provide the data this command evaluates. Framework and contract files provide format and rule context. Procedural instructions are inline in "What This Step Does" and "How to End" below.

- `docs/specs/_status.md`
- `docs/specs/units/stable/s_unit_{unit}.md`
- Stable-layer appendices and rule files referenced by the unit
- The unit's entry in `docs/specs/repository_mapping.md`
- Current implementation and test files
- Existing `_stable_verify_result/unit/{unit}.md` (if an update is needed)
- `framework/process_snapshot_contract.md` (for stable verify result file format and validation rules)

## Pre-Execution Self-Check (MANDATORY)

Before executing this step, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm the target unit's `Active Layer` is `stable`. (The `Next Command` may be any value except `unit_promote` for stable-layer units, per `framework/core/status.md` "Valid Next Commands" — `unit_stable_verify` allows semantics. `unit_promote` cannot occur for `Active Layer=stable` units because `unit_promote` is a candidate-layer command — see status.md legal combination rules.)
        If `docs/specs/_stable_verify_result/unit/{unit}.md` already exists WITHOUT
        receipt fields (`evaluation_mode` absent), check the review result file at
        `docs/specs/_independent_evaluation/results/unit/{unit}/unit_stable_verify_advancing.md`:
        * If the result file exists and `reviewer_result` is `pass`: update
          `_stable_verify_result` with the receipt fields (`evaluation_mode`,
          `reviewer_result`, `reviewer_context`, `review_input_refs`,
          `review_findings`, `human_decision_refs`) and run `command close` —
          do NOT re-run verification or generate a new evaluation request.
        * If the result file exists and `reviewer_result` is `blocked` or
          `needs_human_decision`: route to the appropriate handling defined in the
          advancing outcome's post-review instructions (see How to End).
        * If the result file does **not** exist: check your conversation history or
          ask the user whether the prior independent review (for
          `unit_stable_verify_advancing`) has returned. If the review has NOT yet
          returned: STOP and report the pending review — do NOT re-run verification
          or generate a new evaluation request.
==ATOM_BEGIN:shared_guards==
2. [ ] If `_status.md` is empty (no units registered): STOP, report that no units are registered, and suggest `unit_new` as the first step.
==ATOM_END:shared_guards==
3. [ ] Read `docs/specs/units/stable/s_unit_{unit}.md` — confirm stable-layer truth exists and is the current accepted truth.
4. [ ] Confirm current implementation and test files are accessible.
5. [ ] If any check fails: STOP, report what is missing, and do not proceed.

If all checks pass: proceed to "What This Step Does" below.

> **Policy guard:** Per `framework/operations/entry_routing.md` Implementation Classification point 6, small implementation-only changes on stable should not automatically enter `unit_stable_verify` without explicit user intent or evidence misalignment. Re-read the routing policy before entering this step for routine implementation-side edits.

## What This Step Does

Check current implementation consistency with stable-layer truth.
Output should be `aligned` (consistent), `controlled_repair_required` (repair needed), `controlled_change_required` (change needed), `small_repair_required` (small non-behavior repair), `truth_rejudge_required` (stable-layer truth re-evaluation needed), `truth_text_change_required` (stable-layer truth text must change), or `evidence_incomplete` (insufficient evidence). See the How to End table for advancing vs non-advancing outcomes.

## Note

- This step requires independent review, not self-approval. Use the `unit_stable_verify_advancing` reviewer pack from `framework/core/independent_evaluation.md`. When reporting a review stop, document: (1) the generated evaluation request file path, (2) the trigger instruction from `specflowctl evaluation request`, (3) that the reviewer must not modify repository files, (4) that execution resumes after the reviewer returns `pass`, `blocked`, or `needs_human_decision`.
- Stable verification does not create candidate truth itself. If a change is needed, the result triggers a subsequent `unit_fork`
- For `aligned`, every acceptance item must have `pass` evidence

## Not Allowed

- Modify stable-layer or candidate-layer truth
- Modify lifecycle state
- Modify rule truth
- Modify implementation files — **Exception**: when the current outcome is `small_repair_required`, non-behavior implementation repairs (comments, formatting, documentation strings, naming) are permitted before re-verify

## Allowed Writes

- `docs/specs/_stable_verify_result/unit/{unit}.md` — stable verify result (including independent evaluation receipt after review)

## How to End

| Result | Meaning | Next Step | Command Close Writeback |
|--------|---------|-----------|------------------------|
| `aligned` | Implementation matches stable truth | 1. Write `_stable_verify_result` at `docs/specs/_stable_verify_result/unit/{unit}.md` without independent evaluation receipt.<br>2. Generate evaluation request: `./specflow/tooling/bin/specflowctl-<os>-<arch> evaluation request ... --pack unit_stable_verify_advancing`.<br>3. Request independent review using the `unit_stable_verify_advancing` reviewer pack from `framework/core/independent_evaluation.md`.<br>4. After review returns, handle per result:<br>  - `pass`: update `_stable_verify_result` with receipt fields. Proceed to step 5.<br>  - `blocked`: show the reviewer's findings to the user. Delete `_stable_verify_result`. Run `command close` with outcome `blocked` (keeps `Next Command=unit_stable_verify`). After the issue is fixed, re-run `unit_stable_verify:{unit}`.<br>  - `needs_human_decision`: show the reviewer's findings to the user and ask whether to proceed:<br>    * User decides to proceed: update `_stable_verify_result` with receipt fields, setting `human_decision_refs` to the user's decision reference. Proceed to step 5.<br>    * User decides to fix: delete `_stable_verify_result`. Run `command close` with outcome `blocked` (keeps `Next Command=unit_stable_verify`). After the issue is fixed, re-run `unit_stable_verify:{unit}`.<br>  - review not yet returned: STOP and report the pending review.<br>5. Run `./specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_stable_verify --object-type unit --object {unit} --outcome aligned --apply` (sets `Next Command=unit_fork`).<br>6. Proceed to `unit_fork:{unit}`. | command close sets `Next Command=unit_fork`. |
| `controlled_repair_required` | Repair needed | 1–4. Same two-phase write-and-review flow as `aligned`.<br>5. Run `./specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_stable_verify --object-type unit --object {unit} --outcome controlled_repair_required --apply` (sets `Next Command=unit_fork`).<br>6. Proceed to `unit_fork:{unit}` with repair intent. | command close sets `Next Command=unit_fork`. |
| `controlled_change_required` | Change needed | 1–4. Same two-phase write-and-review flow as `aligned`.<br>5. Run `./specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_stable_verify --object-type unit --object {unit} --outcome controlled_change_required --apply` (sets `Next Command=unit_fork`).<br>6. Proceed to `unit_fork:{unit}` with change intent. | command close sets `Next Command=unit_fork`. |
| `small_repair_required` | Small repair needed, no behavior truth change | Perform the non-behavior repair on implementation files. Then `unit_stable_verify` (re-verify) | command close sets `Next Command=unit_stable_verify` for continued verification. |
| `truth_rejudge_required` | Stable-layer truth needs re-evaluation | If truth text is valid: supplement interpretation evidence and re-verify | command close sets `Next Command=unit_stable_verify` for continued verification. |
| `truth_text_change_required` | Stable-layer truth text must change | 1–4. Same two-phase write-and-review flow as `aligned`.<br>5. Run `./specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_stable_verify --object-type unit --object {unit} --outcome truth_text_change_required --apply` (sets `Next Command=unit_fork`).<br>6. Proceed to `unit_fork:{unit}` with repair intent. | command close sets `Next Command=unit_fork`. |
| `evidence_incomplete` | Evidence insufficient | Supplement evidence and re-verify. **Also used for `stable_verify_invalid` recovery reason code** — when recovery sets this reason, close with `evidence_incomplete`. | command close sets `Next Command=unit_stable_verify` for continued verification. |
| `blocked` | Missing critical input or unresolvable condition | Ask the user to resolve the missing input. command close keeps `Next Command=unit_stable_verify`. | command close keeps `Next Command=unit_stable_verify`. |

Close through `command close`.
Tooling invocation: `specflowctl command close --command unit_stable_verify --object-type unit --object <unit> --outcome <outcome> [--notes <notes>] [--apply]`
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
