# Candidate Plans

本目录存放候选推进过程中的实施计划文件。

规则：

1. 每个模块默认对应一份 `_plans/{module}.md`。
2. 这里的文件不是正式 Spec，也不是行为真相源。
3. 每份 `_plans/{module}.md` 默认承载两类过程信息：
   - `Implementation Tasks`
   - 当前轮实施进度、阻塞与验证重点
4. 每份 `_plans/{module}.md` 还必须记录：
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `system_constraints_stable_file_ref`
   - `system_constraints_stable_version_ref`
   - `system_constraints_stable_fingerprint`
   - `shared_appendix_snapshot`
5. `_plans/{module}.md` 不承载 `cand_check` 的门禁结论，也不承载 `cand_verify` 的提升结论。
6. `cand_plan` 首次执行时，若该文件不存在，应创建计划文件，并写入当前 candidate 与当前正式全局基线状态的绑定信息。
7. `cand_impl` 可以持续回写当前轮实施进度、阻塞与验证重点，但不得改写绑定字段。
8. `cand_verify` 进入前也必须仍然持有当前有效的 `_plans/{module}.md`；`cand_verify` 不回写计划绑定字段，但必须把该文件视为候选链必经门禁的一部分。
9. 只要当前 candidate 原文发生任何变化，当前 `_plans/{module}.md` 即视为过期，不得继续支撑 `cand_impl` 或 `cand_verify`。
10. 若当前正式全局基线已存在，只要当前 candidate 中的 `system_constraints_stable_ref` 不等于当前正式全局基线版本，当前 `_plans/{module}.md` 即视为过期，不得继续支撑 `cand_impl` 或 `cand_verify`。
11. 若当前正式全局基线尚不存在，只要当前 candidate 中的 `system_constraints_stable_ref` 不等于 `none`，当前 `_plans/{module}.md` 即视为过期，不得继续支撑 `cand_impl` 或 `cand_verify`。
12. 若当前 candidate 层 `shared_appendix_refs` 绑定的共享附属展开文件版本、正文、层级或绑定关系变化，当前 `_plans/{module}.md` 即视为过期，不得继续支撑 `cand_impl` 或 `cand_verify`。
13. `_plans/{module}.md` 过期后，必须先回到 `cand_check`；若 `cand_check` 通过，再重新执行 `cand_plan`。即使旧计划看起来仍能描述实施顺序，也不得在缺少当前有效 `_check_result/{module}.md` 的前提下直接继续进入 `cand_impl` 或 `cand_verify`。
14. `spec_version_ref` 固定格式为 `c_{module}@<frontmatter.version>`。
15. `system_constraints_stable_version_ref` 在正式全局基线存在时固定格式为 `s_system_constraints@<frontmatter.version>`；若不存在则固定写 `none`。
16. `spec_fingerprint` 固定为当前 candidate 文件完整原文在首尾空白裁剪后的指纹。
17. `system_constraints_stable_fingerprint` 在正式全局基线存在时按完整原文、首尾空白裁剪后的口径生成；若不存在则固定写 `none`。
18. `shared_appendix_snapshot` 固定按 `spec_policy.md` 第 `12.1` 节定义的规范化口径生成；若当前 candidate 当前层 `shared_appendix_refs=none`，则固定写 `none`。
19. 当 `spec_fork` 开启新一轮 candidate 时，若存在上一轮 `_plans/{module}.md`，必须先删除，避免沿用旧轮次计划。
20. 当 `cand_promote` 完成候选提升时，必须删除对应 `_plans/{module}.md`。
21. `Candidate=no` 时，默认不得保留对应 `_plans/{module}.md`。
