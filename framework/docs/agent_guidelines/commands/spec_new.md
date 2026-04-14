# Spec New Command

## 1. Purpose

本命令用于为**全新模块**建立第一份 `candidate` Spec。

目标只有三个：

1. 为新模块确定第一版完整候选设计。
2. 建立候选推进起点。
3. 把该模块登记进 `docs/specs/_status.md`。

## 2. Scope

本命令默认处理：

1. 新模块首次立项。
2. 当前还没有正式生效版本的模块。
3. 第一份 `candidate` 的建立。
4. 初始化 `Global Constraint Alignment` 中的 `system_constraints_stable_ref` 与 `shared_appendix_refs`。

## 3. Preconditions

执行前必须确认：

1. 已先完成 `spec_policy.md` 第 `8` 节定义的前置自检；若目标模块尚未登记，则至少要确认不存在与该模块冲突的旧状态文件或残留计划文件。
2. 已明确目标模块名。
3. 该模块尚未纳入 `docs/specs/_status.md`。
4. 当前目标是先建立未来设计，而不是先沉淀现状真相。
5. 若本轮会创建或更新 `shared_appendix_refs`，或会同时创建 / 更新 `docs/specs/shared/**`，执行前必须读取 `docs/agent_guidelines/shared_flow_reconcile.md`，并准备在本命令内同步维护对应 Shared Appendix 的 `bound_modules`，再判断是否还需要对其它受影响模块执行统一状态收口。
6. 若本轮会修改 `_status.md`、治理规则文件或其它命中提交触发条件的对象，执行前必须按 `docs/agent_guidelines/command_policy.md` 第 `10` 节读取 Git 收口规则文件，确认本轮是否要求提交以及应按哪类提交收口。

## 4. Procedure

执行步骤固定如下：

1. 若 `docs/specs/system/stable/s_system_constraints.md` 已存在，读取它作为当前正式全局基线；若尚不存在，则按“当前无正式全局基线”的空态继续。
2. 梳理新模块的目标、边界、协议和主流程。
3. 创建 `docs/specs/candidate/c_{module}.md`。
4. 将该文件的 `frontmatter.version` 初始化为 `0.1.0`。
5. 确保该文件覆盖正式 Spec 的核心内容。
6. 在 `Global Constraint Alignment` 中初始化：
   - 若正式全局基线已存在：`system_constraints_stable_ref=s_system_constraints@<current_version>`
   - 若正式全局基线尚不存在：`system_constraints_stable_ref=none`
   - `shared_appendix_refs=none`
   - `shared_mechanism_reuse_summary`
   - `global_constraint_exceptions`
   - `proposed_system_constraints_updates`
   - `promotion_to_system_stable`
7. 若第 6 步写入的 `shared_appendix_refs` 非空，或本轮同时创建 / 更新了 `docs/specs/shared/**`，必须在本命令内同步修正受影响 Shared Appendix 的 `bound_modules`。
8. 若第 7 步命中了 Shared Appendix 变化，且还有其它未在本命令内直接收口、但已受影响的模块，必须在宣称本轮状态已收口前执行 `shared_flow_reconcile`。
9. 更新 `docs/specs/_status.md`：
   - `Stable=no`
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=cand_check`
10. 若本轮改动命中 Git 收口规则文件的提交触发条件，必须按该规则判断并完成当前任务内的 git 收口。

## 5. Stop Conditions

1. 第一份 `candidate` 已生成。
2. `_status.md` 已完成登记。
3. 若命中了 Shared Appendix 变化，Shared Appendix 的 `bound_modules` 已同步维护，且其它受影响模块已完成统一状态收口。
4. 不继续自动进入实现。

## 6. Output Contract

1. 立项判断
2. 新建文件路径
3. 初始化的 candidate 版本
4. 初始化的正式全局基线引用或 `none`
5. `_status.md` 更新结果
6. Git 收口结果
7. 待收口问题

## 7. Non-Goals

1. 生成第一份正式 `stable`
2. 沉淀历史行为
3. 自动进入 `cand_impl`
4. 创建独立的 `system_constraints` candidate 文件

## 8. Examples

```md
spec_new:module_executor
```
