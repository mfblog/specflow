## 基础要求
1. 和用户的所有对话都使用中文。
2. 和用户沟通的时候，对于技术细节的描述，在保证准确性的前提下，尽量通俗。
3. 和用户沟通的时候，避免使用未解释的抽象概念，确保双方对齐。

## 第一性原理

请使用第一性原理思考。不要假设用户已经非常清楚自己想要什么和该怎么得到。请从原始需求和目标出发，如果动机和目标不清晰，先停下来讨论。

## 方案规范

当需要给出修改或重构方案时，必须符合以下规范：

- 不允许给出兼容性或补丁性的方案。
- 不允许过度设计，保持最短路径实现。
- 不允许自行给出用户需求以外的方案。
- 必须确保方案逻辑正确，并经过全链路验证。

## Spec 与 Git 总原则

你在一个采用 Spec 驱动开发的仓库中工作。默认遵守以下原则：

1. 涉及实现、审查、升级时，必须先按对应命令或 policy 对齐目标模块的 Spec 与状态，不得绕开 `docs/specs/` 中的真相文件直接猜测行为。
2. 当不确定是否属于行为变化时，默认视为行为变化；行为变化不得代码先行，必须先遵守 `docs/agent_guidelines/spec_policy.md`。
3. 新模块首版允许先有 `candidate`，之后再由 `cand_promote` 生成第一份 `stable`。
4. 历史模块首次纳管应先通过 `spec_init:{module}` 建立第一份 `stable`。
5. 若本次是代码实现类修改，在确认修改有效后，必须按 `docs/agent_guidelines/git_policy.md` 的当前规则判断是否需要提交；默认应在当前任务内完成提交。
6. `docs/specs/` 中除 `candidate` 层主文件及其附属展开文件外的 Spec 文件，属于行为真相源；其修改默认应立刻纳入 git 历史。
7. `candidate` 层主文件及其附属展开文件属于候选草案层；若本次只修改这类文件，默认不执行 `git commit`，除非用户明确要求，或命中要求提交的命令流程。
8. `docs/agent_guidelines/*.md` 的修改默认也应在当前任务内执行 `git commit`。
9. 不要假设规则；遇到 Spec、命令或提交流程冲突时，回到对应 policy 或命令文件确认。

## 命令

命令是一类标准工作流入口。
当用户的请求明显命中某个命令时，优先按该命令执行。`AGENTS.md` 只负责给出命令索引；具体命中细则与执行步骤以对应命令文件为准。

- 命令总规范见：`docs/agent_guidelines/command_policy.md`。
- 具体命令放在：`docs/agent_guidelines/commands/`。
- 标准命令调用格式为：`{command}:{module}`。
- 其中 `{module}` 默认应为 `docs/specs/_status.md` 中登记的正式模块名；该状态表的职责、字段语义与读取规则以 `docs/agent_guidelines/spec_policy.md` 为准。`spec_init:{module}` 与 `spec_new:{module}` 允许用于首次纳管或首次立项的新模块名。

### Spec 文件指代示例

- 对话中若要直接指某一份 Spec 文件，优先使用文件名前缀写法：
  - `s_module_example`：表示 `stable` 层文件
  - `c_module_example`：表示 `candidate` 层文件
- 示例：
  - “请修改 `c_module_example`”
  - “请查看 `s_module_example`”
  - “请对 `c_module_example` 做收口审查”
- 若用户只说 `module_example`，默认它是模块名，不是具体文件名；Agent 必须先按 `docs/agent_guidelines/spec_policy.md` 中对 `_status.md` 的读取规则读取 `docs/specs/_status.md` 的 `Active Layer` 判定实际落点，再在执行前明确回显本次会改哪一个文件。

- `spec_init:{module}`：为历史模块补建第一份 `stable`。细则见 `docs/agent_guidelines/commands/spec_init.md`。
- `stable_verify:{module}`：核对当前代码是否仍对齐 `stable`。细则见 `docs/agent_guidelines/commands/stable_verify.md`。
- `spec_new:{module}`：为全新模块建立第一份 `candidate`。细则见 `docs/agent_guidelines/commands/spec_new.md`。
- `spec_fork:{module}`：从已有 `stable` 派生一份新的 `candidate`。细则见 `docs/agent_guidelines/commands/spec_fork.md`。
- `cand_check:{module}`：检查 `candidate` 是否已收口到足以进入计划阶段。细则见 `docs/agent_guidelines/commands/cand_check.md`。
- `cand_plan:{module}`：根据已收口的 `candidate` 生成当前轮 `_plans/{module}.md`。细则见 `docs/agent_guidelines/commands/cand_plan.md`。
- `cand_impl:{module}`：按 `candidate` 与升级清单推进代码实现。细则见 `docs/agent_guidelines/commands/cand_impl.md`。
- `cand_verify:{module}`：核对当前代码是否已对齐 `candidate`。细则见 `docs/agent_guidelines/commands/cand_verify.md`。
- `cand_promote:{module}`：将 `candidate` 正式提升为新的 `stable`。细则见 `docs/agent_guidelines/commands/cand_promote.md`。

### Spec Flow 审查触发词

- `spec_flow_review`：用于审查当前待审范围内的 Spec Flow 治理机制是否仍然闭环，以及是否会给现有流程带来副作用。细则见 `docs/agent_guidelines/spec_flow_review.md`。
- `shared_extract_review`：用于判断某段当前写在模块主文件或模块 appendix 中的内容，是否已经达到提取为 Shared Appendix 的标准。细则见 `docs/agent_guidelines/shared_extract_review.md`。

补充规则：

1. `spec_flow_review` 不是 `{command}:{module}` 形式的标准模块命令。
2. 它不进入 `docs/specs/_status.md`，也不参与模块 `stable / candidate` 状态机。
3. 它只审查治理规则本身，不替代 `cand_check`、`stable_verify`、`cand_verify` 等模块或实现侧审查命令。
4. 当用户只说 `spec_flow_review` 时，默认审查完整 Spec Flow 治理规则基线；只有用户明确说明“只审某批改动 / 某些文件 / 某个治理专题”时，才允许缩小范围。
5. `shared_flow_reconcile` 也不是 `{command}:{module}` 形式的标准模块命令；它不进入 `docs/specs/_status.md`，只负责 Shared Appendix 变更后的状态收口，但默认作为 Agent 内部流程触发，不要求用户直接输入该术语。
6. 用户若表达“shared 改完后还有哪些模块受影响”“这些 shared 绑定改动要不要统一回退”这类意图，Agent 应识别为 Shared Appendix 状态收口场景，并在内部执行 `docs/agent_guidelines/shared_flow_reconcile.md`。
7. `shared_extract_review` 也不是 `{command}:{module}` 形式的标准模块命令；它不进入 `docs/specs/_status.md`，只负责 shared 提取边界审查。

## 通用经验

### 规则文档写法

- 规则文档、命令文档和其它机制类文档必须写成可直接阅读的正文：
  1. 文档内容不得依赖未写出的对话上下文；所有依赖必须在文档中直接给出，或提供明确链接。
  2. 文档正文不得写成补丁说明或修改说明，而必须直接写规则本身。

### 复杂逻辑解释

- 和用户沟通过程中，只要遇到复杂逻辑概念，除了文字解释外，应优先补一个最小可理解的 Mermaid 图。
- 图的目标是帮助用户理解“对象是什么、关系是什么、卡点在哪里”，因此应优先使用最小图，避免为了完整而把图画得过大、过乱。
- 图不能替代文字：给图时，仍应配一段通俗解释，明确说明图中的关键节点、主链路和当前问题点。

### 图表语法兼容

- 只要涉及流程图生成或排查相关的渲染错误，就必须遵守稳妥且跨渲染器兼容的语法子集要求。
- 详细规范见：`docs/agent_guidelines/chart_syntax_compatibility.md`。

### Prompt 编写规范

- 只要涉及 Prompt 的设计、编写或修改，就必须遵守结构化、可测试、无未解释概念的严谨要求。
- 详细规范见：`docs/agent_guidelines/prompt_guidelines.md`。

### 审查与评审规范

- 进入候选收口审查时，若用户未指明更具体范围，优先按 `docs/agent_guidelines/commands/cand_check.md` 执行。

### LLM 输出协议与解析

- 只要 Prompt 约定“用 tag 包裹 JSON 输出”，Prompt 侧规范看 `docs/agent_guidelines/prompt_guidelines.md`。
- 实现侧解析流程、错误处理原则与复用入口，必须看 Shared Appendix 或模块当前 `Global Constraint Alignment` 中声明的正式绑定；当前结构化输出 fallback 共享协议见 `docs/specs/shared/candidate/c_shared_structured_output_fallback.md`。
