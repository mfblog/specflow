# Stable Verify Command

## 1. Purpose

本命令用于核对当前代码是否仍对齐指定模块的 `stable` Spec。

目标只有两个：

1. 判断当前实现是否仍满足正式版真相。
2. 在发现偏差时，明确是应先回到 `stable`，还是进入受控升级。

---

## 2. Scope

本命令默认处理：

1. `Active Layer=stable` 的模块日常改动、修复、重构后的回归核对。
2. `stable` 与当前代码的一致性核对。
3. 正式层偏差项与后续动作判断。

本命令默认不处理：

1. 开启新的 `candidate`。
2. 设计新的未来行为。
3. 直接修改代码。
4. 用一句主观判断替代结构化验证证据。

---

## 3. Preconditions

执行前必须确认：

1. 已先完成 `spec_policy.md` 第 `8` 节定义的前置自检，且不存在状态漂移或文件漂移。
2. 目标模块当前 `Active Layer=stable`。
3. `_status.md` 中该模块的 `Next Command=stable_verify`。
4. 已先完成 `spec_policy.md` 第 `8.1` 节定义的 `stable drift reconciliation`。
5. 已明确目标模块名。
6. 该模块存在有效的 `stable`。
7. 当前代码存在需要确认是否仍对齐正式版的实现上下文。
8. 若 `s_{module}.md` 正文明确引用了 stable 层附属展开文件，或 `Global Constraint Alignment.shared_appendix_refs` 显式绑定了 stable 共享附属展开文件，执行前必须一并读取这些文件。
9. 若本轮核对需要同时判断正式技术基线、共享机制或全局例外，执行前还应读取 `docs/specs/system/stable/s_system_constraints.md`；该文件是否必读，应由当前核对场景决定，而不是由统一总入口顺序决定。
10. 若本轮核对会命中实现、审查或非 `candidate` 真相文件修改等提交触发条件，执行前必须按 `docs/agent_guidelines/command_policy.md` 第 `10` 节读取 Git 收口规则文件，确认本轮是否要求提交以及应按哪类提交收口。
11. 若前置自检或读取过程中发现被引用的附属展开文件发生目录漂移，当前命令必须先完成迁移并重新执行前置自检，再继续核对。
12. 若当前 `stable` 层 `shared_appendix_refs` 绑定的 stable 共享附属展开文件已发生版本、正文或绑定关系变化，执行前必须把它视为 stable 漂移来源；不得继续宣称“当前仍对齐 stable”而跳过本命令。

若以上条件不满足，不得直接执行。

---

## 4. Procedure

执行步骤固定如下：

1. 若当前核对场景需要判断正式技术基线、共享机制或全局例外，先读取 `docs/specs/system/stable/s_system_constraints.md` 作为对应场景输入。
2. 读取 `docs/specs/stable/s_{module}.md`；若该文件明确引用了 stable 层附属展开文件，或 `Global Constraint Alignment.shared_appendix_refs` 显式绑定了 stable 共享附属展开文件，必须一并读取。
3. 若在第 2 步或前置自检中发现附属展开文件目录漂移，必须先由当前命令完成迁移并重新执行前置自检；在迁移完成前不得继续核对。
4. 核对当前代码是否满足 `stable` 中的关键协议、主流程、错误处理和验收标准。
4. 生成结构化验证证据矩阵，至少覆盖：
   - `Spec Item`
   - `Expected Behavior`
   - `Implementation Evidence`
   - `Verification Evidence`
   - `Status`
5. 确保验证证据矩阵覆盖全部关键验收点；若原文验收标准过大，先拆成可核对验收点再逐条映射。
6. 输出 `Coverage Summary`，至少包含：
   - `Total`
   - `Covered`
   - `Failed`
   - `Partial`
   - `Not Checked`
7. 对所有 `partial` 与 `not_checked` 项强制补充风险说明，并按 `spec_policy.md` 中的统一降级规则判断其是否可能为非阻塞项。
8. 按 `spec_policy.md` 中统一定义的 `P1 / P2 / P3` 整理偏差项。
9. 给出推进结论：
   - 若存在 `fail`，则结论只能是：`存在偏差，必须先回到 stable 或进入 spec_fork`
   - 存在 `partial` 或 `not_checked` 时，只有满足 `spec_policy.md` 中的非阻塞验证证据降级规则，才允许继续判断
   - 若关键偏差已清空，且验证证据完整，则结论为：`当前仍对齐 stable`
10. 若代码已偏离 `stable`，后续动作只能二选一：
   - 把代码拉回 `stable` 语义
   - 通过 `spec_fork:{module}` 开启新 candidate，把偏离转正为受控升级
11. 更新 `_status.md`：
   - 若当前仍对齐 `stable`：`Next Command=spec_fork`
   - 若存在偏差：保持 `Next Command=stable_verify`
12. 若本轮改动命中 Git 收口规则文件的提交触发条件，必须按该规则判断并完成当前任务内的 git 收口。

补充约束：

1. `stable_verify` 不生成可复用的过程文件结果。
2. 因此，只要正式层代码再次发生新的未核对改动，后续命令必须直接回退到新的 `stable_verify`，不得依据“可能已被最近一次验证覆盖”的口头判断继续推进。

---

## 5. Stop Conditions

默认停止条件如下：

1. 已明确当前代码是否仍对齐 `stable`。
2. 已明确后续动作是保持正式层、回到正式层，还是进入 `spec_fork`。
3. `_status.md` 已同步更新。

---

## 6. Output Contract

输出顺序固定如下：

1. 核对结论
2. 结构化验证证据矩阵
3. `Coverage Summary`
4. 偏差清单
5. 后续动作建议
6. Git 收口结果
7. `_status.md` 更新结果

验证证据矩阵要求：

1. 每一行都必须对应一个明确的 Spec 条目或验收点。
2. `Status` 只能使用：`pass / fail / partial / not_checked`。
3. `partial` 与 `not_checked` 必须附风险说明。
4. `P1 / P2 / P3` 的语义以 `spec_policy.md` 中的统一定义为准，不得在本命令内另起一套解释。
5. 任何被判为非阻塞的 `partial` 或 `not_checked`，都必须满足 `spec_policy.md` 中的统一降级规则，并明确写出理由与残余风险。
6. 必须覆盖全部关键验收点；若未覆盖，必须在 `Coverage Summary` 中反映为 `Not Checked`。
7. 不得只给“目前看起来没问题”这类无证据总结。

---

## 7. Non-Goals

本命令不负责：

1. 创建 `candidate`
2. 用正式层核对替代升级设计
3. 直接宣布未来行为已经成立

---

## 8. Examples

### 8.1 核对 stable 是否仍然对齐

```md
stable_verify:module_example
```
