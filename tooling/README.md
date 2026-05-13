# SpecFlow Tooling

This directory contains the standalone Go CLI that performs deterministic governance actions for `specFlow`.

The tooling layer exists only for fixed execution work whose meaning is already constrained by governance rules.

## Build

`specflow/tooling/bin/` is a local binary cache.
It is ignored by git and must not be committed.

Rebuild local binaries from the repository root:

```bash
cd specflow/tooling
go run ./cmd/specflowctl build-release --repo-root ../..
```

Official platform binaries are GitHub Release assets.
Release tags use the tooling fingerprint form `specflow-tooling-<12-character-fingerprint>`.
The release workflow builds binaries from the tagged source and uploads the binaries plus `SHA256SUMS`.
The release is tied to the tooling input fingerprint, not to every source commit.

Download release binaries for the installed tooling source:

```bash
mkdir -p specflow/tooling/bin
tag="specflow-tooling-$(specflow/tooling/scripts/tooling_fingerprint.sh --short)"
base="https://github.com/Bingordinary/SpecFlow/releases/download/${tag}"
curl -fL -o specflow/tooling/bin/specflowctl-linux-amd64 "${base}/specflowctl-linux-amd64"
curl -fL -o specflow/tooling/bin/specflow-reader-linux-amd64 "${base}/specflow-reader-linux-amd64"
curl -fL -o specflow/tooling/bin/SHA256SUMS "${base}/SHA256SUMS"
chmod +x specflow/tooling/bin/specflowctl-linux-amd64 specflow/tooling/bin/specflow-reader-linux-amd64
(cd specflow/tooling/bin && sha256sum -c SHA256SUMS --ignore-missing)
```

The download commands replace existing files under `specflow/tooling/bin/`.
Rerun them after pulling a `specflow/` update only when the tooling fingerprint changed, the local binaries are missing, or an existing binary reports that it is stale.
Running them on every pull is safe, but it downloads the same Release again when the fingerprint did not change.

Replace `linux-amd64` with the target platform suffix:
`darwin-amd64`, `darwin-arm64`, `linux-amd64`, `linux-arm64`, `windows-amd64.exe`, or `windows-arm64.exe`.

## Governance Boundary

The tooling layer may:

1. collect
2. parse
3. validate
4. rebuild
5. compare
6. cleanup
7. sync
8. render read-only local views
9. maintain mechanical review run-state fields

The tooling layer must not:

1. invent new lifecycle semantics
2. replace command closure judgment
3. replace shared-boundary judgment
4. replace review severity or final conclusion judgment owned by the active review policy
5. become a second semantic source of truth
6. write reader-derived conclusions back into project files

`impact_sync` is a governance concept first.
The current CLI exposes only the deterministic pieces already justified by rules.
For shared-change reconciliation, the current mechanical entry remains `rule sync-impact`, but that entry must first compute `rule_sync` scope and exceptions and only then hand the fixed downstream object set to internal `impact_sync`.

`specflow-reader` is a read-only local view over current truth files.
It may parse `docs/specs/**`, build an in-memory graph, serve local HTML from `specflow/tooling/reader/web`, and refresh that view when truth files change.
It must not edit files, advance lifecycle state, or store semantic conclusions outside process memory.

## Current Command Surface

1. `init`
   - bootstrap framework-managed files
2. `doctor`
   - inspect installation and binary freshness health
3. `upgrade`
   - refresh framework-managed files and managed blocks
4. `build-release`
   - rebuild cross-platform binaries
5. `entry check`
   - inspect registered entry managed-block consistency
6. `entry sync`
   - sync registered project-side entry managed blocks from one explicit source
7. `registry validate`
   - validate `docs/project_standards/_registry.md`
8. `repository-mapping validate`
   - validate `docs/specs/repository_mapping.md` path rules against `_status.md` and declared rule files
9. `review collect-default-scope --flow <review_flow>`
   - collect the deterministic default scope for the explicit review flow
10. `review run-init --flow <review_flow>`
   - create or reuse the full-scope run-state file for the explicit review flow
11. `review run-validate --flow <review_flow>`
   - validate required run-state fields, timestamps, all fixed statuses including closed statuses, baseline slices, score state when present, and dynamic slice parent links
12. `review run-refresh --flow <review_flow>`
   - recompute slice input fingerprints for an open run-state file, mark changed `passed` slices as `stale`, and refresh `last_updated_at`
   - for `spec_flow_review`, the generated baseline includes `supporting_truth_lifecycle_convergence` so stable/candidate supporting truth paths are reviewed as an explicit cross-convergence slice
13. `review run-touch --flow <review_flow>`
   - refresh only `last_updated_at`
14. `command preflight`
   - mechanically verify a standard command's entry state from `_status.md` and required process snapshot validation results
   - this entry does not judge candidate completeness, evidence sufficiency, downgrade, or promotion readiness
15. `command close`
   - close one standard command from explicit standardized outcome flags
   - default mode is dry-run; `--apply` is required before `_status.md` or process files are changed
   - this entry validates fixed state combinations and process gates, but it does not choose outcomes, judge evidence, or repair contradictory caller input
16. `snapshot rebuild`
   - rebuild current process snapshots from bound truth
17. `snapshot validate-process`
   - compare one process file against rebuilt current truth
18. `process cleanup-fallback`
   - execute deterministic layered fallback cleanup for one unit or scenario
19. `process cleanup-success`
   - execute deterministic success cleanup for one unit or scenario
20. `status set-unit`
   - write one deterministic `unit` row in `_status.md` as a low-level status tool
21. `status set-object`
   - write one unified object row in `_status.md` as a low-level status tool
22. `rule sync-impact`
   - compute rule-specific scope, resolve rule-only exceptions into generic impact input, then execute deterministic downstream fallback for the fixed affected objects through internal `impact_sync`
   - when stable landing self-exemption is needed, the caller must pass both `--stable-landing-unit` and exact `--stable-landing-rule-refs`
   - when the same stable landing round retargeted candidate units to those stable landing rule refs, the caller must pass those units through `--retargeted-units` and must select both the old candidate Rule refs and the new stable Rule refs through exact `--rule-refs`
   - the caller may narrow the derived unit subset with `--units`, but at least one rule trigger input must still be provided through `--rule-refs` or `--rule-ids`; retargeted stable landing requires exact `--rule-refs`
23. `rule consumers`
   - read current-layer `unit` and `scenario` frontmatter `rule_refs` and print the consumers for one `rule_id` or exact `rule_ref`
24. `rule release-version`
   - publish an already-existing stable Rule version by retargeting current-layer consumers from `--from-ref` to `--to-ref`
   - candidate current-layer objects are rewritten directly
   - stable current-layer objects are auto-forked to candidate before their candidate `rule_refs` are rewritten
   - same-object stable appendices explicitly linked by the stable main Spec are retargeted into candidate appendices during the auto-fork, including Markdown link targets and direct same-object path literals
   - stale same-object candidate appendices are removed before the auto-fork writes current candidate appendices
   - stable unit forks additionally write `candidate_intent=change`; scenario forks do not write `candidate_intent`

## Reader Command Surface

`specflow-reader` is a separate binary from `specflowctl`.
It starts one local reader server directly and has no public subcommands:

```bash
cd specflow/tooling/bin
./specflow-reader-linux-amd64 --addr 127.0.0.1:17863
```

Rules:

1. running `specflow-reader` starts a local HTTP server and prints the URL.
2. the reader does not open a browser automatically.
3. when `--repo-root` is omitted, it defaults to `../../..` from the current working directory.
4. the server reads front-end files from `specflow/tooling/reader/web`.
5. the server reads current project truth from `docs/specs/**`.
6. each `/api/snapshot` request rebuilds the displayed snapshot from disk before returning data.
7. the server does not watch files and does not expose a server-sent event stream.
8. `/api/source` may return source text only from allowed truth and support files under the requested repository root.
9. the hidden build-fingerprint query command is reserved for freshness checks.

Reader front-end rules:

1. `specflow/tooling/reader/web` is the only runtime source for reader HTML, CSS, and JavaScript.
2. `specflow-reader` does not embed a fallback copy of the front-end.
3. editing front-end files does not require rebuilding `specflow-reader`; refresh the browser after the file change.
4. `doctor` reports missing required reader front-end files as installation failures.
5. the reader front-end supports Chinese and English interface text.
6. language switching affects only reader-owned UI text.
7. Spec document source text, file paths, object IDs, command names, and version values are displayed as source data and must not be translated by the front-end.
8. the selected reader language may be stored in browser-local state and must not be written into project files.
9. the front-end refresh button requests a new snapshot immediately.
10. the front-end also polls `/api/snapshot` on a fixed interval so open pages converge to the latest disk state without relying on filesystem events.
11. the Markdown document panel builds an in-memory side guide from the source document's Markdown headings, lets the reader open or close that guide locally, and uses it only for scrolling inside the currently opened document.
12. the Spec View front-end view shows current candidate main Specs, current stable main Specs for registered objects, and current stable rule Specs from the existing snapshot; it must not create a new review state or write any page conclusion back to project files.

## Review Run-State Commands

The `review run-*` commands require an explicit review flow:

1. `spec_flow_review`
2. `spec_flow_design_review`

They maintain only mechanical fields in:

```text
docs/specs/_governance_review/spec_flow_review.md
docs/specs/_governance_review/spec_flow_design_review.md
```

Rules:

1. timestamps are written from Go runtime UTC time using `YYYY-MM-DDTHH:MM:SSZ`
2. input fingerprints are computed from repository-relative input files
3. `run-refresh` may change `passed` slices to `stale` when inputs change or disappear
4. tooling must not change `pending`, `blocked`, or `skipped_not_in_scope` into a passing judgment
5. tooling may create and validate the `spec_flow_design_review` score-state skeleton
6. tooling must not write findings, severities, non-blocking optimizations, question scores, score basis, hard-blocker judgments, or final conclusions owned by the active review policy
   - `spec_flow_review` final conclusions are `pass | blocked`
   - `spec_flow_design_review` final conclusions are `pass | pass-with-optimization | blocked`
7. each review flow uses one fixed run-state file
8. when the fixed run-state file is missing, tooling creates the file for a new full-scope review
9. when a new full-scope review starts after a closed or invalid run-state file, tooling deletes the old fixed file before writing the new run state
10. `run-validate` checks structural validity only; a closed run-state file can validate successfully while still remaining unavailable for reuse
11. when the fixed run-state file is valid and open, `run-init` applies the owning review policy's age rule:
   - no more than two hours old: reuse automatically
   - for `spec_flow_review`, more than two hours and no more than 24 hours old: stop for a manual reuse-or-delete decision
   - for `spec_flow_review`, more than 24 hours and no more than seven days old: stop for a manual reuse-or-delete decision and recommend deleting the old run state and starting a new run
   - for `spec_flow_design_review`, more than two hours and no more than seven days old: stop for a manual reuse-or-delete decision
   - more than seven days old: delete as expired and create a new run state
12. after reusing an open run-state file, callers must run `review run-refresh` before continuing review work so changed inputs become stale slices instead of hidden drift
13. `review run-refresh` is the authoritative command for updating `input_fingerprint`; callers must not write manual hash output into run-state files
14. `spec_flow_review` baseline run state includes `supporting_truth_lifecycle_convergence` to force explicit review of fork, promote, cleanup, rule release, and tooling paths for stable and candidate supporting truth

## Command Preflight

`command preflight` is the mechanical entry check for standard lifecycle commands.
It validates only facts that are already fixed by governance rules:

1. the current `_status.md` row exists
2. the row's `Next Command` equals the requested command
3. every required input process file validates through `snapshot validate-process`

Usage:

```bash
./specflow/tooling/bin/specflowctl-linux-amd64 command preflight --command unit_plan --object-type unit --object assistant
./specflow/tooling/bin/specflowctl-linux-amd64 command preflight --command scenario_promote --object-type scenario --object checkout_flow
```

Output includes `preflight_result`, `validated_processes`, `failure_layer`, `recommended_next_command`, and `may_continue`.

Rules:

1. lifecycle commands that consume process files should run this before treating a gate, active plan, or verify result as usable
2. a failed preflight is not cleanup by itself; cleanup remains owned by command policy and `process cleanup-fallback`
3. manual hashes, shell checksums, editor display, and temporary scripts may diagnose a mismatch but must not replace this entry

## Tooling Input Set

The default `spec_flow_review` tooling review input set is:

1. the framework tooling policy and this README
2. the current tooling source input set listed below
3. the release helper script input set listed below
4. reader runtime files under `specflow/tooling/reader/web/**`

The current tooling source input set is:

1. `specflow/tooling/cmd/**/*.go`
2. `specflow/tooling/internal/**/*.go`
3. `specflow/tooling/go.mod`
4. `specflow/tooling/manifest.tsv`
5. `specflow/tooling/go.sum` when it exists

The release helper script input set is:

1. `specflow/tooling/scripts/tooling_fingerprint.sh`
2. `specflow/tooling/scripts/tooling_fingerprint.ps1`

The manifest is included because it controls which framework-managed and project-managed files `init`, `upgrade`, and `doctor` inspect or write.
Reader front-end files under `specflow/tooling/reader/web/**` are runtime files, not binary freshness inputs.
Release helper scripts are review inputs because they select release binaries for the installed tooling source.
They are not binary freshness inputs unless they change compiled binary behavior.

## Unified Status Table

`docs/specs/_status.md` now uses the unified object-state table:

1. `Object Type`
2. `Object`
3. `Stable`
4. `Candidate`
5. `Active Layer`
6. `Next Command`
7. `Notes`

Rules:

1. `command close` is the standard command-closing surface for lifecycle state progression
2. `status set-object` and `status set-unit` remain available only as low-level deterministic status write tools
3. standard command files must use `command close` instead of directly calling `status set-object`
4. `command close` may infer only the fixed state write defined by the command name and explicit `--outcome`; it must not infer the outcome itself
5. dry-run is the default; callers must pass `--apply` before `_status.md` or process files are changed
6. ordinary lifecycle progression must not bypass `command close` by editing `_status.md` manually or by using `status set-object` / `status set-unit`

## Command Close

`command close` moves one unit or scenario command from its current `_status.md` row to the one legal next state for an explicit outcome.
It is deterministic state machinery, not a semantic judge.

Required flags:

1. `--command`
2. `--object-type`
3. `--object`
4. `--outcome`

Optional flags:

1. `--reason`
2. `--failure-layer`
3. `--candidate-intent`
4. `--notes`
5. `--stable-before`
6. `--apply`

Execution rules:

1. without `--apply`, the command prints `status_before`, `status_after`, `input_validation_action`, `validation_action`, and `cleanup_action` without changing files
2. with `--apply`, the command writes `_status.md` and executes the success or fallback cleanup required by the fixed transition table
3. the current `_status.md` `Next Command` must equal `--command`, except for creation commands that register a missing object row
4. pass outcomes for check, plan, and verify gates validate the required process file before status progression
5. controlled stable-verify outcomes require the matching `--candidate-intent`
6. promotion recovery requires `--stable-before yes|no`
7. generic `truth_fallback` outcomes require an explicit `--reason`, because the command result owns the fallback reason code
8. `unit_plan` `truth_fallback` requires `--reason truth_incomplete`
9. for commands that consume current process files, non-fallback close outcomes run `command preflight` internally before status progression, cleanup, or success reporting
10. fallback and recovery close outcomes do not require currently valid input process files, because their purpose is to move the object back to the smallest legal recovery point
11. `input_validation_action`, `input_validated_processes`, and `input_validation_mismatches` report the internal close-time preflight result separately from output process validation
12. `command_close_result` is `dry_run`, `applied`, or `failed`; `failed` means the close operation returned an error, even when the caller passed `--apply`

## Usage Examples

Run ordinary governance commands from the repository root using the matching platform binary under `specflow/tooling/bin/`.
For normal use, download the matching `specflowctl-*` and `specflow-reader-*` files from the GitHub Release for the installed tooling fingerprint.
For local tooling development, rebuild them with `build-release`.

When developing the tooling itself, do not assume that ordinary commands may run through `go run`.
The freshness gate requires an embedded build fingerprint for ordinary governance actions.
The supported `go run` recovery and inspection surface remains:

1. `build-release`
2. `doctor`
3. `help`
4. the internal build-fingerprint query command

Examples:

```bash
./specflow/tooling/bin/specflowctl-linux-amd64 doctor
./specflow/tooling/bin/specflow-reader-linux-amd64 --repo-root . --addr 127.0.0.1:17863
./specflow/tooling/bin/specflowctl-linux-amd64 review collect-default-scope --flow spec_flow_review
./specflow/tooling/bin/specflowctl-linux-amd64 review collect-default-scope --flow spec_flow_design_review
./specflow/tooling/bin/specflowctl-linux-amd64 review run-init --flow spec_flow_review
./specflow/tooling/bin/specflowctl-linux-amd64 review run-init --flow spec_flow_design_review
./specflow/tooling/bin/specflowctl-linux-amd64 review run-validate --flow spec_flow_review
./specflow/tooling/bin/specflowctl-linux-amd64 review run-refresh --flow spec_flow_design_review
./specflow/tooling/bin/specflowctl-linux-amd64 review run-touch --flow spec_flow_design_review
./specflow/tooling/bin/specflowctl-linux-amd64 command preflight --command unit_impl --object-type unit --object ai
./specflow/tooling/bin/specflowctl-linux-amd64 command close --command unit_stable_verify --object-type unit --object ai --outcome controlled_repair_required --candidate-intent repair
./specflow/tooling/bin/specflowctl-linux-amd64 command close --command unit_stable_verify --object-type unit --object ai --outcome controlled_repair_required --candidate-intent repair --apply
./specflow/tooling/bin/specflowctl-linux-amd64 snapshot rebuild --object-type unit --object ai
./specflow/tooling/bin/specflowctl-linux-amd64 snapshot validate-process --object-type scenario --object task_execution --process verify
./specflow/tooling/bin/specflowctl-linux-amd64 process cleanup-fallback --object-type unit --object ai --from-command unit_promote --reason evidence_incomplete --failure-layer evidence_layer
./specflow/tooling/bin/specflowctl-linux-amd64 status set-object --type scenario --object task_execution --stable yes --candidate no --active-layer stable --next-command scenario_fork
./specflow/tooling/bin/specflowctl-linux-amd64 rule sync-impact --rule-refs c_b_rule_app_config_topology@0.2.0 --units ai
./specflow/tooling/bin/specflowctl-linux-amd64 rule sync-impact --rule-refs c_b_rule_runtime_model@0.3.0,s_b_rule_runtime_model@0.3.0 --stable-landing-unit skill --stable-landing-rule-refs s_b_rule_runtime_model@0.3.0 --retargeted-units agent
./specflow/tooling/bin/specflowctl-linux-amd64 rule consumers --rule-ref s_b_rule_runtime_model@0.4.0
./specflow/tooling/bin/specflowctl-linux-amd64 rule release-version --rule-id b_rule_runtime_model --from-ref s_b_rule_runtime_model@0.3.0 --to-ref s_b_rule_runtime_model@0.4.0
```

## Freshness Rule

Compiled binaries under `specflow/tooling/bin/` are local cache files.
They must fail closed when the embedded tooling fingerprint no longer matches current source.

The local development recovery path is:

```bash
cd specflow/tooling
go run ./cmd/specflowctl build-release --repo-root ../..
```

The normal user recovery path is to download the matching release binaries again for the installed tooling fingerprint.

The minimal stale-binary recovery and inspection surface remains:

1. `build-release`
2. `doctor`
3. `help`
4. the internal build-fingerprint query command
