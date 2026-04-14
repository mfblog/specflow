# Candidate Plan Command

## 1. Purpose

本命令用于为指定模块生成或更新当前轮实施计划。

它消费的是 `cand_check` 通过后生成的 `_check_result` 放行凭证，而不是失败审查报告。
它不负责兼容历史遗留的多种 gate 写法；仓库中的 `_check_result` 必须先迁移到统一格式。

## 2. Scope

本命令默认处理：

1. 当前轮候选实施范围拆解。
2. 需要跟随变化的代码区域识别。
3. 验证重点、风险与阻塞项梳理。
4. `_plans/{module}.md` 的生成与初始化。

## 3. Preconditions

执行前必须确认：

1. 已先完成前置自检，且不存在状态漂移或文件漂移。
2. `_status.md` 中该模块的 `Next Command=cand_plan`。
3. `docs/specs/_check_result/{module}.md` 已存在且仍有效；这里的“仍有效”不是抽象描述，必须满足 `docs/specs/_check_result/README.md` 与 `spec_policy.md` 第 `12.6` 节定义的完整绑定条件。
4. 当前 candidate 中的 `system_constraints_stable_ref` 与当前正式全局基线状态一致。
5. 若 `c_{module}.md` 正文明确引用了 candidate 层附属展开文件，执行前必须一并读取这些文件。
6. 若模块当前层 `Global Constraint Alignment.shared_appendix_refs` 非空，执行前还必须一并读取这些 Shared Appendix；它们是被模块绑定带入的共享真相对象，不是独立命令目标。
7. `docs/specs/_check_result/{module}.md` 中必须同时满足：
   - `module` 等于目标正式模块名
   - `gate=cand_check`
   - `decision=pass`
   - `allow_next=true`
   - `spec_layer_ref=candidate`
   - `spec_file_ref` 等于当前 candidate 文件路径
   - `spec_version_ref` 等于当前 candidate 的版本引用
   - `spec_fingerprint` 等于当前 candidate 的指纹
   - `next_command=cand_plan`
   - 当前正式全局约束存在时，`system_constraints_stable_file_ref` 等于当前正式全局约束文件路径；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，`system_constraints_stable_version_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，`system_constraints_stable_fingerprint` 等于当前正式全局约束指纹；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，当前 candidate 中的 `system_constraints_stable_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
   - 当前 candidate 当前层 `shared_appendix_refs=none` 时，`shared_appendix_snapshot=none`
   - 当前 candidate 当前层 `shared_appendix_refs` 非空时，`shared_appendix_snapshot` 精确等于按 `spec_policy.md` 第 `12.1` 节规则从当前绑定 Shared Appendix 重新生成的规范化快照
8. 若模块命中 Prompt Adequacy Review，`docs/specs/_check_result/{module}.md` 中还必须同时满足：
   - `prompt_adequacy_review_required=true`
   - `prompt_adequacy_decision=pass`
   - `prompt_adequacy_summary` 满足 `spec_policy.md` 第 `12.2` 节的语义契约
9. 若模块未命中 Prompt Adequacy Review，`docs/specs/_check_result/{module}.md` 中还必须同时满足：
   - `prompt_adequacy_review_required=false`
   - `prompt_adequacy_decision=n/a`
   - `prompt_adequacy_summary` 满足 `spec_policy.md` 第 `12.2` 节的语义契约
10. 若本轮会修改 `_plans/{module}.md`、`_status.md` 或其它命中提交触发条件的过程 / 治理文件，执行前必须按 `docs/agent_guidelines/command_policy.md` 第 `10` 节读取 Git 收口规则文件，确认本轮是否要求提交以及应按哪类提交收口。
11. 若前置自检或读取过程中发现被引用的附属展开文件发生目录漂移，当前命令必须先完成迁移并重新执行前置自检，再继续规划。

## 4. Procedure

1. 读取 `docs/specs/candidate/c_{module}.md`；若该文件明确引用了 candidate 层附属展开文件，必须一并读取。
2. 若模块当前层 `Global Constraint Alignment.shared_appendix_refs` 非空，必须一并读取这些 Shared Appendix；它们属于当前 candidate 层真相读取面。
3. 若存在 `stable`，补读 `docs/specs/stable/s_{module}.md`；若 stable 主文件明确引用了 stable 层附属展开文件，也必须一并读取。
4. 若 `docs/specs/system/stable/s_system_constraints.md` 已存在，读取它；若尚不存在，则按“当前无正式全局基线”的空态继续。
5. 读取 `docs/specs/_check_result/{module}.md` 放行凭证。
6. 若在第 1-3 步或前置自检中发现附属展开文件目录漂移，必须先由当前命令完成迁移并重新执行前置自检；在迁移完成前不得继续生成或更新计划文件。
7. 按 `docs/specs/_check_result/README.md` 与 `spec_policy.md` 第 `12.6` 节逐项校验该文件的完整绑定关系，至少包括：
   - `module`
   - `gate=cand_check`
   - `decision=pass`
   - `allow_next=true`
   - `spec_layer_ref=candidate`
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `next_command=cand_plan`
   - `system_constraints_stable_file_ref`
   - `system_constraints_stable_version_ref`
   - `system_constraints_stable_fingerprint`
   - `shared_appendix_snapshot`
   - 当前 candidate 中的 `system_constraints_stable_ref`
   - 若命中 Prompt Adequacy Review，还必须校验 `prompt_adequacy_review_required`、`prompt_adequacy_decision`、`prompt_adequacy_summary`
8. 若上述任一绑定条件不满足，必须立刻停止当前命令，并把 `_status.md` 中该模块的 `Next Command` 回退为 `cand_check`；不得继续沿用旧 gate 进入计划阶段。
9. 只有当完整绑定关系与门禁放行字段都成立时，才允许继续执行 `cand_plan`。
10. 梳理当前轮实施任务项、依赖关系和验证重点。
11. 创建或更新 `docs/specs/_plans/{module}.md`，至少写入：
   - `Implementation Tasks`
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `system_constraints_stable_file_ref=docs/specs/system/stable/s_system_constraints.md|none`
   - `system_constraints_stable_version_ref=...|none`
   - `system_constraints_stable_fingerprint=...|none`
   - `shared_appendix_snapshot=...|none`
12. 更新 `_status.md` 中该模块的 `Next Command=cand_impl`。
13. 若本轮改动命中 Git 收口规则文件的提交触发条件，必须按该规则判断并完成当前任务内的 git 收口。

## 5. Stop Conditions

1. 当前轮计划已形成。
2. `docs/specs/_plans/{module}.md` 已落盘。
3. 若放行凭证失效，当前命令已停止，且 `_status.md` 已回退为 `cand_check`。
4. 不继续自动进入代码实现。

## 6. Output Contract

1. 差异或落地摘要
2. 计划文件路径
3. 当前轮计划
4. 风险与依赖
5. 若放行凭证失效，必须明确输出失效原因与 `_status.md` 回退结果
6. Git 收口结果
7. 下一步建议

## 7. Non-Goals

1. 直接改代码
2. 创建或读取独立的 `system_constraints` candidate 文件

## 8. Examples

```md
cand_plan:module_example
```
