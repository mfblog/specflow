# Scenario Stable Verify Command

## 1. Purpose

`scenario_stable_verify:{scenario}` checks whether current repository truth still aligns with the stable scenario truth.

## 2. Lifecycle-State Advance Inheritance

Lifecycle-state advancement follows `specflow/framework/command_policy.md` Sections 8.5 and 8.8.
This file states only `scenario_stable_verify`-local entry, output, and stop rules.

## 3. Preconditions

1. `_status.md` says `Object Type=scenario`, `Active Layer=stable`, `Next Command=scenario_stable_verify`
2. current stable scenario file exists
3. if `_verify_result/scenario/{scenario}.md`, `_status.md`, or other commit-triggering governance files may change, read the git policy first

## 4. Procedure

1. read stable scenario truth
2. revalidate current bound units, shared contracts, and stable baseline
3. if still aligned, advance `Next Command=scenario_fork`
4. if drift exists, keep `Next Command=scenario_stable_verify`
5. perform git close-out if required

## 5. Output Contract

The output must report:

1. stable alignment result
2. whether any `_verify_result/scenario/{scenario}.md` write, delete, or keep action occurred
3. `_status.md` update result
4. `round conclusion`
5. `current state`
6. `next step`
7. `why this next step`
8. `next-stage entry gap`
9. git close-out result
10. the `user-facing close-out block` required by `specflow/framework/command_policy.md` Section 8.6
11. if a future extension introduces a checkpoint stop, the same close-out block must also report `resume signal`

## 6. Non-Goals

1. scenario candidate authoring
2. unit implementation repair
