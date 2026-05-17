# Rule Escape

`rule_escape` is the internal rule-governance flow for stopping unsafe rule work and routing it to the smallest legal next action.

It is used when a rule-governance request is ambiguous, combines multiple rule actions, or returns from another rule flow because current repository truth was not sufficient to close safely.

## 1. Scope

`rule_escape` may:

1. route a request to exactly one rule flow
2. decompose a complex request into a stable sequence of rule flows
3. raise a clarification checkpoint
4. raise a decision checkpoint
5. raise a prerequisite-action checkpoint
6. run rule-governance recovery before rerouting when another rule flow already mutated files and can no longer close safely

`rule_escape` must not:

1. write rule truth as a substitute for the routed rule flow
2. write unit truth as a substitute for unit lifecycle or `rule_bind`
3. keep unresolved truth only in chat
4. create a durable command chain outside the standard rule flows
5. resume an old decomposition after the current handling has stopped

## 2. Required Reads

Before routing or checkpointing, read only the smallest durable truth needed for the decision:

1. `specflow/framework/spec_policy.md`
2. `specflow/framework/command_policy.md`
3. `specflow/framework/checkpoint_protocol.md`
4. `specflow/framework/recovery_policy.md` when control returned after mutation
5. `docs/specs/_status.md` when existing units are named or affected
6. the current-layer unit main Specs needed to judge unit-local truth, binding, or writeback legality
7. the relevant rule files
8. `docs/specs/repository_mapping.md` when path ownership or rule object registration matters
9. `docs/specs/rules/stable/s_g_rule_repository_baseline.md` when the request may become a repository-wide default rule

Consumer discovery must use only current-layer unit frontmatter `rule_refs`.

## 3. Routing Decisions

Route to:

1. `rule_new` when independent rule truth must be authored from the start or reopened at candidate layer
2. `rule_extract` when existing unit-local formal truth must move into a rule
3. `rule_bind` when a unit must consume, remove, or retarget an existing rule binding
4. `rule_topology` when rule files or rule bindings need structural change or terminal-state resolution
5. `rule_sync` when rule truth or binding has already changed and only downstream impact must be reconciled
6. unit lifecycle when the change is unit-local behavior truth
7. repository mapping governance when the change is path ownership or object registration

If more than one rule flow is required, `rule_escape` may produce an execution-local `remaining_steps_contract` only when the step order cannot change the resulting formal truth.

## 4. Procedure

1. Identify the smallest distinct actions inside the request.
2. If control returned from another rule flow after file mutation, use `recovery_policy.md` Section 6.5 before any new routing decision unless that returning flow already completed recovery.
3. Test whether the current request can route to exactly one rule flow without ambiguity.
4. If exactly one flow is legal, route to that flow and stop.
5. If multiple flows are involved, test whether their order is stable from current repository truth.
6. If the order is stable, emit an execution-local `remaining_steps_contract` with:
   - `step_order`
   - `current_step`
   - `remaining_steps`
   - `closure_rule`
   - `durability=execution_local`
   - `resume_rule=rerun_natural_language_routing_from_current_truth_if_interrupted`
7. Route only the first legal step after emitting that contract.
8. If the order is not stable, raise a checkpoint instead of guessing.
9. If the boundary between unit-local truth and rule truth is unclear, raise a checkpoint instead of writing truth.
10. If the request crosses out of rule governance, return to the owning unit lifecycle or repository mapping route.

## 5. Checkpoints

Checkpoint `target_objects` may list only:

1. `unit:{unit}`
2. `none`

No scenario target is supported.

Use:

1. `clarification` when the requested meaning is unclear
2. `decision` when the user must choose between two valid formal landing points
3. `prerequisite_action` when a legal upstream action, such as `unit_fork:{unit}`, must happen before writeback

The checkpoint must state the blocking question or action, why it blocks closure, the required writeback target when one exists, and the resume signal.

## 6. Stop Conditions

Stop when one of these is true:

1. the request has been reduced to one legal next flow
2. a stable execution-local sequence has been emitted and the first flow is selected
3. a checkpoint has been raised
4. recovery completed and routing must restart from current repository truth
5. the request belongs to unit lifecycle or repository mapping governance instead of rule governance

## 7. Output Contract

The output must report:

1. the recognized request shape
2. why direct routing was possible or unsafe
3. the selected next flow when one exists
4. the execution-local `remaining_steps_contract` when decomposition is safe
5. the checkpoint fields when checkpointing is required
6. any rule-governance recovery that was performed before rerouting
7. confirmation that `target_objects` contains only `unit:{unit}` entries or `none`
