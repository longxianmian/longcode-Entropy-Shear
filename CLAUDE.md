# CLAUDE.md — Claude Code 入口

> 本文件给 Claude Code 阅读。其他 AI 编程代理请读 [`AGENTS.md`](AGENTS.md)。

## 在做任何动作前，先读这两份文件

1. [`LONGMA_SOFT_GUARD.md`](LONGMA_SOFT_GUARD.md) — 人类可读总章（详细约束）。
2. [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json) — 机器可读锚点；与 MD 冲突时**以 JSON 为准**。

## 当前轮唯一目标

> 在现有 Entropy Shear Core 之上准备开发熵剪旗舰版（Flagship），但**本轮只建立开发期软约束包 P0，不开发任何核心代码**。

## 本轮允许写入的文件（且仅这些）

- `LONGMA_TASK_ANCHOR.json`
- `LONGMA_SOFT_GUARD.md`
- `AGENTS.md`
- `CLAUDE.md`
- `.claude/settings.json`

其它一切只允许 Read / ls / grep。

## Claude Code 专属操作守则

1. **写动作前先读真实文件**：用 `Read` 读取目标文件（若已存在），用 `Bash: ls` / `Bash: find` 验证父目录真实存在。绝不依赖记忆中的目录结构。
2. **引用接口前先 grep**：在引用 `internal/` 下的函数签名、字段名、错误码、`schemas/` 中的字段、`openapi.yaml` 中的端点之前，用 `Bash: grep -rn` 在真实代码中验证。不得引用未读过的代码或文档；**不得参照训练数据中的"熵剪 / 龙码常数"印象**，一切以当前仓库文件为准。
3. **越界即 HOLD**：写动作目标若落在 [`LONGMA_SOFT_GUARD.md`](LONGMA_SOFT_GUARD.md) 第 3 节的禁止路径内，立即停下并把决策权交还给用户。`.claude/settings.json` 已对常见禁止路径配置了 `permissions.deny` 兜底，但这**只是兜底**，不是免责。
4. **不要自行 git 提交**：不自行 `git commit` / `git push` / `git tag` / 任何 release 动作；本轮所有产物都应停留在 working tree，由用户决定是否落库。
5. **完成必须给证据**：交付时必须输出 changed files、`git diff --stat`、执行过的校验命令、边界检查、风险与 HOLD 项。少一项视为未完成。
6. **进入旗舰版 P0 开发前必须先换锚点**：进入熵剪旗舰版 P0 开发前，必须先更新 [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json)，把授权半径切换为具体开发任务（替换 `target_constant.single_goal`、`scope_id`、`authorization_radius.allowed_files`、`current_round_deliverables_only`、`forbidden_boundaries` 等字段，并同步更新 `.claude/settings.json` 的 `permissions.deny`）。**不得沿用本轮"只建立软约束包"的授权半径**去做开发；沿用即触发 AP1（目标漂移）+ AP2（行动越权），立即 HOLD。锚点未完成切换前，仅可 Read / ls / grep，不得对 Core / Pro / Flagship 任何代码进行 Write / Edit / MultiEdit。

## 五种反模式（速查）

| # | 反模式 | 触发即 HOLD |
|---|---|---|
| AP1 | 目标漂移 | 偏离上文唯一目标 |
| AP2 | 行动越权 | 写禁止路径 / 自行 `git commit` / `push` / `tag` / 改外部依赖 |
| AP3 | 路径漂移 | 对真实不存在的目录或文件操作 |
| AP4 | 幻觉判断 | 引用未读过的代码、字段、文档 |
| AP5 | 嘴炮完成 | 完成无证据 |

## 边界声明

本软约束包**不是** Entropy Shear Core 的产品功能。它**不进入** `policies/`、**不修改** `policies/manifest.json`、**不代表** Core 实现 Longma Constant。Longma Constant 仍然属于 Pro / Flagship 路线，详见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md) 与 [`README.md`](README.md) 的 Edition Boundary 段落。
