# Unit Stable Verify

`unit_stable_verify:{unit}` 检查当前实现是否仍然符合稳定层 truth。

## 输入

- `docs/specs/_status.md`
- `docs/specs/units/stable/s_unit_{unit}.md`
- 稳定层附录和本 unit 引用的 rule 文件
- `docs/specs/repository_mapping.md` 中本 unit 的条目
- 当前的实现文件和测试文件
- 已有的 `_stable_verify_result/unit/{unit}.md`（如需要更新）

## 本步骤做什么

检查当前实现与稳定层 truth 的一致性。
输出应为 `aligned`（一致）、`controlled_repair_required`（需修复）、或 `controlled_change_required`（需变更）。

## 注意

- 本步骤需要独立评审，不能自评通过
- 稳定验证不创建候选 truth 本身。如需变更，结果是触发后续的 `unit_fork`
- `aligned` 要求的每个 acceptance item 必须有 `pass` 证据

## 不允许

- 修改稳定层或候选层 truth
- 修改实现文件
- 修改 lifecycle 状态
- 修改 rule truth

## 如何结束

| 结果 | 含义 | 下一步 |
|------|------|--------|
| `aligned` | 实现与稳定 truth 一致 | `unit_fork` |
| `controlled_repair_required` | 需要修复 | `unit_fork` with repair intent |
| `controlled_change_required` | 需要变更 | `unit_fork` with change intent |
| `small_repair_required` | 需要小范围修复，不改变行为 truth | `unit_stable_verify`（重新验证） |
| `truth_rejudge_required` | 稳定层 truth 需要重新判断 | `unit_stable_verify`（重新验证） |
| `evidence_incomplete` | 证据不足 | 补充证据后重新验证 |

通过 `command close` 关闭。
