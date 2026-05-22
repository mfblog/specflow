# Spec Authoring Baseline

This file defines the semantic authoring baseline for formal Spec truth.

It controls what a formal Spec must make understandable before it can be used by downstream check, plan, implementation, verification, or promotion work.

It does not control prose style, heading names, diagram use, narrative order, or whether the author uses tables, lists, examples, or paragraphs.

## 1. Authoring Freedom

Spec authors may choose the readable shape of a document.

A Spec may use:

1. prose
2. lists
3. tables
4. diagrams
5. examples
6. appendix files
7. any section order that remains understandable

This freedom does not allow formal behavior truth to depend on:

1. unstated chat context
2. author memory
3. rejected alternatives
4. future implementer invention
5. undocumented assumptions

Format compliance is not enough for downstream readiness.
A Spec can have valid frontmatter, valid references, and valid acceptance item shape while still failing this baseline.

## 2. Handoff Completeness

A formal Spec is handoff-complete only when the current main Spec and its explicitly referenced appendix files explain enough for the next lifecycle step to continue without inventing missing design.

When the Spec describes behavior that affects implementation, it must make the following information clear in the document body or explicitly linked appendix truth:

1. the intended user, actor, or caller
2. the unit responsibility and why the unit owns it
3. the entry point or trigger that starts the behavior
4. the normal path from input to result
5. the boundaries crossed on that path
6. the data, state, or durable truth each step reads or writes
7. the owner of each read/write responsibility
8. the output artifact or observable result
9. the way failures, gaps, or unavailable dependencies are exposed
10. the verification surface and success condition

The Spec does not need to use these exact labels.
It must still answer these questions clearly enough that a reviewer, planner, or implementer can find the answer in formal truth.

## 3. Decision Closure

A formal Spec must close implementation-affecting decisions that belong to the current unit or current rule.

The downstream executor must not be forced to choose:

1. which object owns a responsibility
2. which entry point starts the behavior
3. where state or durable truth lives
4. which object writes data and which object reads it
5. how ordered steps connect
6. how boundary failures are reported
7. what the result shape means
8. how acceptance proves the stated responsibility

If a decision is intentionally not made in the current round, the Spec must state that boundary directly and explain whether the missing decision is:

1. out of scope
2. owned by another formal object
3. a required future candidate
4. a current blocker
5. represented as an explicit gap in the current behavior

## 4. Coordinated Behavior Detail

When a formal truth statement says that a result is produced by multiple objects, multiple boundaries, multiple steps, or shared state, the Spec must explain how those parts connect.

The required detail is determined by downstream need, not by the names of the objects.

At minimum, the Spec must identify:

1. the participating objects or roles
2. the direction of each important interaction
3. the data or state crossing each boundary
4. the consistency source when one part writes and another part reads
5. the observable signal that proves the connection happened
6. the failure or gap signal when the connection cannot happen

A statement that something "supports", "uses", "integrates with", "connects to", or "is traceable through" another thing is not sufficient by itself.
The Spec must also explain the connection path.

## 5. Appendix Handoff

Appendix files may carry detailed truth for one unit, but they do not weaken the handoff baseline.

If the main Spec delegates behavior, boundary, state, flow, output, or verification details to an appendix, that appendix must provide the missing handoff information directly.

An appendix that is used as implementation truth must not contain only:

1. background explanation
2. motivation
3. principles
4. a list of desired qualities
5. a patch-note description of what changed

It must state the current rule or design as directly readable truth.

## 6. Failure And Gap Expression

If a missing dependency, unavailable state, unsupported path, or known limitation affects the behavior, the Spec must say how that condition appears to the caller, reviewer, report, trace, test, or downstream process.

It is not enough to say that the condition is an error.
The Spec must identify the observable surface for the error or gap.

If the current round intentionally exposes a gap instead of fixing it, the Spec must say:

1. where the gap is observed
2. why the gap is acceptable in the current round
3. which acceptance item or finding proves that the gap is not silent

## 7. Downstream Readiness

This baseline is satisfied only when the next legal lifecycle step can proceed from formal truth.

For `unit_check`, failure to satisfy this baseline is a truth completeness problem.
The command must not pass by assuming that `unit_plan`, `unit_impl`, or a future human reader will fill in missing behavior, boundary, state, output, or verification decisions.
