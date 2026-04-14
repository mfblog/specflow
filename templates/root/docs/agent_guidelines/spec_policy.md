# Spec 驱动开发策略（Spec-Driven Development Policy）

## 1. Purpose

本文件定义本仓库中正式模块 Spec 与全局系统约束 Spec 的工作方式。

它要回答四件事：

1. 正式模块的 Spec 与全局系统约束 Spec 由哪些对象组成。
2. 这些对象在仓库里由哪些文件承载。
3. Agent 在实现、审查、升级前应如何读取和使用这些文件。
4. 什么样的 candidate Spec 才足以驱动后续计划与实现。

本文件是给执行者直接使用的规则文档，不依赖其它上下文讨论。

---

## 2. Core Objects

### 2.1 Version Layers

正式模块默认只有两个版本层：

1. `stable`
   - 当前正式生效的模块真相。
2. `candidate`
   - 未来准备落地的候选真相。

版本层回答的是：

1. 当前正式基线是什么。
2. 当前工作是在对齐正式版，还是在推进下一版。

### 2.2 System Constraints

`system_constraints` 是全局唯一的系统级约束对象。  
它不是普通模块，也不进入 `docs/specs/_status.md`。

它回答的是：

1. 当前项目正式承认的技术基线是什么。
2. 当前项目有哪些共享机制应优先复用。
3. 遇到某类工程问题时，默认优先选择什么方案。
4. 哪些全局做法被禁止，哪些例外必须显式登记。

它不回答：

1. 单个模块内部状态机怎么设计。
2. 单个模块内部实现如何拆函数。
3. 完整系统拓扑图必须一次性长成什么样。

`system_constraints` 只有一个正式生效层：

1. `docs/specs/system/stable/s_system_constraints.md`
   - 当前正式生效的全局系统约束。

它不对应自己的 `_check_result`、`_plans`、`_verify_result` 这类过程文件。  
它也没有独立的 candidate 文件。

它的生命周期规则固定如下：

1. 它不是命令入口目标，用户不应直接执行 `*:system_constraints`。
2. 当模块命令需要判断当前正式技术基线、共享机制或全局例外，且该文件已存在时，应读取 `docs/specs/system/stable/s_system_constraints.md` 作为对应场景的正式基线输入。
3. 模块若需要提出新的全局约束变化，只能把提案写在自己的 candidate 中。
4. `docs/specs/system/stable/s_system_constraints.md` 只允许在模块 `cand_promote:{module}` 时，作为联动副产品被首次创建或后续更新。
5. 因此，`system_constraints` 的演进依附于模块命令链，不单独形成一条独立命令路径。

### 2.3 Module

`module_xxx` 是正式模块名。  
它是命令入口和状态表中的唯一稳定模块标识，不等于具体文件名。

### 2.4 Module Spec File

一个模块在一个版本层中只对应一份正式 Spec 文件：

1. `stable` 对应 `s_{module}.md`
2. `candidate` 对应 `c_{module}.md`

一份正式 Spec 文件必须同时承载：

1. 模块目标与边界
2. 关键术语
3. 数据结构与协议
4. 状态机与主流程
5. 边界条件与错误处理
6. 可验证性与验收标准

### 2.4A Formal Module vs Supporting File

仓库中的 Spec 文件默认分成三类：`正式模块文件`、`单模块附属展开文件` 与 `共享附属展开文件`。

#### 2.4A.1 正式模块文件

只有同时满足以下条件的对象，才算正式模块：

1. 模块名是稳定的 `module_xxx`。
2. 它是标准命令 `{command}:{module}` 的合法目标。
3. 它在 `docs/specs/_status.md` 中登记为一行正式模块状态。
4. 它有自己对应版本层的正式 Spec 文件：
   - `docs/specs/stable/s_{module}.md`
   - 或 `docs/specs/candidate/c_{module}.md`
5. 该正式 Spec 文件正文本身必须满足本文件第 `4` 节定义的正式模块最小内容要求，而不是把核心行为真相长期外包给附属文件代写。

补充说明：

1. “放在 `docs/specs/candidate/` 目录下”本身，不足以证明它是正式模块。
2. “文件名长得像 `c_module_xxx.md`”本身，也不足以证明它是正式模块。
3. 只有命令入口、状态登记和正式 Spec 文件三者同时成立时，才视为正式模块。

#### 2.4A.2 单模块附属展开文件

以下对象属于单模块附属展开文件，而不是正式模块：

1. 某个正式模块的附录、专题展开、Prompt 原文模板、补充示例、复杂对象展开说明。
2. 它们可以帮助解释主模块，但不单独进入命令链。
3. 它们不得单独进入 `docs/specs/_status.md`。
4. 它们不得生成自己的 `_check_result`、`_plans`、`_verify_result`。

单模块附属展开文件允许：

1. 与主模块一起被 `cand_check`、`cand_plan`、`cand_impl`、`cand_verify` 读取。
2. 被主模块正文明确引用为“共同约束”或“详细展开”。
3. 存放在 `docs/specs/candidate/`、`docs/specs/stable/` 的专用子目录中，例如 `appendix/`。

单模块附属展开文件不允许：

1. 使用 `id: module_xxx` 把自己伪装成独立正式模块。
2. 仅靠 frontmatter 中看似模块名的字段，就要求进入 `_status.md`。
3. 在主模块未正式切边界前，擅自变成独立命令目标。
4. 继续放在 `docs/specs/candidate/` 或 `docs/specs/stable/` 根目录，与正式模块主文件混放。

#### 2.4A.2A 单模块附属展开文件的目录规则

单模块附属展开文件的目录规则固定如下：

1. `candidate` 层附属文件必须放在 `docs/specs/candidate/appendix/` 或等价专用子目录。
2. `stable` 层附属文件必须放在 `docs/specs/stable/appendix/` 或等价专用子目录。
3. `docs/specs/candidate/` 与 `docs/specs/stable/` 根目录默认只放正式模块主文件：
   - `c_{module}.md`
   - `s_{module}.md`
4. 若某个文件属于附属展开文件，却仍放在根目录，视为目录漂移。
5. 目录漂移的修复责任固定归属于“当前命令链里第一个在前置自检或必读环节发现该漂移的标准命令”：
   - 该命令必须负责把文件迁移到对应 `appendix/` 或等价专用子目录，并同步修正主文件引用
   - 迁移完成后，必须重新执行本命令的前置自检与必读步骤，确认漂移已消失后才允许继续
6. 不允许把目录漂移留给一个未定义的“治理修正规则”或下游命令兜底处理。

这样做的目的固定只有两个：

1. 让执行者只看路径就能先区分“主文件”与“附属文件”。
2. 避免附属文件因为文件名像模块而被误读成独立正式模块。

#### 2.4A.2B 单模块附属展开文件的读取与清理责任

单模块附属展开文件一旦被主模块正文明确引用为“共同约束”或“详细展开”，就进入当前版本层的真相读取面。

固定规则如下：

1. 执行者不得只读取主文件而跳过被主文件明确引用的附属展开文件。
2. `candidate` 层被引用的附属展开文件，必须随 `cand_check`、`cand_plan`、`cand_impl`、`cand_verify` 一起读取。
3. `stable` 层被引用的附属展开文件，必须随 `stable_verify`、`spec_fork` 一起读取。
4. `spec_fork` 开启新一轮 candidate 时，若存在上一轮 candidate 附属展开文件，必须先删除或按本轮内容重建，禁止把旧轮次 appendix 直接沿用为新轮次真相。
5. `cand_promote` 完成候选提升时，必须同步处理本轮 candidate 附属展开文件：
   - 若其内容已并入新的 stable 主文件或 stable 附属展开文件，则删除旧的 candidate 附属展开文件
   - 若本轮提升后 `Candidate=no`，则不得残留该模块未归属的 candidate 附属展开文件
6. `Candidate=no` 时，不得残留属于该模块当前轮 candidate 的附属展开文件。

#### 2.4A.3 单模块附属展开文件的 frontmatter 约束

单模块附属展开文件若需要 frontmatter，默认应至少使用：

1. `module: module_xxx`
   - 表示它归属哪个正式模块
2. `layer: stable | candidate`
   - 表示它服务哪个版本层
3. `spec_version_ref: s_{module}@... | c_{module}@...`
   - 表示它当前绑定哪一份主 Spec

默认不应再使用：

1. `id: module_xxx`
   - 因为该字段会被误读为“独立正式模块标识”

若某个附属展开文件未来需要升格为正式模块，必须先：

1. 明确新的正式模块边界。
2. 为它创建自己的 `c_{module}.md` 或 `s_{module}.md`。
3. 再通过正式模块命令链把它纳入 `_status.md`。

#### 2.4A.4 共享附属展开文件（Shared Appendix）

以下对象属于共享附属展开文件，而不是正式模块：

1. 被多个正式模块共同引用的共享协议正文、共享输出协议、共享对象展开说明、共享失败语义、共享 few-shot 或共享复用边界说明。
2. 它们服务于多个模块，但不单独进入 `_status.md`，也不形成自己的命令链。
3. 它们是正式真相对象，不是随手参考文档。

共享附属展开文件固定语义如下：

1. 它不是正式模块，也不是 `{command}:{module}` 命令目标。
2. 它允许存在 `candidate` 与 `stable` 两层。
3. 它只有在被某模块当前层的 `Global Constraint Alignment` 中通过 `shared_appendix_refs` 显式绑定后，才进入该模块当前层的默认真相读取面。
4. 它的生命周期依附于引用它的模块命令链，不单独形成一条独立命令路径。
5. `bound_modules` 只用于声明“这份共享对象预期服务哪些模块”，不能替代模块当前层 `shared_appendix_refs` 的正式绑定语义。

共享边界补充规则如下：

1. `shared` 不是“未来可能复用的内容”，而是多个正式模块共同依赖、且应只有一份正式定义的正式行为真相。
2. 第一次出现的内容，默认先留在当前模块主文件或模块 appendix 中；不得仅因“将来可能复用”就提前抽成 shared。
3. 只有当第二个正式模块出现同一条正式行为真相需求时，才进入 shared 候选判定。
4. shared 候选的核心边界固定为：这条内容是否应在仓库里只有一份正式定义；若只是主题相似、实现类似或结构接近，则不构成 shared。
5. `shared` 采用“一个共享对象一个文件”的模型；同一仓库可同时存在多个 `c_shared_xxx.md` 与 `s_shared_xxx.md`。
6. 不允许把多个不相干共享主题长期堆进一个总 shared 文件，只因为它们都属于 shared。
7. shared 提取边界的正式审查流程固定为 `docs/agent_guidelines/shared_extract_review.md`；模块命令只负责在命中条件时提示或报出当前范围内已确认的冲突，不替代该流程做完整跨模块判定。

#### 2.4A.4A 共享附属展开文件的目录规则

共享附属展开文件的目录规则固定如下：

1. `candidate` 层共享附属展开文件必须放在 `docs/specs/shared/candidate/`。
2. `stable` 层共享附属展开文件必须放在 `docs/specs/shared/stable/`。
3. 它们不得继续放在某个模块的 `appendix/` 下伪装成单模块附录。
4. 若某个共享附属展开文件仍放在某个模块 appendix 路径下，视为目录漂移；发现该漂移的标准命令必须先完成迁移并修正引用，之后才允许继续下游流程。

#### 2.4A.4B 共享附属展开文件的读取、失效与清理责任

共享附属展开文件固定规则如下：

1. 若某模块当前层 `Global Constraint Alignment.shared_appendix_refs` 非空，则执行者在读取该模块当前层真相时，必须一并读取这些共享附属展开文件。
2. `cand_check`、`cand_plan`、`cand_impl`、`cand_verify`、`stable_verify`、`spec_fork` 都不得只读模块主文件而跳过已绑定的共享附属展开文件。
3. 若共享附属展开文件的当前层正文、版本引用或绑定关系变化，所有仍绑定旧 Shared Appendix 快照的模块候选侧过程文件默认失效，并统一回退为 `cand_check`。
4. 若 stable 层共享附属展开文件变化，所有声明“当前仍对齐 stable”的相关核对默认都必须重新读取并重新判断。
5. 共享附属展开文件不得按“某个模块 promote 完就顺手删除”处理；只有在无模块继续绑定、或已被新共享附属展开文件替代、或已被正式全局基线完整吸收且不再需要展开正文时，才允许清理。
6. 若 `bound_modules` 与模块当前层 `shared_appendix_refs` 的真实绑定集合不一致，属于治理漂移；发现该绑定变化并正在修改模块当前层 `shared_appendix_refs`、或正在创建 / 更新对应 Shared Appendix 的当前命令，必须在同一次命令内同步修正受影响 Shared Appendix 的 `bound_modules`；`shared_flow_reconcile` 与 `spec_flow_review` 负责继续报出漏修情况，但不替代当前命令的主维护责任。
7. 只要当前任务改动了 `docs/specs/shared/**`，或改动了任一模块当前层 `shared_appendix_refs`，在宣称状态已收口前都必须完成一次 Shared Appendix 状态收口：
   - 若当前标准命令已经基于最新 Shared Appendix 真相，为当前目标模块重算并写回新的过程文件绑定快照，或已把当前目标模块直接收口到新的 stable 真相，则该目标模块视为已在当前命令内完成收口
   - 对其余未在当前命令内直接收口、但已受影响的模块，必须执行 `shared_flow_reconcile`
8. 只要模块当前层 `shared_appendix_refs` 新增、删除或切换了某个 Shared Appendix 绑定，当前命令除更新模块主文件外，还必须同步维护对应 Shared Appendix frontmatter 中的 `bound_modules`：
   - 新增绑定时，把当前模块加入对应 Shared Appendix 的 `bound_modules`
   - 删除绑定时，把当前模块从不再绑定的 Shared Appendix 的 `bound_modules` 中移除
   - 切换绑定版本或切换到另一份 Shared Appendix 时，同时完成旧绑定移除与新绑定加入
9. 上一条中的 `bound_modules` 同步，只回答“当前正文预期服务哪些模块”，不得替代模块当前层 `shared_appendix_refs` 的正式绑定语义。

#### 2.4A.4C 共享附属展开文件的 frontmatter 约束

共享附属展开文件若需要 frontmatter，默认应至少使用：

1. `shared_id: shared_xxx`
   - 表示共享对象的稳定标识
2. `layer: stable | candidate`
   - 表示它服务哪个版本层
3. `shared_version: <semver>`
   - 表示当前共享对象的版本
4. `bound_modules`
   - 表示当前正文期望服务哪些模块
5. `system_constraints_stable_ref`
   - 表示当前共享对象写作时绑定的正式全局基线版本；若尚无正式全局基线，则写 `none`

补充规则：

1. `bound_modules` 是声明性清单，不是门禁绑定来源；模块当前层是否真正绑定该共享对象，只以模块当前层 `shared_appendix_refs` 为准。
2. 共享附属展开文件的失效判断，默认同时比较：
   - 模块当前层 `shared_appendix_refs` 的版本引用
   - 共享附属展开文件当前正文指纹
   - 模块当前层与共享附属展开文件当前 `layer` 是否匹配

默认不应使用：

1. `id: module_xxx`
2. `module: module_xxx`
3. `spec_version_ref: c_{module}@... | s_{module}@...`

因为这些字段都会把共享对象误读成单模块附录。

### 2.5 Check Result File

`_check_result/{module}.md` 是当前 candidate 进入后续候选链前的放行凭证文件。

它只负责回答：

1. 当前 candidate 是否已被放行进入后续候选链
2. 这份放行凭证绑定的是哪一份 candidate
3. 这份放行凭证绑定的是哪一版正式全局基线
4. 当前 `cand_plan` 是否仍可依赖它继续推进

它不是正式行为真相源，也不是实施计划文件。

### 2.6 Verify Result File

`_verify_result/{module}.md` 是当前 candidate 最近一次实现验证结果文件。

它只负责回答：

1. 当前 `cand_verify` 的结论是什么
2. 是否允许进入 `cand_promote`
3. 当前阻塞摘要是什么
4. 这份结果绑定的是哪一份 candidate

它不是正式行为真相源，也不是实施计划文件。

### 2.7 Plan File

`_plans/{module}.md` 是当前候选轮次的实施计划文件。

它只负责回答：

1. 当前轮准备先实现什么
2. 实现顺序和依赖是什么
3. 代码落点和验证重点是什么
4. 当前轮暂不处理什么

它不是正式行为真相源，也不承载门禁快照。

### 2.8 Status File

`docs/specs/_status.md` 是正式模块状态登记表。  
它是运行时索引文件，不是治理规则真相源。

它默认只记录五类最小状态事实：

1. 正式模块名
2. 是否存在 `stable`
3. 是否存在 `candidate`
4. 当前默认对齐层 `Active Layer`
5. 当前最小可行动作 `Next Command`

它不负责：

1. 承载模块真相
2. 承载计划内容
3. 承载 `cand_check` 或 `cand_verify` 的审查结论
4. 定义读取规则、门禁规则或回退规则正文

---

## 3. Repository File Model

当前仓库中的正式模块文件与系统级约束文件职责如下：

```text
docs/
  specs/
    system/
      stable/
        s_system_constraints.md
    shared/
      stable/
        s_shared_xxx.md
      candidate/
        c_shared_xxx.md
    stable/
      s_module_xxx.md
    candidate/
      c_module_xxx.md
    _check_result/
      module_xxx.md
    _verify_result/
      module_xxx.md
    _plans/
      module_xxx.md
    _status.md
```

各路径职责固定如下：

### 3.1 `docs/specs/system/stable/s_system_constraints.md`

1. 若存在，则承载当前正式生效的全局系统约束。
2. 它是“已经被 promote 沉淀出来之后”的正式工程基线，不是候选链默认必需的仓库初始化前置物。
3. 它必须包含 `frontmatter.version`，用于生成固定版本引用 `s_system_constraints@<frontmatter.version>`。

### 3.2 `docs/specs/stable/s_{module}.md`

1. 承载该模块当前正式版本的完整 Spec。
2. 它必须包含 `frontmatter.version`，用于生成固定版本引用 `s_{module}@<frontmatter.version>`。

### 3.3 `docs/specs/candidate/c_{module}.md`

1. 承载该模块当前候选版本的完整 Spec。
2. 它必须包含 `frontmatter.version`，用于生成固定版本引用 `c_{module}@<frontmatter.version>`。
3. 它必须在 `Global Constraint Alignment` 中显式记录当前对齐状态：
   - 若正式全局基线已存在，则记录对应版本
   - 若正式全局基线尚不存在，则显式写 `none`

### 3.3A `docs/specs/shared/stable/s_shared_xxx.md`

1. 若存在，则承载当前正式生效的共享附属展开真相。
2. 它不是正式模块，不进入 `_status.md`。
3. 它只有在模块当前层 `shared_appendix_refs` 显式绑定后，才进入对应模块的默认真相读取面。

### 3.3B `docs/specs/shared/candidate/c_shared_xxx.md`

1. 承载当前候选中的共享附属展开真相。
2. 它不是正式模块，不独立形成命令入口。
3. 它只有在模块当前层 `shared_appendix_refs` 显式绑定后，才进入对应模块的候选侧真相读取面。

### 3.4 `docs/specs/_check_result/{module}.md`

1. 承载该模块当前 candidate 最近一次通过 `cand_check` 后生成的 `cand_plan` 放行快照。
2. 它是候选收口门禁输入，不是正式行为真相源。
3. 它属于过程文件，只在当前 candidate 内容仍匹配当前正式全局基线状态时有效：
   - 若正式全局基线已存在，则要求版本与指纹仍匹配
   - 若正式全局基线尚不存在，则要求相关绑定字段保持 `none`

### 3.5 `docs/specs/_verify_result/{module}.md`

1. 承载该模块当前 candidate 最近一次 `cand_verify` 的结果。
2. 它是候选提升门禁输入，不是正式行为真相源。
3. 它属于过程文件，只在当前 candidate、当前实现上下文与当前正式全局基线状态仍匹配时有效。

### 3.6 `docs/specs/_plans/{module}.md`

1. 承载该模块当前候选轮次的实施计划与实施过程信息。
2. 它是候选推进输入，不是正式行为真相源。
3. 它属于过程文件，只在当前候选轮次与当前正式全局基线状态仍匹配时有效。

### 3.7 `docs/specs/_status.md`

它是正式模块的状态登记表，不是规则正文承载点。

#### 3.7.1 固定职责

它只回答四件事：

1. 模块有没有 `stable`
2. 模块有没有 `candidate`
3. 当前默认对齐哪一层
4. 下一步命令是什么

它不回答：

1. candidate 是否已经闭环
2. 实现是否已经对齐 candidate
3. 模块协议细节是什么
4. 过程文件为什么有效或无效
5. 命令为什么允许或不允许越过

#### 3.7.2 字段语义

`_status.md` 的表头字段固定按如下解释：

1. `Module`
   - 正式模块名 `module_xxx`
2. `Stable`
   - 当前是否存在正式层真相文件
3. `Candidate`
   - 当前是否存在候选层真相文件
4. `Active Layer`
   - 当前默认应对齐的版本层，只能指向 `stable` 或 `candidate`
5. `Next Command`
   - 当前最小可行动作，不是提示语，也不是建议列表
6. `Notes`
   - 只用于记录简短状态备注，不承载门禁规则正文

#### 3.7.3 与其它对象的关系

`_status.md` 与其它对象的关系固定如下：

1. 它不替代 `stable` 或 `candidate` 真相文件。
2. 它不替代 `_check_result`、`_plans`、`_verify_result` 的门禁与过程语义。
3. 它不记录 `system_constraints`，因为 `system_constraints` 不是普通模块，也不进入状态表。
4. 它只负责给执行者一个“当前默认落点”和“当前最小可行动作”的索引入口。

#### 3.7.4 读取时机

出现以下任一场景时，执行者必须读取 `_status.md`：

1. 需要判断目标模块当前默认对齐层。
2. 需要判断当前命令是否与 `Next Command` 一致。
3. 需要判断状态登记与实际文件、过程文件、实现上下文是否一致。
4. 需要决定用户只说模块名时，本次应先对齐 `stable` 还是 `candidate`。

补充说明：

1. 读取 `_status.md` 的目的，是取得状态索引，而不是读取治理规则正文。
2. `_status.md` 不单独定义与 `s_system_constraints.md` 或其它治理文件之间的统一总入口顺序。

#### 3.7.5 维护责任

默认维护规则如下：

1. `_status.md` 默认由标准命令同步维护。
2. 当执行 `spec_init`、`stable_verify`、`spec_new`、`spec_fork`、`cand_check`、`cand_plan`、`cand_impl`、`cand_verify`、`cand_promote` 后，必须同步更新该文件。
3. 人类可以审阅、纠正或覆盖该文件。
4. 该文件只登记正式模块。
5. 若发现状态登记和实际文件或实现上下文失配，触发该失配的当前命令负责修正，不能只报错不回写。

---

## 4. What a Formal Spec Must Contain

无论是 `stable` 还是 `candidate`，正式模块 Spec 正文默认都必须覆盖以下内容：

1. `Context & Motivation`
2. `Terminology`
3. `Data Structures / Protocols`
4. `State Machine / Business Flow`
5. `Edge Cases & Error Handling`
6. `Testability / Acceptance Criteria`

除了章节齐全之外，正式 Spec 还必须覆盖会影响实现结果的关键行为真相。

这里的“关键行为真相”固定指：

1. 缺失后会让不同实现者做出不同模块行为的事实。
2. 缺失后会让关键分支、关键归属、关键输入来源、关键错误语义或关键验收判断变得依赖执行者自行发明的事实。
3. 已经属于模块正式行为的一部分，但若不写入 Spec，就会继续漂在默认上下文、README 愿景、历史口头共识或作者隐含理解中的事实。

默认不按配置、持久化、环境变量、状态机或其它对象类型做封闭枚举；统一判断标准只有一条：

1. 该事实是否会影响实现结果，或影响实现者对关键决策的稳定判断。

若模块内容涉及技术选型、共享基础设施、跨模块复用、全局例外申请或全局约束提案，还应补充 `Global Constraint Alignment` 或等价章节。

该章节至少必须覆盖：

1. `system_constraints_stable_ref`
   - 取值只能是：
     - `s_system_constraints@<frontmatter.version>`
     - `none`
   - 若正式全局基线已存在，表示本模块当前正式或候选真相所依据的正式全局基线版本
   - 若正式全局基线尚不存在，必须显式写 `none`
2. `shared_appendix_refs`
   - 当前模块当前层显式绑定了哪些共享附属展开文件；没有则显式写 `none`
   - 取值必须使用稳定版本引用，例如：
     - `c_shared_xxx@<shared_version>`
     - `s_shared_xxx@<shared_version>`
   - 若当前模块层为 `candidate`，只能绑定 `c_shared_xxx@...`
   - 若当前模块层为 `stable`，只能绑定 `s_shared_xxx@...`
3. `shared_mechanism_reuse_summary`
   - 当前模块对正式全局基线与共享附属展开文件的复用摘要；没有额外复用则显式写 `none`
   - 它只负责摘要说明，不承载共享协议正文
4. `global_constraint_exceptions`
   - 当前是否存在例外；没有则显式写 `none`
5. `proposed_system_constraints_updates`
   - 当前模块提议修改哪些全局规则；没有则显式写 `none`
6. `promotion_to_system_stable`
   - 枚举值：`none | with_module`
   - `none` 表示本模块 promote 时不更新 `s_system_constraints.md`
   - `with_module` 表示本模块 promote 时把本模块 candidate 中的全局提案吸收到 `s_system_constraints.md`

补充规则：

1. 若模块正文当前层行为明确依赖某份共享附属展开文件，却未在 `shared_appendix_refs` 中登记，默认视为 `Global Constraint Alignment` 不完整。
2. 若 `shared_appendix_refs` 已登记某份共享附属展开文件，则执行者不得只在正文里写“本模块会复用某共享协议”而跳过对应版本绑定。
3. `shared_mechanism_reuse_summary` 只回答“当前模块复用了什么”，不回答“共享协议正文是什么”。
4. 若模块 stable 已写 `shared_appendix_refs=none`，则正文不得继续宣称自己仍依赖独立 Shared Appendix；此时共享规则必须已经被 stable 主文或 `s_system_constraints` 吸收。

允许：

1. 在不改变语义的前提下调整章节顺序
2. 为复杂模块增加补充章节，例如 `Facade`、`Examples`、`Integration Notes`

不允许：

1. 缺少上述核心内容，却直接进入计划或实现
2. 把“模块目标与边界”完全丢给实现阶段临时发明
3. 只写愿景、背景或备注，而没有真正的结构与行为定义
4. 把全局约束提案写到模块外部的独立 candidate 文件里

---

## 5. Implementation Constraint Rules

正式 Spec 除了要“章节齐全”，还必须满足“足以约束实施”的要求。

这里的“约束实施”不是要求 Spec 写成逐行伪代码，而是要求：

1. 实现者能够根据 Spec 稳定落地代码。
2. 不同实现者按同一份 Spec 实现时，用户可观察到的核心行为应保持一致。
3. `cand_plan` 只负责拆任务，不负责补写缺失的行为真相。
4. `cand_impl` 只负责按 Spec 与计划推进实现，不负责临时决定关键行为。

当模块处于 `cand_check` 时，候选收口固定同时检查两个门槛：

1. `可推进性`
   - candidate 是否足以稳定进入 `cand_plan`
   - 主流程、关键协议、关键边界、错误语义与验收口径是否足以避免后续计划和实现分叉
2. `内容完整性`
   - candidate 是否已经承认会影响实现结果的关键行为真相
   - 是否仍有关键依据留在文档外部、默认上下文、README 愿景、历史口头共识或作者隐含理解中

这两个门槛在阻塞语义上并列成立，任一关键失败都不得通过 `cand_check`；但执行顺序固定为：

1. 先判 `可推进性`
2. 再判 `内容完整性`
3. 最后合并总体门禁结论

### 5.1 Required Constraint Surface

正式 Spec 默认至少要把以下五类写清：

1. 输入输出语义
2. 主流程
3. 分支与边界
4. 错误语义
5. 验收口径

补充要求：

1. 上述五类只回答“最小约束面是什么”，不代表只要章节存在就等于已收口。
2. 只要某项事实会影响实现结果，它就属于必须被正式承认的关键行为真相。
3. 不得因为“这不是本轮新增改动”或“章节已经齐全”，就把既有关键真相排除在 candidate 收口面之外。

### 5.1A Completeness Layers

为避免 `cand_check` 滑向鸡蛋里挑骨头，`内容完整性` 固定分成三层：

1. `关键层`
   - 缺失后会改变实现结果
   - 缺失后不同实现者可能做出不同外部行为
   - 缺失后会影响主流程、关键分支、关键归属、关键输入来源、关键错误语义或关键验收判断
   - 默认至少对应阻塞态 `P1`
2. `重要层`
   - 不直接改变实现结果
   - 但会显著影响复审稳定性、后续维护、边界理解或复验效率
   - 若继续积累，容易在后续轮次演化成实现分叉
   - 默认对应 `P2`，不单独阻塞
3. `展开层`
   - 只影响表达友好度、例子充分性、阅读成本或章节观感
   - 不影响实现结果，也不影响复审结论
   - 默认对应 `P3`，或可不报

补充约束：

1. 不得把 `关键层` 问题降成 `重要层` 或 `展开层`。
2. 不得把 `展开层` 问题上纲成阻塞项。

### 5.2 Allowed Implementation Freedom

以下内容默认允许实现层自行决定，只要不改变外部行为语义：

1. 私有函数如何拆分
2. 局部命名与局部变量组织
3. 无语义差异的内部代码结构
4. 等价的内部调用编排
5. 不影响实现结果的展开顺序、例子丰富度和表达友好度

### 5.3 Blocking Gaps

以下留白默认属于阻塞项，不能视为“已收口”：

1. 会影响接口语义的留白
2. 会影响流程分支或关键步骤顺序的留白
3. 会影响状态归属或状态流转的留白
4. 会影响失败处理、跳过、中断、重试语义的留白
5. 会影响验收判断的留白
6. 未说明当前 candidate 对齐哪一版 `s_system_constraints`
7. 已经属于模块正式行为的一部分，但关键行为真相仍留在文档外部，导致实现者必须自行补发明
8. 只描述结果，不写清决定结果成立的关键依据，导致实现和验收口径不稳定

### 5.4 Non-Blocking Gaps

以下留白默认可不阻塞推进：

1. 只影响内部代码写法的留白
2. 不影响外部行为一致性的措辞留白
3. 不改变实现结果的结构美化建议
4. 不影响实现结果的例子补充、章节展开与阅读友好度优化

---

## 6. Lifecycle

正式模块的生命周期既包括版本层如何出现和消失，也包括过程文件如何生成和清理。  
`system_constraints` 只有正式基线，不参与自己的过程文件生命周期。

### 6.1 `system_constraints` 的正式基线

1. `docs/specs/system/stable/s_system_constraints.md`
   - 若存在，记录当前正式全局约束。

默认规则：

1. `s_system_constraints.md` 默认允许不存在；这表示仓库尚未沉淀出正式全局基线。
2. 在该文件不存在时，模块命令仍可继续推进；此时模块 candidate 的 `system_constraints_stable_ref` 必须显式写 `none`。
3. 模块若提议修改全局规则，只能把提案写在自己的 candidate 中。
4. `s_system_constraints.md` 只允许在模块 `cand_promote:{module}` 时，由该模块 candidate 中已收口且标记为 `promotion_to_system_stable: with_module` 的提案首次创建或后续更新。
5. 只要命令改动了 `s_system_constraints.md` 的正文内容，就必须同步递增其 `frontmatter.version`；只读不改时不得变更版本号。
6. 若 `s_system_constraints.md` 已存在，则所有模块至少不得违反它。
7. 若 `s_system_constraints.md` 的版本已前移，而某模块 candidate 中的 `system_constraints_stable_ref` 仍落后，则该模块候选链的过程文件默认失效，并统一回退为 `cand_check`。
8. 若模块 candidate 中 `shared_appendix_refs` 绑定的共享附属展开文件版本、正文或绑定关系已变化，则该模块候选链的过程文件默认失效，并统一回退为 `cand_check`。

### 6.1A 共享附属展开文件的分层绑定

共享附属展开文件的分层绑定固定如下：

1. `candidate` 模块若当前层依赖共享附属展开文件，只能绑定 `c_shared_xxx@...`。
2. `stable` 模块若当前层依赖共享附属展开文件，只能绑定 `s_shared_xxx@...`。
3. 不允许在 `candidate` 模块中继续直接绑定 `s_shared_xxx@...`，也不允许在 `stable` 模块中保留 `c_shared_xxx@...`。
4. 若某个已有 `stable` 的模块在执行 `spec_fork` 后仍计划继续依赖共享附属展开文件，则本轮必须先从对应 `s_shared_xxx` 派生出当前轮 `c_shared_xxx`，再让 candidate 显式绑定；不得直接沿用 stable 层引用。
5. 若当前轮尚未形成可读的 `c_shared_xxx`，则 candidate 不得先写“继续复用同一共享附录”再把具体绑定留给后续命令猜测。

### 6.1B 共享附属展开文件的进入与提升

共享附属展开文件的进入与提升固定如下：

1. `c_shared_xxx` 的创建或更新，属于当前轮共享机制演进的一部分；它可以与模块 candidate 同轮推进，但不形成独立命令链。
2. `s_shared_xxx` 只允许在模块 `cand_promote:{module}` 中作为联动副产品被首次创建或后续更新。
3. 共享附属展开文件不得独立 promote；它们只能在模块 `cand_promote` 中被联动迁移到 stable、被稳定结论吸收到 `s_system_constraints.md`，或被 stable 主文吸收后删除。

### 6.2 `stable` 的建立与更新

1. `spec_init:{module}`
   - 创建 `docs/specs/stable/s_{module}.md`
   - 若命中 `Global Constraint Alignment` 触发条件，则补齐该章节
2. `stable_verify:{module}`
   - 核对当前代码是否仍对齐 `docs/specs/stable/s_{module}.md`
   - 输出结构化验证证据
   - 更新 `_status.md`
3. `cand_promote:{module}`
   - 生成或更新 `docs/specs/stable/s_{module}.md`

### 6.3 `candidate` 的建立与重建

1. `spec_new:{module}`
   - 创建 `docs/specs/candidate/c_{module}.md`
   - 首版 candidate 版本固定从 `0.1.0` 起
   - 初始化 `Global Constraint Alignment`
2. `spec_fork:{module}`
   - 以当前 `stable` 为基线生成新的 `docs/specs/candidate/c_{module}.md`
   - 同步初始化本轮 candidate 版本
   - 同步写入当前 `system_constraints_stable_ref`
   - 若当前 stable 仍依赖共享附属展开文件，必须先把对应 `s_shared_xxx` 派生为当前轮 `c_shared_xxx`，再写入 candidate 的 `shared_appendix_refs`
   - 清理上一轮 candidate 的 `_check_result`、`_verify_result`、`_plans`

candidate 版本规则固定如下：

1. `spec_new:{module}` 创建首版 candidate 时，`frontmatter.version` 固定为 `0.1.0`。
2. `spec_fork:{module}` 创建新一轮 candidate 时，必须先确定“本轮计划提升成哪个正式版本”，再把该目标版本写入 candidate。
3. 若本轮目标是兼容性新增能力，默认把当前 `stable` 的 `MINOR` 加一，并把 `PATCH` 置为 `0`。
4. 若本轮目标是正式契约不兼容变化，默认把当前 `stable` 的 `MAJOR` 加一，并把 `MINOR`、`PATCH` 置为 `0`。
5. 若本轮只是对当前正式契约做兼容性修正或对齐，默认把当前 `stable` 的 `PATCH` 加一。
6. candidate 版本不是临时流水号，而是“本轮若成功 promote，默认将成为的新 stable 版本号”。
7. 若 candidate 在收口过程中确认版本级别判断错误，必须在 candidate 中显式改正版本号，并按过程文件失效规则回退后续门禁。

### 6.4 candidate 的检查与计划

1. `cand_check:{module}`
   - 检查 `docs/specs/candidate/c_{module}.md` 是否已收口
   - 检查 `system_constraints_stable_ref` 是否与当前正式全局基线状态一致
   - 检查 `shared_appendix_refs` 与当前共享附属展开文件绑定状态是否一致
   - 若正式全局基线已存在且版本已变化，但 candidate 仍兼容，允许 `cand_check` 仅自动前移 `system_constraints_stable_ref`
   - 若正式全局基线尚不存在，则要求 `system_constraints_stable_ref=none`
   - 若通过，则创建或更新 `docs/specs/_check_result/{module}.md` 作为放行凭证
   - 若未通过，则不得写入失败态 `_check_result/{module}.md`；若旧放行凭证已不再成立，必须删除
2. `cand_plan:{module}`
   - 读取 `docs/specs/_check_result/{module}.md` 放行凭证
   - 创建或更新 `docs/specs/_plans/{module}.md`

### 6.5 计划驱动实现与验证

1. `cand_impl:{module}`
   - 读取 `docs/specs/candidate/c_{module}.md`
   - 读取 `docs/specs/_plans/{module}.md`
   - 在实现推进后按规则回写计划文件
2. `cand_verify:{module}`
   - 核对当前代码是否对齐 `docs/specs/candidate/c_{module}.md`
   - 读取 `docs/specs/_plans/{module}.md`，确认当前轮计划门禁仍存在且仍绑定当前 candidate
   - 创建或更新 `docs/specs/_verify_result/{module}.md`

### 6.6 候选轮次结束与清理

1. `cand_promote:{module}`
   - 让当前 `candidate` 提升为新的 `stable`
   - 若该模块 candidate 中 `promotion_to_system_stable=with_module`，则同步首次创建或后续更新 `docs/specs/system/stable/s_system_constraints.md`
   - 若该模块 candidate 中 `shared_appendix_refs` 非空，必须在同一次 `cand_promote` 中逐项决定去向：迁移到 `docs/specs/shared/stable/`、吸收到 `s_system_constraints.md`、吸收到模块 `stable` 主文后删除，或因无法收口而阻塞本轮提升
   - 若模块 promote 后仍依赖共享附属展开文件，新的 `stable` 主文件必须同步写入 `shared_appendix_refs`
   - 必须先把 `_status.md` 更新到 `Candidate=no`
   - 之后才允许物理删除：
     - `docs/specs/candidate/c_{module}.md`
     - `docs/specs/_check_result/{module}.md`
     - `docs/specs/_verify_result/{module}.md`
     - `docs/specs/_plans/{module}.md`

提升时的版本与联动顺序固定如下：

1. `cand_promote:{module}` 生成的新 `stable`，其 `frontmatter.version` 默认必须等于当前 candidate 的 `frontmatter.version`。
2. 若本轮同时联动提升 `s_system_constraints.md`，必须在同一次 `cand_promote` 中完成全局约束正文吸收与版本创建或递增，再生成或更新模块 `stable`。
3. 上一条属于 `cand_promote` 内部的原子收口步骤；只要提升尚未结束，不把这一步视为候选链已中途失效，也不因此触发回退。
4. 但“原子收口”只描述目标动作必须在同一次 `cand_promote` 内完成，不表示仓库物理上不可能停在中间态；若命令中断、崩溃或被人工打断，执行者必须按本节后续恢复规则处理，不得把中间态继续当作有效 promote 完成态。
5. 若模块 `stable` 保留 `Global Constraint Alignment` 或等价章节：
   - 若本轮已联动创建或更新正式全局基线，则其中引用的 `system_constraints_stable_ref` 必须写成联动提升后的最新正式版本
   - 若本轮仍未形成正式全局基线，则该字段必须显式写 `none`
   - 若本轮提升后模块 `stable` 仍依赖共享附属展开文件，则 `shared_appendix_refs` 必须写成 stable 层版本引用；不得省略
6. `cand_promote` 的正式收口默认只覆盖模块 `stable`、必要时联动更新的 `s_system_constraints.md`，以及当前轮候选过程文件清理；本仓库当前不要求同步维护根目录 `VERSION` 或 Git Tag。
7. 若发现 `cand_promote` 中途停在以下任一状态，必须视为“提升未完成的恢复态”，不得宣称 promote 已完成：
   - `s_system_constraints.md` 已前移，但模块 `stable` 尚未更新到当前 candidate 版本
   - `s_system_constraints.md` 已前移，但模块 candidate 或候选侧过程文件仍残留旧轮次绑定
   - 模块 `stable` 已写入，但 candidate、过程文件或 `_status.md` 尚未完成收口清理
8. 进入“提升未完成的恢复态”后，当前最小可行动作固定回退为 `cand_check`：
   - 必须保证 `docs/specs/candidate/c_{module}.md` 仍存在；若已被误删，必须先按当前 `stable` 与当轮候选事实重建到可读状态
   - 必须同步把 `_status.md` 恢复到 candidate 语义：
     - `Candidate=yes`
     - `Active Layer=candidate`
     - `Next Command=cand_check`
     - `Stable` 按当前仓库中是否已经存在可读取的 `docs/specs/stable/s_{module}.md` 取值：若已存在则写 `yes`，否则写 `no`
   - 先重新确认 candidate 与最新正式全局基线是否仍兼容
   - 再重新生成候选侧放行凭证与计划/验证链路
   - 最后重新执行 `cand_promote`
9. 上一条的目的，是把所有 promote 中断态统一收敛回同一个稳定入口，避免执行者在 `cand_check / cand_verify / cand_promote` 之间自行猜测。

### 6.6A 命令与产物生命周期总表

| 对象 | 进入/创建者 | 继续依赖 | 失效或回退条件 | 回退者/清理者 |
|---|---|---|---|---|
| `docs/specs/stable/s_{module}.md` | `spec_init`、`cand_promote` | `stable_verify`、`spec_fork`、所有 `stable` 侧核对 | 正式层实现发生新的未核对改动时，不是 `stable` 失效，而是“已对齐声明”失效 | `stable drift reconciliation` 负责把 `_status.md` 回退到 `stable_verify` |
| `docs/specs/candidate/c_{module}.md` | `spec_new`、`spec_fork` | `cand_check`、`cand_plan`、`cand_impl`、`cand_verify`、`cand_promote` | candidate 原文变化；当前 `system_constraints_stable_ref` 与正式全局基线状态不一致 | 对应候选侧命令负责回退 `_status.md`；`cand_promote` 仅在 `_status.md` 已切到 `Candidate=no` 后负责清理旧轮次文件，或由下一轮 `spec_fork` 清理 |
| `docs/specs/_check_result/{module}.md` | `cand_check` 通过时创建 | `cand_plan`、`cand_impl`、`cand_verify` | candidate 变更；当前 candidate 绑定的 stable 基线状态失配；字段不完整；新一次 `cand_check` 未通过 | 候选侧自检负责回退到 `cand_check`；失败态 `cand_check`、`spec_fork`、`cand_promote` 负责删除 |
| `docs/specs/_plans/{module}.md` | `cand_plan` | `cand_impl`、`cand_verify` | candidate 原文变更；当前 candidate 绑定的 stable 基线状态失配；字段不完整 | 候选侧自检负责回退到 `cand_check`；`spec_fork`、`cand_promote` 负责删除 |
| `docs/specs/_verify_result/{module}.md` | `cand_verify` | `cand_promote` | 实现再次改动；candidate 变更；当前 candidate 绑定的 stable 基线状态失配；字段不完整 | 候选侧自检负责按原因回退到 `cand_verify`、`cand_impl` 或 `cand_check`；`spec_fork`、`cand_promote` 负责删除 |
| `docs/specs/_status.md` | 所有标准命令都会更新 | 下一步命令路由与默认落点判断 | 与实际文件、实现上下文或过程文件有效性失配 | 触发当前发现失配的命令负责修正，不能只报错不回写 |
| `docs/specs/system/stable/s_system_constraints.md` | 仓库正式基线，首次创建与联动更新者都为 `cand_promote` | 已存在正式全局基线后的相关模块命令 | 正文变更时表示正式基线升级 | `cand_promote` 负责联动更新 |
| `docs/specs/shared/candidate/c_shared_xxx.md` | 当前轮共享机制演进任务；若模块从 stable 派生 candidate 且继续依赖共享附属展开文件，则由同轮 `spec_fork` 前置派生 | 绑定了它的 candidate 模块命令 | 正文、版本、层级或模块绑定关系变化 | `shared_flow_reconcile` 负责统一回退引用模块到 `cand_check`；`cand_promote` 或后续共享演进负责清理 |
| `docs/specs/shared/stable/s_shared_xxx.md` | 模块 `cand_promote` 联动创建或更新 | 绑定了它的 stable 模块核对与后续 `spec_fork` | 正文、版本、层级或模块绑定关系变化 | `shared_flow_reconcile` 负责统一回退引用模块到 `stable_verify`；后续 `cand_promote` 或治理调整负责清理 |

### 6.7 `_check_result` 生命周期

`docs/specs/_check_result/{module}.md` 默认按当前 candidate 内容与当前正式全局基线状态管理：

1. `cand_check` 只有在结论为 `pass` 时，才创建或覆盖该文件。
2. 该文件一旦存在，表示“当前 candidate 已被放行进入后续候选链”。
3. `cand_check` 未通过时，不得写入失败态文件；若旧放行凭证已不再成立，必须删除。
4. 只要当前 candidate 内容发生变化，旧 `_check_result/{module}.md` 即视为失效。
5. 若 `s_system_constraints.md` 已存在，只要当前 candidate 中的 `system_constraints_stable_ref` 不等于当前 `s_system_constraints` 版本，旧 `_check_result/{module}.md` 即视为失效。
6. 若 `s_system_constraints.md` 尚不存在，只要当前 candidate 中的 `system_constraints_stable_ref` 不等于 `none`，旧 `_check_result/{module}.md` 即视为失效。
7. 当执行 `spec_fork` 开启新一轮 candidate 时，必须删除上一轮 `_check_result/{module}.md`。
8. 当执行 `cand_promote` 完成候选提升时，必须删除对应 `_check_result/{module}.md`。
9. 只要 `_check_result/{module}.md` 失效，且模块仍处于 `Candidate=yes`，就必须把 `_status.md` 中该模块的 `Next Command` 回退为 `cand_check`。

### 6.8 `_verify_result` 生命周期

`docs/specs/_verify_result/{module}.md` 默认按当前 candidate、当前实现上下文与当前正式全局基线状态管理：

1. 首次执行 `cand_verify` 时，若文件不存在，应创建结果文件。
2. 新一次 `cand_verify` 执行后，应覆盖旧结果，而不是持续追加历史噪音。
3. 只要当前 candidate 内容变化，旧 `_verify_result/{module}.md` 即视为失效。
4. 只要当前实现出现新的未核对改动，旧 `_verify_result/{module}.md` 即视为失效。
5. 若 `s_system_constraints.md` 已存在，只要当前 candidate 中的 `system_constraints_stable_ref` 不等于当前 `s_system_constraints` 版本，旧 `_verify_result/{module}.md` 即视为失效。
6. 若 `s_system_constraints.md` 尚不存在，只要当前 candidate 中的 `system_constraints_stable_ref` 不等于 `none`，旧 `_verify_result/{module}.md` 即视为失效。
7. 当执行 `spec_fork` 开启新一轮 candidate 时，必须删除上一轮 `_verify_result/{module}.md`。
8. 当执行 `cand_promote` 完成候选提升时，必须删除对应 `_verify_result/{module}.md`。
9. 只要 `_verify_result/{module}.md` 失效，不得继续把 `Next Command` 停在 `cand_promote`；必须根据失效原因回退到真正可执行的上游动作：
   - 若是实现再次发生未核对改动，默认回退为 `cand_verify`
   - 若验证已暴露实现偏差，默认回退为 `cand_impl`
   - 若 candidate 或当前正式全局基线已变化到需要重新收口，默认回退为 `cand_check`

### 6.9 `_plans` 生命周期

`docs/specs/_plans/{module}.md` 默认按候选轮次与当前正式全局基线状态管理：

1. 首次执行 `cand_plan` 时，若该轮计划文件不存在，应创建计划文件。
2. 计划文件至少包含：
   - `Implementation Tasks`
   - `spec_file_ref`
   - `spec_version_ref`
   - `spec_fingerprint`
   - `system_constraints_stable_file_ref`
   - `system_constraints_stable_version_ref`
   - `system_constraints_stable_fingerprint`
3. 在 `cand_plan` 阶段，必须补齐当前轮任务拆解，并写入当前 candidate 的绑定信息。
4. 在 `cand_impl` 阶段，计划文件持续回写当前轮进度、阻塞与验证重点，但不得改写绑定字段。
5. 只要当前 candidate 原文发生任何变化，现有 `_plans/{module}.md` 即视为失效。
6. 若 `s_system_constraints.md` 已存在，只要当前 candidate 中的 `system_constraints_stable_ref` 不等于当前 `s_system_constraints` 版本，现有 `_plans/{module}.md` 即视为失效。
7. 若 `s_system_constraints.md` 尚不存在，只要当前 candidate 中的 `system_constraints_stable_ref` 不等于 `none`，现有 `_plans/{module}.md` 即视为失效。
8. `_plans/{module}.md` 失效后，不得继续执行 `cand_impl` 或 `cand_verify`；必须先回到 `cand_check`，若 `cand_check` 通过，再重新执行 `cand_plan`。即使旧计划看起来仍能描述实施顺序，也不得在缺少当前有效 `_check_result/{module}.md` 的前提下继续下游命令。
9. 当执行 `spec_fork` 开启新一轮 candidate 时，若存在上一轮 `_plans/{module}.md`，必须先删除，避免沿用旧轮次计划。
10. 当执行 `cand_promote` 完成候选提升时，必须删除该轮 `_plans/{module}.md`。
11. `Candidate=no` 时，默认不得保留对应 `_plans/{module}.md`。
12. 只要 `_plans/{module}.md` 失效，且模块仍处于 `Candidate=yes`，就必须把 `_status.md` 中该模块的 `Next Command` 回退为 `cand_check`。

---

## 7. Read Guidance for AI Agents

AI Agent 在执行实现、审查或规划前，应按当前命令或任务需要读取对应对象；本仓库默认不定义 `docs/specs/system/stable/s_system_constraints.md` 与 `docs/specs/_status.md` 的统一总入口顺序。

### 7.0A 仓库入口索引文件

本仓库允许存在会变化的入口索引文件名，但不允许入口职责漂移到“无人负责”的状态。

固定规则如下：

1. 仓库入口索引文件的职责边界按职责识别，不按固定文件名识别；但默认正式入口集合不由执行者临时按职责扩张，而是以 `docs/agent_guidelines/entry_index_registry.md` 的登记结果为准。
2. 只要某个仓库内文件承担以下任一职责，就视为入口索引文件：
   - 给用户或执行者列出标准命令入口
   - 说明自然语言请求如何命中标准命令或治理流程
   - 把请求路由到 `docs/agent_guidelines/command_policy.md`、具体命令文件或 `spec_flow_review`
3. 第 2 条只回答“什么样的文件属于入口索引文件职责范畴”，不直接赋予该文件进入默认治理集合、默认审查范围或默认同步集合的资格。
4. 某个同职责文件只有先登记进 `docs/agent_guidelines/entry_index_registry.md`，才进入默认正式入口集合；在登记之前，它可以被识别为“同职责文件”，但不得自动进入 `spec_flow_review` 默认待审范围，也不得自动成为入口同步目标。
5. 入口索引文件不是治理真相源；它的职责是稳定路由，不是改写治理规则。
6. 入口索引文件可以有多个，也可以改名；但它们对标准命令集合、命中边界和治理流程入口的描述，必须与正式治理文件一致。
7. 若仓库内不存在任何已登记且承担上述职责的文件，则视为入口职责缺失，属于治理闭环问题。

固定读取规则如下：

1. 当任务需要判断当前正式技术基线、共享机制、全局默认做法或全局例外时，读取 `docs/specs/system/stable/s_system_constraints.md`。
2. 当任务需要判断目标模块当前 `Active Layer`、`Next Command`、默认落点或状态登记是否一致时，按本文件第 `3.7` 节的规则读取 `docs/specs/_status.md`。
3. 当目标模块当前版本层为 `stable`，且任务需要核对该模块正式行为时，读取 `docs/specs/stable/s_{module}.md`。
4. 当目标模块当前版本层为 `candidate`，且任务需要核对候选行为时，读取 `docs/specs/candidate/c_{module}.md`。
4A. 当目标模块当前层 `Global Constraint Alignment.shared_appendix_refs` 非空时，还必须读取这些共享附属展开文件。
5. 当目标模块当前版本层为 `candidate` 且存在 `stable`，并且任务需要了解当前正式基线时，补读 `docs/specs/stable/s_{module}.md`。
6. 当当前命令为 `cand_plan` 时，还必须读取 `docs/specs/_check_result/{module}.md` 放行凭证。
7. 当当前命令为 `cand_impl` 时，还必须同时读取 `docs/specs/_check_result/{module}.md` 与 `docs/specs/_plans/{module}.md`。
8. 当当前命令为 `cand_verify` 时，还必须同时读取 `docs/specs/_check_result/{module}.md` 与 `docs/specs/_plans/{module}.md`，确认当前 candidate 仍保有有效放行凭证，且本轮计划门禁仍成立。
9. 当当前命令为 `cand_promote` 时，还必须读取 `docs/specs/_verify_result/{module}.md`。
10. 当任务涉及实现、审查、升级、`cand_promote`，或修改 `docs/agent_guidelines/*.md`、`docs/agent_guidelines/commands/*.md`、`docs/agent_guidelines/entry_index_registry.md`、已登记入口索引文件、`docs/specs/` 中非 `candidate` 真相文件时，还必须读取 `docs/agent_guidelines/git_policy.md`，判断当前任务是否要求提交以及应按哪类提交收口。

补充说明：

1. `stable_verify` 没有独立过程文件输入。
2. `cand_check` 不读取 `_plans/{module}.md`。
3. `docs/specs/system/stable/s_system_constraints.md` 与 `docs/specs/_status.md` 解决的是不同问题：
   - 前者回答当前正式技术基线与共享机制
   - 后者回答模块状态与下一步命令
4. 模块命令读取 `system_constraints` 时，只读取当前正式基线，不读取独立候选文件；若正式基线尚不存在，则按“当前无正式全局基线”的空态继续执行对应规则。
5. `cand_plan`、`cand_impl`、`cand_verify` 与 `cand_promote` 不得只看 `_status.md` 就判断门禁已通过；还必须读取对应过程文件并校验其有效性。
6. `cand_plan` 读取 `_check_result/{module}.md` 时，除校验绑定信息外，还必须确认 `decision=pass`、`allow_next=true`、`next_command=cand_plan`；它消费的是放行凭证，不是失败审查记录。
7. `cand_impl` 与 `cand_verify` 读取 `_check_result/{module}.md` 时，至少还必须确认：
   - `decision=pass`
   - `allow_next=true`
   - `next_command=cand_plan`
   - 当前 candidate 原文仍与该放行凭证绑定的 `spec_version_ref`、`spec_fingerprint` 完全一致
8. `cand_verify` 读取 `_plans/{module}.md` 时，还必须确认：
   - 当前计划文件仍满足本文件第 `12.8` 节的完整绑定条件
   - 当前 candidate 原文仍与计划文件绑定的 `spec_version_ref`、`spec_fingerprint` 完全一致
9. `cand_promote` 读取 `_verify_result/{module}.md` 时，除校验绑定信息外，还必须确认 `decision=pass`、`allow_next=true`、`next_command=cand_promote`。
10. `_check_result/{module}.md` 的 `next_command` 固定表示“候选链入链入口”，不是模块当前阶段状态；模块当前最小可行动作仍只由 `_status.md` 的 `Next Command` 表达。
11. 上一条中的“已登记入口索引文件”，固定按 `docs/agent_guidelines/entry_index_registry.md` 的当前登记结果解释，不得由执行者按职责临时扩缩。
12. `git_policy.md` 解决的是“本轮改动如何进入版本历史”的问题；它不是命令名，但命中其触发条件时属于必读治理规则，不得用“命令文件里没写”作为跳过理由。

---

## 8. Preflight Consistency Rules

所有命令在真正执行前，都必须先完成一次状态一致性自检。

默认至少检查以下项目：

1. 除 `spec_init` 与 `spec_new` 外，`_status.md` 中必须存在目标模块登记。
2. `Active Layer` 与当前准备读取的 Spec 层一致。
3. `Stable=yes` 时，`docs/specs/stable/s_{module}.md` 必须存在。
4. `Candidate=yes` 时，`docs/specs/candidate/c_{module}.md` 必须存在。
5. 若主文件明确引用了当前层附属展开文件，则这些文件也必须存在于对应 appendix/ 或等价专用子目录中。
6. `Next Command=cand_impl` 时，`docs/specs/_check_result/{module}.md` 与 `docs/specs/_plans/{module}.md` 都必须存在且仍有效。
7. `Candidate=no` 时，不得残留 `docs/specs/_check_result/{module}.md`、`docs/specs/_verify_result/{module}.md`、`docs/specs/_plans/{module}.md`，以及该模块当前轮 candidate 附属展开文件。
8. `Next Command=cand_plan` 时，`docs/specs/_check_result/{module}.md` 必须存在且仍有效。
9. `Next Command=cand_verify` 时，`docs/specs/_check_result/{module}.md` 与 `docs/specs/_plans/{module}.md` 都必须存在且仍有效。
10. `Next Command=cand_promote` 时，`docs/specs/_verify_result/{module}.md` 必须存在且仍有效。
11. `Next Command=cand_plan` 或 `Next Command=cand_promote` 时，除过程文件仍有效外，还必须满足对应门禁文件的 `decision`、`allow_next` 与 `next_command` 放行条件。
12. `Next Command=cand_impl` 时，除过程文件仍有效外，还必须确认当前 candidate 仍保有有效 `_check_result/{module}.md`。
13. `Next Command=cand_verify` 时，除过程文件仍有效外，还必须确认当前 candidate 仍保有有效 `_check_result/{module}.md`，且当前 `_plans/{module}.md` 仍覆盖本轮 candidate。
14. `Active Layer=stable` 且正式层代码已发生新的未核对改动时，必须先把 `Next Command` 回退为 `stable_verify`，不得直接进入 `spec_fork`。
15. `Next Command` 必须与当前准备执行的命令一致；若不一致，默认不得越过，除非当前命令文件明示允许越过并要求先解释原因。
16. 模块 candidate 中必须存在 `system_constraints_stable_ref`，且格式正确。
17. 若 `s_system_constraints.md` 已存在，则当前 candidate 中的 `system_constraints_stable_ref` 必须等于当前 `s_system_constraints` 版本；否则候选侧过程文件默认失效，并统一回退为 `cand_check`。
18. 若 `s_system_constraints.md` 尚不存在，则当前 candidate 中的 `system_constraints_stable_ref` 必须等于 `none`；否则候选侧过程文件默认失效，并统一回退为 `cand_check`。
19. 只要过程文件已失效，执行者除了停止当前命令外，还必须把 `_status.md` 中该模块的 `Next Command` 回退到当前最小可行动作；不得只报“文件过期”而不修正状态。

### 8.1 Stable Drift Reconciliation Rules

当模块当前 `Active Layer=stable` 时，所有 `stable` 侧命令都必须先完成一次正式层漂移回退判断。

固定规则如下：

1. `stable_verify` 不保留可复用的过程文件凭据。
2. 因此，`stable drift reconciliation` 不再判断“最近一次 stable_verify 是否覆盖当前实现”。
3. 若仓库中存在该模块的正式层实现改动，必须直接把 `_status.md` 中该模块的 `Next Command` 纠正为 `stable_verify`。
4. 在完成上述状态纠正前，不得继续宣称“当前仍对齐 stable”，也不得继续执行 `spec_fork`。
5. 若当前 `stable` 模块层 `shared_appendix_refs` 绑定的 stable 共享附属展开文件版本、正文或绑定关系已变化，也必须直接把 `_status.md` 中该模块的 `Next Command` 纠正为 `stable_verify`。
6. 若 stable 共享附属展开文件的 `bound_modules` 与真实模块绑定集合不一致，这属于治理漂移；默认应由 `shared_flow_reconcile` 或 `spec_flow_review` 报出并明确回补责任，但不单独替代第 5 条中的状态回退判定。

### 8.2 Candidate Drift Reconciliation Rules

当模块当前 `Active Layer=candidate` 时，所有候选侧命令都必须把“过程文件是否仍覆盖当前 candidate、当前实现与当前正式全局基线状态”当成固定前置检查。

固定规则如下：

1. `_check_result/{module}.md` 失效时，必须把 `Next Command` 回退为 `cand_check`；不得继续执行 `cand_plan`、`cand_impl` 或 `cand_verify`。
2. `_plans/{module}.md` 失效时，必须把 `Next Command` 回退为 `cand_check`。
3. `_verify_result/{module}.md` 失效时，不得继续保留 `Next Command=cand_promote`：
   - 若是实现再次发生新的未核对改动，回退为 `cand_verify`
   - 若验证已明确实现与 candidate 有偏差，回退为 `cand_impl`
   - 若 candidate 或正式全局基线已变化到需要重新收口，回退为 `cand_check`
4. 若 `s_system_constraints.md` 已存在且版本前移，而模块 candidate 仍绑定旧版 `system_constraints_stable_ref`，则 `_check_result`、`_plans` 与 `_verify_result` 都必须按失效处理，并统一回退为 `cand_check`。
5. 若 `s_system_constraints.md` 尚不存在，但模块 candidate 没有显式绑定 `none`，则 `_check_result`、`_plans` 与 `_verify_result` 都必须按失效处理，并统一回退为 `cand_check`。
5A. 若 candidate 层 `shared_appendix_refs` 指向的共享附属展开文件版本、正文、层级或绑定关系已变化，且当前过程文件中的 `shared_appendix_snapshot` 与当前 Shared Appendix 真相重新生成的规范化快照不一致，则 `_check_result`、`_plans` 与 `_verify_result` 都必须按失效处理，并统一回退为 `cand_check`。
6. 只要 candidate 原文变化导致旧 `_check_result/{module}.md` 失效，即使现有 `_plans/{module}.md` 看起来仍能描述实施顺序，也不得继续沿用该计划直接推进下游命令。
7. `Next Command` 必须始终指向“当前最小可行动作”，不能指向一个已知会再次失败的下游命令。
8. 若 `cand_promote` 已经联动前移 `s_system_constraints.md`，但模块 `stable`、candidate 清理或 `_status.md` 仍未收口完成，该中断态也按候选侧失配处理，并统一回退为 `cand_check`；不得继续保留 `Next Command=cand_promote`。
9. 上一条中的回退，不得只修改 `Next Command`；还必须把 `_status.md` 一并恢复到 `Candidate=yes`、`Active Layer=candidate` 的 candidate 语义。

---

## 9. Candidate Sufficiency Rules

一个 candidate Spec 是否足以进入实施计划，至少要满足以下条件：

1. 模块目标与边界已清楚
2. 核心术语已清楚
3. 数据结构与协议已清楚
4. 状态机与主流程已清楚
5. 边界条件与错误处理已清楚
6. 验收标准已清楚
7. 涉及技术选型、共享基础设施、跨模块复用、全局例外申请或全局约束提案时，已写清 `Global Constraint Alignment`
8. 已写清当前对齐的 `system_constraints_stable_ref`；若正式全局基线尚不存在，则显式写 `none`
9. 若当前层依赖共享附属展开文件，已写清 `shared_appendix_refs`
10. 不再依赖实现阶段临时发明关键行为规则

默认判断规则：

1. 只有背景和愿景，没有结构与协议：不得进入 `cand_plan`
2. 主流程仍要求实现者自行发明：不得进入 `cand_plan`
3. 错误语义和边界条件缺失到会影响实现判断：不得进入 `cand_plan`
4. 输入输出语义缺失到会导致不同实现者写出不同对外结果：不得进入 `cand_plan`
5. 验收口径缺失到无法判断实现是否对齐 candidate：不得进入 `cand_plan`
6. 涉及技术选型、共享基础设施或跨模块复用，但未说明如何对齐 `system_constraints`：不得进入 `cand_plan`
7. 未写清当前 `system_constraints_stable_ref`，或在“无正式全局基线”场景下未显式写 `none`：不得进入 `cand_plan`
8. 当前层行为明确依赖共享附属展开文件，但未写清 `shared_appendix_refs`：不得进入 `cand_plan`

---

## 10. Behavior Change Rule

以下情况都属于行为变化：

1. 对外接口语义变化
2. 状态机或主流程变化
3. 错误语义或重试语义变化
4. 存储结构变化
5. 跨模块边界变化
6. 关键对象、关键接缝或关键控制流变化
7. 全局技术基线、共享机制、默认选型或全局例外变化

行为变化的执行要求如下：

1. 不得代码先行
2. 历史模块已有正式版本时：
   - 先更新 `candidate`
   - 再执行 `cand_check`
   - 再生成或更新 `_plans/{module}.md`
   - 最后推进代码
3. 新模块首版时：
   - 先建立 `candidate`
   - 再执行 `cand_check`
   - 再生成 `plan`
   - 最后推进代码

---

## 11. Status Consumption Rule

`_status.md` 的对象职责、字段语义、读取时机与维护责任以上游定义为准：

1. 以本文件第 `3.7` 节为主，不再去 `_status.md` 文件正文内寻找规则。
2. 命令执行时读取 `_status.md`，读取的是状态登记结果，不是治理规则正文。
3. 若命令文件、审查流程或执行者理解与第 `3.7` 节不一致，以第 `3.7` 节为准，并应在当前任务内修正文档漂移。

---

## 12. Verification Evidence Rules

`cand_verify` 与 `stable_verify` 默认都必须输出结构化验证证据，而不是只给主观总结。

### 12.1 Fixed Snapshot Fields

`Check Result Snapshot` 至少包含以下固定字段：

1. `module`
2. `gate`
3. `decision`
4. `allow_next`
5. `next_command`
6. `blocking_summary`
7. `coverage_summary`
8. `prompt_adequacy_review_required`
9. `prompt_adequacy_decision`
10. `prompt_adequacy_summary`
11. `spec_layer_ref`
12. `spec_file_ref`
13. `spec_version_ref`
14. `spec_fingerprint`
15. `system_constraints_stable_file_ref`
16. `system_constraints_stable_version_ref`
17. `system_constraints_stable_fingerprint`
18. `shared_appendix_snapshot`

`Verify Result Snapshot` 至少包含以下固定字段：

1. `gate`
2. `decision`
3. `allow_next`
4. `next_command`
5. `blocking_summary`
6. `coverage_summary`
7. `spec_layer_ref`
8. `spec_file_ref`
9. `spec_version_ref`
10. `spec_fingerprint`
11. `verification_scope_ref`
12. `system_constraints_stable_file_ref`
13. `system_constraints_stable_version_ref`
14. `system_constraints_stable_fingerprint`
15. `shared_appendix_snapshot`

`Plan Binding Snapshot` 至少包含以下固定字段：

1. `spec_file_ref`
2. `spec_version_ref`
3. `spec_fingerprint`
4. `system_constraints_stable_file_ref`
5. `system_constraints_stable_version_ref`
6. `system_constraints_stable_fingerprint`
7. `shared_appendix_snapshot`

`shared_appendix_snapshot` 的固定语义如下：

1. 若当前模块当前层 `shared_appendix_refs=none`，固定写 `none`。
2. 若当前模块当前层存在共享附属展开文件绑定，必须写成按绑定顺序展开的规范化快照字符串。
3. 每个绑定项固定格式为：`<shared_ref>#<shared_fingerprint>`。
4. 多个绑定项之间固定使用 `|` 连接。
5. 其中：
   - `<shared_ref>` 使用模块当前层 `shared_appendix_refs` 中的版本引用原文，例如 `c_shared_xxx@1.2.0`
   - `<shared_fingerprint>` 使用对应 Shared Appendix 当前完整原文在首尾空白裁剪后的指纹
6. 该字段回答的是“当前过程文件绑定的是哪组 Shared Appendix 及其正文状态”，不是声明性 `bound_modules` 清单。
7. 消费该字段时，必须用当前模块当前层 `shared_appendix_refs` 与当前 Shared Appendix 实际正文重新生成同口径快照，再做精确比较；不得只比较版本引用，不得跳过正文指纹。

### 12.2 `prompt_adequacy_summary` 语义契约

`prompt_adequacy_summary` 虽然仍是单字段文本，但其正文语义必须满足固定契约。

最小要求如下：

1. 明确写出当前模块是否命中 Prompt Adequacy Review。
2. 明确写出 `基础充分性审查` 的结论。
3. 明确写出 `结构化输出审查` 是否适用；若适用，必须给出结论。
4. 若命中结构化输出审查，必须说明 `Few-shot Example Requirement` 是否被触发，以及是否满足。
5. 明确写出 `排序审查` 的结论。
6. 明确写出当前阻塞项；若无，则写 `none`。

补充说明：

1. `基础充分性审查`、`结构化输出审查`、`排序审查` 的定义以 `prompt_guidelines.md` 为准。
2. 只写“Prompt Adequacy Review passed”或等价空泛总结，默认视为不满足本节契约。

### 12.2A `cand_check` 非通过输出契约

`cand_check` 若结论为 `blocked` 或 `fix_required`，其终端审查输出必须满足固定解释契约；这里规范的是审查输出本身，不是失败态 `_check_result` 文件，因为失败态结果默认不落库。

最小要求如下：

1. 必须输出结构化 `findings`，不得只写抽象结论或一句话建议。
2. 每条 finding 至少必须写清：
   - 当前约束面的背景是什么
   - candidate 具体缺了什么、冲突在哪里，或哪里仍依赖实现阶段发明
   - 该缺口为什么会阻塞 `cand_plan` 或导致后续实现分叉
   - 当前最佳修复建议是什么
   - 为什么该建议是最佳修复路径，而不是局部补丁
   - 当前问题是否阻塞
3. 若 `Prompt Adequacy Review=n/a`，仍必须说明为什么当前模块不命中 Prompt 触发条件，以及哪些 Prompt 子审查项不适用。
4. 所有非通过建议都必须面向恢复 candidate 的正确性与完整性，不得写成“为了通过本轮审查先补一句说明”这类补丁式修法。

补充说明：

1. `cand_check` 的具体 finding 字段、允许分类与 Markdown 渲染格式，以 `docs/agent_guidelines/commands/cand_check.md` 为准。
2. 本节的作用，是确保上游总策略承认“`cand_check` 非通过输出必须结构化且禁止补丁式建议”这一固定语义，避免命令文件与总策略漂移。

### 12.3 验证偏差分级规则

`stable_verify` 与 `cand_verify` 的偏差分级固定统一为 `P1 / P2 / P3`。

定义如下：

1. `P1`
   - 已影响或可能直接影响关键协议、主流程、错误语义、关键验收点或正式全局基线对齐。
   - 会改变 `是否仍对齐 stable`、`是否允许 cand_promote` 这类门禁结论。
2. `P2`
   - 不直接推翻当前门禁结论，但已经影响非关键验收点的稳定判断，或要求后续补证、补实现、补收口。
   - 若继续放任，容易在后续轮次演化成 `P1`。
3. `P3`
   - 不改变当前门禁结论，也不改变实现语义。
   - 主要影响验证证据的可读性、复审效率或低风险尾项记录。

固定规则：

1. 关键协议、主流程、错误语义、关键验收点、`system_constraints` 对齐问题，不得降成 `P3`。
2. 任何会直接改变 `allow_next` 或 `next_command` 的问题，至少是 `P1`。
3. `P3` 只能用于不改变当前放行结论的低风险问题，不得被拿来包装真实阻塞项。

### 12.4 非阻塞验证证据降级规则

`partial` 或 `not_checked` 不是天然可放行项；只有同时满足固定条件时，才允许被判为“非阻塞”。

#### 12.4.1 `partial` 可降级条件

仅当以下条件同时满足时，`partial` 才允许降为非阻塞：

1. 对应 Spec 条目已拿到足以支撑当前门禁结论的主要证据。
2. 剩余未覆盖部分不涉及关键协议、主流程、错误语义、关键验收点或正式全局基线对齐。
3. 当前没有任何反向证据表明该条目实际不满足 Spec。
4. 验证输出已明确写清：
   - 为什么仍是 `partial`
   - 为什么该残余缺口不影响当前门禁
   - 当前残余风险是什么
   - 后续是否还需要补证；若不需要，也要写清原因

#### 12.4.2 `not_checked` 不可降级条件

出现以下任一情况时，`not_checked` 一律不得视为非阻塞：

1. 对应关键协议。
2. 对应主流程或关键步骤顺序。
3. 对应错误处理、跳过、中断、重试等失败语义。
4. 对应关键验收点。
5. 对应 `system_constraints` 对齐或其它会改变当前门禁判断的条目。

除上述列举外，`not_checked` 在当前仓库规则下默认仍按阻塞处理；本仓库当前不把 `not_checked` 作为可稳定放行的非阻塞类型，除非上游规则未来显式放开。

#### 12.4.3 门禁结论约束

1. `stable_verify` 只有在不存在 `fail`，且所有关键条目都已被完整覆盖时，才允许宣称“当前仍对齐 stable”。
2. `cand_verify` 只有在不存在 `fail`，且所有会影响 `cand_promote` 的关键条目都已被完整覆盖时，才允许写出 `decision=pass`、`allow_next=true`、`next_command=cand_promote`。
3. 被降为非阻塞的条目，仍必须保留在验证证据矩阵和偏差清单中，不得当作“没有问题”。

### 12.5 `spec_version_ref` 固定格式

1. `stable` 固定为：`s_{module}@<frontmatter.version>`
2. `candidate` 固定为：`c_{module}@<frontmatter.version>`
3. `system_constraints` 正式层固定为：`s_system_constraints@<frontmatter.version>`

### 12.6 `_check_result` 有效性判定

`docs/specs/_check_result/{module}.md` 只有在以下条件同时满足时，才可视为当前有效：

1. `gate=cand_check`
2. `module` 等于目标正式模块名
3. `decision=pass`
4. `allow_next=true`
5. `spec_layer_ref=candidate`
6. `spec_file_ref` 等于当前 candidate 文件路径
7. `spec_version_ref` 等于当前 candidate 的版本引用
8. `spec_fingerprint` 等于当前 candidate 的指纹
9. `next_command=cand_plan`
10. 当前正式全局约束存在时，`system_constraints_stable_file_ref` 等于当前正式全局约束文件路径；若不存在，则该字段等于 `none`
11. 当前正式全局约束存在时，`system_constraints_stable_version_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
12. 当前正式全局约束存在时，`system_constraints_stable_fingerprint` 等于当前正式全局约束指纹；若不存在，则该字段等于 `none`
13. 当前正式全局约束存在时，当前 candidate 中的 `system_constraints_stable_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
14. 当前 candidate 当前层 `shared_appendix_refs=none` 时，`shared_appendix_snapshot=none`
15. 当前 candidate 当前层 `shared_appendix_refs` 非空时，`shared_appendix_snapshot` 必须精确等于按本节第 `12.1` 节规则从当前绑定 Shared Appendix 重新生成的规范化快照
16. 若模块命中 Prompt Adequacy Review，则必须满足：
   - `prompt_adequacy_review_required=true`
   - `prompt_adequacy_decision=pass`
   - `prompt_adequacy_summary` 满足本节第 `12.2` 节的语义契约
17. 若模块未命中 Prompt Adequacy Review，则必须满足：
   - `prompt_adequacy_review_required=false`
   - `prompt_adequacy_decision=n/a`
   - `prompt_adequacy_summary` 满足本节第 `12.2` 节的语义契约

任一不满足，默认视为过期结果。

### 12.7 `_verify_result` 有效性判定

`docs/specs/_verify_result/{module}.md` 只有在以下条件同时满足时，才可视为当前有效：

1. `gate=cand_verify`
2. `spec_layer_ref=candidate`
3. `spec_file_ref` 等于当前 candidate 文件路径
4. `spec_version_ref` 等于当前 candidate 的版本引用
5. `spec_fingerprint` 等于当前 candidate 的指纹
6. `next_command` 等于当前命令期望值
7. `verification_scope_ref` 仍覆盖当前实现上下文
8. 当前正式全局约束存在时，`system_constraints_stable_file_ref` 等于当前正式全局约束文件路径；若不存在，则该字段等于 `none`
9. 当前正式全局约束存在时，`system_constraints_stable_version_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
10. 当前正式全局约束存在时，`system_constraints_stable_fingerprint` 等于当前正式全局约束指纹；若不存在，则该字段等于 `none`
11. 当前正式全局约束存在时，当前 candidate 中的 `system_constraints_stable_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
12. 当前 candidate 当前层 `shared_appendix_refs=none` 时，`shared_appendix_snapshot=none`
13. 当前 candidate 当前层 `shared_appendix_refs` 非空时，`shared_appendix_snapshot` 必须精确等于按本节第 `12.1` 节规则从当前绑定 Shared Appendix 重新生成的规范化快照

任一不满足，默认视为过期结果。

### 12.8 `_plans` 有效性判定

`docs/specs/_plans/{module}.md` 只有在以下条件同时满足时，才可视为当前有效：

1. `spec_file_ref` 等于当前 candidate 文件路径
2. `spec_version_ref` 等于当前 candidate 的版本引用
3. `spec_fingerprint` 等于当前 candidate 的指纹
4. 当前正式全局约束存在时，`system_constraints_stable_file_ref` 等于当前正式全局约束文件路径；若不存在，则该字段等于 `none`
5. 当前正式全局约束存在时，`system_constraints_stable_version_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
6. 当前正式全局约束存在时，`system_constraints_stable_fingerprint` 等于当前正式全局约束指纹；若不存在，则该字段等于 `none`
7. 当前正式全局约束存在时，当前 candidate 中的 `system_constraints_stable_ref` 等于当前正式全局约束版本引用；若不存在，则该字段等于 `none`
8. 当前 candidate 当前层 `shared_appendix_refs=none` 时，`shared_appendix_snapshot=none`
9. 当前 candidate 当前层 `shared_appendix_refs` 非空时，`shared_appendix_snapshot` 必须精确等于按本节第 `12.1` 节规则从当前绑定 Shared Appendix 重新生成的规范化快照

任一不满足，默认视为过期计划。

---

## 13. Prohibitions

以下行为一律禁止：

1. 用章节完整替代真实收口
2. candidate 未闭环就进入 `cand_plan`
3. 没有 `plan` 就进入 `cand_impl`
4. 把 `_plans/{module}.md` 当作正式真相源
5. 在实现阶段临时补发明关键主流程、关键状态归属或关键接口语义
6. 把 `module_xxx` 与具体文件路径混为一体
7. 在状态漂移或文件漂移时继续执行后续命令
8. 在缺少验证证据时推进到提升
9. 把“当前看起来没问题”当作验证通过依据
10. 继续维护独立的 `c_system_constraints.md` 候选层
