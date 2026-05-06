# CLAUDE.md — Claude Code 入口（熵剪旗舰版 + LLM 编程大脑网关 P0 软约束）

> 本文件给 Claude Code 阅读。其他 AI 编程代理请读 [`AGENTS.md`](AGENTS.md)。

## 在做任何动作前，先读这两份文件

1. [`LONGMA_SOFT_GUARD.md`](LONGMA_SOFT_GUARD.md) — 人类可读总章（详细约束）。
2. [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json) — 机器可读锚点；与 MD 冲突时**以 JSON 为准**。

## 当前轮唯一目标

> 在不破坏 Core、不破坏 Flagship P0/P1 推理内核（v0.4.1-flagship-p1-hardening 已冻结）、不引入真实密钥、不修改现有 /shear 接口的前提下，开发一个可被 Claude Code / Codex 未来接入的编程大脑网关骨架（FLAGSHIP_CODER_GATEWAY_P0）：独立服务、Anthropic Messages 兼容形态最小骨架、生成前与生成后均通过旗舰版 reasoner 治理、占位 mock LLM provider，限 G1–G12 共 12 项。

## 本轮允许写入的代码 / 文档路径（且仅这些）

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

其它一切只允许 Read / ls / grep。Core 引擎、Core 对外契约、policies、ledger、产品 README / docs / SUPPORT、构建产物、`.gitignore`、AGENTS.md / CLAUDE.md / LONGMA_SOFT_GUARD.md、`internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`、`cmd/flagship-server/**`、`tests/flagship/**`、`schemas/flagship/**`、`examples/flagship/**`、既有 `docs/FLAGSHIP_P0_DEV.md` / `docs/FLAGSHIP_P0_FREEZE_REPORT.md` / `docs/FLAGSHIP_P1_HARDENING_REPORT.md` 全部禁止改动。

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

## Claude Code 专属操作守则

1. **写动作前先读真实文件**：用 `Read` 读取目标文件（若已存在），用 `Bash: ls` / `Bash: find` 验证父目录真实存在。绝不依赖记忆中的目录结构。本轮 `internal/flagship/{gateway,provider,coder}`、`cmd/flagship-coder-gateway`、`schemas/flagship-gateway`、`examples/flagship-gateway`、`tests/flagship-gateway` 等目录尚未创建；第一次写入前要用 `Bash: mkdir -p` 显式建目录，并把动作显式同步给用户确认。
2. **引用接口前先 grep**：在引用 `internal/flagship/reasoner` 等包的函数签名、字段名、错误码之前，用 `Bash: grep -rn` 在真实代码中验证。**严禁参照训练数据中的"熵剪 / 龙码常数 / 三态推理 / Anthropic Messages 形状"印象**，一切以当前仓库文件为准；P0/P1 主契约（矩阵数值、`λ μ T1 T2`、风险因子、verdict 大写、字段名、reason_code）以 `v0.4.1-flagship-p1-hardening` 实际仓库为准；Anthropic Messages 兼容形态以 Claude Code 实际期待的最小子集为准（不得自行扩展高级字段）。
3. **不改 Core 引擎与对外契约 / 不改 P0/P1 已冻结的推理内核**：禁止修改 `internal/{api,engine,errors,ledger,policy,schema,signature}/**`、`internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`、`cmd/{server,verify-ledger,validate-policy,hash-policy}/**`、`cmd/flagship-server/**`、`schemas/*.json`（顶层）、`schemas/flagship/**`、`sdk/**`、`policies/**`、`integrations/**`、`openapi.yaml`、`tests/*.go`（顶层）、`tests/flagship/**`、`ledger/**`、`examples/*.json`（顶层）、`examples/flagship/**`、`docs/{AGENT_TOOL_GATE_GUIDE,INTEGRATION_GUIDE,P1_RELEASE_CHECKLIST,POLICY_PACK_GUIDE,PRODUCT_MATRIX,WHITEPAPER}.md`、`docs/FLAGSHIP_P0_DEV.md`、`docs/FLAGSHIP_P0_FREEZE_REPORT.md`、`docs/FLAGSHIP_P1_HARDENING_REPORT.md`、`README.md` / `README_CN.md` / `SUPPORT.md`、`Dockerfile` / `docker-compose.yml`、`go.mod` / `go.sum`、`.gitignore`、`AGENTS.md` / `CLAUDE.md` / `LONGMA_SOFT_GUARD.md`。所有网关 P0 代码 / 测试 / docs 只能落入 `allowed_files`；网关代码只能在 runtime 通过 import 方式调用 `internal/flagship/reasoner`。
4. **不接入真实 LLM、不读真实 key、不引入新依赖、不部署生产**：本轮 mock provider 占位；不接入任何 LLM；不读取 `.env` / `.env.*` / `secrets/`；不 `go get` / 不升级 `go.mod` / 不 `go mod tidy` 引入新包（含真实 LLM provider SDK / jsonschema 等运行期校验库）；不动 docker compose / 不 docker push / 不 `git tag`。
5. **不改 P0/P1 主契约**：矩阵数值、`λ μ T1 T2`、风险因子、verdict 大写、`AlignmentTask` / `PermitToken` / `RejectInstruction` / `AuditRecord` 字段一律不动；`TestDefaultMatrixFixed` 的守门必须继续通过。
6. **越界即 HOLD**：写动作目标若落在 [`LONGMA_SOFT_GUARD.md`](LONGMA_SOFT_GUARD.md) 第 3 节的禁止路径内（含 P0/P1 已冻结的推理内核源码 / 服务 / 测试 / schema / 示例 / docs），或要在 G1–G12 之外做能力扩展，立即停下并把决策权交还给用户。`.claude/settings.json` 已对常见禁止路径配置了 `permissions.deny` 兜底，但这**只是兜底**，不是免责。
7. **不要自行 git 提交**：不自行 `git commit` / `git push` / `git tag` / 任何 release 动作；本轮所有产物都应停留在 working tree，由用户决定是否落库。
8. **完成必须给证据**：交付时必须输出 changed files、`git diff --stat`、`go build` / `go vet` / `go test ./tests/flagship-gateway/...` / `go test ./...` 的输出、P0/P1 主契约检查（`TestDefaultMatrixFixed` 等仍通过、P0/P1 源码未被改动）、边界检查、G1–G12 各项状态、风险与 HOLD 项。少一项视为未完成。
9. **进入网关 P1 / 旗舰版下一阶段前必须先换锚点**：进入网关 P1 或更后阶段前，必须先更新 [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json)，把 `target_constant.single_goal`、`scope_id`、`scope_kind`、`authorization_radius.allowed_files`、`forbidden_boundaries`、`flagship_coder_gateway_p0_scope` 等字段切换为下一阶段的真实开发目标，并同步更新 `.claude/settings.json` 的 `permissions.deny`、本文件、`AGENTS.md`、`LONGMA_SOFT_GUARD.md`。**不得沿用本轮"Coder Gateway P0 骨架（G1–G12）"的授权半径**去做下一阶段开发；沿用即触发 AP1（目标漂移）+ AP2（行动越权），立即 HOLD。锚点未完成切换前，仅可在本轮 `allowed_files` 内写入；超出范围只允许 Read / ls / grep（含 import P0/P1 已冻结的旗舰版 reasoner 包），不得对 Core / Pro / Flagship 任何代码进行 Write / Edit / MultiEdit。

## 五种反模式（速查）

| # | 反模式 | 触发即 HOLD |
|---|---|---|
| AP1 | 目标漂移 | 偏离上文唯一目标（含把网关扩张为完整龙码 AIOS / 知识资产化 / 能力工具系统 / 后台管理 / 多租户 SaaS / 规则自动生成；或在本轮中接入真实 LLM、加 streaming / tool_use / citation / cache_control / 鉴权 / 限流 / TLS、引入新依赖；或回头改 P0/P1 已冻结的推理内核源码 / 服务 / 测试 / schema / 示例 / docs；或加 G1–G12 之外的能力） |
| AP2 | 行动越权 | 写禁止路径（含 P0/P1 已冻结的旗舰版推理内核源码 / 服务 / 测试 / schema / 示例 / docs）/ 自行 `git commit` / `push` / `tag` / 改外部依赖 / 接入真实 LLM / 读取真实 API key / 改 Core /shear / 改 `policies/manifest.json` / 引入 jsonschema 等新依赖 / 对接 Core ledger / signature / 加鉴权 / 限流 / TLS |
| AP3 | 路径漂移 | 对真实不存在的目录或文件操作 |
| AP4 | 幻觉判断 | 引用未读过的代码、字段、文档；引用训练数据中的"熵剪 / 龙码常数 / 三态推理 / Anthropic Messages 形状"印象；自行改矩阵数值或阈值；自行扩展 Anthropic Messages 高级字段 |
| AP5 | 嘴炮完成 | 完成无证据 |

## 边界声明

熵剪旗舰版 + LLM 编程大脑网关 P0 骨架**不是** Entropy Shear Core 的对外能力。它**不进入** `policies/`、**不修改** `policies/manifest.json`、**不被** `openapi.yaml` 与 `sdk/` 暴露、**不被** README / README_CN / SUPPORT 宣传为 Core 能力。Longma Constant 与本网关骨架仍仅属于 Pro / Flagship 路线，详见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md) 与 [`README.md`](README.md) 的 Edition Boundary 段落。
