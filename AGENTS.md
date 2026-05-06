# AGENTS.md — AI 编程代理入口

> 本文件给所有非 Claude Code 的 AI 编程代理（包括 Codex 等）阅读。Claude Code 请同时阅读 [`CLAUDE.md`](CLAUDE.md)。

## 在做任何动作前，先读这两份文件

1. [`LONGMA_SOFT_GUARD.md`](LONGMA_SOFT_GUARD.md) — 人类可读总章。
2. [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json) — 机器可读锚点；与上文冲突时**以本锚点为准**。

## 当前轮唯一目标

> 在现有 Entropy Shear Core 之上准备开发熵剪旗舰版（Flagship），但**本轮只建立开发期软约束包 P0，不开发任何核心代码**。

本轮允许写入的文件**只**有：

- `LONGMA_TASK_ANCHOR.json`
- `LONGMA_SOFT_GUARD.md`
- `AGENTS.md`
- `CLAUDE.md`
- `.claude/settings.json`

其它一切只允许 read / list / grep，**不允许写**。

## 五条最低红线

| # | 反模式 | 触发即 HOLD |
|---|---|---|
| AP1 | 目标漂移 | 偏离上文唯一目标 |
| AP2 | 行动越权 | 写禁止路径 / 自行 `git commit` / `push` / `tag` / 改外部依赖 |
| AP3 | 路径漂移 | 对真实不存在的目录或文件操作（写入前必须 `ls` / Read 验证） |
| AP4 | 幻觉判断 | 引用未经 Read / grep 验证的代码、字段、文档 |
| AP5 | 嘴炮完成 | 完成无证据 |

## 完成必须给出

1. changed files（清单 + 用途）
2. `git diff --stat`（含 untracked 列举）
3. 执行过的校验命令及结果
4. 边界检查（是否触碰禁止路径、是否改 `policies/manifest.json`、是否动 Core 引擎）
5. 风险与 HOLD 项

## 边界声明

本软约束包**不是** Entropy Shear Core 的产品功能。它**不进入** `policies/`、**不修改** `policies/manifest.json`、**不代表** Core 实现 Longma Constant。Longma Constant 仍然属于 Pro / Flagship 路线，详见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md) 与 [`README.md`](README.md) 的 Edition Boundary 段落。
