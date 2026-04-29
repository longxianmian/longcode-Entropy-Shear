# 熵剪 Entropy Shear

> **大模型给你答案。熵剪告诉你能信哪一个。**

**熵剪不是 AI。它是一个确定性的裁决引擎。**

输入 `policy + facts`，返回三种无可辩驳的裁决之一：**`Yes`**（通过）、**`No`**（拒绝）、**`Hold`**（暂缓 / 需人工处理）。

不猜测。不产生幻觉。不调用大模型。
当逻辑陷入死结，它会承认 *"我判不了"* 并输出 `Hold`，而不是编造一个答案。

阅读其他语言：**中文** · [English](README.md)

---

## 为什么需要它

| 你的痛点 | 熵剪的解法 |
|---|---|
| **AI Agent 安全** | 每次 Agent 动作经由 `/shear` 裁决：`Yes` 放行 / `No` 拦截 / `Hold` 暂缓转人工。 |
| **业务规则全在代码里** | 把规则抽到 Policy JSON。改规则不动代码、不发版、不重启。 |
| **审计需要证据，不是截图** | 每次裁决哈希存证，追加写入可独立校验的 JSONL 链，离线在线均可验。 |
| **模糊地带不好处理** | `Hold` 是正式裁决态，把死结显式标记，而不是蒙混过关或硬冲拒绝。 |
| **合规 / 风控 / 知识产权** | 标准 Policy Pack 自带内容寻址 hash — 钉版本即可证明哪一版规则做出了哪一次裁决。 |

---

## 5 分钟快速体验

```bash
# 启动服务
docker compose up -d --build

# 健康检查
curl -s http://127.0.0.1:8080/health

# 第一次裁决
curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/cityone-request.json | jq

# 校验账本链
curl -s http://127.0.0.1:8080/ledger/verify | jq
```

完成。单二进制、JSONL 账本、零数据库、零鉴权脚手架。

本地 Go 开发：

```bash
go run ./cmd/server   # 监听 :8080，账本写到 ./ledger/shear-chain.jsonl
```

| 环境变量 | 默认值 |
|---|---|
| `ENTROPY_SHEAR_ADDR`   | `:8080` |
| `ENTROPY_SHEAR_LEDGER` | `ledger/shear-chain.jsonl` |

---

## 核心接口

| 方法 | 路径 | 用途 |
|---|---|---|
| `GET`  | `/health`            | 服务健康检查 |
| `POST` | `/shear`             | 单次裁决 |
| `GET`  | `/ledger/{shear_id}` | 查询单条账本记录 |
| `GET`  | `/ledger/verify`     | 整链复算校验 |

### 请求体

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

### 响应体

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

`signature` 与 `shear_id` 是审计凭据。每次 `/shear` 调用都会向 `ledger/shear-chain.jsonl` 追加一条记录。

### 三态语义

| Verdict | 含义 |
|---|---|
| `Yes`  | 允许 / 通过 / 可执行 |
| `No`   | 拒绝 / 不通过 / 不可执行 |
| `Hold` | 证据不足 / 需人工复核 / 待补充信息 |

`Hold` 是合法状态，不是错误。当全部规则未命中时，输出 `policy.default_effect`（一般为 `Hold`）。

### 8 个支持的操作符

`==`  `!=`  `>`  `<`  `>=`  `<=`  `in`  `contains`

字段路径用点号访问：`user.profile.level`。**路径不存在或类型不匹配时条件视为 `false`，不抛异常**。

### 错误码

| HTTP | code | 触发情形 |
|---|---:|---|
| 400 | `invalid_json` | JSON 解析失败 |
| 422 | `policy_schema_violation` / `facts_schema_violation` | 必填字段缺失或非法 |
| 422 | `unsupported_operator` | 操作符不在 8 个之内 |
| 200 | *（落入 `Hold`）* | 单条规则评估异常 — 在 trace 中记录原因 |
| 404 | `not_found` | 未知 `shear_id` |
| 503 | `ledger_unavailable` / `service_unavailable` | 磁盘 / 引擎故障 |

> 单条规则异常**不能**让服务崩溃 — 引擎落入下一条规则或 `default_effect`，原因写入 `trace`。

### 账本（Ledger）

- 文件：`ledger/shear-chain.jsonl`，每行一条 JSON 记录；
- 仅追加，不修改、不删除；
- 每条记录的 `previous_shear_hash` 等于上一条 `current_shear_hash`，首条用 `sha256:genesis`；
- `current_shear_hash` 含 timestamp（链式推进），response 中的 `signature` 不含 timestamp（同输入幂等）；
- 篡改任意一行，`GET /ledger/verify` 与 `cmd/verify-ledger` 都会报告 `broken_at`（1-indexed）。

离线校验：

```bash
go run ./cmd/verify-ledger
go run ./cmd/verify-ledger -ledger /path/to/shear-chain.jsonl
```

链完整退出码 0；篡改退出码 1；I/O 错误退出码 2。**不会创建任何文件或目录**。

---

## 真实应用场景

### 生产级 AI Agent 工具调用闸门

Agent 每次产出动作都经由 `/shear` 裁决再执行：

| 动作 | Verdict | 调用方处理 |
|---|---|---|
| `list_active_users`（只读） | **Yes**  | 执行 |
| `delete_production_data`    | **No**   | 拒绝并暴露原因 |
| `transfer_funds`            | **Hold** | 推入人工确认队列 |

可运行样板（Node + Python）：[`integrations/agent-tool-gate/`](integrations/agent-tool-gate/)。

### AI 客服分流

退款诉求？`/shear` 检查订单号 — 缺信息 → `Hold` 转补证；医疗建议 → `No` 引导专业人士；高置信 FAQ → `Yes` 由 AI 答复。

策略包与 facts 示例：[`policies/ai-customer-service/`](policies/ai-customer-service/) + [`integrations/ai-customer-service-gate/`](integrations/ai-customer-service-gate/)。

### 标书风控（招投标合规）

| 投标状态 | Verdict |
|---|---|
| 必需文件缺失 | **No** — 拦截，列出缺失项 |
| 资质证据不完整 | **Hold** — 转合规人工复核 |
| 关键风险项均通过 | **Yes** — 通过并完整留痕 |

策略包：[`policies/bid-risk/`](policies/bid-risk/)；请求示例：[`examples/bid-risk-request.json`](examples/bid-risk-request.json)。

### 教育 AI 安全（小学 / 中学 / 大学）

三态裁决天然映射不同年龄段的内容安全策略。下表给出策略方向参考；具体 pack 应自行起草并 hash 钉版。

| 学段 | 具体闸门 | Yes | No | Hold |
|---|---|---|---|---|
| **小学** | AI 作业辅导内容安全 | 题目在年级范围、内容适龄 | 含成人 / 暴力 / 越界内容 | 超纲题或考试期 — 转任课老师 / 家长 |
| **中学** | 考试期答题资格 | 练习模式、允许辅助的题型 | 在线考试窗口、被评估学科 | 题目与考试范围擦边 — 监考人工复核 |
| **大学** | AIGC / 学术诚信合规 | AIGC 占比低于阈值、查重 OK | AIGC 超过硬上限或命中已知抄袭 | 临界值 — 提交导师人工审阅 |

### 跨渠道身份路由

CityOne 类营销系统：LINE / 微信 / App 用户落地，统一由 `/shear` 决定路由。规则改 JSON 即生效，免发版。

示例：[`examples/cityone-request.json`](examples/cityone-request.json)。

### 企业准入

VIP / 会员 → `Yes` + 路由；黑名单标签 → `No`；身份信息不足 → `Hold` 要求补证。

策略包：[`policies/permission-gate/`](policies/permission-gate/)。

---

## SDK 与策略工具

```ts
// Node 18+ / TypeScript — 无构建步骤，使用全局 fetch
import { EntropyShearClient } from "@longcode/entropy-shear-client";
const r = await new EntropyShearClient({ baseUrl: "http://127.0.0.1:8080" })
  .shear({ policy, facts });
```

```python
# Python 3.9+ — 仅 stdlib
from entropy_shear_client import EntropyShearClient
r = EntropyShearClient(base_url="http://127.0.0.1:8080").shear(policy=policy, facts=facts)
```

完整客户端：[`sdk/js/`](sdk/js/) · [`sdk/python/`](sdk/python/)。

离线校验和给 policy pack 打 hash：

```bash
go run ./cmd/validate-policy --file policies/agent/agent-action-policy.v1.json
go run ./cmd/hash-policy     --file policies/agent/agent-action-policy.v1.json
```

`policies/manifest.json` 记录每个 pack 的 SHA-256；`tests/policy_pack_test.go` 每次测试都复算并比对，**任何漂移都会让 CI 失败**。

---

## 开发状态

当前版本：**`v0.2.0-p2`**（P0 + P1 + P2）

| 阶段 | 标签 | 交付 |
|---|---|---|
| **P0** *（已冻结）* | `p0-freeze-20260428` | 裁决内核 — `/shear`、三态、trace、signature、JSONL ledger、Docker、单测 |
| **P1**             | `v0.1.0-p1`          | OpenAPI 3.0.3、JSON Schemas、JS / Python SDK、离线 `cmd/verify-ledger`、Agent 三态全覆盖示例 |
| **P2** *（当前）*  | `v0.2.0-p2`          | 版本化 policy pack + manifest、`cmd/{validate,hash}-policy`、Agent Tool Gate 样板（Node + Python）、AI 客服 Gate、4 篇开发文档 |

完整发布说明与验收命令：[`docs/P1_RELEASE_CHECKLIST.md`](docs/P1_RELEASE_CHECKLIST.md)、[`docs/INTEGRATION_GUIDE.md`](docs/INTEGRATION_GUIDE.md)。

### 熵剪永远不会做的事

- ❌ 不调用任何 LLM — 无幻觉、无 prompt injection 攻击面
- ❌ 不自动生成规则 — 规则的话语权永远在你这边
- ❌ 不依赖外部数据库 — 一个二进制、一个 JSONL
- ❌ 不实现用户 / 鉴权 / 多租户系统 — 它站在你的鉴权**之后**，而非之前
- ❌ 不与任何特定平台耦合（不是 LIOS 子模块、不是任何专有宿主组件）
- ❌ 不把业务流程硬编码进引擎
- ❌ 不持久化原始业务数据 — 账本里只有 hash

### 仓库结构

```
entropy-shear/
  cmd/                         server / verify-ledger / validate-policy / hash-policy
  internal/                    api / engine / schema / ledger / signature / policy / errors
  examples/                    POST /shear 请求体（cityone、agent ×4、ai、bid、gate）
  schemas/                     JSON Schema (draft 2020-12)：Policy / Request / Response / Ledger
  sdk/                         js (TypeScript / fetch) + python (urllib stdlib)
  openapi.yaml                 OpenAPI 3.0.3 接口规范
  policies/                    版本化 policy 模板库 + manifest.json（hash 钉版）
    agent/  ai-customer-service/  bid-risk/  permission-gate/
  integrations/                可运行的接入样板（不实现 LLM、不实现 Agent 本身）
    agent-tool-gate/{node,python}   ai-customer-service-gate/
  docs/                        WHITEPAPER / INTEGRATION_GUIDE / POLICY_PACK_GUIDE
                               AGENT_TOOL_GATE_GUIDE / P1_RELEASE_CHECKLIST
  ledger/shear-chain.jsonl     追加式账本（运行时生成）
  tests/                       go test ./...（engine / handler / examples / schemas / packs）
  Dockerfile  docker-compose.yml  NOTICE.md  LICENSE
```

### 验收命令

```bash
go test ./...
docker compose up -d --build
curl -s http://127.0.0.1:8080/health | jq
curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/agent-action-yes-request.json | jq    # → Yes
curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/agent-action-no-request.json | jq     # → No
curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/agent-action-hold-request.json | jq   # → Hold
curl -s http://127.0.0.1:8080/ledger/verify | jq
go run ./cmd/verify-ledger
go run ./cmd/validate-policy --file policies/agent/agent-action-policy.v1.json
go run ./cmd/hash-policy     --file policies/agent/agent-action-policy.v1.json
```

---

## 许可证

核心引擎采用 **Apache License 2.0** 开源 — 详见 [`LICENSE`](LICENSE) 与 [`NOTICE.md`](NOTICE.md)。

商业部署支持、私有化授权、SLA 保障与定制集成，请见 [`SUPPORT.md`](SUPPORT.md)。

---

## 商业化支持

核心引擎 Apache 2.0 永久开源免费。需要更多的团队可选：

- **Entropy Shear Cloud** — 按量计费的托管 API，免去运维。
- **Entropy Shear On-Prem** — 年付授权 + SLA，面向金融、医疗、合规场景。
- **定制 Policy Pack 开发** — 行业专属规则包（标书风控、合规审查、教育 AI 安全、Agent 治理），可签 NDA。

👉 [商业授权与支持 →](SUPPORT.md)

---

阅读其他语言：**中文** · [English](README.md)
