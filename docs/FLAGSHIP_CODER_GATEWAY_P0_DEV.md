# 熵剪旗舰版 + LLM 编程大脑网关 P0 — 开发笔记

> 本文件只描述 **闭源开发分支** 上的网关 P0 实现细节。它**不是** Entropy
> Shear Core 的对外能力。本网关**不进入** `policies/`、**不修改**
> `policies/manifest.json`、**不被** `openapi.yaml` / `sdk/` 暴露、**不被**
> `README.md` / `README_CN.md` / `SUPPORT.md` 宣传为 Core 能力。详见
> [`../LONGMA_TASK_ANCHOR.json`](../LONGMA_TASK_ANCHOR.json) 与
> [`../LONGMA_SOFT_GUARD.md`](../LONGMA_SOFT_GUARD.md)。

## 1. 目标

在不破坏 Core、不破坏 Flagship P0/P1 推理内核（`v0.4.1-flagship-p1-hardening`
已冻结）、不引入真实密钥、不修改现有 `/shear` 接口的前提下，提供一个可被
Claude Code / Codex 未来接入的编程大脑网关骨架（FLAGSHIP_CODER_GATEWAY_P0）：
独立服务、Anthropic Messages 兼容形态最小骨架、生成前与生成后均通过旗舰版
reasoner 治理、占位 mock LLM provider，限 G1–G12 共 12 项。本轮全部走
**确定性逻辑**，不接入任何真实 LLM。

## 2. 模块拓扑

```
internal/flagship/
  coder/       Anthropic 请求适配 + 响应包装 + GatewayAuditRecord
    audit.go      GatewayAuditRecord + builder
    adapter.go    BuildPreGovernanceInput / BuildPostGovernanceInput
    assistant.go  BuildAssistantContent — verdict → (text, stop_reason)
  provider/    LLM Provider 抽象 + 占位 mock
    provider.go   Provider 接口、ProviderMessage / GenerateRequest / GenerateResponse
    mock.go       MockProvider（确定性占位）
  gateway/     HTTP 表面 + 治理编排
    types.go      Anthropic Messages 兼容请求 / 响应类型集
    governance.go runGovernance：pre → mock generate → post → 包装响应
    server.go     Mux() + 4 个 handler + writeJSON / writeError

cmd/flagship-coder-gateway/
  main.go        二进制入口（默认 :9091，env ENTROPY_SHEAR_FLAGSHIP_GATEWAY_ADDR）
  main_test.go   handler 测试（405 / bad JSON / DisallowUnknownFields / /health 等）

schemas/flagship-gateway/      JSON Schema draft-07
examples/flagship-gateway/     YES / HOLD / NO 三组请求 + count_tokens 与 models 样例
tests/flagship-gateway/        gateway_e2e / adapter / mock_provider / count_tokens 测试
```

依赖方向（无循环）：
- `gateway → {coder, provider, reasoner}`
- `coder → {reasoner}`
- `provider → {}`（仅标准库）

`reasoner / state / mapper / rules / hold / output / cmd/flagship-server / tests/flagship`
都是 **P0/P1 已冻结**，本网关只 **import**，不修改任何源文件。

## 3. G1–G12 与 GD-1–GD-12 落地位置

| ID | 含义 | 落地位置 |
|---|---|---|
| **G1** 独立网关服务 | [`cmd/flagship-coder-gateway/main.go`](../cmd/flagship-coder-gateway/main.go) — 默认 `:9091`，env `ENTROPY_SHEAR_FLAGSHIP_GATEWAY_ADDR`；与 Flagship reasoner :9090 / Core :8080 完全独立 |
| **G2** Messages 最小骨架 | [`internal/flagship/gateway/types.go`](../internal/flagship/gateway/types.go) — `MessagesRequest`、`MessagesResponse`、`ContentBlock`、`Usage` 等 |
| **G3** GET /health | [`internal/flagship/gateway/server.go`](../internal/flagship/gateway/server.go) `healthHandler` |
| **G4** GET /v1/models | `server.go` `modelsHandler` + `governance.go` `ModelID` |
| **G5** POST /v1/messages/count_tokens | `server.go` `countTokensHandler` + `governance.go` `approxTokens` |
| **G6** POST /v1/messages | `server.go` `messagesHandler` → `governance.go` `runGovernance` |
| **G7** 生成前治理 | `governance.go` 第一段 `reasoner.Reason(preInput)`；非 YES 直接短路返回 |
| **G8** Mock LLM Provider | [`internal/flagship/provider/mock.go`](../internal/flagship/provider/mock.go) `MockProvider`；`provider.go` 的 `Provider` 接口 |
| **G9** 生成后审查 | `governance.go` 第二段 `reasoner.Reason(postInput)` |
| **G10** Claude Code 可识别 assistant message | [`internal/flagship/coder/assistant.go`](../internal/flagship/coder/assistant.go) `BuildAssistantContent` + `governance.go` `buildResponse` |
| **G11** AuditRecord 结构 | [`internal/flagship/coder/audit.go`](../internal/flagship/coder/audit.go) `GatewayAuditRecord`；只返回对象，不写盘 |
| **G12** 最小测试用例 | [`cmd/flagship-coder-gateway/main_test.go`](../cmd/flagship-coder-gateway/main_test.go) + [`tests/flagship-gateway/`](../tests/flagship-gateway/) 四份测试文件 |

| 决策 | 落地位置 |
|---|---|
| **GD-1** 请求最小字段集 + 忽略 anthropic 头 | `gateway/types.go` `MessagesRequest`；`server.go` 不读取 anthropic headers；测试 `TestMessagesHandlerIgnoresAnthropicHeaders` |
| **GD-2** 响应最小字段集 + verdict / gateway_audit 扩展 | `gateway/types.go` `MessagesResponse`；`governance.go` `buildResponse` |
| **GD-3** stop_reason 取值 | `coder/assistant.go` `StopReasonEndTurn` / `StopReasonRefusal`；YES → end_turn / NO → refusal / HOLD → end_turn |
| **GD-4** Pre 映射规则 | `coder/adapter.go` `BuildPreGovernanceInput`；request_id 前缀 `gw-pre-` |
| **GD-5** Post 映射规则 + 候选关键字 | `coder/adapter.go` `BuildPostGovernanceInput` + `candidateConstraints`；request_id 前缀 `gw-post-` |
| **GD-6** count_tokens 算法 | `governance.go` `approxTokens` / `ceilDiv`：`ceil(JSON 字节长度 / 4)` |
| **GD-7** Mock Provider 输出策略 | `provider/mock.go` `Generate`：`[mock-candidate sha:<first8>] ...` |
| **GD-8** 端口 + env | `cmd/flagship-coder-gateway/main.go` `defaultAddr=":9091"` + `ENTROPY_SHEAR_FLAGSHIP_GATEWAY_ADDR` |
| **GD-9** Models 列表 | `gateway/server.go` `modelsHandler` 返回 `flagship-coder-mock-1` 单条 |
| **GD-10** HOLD/NO 一律 200 | `governance.go` 不修改 status，由 `server.go writeJSON(..., http.StatusOK, ...)` 统一返回 200；只有 method/JSON 错误返 4xx |
| **GD-11** GatewayAuditRecord 字段 | `coder/audit.go` `GatewayAuditRecord` |
| **GD-12** content block type | `gateway/types.go` `ContentBlock.Type`（仅 "text"），schema enum 强制 |

## 4. 请求生命周期

```
HTTP POST /v1/messages
    │
    ▼
server.go messagesHandler
    │  - method 检查 → 405
    │  - dec.DisallowUnknownFields() + Decode → 400
    │
    ▼
governance.go runGovernance(req, mockProvider, now)
    │
    ├─ requestSeedID(req)  // sha256 前 16 hex，作为 pre/post id 派生种子
    ├─ lowerToPreFields(req) → coder.PreGovernanceFields
    ├─ coder.BuildPreGovernanceInput(seed, fields) → reasoner.Input (id="gw-pre-...")
    ├─ reasoner.Reason(preInput) → preOut (frozen kernel)
    │
    ├─ if preOut.Verdict != YES:
    │     coder.BuildAssistantContent(verdict, "", preOut.AlignmentTasks, preOut.RejectInstruction, "pre")
    │     → (text, stopReason)
    │     coder.NewGatewayAuditRecord(... post=nil ...)
    │     return MessagesResponse (HTTP 200, verdict=HOLD/NO)
    │
    ├─ provider.MockProvider.Generate(...) → genResp (deterministic candidate)
    │
    ├─ coder.BuildPostGovernanceInput(seed, preInput, genResp.Text) → postInput (id="gw-post-...")
    ├─ reasoner.Reason(postInput) → postOut
    │
    ├─ if postOut.Verdict != YES: candidate = ""  // 防止前端误执行被拒候选
    ├─ coder.BuildAssistantContent(postOut.Verdict, candidate, postOut.AlignmentTasks, postOut.RejectInstruction, "post")
    ├─ coder.NewGatewayAuditRecord(... post=&postOut.AuditRecord ...)
    └─ return MessagesResponse (HTTP 200)
```

## 5. 端口与启动

```sh
# 默认 :9091
go run ./cmd/flagship-coder-gateway

# 覆盖
ENTROPY_SHEAR_FLAGSHIP_GATEWAY_ADDR=:18080 go run ./cmd/flagship-coder-gateway

# 与 Flagship reasoner（:9090）/ Core server（:8080）互不影响。
```

## 6. 三个示例 + 期望 verdict

| 示例文件 | 期望 verdict | 触发原因 |
|---|---|---|
| [`examples/flagship-gateway/messages-yes-request.json`](../examples/flagship-gateway/messages-yes-request.json) | YES | 完整 system + user message，元素状态全高 |
| [`examples/flagship-gateway/messages-hold-request.json`](../examples/flagship-gateway/messages-hold-request.json) | HOLD | 无 system → Goal 描述空，Score 落入 [T2, T1) |
| [`examples/flagship-gateway/messages-no-request.json`](../examples/flagship-gateway/messages-no-request.json) | NO | system 含 "permission denied" 关键字 → hard constraint → ReasonPermissionDenied |

```sh
curl -s -X POST http://localhost:9091/v1/messages \
     -H 'Content-Type: application/json' \
     --data @examples/flagship-gateway/messages-yes-request.json | jq .
```

## 7. 测试 + 验收命令

```sh
go build ./internal/flagship/... ./cmd/flagship-server/... ./cmd/flagship-coder-gateway/...
go vet  ./internal/flagship/... ./cmd/flagship-coder-gateway/... ./tests/flagship-gateway/...
go test ./tests/flagship-gateway/... ./cmd/flagship-coder-gateway/...
go test ./...
```

JSON 校验：

```sh
python3 -c "import json; [json.load(open(f)) for f in [
  'schemas/flagship-gateway/messages-request.schema.json',
  'schemas/flagship-gateway/messages-response.schema.json',
  'schemas/flagship-gateway/count-tokens-request.schema.json',
  'schemas/flagship-gateway/count-tokens-response.schema.json',
  'schemas/flagship-gateway/models-response.schema.json',
  'examples/flagship-gateway/messages-yes-request.json',
  'examples/flagship-gateway/messages-hold-request.json',
  'examples/flagship-gateway/messages-no-request.json',
  'examples/flagship-gateway/count-tokens-request.json',
  'examples/flagship-gateway/models-response.json'
]]; print('flagship-gateway JSON: OK')"
```

## 8. 边界

- **不修改 P0/P1 已冻结推理内核**（`internal/flagship/{mapper,rules,state,hold,output,reasoner}/**`、`cmd/flagship-server/**`、`tests/flagship/**`、`schemas/flagship/**`、`examples/flagship/**`、3 份 P0/P1 docs）。网关只通过 **runtime import** 调用 `reasoner.Reason()`。
- **不引入新依赖**：`go.mod` / `go.sum` 未触；本网关只用标准库（`crypto/sha256`、`encoding/hex`、`encoding/json`、`net/http`、`net/http/httptest`、`time` 等）。
- **不接入真实 LLM**：仅 `MockProvider`，零外部网络 / 零真实 API key / 零 `.env` 读取。
- **不动 Core /shear / openapi / sdk / policies / Docker / README / SUPPORT**。
- **不实现 Anthropic 高级能力**：streaming / tool_use / citation / cache_control / thinking / images / documents 全部未实现，schema 已用 enum 闭合。
- **不做鉴权 / 限流 / TLS**：HTTP 裸服务；Anthropic 请求头允许出现但不校验。

下一阶段（网关 P1 或 P2）必须先做新一轮锚点滚动；不得沿用本轮 G1–G12 授权半径。
