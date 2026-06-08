# Unit Entry Commands

本文件覆盖 `unit_init:{unit}`、`unit_new:{unit}`、`unit_fork:{unit}` 三个入口命令。

| 命令 | 用途 |
|------|------|
| `unit_init` | 将已有能力直接记录为首个稳定 truth（无需候选层） |
| `unit_new` | 为一个全新的 unit 创建首个候选 truth |
| `unit_fork` | 从现有稳定 truth 分支出一个候选轮次进行变更或修复 |

## 输入

- `docs/specs/_status.md`（如果目标 unit 可能已注册）
- `docs/specs/repository_mapping.md`（如果需确认路径所有权或注册）
- `docs/specs/units/stable/s_unit_{unit}.md` + 稳定层附录（仅 `unit_fork`）

## 各命令的要求

### unit_init
已有被接受的能力必须足够明确，能在不选择新行为/acceptance/ownership 的前提下写出稳定 truth。

### unit_new
候选 truth 必须足够明确，能写出第一个候选 Spec 及其 source 字段。

### unit_fork
- 当前稳定 truth 是候选轮次的基线
- 需确定 `candidate_intent`（`change` 或 `repair`）
- 如果存在有效的稳定 verify 结果：
  - `controlled_repair_required` → 写 `repair`
  - `controlled_change_required` → 写 `change`
  - `aligned` → 不强制特定 intent
- 每个稳定层附录必须有对应的同名候选层附录

## 不允许

- 修改实现文件
- 手工修改 lifecycle 状态
- 修改 rule truth 或 global rules
- 修改其他 unit 的 truth
- `unit_new` / `unit_fork` 期间修改稳定层 truth
- `unit_init` 期间修改候选层 truth
- 引入尚未在 Required Context 中决策的行为/acceptance/ownership/rule

## 注意

`unit_check` 是这三个命令的可选后续质量门禁，不是必选步骤。

## 如何结束

| 命令 | 成功结果 | 下一步 |
|------|---------|--------|
| `unit_init` | `stable_created` | `unit_fork` |
| `unit_new` | `candidate_created` | `unit_verify` |
| `unit_fork` | `candidate_created` | `unit_verify` |

通过 `command close` 关闭。关闭前确保所有写入完成，无未解决的 rule-governance 或 ownership 问题。
