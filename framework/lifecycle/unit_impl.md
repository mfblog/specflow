# Unit Implementation

`unit_impl` is the implementation phase between candidate truth validation and verification.

## 输入

- `docs/specs/units/candidate/c_unit_{unit}.md`
- 当前 unit 的候选层附录文件
- 当前 unit 引用的稳定层 truth 和 rule 文件
- `docs/specs/_check_result/unit/{unit}.md`（如存在）

## 本步骤做什么

根据候选 Spec 中的 acceptance items 实现代码。
实现过程中如果发现 Spec 缺失或错误，停止并问用户。

## 不允许

- 修改 Spec 文件（候选或稳定层）
- 修改 lifecycle 状态
- 实现超出候选 Spec 范围的行为
- 修改 rule truth 或 global rules

## 如何结束

所有 acceptance items 实现完成后，进入 `unit_verify:{unit}`。
无需特殊 close 命令。
