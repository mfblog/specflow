<!-- DEPRECATED: Templates are no longer used. Guidance is now generated programmatically in tooling/internal/contextcard/card.go -->

# Context Card: unit/{unit}

## STATUS
- Stage: unit_promote | Next: unit_fork (on success)
- Layer: candidate → stable

## GUIDANCE
Promote the verified candidate truth to stable truth. This is a mechanical operation — no new independent review needed (the evidence was already verified in unit_verify).

> **Pre-check:** Read `docs/specs/_status.md`. Confirm Next=unit_promote. Read `docs/specs/_verify_result/unit/{unit}.md` and confirm it has outcome=ready_to_promote.

**Execution steps:**
1. Read the candidate spec at `docs/specs/units/candidate/c_unit_{unit}.md` + appendices.
2. Copy candidate spec to stable path: `docs/specs/units/stable/s_unit_{unit}.md`.
3. Strip candidate-only frontmatter fields: `candidate_intent`, `evidence_appendix_ref`, `source_basis`, `repair_basis` (see `framework/candidate_intent.md`).
4. Rewrite Markdown body references: change `c_unit_*` to `s_unit_*` in doc links.
5. Copy non-evidence appendices: re-prefix from `c_unit_*` to `s_unit_*` under `docs/specs/units/stable/appendix/`.
6. Do NOT promote evidence appendix files (those stay in candidate layer).
7. Write promotion summary to `docs/specs/_verify_result/stable/unit/{unit}.md` per `framework/process_snapshot_contract.md` Section 8.
8. Do NOT delete candidate-layer evidence files until after `command close --apply`.

**Close (MUST include --apply flag):**
`specflowctl command close --command unit_promote --object-type unit --object {unit} --outcome promoted --apply`

The `--apply` flag is required — it tells the system to commit the transition. Without it, the candidate evidence will not be cleaned up and the status will not advance.

> If the current request does not involve spec changes: not applicable — promote is a promotion operation, not implementation.

## WRITES (owned by this unit)
- docs/specs/units/stable/s_unit_{unit}.md (promoted stable spec)
- docs/specs/units/stable/appendix/s_unit_{unit}_*.md (non-evidence appendices)
- docs/specs/_verify_result/stable/unit/{unit}.md (promotion summary)

## READS (read-only context)
- docs/specs/_status.md
- docs/specs/_verify_result/unit/{unit}.md
- docs/specs/units/candidate/c_unit_{unit}.md + appendices
- docs/specs/units/stable/s_unit_{unit}.md
- framework/lifecycle/unit_promote.md
- framework/spec_writing_guide.md
- framework/candidate_intent.md
- framework/process_snapshot_contract.md

## BLOCKED
- Introducing behavior, acceptance, ownership, or rule implications beyond the verified scope
- Modifying implementation files
- Manually modifying lifecycle state
- Deleting candidate-layer evidence before command close --apply completes

## CLOSE
specflowctl command close --command unit_promote --object-type unit --object {unit} --outcome promoted --apply

## Next Steps
promoted → stable layer updated, next step unit_fork
Re-run: specflowctl context card
