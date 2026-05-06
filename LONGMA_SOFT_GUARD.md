# 熵剪旗舰版编程大脑网关 P1 Claude Code 端到端适配验证软约束（LONGMA_SOFT_GUARD）

> 本文件不是熵剪 Core 的产品功能。它只在 AI 编程代理（Claude Code / Codex 等）开发熵剪旗舰版编程大脑网关 P1 Claude Code 端到端适配验证时生效，用于约束开发过程的目标边界与执行边界，治理目标漂移、行动越权、路径漂移、幻觉判断与嘴炮完成。

## 0. 边界声明（必读）

熵剪旗舰版编程大脑网关 P1 Claude Code 端到端适配验证：

- **不属于** Entropy Shear Core 的对外能力。
- **不进入** `policies/`、**不修改** `policies/manifest.json`、**不修改** `tests/policy_pack_test.go`。
- **不被** `openapi.yaml` 与 `sdk/` 暴露。
- **不被** `README.md` / `README_CN.md` / `SUPPORT.md` / `docs/PRODUCT_MATRIX.md` 等产品文档宣传为 Core 能力。
- **不接入** 任何真实 LLM；本轮全部走 mock provider，不读真实 API key、不写 `.env`、不读 `secrets/`。
- **不修改 P0/P1 已冻结的旗舰版推理内核源码**（`internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`、`cmd/flagship-server/**`、`tests/flagship/**`、`schemas/flagship/**`、`examples/flagship/**`、对应的 dev / freeze / hardening 报告）。
- **不修改 Gateway P0 已冻结的 schemas 与 dev / freeze 报告**（`schemas/flagship-gateway/**`、`docs/FLAGSHIP_CODER_GATEWAY_P0_DEV.md`、`docs/FLAGSHIP_CODER_GATEWAY_P0_FREEZE_REPORT.md`）。网关 P1 只能 runtime import 已冻结 reasoner 包；本轮代码动作只允许在网关本身的 gateway / coder / provider / cmd / tests / examples 子目录里展开。
- **不改变 P0/P1 主契约 与 Gateway P0 主接口**：请求 / 响应 JSON 形状、verdict 大写 `YES` / `HOLD` / `NO`、矩阵数值、`λ μ T1 T2`、风险因子、`AlignmentTask` / `PermitToken` / `RejectInstruction` / `AuditRecord` 字段集、`stop_reason` 取值、HOLD/NO 一律 200 的语义、`GatewayAuditRecord` 字段，全部以 `v0.4.1-flagship-p1-hardening` 与 `v0.5.0-flagship-coder-gateway-p0` 实际仓库为准，本轮不得改、不得猜。
- 与 [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json) 互为正副本：本文件给人读，锚点 JSON 给机器读，二者冲突时**以锚点 JSON 为准**。

Longma Constant、龙码三态逻辑推理内核与本网关骨架仍仅属于 Pro / Flagship 路线，详见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md) 与 [`README.md`](README.md) 的 Edition Boundary 段落。

## 1. 唯一目标（target constant）

> **在不接真实 LLM、不读取真实 API key、不破坏 Core、不破坏 Flagship P0/P1 推理内核（`v0.4.1-flagship-p1-hardening` 已冻结）、不改变 Gateway P0 主接口（`v0.5.0-flagship-coder-gateway-p0` 已冻结的请求 / 响应 JSON 形状、verdict 大写、stop_reason 取值、`GatewayAuditRecord` 字段）的前提下，验证 Claude Code 是否可以通过当前网关接口完成最小端到端调用（FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E，限 E1–E9 共 9 项）。**

本轮允许触达的路径**只**有：

1. `cmd/flagship-coder-gateway/**`
2. `internal/flagship/gateway/**`
3. `internal/flagship/coder/**`
4. `internal/flagship/provider/**`
5. `tests/flagship-gateway/**`
6. `examples/flagship-gateway/**`
7. `docs/FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E.md`
8. `LONGMA_TASK_ANCHOR.json`

> 注：本轮**不再允许**写入 `schemas/flagship-gateway/**` 与既有 `docs/FLAGSHIP_CODER_GATEWAY_P0_DEV.md` / `docs/FLAGSHIP_CODER_GATEWAY_P0_FREEZE_REPORT.md`（Gateway P0 已冻结）。这是 Gateway P1 相对 Gateway P0 的主接口契约收紧。

非目标（本轮严禁触达）：

- 不接真实 Claude / OpenAI / Gemini / Qwen API
- 不读取任何真实 API key
- 不写 `.env` / 不改 `.env.*` / 不读 `secrets/`
- 不做真实 API Key 管理（KMS / vault / 任何外部 secret store）
- 不做鉴权 / 限流 / TLS
- 不做 streaming（SSE / chunked）
- 不做 tool_use / tool_result（function calling）
- 不做 citation / cache_control / thinking / images / documents 等 Anthropic 高级能力
- 不做真实 Claude Code 自动替换配置（不动用户机器上的 Claude Code 设置）
- 不引入任何新依赖（不 `go get` / 不 `go mod tidy` / 不升级既有依赖）
- 不修改 `go.mod` / `go.sum`
- 不修改现有 Core /shear 接口
- 不修改 `policies/manifest.json`
- 不修改 `openapi.yaml`
- 不修改 `sdk/**`
- 不修改 `Dockerfile` / `docker-compose.yml`
- 不修改 `README.md` / `README_CN.md` / `SUPPORT.md`
- 不修改 P0/P1 已冻结推理内核源码 / 服务 / 测试 / schema / 示例 / 三份 P0/P1 docs
- 不修改 Gateway P0 已冻结 schemas / 既有 dev / freeze 报告
- 不接入 Core ledger / Core signature
- 不改变 P0/P1 主契约 与 Gateway P0 主接口
- 不在本轮纳入 E1–E9 之外的任何能力 / 扩展 / 重构
- 不把龙码常数作为 Core 版正式能力发布

## 2. 授权半径

允许写入的文件 = 第 1 节八条路径，且**仅这八条**。其它一切只允许 Read / ls / grep（用于验证）；Go runtime 可通过 import 方式调用旗舰版 reasoner 包（read-only 引用，不改源码），**不允许写**。

落入 `allowed_files` 内的合规子能力实现范围（限 Gateway P1 Claude E2E 9 项 E1–E9）：

- **E1** 编写 Claude Code 接入本地网关的运行说明（仅指引，不动用户机器配置）
- **E2** 验证 `GET /v1/models`
- **E3** 验证 `POST /v1/messages/count_tokens`
- **E4** 验证 `POST /v1/messages` 三态（YES / HOLD / NO）
- **E5** 验证 Claude Code 需要的请求头（anthropic-version / anthropic-beta / authorization / x-api-key / content-type / user-agent / x-app 等）可被接收并忽略
- **E6** 验证 HOLD / NO 统一 HTTP 200 且 `body.verdict` 表达三态
- **E7** 补充最小 e2e 脚本或测试样例
- **E8** 输出 Claude Code 联调记录文档
- **E9** 保持 Mock Provider，不接真实 Provider

E1–E9 之外的任何能力 / 扩展 / 重构都属于 AP1（目标漂移），立即 HOLD。

## 3. 禁止边界

下列路径**禁止**修改：

- `internal/{api,engine,errors,ledger,policy,schema,signature}/**`（Core 引擎）
- `internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`（**P0/P1 已冻结的旗舰版推理内核源码**）
- `cmd/{hash-policy,server,validate-policy,verify-ledger}/**`（Core 命令行）
- `cmd/flagship-server/**`（**P0/P1 已冻结的旗舰版 reasoner 服务**）
- `schemas/*.json`（顶层 schema）
- `schemas/flagship/**`（**P0 已冻结**）
- `schemas/flagship-gateway/**`（**Gateway P0 已冻结**）
- `sdk/**`
- `policies/**`（含 `policies/manifest.json`）
- `integrations/**`
- `docs/{AGENT_TOOL_GATE_GUIDE,INTEGRATION_GUIDE,P1_RELEASE_CHECKLIST,POLICY_PACK_GUIDE,PRODUCT_MATRIX,WHITEPAPER}.md`（既有 Core 产品文档）
- `docs/FLAGSHIP_P0_DEV.md`、`docs/FLAGSHIP_P0_FREEZE_REPORT.md`、`docs/FLAGSHIP_P1_HARDENING_REPORT.md`（**P0/P1 三份已冻结的旗舰版 dev / freeze / hardening 报告**）
- `docs/FLAGSHIP_CODER_GATEWAY_P0_DEV.md`、`docs/FLAGSHIP_CODER_GATEWAY_P0_FREEZE_REPORT.md`（**Gateway P0 两份已冻结的 dev / freeze 报告**）
- `tests/*.go`（顶层 Core 测试）
- `tests/flagship/**`（**P0/P1 已冻结的旗舰版测试**）
- `ledger/**`
- `examples/*.json`（顶层既有示例）
- `examples/flagship/**`（**P0 已冻结**）
- `README.md`、`README_CN.md`、`SUPPORT.md`
- `openapi.yaml`
- `go.mod`、`go.sum`
- `Dockerfile`、`docker-compose.yml`
- `AGENTS.md`、`CLAUDE.md`、`LONGMA_SOFT_GUARD.md`、`.gitignore`

下列动作**禁止**自行执行（必须用户当轮显式指令）：

- `git commit` / `git push` / `git tag` / 任何 release 动作
- 改动 CI、外部依赖、镜像仓库、域名、密钥
- 安装新包、`go get`、`go mod tidy` 引入新包、改动构建产物、引入真实 LLM provider SDK 或 jsonschema 等新库
- 接入真实 LLM、调用任何外部模型 API
- 读取任何真实 API key（含 `.env` / `.env.*` / `secrets/`）
- 对接 Core ledger / Core signature
- 加鉴权 / 限流 / TLS
- 实现 Anthropic Messages 的 streaming / tool_use / citation / cache_control 等高级能力
- 改变 P0/P1 主契约 或 Gateway P0 主接口
- 修改用户机器上的真实 Claude Code 配置
- 在本轮纳入 E1–E9 之外的能力
- 删除或重写禁止路径下的任何已有文件

## 4. 治理的五种反模式

### AP1 目标漂移
**定义**：在执行过程中悄悄扩展、迁移或替换本轮目标。
**典型表现**：把"Claude E2E P1 验证（E1–E9）"扩张成"完整龙码 AIOS / 多租户 SaaS / 规则自动生成 / 知识资产化"；或在本轮里顺手接真实 LLM、读真实 API key、做鉴权 / 限流 / TLS、实现 streaming / tool_use / citation / cache_control；或回头改 P0/P1 已冻结推理内核源码、Gateway P0 已冻结 schemas / docs；或顺手把网关能力以 policy pack 形式塞进 Core；或加 E1–E9 之外的能力。
**守则**：任何越出第 1 节单一目标或 E1–E9 范围的动作必须先停下，把"我想做 X，是否授权"显式问出来。

### AP2 行动越权
**定义**：未经用户当轮显式指令对禁止路径或外部系统执行写动作。
**典型表现**：自行 `git commit` / `git push` / `git tag`、改 CI、装包、改 Core `/shear` 接口、改 `policies/manifest.json`、接入真实 LLM、对 `.env` / `secrets/` 读取、写入 P0/P1 已冻结推理内核源码 / 服务 / 测试 / schema / 示例 / docs、写入 Gateway P0 已冻结 schemas / dev / freeze 报告、改 `cmd/flagship-server`、改 `go.mod` / `go.sum`、修改用户机器上的 Claude Code 配置。
**守则**：写动作前先把目标路径与第 3 节比对一遍；越界则停。

### AP3 路径漂移
**定义**：对真实不存在的目录或文件进行操作；路径凭空臆造。
**典型表现**：把代码写到 `internal/longma/`（不存在）；引用 `cmd/longma-coder/`（不存在）；把 dev 文档写到 `docs/FLAGSHIP-CODER-GATEWAY-P1-CLAUDE-E2E.md`（连字符错误，正确文件名是 `docs/FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E.md`）；或把 P1 的代码塞进 `schemas/flagship-gateway/`（已冻结，不允许）。
**守则**：写入前用 Read / `ls` / `find` 在真实仓库验证父目录存在；不依赖记忆。本轮 `internal/flagship/{gateway,coder,provider}`、`cmd/flagship-coder-gateway`、`tests/flagship-gateway`、`examples/flagship-gateway` 均已存在（Gateway P0 已落地），不得自行新建错误路径或重命名既有目录。

### AP4 幻觉判断
**定义**：凭空假设接口签名、函数行为、模块结构、字段名、版本号、文档原文。
**典型表现**：声称 `internal/flagship/reasoner` 提供某函数而该函数从未存在；引用 `Anthropic Messages` 不存在的字段或自行扩展高级字段；参照训练数据中的"五行 / 龙码常数 / 三态推理 / Anthropic Messages 形状 / Claude Code 行为"印象写代码或文档；自行改矩阵数值、阈值或 verdict 大小写。
**守则**：任何主张要能给出**文件路径 + 行号**；不得引用未读过的代码或文档；P0/P1 主契约（矩阵、`λ μ T1 T2`、风险因子、verdict 大写、字段名、reason_code）以 `v0.4.1-flagship-p1-hardening` 实际仓库为准；Gateway P0 主接口（请求 / 响应 JSON 形状、`stop_reason` 取值、HOLD/NO 一律 200、`GatewayAuditRecord` 字段）以 `v0.5.0-flagship-coder-gateway-p0` 实际仓库为准；Claude Code 实测期待以真实运行结果为准，不得凭印象推断。

### AP5 嘴炮完成
**定义**：声称完成而无可验证证据。
**典型表现**：只说"已完成"，不给 diff、不给文件清单、不给 `go build` / `go vet` / `go test` 输出，不给 Claude Code 端到端实测记录（curl / shell / 截图描述），不报告 E1–E9 各项状态，不报告 HOLD 项。
**守则**：完成必须给出第 6 节的全部证据；少一项视为未完成。

## 5. HOLD 条件

出现以下任一情况，立即 HOLD 并向用户报告，**不要**自行决策推进：

1. 拟动作目标超出第 1 节单一目标（例如要做完整龙码 AIOS / 多租户 / 规则自动生成 / 知识资产化）。
2. 拟动作要在 Gateway P1 Claude E2E 中纳入 E1–E9 之外的任何能力 / 扩展 / 重构（含 streaming / tool_use / citation / cache_control / 鉴权 / 限流 / TLS / 真实 LLM）。
3. 拟修改文件不在第 2 节授权半径。
4. 拟修改文件落在第 3 节禁止路径（含 P0/P1 已冻结的推理内核源码 / 服务 / 测试 / schema / 示例 / docs，以及 Gateway P0 已冻结 schemas / dev / freeze 报告）。
5. 拟改动会改变 P0/P1 主契约（矩阵数值、`λ μ T1 T2`、风险因子、verdict 大写、字段名）或 Gateway P0 主接口（请求 / 响应 JSON 形状、`stop_reason` 取值、HOLD/NO 一律 200、`GatewayAuditRecord` 字段）。
6. 需要新增产品级能力、policy pack 或动 Core 引擎 / Core 对外契约。
7. 需要 `git commit` / `push` / `tag` / release，但用户当轮没有显式指令。
8. 需要接入真实 LLM 或调用外部模型 API、读取真实 API key / `.env` / `secrets/`。
9. 需要新增依赖（`go get` / `go mod tidy` 引入新包，含真实 LLM provider SDK / jsonschema 等）或升级既有依赖。
10. 需要对接 Core ledger / Core signature。
11. 需要做鉴权 / 限流 / TLS。
12. 需要修改用户机器上的真实 Claude Code 配置。
13. 无法在不修改 P0/P1 已冻结源码或 Gateway P0 已冻结契约的前提下完成本轮任务（说明任务设计本身越界）。
14. 无法在真实仓库中验证某个被引用的文件 / 接口 / 字段 / 路径。
15. 完成证据缺任一项。

## 6. 完成证据（缺一不可）

完成时必须输出：

1. **changed files** — 本轮新增 / 修改的全部文件清单及用途。
2. **git diff --stat** — 实际触达哪些文件、增删行数（含 untracked 列举）。
3. **检查结果** — 至少 `go build ./internal/flagship/... ./cmd/flagship-server/... ./cmd/flagship-coder-gateway/...`、`go vet ./internal/flagship/... ./cmd/flagship-coder-gateway/... ./tests/flagship-gateway/...`、`go test ./tests/flagship-gateway/... ./cmd/flagship-coder-gateway/...` 与 `go test ./...` 各跑一次的输出（含通过 / 失败列表）。
4. **Claude Code 端到端实测记录** — curl / shell 输出 / 关键响应片段，至少覆盖 `/v1/models`、`/v1/messages/count_tokens`、`/v1/messages` 三个端点，与 YES / HOLD / NO 三态。
5. **P0/P1 + Gateway P0 主契约检查** — 实测 P0/P1 守门测试（`TestDefaultMatrixFixed` 等）继续通过；Gateway P0 主接口（请求 / 响应 JSON 形状 / `stop_reason` / HOLD-NO-200 / `GatewayAuditRecord` 字段）未变；P0/P1 与 Gateway P0 各 freeze / dev 报告对应的源码 / schema / docs 均未被改动。
6. **边界检查** — 是否触碰禁止路径（含 P0/P1 已冻结的推理内核源码 / 服务 / 测试 / schema / 示例 / docs，以及 Gateway P0 已冻结 schemas / dev / freeze 报告）、是否修改 `policies/manifest.json`、是否动 Core 引擎代码、是否动 Core `/shear` 接口、是否接入真实 LLM、是否读取真实 API key、是否引入新依赖、是否动 `go.mod` / `go.sum`、是否在 E1–E9 之外做了能力扩展。
7. **E1–E9 状态** — 每项的实现状态（已实现 / 未实现 / 部分实现）+ 触达文件 + 对应实测证据。
8. **风险与 HOLD 项** — 已知风险、待澄清项、需要用户决策的事项；特别是 Claude Code 实测中暴露的 P0 不足（如真实期待 streaming / tool_use 等）。

## 7. 与 Core 产品边界的关系

熵剪旗舰版编程大脑网关 P1 Claude Code 端到端适配验证只在闭源开发分支上演化，**不写入** `policies/`、**不出现在** `policies/manifest.json`、**不被** `tests/policy_pack_test.go` 校验、**不出现在** `openapi.yaml` 与 `sdk/`、**不在** README / README_CN / SUPPORT / 既有 `docs/` 里被宣传为 Core 能力。

未来若把本网关与 Core 拼装为完整 Pro / Flagship 商用产品，将是另一轮独立任务，并发生在 Core 范围之外（独立模块或独立分发）。本软约束包不构成对那一轮的承诺，也不锁死其设计。

## 8. 落地实施说明

- `.claude/settings.json` 通过 `permissions.deny` 把第 3 节的禁止路径机器化兜底（其中既有 Core 顶层文件已具名 deny；P0/P1 已冻结的旗舰版推理内核源码 `internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`、其服务 `cmd/flagship-server/**`、其测试 `tests/flagship/**`、其 dev / freeze / hardening 报告自 P1 dev 锚点起被显式 deny；Gateway P0 已冻结的 `schemas/flagship-gateway/**` 与两份 dev / freeze 报告自本锚点起被显式 deny；`schemas/flagship/**` 与 `examples/flagship/**` 沿袭 P1 deny；专门为 `cmd/flagship-coder-gateway/**`、`internal/flagship/{gateway,coder,provider}/**`、`tests/flagship-gateway/**`、`examples/flagship-gateway/**`、`docs/FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E.md` 让出可写空间），但这只是兜底，不是免责：守则依然以本文件与锚点 JSON 为准。
- 本仓库 `.gitignore` 将整个 `.claude/` 目录排除，因此 `.claude/settings.json` 的 deny 配置只在拥有该文件的工作副本上生效，不随 git 分发到其他副本。如果需要让兜底配置在团队 / CI 中生效，需要单独决定是否调整 `.gitignore`，本软约束包不擅自改 `.gitignore`。
- 本轮 `cmd/flagship-coder-gateway`、`internal/flagship/{gateway,coder,provider}`、`tests/flagship-gateway`、`examples/flagship-gateway` 均已存在（Gateway P0 已落地），不得自行新建错误路径或重命名既有目录。`docs/FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E.md` 是新文件，第一次写入由代理自身创建并对用户透明。

## 9. 进入网关 P2 / 旗舰版下一阶段前的强制门槛

进入熵剪旗舰版编程大脑网关 P2 或更后阶段前，**必须先更新** [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json)：

- 将 `target_constant.single_goal` 从"网关 P1 Claude E2E（E1–E9）"替换为下一阶段当轮的真实开发目标。
- 将 `scope_id` / `scope_kind` 替换为下一阶段对应的标识（例如 `flagship-coder-gateway-p2-…`）。
- 将 `authorization_radius.allowed_files` 替换为下一阶段真正需要触达的路径。
- 同步更新 `target_constant.current_round_deliverables_only`、`forbidden_boundaries`、`completion_evidence_required`、`flagship_coder_gateway_p1_scope`（或新增 `flagship_coder_gateway_p2_scope` 等）、`core_product_boundary` 字段。
- 同步更新 `.claude/settings.json` 的 `permissions.deny` 列表，使机器兜底与新一轮授权半径一致。
- 同步更新本文件、`AGENTS.md`、`CLAUDE.md`，使配套文档的"当前轮唯一目标"措辞与新锚点保持一致。

**不得沿用本轮"网关 P1 Claude E2E（E1–E9）"的授权半径**去做下一阶段开发。沿用即视为 AP1（目标漂移）与 AP2（行动越权）同时触发，立即 HOLD。

锚点未完成切换前，AI 编程代理仍受本软约束包约束，只能在本轮 `allowed_files` 内写入；超出该范围只能 Read / ls / grep（含 import P0/P1 已冻结的旗舰版 reasoner 包），**不得对 Core / Pro / Flagship 任何代码进行 Write / Edit / MultiEdit**。
