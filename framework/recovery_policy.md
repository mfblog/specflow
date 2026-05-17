# Recovery Policy

Recovery handles incomplete command cleanup or invalidated process evidence.

The only supported lifecycle object is `unit`.

## 1. Candidate Fallback

Candidate unit fallback may delete current-round process files according to the failure layer:

1. `truth_layer` deletes check, plan, and verify process files and sets `Next Command=unit_check`
2. `gate_layer` deletes check process files and sets `Next Command=unit_check`
3. `plan_layer` deletes active plan, draft plan, and verify process files and sets `Next Command=unit_plan`
4. `implementation_layer` deletes verify process files and sets `Next Command=unit_impl`
5. `evidence_layer` deletes verify process files and sets `Next Command=unit_verify`

## 2. Success Cleanup

Successful `unit_fork` and `unit_promote` cleanup may delete obsolete candidate process files only after `_status.md` has been updated to the correct legal state.

## 3. Unit Dependency Recovery

If promotion changes a stable unit version, every dependent unit that still references the older stable ref must be rerouted before the round is closed.

## 4. Rejection

Recovery must reject:

1. `object-type=scenario`
2. `scenario_*` commands
3. scenario process paths
4. scenario truth paths

## 5. Post-Mutation Recovery Boundary

Ordinary process invalidation uses the candidate fallback rules from Section 1.

Post-mutation recovery is different.

It is required only when an active command or governance flow has already written, updated, or deleted truth files and can no longer safely close.

Rules:

1. do not claim the round succeeded
2. restore repository truth to the captured pre-mutation state
3. delete new files created only by the interrupted or unsafe round
4. update `_status.md` to the smallest legal restart state for affected units
5. rerun routing from current repository truth after recovery

## 6. Incomplete Promotion Recovery

`unit_promote` must capture a recovery baseline before the first truth-file mutation.

The baseline must cover every file the round may overwrite, retarget, or delete.

At minimum it must cover:

1. the target unit row in `docs/specs/_status.md`
2. `docs/specs/units/candidate/c_unit_{unit}.md`
3. `docs/specs/units/stable/s_unit_{unit}.md` when it already exists
4. current-round unit process files
5. candidate appendix files for the target unit
6. stable appendix files for the target unit when the round may update or delete them
7. rule files and unit files that the promotion may retarget or rewrite in the same round
8. every affected `_status.md` row when dependency retargeting may change another unit's legal next step

Incomplete promotion recovery is required when both are true:

1. `unit_promote` has already mutated at least one promotion target, retarget target, or deletion target
2. the round can no longer safely complete promotion closure

Recovery procedure:

1. stop claiming promotion success
2. restore every mutated or deleted file covered by the recovery baseline to its exact pre-mutation bytes
3. delete every new file created only by the incomplete promotion round when that file did not exist in the baseline
4. restore `_status.md` for the target unit to candidate semantics:
   - keep `Candidate=yes`
   - keep `Active Layer=candidate`
   - set `Next Command=unit_check`
5. keep `Stable=yes|no` consistent with the pre-round state from the recovery baseline
6. delete the target unit's candidate-side process files because they are no longer safe for reuse

After incomplete promotion recovery completes, the only safe claim is that promotion did not complete and the unit must restart from `unit_check`.

### 6.5 Rule-Governance Recovery

Rule-governance recovery applies to `rule_new`, `rule_extract`, `rule_bind`, and `rule_topology`.

It is required when both are true:

1. the rule-governance flow has already mutated rule truth files, `docs/specs/repository_mapping.md`, or downstream unit candidate files
2. the flow can no longer safely close and must return control through `rule_escape` or natural-language routing

Before the first file mutation, a rule-governance flow must capture a recovery baseline for every file it may touch.

The baseline must include:

1. target rule files under `docs/specs/rules/**`
2. stable sibling rule files when the flow may create, update, or delete them
3. candidate unit files whose `rule_refs` or rule-reuse prose may be rewritten
4. `docs/specs/repository_mapping.md` when the rule object map may change
5. affected `_status.md` rows when downstream unit fallback may be written

Recovery procedure:

1. stop claiming the rule-governance flow succeeded
2. restore every mutated file covered by the recovery baseline to its exact pre-mutation bytes
3. delete every new file created only by the interrupted rule-governance round when that file did not exist in the baseline
4. restore `docs/specs/repository_mapping.md` when it was modified
5. if any downstream candidate unit was restored, rerun routing from current repository truth before claiming a new command entry

After rule-governance recovery completes, the next legal entry is natural-language routing from current repository truth.

## 7. Reason Codes

Use `promotion_recovery` only when incomplete promotion recovery restored a unit promotion round.

Use `rule_governance_recovery` only when rule-governance recovery restored a rule-governance flow.
