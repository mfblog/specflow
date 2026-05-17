# Downgrade Policy

Downgrade means a command may continue with a clearly bounded evidence limitation.

Only unit commands may use this policy.

## 1. Allowed Commands

Downgrade may be considered for:

1. `unit_verify`
2. `unit_stable_verify`

## 2. Rule

A downgrade is allowed only when the remaining uncertainty does not weaken the unit claim being made.

If the limitation affects behavior truth, rule truth, unit dependency truth, or acceptance item coverage, do not downgrade. Route to the smallest legal fallback command.

## 3. Reporting

The command result must state:

1. what could not be proven
2. why the remaining uncertainty is bounded
3. why the unit claim remains valid

## 4. Rejection

No scenario downgrade path is supported.
