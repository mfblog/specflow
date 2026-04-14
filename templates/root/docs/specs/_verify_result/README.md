# Candidate Verify Results

本目录存放候选实现验证结果文件。

规则：

1. 每个处于 `candidate` 升级链的模块默认对应一份 `_verify_result/{module}.md`。
2. 这里的文件不是正式 Spec，也不是行为真相源。
3. 每份 `_verify_result/{module}.md` 默认承载当前 candidate 最近一次 `cand_verify` 的结果。
4. `Verify Result Snapshot` 必须使用固定字段：
   - `gate`
   - `decision`
   - `allow_next`
   - `next_command`
   - `blocking_summary`
   - `coverage_summary`
   - `spec_layer_ref`
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `verification_scope_ref`
   - `system_constraints_stable_file_ref`
   - `system_constraints_stable_version_ref`
   - `system_constraints_stable_fingerprint`
   - `shared_appendix_snapshot`
5. `gate` 固定为 `cand_verify`。
6. `next_command` 只能是 `cand_promote`、`cand_verify`、`cand_impl` 或 `cand_check`。
7. `cand_verify` 首次执行时，若文件不存在，应创建该文件。
8. 后续 `cand_verify` 应覆盖更新该文件，而不是持续追加历史噪音。
9. 只要当前 candidate 内容变化，当前 `_verify_result/{module}.md` 即视为过期，不得继续支撑 `cand_promote`。
10. 只要当前实现出现新的未核对改动，当前 `_verify_result/{module}.md` 即视为过期，不得继续支撑 `cand_promote`。
11. 若当前正式全局基线已存在，只要当前 candidate 中的 `system_constraints_stable_ref` 不等于当前正式全局基线版本，当前 `_verify_result/{module}.md` 即视为过期，不得继续支撑 `cand_promote`。
12. 若当前正式全局基线尚不存在，只要当前 candidate 中的 `system_constraints_stable_ref` 不等于 `none`，当前 `_verify_result/{module}.md` 即视为过期，不得继续支撑 `cand_promote`。
13. 若当前 candidate 层 `shared_appendix_refs` 绑定的共享附属展开文件版本、正文、层级或绑定关系变化，当前 `_verify_result/{module}.md` 即视为过期，不得继续支撑 `cand_promote`。
14. 当 `spec_fork` 开启新一轮 candidate 时，必须删除上一轮 `_verify_result/{module}.md`。
15. 当 `cand_promote` 完成候选提升时，必须删除对应 `_verify_result/{module}.md`。
16. 消费该文件时，不得只看文件存在；还必须同时校验：
   - `gate=cand_verify`
   - `spec_layer_ref=candidate`
   - `spec_file_ref` 等于当前 candidate 文件路径
   - `spec_version_ref` 等于当前 candidate 的版本引用
   - `spec_fingerprint` 等于当前 candidate 的指纹
   - `next_command` 等于当前命令期望值
   - `verification_scope_ref` 仍覆盖当前实现上下文
   - 当前正式全局约束存在时，`system_constraints_stable_file_ref` 等于当前正式全局约束文件路径；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，`system_constraints_stable_version_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，`system_constraints_stable_fingerprint` 等于当前正式全局约束指纹；若不存在，则该字段等于 `none`
   - 当前正式全局约束存在时，当前 candidate 中的 `system_constraints_stable_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
   - 当前 candidate 当前层 `shared_appendix_refs=none` 时，`shared_appendix_snapshot=none`
   - 当前 candidate 当前层 `shared_appendix_refs` 非空时，`shared_appendix_snapshot` 精确等于按 `spec_policy.md` 第 `12.1` 节规则从当前绑定 Shared Appendix 重新生成的规范化快照
   - 若下游准备执行 `cand_promote`，则还必须满足 `decision=pass`、`allow_next=true`、`next_command=cand_promote`
17. `spec_version_ref` 固定格式为 `c_{module}@<frontmatter.version>`。
18. `system_constraints_stable_version_ref` 在正式全局基线存在时固定格式为 `s_system_constraints@<frontmatter.version>`；若不存在则固定写 `none`。
19. `spec_fingerprint` 固定为当前 candidate 文件完整原文在首尾空白裁剪后的指纹。
20. `system_constraints_stable_fingerprint` 在正式全局基线存在时按完整原文、首尾空白裁剪后的口径生成；若不存在则固定写 `none`。
21. `shared_appendix_snapshot` 固定按 `spec_policy.md` 第 `12.1` 节定义的规范化口径生成；若当前 candidate 当前层 `shared_appendix_refs=none`，则固定写 `none`。
22. `verification_scope_ref` 的最小语义是“本次 `cand_verify` 覆盖的是当前 candidate 与当次实现状态”；若验证后代码再次变动，则必须视为不再覆盖当前实现。
23. `cand_verify` 只负责验证模块实现是否对齐当前 candidate 体系，不负责推进 `system_constraints` 的独立状态机。
