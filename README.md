# specFlow

`specFlow` 是一套以 Spec 为真相源的开发治理框架。

它解决的不是“怎么写一份文档”，而是“怎么让人和 AI 围绕同一份真相推进设计、计划、实现、验证和提升”。

## 目录

- `framework/`
  - 框架规则正文源文件。
- `templates/root/`
  - 安装到目标仓库根目录和固定路径的模板文件。
- `tooling/`
  - 初始化、升级、检查工具。

## 快速开始

在引入 `specflow/` 后，先进入仓库根目录运行：

```bash
./specflow/tooling/init.sh
```

Windows PowerShell:

```powershell
.\specflow\tooling\init.ps1
```

初始化完成后，目标仓库会得到：

- `AGENTS.md`
- `GEMINI.md`
- `.githooks/pre-commit`
- `docs/agent_guidelines/**`
- `docs/specs/**`

后续可使用：

- `./specflow/tooling/doctor.sh`
- `./specflow/tooling/upgrade.sh`

Windows PowerShell:

- `.\specflow\tooling\doctor.ps1`
- `.\specflow\tooling\upgrade.ps1`
