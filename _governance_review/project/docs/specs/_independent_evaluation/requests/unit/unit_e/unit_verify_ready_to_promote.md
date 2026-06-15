# Independent Evaluation Request

## Request

- `object_type`: `unit`
- `object_ref`: `unit_e`
- `reviewer_pack`: `unit_verify_ready_to_promote`
- `review_title`: `Unit Verify Ready-To-Promote Review`
- `process_kind`: `verify`
- `process_file`: `docs/specs/_verify_result/unit/unit_e.md`
- `request_file`: `docs/specs/_independent_evaluation/requests/unit/unit_e/unit_verify_ready_to_promote.md`
- `created_at`: `2026-06-15T09:50:47Z`

## Reviewer Role

You are the independent reviewer for this request. Do not modify repository files.
Review Subject lists all files you may need to examine (paths only). Evaluation Questions are the authoritative review criteria.

Use Evaluation Questions as the authoritative review criteria.

## Review Goal

Decide whether candidate verification is ready for promotion.


## Allowed Inputs

- user goal or exact `unit_verify:unit_e` target.
- candidate unit truth, valid check result, and active plan.
- verify result under review.
- evidence refs needed to inspect acceptance coverage, retirement evidence, and package-aware delta verification.

## Forbidden Inputs

- unrecorded executor claims that tests passed.
- implementation changes not represented by plan or evidence refs.
- promotion judgment not grounded in verify evidence.

## Review Subject (artifact under review)

- docs/specs/_check_result/unit/unit_e.md
- docs/specs/_verify_result/unit/unit_e.md
- docs/specs/rules/stable/s_g_rule_default.md
- docs/specs/rules/stable/s_g_rule_security.md
- docs/specs/units/candidate/appendix/c_unit_unit_e_evidence.md
- docs/specs/units/candidate/c_unit_unit_e.md
- docs/specs/units/stable/s_unit_unit_e.md
- framework/core/independent_evaluation.md
- framework/lifecycle/unit_verify.md
- framework/spec_writing_guide.md

## Review Evidence Refs

- go test ./...

## Evaluation Questions

- Does the verify result cover every executable acceptance item?
- Does each executable acceptance item have inspectable evidence refs that prove the candidate behavior through the declared verification surface?
- Does the verify result reject weak evidence as sufficient by itself, including generic test success, absent old strings, present new files, or present new fields?
- For primary protocol, default page, primary presentation, API, or artifact-generation changes, does the evidence inspect real generated artifacts, API return values, DOM/screenshots, rendered text, CLI output, or tests proving the mainline path uses the candidate protocol?
- For acceptance items that declare `affects`, does the `scope_verification` record confirm that all affected files, appendices, rules, and dependencies were verified?
- Does the implementation contain dead code, unnecessary abstractions, or duplicated logic that could be simplified?
- Is the implementation concise and proportional to the acceptance item scope? (For replacement scenes: is the new code volume proportionate to the replaced code volume?)
- Does the implementation introduce over-engineering (layers, interfaces, or patterns not justified by current requirements)?
- Are the old code paths declared in `affects.files` fully removed (not left as dead wrappers, compatibility stubs, or commented-out code)?
- Is there any remaining module, test, or configuration that references the deleted paths? The reviewer records findings for each dimension. An outcome of `pass` requires all functional and scope questions to pass. Code quality and retirement questions may produce `quality_concern` findings that are recorded in the review findings but do not automatically block promotion; the executor may address them in the current round or defer to a follow-up round.

## Reviewer Output

Return exactly one reviewer result:

```text
pass | blocked | needs_human_decision
```

If the result is `blocked` or `needs_human_decision`, include concrete blocking findings. If the result is `pass`, include no findings.

## Executor Receipt After Pass

Only the executor writes this receipt into process evidence after receiving reviewer result `pass`.

```yaml
evaluation_mode: independent
reviewer_result: pass
reviewer_context: minimal_context
review_input_refs: unit_verify_ready_to_promote;docs/specs/_independent_evaluation/requests/unit/unit_e/unit_verify_ready_to_promote.md;docs/specs/_check_result/unit/unit_e.md;docs/specs/_verify_result/unit/unit_e.md;docs/specs/rules/stable/s_g_rule_default.md;docs/specs/rules/stable/s_g_rule_security.md;docs/specs/units/candidate/appendix/c_unit_unit_e_evidence.md;docs/specs/units/candidate/c_unit_unit_e.md;docs/specs/units/stable/s_unit_unit_e.md;framework/core/independent_evaluation.md;framework/lifecycle/unit_verify.md;framework/spec_writing_guide.md;go test ./...
review_findings: none
human_decision_refs: none
```

## Trigger Instruction

Read and execute this independent evaluation request: docs/specs/_independent_evaluation/requests/unit/unit_e/unit_verify_ready_to_promote.md
