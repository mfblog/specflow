# Spec Fork Command

## 1. Purpose

本命令用于从已有 `stable` 派生一份新的 `candidate` Spec。

目标只有三个：

1. 以当前正式版为基线开启新一轮候选设计。
2. 为后续 `cand_check / cand_plan / cand_impl` 提供候选真相。
3. 同步更新 `docs/specs/_status.md`。

## 2. Scope

本命令默认处理：

1. 已有 `stable` 模块的新一轮升级。
2. 从正式真相派生候选真相。
3. 初始化当前轮 candidate 的 `system_constraints_stable_ref`。

## 3. Preconditions

执行前必须确认：

1. 已先完成 `spec_policy.md` 第 `8` 节定义的前置自检，且不存在状态漂移或文件漂移。
2. 已先完成 `stable drift reconciliation`。
3. `_status.md` 中该模块的 `Next Command=spec_fork`。
4. 该模块已经存在 `stable`。
5. 若 `s_{module}.md` 正文明确引用了 stable 层附属展开文件，执行前必须一并读取这些文件。
6. 若模块当前 stable 层 `Global Constraint Alignment.shared_appendix_refs` 非空，执行前还必须一并读取这些 Shared Appendix；它们是被模块绑定带入的共享真相对象，不是独立命令目标。
7. 若本轮会修改 `_status.md`、清理过程文件或其它命中提交触发条件的对象，执行前必须按 `docs/agent_guidelines/command_policy.md` 第 `10` 节读取 Git 收口规则文件，确认本轮是否要求提交以及应按哪类提交收口。
8. 若前置自检或读取过程中发现被引用的附属展开文件发生目录漂移，当前命令必须先完成迁移并重新执行前置自检，再继续派生 candidate。
9. 若本轮会改写当前模块 candidate 层的 `shared_appendix_refs`，或会同时创建 / 更新 `docs/specs/shared/**`，执行前必须读取 `docs/agent_guidelines/shared_flow_reconcile.md`，并准备在本命令内同步维护对应 Shared Appendix 的 `bound_modules`，再判断是否还需要对其它受影响模块执行统一状态收口。

## 4. Procedure

1. 若 `docs/specs/system/stable/s_system_constraints.md` 已存在，读取它作为当前正式全局基线；若尚不存在，则按“当前无正式全局基线”的空态继续。
2. 读取 `docs/specs/stable/s_{module}.md`；若该文件明确引用了 stable 层附属展开文件，必须一并读取。
3. 若模块当前 stable 层 `Global Constraint Alignment.shared_appendix_refs` 非空，必须一并读取这些 Shared Appendix；它们属于当前 stable 层真相读取面。
4. 若在第 2-3 步或前置自检中发现附属展开文件目录漂移，必须先由当前命令完成迁移并重新执行前置自检；在迁移完成前不得继续派生。
5. 先按版本规则确定“本轮目标正式版本”：
   - 兼容性新增能力：默认取当前 `stable` 的下一个 `MINOR`
   - 不兼容变化：默认取当前 `stable` 的下一个 `MAJOR`
   - 兼容性修正或对齐：默认取当前 `stable` 的下一个 `PATCH`
6. 以该文件为基线生成 `docs/specs/candidate/c_{module}.md`。
7. 把 candidate 的 `frontmatter.version` 写为本轮目标正式版本。
8. 在 `Global Constraint Alignment` 中写入当前 `system_constraints_stable_ref`，并重新核对 `shared_appendix_refs`：
   - 若正式全局基线已存在，写当前版本
   - 若正式全局基线尚不存在，写 `none`
   - 若当前 stable 仍显式绑定了 `s_shared_xxx@...`，且本轮 candidate 仍计划继续依赖该共享对象，则必须先派生出对应 `c_shared_xxx@...`，再把 candidate 写成该 candidate 层引用
   - 若当前轮不再复用共享附属展开文件，或当前轮尚未形成可读的 candidate 层共享附属展开文件，则必须显式写 `shared_appendix_refs=none`，不得直接沿用 stable 层引用
9. 只要第 8 步新增、删除或切换了 Shared Appendix 绑定，或本轮同时创建 / 更新了 `docs/specs/shared/**`，必须在本命令内同步修正受影响 Shared Appendix 的 `bound_modules`。
10. 删除旧的 `_check_result/{module}.md`、`_verify_result/{module}.md`、`_plans/{module}.md`，以及该模块上一轮 candidate 附属展开文件。
11. 若第 9 步命中了 Shared Appendix 变化，且还有其它未在本命令内直接收口、但已受影响的模块，必须在宣称本轮状态已收口前执行 `shared_flow_reconcile`。
12. 更新 `docs/specs/_status.md`：
   - `Stable=yes`
   - `Candidate=yes`
   - `Active Layer=candidate`
   - `Next Command=cand_check`
13. 若本轮改动命中 Git 收口规则文件的提交触发条件，必须按该规则判断并完成当前任务内的 git 收口。

## 5. Stop Conditions

1. 新一轮 `candidate` 已生成。
2. 上一轮过程文件已清理。
3. 若命中了 Shared Appendix 变化，Shared Appendix 的 `bound_modules` 已同步维护，且其它受影响模块已完成统一状态收口。
4. `_status.md` 已同步更新。

## 6. Output Contract

1. 派生判断
2. 新建文件路径
3. 初始化的 candidate 版本
4. 写入的正式全局基线引用或 `none`
5. 清理结果
6. Git 收口结果
7. `_status.md` 更新结果

## 7. Non-Goals

1. 直接修改 `stable`
2. 直接生成 `plan`
3. 直接进入实现
4. 创建独立的 `system_constraints` candidate 文件

## 8. Examples

```md
spec_fork:module_example
```
