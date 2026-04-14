## specFlow 入口说明

本文件只负责把用户请求路由到 specFlow 的正式流程。
它不是业务规则正文，也不承载项目成员的个人偏好。

## 核心原则

1. 涉及实现、审查、升级时，必须先按对应命令或 policy 对齐目标模块的 Spec 与状态，不得绕开 `docs/specs/` 中的真相文件直接猜测行为。
2. 当不确定是否属于行为变化时，默认视为行为变化；行为变化不得代码先行，必须先遵守 `docs/agent_guidelines/spec_policy.md`。
3. 新模块首版允许先有 `candidate`，之后再由 `cand_promote` 生成第一份 `stable`。
4. 历史模块首次纳管应先通过 `spec_init:{module}` 建立第一份 `stable`。
5. 若本次是代码实现类修改，在确认修改有效后，必须按 `docs/agent_guidelines/git_policy.md` 的当前规则判断是否需要提交；默认应在当前任务内完成提交。
6. `docs/specs/` 中除 `candidate` 层主文件及其附属展开文件外的 Spec 文件，属于行为真相源；其修改默认应立刻纳入 git 历史。
7. `candidate` 层主文件及其附属展开文件属于候选草案层；若本次只修改这类文件，默认不执行 `git commit`，除非用户明确要求，或命中要求提交的命令流程。
8. `docs/agent_guidelines/*.md` 的修改默认也应在当前任务内执行 `git commit`。
9. 不要假设规则；遇到 Spec、命令或提交流程冲突时，回到对应 policy 或命令文件确认。

## 命令入口

- 命令总规范见：`docs/agent_guidelines/command_policy.md`
- 具体命令放在：`docs/agent_guidelines/commands/`
- 标准命令调用格式为：`{command}:{module}`
- `{module}` 默认应为 `docs/specs/_status.md` 中登记的正式模块名；该状态表的职责、字段语义与读取规则以 `docs/agent_guidelines/spec_policy.md` 为准

### Spec 文件指代示例

- `s_module_example`：表示 `stable` 层文件
- `c_module_example`：表示 `candidate` 层文件
- 若用户只说 `module_example`，默认它是模块名，不是具体文件名；执行前必须先按 `docs/specs/_status.md` 的 `Active Layer` 判定实际落点

### 标准命令

- `spec_init:{module}`：为历史模块补建第一份 `stable`
- `stable_verify:{module}`：核对当前代码是否仍对齐 `stable`
- `spec_new:{module}`：为全新模块建立第一份 `candidate`
- `spec_fork:{module}`：从已有 `stable` 派生一份新的 `candidate`
- `cand_check:{module}`：检查 `candidate` 是否已收口到足以进入计划阶段
- `cand_plan:{module}`：根据已收口的 `candidate` 生成当前轮 `_plans/{module}.md`
- `cand_impl:{module}`：按 `candidate` 与升级清单推进代码实现
- `cand_verify:{module}`：核对当前代码是否已对齐 `candidate`
- `cand_promote:{module}`：将 `candidate` 正式提升为新的 `stable`

### 治理审查入口

- `spec_flow_review`：审查当前待审范围内的 specFlow 治理机制是否仍然闭环，以及是否会给现有流程带来副作用
- `shared_extract_review`：判断某段当前写在模块主文件或模块 appendix 中的内容，是否已经达到提取为 Shared Appendix 的标准

补充规则：

1. `spec_flow_review` 不是 `{command}:{module}` 形式的标准模块命令。
2. 它不进入 `docs/specs/_status.md`，也不参与模块 `stable / candidate` 状态机。
3. 它只审查治理规则本身，不替代 `cand_check`、`stable_verify`、`cand_verify` 等模块或实现侧审查命令。
4. `shared_flow_reconcile` 不是标准模块命令；它不进入 `docs/specs/_status.md`，只负责 Shared Appendix 变更后的状态收口。
5. `shared_extract_review` 也不是标准模块命令；它不进入 `docs/specs/_status.md`，只负责 shared 提取边界审查。

## 文档与图表要求

1. 规则文档、命令文档和其它机制类文档必须写成可直接阅读的正文，不得依赖未写出的对话上下文。
2. 文档正文不得写成补丁说明，而必须直接写规则本身。
3. 只要涉及复杂流程、状态机、版本关系、边界判定，优先补最小可理解的 Mermaid 图帮助对齐。
4. 只要涉及图表生成或渲染兼容问题，遵守 `docs/agent_guidelines/chart_syntax_compatibility.md`。
5. 只要涉及 Prompt 的设计、编写或修改，遵守 `docs/agent_guidelines/prompt_guidelines.md`。
6. 只要 Prompt 约定“用 tag 包裹 JSON 输出”，实现侧解析流程与错误处理必须按正式绑定的 Shared Appendix 或模块当前 `Global Constraint Alignment` 执行。
