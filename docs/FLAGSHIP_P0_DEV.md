# 熵剪旗舰版 P0 — 龙码三态逻辑推理内核（开发笔记）

> 本文件只描述 **闭源开发分支** 上的 P0 内核实现细节。它**不是** Entropy
> Shear Core 的对外能力。本内核**不进入** `policies/`、**不修改**
> `policies/manifest.json`、**不被** `openapi.yaml` / `sdk/` 暴露、**不被**
> `README.md` / `README_CN.md` / `SUPPORT.md` 宣传为 Core 能力。详见
> [`LONGMA_TASK_ANCHOR.json`](../LONGMA_TASK_ANCHOR.json) 与
> [`LONGMA_SOFT_GUARD.md`](../LONGMA_SOFT_GUARD.md)。

## 1. 目标

在不破坏现有 Entropy Shear Core 的前提下，提供熵剪旗舰版 P0 的**最小可运行
推理器**（Minimum Verifiable Reasoner）。它接收多源输入，将其映射到龙码五元
（Goal / Fact / Evidence / Constraint / Action），按 5×5 矩阵 + 加权冲突消解
计算分数，并以 YES / HOLD / NO 三态返回。本轮全部走**确定性逻辑**，不接入
任何 LLM。

## 2. 模块拓扑

```
internal/flagship/
  mapper/    多源输入 → 五元状态向量
  rules/     原子校验规则 + 硬冲突识别
  state/     5×5 矩阵、干涉模型 Lite、Score 公式、FSM
  hold/      HOLD 时生成 AlignmentTask
  output/    PermitToken / RejectInstruction / AuditRecord
  reasoner/  对外 API：types.go 聚合类型；reasoner.go 编排

cmd/flagship-server/  HTTP 服务（POST /flagship/reason，端口默认 :9090）
schemas/flagship/     输入 / 输出 JSON Schema
examples/flagship/    YES / HOLD / NO 三组示例请求
tests/flagship/       mapper / state / e2e 测试
```

依赖方向：`reasoner → {mapper, rules, state, hold, output}`；
`hold → {mapper, rules}`；其它包均为叶子，互不引入循环。

## 3. R1–R8 决策落地位置

| 决策 | 内容 | 落地位置 |
|---|---|---|
| **R1** 5×5 矩阵 v0（行=Ei 对 列=Ej 的影响） | 固定数值，不得改 | [`internal/flagship/state/matrix.go`](../internal/flagship/state/matrix.go) — `DefaultMatrix` |
| **R2** 五元干涉模型 Lite | `relation = min(state(Ei),state(Ej)) * factor`；正向促进 1.00 / 约束制衡 0.80 / 过度压制 1.20（保留）/ 反向失衡 1.30（保留） / 无影响 0；P0 按矩阵符号决定基础关系 | [`internal/flagship/state/matrix.go`](../internal/flagship/state/matrix.go) — `Factor*` 常量、`InteractionFactor`、`Relation` |
| **R3** 加权冲突消解 v0 | `Score = ΣWi*state(Ei) + λ*ΣM[i][j]*relation(Ei,Ej) − μ*HardPenalty`；λ=0.20、μ=1.00、T1=0.70、T2=0.35；默认权重 Goal/Fact/Evidence/Constraint/Action=0.25/0.20/0.25/0.20/0.10；风险 r ∈ {0.25, 0.50, 0.75, 1.00} 触发 Evidence_boost / Constraint_boost / Action_decay 后归一化；非法 weights 回落默认且写入 trace | [`internal/flagship/state/matrix.go`](../internal/flagship/state/matrix.go) — `Lambda` `Mu` `T1` `T2` `DefaultWeights` `RiskFactors` `ResolveWeights` `ApplyRiskBoost` `Compute` |
| **R4** AlignmentTask 字段 | `id / target_element / reason_code / gap / required_action / expected_evidence_kind / prompt / priority`；`priority ∈ {low, medium, high}`；HOLD 至少 1 条 | [`internal/flagship/hold/align.go`](../internal/flagship/hold/align.go) — `AlignmentTask`、`Generate` |
| **R5** PermitToken / RejectInstruction 字段 | PermitToken: `id, verdict, scope, valid_until, reason_code, audit_id`；P0 不签名、不接 ledger。RejectInstruction: `id, verdict, reason_code, conflicting_items, remediation_steps, audit_id`；`remediation_steps` 必须是字符串数组 | [`internal/flagship/output/token.go`](../internal/flagship/output/token.go) |
| **R6** AuditRecord v0 | `audit_id / request_id / input_digest / five_element_digest / triggered_rule_ids / matrix_digest / state_machine_result / permit_token_id? / reject_instruction_id? / timestamp`；P0 只生成结构，不写盘 JSONL，不接 Core ledger；digest 用 SHA-256 + Go canonical JSON | [`internal/flagship/output/audit.go`](../internal/flagship/output/audit.go) |
| **R7** 端口与路由 | 默认 `:9090`；环境变量 `ENTROPY_SHEAR_FLAGSHIP_ADDR`；`POST /flagship/reason`；可选 `GET /health`，仅在 flagship-server 内 | [`cmd/flagship-server/main.go`](../cmd/flagship-server/main.go) |
| **R8** Verdict 字面值 | 全大写 `YES` / `HOLD` / `NO`，不复用 Core 的 Yes/Hold/No | [`internal/flagship/state/matrix.go`](../internal/flagship/state/matrix.go) — `Verdict*` 常量；输出层亦使用大写 |

## 4. FSM 规则

```
if HasHardConflict           → NO   (override score)
elif Score >= T1 (=0.70)     → YES
elif Score >= T2 (=0.35)     → HOLD
else                         → NO
```

硬冲突触发条件（任意一条满足即记 hard penalty 并 force NO）：

- 约束 `severity == "hard"` 且 `satisfied != true`；或
- 约束 `kind ∈ {"permission", "forbid", "governance"}` 且 `satisfied != true`，
  无视 severity（实现 R3 中"权限拒绝 / 明确禁止 / 治理主体拒绝 → NO"）。

## 5. 多源输入 → 五元映射规则

| 元 | 规则 |
|---|---|
| Goal | 缺失或 id 空 → 0；只有 id → 0.5；id+description → 1.0 |
| Fact | 无 fact → 0；否则 (key 与 value 都非空的 fact 数 / 总数) |
| Evidence | 无 evidence → 0；否则 mean(`confidence * (1.0 if verified else 0.5)`) |
| Constraint | 无约束 → 1.0（不受阻）；否则 satisfied / 总数（`Satisfied==nil` 视为未满足） |
| Action | 无 action → 0；否则 mean(`1 - cost`，clamp 到 [0,1]） |

实现见 [`internal/flagship/mapper/mapper.go`](../internal/flagship/mapper/mapper.go)。

## 6. HTTP 端点

```
POST /flagship/reason
Content-Type: application/json
Body: 见 schemas/flagship/reason-input.schema.json

GET  /health
Response: {"status":"ok","module":"flagship"}
```

启动：

```sh
go run ./cmd/flagship-server                    # 监听 :9090
ENTROPY_SHEAR_FLAGSHIP_ADDR=:18080 go run ./cmd/flagship-server
```

`/health` 仅在 flagship-server 内响应，不影响 Core `/health`。

## 7. 示例

| 用途 | 文件 | 期望 verdict |
|---|---|---|
| YES | [`examples/flagship/reason-yes-request.json`](../examples/flagship/reason-yes-request.json) | YES（带 PermitToken） |
| HOLD | [`examples/flagship/reason-hold-request.json`](../examples/flagship/reason-hold-request.json) | HOLD（带 AlignmentTasks） |
| NO | [`examples/flagship/reason-no-request.json`](../examples/flagship/reason-no-request.json) | NO（带 RejectInstruction） |

```sh
curl -s -X POST http://localhost:9090/flagship/reason \
     -H 'Content-Type: application/json' \
     --data @examples/flagship/reason-yes-request.json | jq .
```

## 8. 测试

```sh
go build ./internal/flagship/... ./cmd/flagship-server/...
go test  ./tests/flagship/...
go test  ./...
```

JSON 校验：

```sh
python3 -c "import json; [json.load(open(f)) for f in [
  'schemas/flagship/reason-input.schema.json',
  'schemas/flagship/reason-output.schema.json',
  'examples/flagship/reason-yes-request.json',
  'examples/flagship/reason-hold-request.json',
  'examples/flagship/reason-no-request.json'
]]; print('flagship JSON: OK')"
```

## 9. 边界

- **不修改** Core 引擎（`internal/{api,engine,errors,ledger,policy,schema,signature}`），不动 Core `/shear`、不动 `policies/manifest.json`。
- **不引入新依赖**，`go.mod` 未触碰；本内核只用标准库。
- **不接入真实 LLM**，全部走确定性逻辑。
- **不写入** `policies/`、不出现在 `openapi.yaml` / `sdk/`、不被 `README` 宣传为 Core 能力。

下一阶段（旗舰版 P1 或更后）必须先做新一轮锚点滚动；不得沿用 P0 的授权半径。
