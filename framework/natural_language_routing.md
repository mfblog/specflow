# Natural Language Routing

This file routes user requests into specFlow owners.

The supported owners are:

1. `unit`
2. `rule`
3. `repository_mapping`
4. framework governance flows
5. implementation-only work after permission is proven

`scenario` is not a supported owner.

## 1. Exact Entries

If the request exactly matches `unit_advance:{unit}`, read `specflow/framework/advance_policy.md`.

If the request exactly matches a supported unit command from `specflow/framework/command_policy.md`, read that command file.

If the request exactly matches a rule governance entry, read that rule file.

If the request exactly matches `spec_flow_migrate`, read `specflow/framework/spec_flow_migrate.md`.

If the request uses `scenario_*`, `scenario_advance:{id}`, or `object-type=scenario`, stop and report that scenario lifecycle support has been removed.

## 2. Routing Inputs

Before choosing a route, read the smallest necessary durable truth:

1. `docs/specs/_status.md` for existing unit state
2. `docs/specs/repository_mapping.md` for path ownership
3. current-layer unit truth when a named unit is involved
4. current-layer rule truth when rule governance is involved
5. `specflow/framework/implementation_change_policy.md` for implementation-only work

Do not guess ownership from directory shape alone.

## 3. Unit Route

Route to a unit when the request changes or validates one independently governed engineering responsibility.

The responsibility may be local or end-to-end.

If the user describes a complete workflow result, model it as a unit whose responsibility is that complete result. Do not create a scenario owner.

## 4. Rule Route

Route to rule governance when the request changes shared constraints, reusable prohibitions, mandatory process behavior, or rule binding.

For non-exact rule-governance requests, read `specflow/framework/rule_escape.md` first.
`rule_escape.md` owns selecting `rule_new`, `rule_extract`, `rule_bind`, `rule_topology`, `rule_sync`, unit lifecycle, or repository mapping governance.

Rule consumer discovery must use current-layer unit `rule_refs`.

## 5. Repository Mapping Route

Route to repository mapping when the request changes path ownership, object registration, implementation path registration, or support-surface boundaries.

Repository mapping does not change unit behavior or rule meaning by itself.

## 6. Implementation-Only Route

If the request asks only for implementation-side edits and does not require truth, boundary, shared rule, system rule, migration, governance, or guidance work, enter `specflow/framework/implementation_change_policy.md`.

Implementation permission must be proven before editing implementation files.

## 7. Hard Stops

Stop and ask or reroute when:

1. the target unit is unclear
2. path ownership is unclear
3. a behavior or rule decision exists only in chat and has not been written to durable truth
4. implementation permission is not proven
5. a rule or repository mapping change is required first
6. the request tries to use scenario lifecycle concepts

## 8. User-Facing Reports

Reports must use plain project language.

They should state:

1. current state
2. next action
3. reason
4. expected result
5. remaining gap

Internal file names may appear in a separate execution note, but the user-facing answer must not require the user to understand internal policy names.

## 9. Response Baseline

Natural-language routing output inherits the framework output baseline defined by `specflow/framework/output_baseline.md`.
This routing file may tighten or clarify only:

1. user-facing wording
2. answer ordering
3. main-answer and execution-note separation
4. Mermaid usage in user-facing explanation

Output wording rules must not:

1. route the request
2. change the chosen first step
3. create a new command or governance-flow result
4. affect lifecycle state
5. replace the required ordinary-language fields in Section 11.1

## 10. Execution Notes

Execution notes are optional trace material.

When traceability is needed, the execution note may include:

1. the recognized intent
2. the routed first step or checkpoint type
3. the repository truth used to make the route
4. any missing intent or boundary input that blocked routing
5. the smallest legal next step
6. why that next step is legal

Execution note rules:

1. the execution note must appear after the user-facing main answer
2. it must be short enough that it does not become the answer body
3. it must not be required for the user to understand the current state, next action, reason, expected result, or remaining blocker

## 11. Output Contract

When natural-language routing is the active entry, the output must include:

1. a user-facing main answer
2. an execution note when traceability is needed

The user-facing main answer must be understandable without internal specFlow knowledge.
It must use ordinary project language first and internal names only as trace details.
It must not present internal command names, lifecycle state names, object-family names, policy-file names, or formal route names as the recommended action.

Natural-language routing also inherits:

1. the framework output baseline defined by `specflow/framework/output_baseline.md`

### 11.1 User-Facing Report Contract

The user-facing part of the output must include ordinary-language statements for:

1. `current state`
   - what the repository truth says now, expressed through project structure language
2. `next action`
   - what will be done first, expressed as a plain engineering action
3. `why this action`
   - why this is the legal and useful next action
4. `expected result`
   - what should be true after the next action completes
5. `remaining gap`
   - what still cannot be claimed after this step, if anything

These five fields satisfy the framework output baseline defined by `specflow/framework/output_baseline.md` Section 3.
Fields that are not applicable in a routing-only context, such as `round conclusion` and `completed actions` when the output describes a routing decision without executing work, are covered by the baseline's escape hatch for non-applicable items.

The main answer must not make the user understand internal object-family names, command names, lifecycle state names, or governance-flow names before they can evaluate the state or next action.
Internal names may be included only in the execution note after the project-structure explanation.

If the output starts an existing standard command, the command's own output contract controls the final close-out.
That command output must still follow the user-facing language separation required by `specflow/framework/command_policy.md` and the framework output baseline defined by `specflow/framework/output_baseline.md`.
