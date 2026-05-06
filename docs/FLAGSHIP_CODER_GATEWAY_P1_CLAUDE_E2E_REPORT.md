# 熵剪旗舰版编程大脑网关 P1 Claude Code E2E 验证报告

> 本文件是熵剪旗舰版编程大脑网关 P1 Claude Code 端到端适配验证（E1–E9）的
> 正式冻结归档。本报告**只描述闭源开发分支上的网关 P1 Claude E2E 冻结事
> 实**，不属于 Entropy Shear Core 对外能力，**不进入** `policies/`、**不修
> 改** `policies/manifest.json`、**不被** `openapi.yaml` / `sdk/` 暴露、
> **不被** `README.md` / `README_CN.md` / `SUPPORT.md` 宣传为 Core 能力。详
> 见 [`../LONGMA_TASK_ANCHOR.json`](../LONGMA_TASK_ANCHOR.json) 与
> [`../LONGMA_SOFT_GUARD.md`](../LONGMA_SOFT_GUARD.md)。

签发日期：2026-05-06 · 主分支：`main` · 仓库：`entropy-shear`

---

## 1. 版本信息

| 项 | 值 |
|---|---|
| **main commit** | `1b6f170c9ab0a1886246fe9e7085b5bd613e565c` |
| **commit subject** | `test: add flagship coder gateway claude e2e validation (#7)` |
| **tag** | `v0.5.1-flagship-coder-gateway-claude-e2e`（annotated tag，指向 `1b6f170`，message: `flagship coder gateway claude e2e validation`） |
| **PR** | `#7`，源分支 `feat/flagship-coder-gateway-p1-claude-e2e` → `main` |
| **合并方式** | squash merge |
| **基线** | `v0.5.0-flagship-coder-gateway-p0`（commit `c5a9f97`，PR `#5`） |
| **冻结状态** | ✅ **已冻结**：working tree clean、`origin/main = HEAD`、tag 已落、anchor `flagship_coder_gateway_p1_claude_e2e_freeze.status = "frozen"` |

`git log --oneline --decorate -8` 实测：

```
1b6f170 (HEAD -> main, tag: v0.5.1-flagship-coder-gateway-claude-e2e, origin/main) test: add flagship coder gateway claude e2e validation (#7)
7befec8 docs: add flagship coder gateway p0 freeze report (#6)
c5a9f97 (tag: v0.5.0-flagship-coder-gateway-p0) feat: add flagship coder gateway p0 (#5)
912df9c docs: add flagship p1 hardening report (#4)
a6d6ed0 (tag: v0.4.1-flagship-p1-hardening) fix: harden flagship p1 reasoner edges (#3)
a184b0a docs: add flagship p0 freeze report (#2)
2ded4fb (tag: v0.4.0-flagship-p0) feat: add flagship p0 tristate reasoner (#1)
399a7c9 chore: add longma soft guard for flagship development
```

> Flagship 演进：`v0.4.0-flagship-p0` → `v0.4.1-flagship-p1-hardening` → `v0.5.0-flagship-coder-gateway-p0` → **`v0.5.1-flagship-coder-gateway-claude-e2e`**。本轮是网关骨架接入 Claude Code 验证的首个冻结点；commit type 选 `test:` 反映本轮主要交付是测试 + 验证文档，未改动 Gateway P0 主接口。

---

## 2. P1 Claude Code E2E 已完成验证项（E1–E9，9/9）

| ID | 验证项 | 落地位置 | 状态 |
|---|---|---|---|
| **E1** | Claude Code 接入本地网关运行说明（含明确"不动用户机器配置") | [`docs/FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E.md`](FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E.md) §2（启动）+ §3（curl 样例 4 个端点 + 三态）+ §4（用户侧手动 export `ANTHROPIC_BASE_URL`，本轮代理不动 `~/.claude` / Claude Code settings / shell rc） | ✅ |
| **E2** | 验证 `GET /v1/models` | [`tests/flagship-gateway/claude_e2e_test.go`](../tests/flagship-gateway/claude_e2e_test.go) `TestClaudeE2EModelsListShape`（带 7 个 Claude 头）+ `TestClaudeE2EModelsAcceptsHeaders` + `TestClaudeE2EModelsWithoutHeaders`（无头降级） | ✅ |
| **E3** | 验证 `POST /v1/messages/count_tokens` | `claude_e2e_test.go::TestClaudeE2ECountTokensWithHeaders`（strip max_tokens 以符合 P0 count-tokens schema，带全 header，input_tokens > 0） | ✅ |
| **E4** | 验证 `POST /v1/messages` 三态 | `TestClaudeE2EMessagesYes`（Claude 风格 + system + 多轮 messages + metadata → YES，model 回显 `claude-sonnet-4-5`，post audit 存在）+ `TestClaudeE2EMessagesHold` + `TestClaudeE2EMessagesNo` + Claude 风格 YES 样例 [`examples/flagship-gateway/messages-claude-code-style-request.json`](../examples/flagship-gateway/messages-claude-code-style-request.json) | ✅ |
| **E5** | 验证 anthropic-version / anthropic-beta / authorization / x-api-key / content-type / user-agent / x-app 等请求头被接收并忽略（不导致 4xx） | `TestClaudeE2EEachHeaderAccepted`（7 个 header 各自单独 200，子测试 7/7）+ `TestClaudeE2EUnknownFutureHeadersAccepted`（未来未知 header 含 `x-anthropic-experimental` / `x-claude-code-session-id` 也 200） | ✅ |
| **E6** | HOLD / NO 统一 HTTP 200 + body.verdict（GD-10）+ stop_reason（GD-3） | `TestClaudeE2EVerdictMatrixAllReturn200`（YES/HOLD/NO 三态都断言 HTTP 200 + body.verdict + body.stop_reason，子测试 3/3）；HOLD/NO 各自单测 status assertion | ✅ |
| **E7** | 最小 e2e 脚本 / 测试样例 | `tests/flagship-gateway/claude_e2e_test.go`（10 testfunc / 17 子用例）+ `examples/flagship-gateway/messages-claude-code-style-request.json`（28 行 Claude 风格 YES 样例） | ✅ |
| **E8** | Claude Code 联调记录文档 | [`docs/FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E.md`](FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E.md) §5（已验证接口表 + 已验证 header 列表 + 已验证三态矩阵）+ §6（GH-1 ～ GH-9 P2 决策项） | ✅ |
| **E9** | 保持 Mock Provider 不接真实 Provider | [`internal/flagship/provider/{provider,mock}.go`](../internal/flagship/provider/) **零修改**；未引入任何 LLM SDK；`go.mod` / `go.sum` 零修改 | ✅ |

**P0/P1 + Gateway P0 主契约守门**：本轮**未改变** P0/P1 主契约任一字段、矩阵数值、`λ μ T1 T2`、风险因子、verdict 大写、`AlignmentTask` / `PermitToken` / `RejectInstruction` / `AuditRecord` 字段集；**未改变** Gateway P0 主接口（请求 / 响应 JSON 形状、`stop_reason` 取值、HOLD/NO 一律 200、`GatewayAuditRecord` 字段）。`TestDefaultMatrixFixed` + Gateway P0 既有 e2e 测试在 `go test ./...` 中继续 PASS。

---

## 3. 明确未做边界（P1 Claude E2E 严守的不变量）

| 边界 | 状态 |
|---|---|
| 未接真实 Claude / OpenAI / Gemini / Qwen API | ✅ 全 mock provider |
| 未读取真实 API key | ✅ `authorization` / `x-api-key` 仅作为 header 接收并忽略，不校验、不入库 |
| 未读取 `.env` / `.env.*` / `secrets/` | ✅ 零文件读取 |
| 未做真实 Claude Code 自动替换配置 | ✅ 软约束 §3 明确禁止；本轮代理零修改 `~/.claude/` / Claude Code settings / 用户 shell rc |
| 未做 streaming（SSE / chunked） | ✅ |
| 未做 tool_use / tool_result（function calling） | ✅ |
| 未做 citation / cache_control / thinking / images / documents | ✅ Gateway P0 schema enum 闭合 |
| 未做鉴权 / 限流 / TLS | ✅ HTTP 裸服务 |
| 未引入任何新依赖 | ✅ 仅标准库；`go.mod` / `go.sum` 零修改 |
| 未修改 Core `/shear`（`internal/api`、`internal/engine`、`internal/schema`、`internal/policy`、`internal/errors`） | ✅ |
| 未修改 P0/P1 已冻结推理内核（`internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`、`cmd/flagship-server/**`、`tests/flagship/**`、`schemas/flagship/**`、`examples/flagship/**`、3 份 P0/P1 docs） | ✅ |
| 未修改 Gateway P0 已冻结 schema（`schemas/flagship-gateway/**`） | ✅ 5 份 schema 整体未触 |
| 未修改 Gateway P0 已冻结 docs（`docs/FLAGSHIP_CODER_GATEWAY_P0_DEV.md`、`docs/FLAGSHIP_CODER_GATEWAY_P0_FREEZE_REPORT.md`） | ✅ |
| 未修改 `policies/manifest.json` | ✅ `policies/**` 整树未触 |
| 未修改 `go.mod` / `go.sum` | ✅ |
| 未修改 `Dockerfile` / `docker-compose.yml` | ✅ |
| 未修改 `openapi.yaml` | ✅ |
| 未修改 `sdk/**` | ✅ |
| 未修改 `README.md` / `README_CN.md` / `SUPPORT.md` | ✅ |
| 未在 E1–E9 之外做任何能力扩展 | ✅ |

---

## 4. 测试与验收结果（在 `main @ 1b6f170` 实测）

```
$ git status --short
(空)                                                              ✓ 工作树干净

$ go build ./internal/flagship/... ./cmd/flagship-server/... ./cmd/flagship-coder-gateway/...
(无输出)                                                          ✓

$ go vet ./internal/flagship/... ./cmd/flagship-coder-gateway/... ./tests/flagship-gateway/...
(无输出)                                                          ✓

$ go test ./tests/flagship-gateway/... ./cmd/flagship-coder-gateway/...
ok  entropy-shear/tests/flagship-gateway       1.322s
ok  entropy-shear/cmd/flagship-coder-gateway   (cached)            ✓

$ go test ./...
ok  entropy-shear/cmd/flagship-coder-gateway (cached)
ok  entropy-shear/cmd/flagship-server        (cached)  ← P0/P1 守门 PASS
ok  entropy-shear/tests                      (cached)  ← Core 回归 PASS
ok  entropy-shear/tests/flagship             (cached)  ← P0/P1 内核测试 PASS
ok  entropy-shear/tests/flagship-gateway     (cached)  ← Gateway P0+P1 全绿
                                                                  ✓

$ python3 -c "json.load × 1 file"
claude e2e JSON: OK                                                ✓
```

新增 10 个 testfunc / 17 子用例全 PASS：

| 测试 | 覆盖 |
|---|---|
| `TestClaudeE2EModelsListShape` | E2 — `GET /v1/models` 带 7 个 Claude 头，data 含 `flagship-coder-mock-1` |
| `TestClaudeE2EModelsAcceptsHeaders` | E2 + E5 — Claude 头 + content-type 验证 |
| `TestClaudeE2EModelsWithoutHeaders` | E2 边界 — 不带任何头降级也 200 |
| `TestClaudeE2ECountTokensWithHeaders` | E3 — strip max_tokens + 全 header，input_tokens > 0 |
| `TestClaudeE2EMessagesYes` | E4 — Claude 风格 → YES + model 回显 + post audit |
| `TestClaudeE2EMessagesHold` | E4 + E6 — HOLD HTTP 200 + post audit nil |
| `TestClaudeE2EMessagesNo` | E4 + E6 — NO HTTP 200 + stop_reason=refusal |
| `TestClaudeE2EEachHeaderAccepted` | E5 — 7 个 header 各自单独 200（子测试 7/7） |
| `TestClaudeE2EUnknownFutureHeadersAccepted` | E5 — 未来未知 header 也 200 |
| `TestClaudeE2EVerdictMatrixAllReturn200` | E6 — YES/HOLD/NO 三态都 HTTP 200（子测试 3/3） |

JSON 校验范围：`examples/flagship-gateway/messages-claude-code-style-request.json` → `claude e2e JSON: OK`。

---

## 5. Core / Reasoner / Gateway P0 schema 隔离结论

**完全隔离，未触碰任何 Core 资产、任何 P0/P1 已冻结推理内核源码、任何 Gateway P0 已冻结 schema 与 dev / freeze 报告。**

import 图扫描（`grep -rn` `entropy-shear/internal/{api,engine,policy,schema,ledger,signature}` 在 `tests/flagship-gateway/`）：**零命中**。新增 `claude_e2e_test.go` 仅依赖：

- Go 标准库（`bytes` / `encoding/json` / `net/http` / `net/http/httptest` / `strings` / `testing`）
- `entropy-shear/internal/flagship/gateway`（Gateway P0 已冻结 + P1 可改子目录，未改）
- `entropy-shear/internal/flagship/provider`（未改）
- 既有 `tests/flagship-gateway/gateway_e2e_test.go` 内的 `loadExample` / `decodeResp` 等 helper（同 package 共享）

写入审计：

- `internal/{api,engine,errors,ledger,policy,schema,signature}/**`、`cmd/{server,verify-ledger,validate-policy,hash-policy}/**`、Core `/shear` —— 未触
- `internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`（**P0/P1 已冻结**）、`cmd/flagship-server/**`、`tests/flagship/**`、`schemas/flagship/**`、`examples/flagship/**` —— 未触
- `schemas/flagship-gateway/**`（**Gateway P0 已冻结**）、`docs/FLAGSHIP_CODER_GATEWAY_P0_DEV.md`、`docs/FLAGSHIP_CODER_GATEWAY_P0_FREEZE_REPORT.md` —— 未触
- `internal/flagship/{gateway,coder,provider}/**`、`cmd/flagship-coder-gateway/**` —— 未触（本轮无 production 代码改动）
- `docs/FLAGSHIP_P0_DEV.md` / `FLAGSHIP_P0_FREEZE_REPORT.md` / `FLAGSHIP_P1_HARDENING_REPORT.md` —— 未触
- `policies/**`、`integrations/**`、`sdk/**`、`openapi.yaml`、顶层 `examples/*.json` 与 `tests/*.go`、`README*` / `SUPPORT.md`、`go.mod` / `go.sum`、`Dockerfile` / `docker-compose.yml`、`.gitignore`、`AGENTS.md` / `CLAUDE.md` / `LONGMA_SOFT_GUARD.md` —— 全部未触
- 用户机器上的真实 Claude Code 配置（`~/.claude/` / Claude Code settings / shell rc） —— 未触

锚点 `flagship_p0_freeze` / `flagship_p1_hardening_freeze` / `flagship_coder_gateway_p0_freeze` 历史指针保留不动。本轮 P1 Claude E2E 仅在 3 个新增产出（claude_e2e_test.go + Claude 风格 example + 本验证 dev doc）+ 1 个本报告 + 1 个锚点字段（本报告同时新增的 `flagship_coder_gateway_p1_claude_e2e_freeze` 块）内活动。

---

## 6. 后续 HOLD 项（冻结期不动；进入下一轮须先做新一轮锚点滚动）

| # | HOLD 项 | 性质 | 备注 |
|---|---|---|---|
| EH-1 | 真实 Claude Code 客户端尚未手动接入验证 | 验证缺口 | 本轮所有 e2e 通过 Go test + curl 完成；尚未启动真实 Claude Code 二进制把 `ANTHROPIC_BASE_URL` 指向本网关并跑一轮 UI 互动。需用户在自己机器上手动操作（本轮代理依据软约束不动用户配置）。 |
| EH-2 | `/v1/models` 真实模型别名映射 | Anthropic 兼容深化 | Claude Code 启动时若校验 model id 是否存在于 `/v1/models.data`（real Anthropic 列表含十几个 model），可能失败；当前仅有 `flagship-coder-mock-1`。 |
| EH-3 | count_tokens 真实 tokenizer 差异 | 精度问题 | 当前 GD-6 占位 `ceil(JSON 字节长度/4)`；中文 / Unicode 输入下 byte 长度大于 token 数的 N 倍。Claude Code 严格依赖 token 预算时可能误判。 |
| EH-4 | streaming 支持（SSE / chunked） | Anthropic 兼容深化 | Claude Code 重度依赖 streaming 体验；P0 一次性返回完整响应 |
| EH-5 | tool_use / tool_result 支持（function calling） | Anthropic 兼容深化 | Claude Code 工具调用功能完全无法工作；当前 schema enum 已闭合 |
| EH-6 | 真实 Provider 接入（Claude / OpenAI / Gemini / Qwen 等） | 战略决策 | mock 候选不是真实回答 |
| EH-7 | API Key 管理（`.env` / KMS / vault 集成） | 战略决策 | 与 EH-6 配套 |
| EH-8 | 鉴权 / 限流 / TLS（生产化） | 生产化决策 | 与 EH-6 / EH-7 配套；进入生产前必做 |
| EH-9 | Gateway 审计链沉淀（持久化 `GatewayAuditRecord`） | 工程决策 | 当前仅返回对象，不写盘 |

> 所有 EH 项都属于 Gateway P2 / P3+。本冻结期不应动；若要解任一项，必须先新建 [`../LONGMA_TASK_ANCHOR.json`](../LONGMA_TASK_ANCHOR.json) 的下一轮锚点（更新 `single_goal` / `scope_id` / `authorization_radius.allowed_files` / `forbidden_boundaries` / `flagship_coder_gateway_p2_scope`），并同步配套软约束文档（`LONGMA_SOFT_GUARD.md` / `CLAUDE.md` / `AGENTS.md` / `.claude/settings.json`）。**在新锚点未完成前，本仓库不得对任何 Core / Pro / Flagship 代码做 Write / Edit / MultiEdit。**

---

## 7. 下一步建议

**最小下一步**：由用户在自己机器上手动 `export ANTHROPIC_BASE_URL=http://localhost:9091` 后启动真实 Claude Code 二进制，跑一轮 UI 互动作为 EH-1 的补强证据；观察哪些字段 / 行为暴露 P0 / P1 不足，作为 Gateway P2 锚点滚动 single_goal 的输入材料。**本仓库在用户提交真实接入观察之前一律保持只读**；任何后续动作（接 EH-2 ～ EH-9 任一项）都必须先做新一轮锚点滚动。
