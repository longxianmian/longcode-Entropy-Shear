# 熵剪 Entropy Shear · P0

> 给 AI 与业务系统加一道**确定性裁决闸门**：输入 `policy + facts`，输出 `Yes / No / Hold` + 完整 trace + 不可篡改 signature + 追加式 ledger。

熵剪不是 AI、不是 LIOS 子模块、不是聊天机器人。它是一个独立的、确定性的三态裁决引擎。本仓库实现技术开发文档 v1.0 的 **P0 范围**。

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
  examples/
    cityone-policy.json           # CityOne 会员准入
    cityone-request.json          # 完整 POST /shear 请求
    ai-customer-service-policy.json
    bid-risk-policy.json
    agent-action-policy.json
  ledger/shear-chain.jsonl        # 账本（运行时生成）
  tests/                          # 单元测试 + handler 集成测试
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
