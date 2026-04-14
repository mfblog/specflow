# Candidate Promote Command

## 1. Purpose

本命令用于把指定模块的 `candidate` 正式提升为新的 `stable`。

## 2. Scope

本命令默认处理：

1. 候选版本提升为正式版本。
2. 状态文件更新。
3. 当前轮候选与过程文件清理。
4. 在需要时联动更新 `s_system_constraints.md`。

## 3. Preconditions

执行前必须确认：

1. 已先完成前置自检，且不存在状态漂移或文件漂移。
2. `_status.md` 中该模块的 `Next Command=cand_promote`。
3. `docs/specs/_verify_result/{module}.md` 中存在最新 `Verify Result Snapshot`，且仍覆盖当前 candidate、当前实现与当前正式全局基线状态；这里的“仍覆盖”必须满足 `docs/specs/_verify_result/README.md` 与 `spec_policy.md` 第 `12.7` 节定义的完整绑定条件。
4. 当前代码实现已完成必要对齐，且验证结果不再存在阻塞项。
5. 当前 candidate 中的 `system_constraints_stable_ref` 与当前正式全局基线状态一致。
6. 若 `c_{module}.md` 正文明确引用了 candidate 层附属展开文件，或 `Global Constraint Alignment.shared_appendix_refs` 显式绑定了共享附属展开文件，执行前必须一并读取这些文件，并明确它们在提升后是并入 stable 主文件、迁移到 stable appendix / shared stable，还是删除。
7. `docs/specs/_verify_result/{module}.md` 中必须同时满足：
   - `gate=cand_verify`
   - `decision=pass`
   - `allow_next=true`
   - `spec_layer_ref=candidate`
   - `spec_file_ref` 等于当前 candidate 文件路径
   - `spec_version_ref` 等于当前 candidate 的版本引用
   - `spec_fingerprint` 等于当前 candidate 的指纹
   - `next_command=cand_promote`
   - `verification_scope_ref` 仍覆盖当前实现上下文
   - 当前正式全局约束存在时，`system_constraints_stable_file_ref` 等于当前正式全局约束文件路径；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，`system_constraints_stable_version_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，`system_constraints_stable_fingerprint` 等于当前正式全局约束指纹；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，当前 candidate 中的 `system_constraints_stable_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
   - 当前 candidate 当前层 `shared_appendix_refs=none` 时，`shared_appendix_snapshot=none`
   - 当前 candidate 当前层 `shared_appendix_refs` 非空时，`shared_appendix_snapshot` 精确等于按 `spec_policy.md` 第 `12.1` 节规则从当前绑定 Shared Appendix 重新生成的规范化快照
8. 执行前必须按 `docs/agent_guidelines/command_policy.md` 第 `10` 节读取 Git 收口规则文件，确认本轮升级改动是否要求提交以及应按哪类提交收口。

## 4. Procedure

1. 读取并复核 `docs/specs/_verify_result/{module}.md` 中最新 `Verify Result Snapshot`。
2. 读取 `docs/specs/candidate/c_{module}.md`；若该文件明确引用了 candidate 层附属展开文件，必须一并读取，并确认这些文件的提升去向。
3. 按 `docs/specs/_verify_result/README.md` 与 `spec_policy.md` 第 `12.7` 节逐项校验该文件的完整绑定关系，至少包括：
   - `gate=cand_verify`
   - `decision=pass`
   - `allow_next=true`
   - `spec_layer_ref=candidate`
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `next_command=cand_promote`
   - `verification_scope_ref`
   - `system_constraints_stable_file_ref`
   - `system_constraints_stable_version_ref`
   - `system_constraints_stable_fingerprint`
   - `shared_appendix_snapshot`
   - 当前 candidate 中的 `system_constraints_stable_ref`
4. 若发现 `_verify_result/{module}.md` 已失效，必须先识别失效原因，再立刻停止当前命令并回退 `_status.md`：
   - 若是验证后代码再次发生新的未核对改动，回退为 `cand_verify`
   - 若是验证已明确实现与 candidate 有偏差，回退为 `cand_impl`
   - 若是 candidate 或正式全局基线已变化到需要重新收口，回退为 `cand_check`
5. 只有当完整绑定关系、验证覆盖范围与门禁放行字段都成立时，才允许继续提升。
6. 确认当前 candidate 的 `frontmatter.version` 就是本轮准备提升成的新 `stable` 版本。
7. 若当前模块 candidate 中 `promotion_to_system_stable=with_module`，则在同一次 `cand_promote` 中先把 `proposed_system_constraints_updates` 吸收到 `docs/specs/system/stable/s_system_constraints.md`；若该文件尚不存在，则本轮创建它，若已存在，则按语义化版本规则递增其 `frontmatter.version`。
7A. 若当前 candidate 中 `shared_appendix_refs` 非空，必须对每个绑定项逐一做出强制决议；不得用“后续再看”带过：
   - 迁移到 `docs/specs/shared/stable/s_shared_xxx.md`
   - 把稳定结论吸收到 `s_system_constraints.md`
   - 把稳定结论吸收到模块 `stable` 主文后删除共享附属展开文件
   - 若当前轮无法完成上述任一收口，则本轮 `cand_promote` 必须停止，不得继续提升
7B. 共享附属展开文件不得独立 promote；它们只允许在模块 `cand_promote` 中被联动保留、迁移到 `docs/specs/shared/stable/`，或把稳定结论吸收到 `s_system_constraints.md`。
8. 上一步属于 `cand_promote` 内部的原子收口步骤；在本命令完成前，不把它视为候选链中途失效，也不因此触发回退。
8A. 若本命令已经基于最新 Shared Appendix 真相，为当前目标模块写出新的 stable 真相与 stable 层 `shared_appendix_refs`，则当前目标模块视为已在本命令内完成 Shared Appendix 状态收口；后续为本轮执行 Shared Appendix 状态收口时，不得再把当前目标模块按“本轮刚改过 shared”机械回退到 `stable_verify`。
9. 但若命令在第 7 步之后中断、崩溃或被人工打断，必须把当前仓库视为“提升未完成的恢复态”，不得把它当作 promote 已完成。
10. 生成或更新 `docs/specs/stable/s_{module}.md`：
   - `frontmatter.version` 必须等于当前 candidate 的 `frontmatter.version`
   - 若保留 `Global Constraint Alignment` 或等价章节：
     - 本轮若已联动创建或更新正式全局基线，其中 `system_constraints_stable_ref` 必须写成当前最新正式全局基线版本
     - 本轮若仍未形成正式全局基线，该字段必须显式写 `none`
     - 若本轮提升后模块 `stable` 仍依赖共享附属展开文件，`shared_appendix_refs` 必须同步写成 stable 层引用；不得省略，也不得继续保留 `c_shared_xxx@...`
11. 若当前轮 candidate 存在附属展开文件，必须在本轮提升内同步完成以下其一：
   - 把仍需保留的内容迁移到 `docs/specs/stable/appendix/` 或等价专用子目录，并更新 stable 主文件引用
   - 把内容吸收进新的 `docs/specs/stable/s_{module}.md`
   - 删除本轮已不再需要的 candidate 附属展开文件
12. 在 `_status.md` 尚未完成 `Candidate=no` 之前，不得删除 `docs/specs/candidate/c_{module}.md`；这样即使命令在最终收口前中断，恢复态仍有可执行的 `cand_check` 入口。
13. 更新 `_status.md`：
   - `Stable=yes`
   - `Candidate=no`
   - `Active Layer=stable`
   - `Next Command=spec_fork`
14. 只有在第 13 步已经完成后，才允许执行物理删除：
   - `docs/specs/candidate/c_{module}.md`
   - 该模块本轮 candidate 附属展开文件
   - `docs/specs/_check_result/{module}.md`
   - `docs/specs/_verify_result/{module}.md`
   - `docs/specs/_plans/{module}.md`
15. 若在第 7 步之后、第 14 步完成之前发现仓库处于“提升未完成的恢复态”，必须立刻停止继续宣称提升成功，并执行以下恢复规则：
   - 把 `_status.md` 恢复到 candidate 语义：
     - `Candidate=yes`
     - `Active Layer=candidate`
     - `Next Command=cand_check`
     - `Stable` 按当前仓库中是否已经存在可读取的 `docs/specs/stable/s_{module}.md` 取值：若已存在则写 `yes`，否则写 `no`
   - 保留或重建 `docs/specs/candidate/c_{module}.md`，确保后续 `cand_check` 有真实可读的 candidate 输入
   - 把现有 `_check_result/{module}.md`、`_plans/{module}.md`、`_verify_result/{module}.md` 一律按失效结果处理；若仍残留，应删除或在后续 `cand_check` 前明确清理
   - 后续固定从 `cand_check` 重走，不得直接续跑 `cand_promote`
16. 若本轮改动命中 Git 收口规则文件的提交触发条件，必须按该规则判断并完成当前任务内的 git 收口。
17. 若本轮同时改动了 `docs/specs/shared/**`，或改动了任一模块当前层 `shared_appendix_refs`，则在宣称本轮状态已完全收口前，还必须确保其它未在本命令内直接收口、但受影响的模块完成 Shared Appendix 状态收口；该收口由 `shared_flow_reconcile` 负责。

## 5. Stop Conditions

1. 提升完成或已明确阻塞。
2. 状态表已同步更新。
3. 该轮候选清理已完成。
4. 若验证结果失效，当前命令已停止，且 `_status.md` 已按失效原因回退到 `cand_verify`、`cand_impl` 或 `cand_check`。
5. 若命令进入“提升未完成的恢复态”，当前命令已停止，且 `_status.md` 已恢复到 `Candidate=yes`、`Active Layer=candidate`、`Next Command=cand_check` 的 candidate 语义，同时 candidate 仍保留或已被重建到可重审状态。

## 6. Output Contract

1. 提升结论
2. 正式版本号确认结果
3. 文件与状态更新结果
4. `system_constraints` 联动提升结果
5. 清理结果
6. 若验证结果失效，必须明确输出失效原因与 `_status.md` 回退结果
7. 若进入“提升未完成的恢复态”，必须明确输出已完成到哪一步、为什么统一恢复到 `Candidate=yes`、`Active Layer=candidate`、`Next Command=cand_check`
8. Git 收口结果
9. 后续状态说明

## 7. Non-Goals

1. 重新设计 `candidate`
2. 用提升动作替代实现验证
3. 自动开启下一轮候选
4. 维护独立的 `system_constraints` candidate 文件

## 8. Examples

```md
cand_promote:module_example
```
