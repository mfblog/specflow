# Process Snapshot Contract

Process files record what a unit command checked in one round.

They are evidence, not behavior truth.

This file defines the supported process paths, process snapshots, unit-check checklist fields, freshness rules, and validation rules.

## 1. Supported Process Paths

Supported unit process paths:

1. unit check checklist: `docs/specs/_check_work/unit/{unit}.md`
2. check result: `docs/specs/_check_result/unit/{unit}.md`
3. verify result: `docs/specs/_verify_result/unit/{unit}.md`
4. stable promotion summary: `docs/specs/_verify_result/stable/unit/{unit}.md`
5. stable verify result: `docs/specs/_stable_verify_result/unit/{unit}.md`

Active plan (`docs/specs/_plans/active/{unit}.md`) and draft plan (`docs/specs/_plans/draft/{unit}.md`) are agent-internal artifacts. They are not SpecFlow process evidence and are not consumed by SpecFlow lifecycle gates.

No `scenario` process path is supported.

`_check_work` is a command-local checklist path for `unit_check`.
It is not a pass gate and is not consumed by downstream commands.
Downstream handoff commands consume only `_check_result` and `_verify_result` according to their command contracts.
`unit_stable_verify` close consumes only tool-valid `_stable_verify_result` for lifecycle advancement from stable verification.
Other process files may carry command-owned review records only when their owning command defines that adoption in its command file.

## 2. Common Fields

Check and verify process YAML must identify the unit and command gate:

```yaml
object_type: unit
object_ref: {unit}
gate: unit_check|unit_verify
decision: pass
allow_next: true|false
next_command: unit_verify|unit_promote|none
truth_layer_ref: candidate
truth_file_ref: docs/specs/units/candidate/c_unit_{unit}.md
truth_version_ref: c_unit_{unit}@x.y.z
truth_fingerprint: {fingerprint}
acceptance_behavior_fingerprint: {fingerprint}
unit_appendix_snapshot: none | list
unit_snapshot: none | list
rule_snapshot: none | list
evaluation_mode: independent
reviewer_result: pass
reviewer_context: minimal_context
review_input_refs: {reviewer_pack};{request_file};{durable_input_refs}
review_findings: none
human_decision_refs: none | list
```

`_check_result` and candidate `_verify_result` are consumable evidence only for advancing pass gates.
Non-advancing command outcomes such as `blocked` or `fix_required` must not be stored as these process snapshots.

Stable verify process YAML must identify stable truth and the current implementation-alignment decision:

```yaml
object_type: unit
object_ref: {unit}
gate: unit_stable_verify
decision: aligned|controlled_repair_required|controlled_change_required|small_repair_required|evidence_incomplete|truth_rejudge_required
allow_next: true|false
next_command: unit_fork|unit_stable_verify
blocking_summary: none|{summary}
coverage_summary: {summary}
truth_layer_ref: stable
truth_file_ref: docs/specs/units/stable/s_unit_{unit}.md
truth_version_ref: s_unit_{unit}@x.y.z
truth_fingerprint: {fingerprint}
acceptance_behavior_fingerprint: {fingerprint}
repository_mapping_snapshot: present
implementation_surface_refs: {refs}
evidence_refs: {refs}
evaluation_mode: independent
reviewer_result: pass
reviewer_context: minimal_context
review_input_refs: {reviewer_pack};{request_file};{durable_input_refs}
review_findings: none
human_decision_refs: none | list
```

Plan process YAML (defined in `docs/specs/_plans/active/{unit}.md`) is an agent-internal artifact. It is not SpecFlow process evidence and is not consumed by `unit_verify` or `unit_promote`.

When the agent creates an internal plan and chooses to reference it in verify evidence, the candidate verify process YAML may include plan-related fields:

```yaml
acceptance_item_evidence_matrix:
  - id: {acceptance_item_id}
    status: pass|fail|partial|not_checked|not_runnable_yet
    evidence_refs: {refs}
```

```yaml
acceptance_item_evidence_matrix:
  - id: {acceptance_item_id}
    status: pass|fail|partial|not_checked|not_runnable_yet
    evidence_refs: {refs}
    scope_verification:              # required when the acceptance item declares affects
      files:
        - path: {file_path}
          status: pass|fail|not_checked
          evidence_refs: {refs}
      appendices:
        - name: {appendix_name}
          status: pass|fail|not_checked
          evidence_refs: {refs}
      rules:
        - name: {rule_name}
          status: pass|fail|not_checked
          evidence_refs: {refs}
      dependencies:
        - name: {dependency_name}
          status: pass|fail|not_checked
          evidence_refs: {refs}
```

When the agent created an internal plan and chooses to reference it, these optional fields may also appear:

```yaml
active_plan_file_ref: docs/specs/_plans/active/{unit}.md | none
active_plan_fingerprint: {fingerprint} | none
retirement_evidence_matrix: none | list
package_delta_verification: none | list
```

Each `acceptance_item_evidence_matrix` item must include `id`, `status`, and `evidence_refs`.
For executable candidate acceptance items, promotion-ready verify evidence requires `status: pass` and non-empty durable `evidence_refs`.
Items whose current truth records `not_runnable_yet: yes` must use `status: not_runnable_yet`; they may use `evidence_refs: none`.
Generic test success, renamed files, new fields, or absent old strings are not sufficient by themselves for semantic replacement claims.

`retirement_evidence_matrix` is optional. When no plan exists or the plan has no retirement targets, it must be literal `none`.
When present, each item must include `id`, `result`, `mainline_dependency`, and `evidence_refs`.
Promotion-ready evidence requires `result: pass` and `mainline_dependency: not_required` for every retirement target.
Valid `result` values are `pass`, `fail`, and `not_checked`.
Valid `mainline_dependency` values are `not_required`, `still_required`, and `unknown`.
Tooling must not delete implementation code or judge whether a retained compatibility path is business-safe.

`package_delta_verification` is optional. When no plan exists, it must be literal `none`.
When present, it must contain exactly one item for each planned change scope id:

```yaml
package_delta_verification:
  - planned_change_scope_id: pcs.<slug>
    result: pass|fail|not_checked
    evidence_refs: {refs}
```

Advancing check (when run), verify, and stable verify files must include the independent evaluation receipt.
Tooling validates the receipt fields mechanically; it does not prove reviewer session isolation and does not judge whether the reviewer made a good semantic decision.
Independent evaluation request files under `docs/specs/_independent_evaluation/requests/**` are handoff instructions only.
They are not process snapshots, are not lifecycle evidence, and are not consumed by `command close`.

Receipt requirements:

1. `evaluation_mode` must be `independent`.
2. `reviewer_result` must be `pass` for advancing evidence.
3. `reviewer_context` must be `minimal_context`.
4. `review_findings` must be `none`.
5. `review_input_refs` must contain the reviewer pack name from `framework/core/independent_evaluation.md`, the generated request file path, and at least one durable input ref supplied to the reviewer.
6. `human_decision_refs` must be `none` or durable human-confirmation refs, not chat-only conclusions.

Freshness reuse fields are conditional. They are required only when deterministic validation reports `text_drift` and the process evidence is being reused instead of recreated:

```yaml
freshness_impact: text_drift
evidence_reuse: accepted
freshness_current_fingerprint: {current truth/spec fingerprint}
freshness_review_mode: independent
freshness_reviewer_result: pass
freshness_reviewer_context: minimal_context
freshness_review_input_refs: freshness_text_drift_reuse;{request_file};{durable_input_refs}
freshness_review_findings: none
```

`acceptance_behavior_fingerprint` is the normalized SHA-256 of the full formal acceptance item behavior fields: `id`, `target`, `verification_surface`, `implementation_surface`, `verification_method`, `pass_condition`, `not_runnable_yet`, and `not_runnable_yet_reason`.
Advancing check (when run), verify, and stable verify evidence must record it.
Evidence without this field uses the old snapshot schema and is not current valid advancing evidence until it is migrated or recreated.
It must not be treated as accepted `text_drift` reuse.

Tooling classifies freshness impact mechanically. It does not judge whether a changed paragraph preserves business meaning.
Text drift evidence reuse requires the freshness receipt above and independent reviewer confirmation using reviewer pack `freshness_text_drift_reuse`.

## 3. Dependency Snapshots

Candidate check (when run) and verify process files must record the current package snapshots:

```yaml
unit_appendix_snapshot: none
unit_snapshot: none
rule_snapshot: none
```

or lists.

`unit_appendix_snapshot` records the current-layer appendix files owned by the unit through appendix path and appendix frontmatter.

For an active candidate unit, every stable appendix `s_unit_{unit}_{name}.md` must have a corresponding candidate appendix `c_unit_{unit}_{name}.md`.
Candidate-only appendices are allowed.

`unit_snapshot` records stable unit dependencies resolved from current unit `unit_refs`.

`rule_snapshot` records stable global rules plus bound shared rules resolved from current unit `rule_refs`.
Stable global rules are included even when the unit has `rule_refs: none`.

Tooling must validate these snapshots against current truth.

## 4. Fallback Layers

Process validation failure maps to these layers:

1. truth mismatch -> `truth_layer`
2. check schema or gate evidence mismatch -> `gate_layer`
3. verify evidence mismatch -> `evidence_layer`
4. stable verify evidence mismatch -> `evidence_layer`, with `unit_stable_verify` as the restart command
5. unaccepted text drift -> `freshness_layer`, with no automatic status reroute

The legal fallback commands are:

1. `truth_layer` -> candidate truth repair or `unit_check` (optional)
2. `gate_layer` -> candidate truth repair or `unit_check` (optional)
3. `evidence_layer` -> `unit_verify`
4. `stable_verify` validation failure -> `unit_stable_verify`

`freshness_layer` is not a fallback layer. It means the process file may be reusable after independent freshness review, so tooling must not delete process files or advance `_status.md` from that result alone.

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
7. `acceptance_behavior_fingerprint`

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

## 9. Stable Verify Result

`docs/specs/_stable_verify_result/unit/{unit}.md` stores current evidence produced by `unit_stable_verify`.

It is:

1. a gate-bearing process file for advancing `unit_stable_verify` to `unit_fork`
2. an implementation-alignment claim against current stable truth
3. a compact evidence record, not a work-state protocol

It is not:

1. a stable promotion summary
2. behavior truth
3. a slice work-state file
4. a separate checklist file

Each stable verify result must record:

1. `object_type`, `object_ref`, `gate`, `decision`, `allow_next`, and `next_command`
2. `blocking_summary` and `coverage_summary`
3. stable truth ref, version ref, and fingerprint
4. `repository_mapping_snapshot`
5. `unit_appendix_snapshot`, `unit_snapshot`, and `rule_snapshot`
6. `acceptance_item_set`
7. `acceptance_item_evidence_matrix`
8. `implementation_surface_refs`
9. `evidence_refs`
10. independent evaluation receipt fields

For `decision: aligned`, every executable acceptance item must have evidence status `pass`.
Items marked `not_runnable_yet: yes` in stable truth must use evidence status `not_runnable_yet`.

`controlled_repair_required` and `controlled_change_required` may advance to `unit_fork` only when the matching stable verify result validates and the command close outcome matches the stored `decision`.

## 10. Unit Check Checklist

`docs/specs/_check_work/unit/{unit}.md` records resumable `unit_check` progress for one target candidate.

It is:

1. an optional process checklist file
2. a resume aid for the current `unit_check` round
3. a place to record checklist item status, input fingerprints, finding references, blocked reason, and next resume step

It is not:

1. a Spec
2. behavior truth
3. a downstream pass gate
4. a substitute for `_check_result/unit/{unit}.md`
5. a place for tooling to decide semantic pass, finding severity, or final conclusion

### 10.1 Required Run Fields

The checklist run table must record:

```yaml
work_flow: unit_check
work_id: YYYYMMDD-HHMMSS-unit_check-{unit}
object_type: unit
object_ref: {unit}
status: in_progress|blocked_on_finding|ready_for_final|closed_pass|closed_blocked|closed_fix_required
created_at: YYYY-MM-DDTHH:MM:SSZ
last_updated_at: YYYY-MM-DDTHH:MM:SSZ
truth_layer_ref: candidate
truth_file_ref: docs/specs/units/candidate/c_unit_{unit}.md
truth_version_ref: c_unit_{unit}@x.y.z
truth_fingerprint: {fingerprint}
checklist_table: present
finding_refs: none|{refs}
blocked_reason: none|{reason}
resume_next_step: {step}
```

`truth_fingerprint` uses the same normalized SHA-256 contract as Section 6.

### 10.2 Checklist Fields

The checklist table uses these fields:

```text
item_id
status
question
input_files
input_fingerprint
finding_refs
result_summary
```

Allowed `status` values:

1. `pending`
2. `clear`
3. `incomplete`
4. `blocked`
5. `stale`
6. `not_applicable`

`clear`, `incomplete`, and `blocked` are semantic statuses set by the executor.
Tooling may create `pending`, preserve executor-set statuses, and mechanically mark `clear` items as `stale` when inputs change.

### 10.3 Baseline Checklist

1. `goal_and_responsibility`
2. `dependency_truth_surface`
3. `main_flow_and_state`
4. `boundary_and_protocol`
5. `data_artifact_and_output`
6. `error_edge_and_gap`
7. `acceptance_and_testability`
8. `implementation_handoff`

### 10.4 Freshness And Stale Rules

Before a `unit_check` pass gate is written, the checklist file may be refreshed and validated if the executor used it during the round.
The pass gate still depends only on `_check_result` validation.

Refresh rules:

1. recompute the checklist truth fingerprint from the current target candidate
2. recompute each checklist item `input_fingerprint` from the current `input_files`
3. if an item is `clear` and its input fingerprint changes, mark that item `stale`
4. if an item is `clear` and one of its input files is missing, mark that item `stale`
5. update `last_updated_at`

If truth drift, binding drift, fallback cleanup, unit fork, unit promote, rule release, or project-instance migration invalidates the target candidate's prior check state, the old `_check_work` file must be deleted or marked unusable by the owning cleanup path.

### 10.5 Tooling Boundary

`specflowctl process check-work-*` commands may:

1. create the checklist skeleton
2. validate field presence, legal values, checklist table shape, object type, and repository-relative input paths
3. refresh timestamps
4. compute input fingerprints
5. mark stale checklist items caused by input fingerprint change

They must not:

1. mark a semantic item as `clear`, `incomplete`, or `blocked`
2. write finding content
3. choose finding severity
4. decide `blocked` versus `fix_required`
5. decide whether the candidate is good enough to pass
6. write `_check_result/unit/{unit}.md`

## 11. Process Validation

When a command writes or consumes a supported unit process file, it must rebuild the current snapshot from current bound truth and compare it against the stored fields exactly.

At minimum, validation must rebuild:

1. `truth_layer_ref`, `truth_file_ref`, `truth_version_ref`, and `truth_fingerprint` for check and verify files
2. `spec_file_ref`, `spec_version_ref`, and `spec_fingerprint` for active plan files
3. stable truth refs and fingerprints for stable verify files
4. `repository_mapping_snapshot` for stable verify files
5. `unit_appendix_snapshot` from current-layer unit appendix files owned by path and appendix frontmatter
6. `unit_snapshot` from current unit `unit_refs`
7. `rule_snapshot` from stable global rules and current unit `rule_refs`
8. `acceptance_item_set` from the current unit truth for check, verify, and stable verify files
9. `acceptance_item_plan_coverage` against the current candidate acceptance item ids for active plan files
10. `stable_candidate_diff_refs`, `implementation_gap_refs`, `planned_change_scope`, `package_constraint_review`, `package_constraint_refs`, `package_constraint_summary`, and `retirement_targets` shape, ids, target fields, package refs, and acceptance item refs for active plan files
11. `active_plan_file_ref` and `active_plan_fingerprint` for candidate verify files
12. `acceptance_item_evidence_matrix` status and evidence refs against the current acceptance item ids for verify and stable verify files
13. `retirement_evidence_matrix` against the current active plan retirement target ids for candidate verify files
14. `package_delta_verification` against the current active plan `planned_change_scope` ids for candidate verify files
15. independent evaluation receipt fields for check, active plan, verify, and stable verify files
16. acceptance behavior fingerprint and conditional freshness reuse receipt fields when text drift is being reused

Deterministic validation rule:

1. when deterministic snapshot validation tooling is available, process-file writeback and process-file consumption must use it as the deterministic validation step
2. for current unit process files, the required validation command is:

```text
<tooling-root>/bin/specflowctl-<os>-<arch> snapshot validate-process --repo-root <repo-root> --object-type unit --object <unit> --process check|plan|verify|stable_verify
```

3. `plan` validates only `docs/specs/_plans/active/{unit}.md`
4. `docs/specs/_plans/draft/{unit}.md` is not a downstream-consumable handoff file
5. `_check_work` is validated by `specflowctl process check-work-validate`, not by `snapshot validate-process`
6. `stable_verify` validates only `docs/specs/_stable_verify_result/unit/{unit}.md`
7. a command that writes a covered downstream process file must run the matching validation command after writeback and before reporting a pass gate, active handoff, verification pass, or lifecycle advance
8. a command that consumes a covered downstream process file must run the matching validation command before treating that file as current and consumable
9. if the required validation tooling is missing, unsupported for the target process kind, stale, or fails to execute, the command must report that deterministic process validation is unavailable and must not claim lifecycle progression from that process file
10. manual hash output, shell checksum output, editor display, conversation-derived values, and temporary script results are diagnostic only; they must not replace the tooling result for lifecycle progression

Freshness impact is defined by `framework/core/freshness.md`.
When only normalized text drift exists and freshness reuse is accepted, the process file remains valid for downstream use.
When freshness reuse is missing or rejected, validation reports `freshness_layer` without recommending lifecycle fallback.
When behavior, acceptance, dependency, or schema drift is found, existing fallback and recovery rules still apply.
