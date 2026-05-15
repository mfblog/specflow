---
id: repository_mapping
version: 0.1.0
---

# Repository Mapping

This file records the current repository-structure truth.

It is not a command object.
It has no `stable`, `candidate`, `fork`, `verify`, or `promote` lifecycle.
It is read by `unit`, `scenario`, rule-governance, review, and implementation flows when they need to understand path ownership, governed object boundaries, support surfaces, or repository-level drift.

## 1. Project Overview

Record the repository's purpose, main technology stack, runtime entry points, and any major repository conventions that a human or agent must know before changing governed files.

## 2. Object Registry

This table is the only structured registry that connects current `unit`, `scenario`, and `rule` objects to implementation paths in this repository.

The table header is fixed:

| kind | id | scope | registration_state | implementation_paths | spec_files | responsibility |
|---|---|---|---|---|---|---|
| unit | example_unit | capability | planned | none | none | Replace this row with the first registered unit. |

Rules:

1. `kind` must be `unit`, `scenario`, or `rule`.
2. `id` must be the formal object ID.
3. `scope` must be `capability` for `unit`, `flow` for `scenario`, and `bound` or `global` for `rule`.
4. `registration_state` must be `planned` or `landed`.
5. `planned` rows must use `implementation_paths=none`.
6. `landed` rows must list the concrete implementation path or paths.
7. `spec_files` lists the related Spec or rule documents, or `none` when no document file exists yet.
8. Multiple paths in `implementation_paths` or `spec_files` must be separated with `;`.

## 3. Boundary Rules

State how this repository decides what becomes a governed `unit`, what remains a support surface, and what should stay outside specFlow governance.

These rules must be specific enough for a later command to decide whether a changed path belongs to:

1. one formal `unit`
2. one formal `scenario`
3. one `rule`
4. a support surface
5. ignored or external content

## 4. Path Ownership

Record the path rules that connect repository files to formal objects.

The map must make these decisions resolvable:

1. which source paths are governed by each `unit`
2. which truth files belong to each `unit`, `scenario`, and `rule`
3. which paths are support surfaces
4. which paths are ignored
5. which precedence rule applies when a path could match more than one category

## 5. Rule Alignment

Record which global rules and rules shape repository-level structure.

At minimum, state the current stable `g_` rule reference and any repository-wide Rule assumptions that affect path ownership or object boundaries.

## 6. Drift Handling

Record what counts as repository-structure drift.

At minimum, treat these cases as drift:

1. a governed path no longer maps to a formal object
2. one path maps to multiple formal objects
3. a declared formal object has no matching truth file
4. a support surface is used as if it were a command object
5. a command tries to modify implementation paths outside the target object's declared ownership

When drift is detected, the consuming flow must stop boundary-sensitive work and update this file before continuing.
