# AGENTS.md — AI 编程代理入口（熵剪旗舰版编程大脑网关 P1 Claude Code 端到端适配验证软约束）

> 本文件给所有非 Claude Code 的 AI 编程代理（包括 Codex 等）阅读。Claude Code 请同时阅读 [`CLAUDE.md`](CLAUDE.md)。

## 在做任何动作前，先读这两份文件

1. [`LONGMA_SOFT_GUARD.md`](LONGMA_SOFT_GUARD.md) — 人类可读总章。
2. [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json) — 机器可读锚点；与上文冲突时**以本锚点为准**。

## 当前轮唯一目标

> 在不接真实 LLM、不读取真实 API key、不破坏 Core、不破坏 Flagship P0/P1 推理内核（`v0.4.1-flagship-p1-hardening` 已冻结）、不改变 Gateway P0 主接口（`v0.5.0-flagship-coder-gateway-p0` 已冻结的请求 / 响应 JSON 形状、verdict 大写、stop_reason 取值、GatewayAuditRecord 字段）的前提下，验证 Claude Code 是否可以通过当前网关接口完成最小端到端调用（FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E，限 E1–E9 共 9 项）。

本轮允许写入的代码 / 文档路径**只**有：

- `cmd/flagship-coder-gateway/**`
- `internal/flagship/gateway/**`
- `internal/flagship/coder/**`
- `internal/flagship/provider/**`
- `tests/flagship-gateway/**`
- `examples/flagship-gateway/**`
- `docs/FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E.md`
- `LONGMA_TASK_ANCHOR.json`

> 本轮**不再允许**写入 `schemas/flagship-gateway/**`（Gateway P0 schema 已冻结）与既有 `docs/FLAGSHIP_CODER_GATEWAY_P0_DEV.md` / `docs/FLAGSHIP_CODER_GATEWAY_P0_FREEZE_REPORT.md`。

其它一切只允许 read / list / grep，**不允许写**。Core 引擎、对外契约（API / Schema / SDK / OpenAPI）、policies、ledger、产品 README / docs / SUPPORT、构建产物（Dockerfile / docker-compose / go.mod / go.sum）、`.gitignore`、AGENTS.md / CLAUDE.md / LONGMA_SOFT_GUARD.md、`internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`、`cmd/flagship-server/**`、`tests/flagship/**`、`schemas/flagship/**`、`examples/flagship/**`、`schemas/flagship-gateway/**`、3 份 P0/P1 docs 与 2 份 Gateway P0 docs 全部禁止改动。

## 本轮允许做的（E1–E9，9 项；缺一不可，多一也不可）

- **E1** 编写 Claude Code 接入本地网关运行说明（仅指引，不动用户机器配置）
- **E2** 验证 `GET /v1/models`
- **E3** 验证 `POST /v1/messages/count_tokens`
- **E4** 验证 `POST /v1/messages` 三态（YES / HOLD / NO）
- **E5** 验证 Claude Code 需要的请求头可被接收并忽略
- **E6** 验证 HOLD / NO 统一 HTTP 200 且 `body.verdict` 表达三态
- **E7** 补充最小 e2e 脚本或测试样例
- **E8** 输出 Claude Code 联调记录文档
- **E9** 保持 Mock Provider，不接真实 Provider

## 非目标（本轮严禁触达）

- 不接真实 Claude / OpenAI / Gemini / Qwen API
- 不读取任何真实 API key
- 不写 `.env` / 不改 `.env.*` / 不读 `secrets/`
- 不做真实 API Key 管理
- 不做鉴权 / 限流 / TLS
- 不做 streaming
- 不做 tool_use / tool_result
- 不做 citation / cache_control / thinking / images / documents
- 不做真实 Claude Code 自动替换配置
- 不引入任何新依赖
- 不修改 `go.mod` / `go.sum`
- 不修改现有 Core /shear 接口
- 不修改 `policies/manifest.json`
- 不修改 `openapi.yaml`
- 不修改 `sdk/**`
- 不修改 `Dockerfile` / `docker-compose.yml`
- 不修改 `README.md` / `README_CN.md` / `SUPPORT.md`
- 不修改 P0/P1 已冻结推理内核 / 服务 / 测试 / schema / 示例 / 三份 P0/P1 docs
- 不修改 Gateway P0 已冻结 `schemas/flagship-gateway/**` 与两份 dev / freeze 报告
- 不在本轮纳入 E1–E9 之外的任何能力 / 扩展 / 重构
- 不改变 P0/P1 主契约 与 Gateway P0 主接口
- 不把龙码常数作为 Core 版正式能力发布

## 五条最低红线

| # | 反模式 | 触发即 HOLD |
|---|---|---|
| AP1 | 目标漂移 | 偏离上文唯一目标（例如：从"网关 P1 Claude E2E（E1–E9）"扩张到"完整龙码 AIOS / 多租户 SaaS / 规则自动生成"；或在本轮中接入真实 LLM、加 streaming / tool_use / citation / cache_control / 鉴权 / 限流 / TLS、引入新依赖；或回头改 P0/P1 已冻结推理内核源码 / 服务 / 测试 / schema / 示例 / docs，或改 Gateway P0 已冻结 schemas / dev / freeze 报告；或加 E1–E9 之外的能力） |
| AP2 | 行动越权 | 写禁止路径（含 P0/P1 已冻结推理内核与 Gateway P0 已冻结 schemas / docs）/ 自行 `git commit` / `push` / `tag` / 改外部依赖 / 接入真实 LLM / 读取真实 API key / 改 Core /shear / 改 `policies/manifest.json` / 引入 jsonschema 等新依赖 / 对接 Core ledger / signature / 加鉴权 / 限流 / TLS / 修改用户机器上的 Claude Code 配置 |
| AP3 | 路径漂移 | 对真实不存在的目录或文件操作（写入前必须 `ls` / Read 验证；本轮 `cmd/flagship-coder-gateway`、`internal/flagship/{gateway,coder,provider}`、`tests/flagship-gateway`、`examples/flagship-gateway` 均已存在，不得自行新建错误路径或重命名既有目录） |
| AP4 | 幻觉判断 | 引用未经 Read / grep 验证的代码、字段、文档；引用训练数据中的"熵剪 / 龙码常数 / 三态推理 / Anthropic Messages 形状 / Claude Code 行为"印象；自行改矩阵数值、阈值、verdict 大小写或 Gateway P0 主接口字段 |
| AP5 | 嘴炮完成 | 完成无证据，特别是缺 Claude Code 端到端实测记录 |

## 完成必须给出

1. changed files（清单 + 用途）
2. `git diff --stat`（含 untracked 列举）
3. 执行过的校验命令及结果（至少 `go build ./internal/flagship/... ./cmd/flagship-server/... ./cmd/flagship-coder-gateway/...`、`go vet ./internal/flagship/... ./cmd/flagship-coder-gateway/... ./tests/flagship-gateway/...`、`go test ./tests/flagship-gateway/... ./cmd/flagship-coder-gateway/...` 与 `go test ./...` 各跑一次）
4. Claude Code 端到端实测记录（curl / shell 输出 / 关键响应片段，至少覆盖 `/v1/models`、`/v1/messages/count_tokens`、`/v1/messages` 三个端点，与 YES / HOLD / NO 三态）
5. P0/P1 主契约 + Gateway P0 主接口未破坏的检查（`TestDefaultMatrixFixed` 与既有 Gateway e2e 测试继续通过；矩阵 / 阈值 / verdict 大写 / `stop_reason` / HOLD-NO-200 / `GatewayAuditRecord` 字段未变；P0/P1 与 Gateway P0 各 freeze / dev 报告对应的源码 / schema / docs 均未被改动）
6. 边界检查（是否触碰禁止路径含 P0/P1 已冻结推理内核与 Gateway P0 已冻结 schemas / docs、是否改 `policies/manifest.json`、是否动 Core 引擎、是否接入真实 LLM、是否读取真实 API key、是否引入新依赖、是否动 `go.mod` / `go.sum`、是否在 E1–E9 之外做了能力扩展）
7. E1–E9 各项状态（已实现 / 未实现 / 部分实现）+ 触达文件 + 对应实测证据
8. 风险与 HOLD 项，特别是 Claude Code 实测中暴露的 P0 不足（如真实期待 streaming / tool_use 等）

## 边界声明

熵剪旗舰版编程大脑网关 P1 Claude Code 端到端适配验证**不是** Entropy Shear Core 的对外能力。它**不进入** `policies/`、**不修改** `policies/manifest.json`、**不被** `openapi.yaml` 与 `sdk/` 暴露、**不被** README / README_CN / SUPPORT / `docs/PRODUCT_MATRIX.md` 宣传为 Core 能力。Longma Constant、龙码三态逻辑推理内核与本网关骨架仍仅属于 Pro / Flagship 路线，详见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md) 与 [`README.md`](README.md) 的 Edition Boundary 段落。
