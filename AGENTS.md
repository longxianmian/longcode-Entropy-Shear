# AGENTS.md — AI 编程代理入口（熵剪旗舰版 + LLM 编程大脑网关 P0 软约束）

> 本文件给所有非 Claude Code 的 AI 编程代理（包括 Codex 等）阅读。Claude Code 请同时阅读 [`CLAUDE.md`](CLAUDE.md)。

## 在做任何动作前，先读这两份文件

1. [`LONGMA_SOFT_GUARD.md`](LONGMA_SOFT_GUARD.md) — 人类可读总章。
2. [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json) — 机器可读锚点；与上文冲突时**以本锚点为准**。

## 当前轮唯一目标

> 在不破坏 Core、不破坏 Flagship P0/P1 推理内核（v0.4.1-flagship-p1-hardening 已冻结）、不引入真实密钥、不修改现有 /shear 接口的前提下，开发一个可被 Claude Code / Codex 未来接入的编程大脑网关骨架（FLAGSHIP_CODER_GATEWAY_P0）：独立服务、Anthropic Messages 兼容形态最小骨架、生成前与生成后均通过旗舰版 reasoner 治理、占位 mock LLM provider，限 G1–G12 共 12 项。

本轮允许写入的代码 / 文档路径**只**有：

- `internal/flagship/gateway/**`
- `internal/flagship/provider/**`
- `internal/flagship/coder/**`
- `cmd/flagship-coder-gateway/**`
- `schemas/flagship-gateway/**`
- `examples/flagship-gateway/**`
- `tests/flagship-gateway/**`
- `docs/FLAGSHIP_CODER_GATEWAY_P0_DEV.md`
- `LONGMA_TASK_ANCHOR.json`

> 本轮**不再允许**写入 P0/P1 已冻结的任何旗舰版推理内核源码 / 服务 / 测试 / schema / 示例 / dev / freeze / hardening 报告。网关只能在 runtime 通过 import 方式调用旗舰版 reasoner 包。

其它一切只允许 read / list / grep，**不允许写**。Core 引擎、对外契约（API / Schema / SDK / OpenAPI）、policies、ledger、产品 README / docs / SUPPORT、构建产物（Dockerfile / docker-compose / go.mod / go.sum）、`.gitignore`、AGENTS.md / CLAUDE.md / LONGMA_SOFT_GUARD.md、`internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`、`cmd/flagship-server/**`、`tests/flagship/**`、`schemas/flagship/**`、`examples/flagship/**`、既有 `docs/FLAGSHIP_P0_DEV.md` / `docs/FLAGSHIP_P0_FREEZE_REPORT.md` / `docs/FLAGSHIP_P1_HARDENING_REPORT.md` 全部禁止改动。

## 本轮允许做的（G1–G12，12 项；缺一不可，多一也不可）

- **G1** 独立网关服务（不改 Core server、独立端口）
- **G2** Anthropic Messages 兼容形态最小接口骨架（messages、role、content blocks、stop_reason、usage 占位；不实现 streaming / tool_use / citation / cache_control 等高级能力）
- **G3** `GET /health`（独立于 Flagship reasoner 与 Core 的 /health）
- **G4** `GET /v1/models`（占位 model 列表）
- **G5** `POST /v1/messages/count_tokens`（占位 token 计数估算）
- **G6** `POST /v1/messages`（主端点，串联 G7→G8→G9→G10→G11）
- **G7** 生成前治理：调用 `internal/flagship/reasoner.Reason()`
- **G8** Mock LLM Provider（不接真实 LLM、不读真实 key）
- **G9** 生成后审查：再次调用 `internal/flagship/reasoner.Reason()`
- **G10** Claude Code 可识别的 assistant message 响应
- **G11** 审计结构（仅返回对象，不写 Core ledger、不写 JSONL）
- **G12** 最小测试用例

## 非目标（本轮严禁触达）

- 不接真实 Claude / OpenAI / Gemini / Qwen API
- 不读取任何真实 API key
- 不写 `.env` / 不改 `.env.*` / 不读 `secrets/`
- 不开发完整龙码 AIOS
- 不开发知识资产化操作系统
- 不开发能力工具系统
- 不开发后台管理系统
- 不做多租户 SaaS
- 不做规则自动生成
- 不接入 Core ledger（含 `internal/ledger`、`ledger/`）
- 不接入 Core signature（含 `internal/signature`）
- 不做鉴权 / 限流 / TLS
- 不引入任何新依赖（含真实 LLM provider SDK）
- 不修改 `go.mod` / `go.sum`
- 不修改现有 Core /shear 接口
- 不修改 `policies/manifest.json`
- 不修改 `openapi.yaml`
- 不修改 `sdk/**`
- 不修改 `Dockerfile` / `docker-compose.yml`
- 不修改 `README.md` / `README_CN.md` / `SUPPORT.md`
- 不修改 P0/P1 已冻结的旗舰版推理内核源码 / 服务 / 测试 / schema / 示例 / 三份 dev / freeze / hardening 报告
- 不实现 Anthropic Messages 的 streaming / tool_use / citation / cache_control 等高级能力
- 不在本轮纳入 G1–G12 之外的任何能力 / 扩展 / 重构
- 不改变 P0/P1 主契约（请求 / 响应 JSON 形状、verdict 大写 `YES` / `HOLD` / `NO`、矩阵数值、`λ μ T1 T2`、风险因子、`AlignmentTask` / `PermitToken` / `RejectInstruction` / `AuditRecord` 字段集）
- 不把龙码常数作为 Core 版正式能力发布

## 五条最低红线

| # | 反模式 | 触发即 HOLD |
|---|---|---|
| AP1 | 目标漂移 | 偏离上文唯一目标（例如：从"网关 P0 骨架（G1–G12）"扩张到"完整龙码 AIOS / 多租户 SaaS / 规则自动生成"；或在本轮中接入真实 LLM、加 streaming / tool_use / citation / cache_control / 鉴权 / 限流 / TLS、引入新依赖；或回头改 P0/P1 已冻结的推理内核源码 / 服务 / 测试 / schema / 示例 / docs；或加 G1–G12 之外的能力） |
| AP2 | 行动越权 | 写禁止路径（含 P0/P1 已冻结的旗舰版推理内核源码 / 服务 / 测试 / schema / 示例 / docs）/ 自行 `git commit` / `push` / `tag` / 改外部依赖 / 接入真实 LLM / 读取真实 API key / 改 Core /shear / 改 `policies/manifest.json` / 引入 jsonschema 等新依赖 / 对接 Core ledger / signature / 加鉴权 / 限流 / TLS |
| AP3 | 路径漂移 | 对真实不存在的目录或文件操作（写入前必须 `ls` / Read 验证；本轮 `internal/flagship/{gateway,provider,coder}`、`cmd/flagship-coder-gateway`、`schemas/flagship-gateway`、`examples/flagship-gateway`、`tests/flagship-gateway` 等目录初始为空，第一次写入需显式 `mkdir -p` 并对用户透明） |
| AP4 | 幻觉判断 | 引用未经 Read / grep 验证的代码、字段、文档；引用训练数据中的"熵剪 / 龙码常数 / 三态推理 / Anthropic Messages 形状"印象；自行改矩阵数值或阈值；自行扩展 Anthropic Messages 高级字段 |
| AP5 | 嘴炮完成 | 完成无证据 |

## 完成必须给出

1. changed files（清单 + 用途）
2. `git diff --stat`（含 untracked 列举）
3. 执行过的校验命令及结果（至少 `go build ./internal/flagship/... ./cmd/flagship-server/... ./cmd/flagship-coder-gateway/...`、`go vet ./internal/flagship/... ./cmd/flagship-coder-gateway/... ./tests/flagship-gateway/...`、`go test ./tests/flagship-gateway/...` 与 `go test ./...` 各跑一次）
4. P0/P1 主契约检查（`TestDefaultMatrixFixed` 仍通过；矩阵数值 / 阈值 / verdict 大写 / 输出字段集未变；P0/P1 各 freeze / hardening 报告对应的源码未被改动）
5. 边界检查（是否触碰禁止路径含 P0/P1 已冻结的推理内核源码 / 服务 / 测试 / schema / 示例 / docs、是否改 `policies/manifest.json`、是否动 Core 引擎、是否接入真实 LLM、是否读取真实 API key、是否引入新依赖、是否动 `go.mod` / `go.sum`、是否在 G1–G12 之外做了能力扩展）
6. G1–G12 各项状态（已实现 / 未实现 / 部分实现）+ 触达文件 + 对应测试
7. 风险与 HOLD 项

## 边界声明

熵剪旗舰版 + LLM 编程大脑网关 P0 骨架**不是** Entropy Shear Core 的对外能力。它**不进入** `policies/`、**不修改** `policies/manifest.json`、**不被** `openapi.yaml` 与 `sdk/` 暴露、**不被** README / README_CN / SUPPORT / `docs/PRODUCT_MATRIX.md` 宣传为 Core 能力。Longma Constant 与本网关骨架仍仅属于 Pro / Flagship 路线，详见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md) 与 [`README.md`](README.md) 的 Edition Boundary 段落。
