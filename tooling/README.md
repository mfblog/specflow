# SpecFlow Tooling

This directory now contains the standalone Go CLI that serves as the core for deterministic governance actions.

Compiled binaries are placed under:

1. `specflow/tooling/bin/specflowctl-linux-amd64`
2. `specflow/tooling/bin/specflowctl-linux-arm64`
3. `specflow/tooling/bin/specflowctl-darwin-amd64`
4. `specflow/tooling/bin/specflowctl-darwin-arm64`
5. `specflow/tooling/bin/specflowctl-windows-amd64.exe`
6. `specflow/tooling/bin/specflowctl-windows-arm64.exe`

## Build

To rebuild those binaries from source, run from the repository root:

```bash
go run ./specflow/tooling/cmd/specflowctl build-release --repo-root .
```

## Current Command Surface

The current command surface intentionally covers only high-ROI deterministic actions:

1. `init`
   - installs files from `manifest.tsv`
2. `doctor`
   - checks whether the installed structure, hook, and current-platform binary are healthy
3. `upgrade`
   - refreshes framework-managed files and managed blocks
4. `build-release`
   - rebuilds the platform binaries into `specflow/tooling/bin/`
5. `entry check`
   - verifies managed-block consistency across registered entry files
6. `entry sync`
   - syncs registered entry-file managed blocks from one chosen source
7. `registry validate`
   - validates `docs/project_standards/_registry.md`
8. `review collect-default-scope`
   - collects the default deterministic file scope for `spec_flow_review`
9. `process cleanup-fallback`
   - applies command-defined fallback cleanup for candidate-chain process files
10. `snapshot rebuild`
   - rebuilds the current process snapshot inputs from formal truth files
11. `snapshot validate-process`
   - compares an existing process file snapshot against rebuilt current truth
12. `status set-module`
   - writes or creates one deterministic module row in `docs/specs/_status.md`
13. `process cleanup-success`
   - applies command-defined success-path cleanup for `spec_fork` and `cand_promote`
14. `shared sync-impact`
   - reconciles shared-truth impact by comparing current bindings and process snapshots
   - invalidates candidate modules when shared truth drifted
   - reroutes stable modules to `stable_verify` when stable shared truth drifted
   - reports `bound_modules`-only metadata drift without invalidating modules

## Boundary

This CLI does not try to replace semantic judgment performed by the runtime.

It is intentionally not responsible for:

1. `cand_check` closure judgment
2. shared or module boundary judgment
3. verification evidence judgment
4. severity or downgrade decisions

Those remain in the governance documents and the agent runtime.
