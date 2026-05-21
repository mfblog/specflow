# Process Snapshot Contract

Process files record what a unit command checked in one round.

They are evidence, not behavior truth.

When a supported process file carries slice work state, the generic slice terms and mechanical maintenance boundaries come from `specflow/framework/slice_work_state_protocol.md`.
This file defines only the supported process paths, process snapshots, unit-check work-state fields, freshness rules, and validation rules.

## 1. Supported Process Paths

Supported unit process paths:

1. unit check work state: `docs/specs/_check_work/unit/{unit}.md`
2. check result: `docs/specs/_check_result/unit/{unit}.md`
3. active plan: `docs/specs/_plans/active/{unit}.md`
4. draft plan: `docs/specs/_plans/draft/{unit}.md`
5. verify result: `docs/specs/_verify_result/unit/{unit}.md`
6. stable promotion summary: `docs/specs/_verify_result/stable/unit/{unit}.md`

No `scenario` process path is supported.

`_check_work` is a command-local work-state path for `unit_check`.
It is not a pass gate and is not consumed by `unit_plan`.
Downstream handoff commands consume only `_check_result`, `_plans/active`, and `_verify_result` according to their command contracts.
Other process files may carry command-owned business slice records only when their owning command defines that adoption in its command file.

## 2. Common Fields

Check and verify process YAML must identify the unit and command gate:

```yaml
object_type: unit
object_ref: {unit}
gate: unit_check|unit_verify
decision: pass|blocked|fix_required
allow_next: true|false
next_command: unit_plan|unit_impl|unit_verify|unit_promote|unit_check
truth_layer_ref: candidate|stable
truth_file_ref: docs/specs/units/{layer}/{file}.md
truth_version_ref: c_unit_{unit}@x.y.z
truth_fingerprint: {fingerprint}
```

Plan process YAML must identify the unit truth it planned from:

```yaml
spec_file_ref: docs/specs/units/candidate/c_unit_{unit}.md
spec_version_ref: c_unit_{unit}@x.y.z
spec_fingerprint: {fingerprint}
```

## 3. Dependency Snapshots

Process files may record:

```yaml
unit_appendix_snapshot: none
unit_snapshot: none
rule_snapshot: none
```

or lists.

`unit_appendix_snapshot` records appendix files explicitly used by the unit round.

`unit_snapshot` records stable unit dependencies resolved from current unit `unit_refs`.

`rule_snapshot` records rules resolved from current unit `rule_refs`.

If a snapshot field is present, tooling must validate it against current truth.

## 4. Fallback Layers

Process validation failure maps to these layers:

1. truth mismatch -> `truth_layer`
2. check schema or gate evidence mismatch -> `gate_layer`
3. plan schema or plan coverage mismatch -> `plan_layer`
4. verify evidence mismatch -> `evidence_layer`

The legal fallback commands are:

1. `truth_layer` -> `unit_check`
2. `gate_layer` -> `unit_check`
3. `plan_layer` -> `unit_plan`
4. `evidence_layer` -> `unit_verify`

## 5. Rejection

Tooling must reject:

1. `object_type: scenario`
2. `--object-type scenario`
3. process files under `docs/specs/_check_result/scenario/**`
4. process files under `docs/specs/_verify_result/scenario/**`

It must not convert those files into unit evidence.

## 6. Fingerprint Contract

The process fingerprint algorithm is fixed:

1. normalize file text according to Section 7
2. encode the normalized text as UTF-8
3. compute `sha256`
4. render the result as lowercase hexadecimal

This same fingerprint contract applies to:

1. `truth_fingerprint`
2. `spec_fingerprint`
3. `unit_appendix_snapshot` item fingerprints
4. `unit_snapshot` item fingerprints
5. `rule_snapshot` item fingerprints
6. `stable_truth_fingerprint`

## 7. Text Normalization Rules

Before hashing a markdown truth file, normalize it in this exact order:

1. read the full file text
2. convert all line endings to `LF`
3. if the file does not end with `LF`, append exactly one trailing `LF`
4. do not trim leading spaces
5. do not trim trailing spaces inside lines
6. do not remove blank lines
7. do not reorder frontmatter keys
8. do not apply markdown-aware semantic rewriting

The fingerprint represents the actual normalized file text.

If the text changes, the fingerprint changes.

## 8. Stable Promotion Summary

`docs/specs/_verify_result/stable/unit/{unit}.md` stores the minimal acceptance coverage summary preserved by `unit_promote`.

It is not:

1. a gate-bearing process file
2. a substitute for stable unit truth
3. a current implementation-alignment claim after later code changes

Each stable promotion summary must record:

1. `object_type`
   - must be `unit`
2. `object_ref`
   - the bare unit id
3. `stable_truth_file_ref`
   - the stable unit file written by promotion
4. `stable_truth_version_ref`
   - `s_unit_{unit}@<version>`
5. `stable_truth_fingerprint`
   - the Section 6 fingerprint of `stable_truth_file_ref`
6. `promotion_verify_result_ref`
   - the current-round verify result file consumed before cleanup
7. `acceptance_item_set`
   - the promoted acceptance item set by item id
8. `acceptance_item_coverage_summary`
   - the final promoted status for each acceptance item
9. `key_evidence_source_refs`
   - the smallest useful references to commands, tests, inspections, or manual evidence that proved the promoted items

Later stable verification may read this summary as background, but it must collect current evidence before making a new stable-alignment claim.

## 9. Unit Check Work State

`docs/specs/_check_work/unit/{unit}.md` records resumable `unit_check` progress for one target candidate.
It follows the generic state-carrier and slice-field standards in `specflow/framework/slice_work_state_protocol.md`, with the command-specific fields below.

It is:

1. a process work-state file
2. a resume aid for the current `unit_check` round
3. a place to record slice status, input fingerprints, finding references, blocked reason, and next resume step

It is not:

1. a Spec
2. behavior truth
3. a downstream pass gate
4. a substitute for `_check_result/unit/{unit}.md`
5. a place for tooling to decide semantic pass, finding severity, or final conclusion

### 9.1 Required Run Fields

The work-state run table must record:

```yaml
work_flow: unit_check
work_id: YYYYMMDD-HHMMSS-unit_check-{unit}
object_type: unit
object_ref: {unit}
status: in_progress|blocked_on_finding|ready_for_final|closed_pass|closed_blocked|closed_fix_required
created_at: YYYY-MM-DDTHH:MM:SSZ
last_updated_at: YYYY-MM-DDTHH:MM:SSZ
active_slice: {slice_id}
truth_layer_ref: candidate
truth_file_ref: docs/specs/units/candidate/c_unit_{unit}.md
truth_version_ref: c_unit_{unit}@x.y.z
truth_fingerprint: {fingerprint}
baseline_slice_table: present
dynamic_slice_table: none|present
finding_refs: none|{refs}
blocked_reason: none|{reason}
resume_next_step: {step}
```

`truth_fingerprint` uses the same normalized SHA-256 contract as Section 6.

### 9.2 Slice Table Fields

Baseline and dynamic slice tables use these fields:

```text
slice_id
slice_origin
slice_type
status
review_question
why_added
parent_slice_id
input_files
input_fingerprint
depends_on
finding_refs
result_summary
exit_condition
resume_next_step
```

Allowed `slice_origin` values:

1. `baseline`
2. `dynamic`

Allowed `slice_type` values:

1. `local`
2. `cross_convergence`

Allowed `status` values:

1. `pending`
2. `passed`
3. `blocked`
4. `stale`
5. `skipped_not_applicable`

Every dynamic slice must name an existing `parent_slice_id`.
The parent may be a baseline slice or another dynamic slice.
Dynamic slices must not replace required baseline slices.

### 9.3 Baseline Slice Skeleton

The mechanical work-state skeleton for `unit_check` must include these baseline local slices:

1. `goal_and_responsibility`
2. `dependency_truth_surface`
3. `main_flow_and_state`
4. `boundary_and_protocol`
5. `data_artifact_and_output`
6. `error_edge_and_gap`
7. `acceptance_and_testability`
8. `implementation_handoff`

It must include these baseline cross-check slices:

1. `goal_to_acceptance_convergence`
2. `flow_to_boundary_convergence`
3. `dependency_truth_convergence`
4. `output_to_acceptance_convergence`

### 9.4 Freshness And Stale Rules

Before a `unit_check` pass gate is written, the work-state file must be refreshed and validated.

Refresh rules:

1. recompute the work-state truth fingerprint from the current target candidate
2. recompute each slice `input_fingerprint` from the current `input_files`
3. if a slice is `passed` and its input fingerprint changes, mark that slice `stale`
4. if a slice is `passed` and one of its input files is missing, mark that slice `stale`
5. if a cross-check slice is `passed` and any slice in `depends_on` is `stale`, mark the cross-check slice `stale`
6. update `last_updated_at`

If truth drift, binding drift, fallback cleanup, unit fork, unit promote, rule release, or project-instance migration invalidates the target candidate's prior check state, the old `_check_work` file must be deleted or marked unusable by the owning cleanup path.

### 9.5 Tooling Boundary

`specflowctl process check-work-*` commands may:

1. create the work-state skeleton
2. validate field presence, legal values, slice table shape, dynamic parent links, object type, and repository-relative input paths
3. refresh timestamps
4. compute input fingerprints
5. mark stale slices caused by input or dependency fingerprint change

They must not:

1. mark a semantic slice as `passed`
2. write finding content
3. choose finding severity
4. decide `blocked` versus `fix_required`
5. decide whether the candidate is good enough to pass
6. write `_check_result/unit/{unit}.md`

## 10. Process Validation

When a command writes or consumes a supported unit process file, it must rebuild the current snapshot from current bound truth and compare it against the stored fields exactly.

At minimum, validation must rebuild:

1. `truth_layer_ref`, `truth_file_ref`, `truth_version_ref`, and `truth_fingerprint` for check and verify files
2. `spec_file_ref`, `spec_version_ref`, and `spec_fingerprint` for active plan files
3. `unit_appendix_snapshot` from explicitly referenced unit appendix files
4. `unit_snapshot` from current unit `unit_refs`
5. `rule_snapshot` from current unit `rule_refs`
6. `acceptance_item_set` from the current unit truth for check and verify files
7. `acceptance_item_plan_coverage` against the current candidate acceptance item ids for active plan files

Tool-backed validation rule:

1. when deterministic snapshot validation tooling is available, process-file writeback and process-file consumption must use it as the authoritative validation step
2. for current unit process files, the required validation command is:

```text
specflow/tooling/bin/specflowctl-<os>-<arch> snapshot validate-process --repo-root <repo-root> --object-type unit --object <unit> --process check|plan|verify
```

3. `plan` validates only `docs/specs/_plans/active/{unit}.md`
4. `docs/specs/_plans/draft/{unit}.md` is not a downstream-consumable handoff file
5. `_check_work` is validated by `specflowctl process check-work-validate`, not by `snapshot validate-process`
6. `snapshot validate-process` remains limited to downstream-consumable `check`, `plan`, and `verify` files
7. a command that writes a covered downstream process file must run the matching validation command after writeback and before reporting a pass gate, active handoff, verification pass, or lifecycle advance
8. a command that consumes a covered downstream process file must run the matching validation command before treating that file as current and consumable
9. if the required validation tooling is missing, unsupported for the target process kind, stale, or fails to execute, the command must report that authoritative process validation is unavailable and must not claim lifecycle progression from that process file
10. manual hash output, shell checksum output, editor display, conversation-derived values, and temporary script results are diagnostic only; they must not replace the tooling result for lifecycle progression

If any required field differs after applying only command-owned exceptions, the process file is invalid for downstream use.
