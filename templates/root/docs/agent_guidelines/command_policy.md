# 命令规范（Command Policy）

## 1. Purpose

本文件定义本仓库中标准命令的工作方式。

它要回答四件事：

1. 什么是命令。
2. 命令操作哪些对象。
3. 不同命令各自负责什么。
4. Agent 在收到命令型请求时应如何命中和执行。

---

## 2. What a Command Is

命令是 Agent 的标准工作流入口。  
它不是 Shell 命令，也不是业务规则本身。

通俗讲：

1. `Spec` 是真相。
2. `Command` 是动作。

---

## 3. Objects Operated by Commands

命令默认会操作以下对象：

1. `stable`
2. `candidate`
3. `plan`
4. `check_result`
5. `verify_result`
6. `status`

补充说明：

1. 这里的 `status` 固定指 `docs/specs/_status.md`；它是状态索引文件，不是行为真相源，但所有标准命令都必须按规则同步维护它。
2. `system_constraints` 是全局唯一的系统级约束对象，不属于普通命令直接操作的六类过程对象。
2A. `shared appendix` 是共享附属展开对象，不属于普通命令直接操作的六类过程对象，也不是独立命令目标。
3. 当模块命令需要判断正式技术基线、共享机制或全局例外，且仓库中已存在 `docs/specs/system/stable/s_system_constraints.md` 时，应读取它作为该场景的上游约束输入。
4. `system_constraints` 不进入 `docs/specs/_status.md`，也不生成自己的 `_check_result`、`_plans`、`_verify_result`。
5. 模块若需要提出新的全局约束变化，只能把提案写在自己的 candidate 中。
6. `s_system_constraints.md` 只允许在模块 `cand_promote` 时作为联动副产品被首次创建或后续更新。

---

## 4. Command Format

标准命令格式如下：

```text
{command}:{module}
```

其中：

1. `{command}` 是稳定命令名。
2. `{module}` 是正式模块名，例如 `module_example`。

补充规则：

1. `{module}` 必须使用正式模块名，不得使用文件名前缀。
2. `system_constraints` 不是合法命令目标，不得写成 `spec_new:system_constraints`、`cand_check:system_constraints` 等形式。
3. `{module}` 默认只指向已被机制承认为“正式模块”的对象；附录、专题展开文件、Prompt 原文模板等附属文件不是合法命令目标。
3A. `shared appendix` 也不是合法命令目标；执行者不得写成 `cand_check:shared_xxx` 或等价形式。
3B. `shared_flow_reconcile` 不是 `{command}:{module}` 形式的标准模块命令；执行者不得写成 `shared_flow_reconcile:module_xxx`。
3C. `shared_extract_review` 也不是 `{command}:{module}` 形式的标准模块命令；执行者不得写成 `shared_extract_review:module_xxx`。
4. 对 `spec_init:{module}` 与 `spec_new:{module}` 这两个首版入口，`{module}` 允许指向“尚未进入 `_status.md`、但模块名已明确且不存在冲突”的新目标。
5. 除上一条首版入口例外外，某个文件若尚未作为独立正式模块进入 `_status.md`，则不得仅因文件名、路径或 frontmatter 看起来像模块，就被当成 `{module}` 命令目标。

---

## 5. Standard Commands

### 5.1 Version Commands

1. `spec_init:{module}`
2. `stable_verify:{module}`
3. `spec_new:{module}`
4. `spec_fork:{module}`
5. `cand_promote:{module}`

### 5.2 Candidate Commands

1. `cand_check:{module}`
2. `cand_plan:{module}`
3. `cand_impl:{module}`
4. `cand_verify:{module}`

补充要求：

1. 任何面向执行者的命令索引文档，都必须完整列出以上标准命令，不得遗漏 `stable_verify` 这类非候选侧命令。
2. 若命令索引与本节不一致，以本节和对应命令文件为准，并应在当前任务内修正文档漂移。
3. “命令索引文档”按职责定义，但其默认登记集合以 `docs/agent_guidelines/entry_index_registry.md` 为准；新增同职责入口文件时，必须先更新登记表。

---

## 6. Responsibilities of Each Command Type

### 6.1 Version Commands

版本命令负责建立、开启或切换版本层。

1. `spec_init`
   - 为历史模块建立第一份 `stable`
2. `stable_verify`
   - 核对当前代码是否仍对齐 `stable`
   - 输出结构化验证证据，并更新 `_status.md`
3. `spec_new`
   - 为新模块建立第一份 `candidate`
4. `spec_fork`
   - 从现有 `stable` 开启新一轮 `candidate`
5. `cand_promote`
   - 让当前 `candidate` 提升为新的 `stable`
   - 在需要时联动更新 `s_system_constraints.md`

### 6.2 Candidate Commands

候选命令负责把 candidate 从设计推进到实现，再推进到提升。

1. `cand_check`
   - 检查 candidate Spec 是否已收口到足以稳定约束实施
   - 检查当前 `system_constraints_stable_ref` 是否与正式全局基线状态一致
   - 对命中 Prompt 触发条件的模块执行 Prompt Adequacy Review
   - 若通过，则产出 `_check_result/{module}.md` 作为当前 candidate 进入后续候选链的放行凭证
2. `cand_plan`
   - 读取 `_check_result/{module}.md` 放行凭证
   - 生成或更新当前轮 `_plans/{module}.md`
3. `cand_impl`
   - 在确认当前 candidate 仍保有有效 `_check_result/{module}.md` 后，依据 candidate Spec 与 `_plans/{module}.md` 推进实现
4. `cand_verify`
   - 在确认当前 candidate 仍保有有效 `_check_result/{module}.md` 与 `_plans/{module}.md` 后，核对实现是否对齐 candidate Spec
   - 产出 `_verify_result/{module}.md`

---

## 7. Default Lifecycle Order

正式模块的默认命令顺序分为两条主链：

1. `stable` 维护链
   - `spec_init`
   - `stable_verify`
   - `spec_fork`
2. `candidate` 升级链
   - `spec_new`
   - `spec_fork`
   - `cand_check`
   - `cand_plan`
   - `cand_impl`
   - `cand_verify`
   - `cand_promote`

---

## 8. Gate Rules

以下规则属于统一门禁，所有命令默认都要遵守：

1. 未通过前置自检，不得执行对应命令。
2. 未通过 `cand_check`，不得进入 `cand_plan`。
3. 当前 candidate 若不存在有效 `_check_result/{module}.md`，不得进入 `cand_plan`、`cand_impl` 或 `cand_verify`。
4. 不存在有效 `_plans/{module}.md`，不得进入 `cand_impl` 或 `cand_verify`。
5. 未完成 `cand_verify` 或仍有阻塞项，不得执行 `cand_promote`。
6. `Active Layer=stable` 时，若发生实现改动但未做 `stable_verify`，不得宣称“仍对齐正式版”。
7. 对 `Active Layer=stable` 的模块，执行 `stable_verify`、`spec_fork`、以及任何声明“当前仍对齐 stable”的工作前，必须先完成 `stable drift reconciliation`。
8. `Next Command` 是默认允许推进的下一个动作；若与当前命令不一致，默认不得越过。
   - `_status.md` 的对象职责、字段语义与读取规则以 `spec_policy.md` 为准；本文件只消费 `Next Command` 的命令层含义。
9. 过程文件是否有效，不得只靠文件存在判断；必须校验绑定的 Spec 层、Spec 文件、版本引用、指纹以及命令期望字段。
10. 所有模块 candidate 都必须显式记录 `system_constraints_stable_ref`。
10A. 若模块当前层行为依赖共享附属展开文件，则该模块当前层 Spec 还必须显式记录 `shared_appendix_refs`。
11. 若 `docs/specs/system/stable/s_system_constraints.md` 已存在，而当前模块 candidate 中的 `system_constraints_stable_ref` 不等于当前 `s_system_constraints` 版本，则该模块的候选侧过程文件默认失效，并统一回退为 `cand_check`。
12. 若 `docs/specs/system/stable/s_system_constraints.md` 尚不存在，而当前模块 candidate 中的 `system_constraints_stable_ref` 不等于 `none`，则该模块的候选侧过程文件默认失效，并统一回退为 `cand_check`。
12A. 若模块当前层 `shared_appendix_refs` 绑定的共享附属展开文件版本、正文或绑定关系已变化，则该模块候选侧过程文件默认失效，并统一回退为 `cand_check`。
12B. 若 `Active Layer=stable` 的模块当前层 `shared_appendix_refs` 绑定的 stable 共享附属展开文件版本、正文或绑定关系已变化，则该模块不得继续宣称“仍对齐 stable”，并统一回退为 `stable_verify`。
13. `cand_verify` 不负责推进 `system_constraints` 的独立状态机；它只负责验证模块实现是否已对齐当前 candidate 体系。
14. `cand_promote` 若确认本轮模块提升同时带着已收口的全局约束提案，应同步把该模块 candidate 中的提案吸收到 `docs/specs/system/stable/s_system_constraints.md`。
15. `cand_plan`、`cand_impl` 与 `cand_verify` 读取 `_check_result/{module}.md` 时，不得只看文件存在；必须同时满足对应命令所要求的绑定关系以及 `decision=pass`、`allow_next=true`。
16. `cand_promote` 读取 `_verify_result/{module}.md` 时，不得只看文件存在；必须同时满足 `decision=pass`、`allow_next=true`、`next_command=cand_promote`。
17. 命中 Prompt 触发条件的 candidate，未通过 Prompt Adequacy Review，不得进入 `cand_plan`，也不得继续沿用旧的候选侧放行凭证推进下游命令。
18. `Prompt Adequacy Review` 的 `n/a` 只允许在未命中 Prompt 触发条件时出现；一旦命中，必须明确给出 `pass|blocked|fix_required`。
19. `cand_check` 未通过时，不得写入失败态 `_check_result/{module}.md`；若旧的 pass gate 已不再成立，必须删除该文件并把 `Next Command` 保持或回退为 `cand_check`。
20. `cand_check` 默认不负责直接修改 candidate 真相文件；唯一允许的自动修正项是：当 candidate 与当前正式全局基线状态仍兼容时，只前移 `system_constraints_stable_ref`，或在“尚无正式全局基线”场景下把它纠正为 `none`。除此之外，任何 candidate 真相修改都必须先由人类或上游命令完成，再重新执行 `cand_check`。

---

## 9. Command File Contract

每个命令文件默认必须包含以下章节：

1. `Purpose`
2. `Scope`
3. `Preconditions`
4. `Procedure`
5. `Stop Conditions`
6. `Output Contract`
7. `Non-Goals`
8. `Examples`

附加要求：

1. 命令文件必须明确说明自己操作的是哪个对象。
2. 命令文件必须明确说明自己的上游前提。
3. 命令文件必须明确说明停止条件。
4. 命令文件不得把 candidate 真相与 `_plans`、`_check_result`、`_verify_result` 的职责混成同一个对象。
5. 若命令文件会消费 `system_constraints`，必须明确它是上游约束输入，不是当前命令的主要产物。
5A. 若命令文件会消费 `shared appendix`，必须明确它们是被模块绑定带入的共享真相对象，而不是独立命令目标。
6. 命令文件若涉及生命周期、回退与清理，不得自创另一套总规则；对象级闭环关系以 `spec_policy.md` 第 `6` 节为准。
7. 若命令文件涉及 Prompt 门禁，必须明确：
   - Prompt 触发条件
   - 审查维度
   - 阻塞标准
   - KV cache 友好排序与语义清晰之间的优先级关系
8. 若命令文件要求回写 Prompt 审查结果，必须明确快照字段的最小契约，不得让执行者自行发明字段含义。
9. 若命令文件完成后还存在强制性的收口动作，例如必须判断是否提交 git 历史，则必须明确引用对应治理规则，不得把这类动作留给执行者自行回忆。

---

## 10. Auxiliary Governance Rules

除模块命令本身外，以下治理规则在命中特定场景时也属于标准执行链的一部分：

术语约定：

1. 本文及各标准命令文件后续提到的“Git 收口规则文件”，固定指 `docs/agent_guidelines/git_policy.md`。
2. 若命令文件写“必须读取 Git 收口规则文件”，其含义固定等于：必须读取并执行 `docs/agent_guidelines/git_policy.md`，判断本轮改动是否要求纳入 git 历史，以及应按哪类提交收口。

1. `docs/agent_guidelines/git_policy.md`
   - 当任务涉及实现、审查、升级、`cand_promote`，或修改 `docs/agent_guidelines/*.md`、`docs/agent_guidelines/commands/*.md`、`docs/agent_guidelines/entry_index_registry.md`、已登记入口索引文件、`docs/specs/` 中非 `candidate` 真相文件时，必须读取并执行。
   - 它回答的是“当前改动是否必须纳入 git 历史，以及应按哪类提交收口”，不是可选建议。
2. `docs/agent_guidelines/shared_flow_reconcile.md`
   - 当任务修改 `docs/specs/shared/**`，或修改任一模块当前层 `shared_appendix_refs` 时，必须完成一次 Shared Appendix 状态收口。
   - 它是 Agent 在命中 Shared Appendix 变更收口场景后执行的内部治理流程，不作为面向用户直接暴露的命令词。
   - 若当前标准命令已经基于最新 Shared Appendix 真相，为当前目标模块重算并写回了新的过程文件绑定快照，或已把当前目标模块直接收口到新的 stable 真相，则该目标模块视为已在当前命令内完成收口，不再由 `shared_flow_reconcile` 二次回退。
   - 对其余未在当前命令内直接收口、但已受 Shared Appendix 变化影响的模块，必须执行 `shared_flow_reconcile` 做统一回退和失效过程文件清理。
   - 它回答的是“Shared Appendix 变化后哪些模块状态与过程文件需要统一回退和清理”，不是治理评审流程。
3. `docs/agent_guidelines/shared_extract_review.md`
   - 当用户主动要求判断某段内容是否该提取为 shared，或模块命令发现 shared 候选信号且用户同意扩展审查范围时，应执行。
   - 它回答的是“某段内容是否已达到 shared 提取标准，以及当前是否应提取为 `c_shared_xxx`”，不是模块状态推进流程。
4. 命令文件若与这些辅助治理规则的触发条件不一致，以辅助治理规则和本文件的最新联动规则为准，并应在当前任务内修正文档漂移。

---

## 11. Command Matching Rules

当用户请求明显命中命令时，Agent 应优先按命令执行。

补充说明：

1. 本节只定义标准模块命令 `{command}:{module}` 及其自然语言命中规则。
2. `spec_flow_review` 不属于普通模块命令，因此不纳入本节的命中表，也不按 `{command}:{module}` 解析。
3. 若仓库内入口索引文件把请求路由到 `spec_flow_review`，执行者应直接转入 `docs/agent_guidelines/spec_flow_review.md`；该流程的待审范围与判定规则仍只由其自身和仓库内正式治理文件定义。
4. `shared_flow_reconcile` 也不属于普通模块命令，因此不纳入本节的命中表，也不按 `{command}:{module}` 解析。
5. 当用户表达“shared 改完后还有哪些模块受影响”“这些 shared 绑定改动要不要统一回退”这类自然语言意图时，执行者应把它识别为 Shared Appendix 状态收口场景，再在内部转入 `docs/agent_guidelines/shared_flow_reconcile.md`；不得要求用户直接输入内部流程名。
6. `shared_extract_review` 也不属于普通模块命令，因此不纳入本节的命中表，也不按 `{command}:{module}` 解析。
7. 若请求明确命中“这段是否应提取为 shared / 是否该抽成公共部分 / 是否需要 shared 边界审查”，执行者应直接转入 `docs/agent_guidelines/shared_extract_review.md`；该流程只负责 shared 提取边界判定，不替代模块命令。

默认命中规则如下：

1. 用户明确写出命令名时，直接命中对应命令。
2. 用户若只说“候选收口审查”，默认命中 `cand_check`。
3. 用户若只说“做实现计划”，默认命中 `cand_plan`。
4. 用户若只说“按 candidate 实现”，默认命中 `cand_impl`。
5. 用户若只说“核对 candidate 是否已实现”，默认命中 `cand_verify`。
6. 用户若只说“核对 stable 是否仍然对齐”，默认命中 `stable_verify`。
