# 熵剪 Entropy Shear · P0 + P1 + P2

> **熵剪 Entropy Shear：生产级 Agent 的轻量三态裁决引擎。**
>
> - **LLM / Agent = 理解、规划、生成。**
> - **熵剪 = 裁决、拦截、暂缓、留痕。**

输入 `policy + facts`，输出 `Yes / No / Hold` + 完整 trace + 不可篡改 signature + 追加式 ledger。熵剪不是 AI、不是 LIOS 子模块、不是聊天机器人 — 它是一个独立的、确定性的三态裁决引擎，做 Agent / AI 客服 / 风控系统的关键动作闸门。

- **P0**（已冻结，tag `p0-freeze-20260428`）：可运行的裁决内核 — `/shear`、三态、trace、signature、JSONL ledger、Docker、测试。
- **P1**（tag `v0.1.0-p1`）：在不破坏 P0 兼容的前提下，补齐 `openapi.yaml`、JSON Schemas、典型场景示例（含 Agent 三态全覆盖）、最小 JS / Python SDK、`cmd/verify-ledger` 离线校验器、测试增强。详见 [§11](#11-p1-产品化增强)。
- **P2**（当前）：把熵剪从「可调用的裁决服务」推进为「可被真实 Agent / AI 客服 / 风控系统接入的治理组件」。新增 policy pack 模板库 + manifest、`cmd/{validate,hash}-policy`、Agent Tool Gate（Node + Python）样板、AI 客服 Gate facts 示例、技术白皮书与接入手册。详见 [§12](#12-p2-接入治理)。

---

## 1. 接口

| 方法 | 路径 | 说明 |
|---|---|---|
| `GET`  | `/health`            | 服务健康检查 |
| `POST` | `/shear`             | 单次裁决（Yes/No/Hold） |
| `GET`  | `/ledger/{shear_id}` | 查询单条裁决账本记录 |
| `GET`  | `/ledger/verify`     | 整链复算校验 |

### 1.1 `POST /shear`

请求体：

```json
{
  "policy": {
    "id": "policy-cityone-v1",
    "version": "1.0.0",
    "rules": [
      {
        "id": "rule-001",
        "priority": 1,
        "condition": { "field": "user.level", "operator": "in", "value": ["member", "vip"] },
        "effect": "Yes",
        "route": "/campaign/member-landing",
        "reason": "会员或 VIP 用户允许进入会员活动页"
      }
    ],
    "default_effect": "Hold",
    "default_reason": "未命中任何规则，需人工复核"
  },
  "facts": {
    "user": { "id": "U-10001", "level": "member", "tags": [] }
  }
}
```

响应体：

```json
{
  "verdict": "Yes",
  "applied_rule_id": "rule-001",
  "route": "/campaign/member-landing",
  "reason": "会员或 VIP 用户允许进入会员活动页",
  "trace": [
    { "rule_id": "rule-001", "evaluated": true, "matched": true,
      "detail": "user.level in [member vip]，实际值为 member，命中" }
  ],
  "signature": "sha256:...",
  "shear_id": "entropy-shear-20260428-000001"
}
```

---

## 2. 三态语义

| Verdict | 语义 |
|---|---|
| `Yes`  | 允许 / 通过 / 可执行 |
| `No`   | 拒绝 / 不通过 / 不可执行 |
| `Hold` | 证据不足 / 需人工复核 / 待补充信息 |

`Hold` 是合法状态，不是错误。当全部规则未命中时，输出 `policy.default_effect`（一般为 `Hold`）。

---

## 3. 8 个支持的操作符

`==` `!=` `>` `<` `>=` `<=` `in` `contains`

字段路径用点号访问：`user.profile.level`。路径不存在或类型不匹配时**条件视为 false，不抛异常**。

---

## 4. 账本（Ledger）

- 文件：`ledger/shear-chain.jsonl`，每行一条 JSON 记录；
- 仅追加，不修改、不删除；
- 每条记录的 `previous_shear_hash` 必须等于上一条的 `current_shear_hash`；
- 第一条 `previous_shear_hash = "sha256:genesis"`；
- 单进程并发写入由互斥锁保证；
- `GET /ledger/verify` 从 genesis 复算每一行，任一不一致即报告 `broken_at`（1-indexed）。

`current_shear_hash` 含 timestamp（链式推进），response 中的 `signature` 不含 timestamp（同输入幂等）。

---

## 5. 本地运行

需要 Go 1.22+：

```bash
go run ./cmd/server
# 监听 :8080，账本写到 ./ledger/shear-chain.jsonl
```

环境变量：

| 变量 | 默认值 |
|---|---|
| `ENTROPY_SHEAR_ADDR`   | `:8080` |
| `ENTROPY_SHEAR_LEDGER` | `ledger/shear-chain.jsonl` |

---

## 6. Docker 运行

```bash
docker compose up -d --build
curl -s http://127.0.0.1:8080/health
```

`./ledger` 卷挂载到容器，账本在宿主机持久化。

---

## 7. 验收命令

```bash
# 单元测试
go test ./...

# 启动服务
docker compose up -d --build

# 健康检查
curl -s http://127.0.0.1:8080/health

# 单次裁决
curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/cityone-request.json | jq

# 账本校验
curl -s http://127.0.0.1:8080/ledger/verify | jq
```

---

## 8. 目录结构

```
entropy-shear/
  cmd/server/main.go              # 启动入口
  internal/
    api/                          # HTTP handler / 错误响应
    engine/                       # 三态裁决核心 + 操作符评估器
    schema/                       # Policy/Rule/Condition/Facts 类型 + 校验
    ledger/                       # JSONL 追加账本 + verify
    signature/                    # 稳定 JSON + SHA-256
    errors/                       # 统一错误码
  cmd/verify-ledger/main.go       # P1：离线账本校验
  examples/                       # 5 个 POST /shear 完整请求体
    cityone-request.json          # CityOne 会员准入 → Yes
    agent-action-request.json     # P1：Agent 动作（只读 → Yes）
    ai-customer-service-request.json # P1：AI 客服（refund 无单号 → Hold）
    bid-risk-request.json         # P1：招投标合规（必需文件缺失 → No）
    permission-gate-request.json  # P1：权限准入（VIP → Yes）
    *-policy.json                 # 仅 policy 部分（用作组合参考）
  schemas/                        # P1：JSON Schema 固化
    policy.schema.json
    shear-request.schema.json
    shear-response.schema.json
    ledger-record.schema.json
  sdk/                            # P1：最小 HTTP SDK（不做本地裁决）
    js/                           #   Node 18+ / TypeScript
    python/                       #   Python 3.9+ stdlib only
  openapi.yaml                    # P1：OpenAPI 3.0.3 接口规范
  cmd/validate-policy/main.go     # P2：校验单个 policy 文件
  cmd/hash-policy/main.go         # P2：输出 policy 的稳定 hash
  internal/policy/                # P2：Load / Hash / Manifest 共享逻辑
  policies/                       # P2：版本化 policy pack 模板库
    manifest.json                 #   全部 pack + hash 索引
    agent/                        #   Agent Tool Gate
    ai-customer-service/          #   AI 客服 Gate
    bid-risk/                     #   招投标合规
    permission-gate/              #   企业准入
  integrations/                   # P2：接入样板（不实现 Agent 本身）
    agent-tool-gate/{node,python} #   Tool Gate 三态调用样板
    ai-customer-service-gate/     #   AI 客服 facts 示例
  docs/                           # P2：白皮书 + 接入手册（4 篇）
    WHITEPAPER.md
    INTEGRATION_GUIDE.md
    POLICY_PACK_GUIDE.md
    AGENT_TOOL_GATE_GUIDE.md
  ledger/shear-chain.jsonl        # 账本（运行时生成）
  tests/                          # 单元测试 + handler / examples / schema 测试
  docs/P1_RELEASE_CHECKLIST.md    # P1 上线检查清单
  Dockerfile
  docker-compose.yml
```

---

## 9. P0 边界

**只做**：`/health`、`POST /shear`、`GET /ledger/{shear_id}`、`GET /ledger/verify`、条件评估器、三态裁决引擎、trace、signature、JSONL ledger、Docker、示例 policy、单元测试。

**不做**：LLM 调用、规则自动生成、用户系统、权限系统、管理后台、数据库依赖、LIOS 依赖、复杂 DSL、原始业务数据全文持久化、多租户。

---

## 10. 错误码

| 错误 | HTTP | code |
|---|---:|---|
| JSON 解析失败 | 400 | `invalid_json` |
| Policy 字段缺失/非法 | 422 | `policy_schema_violation` |
| Facts 字段缺失/非法 | 422 | `facts_schema_violation` |
| 不支持的 operator | 422 | `unsupported_operator` |
| 单条规则评估异常 | 200 | 转 Hold，trace 中记录原因 |
| 账本写入失败 | 503 | `ledger_unavailable` |
| 系统级不可用 | 503 | `service_unavailable` |
| 找不到 shear_id | 404 | `not_found` |

---

## 11. P1 产品化增强

P1 在 P0 已冻结基线上**只做加法**：不改 `/shear` 主接口，不动核心裁决逻辑。

### 11.1 OpenAPI 与 JSON Schema

- `openapi.yaml` — 全部 4 个端点的 OpenAPI 3.0.3 规范，可直接喂给 Swagger Editor / openapi-generator。
- `schemas/policy.schema.json` — Policy 结构（draft 2020-12）。
- `schemas/shear-request.schema.json` — `POST /shear` 请求体。
- `schemas/shear-response.schema.json` — `POST /shear` 响应体。
- `schemas/ledger-record.schema.json` — JSONL 单行记录。

仅描述当前已实现字段，不预留未实现的扩展位。

### 11.2 典型场景示例

每个文件都是完整的 `POST /shear` 请求体（含 policy + canonical facts）。同一个 policy 可通过不同 facts 触发 Yes / No / Hold，详见 `tests/examples_test.go`。

| 文件 | canonical 结果 | 触发的规则 / 默认 |
|---|---|---|
| `examples/cityone-request.json`             | `Yes`  | `rule-001`（VIP/member 准入） |
| `examples/agent-action-request.json`        | `Yes`  | `allow-readonly-action`（兼容示例） |
| `examples/agent-action-yes-request.json`    | `Yes`  | `allow-readonly-action`（只读查询） |
| `examples/agent-action-no-request.json`     | `No`   | `forbid-delete-production-data`（删除生产数据） |
| `examples/agent-action-hold-request.json`   | `Hold` | `hold-payment-action`（转账未人工确认） |
| `examples/ai-customer-service-request.json` | `Hold` | `need-order-id-before-human` |
| `examples/bid-risk-request.json`            | `No`   | `missing-required-file` |
| `examples/permission-gate-request.json`     | `Yes`  | `vip-or-member-allow` |

> **Agent 三态全覆盖**：`agent-action-{yes,no,hold}-request.json` 三个示例分别证明熵剪可以让 Agent 通过、拒绝、暂缓 — 这正是「LLM 负责生成，熵剪负责裁决」的最小可演示集合。

### 11.3 SDK（HTTP 封装，不做本地裁决）

- `sdk/js`：Node 18+ / TypeScript，使用全局 `fetch`，无构建步骤。
- `sdk/python`：Python 3.9+，仅依赖 stdlib `urllib`。

```ts
// JS / TypeScript
import { EntropyShearClient } from "@longcode/entropy-shear-client";
const client = new EntropyShearClient({ baseUrl: "http://127.0.0.1:8080" });
const r = await client.shear({ policy, facts });
```

```python
# Python
from entropy_shear_client import EntropyShearClient
client = EntropyShearClient(base_url="http://127.0.0.1:8080")
r = client.shear(policy=policy, facts=facts)
```

详见 [`sdk/js/README.md`](sdk/js/README.md) 和 [`sdk/python/README.md`](sdk/python/README.md)。

### 11.4 离线账本校验

```bash
go run ./cmd/verify-ledger
go run ./cmd/verify-ledger -ledger /path/to/shear-chain.jsonl
```

输出：

```json
{
  "ok": true,
  "total": 3,
  "broken_at": null,
  "latest_shear_id": "entropy-shear-20260428-000003",
  "latest_hash": "sha256:..."
}
```

链不一致时退出码为 1，I/O 错误时退出码为 2。**不会创建任何文件或目录**。

### 11.5 P1 验收命令

```bash
go test ./...

docker compose up -d --build
curl -s http://127.0.0.1:8080/health

curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/agent-action-request.json | jq

curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/ai-customer-service-request.json | jq

curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/bid-risk-request.json | jq

curl -s http://127.0.0.1:8080/ledger/verify | jq
go run ./cmd/verify-ledger
```

详细清单见 [`docs/P1_RELEASE_CHECKLIST.md`](docs/P1_RELEASE_CHECKLIST.md)。

### 11.6 P1 仍然不做

LLM 调用 · 规则自动生成 · 后台管理系统 · 用户系统 · 权限系统 · 数据库依赖 · 多租户 · LIOS 耦合 · 复杂 DSL · 业务流程硬编码。

---

## 12. P2 接入治理

P2 不是「再多做一些功能」，而是让熵剪**更容易被真实系统接入**。同样在 P0 / P1 已冻结基线上**只做加法**。

### 12.1 Policy Pack 模板库

`policies/` 是版本化、hash 固化的 policy 包：

```
policies/
  manifest.json                       # 全部 pack 的索引 + hash
  agent/
    agent-action-policy.v1.json
    README.md
  ai-customer-service/
    ai-customer-service-policy.v1.json
    README.md
  bid-risk/
    bid-risk-policy.v1.json
    README.md
  permission-gate/
    permission-gate-policy.v1.json
    README.md
```

每个 pack 的 README 列明：facts 结构、verdict 表、边界、复算命令。`policies/manifest.json` 中 `hash` 字段在 `tests/policy_pack_test.go` 中**每次测试时复算并比对**，任何漂移都会让 CI 失败。

详见 [`docs/POLICY_PACK_GUIDE.md`](docs/POLICY_PACK_GUIDE.md)。

### 12.2 策略校验与版本工具

```bash
# 任一 pack 都可以离线校验和打 hash
go run ./cmd/validate-policy --file policies/agent/agent-action-policy.v1.json
go run ./cmd/hash-policy     --file policies/agent/agent-action-policy.v1.json
```

输出：

```json
{ "ok": true, "policy_id": "policy-agent-action-v1", "version": "1.0.0", "rule_count": 3 }
{ "policy_id": "policy-agent-action-v1", "version": "1.0.0", "hash": "sha256:..." }
```

非法 policy 会返回 `ok: false` 并退出码 1；I/O 错误退出码 2。`hash-policy` 拒绝为不通过校验的 policy 出 hash。

### 12.3 Agent Tool Gate 样板（不是 Agent 本身）

`integrations/agent-tool-gate/{node,python}` 提供生产级 Agent 接入熵剪的最小样板：

```
agent.plan() → action → AgentToolGate.gate(action)
                       → POST /shear → Yes / No / Hold
                       → execute / refuse / queue-for-human
```

```bash
docker compose up -d --build
node    integrations/agent-tool-gate/node/example.mjs
python3 integrations/agent-tool-gate/python/example.py
```

两个样板对相同的三个动作产生 allow / deny / hold 三种 decision。详见 [`docs/AGENT_TOOL_GATE_GUIDE.md`](docs/AGENT_TOOL_GATE_GUIDE.md)。

### 12.4 AI 客服 Gate

`integrations/ai-customer-service-gate/facts-examples/` 配 `policies/ai-customer-service/` 使用：

| 文件 | 预期 verdict | 业务动作 |
|---|---|---|
| `refund-missing-order.json` | `Hold` | 要求补充订单号，再转人工 |
| `high-risk-medical.json`    | `No`   | 拒答，引导专业人士 |
| `faq-high-confidence.json`  | `Yes`  | AI 渲染答案 |

测试覆盖：`tests/policy_pack_test.go::TestAICustomerServiceFactsExamples`。

### 12.5 P2 文档

| 文档 | 用途 |
|---|---|
| [`docs/WHITEPAPER.md`](docs/WHITEPAPER.md) | 三态裁决型 AI 治理引擎的技术与定位说明 |
| [`docs/INTEGRATION_GUIDE.md`](docs/INTEGRATION_GUIDE.md) | 一小时接入指南：启动 / 调用 / 设计 policy / 处理三态 / 校验 ledger |
| [`docs/POLICY_PACK_GUIDE.md`](docs/POLICY_PACK_GUIDE.md) | pack 目录约定、manifest 规范、hash 规则、新增 pack 的 lifecycle |
| [`docs/AGENT_TOOL_GATE_GUIDE.md`](docs/AGENT_TOOL_GATE_GUIDE.md) | Tool Gate 接入模式、action facts 结构、三态处理建议、高风险动作分类 |

### 12.6 P2 验收命令

```bash
go test ./...
docker compose up -d --build
curl -s http://127.0.0.1:8080/health | jq
curl -s http://127.0.0.1:8080/ledger/verify | jq
go run ./cmd/verify-ledger
go run ./cmd/validate-policy --file policies/agent/agent-action-policy.v1.json
go run ./cmd/hash-policy     --file policies/agent/agent-action-policy.v1.json
node    integrations/agent-tool-gate/node/example.mjs    # 服务跑起来后
python3 integrations/agent-tool-gate/python/example.py
```

### 12.7 P2 仍然不做

LLM 调用 · 规则自动生成 · 后台管理系统 · 用户系统 · 权限系统 · 数据库依赖 · 多租户 SaaS · LIOS 耦合 · 复杂 DSL · 业务流程硬编码。
