# Candidate Implement Command

## 1. Purpose

本命令用于按 `candidate` 与 `_plans/{module}.md` 推进代码实现。

## 2. Scope

本命令默认处理：

1. 按计划推进实现。
2. 补充必要测试或验证动作。
3. 回写 `_plans/{module}.md` 中的进度状态。

## 3. Preconditions

执行前必须确认：

1. 已先完成前置自检，且不存在状态漂移或文件漂移。
2. `_status.md` 中该模块的 `Next Command=cand_impl`。
3. 已存在当前有效的 `docs/specs/_check_result/{module}.md`；这里的“当前有效”必须满足 `spec_policy.md` 第 `12.6` 节定义的完整绑定条件。
4. 已存在当前有效的 `docs/specs/_plans/{module}.md`；这里的“当前有效”必须满足 `spec_policy.md` 第 `12.8` 节定义的完整绑定条件。
5. 当前 candidate 中的 `system_constraints_stable_ref` 与当前正式全局基线状态一致。
6. 若 `c_{module}.md` 正文明确引用了 candidate 层附属展开文件，执行前必须一并读取这些文件。
7. 若模块当前层 `Global Constraint Alignment.shared_appendix_refs` 非空，执行前还必须一并读取这些 Shared Appendix；它们是被模块绑定带入的共享真相对象，不是独立命令目标。
8. `docs/specs/_check_result/{module}.md` 中必须同时满足：
   - `module` 等于目标正式模块名
   - `gate=cand_check`
   - `decision=pass`
   - `allow_next=true`
   - `next_command=cand_plan`
   - `spec_layer_ref=candidate`
   - `spec_file_ref` 等于当前 candidate 文件路径
   - `spec_version_ref` 等于当前 candidate 的版本引用
   - `spec_fingerprint` 等于当前 candidate 的指纹
   - 当前正式全局约束存在时，`system_constraints_stable_file_ref` 等于当前正式全局约束文件路径；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，`system_constraints_stable_version_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，`system_constraints_stable_fingerprint` 等于当前正式全局约束指纹；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，当前 candidate 中的 `system_constraints_stable_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
   - 当前 candidate 当前层 `shared_appendix_refs=none` 时，`shared_appendix_snapshot=none`
   - 当前 candidate 当前层 `shared_appendix_refs` 非空时，`shared_appendix_snapshot` 精确等于按 `spec_policy.md` 第 `12.1` 节规则从当前绑定 Shared Appendix 重新生成的规范化快照
9. `docs/specs/_plans/{module}.md` 中必须同时满足：
   - `spec_file_ref` 等于当前 candidate 文件路径
   - `spec_version_ref` 等于当前 candidate 的版本引用
   - `spec_fingerprint` 等于当前 candidate 的指纹
   - 当前正式全局约束存在时，`system_constraints_stable_file_ref` 等于当前正式全局约束文件路径；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，`system_constraints_stable_version_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，`system_constraints_stable_fingerprint` 等于当前正式全局约束指纹；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，当前 candidate 中的 `system_constraints_stable_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
   - 当前 candidate 当前层 `shared_appendix_refs=none` 时，`shared_appendix_snapshot=none`
   - 当前 candidate 当前层 `shared_appendix_refs` 非空时，`shared_appendix_snapshot` 精确等于按 `spec_policy.md` 第 `12.1` 节规则从当前绑定 Shared Appendix 重新生成的规范化快照
10. 执行前必须按 `docs/agent_guidelines/command_policy.md` 第 `10` 节读取 Git 收口规则文件，确认本轮实现改动是否要求提交以及应按哪类提交收口。
11. 若前置自检或读取过程中发现被引用的附属展开文件发生目录漂移，当前命令必须先完成迁移并重新执行前置自检，再继续实现。

## 4. Procedure

1. 读取 `docs/specs/candidate/c_{module}.md`；若该文件明确引用了 candidate 层附属展开文件，必须一并读取。
2. 若模块当前层 `Global Constraint Alignment.shared_appendix_refs` 非空，必须一并读取这些 Shared Appendix；它们属于当前 candidate 层真相读取面。
3. 若 `docs/specs/system/stable/s_system_constraints.md` 已存在，读取它；若尚不存在，则按“当前无正式全局基线”的空态继续。
4. 读取 `docs/specs/_check_result/{module}.md`。
5. 读取 `docs/specs/_plans/{module}.md`。
6. 若在第 1-2 步或前置自检中发现附属展开文件目录漂移，必须先由当前命令完成迁移并重新执行前置自检；在迁移完成前不得继续推进实现。
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
9. 若上述任一绑定条件不满足，必须立刻停止当前命令，并把 `_status.md` 中该模块的 `Next Command` 回退为 `cand_check`；不得继续沿用过期放行凭证或过期计划推进实现。
10. 若发现当前 candidate 中的 `system_constraints_stable_ref` 与当前正式全局基线状态不一致，必须立刻回退为 `cand_check`，不得继续实现。
11. 只有当放行凭证与计划文件都仍有效，且 candidate 仍对齐当前正式全局基线状态时，才允许继续执行实现。
12. 按计划中的顺序推进代码实现。
13. 运行必要验证，或明确记录哪些验证无法执行。
14. 回写 `docs/specs/_plans/{module}.md` 中的完成状态、阻塞项和验证结果。
15. 根据本轮实现结果更新 `_status.md`：
   - 若计划内实现已达到可验证范围，则 `Next Command=cand_verify`
   - 若仍有实现阻塞，则保持 `Next Command=cand_impl`
   - 若实现过程中确认 candidate 或正式全局基线已变化到需要重新收口，必须回退为 `cand_check`
16. 若本轮改动命中 Git 收口规则文件的提交触发条件，必须按该规则判断并完成当前任务内的 git 收口。

## 5. Stop Conditions

1. 当前轮计划已按可行范围推进。
2. 计划文件已回写。
3. `_status.md` 已回退或推进到真实可执行的下一步动作。
4. 若放行凭证或计划文件绑定失效，当前命令已停止，且 `_status.md` 已回退为 `cand_check`。

## 6. Output Contract

1. 实现摘要
2. 验证结果
3. 计划回写结果
4. 若放行凭证或计划文件失效，必须明确输出失效原因与 `_status.md` 回退结果
5. Git 收口结果
6. `_status.md` 更新结果

## 7. Non-Goals

1. 重写 candidate 真相
2. 在实现阶段修改 `system_constraints`
3. 读取独立的 `system_constraints` candidate 文件

## 8. Examples

```md
cand_impl:module_example
```
