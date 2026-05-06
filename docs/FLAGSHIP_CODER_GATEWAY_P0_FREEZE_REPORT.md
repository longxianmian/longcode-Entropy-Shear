# 熵剪旗舰版 + LLM 编程大脑网关 P0 冻结报告

> 本文件是熵剪旗舰版 + LLM 编程大脑网关 P0 骨架（G1–G12）的正式冻结归档。
> 本报告**只描述闭源开发分支上的网关 P0 冻结事实**，不属于 Entropy Shear
> Core 对外能力，**不进入** `policies/`、**不修改** `policies/manifest.json`、
> **不被** `openapi.yaml` / `sdk/` 暴露、**不被** `README.md` /
> `README_CN.md` / `SUPPORT.md` 宣传为 Core 能力。详见
> [`../LONGMA_TASK_ANCHOR.json`](../LONGMA_TASK_ANCHOR.json) 与
> [`../LONGMA_SOFT_GUARD.md`](../LONGMA_SOFT_GUARD.md)。

签发日期：2026-05-06 · 主分支：`main` · 仓库：`entropy-shear`

---

## 1. 版本信息

| 项 | 值 |
|---|---|
| **main commit** | `c5a9f9754ca76169a0647282705124fe1f748a4e` |
| **commit subject** | `feat: add flagship coder gateway p0 (#5)` |
| **tag** | `v0.5.0-flagship-coder-gateway-p0`（annotated tag，指向 `c5a9f97`，message: `flagship coder gateway p0`） |
| **PR** | `#5`，源分支 `feat/flagship-coder-gateway-p0` → `main` |
| **合并方式** | squash merge |
| **基线** | `v0.4.1-flagship-p1-hardening`（commit `a6d6ed0`，PR `#3`） |
| **冻结状态** | ✅ **已冻结**：working tree clean、`origin/main = HEAD`、tag 已落、anchor `flagship_coder_gateway_p0_freeze.status = "frozen"` |

`git log --oneline --decorate -7` 实测：

```
c5a9f97 (HEAD -> main, tag: v0.5.0-flagship-coder-gateway-p0, origin/main) feat: add flagship coder gateway p0 (#5)
912df9c docs: add flagship p1 hardening report (#4)
a6d6ed0 (tag: v0.4.1-flagship-p1-hardening) fix: harden flagship p1 reasoner edges (#3)
a184b0a docs: add flagship p0 freeze report (#2)
2ded4fb (tag: v0.4.0-flagship-p0) feat: add flagship p0 tristate reasoner (#1)
399a7c9 chore: add longma soft guard for flagship development
78c093c docs: define core/pro/flagship edition boundaries
```

> Flagship 版本演进：`v0.4.0-flagship-p0`（推理内核） → `v0.4.1-flagship-p1-hardening`（小修加固） → **`v0.5.0-flagship-coder-gateway-p0`**（编程大脑网关骨架）。本轮是网关骨架的首个冻结点，不破坏 P0/P1 已冻结的推理内核。

---

## 2. Coder Gateway P0 已完成能力（G1–G12，12/12）

| ID | 能力 | 落地位置 | 状态 |
|---|---|---|---|
| **G1** | 独立网关服务（不改 Core server、独立端口） | [`cmd/flagship-coder-gateway/main.go`](../cmd/flagship-coder-gateway/main.go) — 默认 `:9091`，env `ENTROPY_SHEAR_FLAGSHIP_GATEWAY_ADDR` | ✅ |
| **G2** | Anthropic Messages 兼容形态最小骨架（messages、role、content blocks、stop_reason、usage 占位） | [`internal/flagship/gateway/types.go`](../internal/flagship/gateway/types.go) | ✅ |
| **G3** | `GET /health`（独立于 Flagship reasoner 与 Core 的 /health） | [`internal/flagship/gateway/server.go`](../internal/flagship/gateway/server.go) `healthHandler` | ✅ |
| **G4** | `GET /v1/models`（占位 model 列表） | `server.go` `modelsHandler` + `governance.go` `ModelID="flagship-coder-mock-1"` | ✅ |
| **G5** | `POST /v1/messages/count_tokens`（占位 token 计数估算） | `server.go` `countTokensHandler` + `governance.go` `approxTokens`（`ceil(JSON 字节长度/4)`） | ✅ |
| **G6** | `POST /v1/messages`（主端点，串联 G7→G8→G9→G10→G11） | `server.go` `messagesHandler` → [`internal/flagship/gateway/governance.go`](../internal/flagship/gateway/governance.go) `runGovernance` | ✅ |
| **G7** | 生成前治理：调用 `internal/flagship/reasoner.Reason()` | `governance.go` 第一段 `reasoner.Reason(preInput)`；非 YES 短路返回（不调 provider） | ✅ |
| **G8** | Mock LLM Provider | [`internal/flagship/provider/mock.go`](../internal/flagship/provider/mock.go) `MockProvider` + [`provider.go`](../internal/flagship/provider/provider.go) `Provider` 接口 | ✅ |
| **G9** | 生成后审查：再次调用 `internal/flagship/reasoner.Reason()` | `governance.go` 第二段 `reasoner.Reason(postInput)` | ✅ |
| **G10** | Claude Code 可识别的 assistant message 响应 | [`internal/flagship/coder/assistant.go`](../internal/flagship/coder/assistant.go) `BuildAssistantContent` + `governance.go` `buildResponse` | ✅ |
| **G11** | `GatewayAuditRecord` 审计结构（仅返回对象，不写 Core ledger、不写 JSONL） | [`internal/flagship/coder/audit.go`](../internal/flagship/coder/audit.go) `GatewayAuditRecord` + `NewGatewayAuditRecord` | ✅ |
| **G12** | 最小测试用例 | [`tests/flagship-gateway/`](../tests/flagship-gateway/) 4 个测试文件 + [`cmd/flagship-coder-gateway/main_test.go`](../cmd/flagship-coder-gateway/main_test.go) 共 28 个 testfunc，全 PASS | ✅ |

**P0/P1 主契约守门**：本轮**未改变** P0/P1 主契约任一字段、矩阵数值、`λ μ T1 T2`、风险因子、verdict 大写、`AlignmentTask` / `PermitToken` / `RejectInstruction` / `AuditRecord` 字段集；`TestDefaultMatrixFixed` 等冻结测试在 `go test ./...` 中继续 PASS。

GD-1 到 GD-12 决策表已在 [`docs/FLAGSHIP_CODER_GATEWAY_P0_DEV.md §3`](FLAGSHIP_CODER_GATEWAY_P0_DEV.md) 列出每条决策的代码落地行号。

---

## 3. 明确未做边界（网关 P0 严守的不变量）

| 边界 | 状态 |
|---|---|
| 不接真实 Claude / OpenAI / Gemini / Qwen API | ✅ 全 mock provider |
| 不读取真实 API key | ✅ 零 `.env` / `.env.*` / `secrets/` 读取 |
| 不写 `.env` / 不改 `.env.*` | ✅ |
| 不开发完整龙码 AIOS | ✅ |
| 不开发知识资产化操作系统 | ✅ |
| 不开发能力工具系统 | ✅ |
| 不接 Core ledger（`internal/ledger`、`ledger/`） | ✅ |
| 不接 Core signature（`internal/signature`） | ✅ |
| 不做鉴权 | ✅ HTTP 裸服务，anthropic-version / authorization / x-api-key header 允许出现但不校验 |
| 不做限流 | ✅ |
| 不做 TLS | ✅ |
| 不引入任何新依赖 | ✅ 仅标准库；`go.mod` / `go.sum` 零修改 |
| 不修改 Core `/shear`（`internal/api`、`internal/engine`、`internal/schema`、`internal/policy`、`internal/errors`） | ✅ |
| 不修改 P0/P1 已冻结推理内核（`internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`、`cmd/flagship-server/**`、`tests/flagship/**`、`schemas/flagship/**`、`examples/flagship/**`、3 份 P0/P1 docs） | ✅ |
| 不修改 `policies/manifest.json` | ✅ `policies/**` 整树未触 |
| 不修改 `go.mod` / `go.sum` | ✅ |
| 不修改 `Dockerfile` / `docker-compose.yml` | ✅ |
| 不修改 `openapi.yaml` | ✅ |
| 不修改 `sdk/**` | ✅ |
| 不修改 `README.md` / `README_CN.md` / `SUPPORT.md` | ✅ |
| 不实现 Anthropic 高级能力（streaming / tool_use / citation / cache_control / thinking / images / documents） | ✅ schema enum 闭合 |
| 不在本轮纳入 G1–G12 之外的能力 | ✅ |

---

## 4. 测试与验收结果（在 `main @ c5a9f97` 实测）

```
$ git status --short
(空)                                                              ✓ 工作树干净

$ go build ./internal/flagship/... ./cmd/flagship-server/... ./cmd/flagship-coder-gateway/...
(无输出)                                                          ✓

$ go vet ./internal/flagship/... ./cmd/flagship-coder-gateway/... ./tests/flagship-gateway/...
(无输出)                                                          ✓

$ go test ./tests/flagship-gateway/... ./cmd/flagship-coder-gateway/...
ok  entropy-shear/tests/flagship-gateway       0.926s
ok  entropy-shear/cmd/flagship-coder-gateway   1.433s              ✓

$ go test ./...
ok  entropy-shear/cmd/flagship-coder-gateway (cached)
ok  entropy-shear/cmd/flagship-server        (cached)  ← P0/P1 守门 PASS
ok  entropy-shear/tests                      (cached)  ← Core 回归 PASS
ok  entropy-shear/tests/flagship             (cached)  ← P0/P1 内核测试 PASS
ok  entropy-shear/tests/flagship-gateway     (cached)               ✓

$ python3 -c "json.load × 10 flagship-gateway files"
flagship-gateway JSON: OK                                          ✓
```

新增 28 个 testfunc 全 PASS：

| 套件 | 数量 | 列表 |
|---|---|---|
| `cmd/flagship-coder-gateway/main_test.go` | 8 | `TestMessagesHandlerRejectsGET` / `BadJSON` / `DisallowsUnknownFields` / `IgnoresAnthropicHeaders`、`TestHealthHandlerGET` / `RejectsPOST`、`TestModelsHandlerRejectsPOST`、`TestCountTokensHandlerRejectsGET` |
| `tests/flagship-gateway/adapter_test.go` | 8 | `Basic` / `NoSystemDescriptionMissing` / `PermissionDeniedHard` / `PostAddsCandidate` / `PostForceReject` / `PostPolicyViolation` / `PreDefaultProducesYes` / `PreEmptySystemLandsInHold` |
| `tests/flagship-gateway/mock_provider_test.go` | 4 | `Name` / `Determinism` / `TextShape` / `DifferentRequestsDiffer` |
| `tests/flagship-gateway/count_tokens_test.go` | 4 | `EmptyMessages` / `NonZero` / `LongInputScalesUp` / `RejectsBadJSON` |
| `tests/flagship-gateway/gateway_e2e_test.go` | 6 | `MessagesYesEndToEnd` / `HoldEndToEnd` / `NoEndToEnd` / `ModelsListShape` / `HealthReachable` / `AuditPreContainsRequestID` |

JSON 校验范围：
`schemas/flagship-gateway/{messages-request,messages-response,count-tokens-request,count-tokens-response,models-response}.schema.json` + `examples/flagship-gateway/{messages-yes-request,messages-hold-request,messages-no-request,count-tokens-request,models-response}.json` 共 10 份 → `flagship-gateway JSON: OK`。

---

## 5. Core 与冻结内核隔离结论

**完全隔离，未触碰任何 Core 资产 / 任何 P0/P1 已冻结推理内核源码。**

import 图扫描（`grep -rn` `entropy-shear/internal/{api,engine,policy,schema,ledger,signature}` 在 `internal/flagship/{gateway,provider,coder}/` + `cmd/flagship-coder-gateway/` + `tests/flagship-gateway/`）：**零命中**。网关代码仅依赖：

- Go 标准库
- `entropy-shear/internal/flagship/{coder,provider,reasoner}`（其中 reasoner 为 P0/P1 已冻结，仅 runtime import，零源码修改）

写入审计：
- `internal/{api,engine,errors,ledger,policy,schema,signature}/**`、`cmd/{server,verify-ledger,validate-policy,hash-policy}/**`、Core `/shear` —— 未触
- `internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`（**P0/P1 已冻结**）、`cmd/flagship-server/**`、`tests/flagship/**`、`schemas/flagship/**`、`examples/flagship/**` —— 未触
- `docs/FLAGSHIP_P0_DEV.md` / `FLAGSHIP_P0_FREEZE_REPORT.md` / `FLAGSHIP_P1_HARDENING_REPORT.md` —— 未触
- `policies/**`、`integrations/**`、`sdk/**`、`openapi.yaml`、顶层 `examples/*.json` 与顶层 `tests/*.go`、`README.md` / `README_CN.md` / `SUPPORT.md`、`go.mod` / `go.sum`、`Dockerfile` / `docker-compose.yml`、`.gitignore`、`AGENTS.md` / `CLAUDE.md` / `LONGMA_SOFT_GUARD.md` —— 全部未触

锚点 `flagship_p0_freeze` / `flagship_p0_dev_progress` / `flagship_p1_hardening_freeze` 历史指针保留不动。本轮网关 P0 仅在网关自身的 6 个新增子目录 + 1 份新 dev 文档 + 1 个锚点字段（本报告同时新增的 `flagship_coder_gateway_p0_freeze` 块）内活动。

---

## 6. Gateway P1 HOLD 项（冻结期不动；进入下一轮须先做新一轮锚点滚动）

| # | HOLD 项 | 性质 | 备注 |
|---|---|---|---|
| GH-1 | 是否接真实 LLM Provider（Claude / OpenAI / Gemini / Qwen / DeepSeek / 通义 等） | 战略决策 | 需先决定支持哪家 + 引入对应 SDK（违反 GD"不引新依赖"，必须先做锚点滚动） |
| GH-2 | 是否做 API Key 管理（`.env` / KMS / vault 集成） | 战略决策 | 与 GH-1 配套；P0 显式不读真实 key |
| GH-3 | 是否做鉴权 / 限流 / TLS（生产化） | 生产化决策 | P0 是裸服务；进入生产前必做 |
| GH-4 | 是否做 streaming（SSE / chunked） | Anthropic 兼容深化 | Claude Code 实际依赖 streaming；P0 仅最小骨架 |
| GH-5 | 是否做 tool_use / tool_result（function calling） | Anthropic 兼容深化 | Claude Code 重度依赖；与 streaming 同优先级 |
| GH-6 | 是否做真实 Claude Code 端到端适配测试（`ANTHROPIC_BASE_URL` 指向本网关） | 验证手段 | 当前只有 unit + e2e 内测，未与真 Claude Code 互联 |
| GH-7 | 是否接 Core ledger / signature 形成证据链 | 战略决策 | P0 仅返回对象；考虑长期可审计性时启动 |
| GH-8 | 是否沉淀 Gateway 审计链（持久化 GatewayAuditRecord） | 工程决策 | 与 GH-7 关联；可独立选 JSONL 落盘 / 接 ledger / 接外部存储 |
| GH-9 | 是否进入"可替代 Claude Code LLM backend 的本地代理模式" | 产品决策 | GH-1 + GH-4 + GH-5 + GH-6 完成后才有意义 |

> 所有 GH 项都属于网关 P1 / P2+。本冻结期不应动；若要解任一项，必须先新建 [`../LONGMA_TASK_ANCHOR.json`](../LONGMA_TASK_ANCHOR.json) 的下一轮锚点（更新 `single_goal` / `scope_id` / `authorization_radius.allowed_files` / `forbidden_boundaries` / `flagship_coder_gateway_p1_scope`），并同步配套软约束文档（`LONGMA_SOFT_GUARD.md` / `CLAUDE.md` / `AGENTS.md` / `.claude/settings.json`）。**在新锚点未完成前，本仓库不得对任何 Core / Pro / Flagship 代码做 Write / Edit / MultiEdit。**

---

## 7. 下一步建议

**最小下一步**：由用户决定是否启动**网关 P1 锚点滚动**——只更新 [`../LONGMA_TASK_ANCHOR.json`](../LONGMA_TASK_ANCHOR.json) 的 `single_goal` / `scope_id` / `authorization_radius.allowed_files` / `flagship_coder_gateway_p1_scope`，并同步 `LONGMA_SOFT_GUARD.md` / `CLAUDE.md` / `AGENTS.md` / `.claude/settings.json`。在新锚点未完成前，本仓库一律保持只读。最优先解 GH 项的候选是 **GH-6**（与真实 Claude Code 端到端互联），它不需要引入 LLM SDK 也不需要鉴权，验证既有 P0 骨架的 Anthropic 兼容性是否真的满足 Claude Code 的期待，再决定 GH-1 / GH-4 / GH-5 的优先级。
