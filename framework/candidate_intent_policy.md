# Candidate Intent Policy

`candidate_intent` explains why a unit candidate exists.

Only unit candidates use candidate intent.

## 1. Values

Supported values:

1. `change`
2. `repair`

## 2. Change

Use `change` when the candidate intentionally changes stable unit behavior, dependency, rule binding, acceptance criteria, or implementation expectations.

## 3. Repair

Use `repair` when the candidate keeps the intended stable behavior but repairs missing, stale, malformed, or insufficient truth.

## 4. Requirement

`unit_fork` must write `candidate_intent`.

No scenario candidate intent path is supported.
