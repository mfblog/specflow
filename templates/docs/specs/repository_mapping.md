---
id: repository_mapping
version: 0.1.0
---

# Repository Mapping

This document records repository path ownership and formal object registration.

It is not a lifecycle object.

## 1. Project Overview

Describe the repository's governed roots and primary delivery surfaces here.

## 2. Object Registry

This table is the only structured registry that connects current or planned `unit` and `rule` objects to implementation paths.

| kind | id | registration_state | implementation_paths | spec_files | responsibility |
|---|---|---|---|---|---|
| unit | example | planned | none | none | Example unit responsibility |
| rule | g_rule_repository_baseline | planned | none | `docs/specs/rules/stable/s_g_rule_repository_baseline.md` | Repository-wide baseline rules |

Rules:

1. `kind` must be `unit` or `rule`.
2. `registration_state` must be `planned` or `landed`.
3. `scope` is not a column.
4. rule global or bound scope is resolved from rule frontmatter or id prefix.

## 3. Boundary Rules

Explain which paths are governed roots, support surfaces, ignored paths, and formal object-owned paths.

## 4. Path Ownership

List implementation paths and support paths that are not fully captured by the Object Registry.

Process support paths include:

1. `docs/specs/_governance_review/**`
