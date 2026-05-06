# AGENTS.md — AI 编程代理入口（熵剪旗舰版 P0 开发软约束）

> 本文件给所有非 Claude Code 的 AI 编程代理（包括 Codex 等）阅读。Claude Code 请同时阅读 [`CLAUDE.md`](CLAUDE.md)。

## 在做任何动作前，先读这两份文件

1. [`LONGMA_SOFT_GUARD.md`](LONGMA_SOFT_GUARD.md) — 人类可读总章。
2. [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json) — 机器可读锚点；与上文冲突时**以本锚点为准**。

## 当前轮唯一目标

> 在不破坏现有 Entropy Shear Core 的前提下，开发熵剪旗舰版 P0 的龙码三态逻辑推理内核最小可运行版本（Minimum Verifiable Reasoner）。

本轮允许写入的代码 / 文档路径**只**有：

- `internal/flagship/**`
- `cmd/flagship-server/**`
- `schemas/flagship/**`
- `examples/flagship/**`
- `tests/flagship/**`
- `docs/FLAGSHIP_P0_DEV.md`
- `LONGMA_TASK_ANCHOR.json`

其它一切只允许 read / list / grep，**不允许写**。Core 引擎、对外契约（API / Schema / SDK / OpenAPI）、policies、ledger、产品 README / docs / SUPPORT、构建产物（Dockerfile / docker-compose / go.mod / go.sum）、`.gitignore`、AGENTS.md / CLAUDE.md / LONGMA_SOFT_GUARD.md 全部禁止改动。

## 非目标（本轮严禁触达）

- 不开发完整龙码 AIOS
- 不开发 LLM Gateway
- 不接入真实 LLM
- 不开发知识资产化操作系统
- 不开发能力工具系统
- 不开发后台管理系统
- 不做多租户 SaaS
- 不做规则自动生成
- 不改现有 Core /shear 接口
- 不改现有 policy + facts 裁决逻辑
- 不改现有 ledger / signature / policy pack 行为
- 不把龙码常数作为 Core 版正式能力发布

## 五条最低红线

| # | 反模式 | 触发即 HOLD |
|---|---|---|
| AP1 | 目标漂移 | 偏离上文唯一目标（例如：从"龙码三态逻辑推理内核 MVR"扩张到"LLM Gateway / 多租户 SaaS / 规则自动生成 / 完整龙码 AIOS"） |
| AP2 | 行动越权 | 写禁止路径 / 自行 `git commit` / `push` / `tag` / 改外部依赖 / 接入真实 LLM / 改 Core /shear / 改 `policies/manifest.json` |
| AP3 | 路径漂移 | 对真实不存在的目录或文件操作（写入前必须 `ls` / Read 验证；本轮 `internal/flagship` 等目录初始为空，第一次写入需显式 `mkdir -p` 并对用户透明） |
| AP4 | 幻觉判断 | 引用未经 Read / grep 验证的代码、字段、文档；引用训练数据中的"熵剪 / 龙码常数 / 三态推理"印象 |
| AP5 | 嘴炮完成 | 完成无证据 |

## 完成必须给出

1. changed files（清单 + 用途）
2. `git diff --stat`（含 untracked 列举）
3. 执行过的校验命令及结果（flagship 子模块至少 `go build ./internal/flagship/... ./cmd/flagship-server/...` 与 `go test ./tests/flagship/...` 各跑一次）
4. 边界检查（是否触碰禁止路径、是否改 `policies/manifest.json`、是否动 Core 引擎、是否接入真实 LLM、是否引入新依赖）
5. 风险与 HOLD 项

## 边界声明

熵剪旗舰版 P0 龙码三态逻辑推理内核**不是** Entropy Shear Core 的对外能力。它**不进入** `policies/`、**不修改** `policies/manifest.json`、**不被** `openapi.yaml` 与 `sdk/` 暴露、**不被** README / README_CN / SUPPORT / `docs/PRODUCT_MATRIX.md` 宣传为 Core 能力。Longma Constant 仍仅属于 Pro / Flagship 路线，详见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md) 与 [`README.md`](README.md) 的 Edition Boundary 段落。
