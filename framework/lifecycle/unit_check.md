# Unit Check

`unit_check:{unit}` 是验证前的质量门禁，检查候选 truth 是否足够清晰、完整。它本身不推进 lifecycle 状态——但 `pass` 结果的 `command close` 会将 `Next Command` 设置为 `unit_impl`。这是 close 操作的副作用，不是 `unit_check` 作为检查步骤的推进行为。

## 输入

- `docs/specs/_status.md`
- `docs/specs/units/candidate/c_unit_{unit}.md`
- 当前 unit 的候选层附录文件
- 当前 unit 引用的稳定层 truth 和 rule 文件

## 本步骤做什么

检查以下 7 个问题。全部通过才算 `pass`：

1. unit 的目标和责任范围是否清晰？
2. 依赖、rule binding、ownership 边界是否明确？
3. 主流程、数据、协议、状态、错误路径是否完整到可以验证？
4. 验证工作能否在不猜测行为/边界/acceptance 的前提下进行？
5. 所有 acceptance items 的格式是否正确（`verification_type`、`evidence_requirements`、`affects`）？
6. 如果是 `candidate_intent: change` + `source_basis: replacement`，是否有至少一个 `verification_type: inspectable` 的 item 且 `evidence_requirements` 包含 `old_code_deleted` 和 `no_remaining_refs`？
7. 所有 `affects` 范围是否正确（不能为空且无理由）？

## 不允许

- 修改实现文件
- 修改稳定层 truth
- 修改 lifecycle 状态
- 修改 rule truth

## 如何结束

| 结果 | 含义 | 下一步 |
|------|------|--------|
| `pass` | Spec 满足条件 | 写入 `_check_result`，需要独立评审通过后进入 `unit_impl` |
| `fix_required` | Spec 需要修复 | 修复候选 Spec 后重新 check |
| `blocked` | 缺少关键输入 | 问用户 |
