---
version: 0.1.0
---

# System Constraints Spec

> 版本说明：本文档描述当前正式生效的全局系统约束。它不是普通模块 Spec，不进入 `docs/specs/_status.md`，默认只允许在模块 `cand_promote` 时作为联动副产品被更新。只要正文被实际改动，必须同步递增 `frontmatter.version`；只读不改时不得变更版本号。

## 1. Context & Scope

本文件只负责回答以下问题：

1. 当前项目正式承认的技术栈基线是什么。
2. 哪些共享机制已经存在，后续模块应优先复用。
3. 遇到某类工程问题时，默认优先选择什么方案。
4. 哪些做法被全局禁止，哪些例外必须显式登记。

本文件不负责：

1. 描述单个模块的内部状态机。
2. 约束单个模块的函数拆分或代码风格细节。
3. 承载模块 candidate 阶段的全局提案草稿。

## 2. Version Semantics

本文件版本号采用 `MAJOR.MINOR.PATCH`：

1. `MAJOR`
   - 全局约束出现不兼容变化
2. `MINOR`
   - 新增全局默认规则、共享机制、兼容性扩展
3. `PATCH`
   - 只修正文案、澄清歧义且不改变正式约束语义

模块 candidate 在 `Global Constraint Alignment` 中引用本文件时，必须使用固定格式：

1. `system_constraints_stable_ref: s_system_constraints@<frontmatter.version>`
2. 若模块当前层还显式绑定了共享附属展开文件，应在同一章节额外登记 `shared_appendix_refs`；该字段不替代 `system_constraints_stable_ref`

## 3. Tech Stack Baseline

> 按目标项目实际情况填写正式基线；若模块需要提出新的全局约束变化，应写在模块自己的 candidate 中，而不是创建独立 system candidate 文件。

1. 主语言：
2. 主框架 / 运行时：
3. 主存储：
4. 缓存：
5. 队列 / 异步任务：
6. 测试体系：

## 4. Shared Mechanisms

> 记录当前项目已承认的共享基础设施或共享机制。若某类机制尚未正式承认，不应在此假装已经存在。

1. 配置管理：
2. 日志 / 审计：
3. 认证 / 授权：
4. 缓存复用：
5. 调度 / 后台任务：
6. 事件或消息机制：
7. ID / 唯一标识生成：
8. 重试 / 降级策略：

## 5. Default Selection Rules

> 只写“默认优先怎么选”，不要把所有历史讨论都堆进来。

1. 当模块需要持久化业务数据时，默认优先：
2. 当模块需要短期共享状态或缓存时，默认优先：
3. 当模块需要后台异步处理时，默认优先：
4. 当模块需要共享日志、审计或追踪时，默认优先：
5. 当模块需要跨模块复用已有机制时，默认要求：
6. 当模块当前需要复用尚未沉淀为正式全局基线的共享机制正文时，默认应通过 Shared Appendix 绑定，而不是挂在某个模块 appendix 下双写

## 6. Global Prohibitions / Exceptions

### 6.1 Prohibitions

1. 禁止在同一类核心能力上并行引入两套互相冲突的主方案，除非例外已登记。
2. 禁止模块在未说明原因的情况下绕过已正式承认的共享机制，私自重造同类基础设施。
3. 禁止把“暂时方便”的实现选择伪装成正式工程基线。

### 6.2 Exceptions

若某模块必须偏离本文件，至少要在该模块 candidate 的 `Global Constraint Alignment` 中明确：

1. 例外点是什么。
2. 为什么现有正式约束不适用。
3. 例外影响范围是什么。
4. 例外是临时过渡还是准备推动全局升级。
5. 若同时偏离了某份 Shared Appendix，也必须一并写清该共享对象的引用与例外关系。
