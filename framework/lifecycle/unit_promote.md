# Unit Promote

`unit_promote:{unit}` promotes verified candidate truth to stable truth.

## Input

> **Reading guidance:** Must Read files are the truth and process data this command evaluates. May Reference files hold the format and policy contracts — read them when a specific check question needs the exact rule text. Procedural instructions are inline in "What This Step Does" and "How to End" below.

### Must Read

- `docs/specs/_status.md`
- `docs/specs/_verify_result/unit/{unit}.md`
- `docs/specs/units/candidate/c_unit_{unit}.md`
- `docs/specs/units/stable/s_unit_{unit}.md`
- Current unit's candidate-layer appendix files

### May Reference

- `framework/spec_writing_guide.md` (evidence appendix promotion restriction and stable truth field rules)
- `framework/candidate_intent.md` (candidate-only frontmatter fields that must be stripped)
- `framework/process_snapshot_contract.md` (stable promotion summary format)

## Pre-Execution Self-Check (MANDATORY)

Before executing this step, you MUST verify:

1. [ ] Read `docs/specs/_status.md` — confirm the target unit's `Next Command` is `unit_promote`.
==ATOM_BEGIN:shared_guards==
2. [ ] If `_status.md` is empty (no units registered): STOP, report that no units are registered, and suggest `unit_new` as the first step.
==ATOM_END:shared_guards==
3. [ ] Read `docs/specs/_verify_result/unit/{unit}.md` — confirm verification passed with `ready_to_promote`.
4. [ ] Confirm both candidate-layer and stable-layer Spec files exist.
5. [ ] If any check fails: STOP, report what is missing, and do not proceed.

If all checks pass: proceed to "What This Step Does" below.

## What This Step Does

0. **Before writing stable truth**, validate that `rule_refs` and `unit_refs` in the candidate frontmatter reference current, existing rule and unit versions. If any ref points to a version that has been superseded or no longer exists as a rule file, update the refs to the current version before writeback. If the ref cannot be resolved, STOP and route through `unit_check:{unit}` for re-validation.

1. Write candidate truth (main Spec + non-evidence appendices) as stable-layer truth. Evidence appendix files (referenced by `evidence_appendix_ref`) must not be promoted to stable truth as behavior-correctness claims. Strip candidate-only frontmatter fields (`candidate_intent`, `evidence_appendix_ref`, `source_basis`, `repair_basis`, and any command-specific fields) when writing stable truth. Rewrite Markdown document references within the promoted spec body and promoted non-evidence appendices from candidate paths (`c_unit_*`) to stable paths (`s_unit_*`). Rewrite the `layer` frontmatter field in each promoted file from `candidate` to `stable`. After rewriting, verify that no `c_unit_*` references remain in any promoted stable file.
2. Update `docs/specs/repository_mapping.md` spec_files to reference the stable Spec path `s_unit_{unit}.md`, replacing the candidate Spec path. If the promoted unit's candidate rule files carry `promotion_owner_unit`, this is also the point to promote those candidate rules to stable alongside the unit (the `rule_sync` impact handoff after `command close` reconciles downstream consumers).
3. Update lifecycle state and refs
4. Clean up candidate-layer evidence files

This is a mechanical operation that does not involve new behavior judgment.
`unit_promote` does not need a new independent review — it consumes the evidence already verified by `unit_verify`.

## Not Allowed

- Introduce behavior, acceptance, ownership, or rule meaning outside the verified scope
- Modify implementation files
- Manually modify lifecycle state
- Delete candidate-layer evidence before `command close --apply` completes

## Allowed Writes

- `docs/specs/units/stable/s_unit_{unit}.md` — stable main Spec (written from candidate truth with frontmatter stripping)
- `docs/specs/units/stable/appendix/s_unit_{unit}_*.md` — stable appendix files (copied from candidate non-evidence appendices)
- `docs/specs/repository_mapping.md` — spec_files update to reference stable Spec
- `docs/specs/_verify_result/stable/unit/{unit}.md` — stable promotion summary (written by tooling)

## How to End

| Result | Meaning | Next Step |
|--------|---------|-----------|
| `promoted` | Promotion succeeded | `command close --command unit_promote --object-type unit --object <unit> --outcome promoted --apply`. After success: `Active Layer=stable`, `Next Command=unit_fork`, candidate-layer evidence is cleaned up. After cleanup, run `framework/governance/impact_sync.md` to check downstream impact: impact_sync identifies units that consume this stable version, detects drift (truth_drift, binding_drift, etc.), and routes affected units to `unit_check` (candidate), `unit_stable_verify` (stable), or marks them as `no_drift_observed`. See `framework/governance/impact_sync.md` for the full procedure. Tooling writes the stable promotion summary at `docs/specs/_verify_result/stable/unit/{unit}.md` per `framework/process_snapshot_contract.md` Section 9 format. |
| `promotion_recovered` | Promotion partially mutated stable truth | Restore candidate state and apply recovery rules from `framework/lifecycle/recovery.md`. |
| `verify_invalid`* | Verify result became invalid between close and apply | Handled by tooling fallback machinery. Sub-types: `truth`, `binding`, `baseline`, `rule`, `gate`, `evidence`. See `recovery.md` for per-type recovery. |

\* `verify_invalid_*` outcomes are handled by the tooling fallback machinery and do not require agent action beyond following recovery.md.

Tooling invocation: `specflowctl command close --command unit_promote --object-type unit --object <unit> --outcome promoted --apply`
==ATOM_BEGIN:close_fallback==
### Manual Command Close (when `specflowctl` is unavailable)

When `specflowctl command close` is unavailable (tooling not installed, broken, or
inaccessible), read `framework/lifecycle/command_close_fallback.md` for the complete
manual command close procedure.
==ATOM_END:close_fallback==
