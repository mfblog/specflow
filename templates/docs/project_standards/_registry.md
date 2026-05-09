# 项目标准注册表

> 本文件是项目级标准的唯一正式入口。只有在这里登记过的 `docs/project_standards/*.md` 文件，才允许影响命令执行。命令不得自行扫描整个目录猜测哪些文件生效。

## Purpose

本文件回答四件事：

1. 当前项目启用了哪些项目级标准。
2. 每个标准的正式文件路径是什么。
3. 哪些命令、内部流程或共享输出消费方会消费它。
4. 它是在哪个已定义的消费面上生效，以及它是在补充说明还是在收紧项目要求。

## Active Standards

| standard_id | type | surface | file | consumed_by | applies_to | effect | conflict_rule | notes |
|---|---|---|---|---|---|---|---|---|

## Rules

1. `type` 只允许使用框架已定义的支持类型。
2. `surface` 必须引用消费它的命令、内部流程或共享输出消费方已正式定义的稳定名称，不得由本注册表自行命名。
3. `file` 必须位于 `docs/project_standards/` 下。
4. `consumed_by` 必须显式写出已声明支持该标准类型与 `surface` 的命令名、内部流程名或共享输出消费方，不得写成 `all`。
5. `applies_to` 必须使用 framework 已定义的正式 selector 语法，不得写项目自造短语。
6. `effect` 只允许：
   - `clarify`
   - `tighten`
7. `conflict_rule` 固定写 `framework_wins`。
8. 未登记的标准文件，即使存在，也不得影响命令执行。
9. 本文件只负责登记启用关系，不负责定义框架接口、命令接口或消费语义。
10. 本文件不得创建新的 `surface`，不得通过注册动作扩展 framework 或 command 接口。
11. 若某条 entry 引用了未被对应命令、内部流程或共享输出消费方正式定义的 `surface`，该 entry 不能作为合法治理输入。
12. 若某条 entry 使用了 framework 未定义的 `applies_to` selector，该 entry 不能作为合法治理输入。
13. 项目级标准可以补充或收紧框架规则，但不得削弱框架底线。

## Applies To Selector

`applies_to` 只允许使用以下 selector 形式：

1. `all_targets_on_surface`
   - 表示所有已经命中该命令已定义消费面的目标都适用
2. `unit:<formal_unit_name>`
   - 表示只适用于一个正式单元
3. `unit_set:<formal_unit_name>,<formal_unit_name>,...`
   - 表示只适用于列出的正式单元集合
   - 单元名必须来自 `docs/specs/_status.md`
   - 逗号分隔列表内部不得带空格
4. `review_scenario:<stable_name>`
   - 表示只适用于消费方已经正式定义过的某个 review scenario
