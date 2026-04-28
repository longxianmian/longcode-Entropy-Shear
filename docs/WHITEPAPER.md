# Entropy Shear · 三态裁决型 AI 治理引擎

> 面向生产级 Agent / AI 客服 / 风控系统的轻量裁决基础设施。

版本：与 `v0.2.0-p2` 同步。

---

## 1. 背景

当 AI Agent 真正进入生产，「能说话」不再是问题，「敢不敢动」才是。

- LLM **生成动作**，但生成的动作未经鉴定就执行 = 事故；
- Guardrails / 关键词过滤 **只能拦截文本**，拦不住动作；
- 二元策略（允许 / 拒绝）面对「证据不全」时只能强行二选一；
- 业务系统里的关键判断散落在代码各处，**事后既无法追溯，也无法举证**。

```
LLM / Agent  =  理解、规划、生成
Entropy Shear =  裁决、拦截、暂缓、留痕
```

熵剪不是 AI，也不是 Agent；它是 Agent 与"实际副作用"之间那道 **确定性闸门**。

---

## 2. 核心问题

> 当一个 Agent 想做一件事，谁来裁决「能做 / 不能做 / 暂缓」？

主流方案的痛点：

| 方案 | 局限 |
|---|---|
| 关键词 / 正则过滤 | 只拦文本，不拦动作；规则爆炸；无审计 |
| 大模型自评（self-critique） | 不确定、不可复算、无签名 |
| Guardrails 类工具 | 与 LLM 厂商耦合；以拦截输出为主，缺独立裁决与留痕 |
| 业务代码 if-else | 散落、难审计、无版本、规则与代码耦合 |
| 二元 RBAC | 没有「证据不足」状态；硬冲到 Yes 或 No 都不安全 |

熵剪的回答：**把「能做 / 不能做 / 暂缓」从一段代码里抽离成一个独立、可校验、可签名、可留痕的裁决服务。**

---

## 3. 三态逻辑

熵剪的最小语义只有三个 verdict：

| Verdict | 业务含义 |
|---|---|
| `Yes`  | 允许执行 |
| `No`   | 拒绝执行 |
| `Hold` | 证据不足，需要人工确认 / 补充信息 |

`Hold` 是熵剪与传统二元策略的关键差别：**模糊不再被强行二选一**，而是被显式地标注为"需要人介入"。

每条 `/shear` 请求都包含：

- 输入 `{policy, facts}`；
- 输出 `verdict + applied_rule_id + reason + trace + signature + shear_id`；
- ledger 中追加一条带 hash 链的记录。

详细字段见 [`openapi.yaml`](../openapi.yaml) 与 [`schemas/`](../schemas)。

---

## 4. 三件套：Trace / Signature / Ledger

熵剪的可治理性来自三件套：

### 4.1 Trace

每一条规则的评估结果都记录到 `trace[]`：

```json
[
  { "rule_id": "forbid-delete-production-data", "matched": false, "detail": "action.name == delete_production_data，实际值为 list_active_users，未命中" },
  { "rule_id": "hold-payment-action",          "matched": false, "detail": "action.category == payment，实际值为 data，未命中" },
  { "rule_id": "allow-readonly-action",        "matched": true,  "detail": "action.mode == readonly，实际值为 readonly，命中" }
]
```

任何一次裁决都可以被解释。「为什么我被拒/被放行/被 Hold」始终有答案。

### 4.2 Signature

`signature = sha256( canonical( policy_id, policy_version, input_hash, verdict, applied_rule_id, trace_hash, previous_shear_hash ) )`

同样的输入对同一条 chain 永远得到同样的签名。第三方审计可以在 **不持有 facts 原文** 的前提下，仅凭 `policy_id / version / facts_hash` 复算并核对结果。

### 4.3 Ledger

`ledger/shear-chain.jsonl` 是一条只追加的 hash 链：

- 每行一条 JSON 记录；
- 每条记录的 `previous_shear_hash` = 上一行的 `current_shear_hash`；
- 第一行 `previous_shear_hash = "sha256:genesis"`；
- 任何一行被篡改，`GET /ledger/verify` 会立刻报告 `broken_at`；
- `cmd/verify-ledger` 提供离线复算，**不写入、不创建目录**，可作为归档检验工具。

这是 AI 时代的「裁决留证」基础设施：每一次允许、拒绝、暂缓都可以**事后举证**。

---

## 5. Agent 动作治理

Agent 在生产中的 95% 风险来自**未经鉴定的工具调用**，不是来自文本。

熵剪给 Agent 一个标准接入面：

```
agent.plan() → action object
              ↓
        AgentToolGate.gate(action)
              ↓
        POST /shear
              ↓
        Yes / No / Hold + trace + signature + shear_id
              ↓
        execute / refuse / queue-for-human
```

参考实现：[`integrations/agent-tool-gate/{node,python}`](../integrations/agent-tool-gate/)。

要点：

- Tool Gate **不实现 Agent**，只是 Agent 与工具执行器之间的薄层。
- `Hold` 不应被自动重试。Hold 表示「需要人」，不是「需要再试一次」。
- Tool Gate 不缓存裁决：每次动作都进 ledger，审计才完整。
- Tool Gate 不替代鉴权；鉴权放在 Agent runtime，Tool Gate 只看 facts。

---

## 6. 与既有方案的差异

| 维度 | 关键词过滤 | Guardrails 类 | 业务 if-else | 熵剪 |
|---|---|---|---|---|
| 拦得住动作 | ✗ | 部分（局限于 LLM 输出） | ✓ | **✓** |
| 三态（含 Hold） | ✗ | ✗ | 通常没有 | **✓** |
| 独立服务，跨语言可用 | 看实现 | 厂商耦合 | 否 | **✓** |
| 输出可解释（trace） | 弱 | 弱 | 弱 | **✓** |
| 可校验签名 | ✗ | ✗ | ✗ | **✓** |
| 不可篡改 ledger | ✗ | ✗ | ✗ | **✓** |
| 规则可版本化、可 hash | 弱 | 弱 | 强耦合 | **✓ + manifest** |

熵剪不是要替代上面任何一个工具，而是补齐它们没做的事：**让"关键动作的允许/拒绝/暂缓"这件事，本身可被审计**。

---

## 7. 应用场景

| 场景 | 熵剪的角色 |
|---|---|
| 生产级 Agent 工具调用 | Tool Gate：决定能否执行 / 是否需要人工确认 |
| AI 客服 | 决定能否回答 / 是否拒答 / 是否补订单号 / 是否转人工 |
| 招投标合规风控 | 决定文件是否合规 / 是否需复核 / 是否通过 |
| 企业准入 | 决定用户是否可访问 / 可领取 / 可执行某动作 |
| 共享 / 会员系统 | 决定能否借用 / 退款 / 进入指定流程 |

每个场景都共享同一个 `/shear` 接口，差异只在 **policy pack** —— 见 [`policies/`](../policies/) 与 [`docs/POLICY_PACK_GUIDE.md`](POLICY_PACK_GUIDE.md)。

---

## 8. 当前与未来

P0 → P1 → P2 已交付：

- P0：可运行的裁决内核，含 `/shear`、三态、trace、signature、JSONL ledger、Docker、测试。
- P1：OpenAPI、JSON Schemas、4 个典型场景示例（含 Agent 三态全覆盖）、JS / Python SDK、`cmd/verify-ledger` 离线校验。
- P2（本版）：policy pack 模板库 + manifest、`cmd/{validate,hash}-policy`、Agent Tool Gate Node/Python 样板、AI 客服 Gate facts 示例、技术白皮书与接入手册。

P2 之后熵剪的边界仍然清晰：

> 不做 LLM、不做规则自动生成、不做后台、不做用户系统、不做权限系统、不接数据库、不做多租户、不接 LIOS、不写死业务流程。

熵剪只做一件事：**当系统需要一个可审计、可签名、可留痕的「Yes / No / Hold」时，给出确定性的回答。**
