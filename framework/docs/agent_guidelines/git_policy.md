# Git 提交流程与版本策略（Git Policy）

## 1. 基本原则

1. `stable` 的变化表示正式契约变化。
2. `candidate` 的变化表示候选推进，不等于正式生效。
3. 因此，候选推进提交和提升提交必须分开定义。
4. `docs/specs/` 中除 `candidate` 层主文件与其附属展开文件外的 Spec 文件，都是仓库中的行为真相源；其修改默认应在当前任务内纳入 git 历史。
5. `candidate` 层主文件与其附属展开文件属于候选草案层；若本次只修改这些层文件，默认不执行 `git commit`，除非用户明确要求，或命中要求提交的命令流程。
6. `docs/agent_guidelines/*.md` 与 `docs/agent_guidelines/commands/*.md` 都属于仓库规则的一部分；其修改默认也应在当前任务内执行 `git commit`。
7. 已登记入口索引文件的修改也属于治理规则修改；其修改默认应在当前任务内执行 `git commit`，并在提交前完成入口文件同步。
8. 当 `Active Layer=stable` 且代码发生新的正式层实现改动时，默认必须把对应模块的 `Next Command` 回退为 `stable_verify`。
9. `docs/specs/system/stable/s_system_constraints.md` 默认视为模块 `cand_promote` 的正式副产品。

---

## 2. 候选推进提交

适用场景：

1. 新模块首版推进
2. 已有模块的候选升级推进

规则：

1. 允许使用 `feat:` 提交 `candidate + code + plan` 的联动修改。
2. 此类提交不强制要求同时修改 `stable`。
3. 但提交对应的行为目标必须能在 `candidate` 中找到来源。
4. 若只是把代码拉回到当前对齐层，可以使用 `fix:`。
5. 若只是结构调整且不改变当前对齐层定义的行为，可以使用 `refactor:`。

---

## 3. 提升提交

适用场景：

1. 执行 `cand_promote:{module}`

规则：

1. 该提交必须更新或创建对应 `stable`。
2. 该提交必须删除该轮候选的 `docs/specs/candidate/c_{module}.md`，以及该模块在 `docs/specs/candidate/appendix/` 或等价专用子目录中的本轮附属展开文件；若本轮联动处理了 Shared Appendix，还必须同步处理对应的 `docs/specs/shared/candidate/*.md` 或 `docs/specs/shared/stable/*.md` 去向。
3. 若存在该轮 `_check_result/{module}.md`、`_verify_result/{module}.md`、`_plans/{module}.md`，也必须同步删除。
4. 若该模块 candidate 中 `promotion_to_system_stable=with_module`，则同一提交还必须更新 `docs/specs/system/stable/s_system_constraints.md`。

---

## 4. 语义化版本规则

版本号采用 `MAJOR.MINOR.PATCH`。

### 4.1 模块 `stable`

1. `MAJOR`
   - 正式契约出现不兼容变化
2. `MINOR`
   - 正式契约新增能力或兼容性行为变化
3. `PATCH`
   - 只修正实现或对齐当前对齐层

### 4.2 `s_system_constraints.md`

1. `MAJOR`
   - 全局约束出现不兼容变化
2. `MINOR`
   - 新增全局默认规则、共享机制、兼容性扩展
3. `PATCH`
   - 只修正文案、澄清歧义且不改变正式约束语义

说明：

1. `candidate` 的内容可以频繁变化。
2. 但只有当它被提升为新的 `stable` 时，才进入正式版本语义。

---

## 5. 提升提交的收口范围

规则：

1. `cand_promote` 的收口范围默认只包括本轮模块 `stable`、必要时联动更新的 `s_system_constraints.md`、本轮联动处理的 Shared Appendix，以及本轮 candidate 主文件、candidate 附属展开文件和候选侧过程文件清理。
2. 本仓库当前不要求在 `cand_promote` 时维护根目录 `VERSION`。
3. 本仓库当前不要求在 `cand_promote` 时创建 Git Tag。

---

## 6. 文档修改与提交要求

### 6.1 `docs/specs/*.md`

若本次任务只修改 `docs/specs/*.md`：

1. 若修改的是 `docs/specs/candidate/c_{module}.md`、`docs/specs/candidate/appendix/`、等价专用子目录中的 candidate 层附属展开文件，或 `docs/specs/shared/candidate/*.md`，默认不执行 `git commit`；除非用户明确要求，或该修改属于要求提交的命令流程。
2. 若修改的是 `docs/specs/stable/*.md`、`docs/specs/stable/appendix/*.md`、等价专用子目录中的 stable 层附属展开文件、`docs/specs/shared/stable/*.md`、`docs/specs/system/stable/*.md`、`docs/specs/_status.md`、`docs/specs/_check_result/*.md`、`docs/specs/_verify_result/*.md` 或 `docs/specs/_plans/*.md`，默认应在当前任务内执行 `git commit`。
3. 若修改的是 `stable`，必须按正式契约变化处理；若命中 `cand_promote`，按提升提交规则处理。
4. 若 `candidate` 修改与对应代码实现、计划文件或提升提交属于同一条命令流程，可以随该流程一并提交。

### 6.2 `docs/agent_guidelines/*.md` 与 `docs/agent_guidelines/commands/*.md`

若本次任务只修改 `docs/agent_guidelines/*.md` 或 `docs/agent_guidelines/commands/*.md`：

1. 默认应纳入 git 历史，因为它们属于仓库规则的一部分。
2. 默认应在当前任务内执行 `git commit`，不再按批次归并。

### 6.3 已登记入口索引文件

若本次任务修改的是 `docs/agent_guidelines/entry_index_registry.md` 中登记的入口索引文件，例如 `AGENTS.md`、`GEMINI.md`：

1. 默认应纳入 git 历史，因为它们会直接影响命令列举、命中说明或治理流程路由。
2. 默认应在当前任务内执行 `git commit`。
3. 提交前必须先完成入口文件同步；若多个已登记入口文件同时被修改且内容仍不一致，必须先显式指定同步源，再继续提交。
