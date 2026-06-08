# Unit Promote

`unit_promote:{unit}` 将已验证的候选 truth 晋升为稳定 truth。

## 输入

- `docs/specs/_status.md`
- `docs/specs/_verify_result/unit/{unit}.md`
- `docs/specs/units/candidate/c_unit_{unit}.md`
- `docs/specs/units/stable/s_unit_{unit}.md`
- 当前 unit 的候选层附录文件

## 本步骤做什么

1. 将候选 truth（main Spec + appendices）写为稳定层 truth
2. 更新 lifecycle 状态和 refs
3. 清理候选层证据文件

这是一个机械性操作，不涉及新的行为判断。
`unit_promote` 不需要新的独立评审——它消费 `unit_verify` 已验证的证据。

## 不允许

- 引入已验证范围之外的行为、acceptance、ownership 或 rule 含义
- 修改实现文件
- 手工修改 lifecycle 状态
- 在 `command close --apply` 执行完之前删除候选层证据

## 如何结束

`promoted` → 执行 `command close --command unit_promote --outcome promoted --apply`。
成功后 `Active Layer=stable`、`Next Command=unit_fork`、候选层证据被清理。
