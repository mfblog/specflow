### Manual Command Close (when `specflowctl` is unavailable)

When `specflowctl command close` is unavailable (tooling not installed, broken, or inaccessible), perform a manual close following these deterministic rules. This is the **only** exception to the rule that `command close` is the sole mechanism for advancing lifecycle state.

**Manual close is scoped to the current lifecycle command only.** It must not be used to skip lifecycle phases, jump ahead in the lifecycle sequence, or perform close operations that involve automatic file mutations that manual file editing cannot reliably reproduce.

**Pre-conditions (mandatory — all must pass):**

1. All required writes from the "How to End" outcome above are complete and correct.
2. All process evidence files are written with the correct schema (see `framework/process_snapshot_contract.md` for file format).
3. For advancing outcomes: the independent evaluation receipt is present in the process evidence, satisfying gate rule requirements from `framework/core/independent_evaluation.md` Section Gate Rules.
4. The `docs/specs/_status.md` file is readable and the target unit's `Next Command` matches the command being closed.

If any pre-condition fails: STOP, report what is missing, and do not perform the manual close.

**Procedure:**

1. From the "How to End" outcome table above, identify your outcome and its Next Step column.
2. Update `docs/specs/_status.md` for the target unit:
   - Set `Next Command` to the value specified in the outcome's Next Step.
   - Set or clear `Notes` per the outcome's Next Step description.
   - For `unit_fork` with outcome `candidate_created`: set `Active Layer` to `candidate`.
   - For `unit_promote` with outcome `promoted`: set `Active Layer` to `stable`, `Stable` to `yes`, `Candidate` to `no`.
   - For `unit_init` with outcome `stable_created`: set `Stable=yes`, `Candidate=no`, `Active Layer=stable`.
   - For `unit_new` with outcome `candidate_created`: set `Stable=no`, `Candidate=yes`, `Active Layer=candidate`.
   - For all other commands and outcomes: do **not** change `Active Layer`, `Stable`, or `Candidate`.
3. If the target unit has **no row** in `_status.md` (applies to `unit_init` and `unit_new`), add a new row with the columns `| unit | {unit} | ... |` and fill values from the mapping above.
4. Perform the cleanup described in the outcome's Next Step column (delete specified evidence files, preserve others).
5. Write the updated `docs/specs/_status.md`.

**Recording the fallback:**

Add the following to the command's process evidence file (if one exists):

```yaml
command_close_fallback: manual
command_close_fallback_recorded_at: <UTC ISO 8601 timestamp>
```

This annotation documents that manual intervention occurred and is consumed by subsequent executors only as advisory context — it is not a lifecycle gate validation input.

For the reference per-outcome state transition mapping across all lifecycle commands, see `framework/lifecycle/overview.md:114-145`.
