# Unit Verify

`unit_verify:{unit}` 验证实现是否满足候选 truth 中的每个 acceptance item。

## 输入

- `docs/specs/_status.md`
- `docs/specs/units/candidate/c_unit_{unit}.md`
- 当前 unit 的候选层附录文件
- 当前 unit 引用的稳定层 truth 和 rule 文件
- 本 unit 的实现文件和测试文件
- `docs/specs/_check_result/unit/{unit}.md`（如存在，参考但不必须）

## 本步骤做什么

1. **功能验证**：验证每个 acceptance item 是否满足，给出可检查的证据
2. **范围验证**：验证 `affects` 中声明的文件/附录/规则/依赖是否正确实现
3. **退役验证**（replacement 场景）：验证旧代码路径已完全删除，无残留引用
4. **代码质量检查**：无死代码、不过度设计、变更量合理

## 验证证据要求

- 每个可执行的 acceptance item 都必须在 `acceptance_item_evidence_matrix` 中有对应条目
- 主协议、API、UI、生成产物的变更必须检查真实产出（截图、API 返回值、CLI 输出等），不能仅靠"测试通过"
- 验证不能自动删除代码或推断业务兼容性安全

## 注意

- 本步骤需要独立评审，**不能自评通过**。必须有一个上下文独立的评审者给出 `pass` 才能 `ready_to_promote`
- verify 过程中发现实现问题可以修复，修完重新验证
- 如果候选 Spec 本身有问题，退回 `unit_check` 修复 Spec

## 不允许

- 修改候选或稳定层 truth
- 修改 lifecycle 状态
- 修改 rule truth

## 如何结束

| 结果 | 含义 | 下一步 |
|------|------|--------|
| `ready_to_promote` | 验证通过，评审通过 | 写入 `_verify_result`，进入 `unit_promote` |
| 其他 | 需要修复 | 修复后重新验证 |
