# CLAUDE.md — Claude Code 入口（熵剪旗舰版编程大脑网关 P1 Claude Code 端到端适配验证软约束）

> 本文件给 Claude Code 阅读。其他 AI 编程代理请读 [`AGENTS.md`](AGENTS.md)。

## 在做任何动作前，先读这两份文件

1. [`LONGMA_SOFT_GUARD.md`](LONGMA_SOFT_GUARD.md) — 人类可读总章（详细约束）。
2. [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json) — 机器可读锚点；与 MD 冲突时**以 JSON 为准**。

## 当前轮唯一目标

> 在不接真实 LLM、不读取真实 API key、不破坏 Core、不破坏 Flagship P0/P1 推理内核（`v0.4.1-flagship-p1-hardening` 已冻结）、不改变 Gateway P0 主接口（`v0.5.0-flagship-coder-gateway-p0` 已冻结的请求 / 响应 JSON 形状、verdict 大写、stop_reason 取值、GatewayAuditRecord 字段）的前提下，验证 Claude Code 是否可以通过当前网关接口完成最小端到端调用（FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E，限 E1–E9 共 9 项）。

## 本轮允许写入的代码 / 文档路径（且仅这些）

- `cmd/flagship-coder-gateway/**`
- `internal/flagship/gateway/**`
- `internal/flagship/coder/**`
- `internal/flagship/provider/**`
- `tests/flagship-gateway/**`
- `examples/flagship-gateway/**`
- `docs/FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E.md`
- `LONGMA_TASK_ANCHOR.json`

> 本轮**不再允许**写入 `schemas/flagship-gateway/**`（Gateway P0 schema 已冻结）与既有 `docs/FLAGSHIP_CODER_GATEWAY_P0_DEV.md` / `docs/FLAGSHIP_CODER_GATEWAY_P0_FREEZE_REPORT.md`。

其它一切只允许 Read / ls / grep。Core 引擎、Core 对外契约、policies、ledger、产品 README / docs / SUPPORT、构建产物、`.gitignore`、AGENTS.md / CLAUDE.md / LONGMA_SOFT_GUARD.md、`internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`、`cmd/flagship-server/**`、`tests/flagship/**`、`schemas/flagship/**`、`examples/flagship/**`、`schemas/flagship-gateway/**`、3 份 P0/P1 docs 与 2 份 Gateway P0 docs 全部禁止改动。

## 本轮允许做的（E1–E9，9 项；缺一不可，多一也不可）

- **E1** 编写 Claude Code 接入本地网关运行说明（在 `docs/FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E.md` 内；不修改用户机器上的真实 Claude Code 配置）
- **E2** 验证 `GET /v1/models`
- **E3** 验证 `POST /v1/messages/count_tokens`
- **E4** 验证 `POST /v1/messages` 三态（YES / HOLD / NO）
- **E5** 验证 Claude Code 需要的请求头（anthropic-version / anthropic-beta / authorization / x-api-key / content-type / user-agent / x-app 等）可被接收并忽略
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
- 不做 streaming（SSE / chunked）
- 不做 tool_use / tool_result（function calling）
- 不做 citation / cache_control / thinking / images / documents 等 Anthropic 高级能力
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
- 不实现 Anthropic Messages 高级能力
- 不在本轮纳入 E1–E9 之外的任何能力 / 扩展 / 重构
- 不改变 P0/P1 主契约（请求 / 响应 JSON 形状、verdict 大写 `YES` / `HOLD` / `NO`、矩阵数值、`λ μ T1 T2`、风险因子、`AlignmentTask` / `PermitToken` / `RejectInstruction` / `AuditRecord` 字段集）
- 不改变 Gateway P0 主接口（请求 / 响应 JSON 形状、`stop_reason` 取值、HOLD/NO 一律 200、`GatewayAuditRecord` 字段）
- 不把龙码常数作为 Core 版正式能力发布

## Claude Code 专属操作守则

1. **写动作前先读真实文件**：用 `Read` 读取目标文件（若已存在），用 `Bash: ls` / `Bash: find` 验证父目录真实存在。绝不依赖记忆中的目录结构。本轮 `cmd/flagship-coder-gateway`、`internal/flagship/{gateway,coder,provider}`、`tests/flagship-gateway`、`examples/flagship-gateway` 均已存在（Gateway P0 已落地），不得自行新建错误路径。`docs/FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E.md` 是新文件，第一次写入需对用户透明。
2. **引用接口前先 grep**：在引用 `internal/flagship/reasoner` / `internal/flagship/coder` / `internal/flagship/provider` / `internal/flagship/gateway` 等包的函数签名、字段名、错误码之前，用 `Bash: grep -rn` 在真实代码中验证。**严禁参照训练数据中的"熵剪 / 龙码常数 / 三态推理 / Anthropic Messages 形状 / Claude Code 行为"印象**，一切以当前仓库文件为准；P0/P1 主契约（矩阵数值、`λ μ T1 T2`、风险因子、verdict 大写、字段名、reason_code）以 `v0.4.1-flagship-p1-hardening` 实际仓库为准；Gateway P0 主接口（请求 / 响应 JSON 形状、`stop_reason` 取值、HOLD/NO 一律 200、`GatewayAuditRecord` 字段）以 `v0.5.0-flagship-coder-gateway-p0` 实际仓库为准；Claude Code 实测期待以真实运行结果为准，不得凭印象推断。
3. **不改 Core 引擎与对外契约 / 不改 P0/P1 已冻结推理内核 / 不改 Gateway P0 已冻结 schemas 与 docs**：禁止修改 `internal/{api,engine,errors,ledger,policy,schema,signature}/**`、`internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`、`cmd/{server,verify-ledger,validate-policy,hash-policy}/**`、`cmd/flagship-server/**`、`schemas/*.json`（顶层）、`schemas/flagship/**`、`schemas/flagship-gateway/**`、`sdk/**`、`policies/**`、`integrations/**`、`openapi.yaml`、`tests/*.go`（顶层）、`tests/flagship/**`、`ledger/**`、`examples/*.json`（顶层）、`examples/flagship/**`、`docs/{AGENT_TOOL_GATE_GUIDE,INTEGRATION_GUIDE,P1_RELEASE_CHECKLIST,POLICY_PACK_GUIDE,PRODUCT_MATRIX,WHITEPAPER}.md`、`docs/FLAGSHIP_P0_DEV.md`、`docs/FLAGSHIP_P0_FREEZE_REPORT.md`、`docs/FLAGSHIP_P1_HARDENING_REPORT.md`、`docs/FLAGSHIP_CODER_GATEWAY_P0_DEV.md`、`docs/FLAGSHIP_CODER_GATEWAY_P0_FREEZE_REPORT.md`、`README.md` / `README_CN.md` / `SUPPORT.md`、`Dockerfile` / `docker-compose.yml`、`go.mod` / `go.sum`、`.gitignore`、`AGENTS.md` / `CLAUDE.md` / `LONGMA_SOFT_GUARD.md`。所有网关 P1 代码 / 测试 / docs 只能落入 `allowed_files`；网关代码只能在 runtime 通过 import 方式调用 `internal/flagship/reasoner` 等已冻结包。
4. **不接入真实 LLM、不读真实 key、不引入新依赖、不部署生产**：本轮 mock provider 占位；不接入任何 LLM；不读取 `.env` / `.env.*` / `secrets/`；不 `go get` / 不升级 `go.mod` / 不 `go mod tidy` 引入新包（含真实 LLM provider SDK / jsonschema 等）；不动 docker compose / 不 docker push / 不 `git tag`。
5. **不改 P0/P1 主契约 与 Gateway P0 主接口**：矩阵数值、`λ μ T1 T2`、风险因子、verdict 大写、`AlignmentTask` / `PermitToken` / `RejectInstruction` / `AuditRecord` 字段、`stop_reason` 取值、HOLD/NO 一律 200、`GatewayAuditRecord` 字段一律不动；`TestDefaultMatrixFixed` 与 Gateway P0 既有 e2e 测试必须继续通过。
6. **越界即 HOLD**：写动作目标若落在 [`LONGMA_SOFT_GUARD.md`](LONGMA_SOFT_GUARD.md) 第 3 节的禁止路径内（含 P0/P1 已冻结的推理内核源码 / 服务 / 测试 / schema / 示例 / docs，以及 Gateway P0 已冻结 schemas / dev / freeze 报告），或要在 E1–E9 之外做能力扩展，立即停下并把决策权交还给用户。`.claude/settings.json` 已对常见禁止路径配置了 `permissions.deny` 兜底，但这**只是兜底**，不是免责。
7. **不要自行 git 提交**：不自行 `git commit` / `git push` / `git tag` / 任何 release 动作；本轮所有产物都应停留在 working tree，由用户决定是否落库。
8. **不动用户机器上的真实 Claude Code 配置**：E1 只允许在 `docs/FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E.md` 写指引；不允许自行执行命令修改用户的 `~/.claude` / Claude Code 设置；用户接入是用户自己的动作。
9. **完成必须给证据**：交付时必须输出 changed files、`git diff --stat`、`go build` / `go vet` / `go test ./tests/flagship-gateway/...` / `go test ./...` 的输出、Claude Code 端到端实测记录（curl / shell 输出 / 关键响应片段）、P0/P1 主契约与 Gateway P0 主接口未破坏的检查、边界检查、E1–E9 各项状态、风险与 HOLD 项。少一项视为未完成。
10. **进入网关 P2 / 旗舰版下一阶段前必须先换锚点**：进入网关 P2 或更后阶段前，必须先更新 [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json)，把 `target_constant.single_goal`、`scope_id`、`scope_kind`、`authorization_radius.allowed_files`、`forbidden_boundaries`、`flagship_coder_gateway_p1_scope` 等字段切换为下一阶段的真实开发目标，并同步更新 `.claude/settings.json` 的 `permissions.deny`、本文件、`AGENTS.md`、`LONGMA_SOFT_GUARD.md`。**不得沿用本轮"网关 P1 Claude E2E（E1–E9）"的授权半径**去做下一阶段开发；沿用即触发 AP1（目标漂移）+ AP2（行动越权），立即 HOLD。锚点未完成切换前，仅可在本轮 `allowed_files` 内写入；超出范围只允许 Read / ls / grep（含 import P0/P1 已冻结的旗舰版 reasoner 包），不得对 Core / Pro / Flagship 任何代码进行 Write / Edit / MultiEdit。

## 五种反模式（速查）

| # | 反模式 | 触发即 HOLD |
|---|---|---|
| AP1 | 目标漂移 | 偏离上文唯一目标（含把网关 P1 扩张为完整龙码 AIOS / 知识资产化 / 能力工具系统 / 后台管理 / 多租户 SaaS / 规则自动生成；或在本轮中接入真实 LLM、加 streaming / tool_use / citation / cache_control / 鉴权 / 限流 / TLS、引入新依赖；或回头改 P0/P1 已冻结推理内核源码 / 服务 / 测试 / schema / 示例 / docs，或改 Gateway P0 已冻结 schemas / dev / freeze 报告；或加 E1–E9 之外的能力） |
| AP2 | 行动越权 | 写禁止路径（含 P0/P1 已冻结推理内核与 Gateway P0 已冻结 schemas / docs）/ 自行 `git commit` / `push` / `tag` / 改外部依赖 / 接入真实 LLM / 读取真实 API key / 改 Core /shear / 改 `policies/manifest.json` / 引入 jsonschema 等新依赖 / 对接 Core ledger / signature / 加鉴权 / 限流 / TLS / 修改用户机器上的 Claude Code 配置 |
| AP3 | 路径漂移 | 对真实不存在的目录或文件操作 |
| AP4 | 幻觉判断 | 引用未读过的代码、字段、文档；引用训练数据中的"熵剪 / 龙码常数 / 三态推理 / Anthropic Messages 形状 / Claude Code 行为"印象；自行改矩阵数值、阈值、verdict 大小写或 Gateway P0 主接口字段 |
| AP5 | 嘴炮完成 | 完成无证据，特别是缺 Claude Code 端到端实测记录 |

## 边界声明

熵剪旗舰版编程大脑网关 P1 Claude Code 端到端适配验证**不是** Entropy Shear Core 的对外能力。它**不进入** `policies/`、**不修改** `policies/manifest.json`、**不被** `openapi.yaml` 与 `sdk/` 暴露、**不被** README / README_CN / SUPPORT 宣传为 Core 能力。Longma Constant、龙码三态逻辑推理内核与本网关骨架仍仅属于 Pro / Flagship 路线，详见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md) 与 [`README.md`](README.md) 的 Edition Boundary 段落。
