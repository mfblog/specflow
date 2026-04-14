# Candidate Check Command

## 1. Purpose

本命令用于检查指定模块的 `candidate` Spec 是否已经收口到既能稳定推进后续计划、又能完整承认关键行为真相的程度。

它本质上是一个审查动作，不是失败结果落库动作。

这里的“收口”默认包括四件事：

1. `可推进性` 已成立：
   - 模块行为已写清到足以稳定进入 `cand_plan`
   - 主流程、关键协议、关键边界、错误语义与验收口径不会让后续计划和实现分叉
2. `内容完整性` 已成立：
   - candidate 已覆盖会影响实现结果的关键行为真相
   - 不把关键依据留在文档外部、默认上下文、README 愿景、历史口头共识或作者隐含理解中
3. 当前 candidate 是否仍建立在**当前正式全局基线状态**上。
4. 若该模块命中 Prompt 相关触发条件，其 Prompt 设计是否已写清到足以稳定约束模型行为，而不是依赖作者隐含上下文。

## 2. Scope

默认只看会影响候选版本落地的内容：

1. `可推进性` 是否成立：
   - candidate 的目标与边界是否清楚
   - 数据结构与协议是否清楚
   - 状态机与主流程是否清楚
   - 输入输出语义是否足以约束实现结果
   - 分支、边界与错误处理是否足以避免实现分叉
   - 验收标准是否足以支撑后续计划与实现对齐
2. `内容完整性` 是否成立：
   - 关键行为真相是否已经在 candidate 中正式承认
   - 是否仍有会影响实现结果的关键依据留在文档外部
   - 当前留白属于 `关键层 / 重要层 / 展开层` 的哪一层
3. `Global Constraint Alignment` 是否写清：
   - `system_constraints_stable_ref`
   - `shared_appendix_refs`
   - `shared_mechanism_reuse_summary`
   - `global_constraint_exceptions`
   - `proposed_system_constraints_updates`
   - `promotion_to_system_stable`
4. 若当前模块当前层显式绑定了共享附属展开文件，绑定关系与正文依赖是否一致。
5. 若当前正式全局基线已存在且 `system_constraints_stable_ref` 落后于它，candidate 是否仍兼容新基线。
6. 若当前正式全局基线尚不存在，candidate 是否已把 `system_constraints_stable_ref` 显式写成 `none`。
7. 若模块命中 Prompt 相关触发条件，Prompt Adequacy 是否足以稳定约束实现，固定按以下审查对象执行：
   - `基础充分性审查`
   - `结构化输出审查`
   - `排序审查`
8. 若当前模块命中 shared 候选信号，是否需要提示用户进入 `shared_extract_review`，或在当前范围内直接报出 shared 双真相源冲突。

补充约束：

1. `cand_check` 不是“最小可推进审查”。
2. `cand_check pass` 的含义固定为：
   - 当前 candidate 可以进入 `cand_plan`
   - 当前 candidate 已具备作为本轮实现真相输入的关键约束完整性
3. 不允许因为“本轮改动点已经写清”或“章节已经齐全”，就忽略当前 candidate 中仍缺失的关键行为真相。

## 3. Preconditions

执行前必须确认：

1. 已先完成前置自检，且不存在状态漂移或文件漂移。
2. `_status.md` 中该模块的 `Next Command=cand_check`。
3. 该模块存在 `candidate`。
4. 若 `c_{module}.md` 正文明确引用了 candidate 层附属展开文件，或 `Global Constraint Alignment.shared_appendix_refs` 显式绑定了共享附属展开文件，执行前必须一并读取这些文件。
5. 若模块涉及 Prompt 设计、Prompt 组装或上下文注入链路，执行前必须同时读取 `docs/agent_guidelines/prompt_guidelines.md`。
6. 若本轮会修改 `_check_result/{module}.md`、`_status.md`、candidate 或其它命中审查 / 治理文件提交触发条件的对象，执行前必须按 `docs/agent_guidelines/command_policy.md` 第 `10` 节读取 Git 收口规则文件，确认本轮是否要求提交以及应按哪类提交收口。
7. 若前置自检或读取过程中发现被引用的附属展开文件发生目录漂移，当前命令必须先完成迁移并重新执行前置自检，再继续审查。

## 4. Procedure

1. 读取 `docs/specs/candidate/c_{module}.md`；若该文件明确引用了 candidate 层附属展开文件，或 `Global Constraint Alignment.shared_appendix_refs` 显式绑定了共享附属展开文件，必须一并读取；若存在 `stable`，同时补读 `docs/specs/stable/s_{module}.md`，且若 stable 主文件明确引用了 stable 层附属展开文件，也必须一并读取。
2. 若 `docs/specs/system/stable/s_system_constraints.md` 已存在，读取它；若尚不存在，则按“当前无正式全局基线”的空态继续。
3. 若在第 1 步或前置自检中发现附属展开文件目录漂移，必须先由当前命令完成迁移并重新执行前置自检；在迁移完成前不得继续给出 `cand_check` 结论。
4. 先判断 `可推进性` 是否成立，至少检查：
   - 目标与边界是否足以支撑 `cand_plan`
   - 主流程、关键协议、关键边界、错误语义与验收口径是否足以避免后续计划和实现分叉
5. 再判断 `内容完整性` 是否成立，固定按以下三个审查对象执行：
   - `Behavior Basis Completeness`
   - `Decision Surface Completeness`
   - `Acceptance Basis Completeness`
5. `内容完整性` 判断时，所有缺口都必须先归入以下三层之一：
   - `关键层`
   - `重要层`
   - `展开层`
6. `内容完整性` 三层固定语义如下：
   - `关键层`
     - 缺失后会改变实现结果
     - 缺失后不同实现者可能做出不同外部行为
     - 缺失后会影响主流程、关键分支、关键归属、关键输入来源、关键错误语义或关键验收判断
   - `重要层`
     - 不直接改变实现结果
     - 但会显著影响复审稳定性、后续维护、边界理解或复验效率
     - 若继续积累，容易在后续轮次演化成实现分叉
   - `展开层`
     - 只影响表达友好度、例子充分性、阅读成本或章节观感
     - 不影响实现结果，也不影响复审结论
7. `内容完整性` 审查对象固定含义如下：
   - `Behavior Basis Completeness`
     - candidate 是否已正式承认模块关键行为的依据，而不是只描述结果
   - `Decision Surface Completeness`
     - 执行者在关键分支、关键来源、关键归属、关键边界上是否仍需自行发明
   - `Acceptance Basis Completeness`
     - 测试或验收所依赖的关键判断依据，是否都能在 candidate 中找到正式落点
8. 判断当前模块是否命中 Prompt Adequacy Review 触发条件。默认命中条件至少包括：
   - 定义了 Prompt 结构、Prompt 组装顺序或 Prompt Block
   - 定义了 system prompt、role prompt、output prompt、system base 或等价层
   - 行为正确性高度依赖模型理解角色、状态、上下文或术语
9. 若命中 Prompt Adequacy Review：
   - 按 `docs/agent_guidelines/prompt_guidelines.md` 逐项检查三个固定审查对象：
     - `基础充分性审查`
     - `结构化输出审查`
     - `排序审查`
   - `基础充分性审查` 固定覆盖：
     - `Role Completeness`
     - `Context Sufficiency`
     - `Concept Clarity`
     - `Execution Context Completeness`
     - `Logical Closure`
   - `结构化输出审查` 仅在 Prompt 要求结构化输出时启用，固定覆盖：
     - `Output Protocol Clarity`
     - `TypeScript Schema Requirement`
     - `Few-shot Example Requirement`
   - 其中 `Few-shot Example Requirement` 只在输出结构、动作语义或边界判断较复杂时才是必审项
   - `排序审查` 固定对应 `KV Cache Friendly Ordering`
   - 若 `基础充分性审查` 或适用的 `结构化输出审查` 存在关键缺口，结论只能是 `blocked` 或 `fix_required`
   - 若仅 `排序审查` 不优，但不影响理解链路，可记改进项，不强制阻塞
10. 处理 `system_constraints_stable_ref` 时，固定按以下分支执行：
   - 若当前正式全局基线已存在，且 candidate 中的 `system_constraints_stable_ref` 不等于当前正式全局基线：
     - 先判断 candidate 是否仍兼容当前正式全局基线
     - 若兼容，允许自动把 `system_constraints_stable_ref` 更新到当前版本；该动作仅属于机械性基线绑定对齐，不得顺手修改 candidate 的任何其它真相内容
     - 若不兼容，结论只能是 `blocked` 或 `fix_required`
   - 若当前正式全局基线尚不存在：
     - candidate 中的 `system_constraints_stable_ref` 必须显式为 `none`
     - 若不是 `none`，结论只能是 `blocked` 或 `fix_required`
11. 处理 `shared_appendix_refs` 时，固定按以下分支执行：
   - 若模块正文当前层行为明确依赖共享附属展开文件，而 `shared_appendix_refs` 缺失、写成 `none` 或版本绑定不完整，结论只能是 `blocked` 或 `fix_required`
   - 若 `shared_appendix_refs` 已登记共享附属展开文件，但正文没有说明当前是哪条行为链路复用它，结论只能是 `blocked` 或 `fix_required`
   - 若 `shared_mechanism_reuse_summary` 与 `shared_appendix_refs`、`system_constraints_stable_ref` 或正文依赖关系矛盾，结论只能是 `blocked` 或 `fix_required`
11A. 处理 shared 候选信号时，固定按以下分支执行：
   - 若当前正文显式宣称“通用 / 统一 / 多模块共用 / 其它模块复用”，或正在定义共享输出协议、共享 fallback、共享对象展开、共享 few-shot、共享失败语义等高共享倾向对象，默认记为 `shared candidate hint`
   - 若当前正文与已有 shared 或当前命令必读范围内的其它正式真相出现明显重名、重职责、重协议语义，也默认记为 `shared candidate hint`
   - 发现 `shared candidate hint` 时，默认只提示用户可能存在 shared 候选，并建议按 `docs/agent_guidelines/shared_extract_review.md` 继续审查；不得自动扩到当前命令必读范围之外，也不得自动创建 `c_shared_xxx`
   - 若当前命令必读范围内已经能确认双真相源，例如当前模块正在重新定义一个已存在的 shared 正式真相，或继续保留模块内定义会直接造成正式语义冲突，则该问题必须直接作为当前命令的阻塞项报出，不等待用户二次确认
12. 形成总体结论时，固定按以下顺序处理：
   - 先给出 `可推进性` 结论
   - 再给出 `内容完整性` 结论
   - 最后合并成总体门禁结论
13. 合并规则固定如下：
   - 若 `可推进性` 未通过，结论只能是 `blocked` 或 `fix_required`
   - 若 `可推进性` 通过，但 `内容完整性` 存在任何 `关键层` 缺口，结论只能是 `blocked` 或 `fix_required`
   - 若 `可推进性` 通过，且 `内容完整性` 只有 `重要层 / 展开层` 问题，允许继续判断是否 `pass`
   - `重要层` 问题默认按 `P2` 处理，不单独阻塞
   - `展开层` 问题默认按 `P3` 处理，或可不报
14. 不允许：
   - 用“章节都在”替代真实收口
   - 用“本轮改动点已写清”覆盖关键真相缺口
   - 把 `关键层` 问题降成 `重要层 / 展开层`
   - 把 `展开层` 问题上纲成阻塞项
15. 若通过，则允许进入 `cand_plan`
16. 若为 `blocked` 或 `fix_required`，则不得进入 `cand_plan`
17. 若为 `blocked` 或 `fix_required`，后续输出中的 `findings` 必须满足本文件第 `6` 节定义的固定契约
17A. 若只命中 `shared candidate hint`，但当前命令必读范围内尚未形成已确认的 shared 双真相源，则该信号默认不得单独阻塞 `cand_check`；执行者应在输出中明确提示用户，是否需要进一步进入 `shared_extract_review`
18. 若结论为 `pass`，创建或更新 `docs/specs/_check_result/{module}.md`，作为当前 candidate 的候选链放行凭证，至少写入：
   - `module={module}`
   - `gate=cand_check`
   - `decision=pass`
   - `allow_next=true`
   - `next_command=cand_plan`
   - `blocking_summary=none`
   - `coverage_summary=n/a`
   - `prompt_adequacy_review_required=true|false`
   - `prompt_adequacy_decision=pass|n/a`
   - `prompt_adequacy_summary=...`，且其正文必须满足 `spec_policy.md` 中约定的最小语义契约
   - `spec_layer_ref=candidate`
   - `spec_file_ref=docs/specs/candidate/c_{module}.md`
   - `spec_version_ref=...`
   - `spec_fingerprint=...`
   - `system_constraints_stable_file_ref=docs/specs/system/stable/s_system_constraints.md|none`
   - `system_constraints_stable_version_ref=...|none`
   - `system_constraints_stable_fingerprint=...|none`
   - `shared_appendix_snapshot=...|none`，且其值必须按 `spec_policy.md` 第 `12.1` 节的规范化口径从当前绑定 Shared Appendix 生成
19. 若结论不是 `pass`：
   - 不得写入失败态 `_check_result/{module}.md`
   - 若仓库中存在旧的 `_check_result/{module}.md`，且其放行条件已不再成立，必须删除该文件
20. 更新 `_status.md`：
   - 若当前可进入 `cand_plan`，则 `Next Command=cand_plan`
   - 若仍未收口，则保持 `Next Command=cand_check`
21. 若本轮改动命中 Git 收口规则文件的提交触发条件，必须按该规则判断并完成当前任务内的 git 收口。

## 5. Stop Conditions

1. 本轮已明确 candidate 是否同时满足 `可推进性` 与 `内容完整性`，从而决定是否可进入 `cand_plan`。
2. 若本轮通过，当前 pass gate 已回写到 `_check_result/{module}.md`；若未通过，仓库中不得残留无效的失败态或过期放行结果。
3. `_status.md` 已同步更新。

## 6. Finding Contract

当 `cand_check` 结论为 `blocked` 或 `fix_required` 时，`findings` 必须按固定结构输出，不得只写抽象结论，也不得只给一句临时修法。

### 6.1 Severity And Blocking

`cand_check` 发现的问题，统一按 `P0 / P1 / P2 / P3` 分级。

补充说明：

1. `可推进性` 与 `内容完整性` 是并列阻塞门槛，不是二选一。
2. 但执行顺序固定为：先判 `可推进性`，再判 `内容完整性`。
3. `内容完整性` 的分层只用于避免过度挑刺，不得拿它稀释真实阻塞缺口。

#### 6.1.1 P0

定义：

1. candidate 已出现主链路断裂、真相冲突或关键门禁失真。
2. 执行者无法稳定判断应该按哪一条流程、协议或验收口径落地。
3. 若继续进入 `cand_plan`，后续计划与实现几乎必然围绕错误真相展开。

典型例子：

1. 同一条主流程在固定顺序、状态机、验收标准三处写成不同版本。
2. 关键协议对象在验收标准里被引用，但 candidate 中根本没有正式定义。
3. `system_constraints` 对齐关系自相矛盾，导致是否兼容正式基线无法判断。
4. 关键行为依据同时在 candidate 内外各有一套说法，执行者无法判断哪一套才是正式真相。

处理要求：

1. 存在任何 `P0`，结论必须为 `blocked`。
2. `P0` 不允许降级成非阻塞观察项。

#### 6.1.2 P1

定义：

1. 主链路还未完全断裂，但实现语义已经明显不稳。
2. 不同实现者、计划编写者或测试编写者很可能据此落出不同结果。
3. 当前问题虽未直接形成真相冲突，但已经不足以稳定约束 `cand_plan`。

典型例子：

1. Prompt 注入顺序写了大概顺序，却没把必注入块、可裁剪块和裁剪边界写实。
2. 状态机节点存在，但进入条件、跳过条件或回退条件仍留给实现阶段自行发明。
3. 验收标准覆盖到了某条链路，但断言口径仍依赖“等价实现”之类未落成协议的说法。
4. 主流程已经写出，但决定结果成立的关键依据没有正式落点，只能靠实现者补全。
5. candidate 描述了一个行为结果，却没有写清哪些事实决定该结果，因此不同实现者可能补出不同实现语义。

处理要求：

1. 默认阻塞当前放行，结论至少为 `fix_required`。
2. 只有在明确说明“不影响当前进入 `cand_plan` 的唯一链路”时，才允许降级成非阻塞项。

#### 6.1.3 P2

定义：

1. 不直接改变当前是否允许进入 `cand_plan` 的结论。
2. 当前 candidate 仍可稳定约束计划与实现。
3. 问题主要影响复审效率、可读性或后续维护成本。

典型例子：

1. 某一约束面表述偏绕，但上下文仍足以稳定理解。
2. 同类条目之间的模板不够统一，但不会导致实现分叉。
3. 关键行为真相已经具备，但部分边界说明仍偏散，主要影响复审效率和后续维护。

处理要求：

1. 默认不阻塞当前结论。
2. 可作为后续优化项记录。

#### 6.1.4 P3

定义：

1. 仅属于低风险表述优化或展示层瑕疵。
2. 不改变 candidate 的约束能力，也不影响复审判断。

典型例子：

1. 字段顺序不够统一，但语义完整。
2. 某些例子可更贴切，但不影响规则落地。
3. 关键真相已被正式承认，只是例子、展开说明或章节组织仍可优化。

处理要求：

1. 默认不阻塞。
2. 不得用 `P3` 包装真实收口缺口。

#### 6.1.5 Blocking Rule

默认阻塞规则如下：

1. 存在任何 `P0`，结论必须为 `blocked`。
2. 若不存在 `P0`，但存在任何阻塞态 `P1`，结论必须为 `fix_required`。
3. `P2` 与 `P3` 不得单独导致 `blocked` 或 `fix_required`。
4. `blocking=true` 只表示“该条问题在当前轮是否阻塞放行”；它不能替代优先级。
5. `可推进性` 未通过，或 `内容完整性` 存在任何 `关键层` 缺口时，不得写出 `pass`。
6. `内容完整性` 中的 `重要层` 与 `展开层` 问题，不得单独推翻已成立的 `可推进性` 结论。

### 6.2 Allowed Categories

`category` 不允许自由发挥，只允许使用以下分类：

1. `Spec Coverage`
2. `State Machine / Flow`
3. `Data Structures / Protocols`
4. `Edge Cases / Error Handling`
5. `Acceptance Criteria`
6. `Global Constraint Alignment`
7. `Prompt Adequacy`
8. `Shared Extraction Boundary`

这些分类的目的，是让每条 finding 能直接对应 candidate 的收口面，而不是漂成泛化评论。

### 6.3 Required Fields

每条 finding 至少必须包含以下字段：

1. `priority`
2. `title`
3. `category`
4. `background`
5. `what_happened`
6. `impact`
7. `best_recommendation`
8. `why_best`
9. `blocking`
10. `constraint_layer`

字段语义固定如下：

1. `priority`
   - 只能使用本节第 `6.1` 节定义的 `P0 | P1 | P2 | P3`。
2. `title`
   - 一句话概括当前缺口或冲突本身。
3. `category`
   - 只能使用本节第 `6.2` 节允许的分类。
4. `background`
   - 必须写“当前问题对象本身”的上下文背景，而不是写审查动作本身的背景。
   - 它的目标，是让不熟悉当前模块的人先看懂：这条问题讨论的对象是什么，它为什么存在，它在当前 candidate 中原本要解决什么约束问题。
   - `background` 不预设单一模板；可以根据问题类型写成模块上下文、流程上下文、协议上下文、对象语义上下文、约束来源上下文等，只要能帮助读者理解“这个问题为什么成立”即可。
   - 允许包含但不强制包含以下信息：
     - 当前对象位于哪条链路、哪类协议或哪组约束里
     - 当前对象原本要承接什么职责、语义或判定工作
     - 该问题为何会在当前模块中出现，而不是一个脱离上下文的孤立缺陷
   - 禁止写成“`cand_check` 需要确认什么”或其它审查者视角的流程说明；那属于审查背景，不属于问题背景。
5. `what_happened`
   - 说明 candidate 当前具体缺了什么、冲突在哪里，或哪里仍依赖实现阶段临时发明。
6. `impact`
   - 必须明确说明它会影响哪条链路、哪个对象、哪类实现判断或哪类测试断言。
   - 若 `blocking=true`，还必须说明为什么这会阻塞 `cand_plan`，或为什么它会让后续实现产生不稳定分叉。
7. `best_recommendation`
   - 必须给出能恢复 candidate 正确性与完整性的最佳修复建议，不得写成“为了通过本轮审查补一小段说明”这类临时补洞动作。
8. `why_best`
   - 必须说明为什么这是当前问题下最合适的修复路径，而不是局部补丁。
9. `blocking`
   - 固定写 `true | false`。
10. `constraint_layer`
   - 只能使用 `critical | important | elaboration`。
   - 用于标记当前问题在 `内容完整性` 里的层级。
   - 若当前问题不属于内容完整性缺口，而是纯粹的流程冲突或基线问题，也必须显式写明其最接近的层级判断；默认仍按是否影响实现结果来选取。

补充约束：

1. 凡涉及 `内容完整性` 的 finding，必须先判断属于 `关键层 / 重要层 / 展开层` 的哪一层，再决定优先级和是否阻塞。
2. `Spec Coverage` 不只回答“有没有提到”，还必须判断“提到的层级是否足以构成关键真相承认”。

### 6.4 Recommendation Rule

关于 `best_recommendation`，固定遵守以下规则：

1. 目标是恢复 candidate 的完整约束能力，不是以最低成本通过当前门禁。
2. 不得给出补丁式修法，例如：
   - 只补一句临时备注
   - 只补能让当前审查话术成立的局部描述
   - 只覆盖当前暴露出来的一处表面缺口，但继续保留根因
3. 所谓“最佳”，指当前问题下最能恢复逻辑正确性与收口完整性的修复路径，不等于改动最大，也不等于过度设计。

### 6.5 Markdown Rendering Contract

`findings` 在 Markdown 中必须按“列表项 + 字段分行”的形式输出，固定写法如下：

1. 每条 finding 必须是 `findings` 下的一个有序列表项，例如 `1.`、`2.`。
2. 列表项首行只允许写 finding 标题，格式固定为：`**[P1] 标题**`。
3. 从第二行开始，所有字段都必须独立成行，使用 `字段名：字段内容` 的形式书写。
4. `category`、`background`、`what_happened`、`impact`、`best_recommendation`、`why_best`、`blocking` 都必须显式出现，不得省略。
5. `constraint_layer` 也必须显式出现，不得省略。
6. 不得额外输出 `priority：P1` 这一独立字段行；优先级只放在标题前缀中，避免视觉重复。
7. `blocking` 必须单独成行，格式固定为 `blocking：true` 或 `blocking：false`。
8. `constraint_layer` 必须单独成行，格式固定为 `constraint_layer：critical|important|elaboration`。

最小示例如下：

```md
findings:

1. **[P1] 主流程失败分支仍依赖实现阶段发明**
   category：State Machine / Flow
   background：当前问题对应模块的主流程定义。这里原本应该给出一次完整执行会经历的状态、关键分支和回退路径，让计划拆解、编码和测试都围绕同一条控制流展开。
   what_happened：当前 candidate 只描述了成功路径，没有定义失败分支的回退条件和停止条件，导致不同实现者可能补出不同控制流。
   impact：这会直接影响 `cand_plan` 的任务拆解，并让后续实现无法稳定对齐同一份 candidate，因此当前不能放行。
   best_recommendation：补齐成功、失败、跳过或中断等关键分支的进入条件、状态归属和回退路径，使实现者无需在实现阶段自行发明控制流。
   why_best：当前缺口不是单句文案不足，而是主流程约束面不完整；只补局部备注会继续留下实现分叉空间。
   blocking：true
   constraint_layer：critical
```

## 7. Output Contract

1. 总体结论
2. 严重度汇总：
   - 至少统计 `P0`
   - 至少统计 `P1`
   - 至少统计 `P2`
   - 至少统计 `P3`
   - 至少统计 `blocking_count`
   - 若没有任何问题，也应显式写 `P0=0, P1=0, P2=0, P3=0, blocking_count=0`
3. 正式全局基线状态匹配结果
4. `Prompt Adequacy Review` 结果：
   - 若为 `n/a`，必须说明为什么该模块不命中触发条件
   - 该说明至少要覆盖：
     - 当前模块主要约束对象是什么
     - 为什么当前 candidate 不涉及 Prompt 设计、Prompt 组装或模型理解依赖链路
     - 因此哪些 Prompt 子审查项不适用
   - 若为 `blocked` 或 `fix_required`，必须明确缺的是：
     - `基础充分性审查`
     - `结构化输出审查`
     - `排序审查`
   - `prompt_adequacy_summary` 必须满足 `spec_policy.md` 中约定的最小语义契约，至少写清：
     - 三个审查对象哪些适用
     - 各自结论是什么
     - 若命中结构化输出审查，`Few-shot Example Requirement` 是否被触发且是否满足
     - 当前阻塞项是什么；若无，则写 `none`
5. 若 `system_constraints_stable_ref` 与当前正式全局基线状态不一致，必须明确说明是：
   - 已由本轮 `cand_check` 自动完成机械性基线绑定对齐
   - 当前 candidate 与新基线不兼容
   - 当前尚无正式全局基线，因此本轮要求并确认 `system_constraints_stable_ref=none`
6. `双门槛结论`：
   - 必须先单独说明 `可推进性` 结论
   - 再单独说明 `内容完整性` 结论
   - 最后说明总体门禁结论
7. 若通过，说明 `Check Result Snapshot` 回写结果；若未通过，明确说明本轮未生成 pass gate，且是否清理了旧 gate
8. 若结论为 `blocked` 或 `fix_required`：
   - 必须输出满足本文件第 `6` 节契约的结构化 `findings`
   - `findings` 必须先按严重度从高到低排序，再按阻塞性优先，最后按主链路优先排序
   - 若没有结构化 `findings`，不得视为合格输出
9. 若结论为 `pass`：
   - 不要求把输出扩写成冗长审查报告
   - 维持当前放行结论、关键摘要与 `Check Result Snapshot` 回写说明即可
10. `下一步建议`
   - 若结论为 `blocked` 或 `fix_required`，这里应汇总当前应优先完成的修复方向，不得简单重复 `findings` 原文
11. Git 收口结果
12. `_status.md` 更新结果

## 8. Non-Goals

1. 直接生成 `plan`
2. 直接进入代码实现
3. 创建、更新或删除独立的 `system_constraints` candidate 文件
4. 要求所有模块无差别地设计 Prompt；本命令只对命中 Prompt 触发条件的模块执行 Prompt Adequacy Review

## 9. Examples

```md
cand_check:module_example
```
