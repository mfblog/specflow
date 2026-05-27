# Freshness Impact

Freshness impact classifies process evidence drift after deterministic snapshot validation.

It does not replace fingerprints. It prevents a text-level fingerprint change from automatically becoming lifecycle fallback when formal behavior evidence is still reusable.

## Levels

`current` means the stored truth or spec fingerprint matches the current normalized file fingerprint.

`text_drift` means only the normalized file text fingerprint changed. File refs, version refs, acceptance behavior fingerprint, acceptance item set, and dependency snapshots still match.

`semantic_drift` means the formal acceptance behavior fingerprint changed while the acceptance item id, verification surface, and runnable set still match.

`acceptance_drift` means the acceptance item set changed.

`dependency_drift` means appendix, unit dependency, rule, or repository mapping snapshots changed.

`schema_drift` means the process file shape, gate fields, coverage fields, receipt fields, or evidence matrix is invalid.

`unknown_drift` means the process file predates the behavior fingerprint needed to distinguish text drift from semantic drift.

## Evidence Reuse

Only `text_drift` can reuse existing process evidence.

Text drift reuse requires all of these:

1. deterministic tooling classifies the process as `text_drift`.
2. the process records `evidence_reuse: accepted`.
3. the process records `freshness_current_fingerprint` for the current truth or spec file.
4. an independent reviewer receipt confirms reuse with reviewer pack `freshness_text_drift_reuse`, `freshness_review_mode: independent`, `freshness_reviewer_result: pass`, `freshness_reviewer_context: minimal_context`, and `freshness_review_findings: none`.

The reviewer judges whether the text-only edit preserves the intended meaning of the existing evidence. Tooling only verifies the mechanical classification and receipt fields.

`semantic_drift`, `acceptance_drift`, `dependency_drift`, `schema_drift`, and `unknown_drift` cannot be accepted through freshness reuse. They must use the existing smallest legal validation, repair, or fallback path.

## Lifecycle Effect

Accepted `text_drift` evidence is current valid evidence for standard lifecycle progression.

Unaccepted `text_drift` is a `freshness_layer` stop. It must not delete process files or reroute `_status.md`; the next action is independent freshness review or evidence recreation.

Manual hashes, chat agreement, and executor-only claims do not establish freshness impact or evidence reuse.
