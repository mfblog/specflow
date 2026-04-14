# Candidate Check Results

本目录存放候选收口通过后生成的候选链放行凭证。

规则：

1. 只有已经通过 `cand_check` 且当前仍保有有效候选链放行结果的模块，才应存在 `_check_result/{module}.md`。
2. 这里的文件不是正式 Spec，也不是行为真相源。
3. 每份 `_check_result/{module}.md` 默认承载当前 candidate 的最新有效候选链放行快照，而不是失败审查记录。
4. `Check Result Snapshot` 必须使用固定字段：
   - `module`
   - `gate`
   - `decision`
   - `allow_next`
   - `next_command`
   - `blocking_summary`
   - `coverage_summary`
   - `prompt_adequacy_review_required`
   - `prompt_adequacy_decision`
   - `prompt_adequacy_summary`
   - `spec_layer_ref`
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `system_constraints_stable_file_ref`
   - `system_constraints_stable_version_ref`
   - `system_constraints_stable_fingerprint`
   - `shared_appendix_snapshot`
5. `gate` 固定为 `cand_check`。
6. `next_command` 固定为 `cand_plan`。
7. `cand_check` 只有在结论为 `pass` 时才创建或覆盖该文件。
8. `cand_check` 未通过时，不得写入失败态文件；若旧 gate 已不再成立，必须删除该文件。
9. 只要当前 candidate 内容发生变化，当前 `_check_result/{module}.md` 即视为过期，不得继续支撑 `cand_plan`、`cand_impl` 或 `cand_verify`。
10. 若当前正式全局基线已存在，只要当前 candidate 中的 `system_constraints_stable_ref` 不等于当前正式全局基线版本，当前 `_check_result/{module}.md` 即视为过期，不得继续支撑 `cand_plan`、`cand_impl` 或 `cand_verify`。
11. 若当前正式全局基线尚不存在，只要当前 candidate 中的 `system_constraints_stable_ref` 不等于 `none`，当前 `_check_result/{module}.md` 即视为过期，不得继续支撑 `cand_plan`、`cand_impl` 或 `cand_verify`。
12. 若当前 candidate 层 `shared_appendix_refs` 绑定的共享附属展开文件版本、正文、层级或绑定关系变化，当前 `_check_result/{module}.md` 即视为过期，不得继续支撑 `cand_plan`、`cand_impl` 或 `cand_verify`。
13. 当 `spec_fork` 开启新一轮 candidate 时，必须删除上一轮 `_check_result/{module}.md`。
14. 当 `cand_promote` 完成候选提升时，必须删除对应 `_check_result/{module}.md`。
15. 消费该文件时，不得只看文件存在；还必须同时校验：
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
   - 若模块命中 Prompt Adequacy Review，则还必须满足：
     - `prompt_adequacy_review_required=true`
     - `prompt_adequacy_decision=pass`
     - `prompt_adequacy_summary` 满足 `spec_policy.md` 中约定的最小语义契约
   - 若模块未命中 Prompt Adequacy Review，则还必须满足：
     - `prompt_adequacy_review_required=false`
     - `prompt_adequacy_decision=n/a`
     - `prompt_adequacy_summary` 满足 `spec_policy.md` 中约定的最小语义契约
16. `spec_version_ref` 固定格式为 `c_{module}@<frontmatter.version>`。
17. `system_constraints_stable_version_ref` 在正式全局基线存在时固定格式为 `s_system_constraints@<frontmatter.version>`；若不存在则固定写 `none`。
18. `spec_fingerprint` 固定为当前 candidate 文件完整原文在首尾空白裁剪后的指纹。
19. `system_constraints_stable_fingerprint` 在正式全局基线存在时按完整原文、首尾空白裁剪后的口径生成；若不存在则固定写 `none`。
20. `shared_appendix_snapshot` 固定按 `spec_policy.md` 第 `12.1` 节定义的规范化口径生成；若当前 candidate 当前层 `shared_appendix_refs=none`，则固定写 `none`。
21. `cand_check` 允许自动前移 `system_constraints_stable_ref`，或在“尚无正式全局基线”场景下把它纠正为 `none`；但该自动修正只适用于机械性基线绑定对齐，不得顺手修改 candidate 的任何其它真相内容。
22. `_check_result/{module}.md` 的 `next_command` 只回答“这份 candidate 的候选链入链入口是 `cand_plan`”，不回答模块当前已经推进到候选链的哪一步。
23. 模块当前最小可行动作始终由 `docs/specs/_status.md` 的 `Next Command` 负责；`cand_impl`、`cand_verify` 继续消费 `_check_result/{module}.md` 时，只校验它是否仍是当前 candidate 的有效放行凭证，不要求它把 `next_command` 改写成当前命令。
