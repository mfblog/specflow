# Recovery

<!-- AGENT: This is an internal framework document. The Agent only needs to read it when tooling redirects there or when a recovery scenario occurs. It is not needed during normal operation. -->

Recovery resets unsafe process evidence to the smallest legal restart point.

It applies to unit lifecycle work, rule-governance work that already mutated files, and impact sync outcomes that invalidate downstream evidence.

## Unit Fallback Targets

> **Note:** This table covers both candidate and stable unit evidence. The `evidence_layer` row routes candidate evidence to `unit_verify` and stable verify evidence to `unit_stable_verify` — see the per-row routing for the correct restart command. `binding_drift` and `rule_drift` in the `truth_layer` row apply to candidate units only; stable units with those reason codes are routed per `framework/governance/impact_sync.md` Fallback Routing item 6 (to `unit_stable_verify`). Stable-layer truth changes follow the separate rules in [Stable Unit Recovery](#stable-unit-recovery) below.

Layer classification maps a failure to the layer whose evidence is invalidated. See `framework/process_snapshot_contract.md` Section 4 (Fallback Layers) for the classification rules: truth mismatch → `truth_layer`, check schema or gate evidence mismatch → `gate_layer`, verify evidence mismatch → `evidence_layer`.

| Failure Layer | Reason Codes | Deletes | Next Command |
|---|---|---|---|
| `truth_layer` | `truth_drift`, `binding_drift` (candidate only), `baseline_drift`, `rule_drift` (candidate only), `truth_incomplete` | check_work, check result (if any), verify result | `unit_check` (clears `Notes`) |
| `gate_layer` | `gate_missing`, `spec_issue` | check_work, check result (if any) | `unit_check` (clears `Notes`) |
| `evidence_layer` | `evidence_incomplete` (candidate-layer only), `stable_verify_invalid` | verify result or stable verify result | `unit_verify` if `evidence_incomplete`; `unit_stable_verify` if `stable_verify_invalid` |

Only reason codes in this table are valid for fallback cleanup.
Do not introduce alternate names for the same invalidated layer.
Use the earliest layer that is invalidated by current repository truth.
Do not delete upstream process files that still validate and are still supported by current truth.

### Freshness Review Required

When impact sync reports `freshness_review_required`, run the deterministic freshness classification first:

```text
./specflow/tooling/bin/specflowctl-<os>-<arch> snapshot validate-process --object-type unit --object <unit> --process check|verify|stable_verify
```

The tooling output classifies the freshness state and returns the resolved branch. If tooling is unavailable, the process snapshot is stale (fingerprint mismatch) but truth has not drifted. This is not a fallback layer — no evidence cleanup is required. The agent must re-read the current truth files to verify the snapshot's claims against the latest content, then:
1. If truth content has changed since the snapshot was recorded: re-classify under the appropriate fallback layer (`truth_layer`).
2. If only timestamp or fingerprint metadata is stale: generate an independent evaluation request with pack `freshness_text_drift_reuse` per `framework/core/freshness.md` section Evidence Reuse and `framework/core/independent_evaluation.md` Handoff Requests. After the reviewer returns `pass`, write the freshness_receipt with `freshness_review_mode: independent` (see `framework/process_snapshot_contract.md` Section 2 — Freshness reuse fields). Do not self-certify `freshness_review_mode: independent`.
3. If the snapshot cannot be verified: delete the stale process evidence and rerun the originating command.

### truth_drift Recovery Procedure

When the failure is classified as `truth_layer` and the cause is `truth_drift` (candidate truth has diverged from the stable baseline — see `framework/lifecycle/unit_verify.md` `truth_fallback` outcome), apply the following procedure:

1. **Determine the baseline and correct candidate truth** — Before deleting any evidence, decide which baseline applies. If the candidate diverged from the stable Spec (e.g., through `truth_fallback` from `unit_verify` at `framework/lifecycle/unit_verify.md:95`), restore `docs/specs/units/candidate/c_unit_{unit}.md` to alignment with the stable Spec at `docs/specs/units/stable/s_unit_{unit}.md`. Remove content that contradicts the stable baseline and restore any content that was incorrectly omitted. If the candidate was intentionally modified during implementation without re-validation (a legitimate scope addition), preserve all intentional changes — the subsequent `unit_check` re-validation will determine their validity. (See `framework/candidate_intent.md` for the distinction between change candidates — which intentionally diverge from the stable layer — and repair candidates — which must not.)
2. **Delete invalid evidence** — Remove process artifacts per the `truth_layer` row in the fallback table above (check_work, check result, verify result).
3. **Reset lifecycle state** — Run `./specflow/tooling/bin/specflowctl-<os>-<arch> command close --command unit_verify --object-type unit --object {unit} --outcome truth_fallback --apply` —
   this sets `Next Command=unit_check` and clears `Notes` per the fallback table.
   (See `unit_verify.md` How to End `truth_fallback` outcome.)
4. **Re-validate** — Run `unit_check:{unit}` on the corrected candidate truth.
5. **Get post-recovery directive** — After re-validation passes, run `./specflow/tooling/bin/specflowctl-<os>-<arch> next --unit {unit}` to obtain the deterministic directive for the next governance step. The directive tells you TASK, READS, WRITES, BLOCKED, and COMPLETION.

## Candidate Recovery

When candidate truth changes, bound rule references change, repository mapping changes the unit boundary (maps to `truth_layer`, reason: `baseline_drift`), or a global rule changes the candidate's constraints:

1. delete downstream evidence that was derived from the prior truth.
2. set the candidate unit's next command to the earliest required command from the fallback target table.
3. keep still-valid upstream evidence only when deterministic validation proves it still matches current truth.
4. rerun impact sync when the change may affect other units.

If a candidate main Spec changes after `unit_check`, the check result (if any) may need revalidation. Verify evidence may still be valid if the spec change does not affect acceptance items or verification scope.

When `unit_verify` reports `spec_issue` (candidate Spec needs repair without implementation change — see `framework/lifecycle/unit_verify.md` How to End), only the spec requires repair. The verify evidence remains valid for the unchanged acceptance items. Do not delete verify evidence. Set next command to `unit_check`. Delete check_work and check_result per the `gate_layer` row in the fallback table above — they contain evidence against a now-invalid spec. The subsequent `unit_check` round will regenerate them against the repaired spec.

## Stable Unit Recovery

Stable truth is not edited in place through recovery.
If a stable unit needs a behavior, boundary, acceptance, or rule-binding change, start a candidate through `unit_fork:{unit}` and use `candidate_intent=change` or `candidate_intent=repair` as appropriate.

Stable implementation drift routes to `unit_stable_verify:{unit}`.
Stable truth drift that requires a new version routes through `unit_fork:{unit}`.

## Promotion Recovery

If promotion has not mutated stable truth yet, recover by resetting the candidate to the earliest invalidated command.

If promotion has already mutated stable truth but closure is incomplete:

1. do not silently keep partial promotion state.
2. restore the unit to a deterministic candidate state when stable truth cannot be proven complete.
3. delete process evidence that references the incomplete promotion result.
4. set the next command to `unit_check` (truth_layer fallback) — partial promotion invalidates all verify evidence.
5. rerun impact sync for any stable dependency or rule consumer that could observe the promotion attempt.

## Rule-Governance Recovery

Rule-governance flows must capture a recovery baseline before the first file mutation.
The baseline must include the files that would need to be restored or revalidated if the flow cannot close.

The baseline is an execution-local checklist (not checked into repository truth) listing: (1) file paths of every rule file about to be mutated, (2) a SHA-256 fingerprint of each file before mutation, (3) the current `_status.md` row for any affected unit. It is not consumed by any lifecycle gate.

If repository truth becomes insufficient before mutation, stop and route through `framework/governance/rules/rule_escape.md`.

If mutation already happened and the rule flow cannot safely close:

1. stop further rule mutation.
2. restore or complete the smallest set of rule files needed to make repository truth deterministic.
3. run `rule_sync` when any affected unit or rule consumer may have downstream drift.
4. apply candidate or stable unit recovery for every affected unit.
5. return to `framework/operations/entry_routing.md` only after the repository is no longer left in a partial rule-governance state.

## Success Cleanup

After a command or rule flow closes successfully:

1. remove process files that the closing command explicitly supersedes.
2. keep evidence that remains current and is still required by the next command.
3. never keep stale downstream evidence as a historical shortcut.
4. ensure `_status.md` names the next legal command for every affected unit.

### Cleanup Mode Reference

| Mode | Deleted | Preserved |
|------|---------|-----------|
| `unit_fork` | Process artifacts (check_work, check result, verify result, stable_verify result) and agent-internal artifacts (plan) for the target unit | Stable unit truth (main Spec + appendices) unchanged, candidate unit truth (main Spec + appendices) intact |
| `unit_promote` | Candidate main Spec, candidate appendix files, process artifacts | Stable main Spec written by promotion; stable appendix files (copied from candidate before cleanup). Candidate appendix files (including evidence appendices) are deleted during cleanup regardless of content. Promotion summary at `docs/specs/_verify_result/stable/unit/{unit}.md` |

The stable promotion summary is written by tooling (`command close --apply`) before cleanup begins, so it is preserved at a separate path that cleanup globs do not match. See `framework/process_snapshot_contract.md` Section 8 for the summary format and `tooling/internal/commandclose/commandclose.go` for implementation details.

## Removed Scenario Lifecycle

Requests that use `scenario_*`, `scenario_advance:{id}`, or `object-type=scenario` are not recoverable lifecycle work.
Stop and report that scenario lifecycle support has been removed.

## Governance Exception (Force Bypass)

`specflowctl command close --force --force-reason <reason>` allows a caller to bypass non-critical validation checks during command close. The bypass records a `forced:` entry in the unit's Notes to enable audit and downstream recovery.

### Bypassable Checks

The following checks may be bypassed with `--force`:

1. **Process validation** — when `snapshot validate-process` reports a mismatch between current truth and the stored process evidence. Common causes: intentional spec amendment during implementation, appendix file changes outside the lifecycle flow, or transient repository state that will be resolved by the next lifecycle command.
2. **Unit fork appendix coverage** — when `unit_fork` with outcome `candidate_created` detects missing candidate appendix files for one or more stable appendix references.

The following checks are NEVER bypassable with `--force`:

1. **Controlled fork intent validation** (`validateControlledStableVerifyForkIntent`) — fork intent validation checks whether stable verify evidence supports the requested fork. Bypassing this gate would allow lifecycle advancement without valid stable-verify evidence. This is a governance hard constraint and must not be overridden.

### Usage Rules

1. `--force-reason` is required. A descriptive reason must explain why the bypass is necessary and what downstream action will resolve the bypassed condition.
2. Force bypass must not be used to avoid fixing legitimate governance gaps. It is intended for transient, non-recurring scenarios where the bypassed check produces a false positive or where the repository state will be corrected by the next lifecycle command.
3. Repeated use of `--force` for the same unit and same check type without resolution is a governance concern and must be flagged during review.

### Notes Recording

When `--force` bypasses a check, the tooling appends a `forced:` entry to the unit's `Notes` in `_status.md`:

```text
forced:<check_type>:<reason>
```

- `check_type` is `validation` (process validation bypass) or `appendix_coverage` (fork appendix coverage bypass).
- `reason` is the `--force-reason` value passed by the caller.
- If `--force-reason` is empty (invalid — tooling rejects this), the reason defaults to `no_reason`.

Multiple `forced:` entries accumulate in Notes, separated by `;`.

### Recovery After Force Bypass

A `forced:` Notes entry indicates that a governance gate was bypassed. The condition that caused the bypass must be resolved before the unit advances to the next lifecycle command:

1. For `forced:validation` — the process validation mismatch must be corrected or the check result must be re-created by re-running `unit_check` through the re-validation path. `snapshot --update-check-result` may be used only when text-only drift caused the mismatch (see `framework/process_snapshot_contract.md` §12.2).
2. For `forced:appendix_coverage` — the missing candidate appendix files must be created, or the `appendix_exc:` exclusion must be confirmed as intentional (see `framework/core/status.md` §Appendix Coverage Exclusions).
3. After resolution, the `forced:` entry should be removed from Notes. The executor must not remove `forced:` entries without verifying that the underlying condition has been resolved.

### Audit Trail

Every `--force` use creates an audit record in `_status.md` Notes. Governance reviews and lifecycle audits MUST inspect `forced:` entries and verify that each one has a documented resolution path. Unresolved `forced:` entries at lifecycle command boundaries are governance findings.
