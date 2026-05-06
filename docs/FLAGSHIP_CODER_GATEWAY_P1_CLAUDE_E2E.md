# 熵剪旗舰版编程大脑网关 P1 — Claude Code 端到端适配验证（开发笔记）

> 本文件只描述 **闭源开发分支** 上的 Claude Code 端到端适配验证细节。它
> **不是** Entropy Shear Core 的对外能力。本网关**不进入** `policies/`、
> **不修改** `policies/manifest.json`、**不被** `openapi.yaml` / `sdk/` 暴露、
> **不被** `README.md` / `README_CN.md` / `SUPPORT.md` 宣传为 Core 能力。详见
> [`../LONGMA_TASK_ANCHOR.json`](../LONGMA_TASK_ANCHOR.json) 与
> [`../LONGMA_SOFT_GUARD.md`](../LONGMA_SOFT_GUARD.md)。

## 1. 目标与边界

在不接真实 LLM、不读取真实 API key、不破坏 Core、不破坏 Flagship P0/P1 推
理内核（`v0.4.1-flagship-p1-hardening` 已冻结）、不改变 Gateway P0 主接口
（`v0.5.0-flagship-coder-gateway-p0` 已冻结）的前提下，验证 Claude Code 是否
可以通过当前网关接口完成最小端到端调用（FLAGSHIP_CODER_GATEWAY_P1_CLAUDE_E2E，
限 E1–E9 共 9 项）。

**显式不做**（再读一遍以防漂移）：

- 不接真实 Claude / OpenAI / Gemini / Qwen / DeepSeek / 通义 / 文心 / 等等 API
- 不读取 `.env` / `.env.*` / `secrets/`，不做真实 API Key 管理
- 不做鉴权 / 限流 / TLS
- 不做 streaming、tool_use、citation、cache_control、thinking、images、documents
- 不做真实 Claude Code 自动替换配置（**不动用户机器上的 `~/.claude` 或 Claude Code 设置**）
- 不引入新依赖；`go.mod` / `go.sum` 不动
- 不修改 P0/P1 已冻结推理内核 / 服务 / 测试 / schema / 示例 / docs
- 不修改 Gateway P0 已冻结的 `schemas/flagship-gateway/**` 与两份 dev / freeze 报告
- 不改变 P0/P1 主契约 与 Gateway P0 主接口（请求 / 响应 JSON 形状、verdict 大写、`stop_reason` 取值、HOLD/NO 一律 200、`GatewayAuditRecord` 字段）

## 2. 启动本地网关

依赖：仅 Go 标准库；`go.mod` / `go.sum` 不动。

```sh
# 默认 :9091
go run ./cmd/flagship-coder-gateway

# 覆盖端口（与 Flagship reasoner :9090 / Core :8080 错开）
ENTROPY_SHEAR_FLAGSHIP_GATEWAY_ADDR=:18091 go run ./cmd/flagship-coder-gateway

# 后台运行（macOS / Linux）
nohup go run ./cmd/flagship-coder-gateway >/tmp/flagship-coder-gateway.log 2>&1 &
```

启动成功后日志：

```
flagship-coder-gateway listening on :9091
```

可用端点：

| 方法 + 路径 | 用途 |
|---|---|
| `GET /health` | 网关存活探针；`{"status":"ok","module":"flagship-coder-gateway"}` |
| `GET /v1/models` | 占位模型列表（`flagship-coder-mock-1`） |
| `POST /v1/messages/count_tokens` | 占位 token 计数（`ceil(JSON 字节长度/4)`） |
| `POST /v1/messages` | 主入口：pre 治理 → mock provider → post 治理 → 包装响应 |

> ⚠️ 本网关**只用 mock provider**，每次都返回固定模板候选 `[mock-candidate sha:<8>] entropy-shear flagship coder gateway P0 placeholder; not from a real LLM.`，不调用任何真实 LLM。

## 3. 用 curl 模拟 Claude Code 请求

以下每个示例都附带 Claude Code 类客户端常用的请求头，验证 P1-E5 的"接收且
忽略"语义。`authorization` 与 `x-api-key` 用占位 fake 值，**不**触达任何真实
密钥。

### 3.1 `GET /v1/models`

```sh
curl -s http://localhost:9091/v1/models \
     -H 'anthropic-version: 2023-06-01' \
     -H 'anthropic-beta: tools-2024-04-04,prompt-caching-2024-07-31' \
     -H 'authorization: Bearer fake-not-validated' \
     -H 'x-api-key: sk-fake-not-validated' \
     -H 'user-agent: claude-cli/0.x (entropy-shear-flagship-coder-gateway-p1-e2e)' \
     -H 'x-app: claude-code' | jq .
```

期待响应：

```json
{
  "data": [
    {
      "id": "flagship-coder-mock-1",
      "type": "model",
      "display_name": "Flagship Coder Mock"
    }
  ]
}
```

### 3.2 `POST /v1/messages/count_tokens`

```sh
curl -s -X POST http://localhost:9091/v1/messages/count_tokens \
     -H 'content-type: application/json' \
     -H 'anthropic-version: 2023-06-01' \
     -H 'authorization: Bearer fake-not-validated' \
     -H 'x-api-key: sk-fake-not-validated' \
     --data @examples/flagship-gateway/count-tokens-request.json | jq .
```

期待响应：`{"input_tokens": <int>}`，整数 ≥ 1。

### 3.3 `POST /v1/messages` — YES（Claude Code 风格请求）

```sh
curl -s -X POST http://localhost:9091/v1/messages \
     -H 'content-type: application/json' \
     -H 'anthropic-version: 2023-06-01' \
     -H 'anthropic-beta: tools-2024-04-04,prompt-caching-2024-07-31' \
     -H 'authorization: Bearer fake-not-validated' \
     -H 'x-api-key: sk-fake-not-validated' \
     -H 'user-agent: claude-cli/0.x' \
     -H 'x-app: claude-code' \
     --data @examples/flagship-gateway/messages-claude-code-style-request.json | jq .
```

期待响应（关键字段）：

```json
{
  "id": "msg_gw-req-<sha16>",
  "type": "message",
  "role": "assistant",
  "content": [{"type": "text", "text": "[mock-candidate sha:...] ..."}],
  "model": "claude-sonnet-4-5",
  "stop_reason": "end_turn",
  "stop_sequence": null,
  "usage": {"input_tokens": <int>, "output_tokens": <int>},
  "verdict": "YES",
  "gateway_audit": {
    "gateway_id": "flagship-coder-gateway-p0",
    "request_id": "gw-req-<sha16>",
    "pre_reasoner_audit": { "audit_id": "audit-<sha>", ... },
    "post_reasoner_audit": { "audit_id": "audit-<sha>", ... },
    "provider_name": "mock",
    "verdict": "YES",
    "timestamp": "2026-..."
  }
}
```

### 3.4 `POST /v1/messages` — HOLD

```sh
curl -s -X POST http://localhost:9091/v1/messages \
     -H 'content-type: application/json' \
     --data @examples/flagship-gateway/messages-hold-request.json | jq '.verdict, .stop_reason, .gateway_audit.post_reasoner_audit'
```

期待：`verdict="HOLD"`、`stop_reason="end_turn"`、`post_reasoner_audit` 为
`null`（pre HOLD 短路，未调用 mock provider）。HTTP 状态仍为 200。

### 3.5 `POST /v1/messages` — NO（硬约束触发）

```sh
curl -s -X POST http://localhost:9091/v1/messages \
     -H 'content-type: application/json' \
     --data @examples/flagship-gateway/messages-no-request.json | jq '.verdict, .stop_reason'
```

期待：`verdict="NO"`、`stop_reason="refusal"`、`content[0].text` 含
`reason_code=FLAGSHIP_PERMISSION_DENIED`。HTTP 状态仍为 200。

## 4. Claude Code 端接入说明（仅指引；不动用户机器配置）

**本轮 AI 编程代理不会、不应、不允许自动修改用户机器上的真实 Claude Code
配置（含 `~/.claude/`、Claude Code 设置文件、shell 环境变量等）。** 真实接入
是用户自己的动作；下面只提供参考做法。

如果用户想让 Claude Code 把 base URL 指向本网关：

```sh
# 用户在自己的 shell 中手动 export，或在自己的工程脚本里临时设置
export ANTHROPIC_BASE_URL=http://localhost:9091

# Claude Code 启动时自动读取该环境变量。占位 key 即可：
export ANTHROPIC_API_KEY=sk-fake-not-validated

# 然后启动 Claude Code（具体命令以用户安装方式为准）
```

> 由于本网关返回的 `model` 字段会原样回显 Claude Code 发来的 `model` 值（
> 如 `claude-sonnet-4-5`），Claude Code 在响应解析侧不需要看到自定义 model
> id 也能拿到 mock 候选文本。但 `/v1/models` 返回的列表里只有
> `flagship-coder-mock-1`；如果 Claude Code 在启动时校验 model 列表，可能需
> 要在用户侧选择"忽略 model 列表"或在 P2 加一个别名映射端点（H 项之一，见 §6）。

不要把 `ANTHROPIC_BASE_URL` 写入用户的 `~/.zshrc` / `~/.bashrc` / `~/.profile`
等持久化文件——这是用户自己的决定。

## 5. 联调验证记录

### 5.1 已验证接口（来自 `tests/flagship-gateway/claude_e2e_test.go`）

| 测试 | 端点 | 行为 | 结果 |
|---|---|---|---|
| `TestClaudeE2EModelsListShape` | `GET /v1/models` | 带 7 个 Claude 头，data 含 `flagship-coder-mock-1` | ✅ PASS |
| `TestClaudeE2EModelsAcceptsHeaders` | `GET /v1/models` | 带全 header，content-type 是 application/json | ✅ PASS |
| `TestClaudeE2EModelsWithoutHeaders` | `GET /v1/models` | 不带任何 Claude header 也 200 | ✅ PASS |
| `TestClaudeE2ECountTokensWithHeaders` | `POST /v1/messages/count_tokens` | 带全 header，input_tokens > 0 | ✅ PASS |
| `TestClaudeE2EMessagesYes` | `POST /v1/messages` | Claude 风格请求，verdict=YES，stop_reason=end_turn，model 回显 `claude-sonnet-4-5`，post audit 存在 | ✅ PASS |
| `TestClaudeE2EMessagesHold` | `POST /v1/messages` | HOLD 短路，HTTP 200，post audit 为 nil | ✅ PASS |
| `TestClaudeE2EMessagesNo` | `POST /v1/messages` | NO 短路，HTTP 200，stop_reason=refusal | ✅ PASS |
| `TestClaudeE2EEachHeaderAccepted` | `POST /v1/messages` | 7 个 Claude header 各自单独发送都 200（子测试 7/7 PASS） | ✅ PASS |
| `TestClaudeE2EUnknownFutureHeadersAccepted` | `POST /v1/messages` | 未知未来 header（含 `x-anthropic-experimental` / `x-claude-code-session-id`）也 200 | ✅ PASS |
| `TestClaudeE2EVerdictMatrixAllReturn200` | `POST /v1/messages` | YES / HOLD / NO 三态都是 HTTP 200，verdict 在 body，stop_reason 与 GD-3 对齐（子测试 3/3 PASS） | ✅ PASS |

### 5.2 已验证 header（E5）

下面这组 header **可被网关接收并忽略**（不导致 4xx，不校验真实 API key）：

- `anthropic-version`（含未来值如 `future-version-9999-99-99`）
- `anthropic-beta`（含 `tools-2024-04-04`、`prompt-caching-2024-07-31`、未来值）
- `authorization`（fake `Bearer ...` 不被校验）
- `x-api-key`（fake `sk-...` 不被校验）
- `content-type: application/json`
- `user-agent`（任意值，含 `claude-cli/0.x ...`）
- `x-app`（如 `claude-code`）
- 未来未知 header（如 `x-anthropic-experimental` / `x-claude-code-session-id`）

### 5.3 已验证三态（E4 + E6）

| Verdict | HTTP status | `body.verdict` | `body.stop_reason` | `body.gateway_audit.post_reasoner_audit` |
|---|---|---|---|---|
| YES | **200** | `"YES"` | `"end_turn"` | 非 nil（含 audit_id） |
| HOLD | **200** | `"HOLD"` | `"end_turn"` | nil（pre 短路） |
| NO | **200** | `"NO"` | `"refusal"` | nil（pre 短路） |

满足 GD-10（HOLD/NO 一律 200）与 GD-3（HOLD 复用 `end_turn`，NO 用 `refusal`）。

## 6. 当前不足（P2 决策项）

下列项是 P1 实测过程中暴露的、Claude Code 真实使用可能依赖、但本轮**显式
不做**的能力。任何启动这些项之前必须先做新一轮锚点滚动。

| ID | 不足 | 影响 | P2 决策 |
|---|---|---|---|
| GH-1 | 不接真实 LLM Provider | mock 候选不是真实回答；用户用 Claude Code 看到的是占位文本 | 是否接 Claude Messages API / OpenAI / Gemini / Qwen 等之一？需先决定支持哪家并选定 SDK 引入策略 |
| GH-2 | 不读取真实 API key | 无法把请求转发给上游 LLM | 与 GH-1 配套，需要决定 key 来源（`.env` / KMS / vault） |
| GH-3 | 不做鉴权 / 限流 / TLS | 网关对任何调用方开放；HTTP 明文 | 生产化前必做 |
| GH-4 | 不做 streaming（SSE） | Claude Code 重度依赖 streaming；当前 P0 一次性返回完整响应，Claude Code 体验会有明显延迟 | 与 GH-1 优先级耦合 |
| GH-5 | 不做 tool_use / tool_result | Claude Code 的工具调用功能完全无法工作；本轮 schema 已用 enum 闭合不允许 | 需要为 `internal/flagship/coder` 加工具循环编排 |
| GH-6 | 真实 Claude Code 端到端互联尚未实测 | 本轮所有验证通过 Go test + curl 完成；尚未看到真实 Claude Code 的 UI 在本地网关下端到端跑通 | 启动 Claude Code 客户端真实接入时需用户在自己机器上手动 export `ANTHROPIC_BASE_URL`（见 §4），并由用户主动提交联调结果 |
| GH-7 | 不接 Core ledger / signature | `GatewayAuditRecord` 仅返回对象，不上链 | 长期可审计性需要时启动 |
| GH-8 | 不沉淀 Gateway 审计链 | audit 仅返回响应，不写盘 | 与 GH-7 关联 |
| GH-9 | 不做模型别名映射 | `/v1/models` 只回 `flagship-coder-mock-1`；如 Claude Code 启动时校验 model id 是否存在于列表，可能需要别名 | 与 GH-1 配套：接真 LLM 后才有意义 |

## 7. 测试 + 验收命令

```sh
go build ./internal/flagship/... ./cmd/flagship-server/... ./cmd/flagship-coder-gateway/...
go vet  ./internal/flagship/... ./cmd/flagship-coder-gateway/... ./tests/flagship-gateway/...
go test ./tests/flagship-gateway/... ./cmd/flagship-coder-gateway/...
go test ./...

# JSON 校验（新增的 Claude 风格示例）
python3 -c "import json; [json.load(open(f)) for f in [
  'examples/flagship-gateway/messages-claude-code-style-request.json'
] if __import__('os').path.exists(f)]; print('claude e2e JSON: OK')"
```

## 8. 边界

- **不修改 P0/P1 已冻结推理内核**（`internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`、`cmd/flagship-server/**`、`tests/flagship/**`、`schemas/flagship/**`、`examples/flagship/**`、3 份 P0/P1 docs）。
- **不修改 Gateway P0 已冻结 schemas 与 docs**（`schemas/flagship-gateway/**`、`docs/FLAGSHIP_CODER_GATEWAY_P0_DEV.md`、`docs/FLAGSHIP_CODER_GATEWAY_P0_FREEZE_REPORT.md`）。
- **不引入新依赖**：`go.mod` / `go.sum` 未触；本轮只用标准库。
- **不接入真实 LLM**：`provider/mock.go` 沿用 P0 不变。
- **不动 Core /shear / openapi / sdk / policies / Docker / README / SUPPORT**。
- **不修改用户机器上的真实 Claude Code 配置**（`~/.claude/` / Claude Code settings / shell rc 等都未触）。
- **未在 E1–E9 之外做能力扩展**。

进入网关 P2 必须先做新一轮锚点滚动；不得沿用本轮 E1–E9 的授权半径。
