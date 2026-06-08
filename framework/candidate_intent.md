# Candidate Intent

`candidate_intent` 说明 unit 候选层存在的原因。仅 unit 的候选层使用。

## 可取值

| Intent | 用途 |
|--------|------|
| `change` | 候选层有意改变稳定层的行为、依赖、rule binding、acceptance 或实现预期 |
| `repair` | 候选层保持稳定层的预期行为，修复缺失/过时/格式错误/不充分的 truth |

`unit_fork` 必须写入 `candidate_intent`。

## Change 候选

### 字段要求

```yaml
candidate_intent: change
source_basis: new_design | existing_implementation | mixed | replacement
evidence_appendix_ref: none | <candidate appendix path>
```

- `repair_basis` 不允许
- 如果行为依赖当前实现/测试/运行时行为，必须使用 `existing_implementation` 或 `mixed` 并提供 evidence appendix
- 如果替换现有行为但不将其作为 selected truth，使用 `replacement` + `evidence_appendix_ref=none`
- `replacement` 场景下，至少需一个 `verification_type: inspectable` 的 acceptance item 且 `evidence_requirements` 包含 `old_code_deleted` 和 `no_remaining_refs`

### 命令行为

- **unit_fork**: 从当前稳定层 main Spec 派生，写入 `candidate_intent=change`，记录与稳定层的行为差异
- **unit_check**: 检查行为差异是否明确、边界清晰、acceptance items 可直接验证、source 字段一致
- **unit_verify**: 验证实现满足候选 truth
- **unit_promote**: 晋升后 `candidate_intent` 元数据不写入稳定层

## Repair 候选

### 字段要求

```yaml
candidate_intent: repair
repair_basis: s_unit_{unit}@<version>
source_basis: new_design
evidence_appendix_ref: none
```

- `repair_basis` 必须命名要恢复的稳定层版本
- 必须包含 `Repair Scope` 章节，说明：正在恢复的 acceptance item ids、观察到的偏离、预期变更的实现面、需证明的验证证据
- `Repair Scope` 不得 redefine 行为 truth
- 晋升时 `Repair Scope` 和 `candidate_intent` 不写入稳定层

### 命令行为

- **unit_fork**: 从稳定层 main Spec 派生，版本号使用下一个 PATCH
- **unit_check**: 必须验证 repair 候选未改变稳定行为 truth。违规（如修改协议/字段/ownership/状态机语义）必须要求 `fix_required` 并建议转为 `change`
- **unit_verify**: 必须证明实现满足 repair basis 和 acceptance items。不得将新行为/宽松 pass 条件视为修复成功
- **unit_promote**: 稳定版本为 repair_basis 的 PATCH 版本；候选专用字段不写入稳定层

## 不允许

- Chat-only 的行为决策成为 truth
- 绕过 source_basis 和 evidence appendix 规则
- 绕过标准 unit 候选命令链
