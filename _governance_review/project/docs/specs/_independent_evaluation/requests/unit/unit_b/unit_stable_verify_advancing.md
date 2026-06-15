# Independent Evaluation Request

## Request

- `object_type`: `unit`
- `object_ref`: `unit_b`
- `reviewer_pack`: `unit_stable_verify_advancing`
- `review_title`: `Stable Verify Advancing Review`
- `process_kind`: `stable_verify`
- `process_file`: `docs/specs/_stable_verify_result/unit/unit_b.md`
- `request_file`: `docs/specs/_independent_evaluation/requests/unit/unit_b/unit_stable_verify_advancing.md`
- `created_at`: `2026-06-15T08:24:14Z`

## Reviewer Role

You are the independent reviewer for this request. Do not modify repository files.
Review Subject lists all files you may need to examine (paths only). Evaluation Questions are the authoritative review criteria.

Use Evaluation Questions as the authoritative review criteria.

## Review Goal

Decide whether the stable verify result supports the stored advancing decision.


## Allowed Inputs

- exact `unit_stable_verify:unit_b` target.
- stable unit truth, stable appendices owned by the unit, rules, and repository mapping snapshot.
- stable verify result under review.
- implementation surface refs and evidence refs needed to inspect stable alignment.
- decision criteria from `framework/lifecycle/unit_stable_verify.md`.

## Forbidden Inputs

- candidate truth unless the stable verify result explicitly cites it as historical context.
- proposed repairs or changes not captured in the stable verify result.
- executor preference for aligned, controlled repair, or controlled change outcomes.

## Review Subject (artifact under review)

- docs/specs/_stable_verify_result/unit/unit_b.md
- docs/specs/repository_mapping.md
- docs/specs/rules/stable/s_g_rule_default.md
- docs/specs/units/stable/s_unit_unit_b.md
- framework/core/independent_evaluation.md
- framework/lifecycle/unit_stable_verify.md

## Review Evidence Refs

- go test ./...
- src/unit_b

## Evaluation Questions

- Does current implementation align with stable truth, or does the stored decision correctly identify the controlled next step?
- Does the evidence matrix cover every current stable acceptance item?
- Are implementation surface refs and evidence refs sufficient for the stored decision?
- If the decision is `truth_text_change_required`, does the evidence prove that the stable truth text must change and cannot be resolved through re-interpretation of the existing text?

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
review_input_refs: unit_stable_verify_advancing;docs/specs/_independent_evaluation/requests/unit/unit_b/unit_stable_verify_advancing.md;docs/specs/_stable_verify_result/unit/unit_b.md;docs/specs/repository_mapping.md;docs/specs/rules/stable/s_g_rule_default.md;docs/specs/units/stable/s_unit_unit_b.md;framework/core/independent_evaluation.md;framework/lifecycle/unit_stable_verify.md;go test ./...;src/unit_b
review_findings: none
human_decision_refs: none
```

## Trigger Instruction

Read and execute this independent evaluation request: docs/specs/_independent_evaluation/requests/unit/unit_b/unit_stable_verify_advancing.md
