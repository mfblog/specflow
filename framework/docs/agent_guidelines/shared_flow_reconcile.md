# Shared Appendix 状态收口流程

## 1. Purpose

本流程用于在 Shared Appendix 发生变化后，统一修正所有引用模块的状态与失效过程文件。

它只回答三件事：

1. 哪些模块当前层仍持有与当前 Shared Appendix 真相不一致的旧绑定快照。
2. 这些模块现在应该回退到哪个最小可行动作。
3. 哪些候选侧过程文件必须删除，避免继续沿用失效门禁。

本流程不是普通模块命令，不属于 `{command}:{module}`，也不进入 `docs/specs/_status.md`。
它默认是 Agent 在命中 Shared Appendix 变更收口场景后执行的内部治理流程，不作为面向用户直接暴露的命令词。

## 2. Scope

本流程默认处理：

1. `docs/specs/shared/candidate/` 与 `docs/specs/shared/stable/` 中共享附属展开文件的版本、正文、层级或绑定关系变化。
2. 各模块当前层 `Global Constraint Alignment.shared_appendix_refs` 与其过程文件绑定快照，是否仍和当前 Shared Appendix 状态一致。
3. Shared Appendix 变更后，对那些仍持有旧快照的模块执行 `_status.md`、`_check_result`、`_plans`、`_verify_result` 的统一回退与清理。
4. 报出哪些 Shared Appendix 的 `bound_modules` 仍与真实模块绑定集合不一致，要求回到负责该绑定变更的命令中补齐修正。

本流程默认不处理：

1. 改写业务模块正文。
2. 替代 `spec_flow_review` 做治理机制评审。
3. 替代 `cand_check`、`stable_verify` 或 `cand_promote` 做模块行为审查。

## 3. Preconditions

执行前必须确认：

1. 已读取 `docs/agent_guidelines/spec_policy.md`、`docs/agent_guidelines/command_policy.md` 与 `docs/specs/_status.md`。
2. 已读取 `docs/specs/shared/candidate/`、`docs/specs/shared/stable/` 下当前存在的 Shared Appendix。
3. 已明确本流程的受影响对象来源：
   - 若当前任务实际改动了 `docs/specs/shared/**`，默认按这些 Shared Appendix 文件集合解释“本轮需要复核的 Shared Appendix 集合”
   - 若当前任务没有改 `docs/specs/shared/**`，但改了任一模块当前层 `shared_appendix_refs`，则必须建立“本轮需要复核的模块绑定集合”
   - 若用户明确点名某份 Shared Appendix，则按用户点名集合解释
4. 已按 `_status.md` 的当前 `Active Layer` 读取各正式模块的当前层主文件；若模块当前层 `shared_appendix_refs` 非空，还必须读取对应 Shared Appendix。
5. 若本轮会修改 `_status.md`、`docs/specs/_check_result/*.md`、`docs/specs/_plans/*.md`、`docs/specs/_verify_result/*.md` 或其它命中提交触发条件的对象，执行前必须按 `docs/agent_guidelines/command_policy.md` 第 `10` 节读取 Git 收口规则文件。
6. 若当前任务中的标准命令已经为某个目标模块基于最新 Shared Appendix 真相重算并写回新的过程文件绑定快照，或已把该目标模块直接收口到新的 stable 真相，必须先把该模块标记为“本轮已直接收口的目标模块”，本流程不再对它做二次回退。

补充说明：

1. 用户若表达“shared 改完后还有哪些模块受影响”“这些 shared 绑定改动要不要统一回退”这类自然语言意图，Agent 应先用这些意图确定当前任务范围，再在内部执行本流程。
2. 执行者不得要求用户先理解或输入 `shared_flow_reconcile` 这个内部流程名，才能触发 Shared Appendix 状态收口。

## 4. Procedure

执行步骤固定如下：

1. 建立 Shared Appendix 当前视图：
   - 以 `docs/specs/shared/candidate/` 与 `docs/specs/shared/stable/` 的现存文件为准
   - 记录每个共享对象的 `shared_id`、`layer`、`shared_version`、当前正文指纹与 `bound_modules`
   - 若当前任务实际改动了 `docs/specs/shared/**`，单独标出“本轮需要复核的 Shared Appendix 集合”
2. 建立模块当前层绑定视图：
   - 读取 `_status.md` 中所有正式模块的当前层
   - 读取对应当前层主文件中的 `Global Constraint Alignment.shared_appendix_refs`
   - 以模块当前层 `shared_appendix_refs` 作为正式绑定来源；`bound_modules` 只作声明性辅助
3. 建立模块当前快照视图：
   - 若模块当前层为 `candidate`，读取现存的 `_check_result/{module}.md`、`_plans/{module}.md`、`_verify_result/{module}.md`
   - 若这些文件存在，提取其中的 `shared_appendix_snapshot`
   - 用模块当前层 `shared_appendix_refs` 与当前 Shared Appendix 实际正文，按 `spec_policy.md` 第 `12.1` 节重新生成规范化快照
4. 对每个模块逐一判断 Shared Appendix 绑定是否仍有效：
   - 若模块属于“本轮已直接收口的目标模块”，本流程跳过，不做二次回退
   - 若模块当前层 `shared_appendix_refs=none`，且不属于本轮绑定变化导致的缺失场景，本流程不改动该模块状态
   - 若引用的 Shared Appendix 文件不存在、层级不匹配、版本引用不匹配，或模块当前层与共享对象绑定关系已变化，则视为失效
   - 若模块当前层为 `candidate`，且任一现存过程文件中的 `shared_appendix_snapshot` 与按当前真相重生成的规范化快照不一致，则视为失效
   - 若模块当前层为 `stable`，且当前 stable 绑定指向的 Shared Appendix 真相已变化到足以使“当前仍对齐 stable”的判断失去依据，则视为失效
5. 命中失效的 candidate 模块统一按以下规则收口：
   - 删除 `docs/specs/_check_result/{module}.md`
   - 删除 `docs/specs/_plans/{module}.md`
   - 删除 `docs/specs/_verify_result/{module}.md`
   - 把 `_status.md` 中该模块的 `Next Command` 统一回退为 `cand_check`
   - 保持该模块 `Candidate=yes`、`Active Layer=candidate`
6. 命中失效的 stable 模块统一按以下规则收口：
   - 不生成候选侧过程文件删除动作
   - 把 `_status.md` 中该模块的 `Next Command` 统一回退为 `stable_verify`
   - 保持该模块 `Stable=yes`、`Active Layer=stable`
7. 若 Shared Appendix 的 `bound_modules` 与真实模块绑定集合不一致：
   - 记录为治理漂移
   - 明确指出应由哪一类绑定变更命令回补修正
   - 不由本流程直接改写 Shared Appendix 正文
   - 不单独据此改变模块状态机
8. 若某模块当前 `_status.md` 已经指向比真实最小可行动作更下游的命令，必须由本流程统一纠正，不得只报告“已失效”而不回写状态。
9. 若本轮改动命中 Git 收口规则文件的提交触发条件，必须按该规则判断并完成当前任务内的 git 收口。

## 5. Stop Conditions

1. 所有 Shared Appendix 当前视图与模块绑定视图已完成对照。
2. 所有命中失效的模块都已回退到最小可行动作。
3. 所有候选侧失效过程文件都已清理。
4. 所有 `bound_modules` 与真实绑定集合的不一致都已被明确报告，并已明确回补责任归属。

## 6. Output Contract

输出必须至少包含：

1. Shared Appendix 变化摘要
2. 本轮已直接收口的目标模块列表
3. 受影响模块列表
4. 每个受影响模块的状态回退结果
5. 被删除的过程文件列表
6. `bound_modules` 与真实绑定集合的不一致项
7. Git 收口结果

## 7. Non-Goals

本流程不负责：

1. 为 Shared Appendix 建立独立状态机
2. 直接修改模块正文或 Shared Appendix 正文来“顺手修好”绑定关系
3. 替代 `cand_check` 重新审通过候选收口
4. 替代 `stable_verify` 重新核对 stable 对齐

## 8. Examples

```md
用户表达“我改了 shared 之后，帮我看看还有哪些模块需要回退状态”
```
