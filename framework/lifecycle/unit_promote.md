# Unit Promote

`unit_promote:{unit}` promotes verified candidate truth to stable truth.

## Input

- `docs/specs/_status.md`
- `docs/specs/_verify_result/unit/{unit}.md`
- `docs/specs/units/candidate/c_unit_{unit}.md`
- `docs/specs/units/stable/s_unit_{unit}.md`
- Current unit's candidate-layer appendix files
- `framework/spec_writing_guide.md` (for evidence appendix promotion restriction and stable truth field rules)
- `framework/candidate_intent.md` (for candidate-only frontmatter fields that must be stripped)
- `framework/process_snapshot_contract.md` (for stable promotion summary format)

## Pre-Execution Self-Check (MANDATORY)

Before executing this step, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm the target unit's `Next Command` is `unit_promote`.
2. [ ] If `_status.md` is empty (no units registered): STOP, report that no units are registered, and suggest `unit_new` as the first step.
3. [ ] Read `docs/specs/_verify_result/unit/{unit}.md` — confirm verification passed with `ready_to_promote`.
4. [ ] Confirm both candidate-layer and stable-layer Spec files exist.
5. [ ] If any check fails: STOP, report what is missing, and do not proceed.

If all checks pass: proceed to "What This Step Does" below.

1. Write candidate truth (main Spec + non-evidence appendices) as stable-layer truth. Evidence appendix files (referenced by `evidence_appendix_ref`) must not be promoted to stable truth as behavior-correctness claims. Strip candidate-only frontmatter fields (`candidate_intent`, `evidence_appendix_ref`, `source_basis`, `repair_basis`, and any command-specific fields) when writing stable truth. Rewrite Markdown document references within the promoted spec body and promoted non-evidence appendices from candidate paths (`c_unit_*`) to stable paths (`s_unit_*`).
2. Update lifecycle state and refs
3. Clean up candidate-layer evidence files

This is a mechanical operation that does not involve new behavior judgment.
`unit_promote` does not need a new independent review — it consumes the evidence already verified by `unit_verify`.

## Not Allowed

- Introduce behavior, acceptance, ownership, or rule meaning outside the verified scope
- Modify implementation files
- Manually modify lifecycle state
- Delete candidate-layer evidence before `command close --apply` completes

## How to End

| Result | Meaning | Next Step |
|--------|---------|-----------|
| `promoted` | Promotion succeeded | `command close --command unit_promote --object-type unit --object <unit> --outcome promoted --apply`. After success: `Active Layer=stable`, `Next Command=unit_fork`, candidate-layer evidence is cleaned up. Tooling writes the stable promotion summary at `docs/specs/_verify_result/stable/unit/{unit}.md` per `framework/process_snapshot_contract.md` Section 8 format. |
| `promotion_recovered` | Promotion partially mutated stable truth | Restore candidate state and apply recovery rules from `framework/lifecycle/recovery.md`. |
| `verify_invalid`* | Verify result became invalid between close and apply | Handled by tooling fallback machinery. Sub-types: `truth`, `binding`, `baseline`, `rule`, `gate`, `evidence`. See `recovery.md` for per-type recovery. |

\* `verify_invalid_*` outcomes are handled by the tooling fallback machinery and do not require agent action beyond following recovery.md.
