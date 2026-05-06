# AGENTS.md — AI 编程代理入口（熵剪旗舰版 P1 第一轮小修加固软约束）

> 本文件给所有非 Claude Code 的 AI 编程代理（包括 Codex 等）阅读。Claude Code 请同时阅读 [`CLAUDE.md`](CLAUDE.md)。

## 在做任何动作前，先读这两份文件

1. [`LONGMA_SOFT_GUARD.md`](LONGMA_SOFT_GUARD.md) — 人类可读总章。
2. [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json) — 机器可读锚点；与上文冲突时**以本锚点为准**。

## 当前轮唯一目标

> 在不破坏 Core、不改变 P0 主契约、不引入新依赖的前提下，对熵剪旗舰版 P0 MVR 做第一轮小修加固（FLAGSHIP_P1_HARDENING_ROUND_1，限 H1–H5 共 5 项）。

本轮允许写入的代码 / 文档路径**只**有：

- `internal/flagship/**`
- `cmd/flagship-server/**`
- `tests/flagship/**`
- `docs/FLAGSHIP_P0_FREEZE_REPORT.md`
- `docs/FLAGSHIP_P0_DEV.md`
- `LONGMA_TASK_ANCHOR.json`

> 相比 P0 第一轮，本轮**不再允许**写入 `schemas/flagship/**` 与 `examples/flagship/**`（已随 `v0.4.0-flagship-p0` 冻结）。

其它一切只允许 read / list / grep，**不允许写**。Core 引擎、对外契约（API / Schema / SDK / OpenAPI）、policies、ledger、产品 README / docs / SUPPORT、构建产物（Dockerfile / docker-compose / go.mod / go.sum）、`.gitignore`、AGENTS.md / CLAUDE.md / LONGMA_SOFT_GUARD.md、`schemas/flagship/**`、`examples/flagship/**` 全部禁止改动。

## 本轮允许做的（H1–H5，5 项；缺一不可，多一也不可）

- **H1** HTTP handler 测试覆盖（405、bad JSON、`DisallowUnknownFields`、`/health` GET / 非 GET）
- **H2** NO low-score 分支测试（`Score < T2` 且无硬冲突 → NO，`reason_code = FLAGSHIP_REASONER_NO_LOW_SCORE`）
- **H3** `state.Compute` 返回的 `NormalizedWeights` 必须是 map 副本，避免共享传入 map 的引用
- **H4** `output.NewRejectInstruction` 生成的 ID 必须纳入 `conflicting_items` 内容摘要，不得只依赖 `len(conflicts)`
- **H5** reasoner 中 `CanonicalJSON` 返回 error 不得静默忽略；出错时必须写入 trace 或 fallback digest reason

## 非目标（本轮严禁触达）

- 不接入真实 LLM
- 不开发 LLM Gateway
- 不开发完整龙码 AIOS
- 不开发知识资产化操作系统
- 不开发能力工具系统
- 不开发后台管理系统
- 不做多租户 SaaS
- 不做规则自动生成
- 不接入 Core ledger（含 `internal/ledger`、`ledger/`）
- 不接入 Core signature（含 `internal/signature`）
- 不做鉴权 / 限流 / TLS
- 不引入 jsonschema 等任何新依赖
- 不修改 `go.mod` / `go.sum`
- 不修改现有 Core /shear 接口
- 不修改 `policies/manifest.json`
- 不修改 `openapi.yaml`
- 不修改 `sdk/**`
- 不修改 `Dockerfile` / `docker-compose.yml`
- 不在本轮纳入 H1–H5 之外的任何加固 / 扩展 / 重构
- 不改变 P0 主契约（请求 / 响应 JSON 形状、verdict 大写 `YES` / `HOLD` / `NO`、矩阵数值、`λ μ T1 T2`、风险因子、`AlignmentTask` / `PermitToken` / `RejectInstruction` / `AuditRecord` 字段集）
- 不把龙码常数作为 Core 版正式能力发布

## 五条最低红线

| # | 反模式 | 触发即 HOLD |
|---|---|---|
| AP1 | 目标漂移 | 偏离上文唯一目标（例如：从"P1 第一轮 H1–H5 小修加固"扩张到"LLM Gateway / 多租户 SaaS / 规则自动生成 / 完整龙码 AIOS"；或在 P1 第一轮中加入 H1–H5 之外的任何加固 / 扩展 / 重构；或回头改 P0 主契约） |
| AP2 | 行动越权 | 写禁止路径（含 `schemas/flagship/**` 与 `examples/flagship/**`）/ 自行 `git commit` / `push` / `tag` / 改外部依赖 / 接入真实 LLM / 改 Core /shear / 改 `policies/manifest.json` / 引入 jsonschema 等新依赖 / 对接 Core ledger / signature / 加鉴权 / 限流 / TLS |
| AP3 | 路径漂移 | 对真实不存在的目录或文件操作（写入前必须 `ls` / Read 验证；本轮 `internal/flagship`、`cmd/flagship-server`、`tests/flagship` 已在 `v0.4.0-flagship-p0` 落地，不得自行新建错误路径） |
| AP4 | 幻觉判断 | 引用未经 Read / grep 验证的代码、字段、文档；引用训练数据中的"熵剪 / 龙码常数 / 三态推理"印象；自行改矩阵数值或阈值 |
| AP5 | 嘴炮完成 | 完成无证据 |

## 完成必须给出

1. changed files（清单 + 用途）
2. `git diff --stat`（含 untracked 列举）
3. 执行过的校验命令及结果（至少 `go build ./internal/flagship/... ./cmd/flagship-server/...`、`go vet ./internal/flagship/... ./cmd/flagship-server/... ./tests/flagship/...`、`go test ./tests/flagship/...` 与 `go test ./...` 各跑一次）
4. P0 主契约检查（`TestDefaultMatrixFixed` 仍通过；verdict 仍为 `YES` / `HOLD` / `NO` 大写；`AlignmentTask` / `PermitToken` / `RejectInstruction` / `AuditRecord` 字段未减）
5. 边界检查（是否触碰禁止路径含 `schemas/flagship/**` 与 `examples/flagship/**`、是否改 `policies/manifest.json`、是否动 Core 引擎、是否接入真实 LLM、是否引入新依赖、是否动 `go.mod` / `go.sum`、是否在 H1–H5 之外做了加固）
6. H1–H5 各项状态（已实现 / 未实现 / 部分实现）+ 触达文件 + 对应测试
7. 风险与 HOLD 项

## 边界声明

熵剪旗舰版 P1 第一轮小修加固**不是** Entropy Shear Core 的对外能力。它**不进入** `policies/`、**不修改** `policies/manifest.json`、**不被** `openapi.yaml` 与 `sdk/` 暴露、**不被** README / README_CN / SUPPORT / `docs/PRODUCT_MATRIX.md` 宣传为 Core 能力。Longma Constant 仍仅属于 Pro / Flagship 路线，详见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md) 与 [`README.md`](README.md) 的 Edition Boundary 段落。
