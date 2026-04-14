# Candidate Verify Command

## 1. Purpose

本命令用于核对当前代码是否已对齐指定模块的 `candidate` Spec。

## 2. Scope

本命令默认处理：

1. candidate 与当前代码的一致性核对。
2. 已完成项、偏差项和阻塞项梳理。
3. 状态推进判断。
4. 若模块 candidate 中存在 `proposed_system_constraints_updates`，则核对其中属于本模块实现范围的内容是否已落地。

## 3. Preconditions

执行前必须确认：

1. 已先完成前置自检，且不存在状态漂移或文件漂移。
2. `_status.md` 中该模块的 `Next Command=cand_verify`。
3. 该模块存在有效的 `candidate`。
4. 已存在当前有效的 `docs/specs/_check_result/{module}.md`；这里的“当前有效”必须满足 `spec_policy.md` 第 `12.6` 节定义的完整绑定条件。
5. 已存在当前有效的 `docs/specs/_plans/{module}.md`；这里的“当前有效”必须满足 `spec_policy.md` 第 `12.8` 节定义的完整绑定条件。
6. 当前 candidate 中的 `system_constraints_stable_ref` 必须匹配当前正式全局基线状态，否则回退到 `cand_check`。
7. 若 `c_{module}.md` 正文明确引用了 candidate 层附属展开文件，执行前必须一并读取这些文件。
8. 若模块当前层 `Global Constraint Alignment.shared_appendix_refs` 非空，执行前还必须一并读取这些 Shared Appendix；它们是被模块绑定带入的共享真相对象，不是独立命令目标。
9. 若本轮会修改 `_verify_result/{module}.md`、`_status.md` 或其它命中实现 / 审查 / 升级提交触发条件的对象，执行前必须按 `docs/agent_guidelines/command_policy.md` 第 `10` 节读取 Git 收口规则文件，确认本轮是否要求提交以及应按哪类提交收口。
10. 若前置自检或读取过程中发现被引用的附属展开文件发生目录漂移，当前命令必须先完成迁移并重新执行前置自检，再继续验证。

## 4. Procedure

1. 读取 `docs/specs/candidate/c_{module}.md`；若该文件明确引用了 candidate 层附属展开文件，必须一并读取；若存在 `stable`，补读 `docs/specs/stable/s_{module}.md`，且若 stable 主文件明确引用了 stable 层附属展开文件，也必须一并读取。
2. 若模块当前层 `Global Constraint Alignment.shared_appendix_refs` 非空，必须一并读取这些 Shared Appendix；它们属于当前 candidate 层真相读取面。
3. 若 `docs/specs/system/stable/s_system_constraints.md` 已存在，读取它；若尚不存在，则按“当前无正式全局基线”的空态继续。
4. 读取 `docs/specs/_check_result/{module}.md`。
5. 读取 `docs/specs/_plans/{module}.md`。
6. 若在第 1-2 步或前置自检中发现附属展开文件目录漂移，必须先由当前命令完成迁移并重新执行前置自检；在迁移完成前不得继续执行验证。
7. 按 `spec_policy.md` 第 `12.6` 节逐项校验该放行凭证的完整绑定关系，至少包括：
   - `module`
   - `gate=cand_check`
   - `decision=pass`
   - `allow_next=true`
   - `next_command=cand_plan`
   - `spec_layer_ref=candidate`
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `system_constraints_stable_file_ref`
   - `system_constraints_stable_version_ref`
   - `system_constraints_stable_fingerprint`
   - `shared_appendix_snapshot`
   - 当前 candidate 中的 `system_constraints_stable_ref`
8. 按 `spec_policy.md` 第 `12.8` 节逐项校验该计划文件的完整绑定关系，至少包括：
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `system_constraints_stable_file_ref`
   - `system_constraints_stable_version_ref`
   - `system_constraints_stable_fingerprint`
   - `shared_appendix_snapshot`
   - 当前 candidate 中的 `system_constraints_stable_ref`
9. 若上述任一绑定条件不满足，必须立刻停止当前命令，并把 `_status.md` 中该模块的 `Next Command` 回退为 `cand_check`；不得继续跳过计划门禁直接执行验证。
10. 核对当前代码是否满足 candidate 中的关键协议、主流程、错误处理和验收标准。
11. 生成结构化验证证据矩阵。
12. 输出 `Coverage Summary`。
13. 对所有偏差项按 `spec_policy.md` 中统一定义的 `P1 / P2 / P3` 分级。
14. 给出推进结论：
   - 存在 `fail`，则不得进入 `cand_promote`
   - 存在 `partial` 或 `not_checked` 时，只有满足 `spec_policy.md` 中的非阻塞验证证据降级规则，才允许继续判断是否可进入 `cand_promote`
   - 若关键偏差已清空，且验证证据完整，则可进入 `cand_promote`
15. 更新 `docs/specs/_verify_result/{module}.md` 中的 `Verify Result Snapshot`，至少记录：
   - `gate=cand_verify`
   - `decision=pass|blocked`
   - `allow_next=true|false`
   - `next_command=cand_promote|cand_verify|cand_impl|cand_check`
   - `blocking_summary=...`
   - `coverage_summary={total,covered,failed,partial,not_checked}`
   - `spec_layer_ref=candidate`
   - `spec_file_ref=docs/specs/candidate/c_{module}.md`
   - `spec_version_ref=...`
   - `spec_fingerprint=...`
   - `verification_scope_ref=...`
   - `system_constraints_stable_file_ref=docs/specs/system/stable/s_system_constraints.md|none`
   - `system_constraints_stable_version_ref=...|none`
   - `system_constraints_stable_fingerprint=...|none`
   - `shared_appendix_snapshot=...|none`
16. 更新 `_status.md`：
   - 可提升时：`Next Command=cand_promote`
   - 若实现有偏差但 candidate 仍成立：`Next Command=cand_impl`
   - 若 candidate 或正式全局基线需要改写后才能继续：`Next Command=cand_check`
   - 若仅是验证证据暂不完整：允许保持 `Next Command=cand_verify`
17. 若本轮改动命中 Git 收口规则文件的提交触发条件，必须按该规则判断并完成当前任务内的 git 收口。

## 5. Stop Conditions

1. 已明确 candidate 与当前代码是否对齐。
2. 已明确下一步是否可进入 `cand_promote`。
3. `_status.md` 已同步更新到真实可执行的下一步动作。
4. 若放行凭证或计划文件失效，当前命令已停止，且 `_status.md` 已回退为 `cand_check`。

## 6. Output Contract

1. 核对结论
2. 结构化验证证据矩阵
3. `Coverage Summary`
4. `Verify Result Snapshot` 回写结果
5. 偏差清单
6. 若放行凭证或计划文件失效，必须明确输出失效原因与 `_status.md` 回退结果
7. 下一步建议
8. Git 收口结果
9. `_status.md` 更新结果

补充要求：

1. `P1 / P2 / P3` 的语义以 `spec_policy.md` 中的统一定义为准，不得在本命令内另起一套解释。
2. 任何被判为非阻塞的 `partial` 或 `not_checked`，都必须满足 `spec_policy.md` 中的统一降级规则，并明确写出理由与残余风险。

## 7. Non-Goals

1. 直接改代码
2. 直接重写 candidate
3. 推进 `system_constraints` 的独立状态机

## 8. Examples

```md
cand_verify:module_example
```
