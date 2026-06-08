# Lifecycle Overview

## 生命周期序列

当需要正式 unit 治理时，标准生命周期为：

```
unit_new / unit_fork → unit_check → unit_impl → unit_verify → unit_promote
```

- `unit_check` 是可选的 pre-verify 质量门禁，验证候选 truth 是否足够清晰
- `unit_impl` 是单元实现阶段，由 `unit_check pass` 自动触发
- `unit_verify` 验证实现是否满足候选 truth
- `unit_promote` 将已验证的候选 truth 晋升为稳定 truth

## 入口方式

`entry_routing.md` 决定一个自然语言请求走哪个生命周期路径。
支持 exact command 匹配（`command:{unit}`）和自然语言两种方式。

## 入口命令

| 命令 | 用途 |
|------|------|
| `unit_init:{unit}` | 已有能力→首个稳定 truth |
| `unit_new:{unit}` | 全新→首个候选 truth |
| `unit_fork:{unit}` | 稳定 truth→候选变更轮次 |
| `unit_check:{unit}` | 候选 truth 质量检查（可选） |
| `unit_verify:{unit}` | 验证实现 vs 候选 truth |
| `unit_promote:{unit}` | 候选 truth→稳定 truth |
| `unit_stable_verify:{unit}` | 检查实现 vs 稳定 truth |

`unit_impl` 是一个自动推进状态，由 `unit_check pass` 设置，不是用户输入命令。`entry_routing.md` 负责在状态为 `Next Command=unit_impl` 时路由到 `framework/lifecycle/unit_impl.md`。

## 命令执行规则

- `command close` 是唯一能推进 lifecycle 状态的操作
- 推进式 evidence（`unit_check pass`、`unit_verify ready_to_promote`、`unit_stable_verify advancing`）需要独立评审 receipt
- `unit_promote` 消费已验证的证据，不需要新的独立评审
- 非推进式结果（blocked、fix_required、evidence_incomplete）不阻塞后续正确证据的推进

## 依赖管理

- 候选 unit 可依赖当前稳定层 unit 版本或当前候选 truth（如果 Context Card 允许）
- 稳定层晋升不能静默改变其他 unit 消费的稳定版本
- 稳定版本变更时，需运行 `governance/impact_sync.md` 检查其影响

## Rule 消费

- Global rules 自动应用于所有当前层的 unit
- Bound rules 仅当 unit 的 `rule_refs` 中显式列出时才应用
- Rule 变更通过 `framework/governance/rule_system.md` 管理

## 生命周期状态

`docs/specs/_status.md` 记录每个 unit 的当前状态（layer、Next Command）。
只有 `command close` 可以修改这个文件。

## Context Card 格式（框架设计者参考）

每个生命周期 Context Card 包含以下部分：

1. **输入** — 本步骤需要读取的文件
2. **本步骤做什么** — 当前命令的目标和执行内容
3. **不允许** — 硬边界
4. **如何结束** — 成功/失败/阻塞的结果及下一步

Context Card 中的 `framework/...` 路径是相对于框架根目录的：
- 已安装项目：`framework/...` → `specflow/framework/...`
- 源代码仓库：`framework/...` → `framework/...`

## 生命周期权限规则

`command close` 是推进 lifecycle 状态的唯一途径。

推进式 evidence 的有效条件是：
1. 当前 Context Card 允许该 evidence 写入
2. 当前 process 文件通过对应的 `snapshot validate-process` 检查
3. process 文件包含有效的独立评审 receipt（当 Context Card 要求时）
4. `command close` 接受该结果和 evidence

有效输入 evidence 是消耗性的——只有当前通过确定性验证的文件才能被消耗。

非推进式结果（blocked、fix_required、checkpoint）不会永久 disqualify 后续工作。修复后，当前 evidence 只要通过验证并携带独立评审 receipt，仍然可以推进。
