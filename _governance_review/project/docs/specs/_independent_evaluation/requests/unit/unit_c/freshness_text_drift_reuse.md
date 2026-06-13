# Independent Evaluation Request

## Request

- `object_type`: `unit`
- `object_ref`: `unit_c`
- `reviewer_pack`: `freshness_text_drift_reuse`
- `review_title`: `Freshness Text-Drift Reuse Review`
- `process_kind`: `check`
- `process_file`: `docs/specs/_check_result/unit/unit_c.md`
- `request_file`: `docs/specs/_independent_evaluation/requests/unit/unit_c/freshness_text_drift_reuse.md`
- `created_at`: `2026-06-13T18:03:31Z`

## Reviewer Role

You are the independent reviewer for this request. Do not modify repository files.
Review Subject lists all files you may need to examine (paths only). Evaluation Questions are the authoritative review criteria.

Use Evaluation Questions as the authoritative review criteria.

## Review Goal

Decide whether text-only drift can reuse existing process evidence.


## Allowed Inputs

- current truth or spec file.
- prior process evidence being reused.
- deterministic freshness classification showing `text_drift`.
- acceptance behavior fingerprint comparison and current fingerprint reported by tooling.

## Forbidden Inputs

- reuse claims when deterministic validation reports `semantic_drift`, `acceptance_drift`, `dependency_drift`, `schema_drift`, or `unknown_drift`.
- executor assertions that the text change is harmless without current file refs.
- unrelated changes outside the file and process evidence under review.

## Review Subject (artifact under review)

- docs/specs/_check_result/unit/unit_c.md
- docs/specs/rules/stable/s_b_rule_policy.md
- docs/specs/rules/stable/s_g_rule_default.md
- docs/specs/units/candidate/appendix/c_unit_unit_c_design.md
- docs/specs/units/candidate/c_unit_unit_c.md
- docs/specs/units/stable/s_unit_unit_a.md
- framework/core/freshness.md
- framework/core/independent_evaluation.md

## Review Evidence Refs

- none

## Evaluation Questions

- Is the change only wording, formatting, or clarification that preserves the acceptance behavior already reviewed?
- Does the prior evidence still answer the same gate question?
- Is recreating evidence unnecessary for semantic safety?

## Mechanical Validation

- `freshness_impact`: `text_drift`
- `evidence_reuse`: `pending_review`
- `freshness_current_fingerprint`: `0cd8cf41a9a34797ea21a5225e99bbf76add05f849214a996c0e1cab4bf2382b`

## Reviewer Output

Return exactly one reviewer result:

```text
pass | blocked | needs_human_decision
```

If the result is `blocked` or `needs_human_decision`, include concrete blocking findings. If the result is `pass`, include no findings.

## Executor Receipt After Pass

Only the executor writes this receipt into process evidence after receiving reviewer result `pass`.

```yaml
freshness_impact: text_drift
evidence_reuse: accepted
freshness_current_fingerprint: 0cd8cf41a9a34797ea21a5225e99bbf76add05f849214a996c0e1cab4bf2382b
freshness_review_mode: independent
freshness_reviewer_result: pass
freshness_reviewer_context: minimal_context
freshness_review_input_refs: freshness_text_drift_reuse;docs/specs/_independent_evaluation/requests/unit/unit_c/freshness_text_drift_reuse.md;docs/specs/_check_result/unit/unit_c.md;docs/specs/rules/stable/s_b_rule_policy.md;docs/specs/rules/stable/s_g_rule_default.md;docs/specs/units/candidate/appendix/c_unit_unit_c_design.md;docs/specs/units/candidate/c_unit_unit_c.md;docs/specs/units/stable/s_unit_unit_a.md;framework/core/freshness.md;framework/core/independent_evaluation.md
freshness_review_findings: none
```

## Trigger Instruction

Read and execute this independent evaluation request: docs/specs/_independent_evaluation/requests/unit/unit_c/freshness_text_drift_reuse.md
