# Output Baseline

## 1. Purpose

This file defines the framework baseline minimum output quality rules for the `specflow_response` / `user_facing_response_clarity` shared surface.

These rules are framework baseline. They apply automatically to every specFlow consumer that produces user-facing output on this shared surface. No registry entry or project-level activation is required.

Project-level output standards may tighten these rules through the registered project standards mechanism defined by `specflow/framework/project_standards_policy.md`.

This file answers four questions:

1. what principle governs user-facing answer content
2. which minimum information a user-facing answer must cover
3. where execution notes belong and what they may contain
4. which output forms are forbidden

---

## 2. Core Principle

User-facing output on this surface must serve user understanding first, not internal tracking.

The user must be able to understand, without knowing internal command names, lifecycle state names, object-family names, file paths, line numbers, or `fallback_reason_code`:

1. what conclusion was reached this round
2. what the executor actually did
3. what state the project is now in
4. what should be done next
5. why that next action is required
6. what the expected result of that action is
7. what still cannot be claimed as complete

Internal tracking information may be preserved, but only in a short execution note after the main answer.

---

## 3. Minimum Information Responsibilities

Every user-facing main answer must cover the following information responsibilities. Expression may vary by context, but the information must not be missing.

1. **Round conclusion** -- whether the result is completion, failure, fix required, user confirmation needed, or explanation only.
2. **Completed actions** -- what was actually checked, modified, verified, compared, or read.
3. **Current state** -- where the project is stuck, expressed through the current project's capability areas, responsibility areas, entry points, or delivery surfaces.
4. **Next step** -- what should be done first, expressed as a plain engineering action.
5. **Reason** -- why the current state requires this step before proceeding to the next one.
6. **Expected result** -- what state or checkable result should be reachable after completing the next step.
7. **Remaining gap** -- what still cannot be claimed as done after this step; if no gap remains, say so explicitly.

When an item is truly not applicable in the current context, it may be omitted. The main answer must still make the current state and next step clear to the user.

Framework consumers (such as command close-out blocks defined by `specflow/framework/command_policy.md` Section 8.6 and NL routing output defined by `specflow/framework/natural_language_routing.md` Section 11.1) may define stricter or context-specific field lists. Those lists inherit this baseline and may add requirements, but must not weaken the minimum information responsibilities defined here.

---

## 4. Main Answer vs Execution Note Separation

Execution notes must be separate from the main answer and serve only as trace material.

Execution notes may contain:

1. internal command names
2. lifecycle state names
3. object-family names
4. file paths and line numbers
5. check dimensions
6. `fallback_reason_code`
7. governance files read or modified

Execution notes must not:

1. be the only source for understanding current state
2. be the only source for understanding the next step
3. replace the main answer for explaining why work cannot continue
4. use internal field listings as a substitute for plain-language explanation
5. stack full review details before the main answer

---

## 5. Forbidden Output Shapes

The following output forms are forbidden on this surface:

1. Starting with an internal command result and requiring the user to understand that command to know the outcome.
2. Saying only `fix_required`, `blocked`, `pass`, or similar internal conclusion without translating to ordinary language.
3. Listing only check fields such as `progressability`, `content completeness`, `Candidate Design Quality` without saying what the user should do next.
4. Saying only "cannot proceed to next step" without stating what needs to be fixed first.
5. Using `fallback_reason_code` as the main answer.
6. Stacking file paths, line numbers, status tables, and internal object names as the answer body opening.

Project-level standards registered on this surface may add additional forbidden shapes through the registered standards mechanism defined by `specflow/framework/project_standards_policy.md`.

---

## 6. Relationship to Other Framework Documents

`specflow/framework/command_policy.md` Section 8.6 defines the user-facing close-out block contract for formal commands. That contract inherits this baseline and adds command-specific fields.

`specflow/framework/natural_language_routing.md` Section 11.1 defines the output contract for natural-language routing. That contract inherits this baseline and adds routing-specific fields.

`specflow/framework/project_standards_policy.md` Section 10.1 documents how project-level standards on the `user_facing_response_clarity` surface may tighten this baseline.

---

## 7. Non-Goals

This file does not:

1. define fixed reply templates
2. change specFlow command result sets
3. change `_status.md` update rules
4. change `_check_result` write-back rules
5. change checkpoint, fallback, review, or promote semantics
6. require Mermaid diagrams in every output
7. require hiding all internal information -- only that internal information must not substitute for user-readable explanation
