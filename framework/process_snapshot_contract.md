# Process Snapshot Contract

Process files record what a unit command checked in one round.

They are evidence, not behavior truth.

## 1. Supported Process Paths

Supported unit process paths:

1. check result: `docs/specs/_check_result/unit/{unit}.md`
2. active plan: `docs/specs/_plans/active/{unit}.md`
3. draft plan: `docs/specs/_plans/draft/{unit}.md`
4. verify result: `docs/specs/_verify_result/unit/{unit}.md`
5. stable promotion summary: `docs/specs/_verify_result/stable/unit/{unit}.md`

No `scenario` process path is supported.

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

## 9. Process Validation

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
5. a command that writes a covered process file must run the matching validation command after writeback and before reporting a pass gate, active handoff, verification pass, or lifecycle advance
6. a command that consumes a covered process file must run the matching validation command before treating that file as current and consumable
7. if the required validation tooling is missing, unsupported for the target process kind, stale, or fails to execute, the command must report that authoritative process validation is unavailable and must not claim lifecycle progression from that process file
8. manual hash output, shell checksum output, editor display, conversation-derived values, and temporary script results are diagnostic only; they must not replace the tooling result for lifecycle progression

If any required field differs after applying only command-owned exceptions, the process file is invalid for downstream use.
