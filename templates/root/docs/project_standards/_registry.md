# 项目标准注册表

> 本文件是项目级标准的唯一正式入口。只有在这里登记过的 `docs/project_standards/*.md` 文件，才允许影响命令执行。命令不得自行扫描整个目录猜测哪些文件生效。

## Purpose

本文件回答四件事：

1. 当前项目启用了哪些项目级标准。
2. 每个标准的正式文件路径是什么。
3. 哪些命令或内部流程会消费它。
4. 它是在补充说明，还是在收紧项目要求。

## Active Standards

| standard_id | type | surface | file | consumed_by | applies_to | effect | conflict_rule | notes |
|---|---|---|---|---|---|---|---|---|

## Rules

1. `type` 只允许使用框架已定义的支持类型。
2. `surface` 必须使用框架允许的稳定命名；若用于 `cand_check` 的通用候选收口扩展，固定写 `candidate_closure_review`。
3. `file` 必须位于 `docs/project_standards/` 下。
4. `consumed_by` 必须显式写出命令名或内部流程名，不得写成 `all`。
5. `effect` 只允许：
   - `clarify`
   - `tighten`
6. `conflict_rule` 固定写 `framework_wins`。
7. 未登记的标准文件，即使存在，也不得影响命令执行。
8. 项目级标准可以补充或收紧框架规则，但不得削弱框架底线。
