# Spec Init Command

## 1. Purpose

本命令用于为**历史模块**补建第一份 `stable` Spec。

目标只有三个：

1. 沉淀当前已经生效的正式行为。
2. 建立该模块的第一份正式真相文件。
3. 把该模块登记进 `docs/specs/_status.md`。

---

## 2. Scope

本命令默认处理：

1. 历史模块首次纳管。
2. 当前已有实现和稳定行为，但还没有进入 Spec 体系的模块。
3. 第一份 `stable` 的建立。

本命令默认不处理：

1. 新模块立项。
2. 从已有 `stable` 派生新候选。
3. 直接建立 `candidate`。

---

## 3. Preconditions

执行前必须确认：

1. 已先完成 `spec_policy.md` 第 `8` 节定义的前置自检；若目标模块尚未登记，则至少要确认不存在与该模块冲突的旧状态文件或残留计划文件。
2. 已明确目标模块名。
3. 该模块尚未纳入 `docs/specs/_status.md`。
4. 当前目标是沉淀现状真相，而不是定义未来设计。
5. 本命令属于首次纳管例外，不要求目标模块已存在既有过程文件或 `stable drift reconciliation` 结果。
6. 若本轮纳管需要同时判断正式技术基线、共享机制或全局例外，执行前还应读取 `docs/specs/system/stable/s_system_constraints.md`；若当前任务只沉淀模块既有正式行为，则不因统一总入口规则而强制固定先读该文件。
7. 若历史模块涉及技术选型、共享基础设施、跨模块复用、全局例外申请或全局约束提案，首版 `stable` 必须补齐 `Global Constraint Alignment` 或等价章节。
8. 若本轮会修改 `stable`、`_status.md` 或其它命中提交触发条件的治理文件，执行前必须按 `docs/agent_guidelines/command_policy.md` 第 `10` 节读取 Git 收口规则文件，确认本轮是否要求提交以及应按哪类提交收口。
9. 若本轮会创建或更新 `shared_appendix_refs`，或会同时创建 / 更新 `docs/specs/shared/**`，执行前必须读取 `docs/agent_guidelines/shared_flow_reconcile.md`，并准备在本命令内同步维护对应 Shared Appendix 的 `bound_modules`，再判断是否还需要对其它受影响模块执行统一状态收口。

若以上条件不满足，不得直接执行。

---

## 4. Procedure

执行步骤固定如下：

1. 梳理模块当前已经生效的行为基线。
2. 若当前任务需要判断正式技术基线、共享机制或全局例外，补读 `docs/specs/system/stable/s_system_constraints.md` 作为该场景输入。
3. 创建 `docs/specs/stable/s_{module}.md`。
4. 确保该文件覆盖正式 Spec 的核心内容：
   - `Context & Motivation`
   - `Terminology`
   - `Data Structures / Protocols`
   - `State Machine / Business Flow`
   - `Edge Cases & Error Handling`
   - `Testability / Acceptance Criteria`
5. 若本模块涉及技术选型、共享基础设施、跨模块复用、全局例外申请或全局约束提案，则补齐 `Global Constraint Alignment` 或等价章节，且至少覆盖：
   - `system_constraints_stable_ref`
     - 若正式全局基线已存在，写对应版本
     - 若正式全局基线尚不存在，写 `none`
   - `shared_appendix_refs`
   - `shared_mechanism_reuse_summary`
   - `global_constraint_exceptions`
   - `proposed_system_constraints_updates`
   - `promotion_to_system_stable`
6. 若第 5 步写入的 `shared_appendix_refs` 非空，或本轮同时创建 / 更新了 `docs/specs/shared/**`，必须在本命令内同步修正受影响 Shared Appendix 的 `bound_modules`。
7. 若第 6 步命中了 Shared Appendix 变化，且还有其它未在本命令内直接收口、但已受影响的模块，必须在宣称本轮状态已收口前执行 `shared_flow_reconcile`。
8. 更新 `docs/specs/_status.md`：
   - `Stable=yes`
   - `Candidate=no`
   - `Active Layer=stable`
   - `Next Command=spec_fork`
9. 若本轮改动命中 Git 收口规则文件的提交触发条件，必须按该规则判断并完成当前任务内的 git 收口。

---

## 5. Stop Conditions

默认停止条件如下：

1. 第一份 `stable` 已生成。
2. `_status.md` 已完成登记。
3. 若命中了 Shared Appendix 变化，Shared Appendix 的 `bound_modules` 已同步维护，且其它受影响模块已完成统一状态收口。
4. 不继续自动开启候选轮次。

---

## 6. Output Contract

输出顺序固定如下：

1. 纳管判断
2. 新建文件路径
3. 本次是否要求写入 `Global Constraint Alignment`，以及触发依据
4. `_status.md` 更新结果
5. Git 收口结果
6. 后续建议

---

## 7. Non-Goals

本命令不负责：

1. 建立第一份 `candidate`
2. 直接进入实现
3. 重新设计现有模块

---

## 8. Examples

### 8.1 为历史模块补建第一份 stable

```md
spec_init:module_example
```
