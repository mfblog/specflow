## SpecFlow Rules

以下内容只定义“当当前仓库采用 specFlow 时”需要额外遵守的规则。

这些规则是对宿主 Agent 通用规则文件的补充，不替代宿主已有的其它规则。

### 1. 请求识别

收到请求后，若命中以下任一对象，应按 specFlow 规则处理：

1. 标准命令：
   - `{command}:{module}`
2. 治理审查：
   - `spec_flow_review`
   - `shared_extract_review`
3. 涉及模块 Spec、状态推进、候选收口、正式提升、Shared Appendix、系统约束的请求

若未命中以上对象，其余行为继续按宿主 Agent 的其它规则执行。

### 2. 标准命令

标准命令格式：

```text
{command}:{module}
```

命令总规范见：

- `docs/agent_guidelines/command_policy.md`

具体命令文件见：

- `docs/agent_guidelines/commands/`

标准命令包括：

1. `spec_init:{module}`
2. `stable_verify:{module}`
3. `spec_new:{module}`
4. `spec_fork:{module}`
5. `cand_check:{module}`
6. `cand_plan:{module}`
7. `cand_impl:{module}`
8. `cand_verify:{module}`
9. `cand_promote:{module}`

治理审查入口包括：

1. `spec_flow_review`
2. `shared_extract_review`

补充规则：

1. `spec_flow_review` 与 `shared_extract_review` 不是 `{command}:{module}` 形式的标准模块命令。
2. `shared_flow_reconcile` 不是用户直接输入的标准命令；它只用于 Shared Appendix 变更后的状态收口。

### 3. 模块与文件判定

`{module}` 默认指正式模块名，不是具体文件名。

若用户直接说模块名，例如 `module_example`，执行前必须先读取：

- `docs/specs/_status.md`

再根据其中的 `Active Layer` 判定实际落点：

1. `Active Layer=stable`
   - 默认落到 `docs/specs/stable/s_{module}.md`
2. `Active Layer=candidate`
   - 默认落到 `docs/specs/candidate/c_{module}.md`

若用户直接说具体文件前缀，则按文件处理：

1. `s_module_example`
   - 指 `stable` 层主文件
2. `c_module_example`
   - 指 `candidate` 层主文件

### 4. 非命令请求的读取顺序

若请求命中了 specFlow 范围，但不是标准命令，默认按下面顺序处理：

1. 先确认它影响哪个模块或哪个治理对象。
2. 读取 `docs/specs/_status.md`，确认目标模块当前的 `Active Layer` 与 `Next Command`。
3. 若任务涉及模块行为真相，读取对应层的主 Spec。
4. 若主 Spec 明确引用了 appendix 或 Shared Appendix，必须一并读取。
5. 若任务涉及全局技术基线、共享机制或全局例外，再读取：
   - `docs/specs/system/stable/s_system_constraints.md`
6. 再决定当前动作是：
   - 只解释
   - 修改 candidate
   - 修改 stable
   - 执行某个标准命令

### 5. 强制约束

1. 不得绕开 `docs/specs/` 中的真相文件直接猜测行为。
2. 当不确定是否属于行为变化时，默认视为行为变化。
3. 行为变化不得代码先行，必须先遵守 `docs/agent_guidelines/spec_policy.md`。
4. 新模块首版允许先有 `candidate`，之后再由 `cand_promote` 生成第一份 `stable`。
5. 历史模块首次纳管应先通过 `spec_init:{module}` 建立第一份 `stable`。
6. `docs/specs/` 中除 `candidate` 层主文件及其附属展开文件外的 Spec 文件，属于行为真相源；其修改默认应纳入 git 历史。
7. `candidate` 层主文件及其附属展开文件属于候选草案层；若本次只修改这类文件，默认不执行 `git commit`，除非用户明确要求，或命中要求提交的命令流程。
8. `docs/agent_guidelines/*.md` 的修改默认也应在当前任务内执行 `git commit`。
9. 遇到 Spec、命令或提交流程冲突时，不要自行猜测，回到对应 policy 或命令文件确认。

### 6. 必读文件

若当前任务命中了 specFlow 范围，至少应知道以下文件各自负责什么：

1. `docs/agent_guidelines/spec_policy.md`
   - 定义 Spec 对象、层次、真相边界、读取规则
2. `docs/agent_guidelines/command_policy.md`
   - 定义标准命令、门禁和默认生命周期
3. `docs/agent_guidelines/git_policy.md`
   - 定义哪些改动默认要提交，哪些可以不提交
4. `docs/specs/_status.md`
   - 记录正式模块当前状态、当前层和默认下一步命令

执行时不要一次性盲读所有文件，只按当前任务需要读取。
