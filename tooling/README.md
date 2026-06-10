# SpecFlow Tooling

This directory contains the standalone Go CLI that performs deterministic governance actions for `specFlow`.

The tooling layer exists only for fixed execution work whose meaning is already constrained by governance rules.
It validates Context Card process evidence mechanically, including independent reviewer receipt fields for advancing gates, but it does not judge business semantics.

In tooling contracts, `<tooling-root>` means `tooling/` in a `source_repo` layout and `specflow/tooling/` in an `installed_project` layout.
Installed-project usage examples below continue to show `specflow/tooling/...` directly.

## Build

`<tooling-root>/bin/` is a local binary cache.
It is ignored by git and must not be committed.

Installed-project rebuild example from the repository root:

```bash
cd specflow/tooling
go run ./cmd/specflowctl build-release --repo-root ../..
```

Source-repository rebuild example from the repository root:

```bash
cd tooling
go run ./cmd/specflowctl build-release --repo-root ..
```

Official platform binaries are GitHub Release assets.
Release tags use the tooling fingerprint form `specflow-tooling-<12-character-fingerprint>`.
The release workflow builds binaries from the tagged source and uploads the binaries plus `SHA256SUMS`.
The release is tied to the tooling input fingerprint, not to every source commit.
The fingerprint includes Go command code, Go internal code, and required tooling metadata.

Pull the repository and install the current platform binaries for the pulled tooling source:

```bash
specflow/tooling/scripts/pull_with_release.sh
```

PowerShell:

```powershell
.\specflow\tooling\scripts\pull_with_release.ps1
```

The script runs a fast-forward pull, computes the current tooling fingerprint, and downloads the current platform's `specflowctl`, `specflow-reader`, and `SHA256SUMS` only when the local binaries are missing, stale, or missing checksums.

Push the current branch and publish a tooling release when the current `main` fingerprint has no release tag:

```bash
specflow/tooling/scripts/push_with_release.sh
```

PowerShell:

```powershell
.\specflow\tooling\scripts\push_with_release.ps1
```

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
9. maintain mechanical review slice work-state fields when the adopting owner defines the state carrier and stale rules
10. maintain mechanical review run-state fields
11. maintain mechanical unit-check checklist fields
12. generate independent evaluation request files without writing reviewer conclusions

The tooling layer must not:

1. invent new lifecycle semantics
2. replace command closure judgment
3. replace shared-boundary judgment
4. replace review severity or final conclusion judgment owned by the active review policy
5. become a second semantic source of truth
6. write reader-derived conclusions back into project files
7. claim that a request file proves reviewer isolation

Slice work-state tooling for review run-state follows `specflow/framework/slice_work_state_protocol.md`.
The protocol defines generic carrier, slice, stale, and tooling-boundary standards for adopting review flows.
Each review policy still owns its own adoption rules, carrier paths, slice catalog, closure criteria, and final conclusions.
The CLI must not infer adoption or create a new durable state carrier from the protocol alone.

`impact_sync` is a governance concept first.
The current CLI exposes only the deterministic pieces already justified by rules.
For shared-change reconciliation, the current mechanical entry remains `rule sync-impact`, but that entry must first compute `rule_sync` scope and exceptions and only then hand the fixed downstream object set to internal `impact_sync`.

`specflow-reader` is a read-only local view over current truth files.
It may parse `docs/specs/**`, build an in-memory graph, serve local HTML from `<tooling-root>/reader/web`, and refresh that view when truth files change.
It must not edit files, advance lifecycle state, or store semantic conclusions outside process memory.

## Current Command Surface

1. `init`
   - bootstrap framework-managed files
2. `doctor`
   - inspect installation and binary freshness health
3. `build-release`
   - rebuild cross-platform binaries
4. `entry check`
   - inspect registered entry managed-block consistency
5. `entry sync`
   - sync registered entry managed blocks from one explicit source
   - in `installed_project`, registered entries resolve at the repository root
   - in `source_repo`, registered entries resolve under `templates/`
6. `evaluation request`
   - generate `docs/specs/_independent_evaluation/requests/unit/{unit}/{reviewer_pack}.md`
   - standard gate requests validate the target process artifact without requiring the not-yet-written independent receipt
   - freshness requests are allowed only for `text_drift` with `evidence_reuse: pending_review`
   - this entry does not create a reviewer session, write receipt fields, or advance lifecycle state
7. `repository-mapping validate`
   - validate `docs/specs/repository_mapping.md` path rules against `_status.md` and declared rule files
8. `review collect-default-scope --flow <review_flow> --layout auto|installed|source`
   - collect the deterministic default scope for the explicit review flow
9. `review run-init --flow <review_flow> --layout auto|installed|source`
   - create or reuse the full-scope run-state file for the explicit review flow
10. `review run-validate --flow <review_flow> --layout auto|installed|source`
   - validate required run-state fields, timestamps, all fixed statuses including closed statuses, baseline slices, score state when present, and dynamic slice parent links
11. `review run-refresh --flow <review_flow> --layout auto|installed|source`
   - recompute slice input fingerprints for an open run-state file, mark changed `passed` slices as `stale`, and refresh `last_updated_at`
   - for `spec_flow_review`, the generated baseline includes `supporting_truth_lifecycle_convergence` so stable/candidate supporting truth paths are reviewed as an explicit cross-convergence slice
12. `review run-touch --flow <review_flow> --layout auto|installed|source`
   - refresh only `last_updated_at`
13. `command preflight`
   - mechanically verify a standard command's entry state from `_status.md` and required process snapshot validation results
   - this entry does not judge candidate completeness, evidence sufficiency, downgrade, or promotion readiness
14. `command close`
   - close one standard command from explicit standardized outcome flags
   - default mode is dry-run; `--apply` is required before `_status.md` or process files are changed
   - this entry validates fixed state combinations and process gates, but it does not choose outcomes, judge evidence, or repair contradictory caller input
15. `snapshot rebuild`
   - rebuild current process snapshots from bound truth
16. `snapshot validate-process`
   - compare one process file against rebuilt current truth
17. `process cleanup-fallback`
   - execute deterministic layered fallback cleanup for one unit
18. `process cleanup-success`
   - execute deterministic success cleanup for one unit
   - `unit_promote` cleanup requires the stable promotion summary at `docs/specs/_verify_result/stable/unit/{unit}.md` before candidate verify evidence or candidate truth files are deleted
   - if `command close --command unit_promote --outcome promoted --apply` reports `status_updated: true` and then a success-cleanup failure, fix the filesystem blocker and rerun `process cleanup-success --object-type unit --object <unit> --mode unit_promote`; do not rerun `unit_promote` close because status already points to `unit_fork`
19. `process check-work-init`
	- create or reuse `docs/specs/_check_work/unit/{unit}.md` for `unit_check`
	- writes only the baseline checklist skeleton, timestamps, truth fingerprint, and input fingerprints
20. `process check-work-validate`
	- validate `unit_check` checklist shape, legal statuses, and repository-relative input paths
21. `process check-work-refresh`
	- recompute checklist fingerprints, mark changed `clear` items as `stale`, and update `last_updated_at`
22. `process check-work-touch`
	- refresh only `last_updated_at` for a valid `unit_check` checklist file
	- the `process check-work-*` commands do not adopt `slice_work_state_protocol.md`; they only maintain an optional unit-check checklist
23. `status set-unit`
   - write one deterministic `unit` row in `_status.md` as a low-level status tool
24. `status set-object`
   - write one unified object row in `_status.md` as a low-level status tool
25. `rule sync-impact`
   - compute rule-specific scope, resolve rule-only exceptions into generic impact input, then execute deterministic downstream fallback for the fixed affected objects through internal `impact_sync`
   - when a rule-governance topology round deleted an exact Rule ref only after proving it has no current-layer unit consumers, the caller may pass that ref through `--deleted-rule-refs`; the command verifies the ref is absent from rule files and current-layer unit `rule_refs`, then reports a no-impact result with no unit fallback
   - when stable landing self-exemption is needed, the caller must pass both `--stable-landing-unit` and exact `--stable-landing-rule-refs`
   - when the same stable landing round retargeted candidate units to those stable landing rule refs, the caller must pass those units through `--retargeted-units` and must select both the old candidate Rule refs and the new stable Rule refs through exact `--rule-refs`
   - the caller may narrow the derived unit subset with `--units`, but at least one rule trigger input must still be provided through `--rule-refs`, `--rule-ids`, or `--deleted-rule-refs`; retargeted stable landing requires exact `--rule-refs`
26. `rule consumers`
   - read current-layer `unit` frontmatter `rule_refs` and print the consumers for one `rule_id` or exact `rule_ref`
27. `rule release-version`
   - publish an already-existing stable Rule version by retargeting current-layer consumers from `--from-ref` to `--to-ref`
   - candidate current-layer objects are rewritten directly
   - stable current-layer objects are auto-forked to candidate only when their current `_status.md` `Next Command` is `unit_fork`
   - same-object stable appendices owned by path and appendix frontmatter are retargeted into candidate appendices during the auto-fork, including Markdown link targets and direct same-object path literals
   - stale same-object candidate appendices are removed before the auto-fork writes current candidate appendices
   - stable unit forks write `candidate_intent=change`; a current effective stable-verify `controlled_repair_required` result fails closed before mutation, while `controlled_change_required` allows the change fork
   - stable auto-fork status routing and process cleanup reuse the `unit_fork` command close contract
28. `unit release-version`
   - publish an already-existing stable unit version by retargeting current-layer `unit_refs` from `--from-ref` to `--to-ref`
   - candidate current-layer units are rewritten directly, unsafe current-round check checklist, check, plan, and verify process files are removed when present, and the unit is routed to `unit_check`
   - stable current-layer units are not rewritten; stale stable-verify evidence is removed when present and the unit is routed to `unit_stable_verify`
   - when no current-layer unit still uses the old stable unit ref, the command reports a no-op result
29. `relation candidates`
   - compute the current candidate advancement relation graph from explicit refs
   - print `relation_result`, `ready_candidates`, `blocked_candidates`, `candidate_cycles`, and `diagnostics`
30. `relation candidate-preflight`
   - check whether one current candidate unit is in the ready set
   - print the same relation fields narrowed to the requested object and fail when the target is blocked
31. `context collect`
   - collect the required context pack for a lifecycle command
   - `context collect --flow lifecycle --command <cmd> --object <obj>` collects the minimum durable truth inputs needed before entering the named lifecycle Context Card
32. `validate write`
   - validate write permission for a file path under the current lifecycle phase
   - `validate write --path <path> --phase <phase> [--unit <unit>]` checks whether the executor may write the given path under the active lifecycle constraints recorded in `_status.md`

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
4. the server reads front-end files from `<tooling-root>/reader/web`.
5. the server reads current project truth from `docs/specs/**`.
6. each `/api/snapshot` request rebuilds the displayed snapshot from disk before returning data.
7. the server does not watch files and does not expose a server-sent event stream.
8. `/api/source` may return source text only from allowed truth and support files under the requested repository root.
9. `/api/source-diff` returns diff hunks between the candidate and stable versions of a Spec document. Request parameter: `path` (repo-relative path to the Spec file). Returns a list of hunks with added/removed lines. Implemented at `/tooling/internal/reader/server.go` and consumed by the front-end diff panel.
10. the hidden build-fingerprint query command is reserved for freshness checks.
11. the snapshot includes `candidate_relations`, which is rebuilt from the same read-only relation calculation used by `specflowctl relation`.

Reader front-end rules:

1. `<tooling-root>/reader/web` is the only runtime source for reader HTML, CSS, and JavaScript.
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
12. the Spec View front-end view shows current candidate main Specs and candidate appendices as candidate truth, shows stable baseline truth separately for active candidate units, and shows current stable main Specs plus stable rule Specs from the existing snapshot; it must not create a new review state or write any page conclusion back to project files.

## Review Run-State Commands

The `review run-*` commands require an explicit review flow:

1. `spec_flow_review`
2. `spec_flow_design_review`

They also accept `--layout auto|installed|source`.

Layout meanings:

1. `installed`
   - resolves to `installed_project`
   - reads framework inputs from `specflow/framework/`
   - reads templates from `specflow/templates/`
   - reads tooling from `specflow/tooling/`
   - reviews real project-instance compatibility under `docs/specs/`
2. `source`
   - resolves to `source_repo`
   - reads framework inputs from local `framework/`
   - reads templates from local `templates/`
   - reads tooling from local `tooling/`
   - reviews template bootstrap compatibility under `templates/docs/specs/` and does not require real project-instance `docs/specs/`
3. `auto`
   - detects one of the two layouts
   - stops on ambiguous detection instead of choosing silently

They maintain only mechanical fields in:

```text
docs/specs/_governance_review/spec_flow_review.md
docs/specs/_governance_review/spec_flow_design_review.md
```

Rules:

1. timestamps are written from Go runtime UTC time using `YYYY-MM-DDTHH:MM:SSZ`
2. run-state files record `review_layout` as `installed_project` or `source_repo`
3. input fingerprints are computed from repository-relative input files
4. `run-refresh` may change `passed` slices to `stale` when inputs change or disappear
5. tooling must not change `pending`, `blocked`, or `skipped_not_in_scope` into a passing judgment
6. tooling may create and validate the `spec_flow_design_review` score-state skeleton
7. tooling must not write findings, severities, non-blocking optimizations, question scores, score basis, hard-blocker judgments, or final conclusions owned by the active review policy
   - `spec_flow_review` final conclusions are `pass | blocked`
   - `spec_flow_design_review` final conclusions are `pass | pass-with-optimization | blocked`
8. each review flow uses one fixed run-state file
9. when the fixed run-state file is missing, tooling creates the file for a new full-scope review
10. when a new full-scope review starts after a closed or invalid run-state file, tooling deletes the old fixed file before writing the new run state
11. `run-validate` checks structural validity only; a closed run-state file can validate successfully while still remaining unavailable for reuse
12. when the fixed run-state file is valid and open, `run-init` applies the owning review policy's age rule:
   - no more than two hours old: reuse automatically
   - for `spec_flow_review`, more than two hours and no more than 24 hours old: stop for a manual reuse-or-delete decision
   - for `spec_flow_review`, more than 24 hours and no more than seven days old: stop for a manual reuse-or-delete decision and recommend deleting the old run state and starting a new run
   - for `spec_flow_design_review`, more than two hours and no more than seven days old: stop for a manual reuse-or-delete decision
   - more than seven days old: delete as expired and create a new run state
13. after reusing an open run-state file, callers must run `review run-refresh` before continuing review work so changed inputs become stale slices instead of hidden drift
14. `review run-refresh` is the authoritative command for updating `input_fingerprint`; callers must not write manual hash output into run-state files
15. an explicit `--layout` that conflicts with an existing open run-state file's `review_layout` fails instead of rewriting that file
16. `spec_flow_review` baseline run state includes `supporting_truth_lifecycle_convergence` to force explicit review of fork, promote, cleanup, rule release, and tooling paths for stable and candidate supporting truth
17. review run-state slice fields follow the generic protocol only through the adoption rules in the active review policy

## Command Preflight

`command preflight` is the mechanical entry check for standard lifecycle commands.
It validates only facts that are already fixed by governance rules:

1. the current `_status.md` row exists
2. the row's `Next Command` equals the requested command
3. every required input process file validates through `snapshot validate-process`

Usage:

```bash
./specflow/tooling/bin/specflowctl-linux-amd64 command preflight --command unit_verify --object-type unit --object demo
```

Output includes `preflight_result`, `validated_processes`, `failure_layer`, `recommended_next_command`, and `may_continue`.

Rules:

1. lifecycle commands that consume process files should run this before treating a gate, active plan, or verify result as usable
2. a failed preflight is not cleanup by itself; cleanup remains owned by lifecycle recovery and `process cleanup-fallback`
3. `unit_promote` validates both the active plan and verify evidence so retirement target drift cannot be hidden between verification and promotion
4. manual hashes, shell checksums, editor display, and temporary scripts may diagnose a mismatch but must not replace this entry

## Relation Commands

`relation candidates` and `relation candidate-preflight` are read-only mechanical relation commands.
They calculate candidate advancement order from already-written explicit references.

Recognized reference inputs are current candidate unit main Specs, same-layer non-evidence appendix files owned by those units, `unit_refs`, `rule_refs`, Markdown `.md` links, and version refs such as `c_unit_trace@0.3.0` or `c_b_rule_runtime_model@0.4.0`.
The candidate appendix named by `evidence_appendix_ref` is reported as reference-only evidence and does not block advancement.
Natural-language prose alone is not a relation input.

Usage:

```bash
./specflow/tooling/bin/specflowctl-linux-amd64 relation candidates
./specflow/tooling/bin/specflowctl-linux-amd64 relation candidate-preflight --object trace
```

Rules:

1. the commands never edit project files
2. the commands never judge candidate completeness, evidence quality, or promotion readiness
3. `candidate-preflight` must fail when the requested candidate is blocked by another current candidate unit, a candidate Rule, or a candidate progression cycle
4. the reader todo panel may use the same result to group candidates as ready, blocked, or cycle
5. the reader todo panel must show `unit_advance:{unit}` only for ready candidates whose recorded next command is `unit_check` or `unit_verify`; promotion-ready candidates must show the explicit `unit_promote:{unit}` command instead

## Tooling Input Set

The default `spec_flow_review` tooling review input set is:

1. the framework tooling policy and this README
2. the current tooling source input set listed below
3. the tooling helper script input set listed below
4. reader runtime files under `<tooling-root>/reader/web/**`

The current tooling source input set is:

1. `<tooling-root>/cmd/**/*.go`
2. `<tooling-root>/internal/**/*.go`
3. `<tooling-root>/go.mod`
4. `<tooling-root>/manifest.tsv`
5. `<tooling-root>/go.sum` when it exists

The tooling helper script input set is every regular file under:

```text
<tooling-root>/scripts/**
```

This includes install, pull-with-release, push-with-release, build-release, and tooling-fingerprint scripts.

The manifest is included because it controls which framework-managed and project-managed files `init` and `doctor` inspect or write.
Reader front-end files under `<tooling-root>/reader/web/**` are runtime files, not binary freshness inputs.
Tooling helper scripts are review inputs because they rebuild or select binaries for the installed tooling source.
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
3. lifecycle Context Cards must use `command close` instead of directly calling `status set-object`
4. `command close` may infer only the fixed state write defined by the command name and explicit `--outcome`; it must not infer the outcome itself
5. dry-run is the default; callers must pass `--apply` before `_status.md` or process files are changed
6. ordinary lifecycle progression must not bypass `command close` by editing `_status.md` manually or by using `status set-object` / `status set-unit`

## Command Close

`command close` moves one unit command from its current `_status.md` row to the one legal next state for an explicit outcome.
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
3. the current `_status.md` `Next Command` must equal `--command`, except for:
   - creation commands that register a missing object row
   - `unit_check` when `Next Command` is `unit_verify` and `Notes` contains `pending_impl` (re-validation during implementation phase)
   - `unit_stable_verify` when `Active Layer` is `stable` and `Next Command` is not `unit_promote` (per status.md allows semantics — stable verify is a check command, not a progression command)
4. pass outcomes for check, plan, and verify gates validate the required process file before status progression
5. controlled stable-verify outcomes require the matching `--candidate-intent`
6. promotion recovery requires `--stable-before yes|no`
7. generic `truth_fallback` outcomes require an explicit `--reason`, because the command result owns the fallback reason code
8. `unit_plan` is a removed command; requests using `unit_plan` as `--command` must be rejected
9. fallback reasons must use the canonical recovery codes: `truth_drift`, `binding_drift`, `baseline_drift`, `rule_drift`, `truth_incomplete`, `gate_missing`, `evidence_incomplete`, or `stable_verify_invalid`
10. each fallback reason is accepted only with its recovery-defined failure layer; `gate_missing` belongs to `gate_layer`
11. `plan_drift` and `implementation_deviation` are impact classifications handled agent-internally; they are not valid fallback cleanup reasons and must not be passed to process cleanup-fallback or fallback close
12. for commands that consume current process files, non-fallback close outcomes run `command preflight` internally before status progression, cleanup, or success reporting
13. fallback and recovery close outcomes do not require currently valid input process files, because their purpose is to move the object back to the smallest legal recovery point
14. `input_validation_action`, `input_validated_processes`, and `input_validation_mismatches` report the internal close-time preflight result separately from output process validation
15. `command_close_result` is `dry_run`, `applied`, or `failed`; `failed` means the close operation returned an error, even when the caller passed `--apply`

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
./specflow/tooling/bin/specflowctl-linux-amd64 review collect-default-scope --flow spec_flow_review --layout auto
./specflow/tooling/bin/specflowctl-linux-amd64 review collect-default-scope --flow spec_flow_design_review --layout auto
./specflow/tooling/bin/specflowctl-linux-amd64 review run-init --flow spec_flow_review --layout auto
./specflow/tooling/bin/specflowctl-linux-amd64 review run-init --flow spec_flow_design_review --layout auto
./specflow/tooling/bin/specflowctl-linux-amd64 review run-validate --flow spec_flow_review --layout auto
./specflow/tooling/bin/specflowctl-linux-amd64 review run-refresh --flow spec_flow_design_review --layout auto
./specflow/tooling/bin/specflowctl-linux-amd64 review run-touch --flow spec_flow_design_review --layout auto
./specflow/tooling/bin/specflowctl-linux-amd64 command preflight --command unit_verify --object-type unit --object ai
./specflow/tooling/bin/specflowctl-linux-amd64 command close --command unit_stable_verify --object-type unit --object ai --outcome controlled_repair_required --candidate-intent repair
./specflow/tooling/bin/specflowctl-linux-amd64 command close --command unit_stable_verify --object-type unit --object ai --outcome controlled_repair_required --candidate-intent repair --apply
./specflow/tooling/bin/specflowctl-linux-amd64 snapshot rebuild --object-type unit --object ai
./specflow/tooling/bin/specflowctl-linux-amd64 snapshot validate-process --object-type unit --object ai --process verify
./specflow/tooling/bin/specflowctl-linux-amd64 snapshot validate-process --object-type unit --object ai --process stable_verify
./specflow/tooling/bin/specflowctl-linux-amd64 process cleanup-fallback --object-type unit --object ai --from-command unit_promote --reason evidence_incomplete --failure-layer evidence_layer
./specflow/tooling/bin/specflowctl-linux-amd64 status set-object --type unit --object ai --stable yes --candidate no --active-layer stable --next-command unit_fork
./specflow/tooling/bin/specflowctl-linux-amd64 rule sync-impact --rule-refs c_b_rule_app_config_topology@0.2.0 --units ai
./specflow/tooling/bin/specflowctl-linux-amd64 rule sync-impact --deleted-rule-refs c_b_rule_unused@0.1.0
./specflow/tooling/bin/specflowctl-linux-amd64 rule sync-impact --rule-refs c_b_rule_runtime_model@0.3.0,s_b_rule_runtime_model@0.3.0 --stable-landing-unit skill --stable-landing-rule-refs s_b_rule_runtime_model@0.3.0 --retargeted-units agent
./specflow/tooling/bin/specflowctl-linux-amd64 rule consumers --rule-ref s_b_rule_runtime_model@0.4.0
./specflow/tooling/bin/specflowctl-linux-amd64 rule release-version --rule-id b_rule_runtime_model --from-ref s_b_rule_runtime_model@0.3.0 --to-ref s_b_rule_runtime_model@0.4.0
./specflow/tooling/bin/specflowctl-linux-amd64 unit release-version --unit assistant --from-ref s_unit_assistant@0.8.0 --to-ref s_unit_assistant@0.9.0
```

## Freshness Rule

Compiled binaries under `<tooling-root>/bin/` are local cache files.
They must fail closed when the embedded tooling fingerprint no longer matches current source.
The fingerprint hashes tooling-root-relative keys such as `cmd/...`, `internal/...`, `go.mod`, and `manifest.tsv`, so identical tooling content has one fingerprint in both layouts.

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
