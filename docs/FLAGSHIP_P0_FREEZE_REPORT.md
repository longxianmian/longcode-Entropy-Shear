# 熵剪旗舰版 P0 冻结报告

> 本文件是熵剪旗舰版 P0 龙码三态逻辑推理内核 MVR 第一轮开发的正式冻结归档。
> 本报告**只描述闭源开发分支上的 P0 内核冻结事实**，不属于 Entropy Shear Core
> 对外能力，**不进入** `policies/`、**不修改** `policies/manifest.json`、
> **不被** `openapi.yaml` / `sdk/` 暴露、**不被** `README.md` / `README_CN.md` /
> `SUPPORT.md` 宣传为 Core 能力。详见
> [`../LONGMA_TASK_ANCHOR.json`](../LONGMA_TASK_ANCHOR.json) 与
> [`../LONGMA_SOFT_GUARD.md`](../LONGMA_SOFT_GUARD.md)。

签发日期：2026-05-06 · 主分支：`main` · 仓库：`entropy-shear`

---

## 1. 版本信息

| 项 | 值 |
|---|---|
| **main commit** | `2ded4fb744d54b4bbe211f236d8d2efaa7eeb799` |
| **commit subject** | `feat: add flagship p0 tristate reasoner (#1)` |
| **tag** | `v0.4.0-flagship-p0`（annotated tag，指向 `2ded4fb`，message: `flagship p0 tristate reasoner`） |
| **PR** | `#1`，源分支 `feat/flagship-p0-reasoner @ 7e45ff7` → `main` |
| **合并方式** | squash merge（原分支 `7e45ff7` 被压缩为 main 提交 `2ded4fb`，PR 编号 `(#1)` 写入 commit subject） |
| **冻结状态** | ✅ **已冻结**：working tree clean、`origin/main = HEAD`、tag 已落、anchor `flagship_p0_dev_progress.round_1.status = "implemented"` |

`git log --oneline --decorate -5` 实测：

```
2ded4fb (HEAD -> main, tag: v0.4.0-flagship-p0, origin/main) feat: add flagship p0 tristate reasoner (#1)
399a7c9 chore: add longma soft guard for flagship development
78c093c docs: define core/pro/flagship edition boundaries
a6ca60a docs: replace LICENSE with verbatim Apache 2.0 text from apache.org
b6e4b72 (tag: v0.3.0-p3) docs: relicense engine under Apache 2.0 and rewrite README for adoption
```

> 上一个 Core release 是 `v0.3.0-p3`；本轮 `v0.4.0-flagship-p0` 是 P0 内核的首个冻结点。

---

## 2. P0 已完成能力（12/12）

| # | 能力 | 落地位置 | 状态 |
|---|---|---|---|
| 1 | 多源输入标准结构 | [`internal/flagship/reasoner/types.go`](../internal/flagship/reasoner/types.go) `Input` + [`internal/flagship/mapper/mapper.go`](../internal/flagship/mapper/mapper.go) 五类原子结构 | ✅ |
| 2 | 五元映射器（Goal / Fact / Evidence / Constraint / Action） | [`internal/flagship/mapper/mapper.go`](../internal/flagship/mapper/mapper.go) `Map()` | ✅ |
| 3 | 原子校验规则执行器 | [`internal/flagship/rules/atoms.go`](../internal/flagship/rules/atoms.go) `Apply()`（7 条原子规则 + dedup） | ✅ |
| 4 | 内部状态计算 | [`internal/flagship/mapper/mapper.go`](../internal/flagship/mapper/mapper.go) 五个 `*State()` + `clamp01` | ✅ |
| 5 | 5×5 状态评估矩阵 | [`internal/flagship/state/matrix.go`](../internal/flagship/state/matrix.go) `DefaultMatrix`（R1 数值锁定） | ✅ |
| 6 | 五元干涉模型 Lite | [`internal/flagship/state/matrix.go`](../internal/flagship/state/matrix.go) `InteractionFactor` + `Relation`（R2 因子锁定） | ✅ |
| 7 | 加权冲突消解 | [`internal/flagship/state/matrix.go`](../internal/flagship/state/matrix.go) `Compute`（λ=0.20、μ=1.00、T1=0.70、T2=0.35） | ✅ |
| 8 | 三态状态机 YES / HOLD / NO | [`internal/flagship/state/matrix.go`](../internal/flagship/state/matrix.go) `Compute`（硬冲突 override → score 分支） | ✅ |
| 9 | HOLD 动态对齐任务生成 | [`internal/flagship/hold/align.go`](../internal/flagship/hold/align.go) `Generate`（HOLD 至少一条 + fallback 兜底） | ✅ |
| 10 | permit_token / reject_instruction | [`internal/flagship/output/token.go`](../internal/flagship/output/token.go) `NewPermitToken` / `NewRejectInstruction` | ✅ |
| 11 | AuditRecord 结构 | [`internal/flagship/output/audit.go`](../internal/flagship/output/audit.go) `NewAuditRecord` + SHA-256 digest | ✅ |
| 12 | 最小测试用例 | [`tests/flagship/`](../tests/flagship/) 三组测试 18 testfunc 全 PASS | ✅ |

附 HTTP 服务（[`cmd/flagship-server/main.go`](../cmd/flagship-server/main.go)）：`POST /flagship/reason`、`GET /health`、默认监听 `:9090`、环境变量 `ENTROPY_SHEAR_FLAGSHIP_ADDR` 覆盖。

---

## 3. 明确未做边界（本轮严守的不变量）

| 边界 | 状态 |
|---|---|
| 不接入真实 LLM | ✅ 全确定性逻辑，零外部模型调用 |
| 不做 LLM Gateway | ✅ |
| 不做完整龙码 AIOS | ✅ |
| 不做知识资产化操作系统 | ✅ |
| 不做能力工具系统 | ✅ |
| 不做后台管理系统 | ✅ |
| 不做多租户 SaaS | ✅ HTTP 单实例无租户隔离 |
| 不做规则自动生成 | ✅ 7 条原子规则全部硬编码 |
| 不接 Core ledger | ✅ AuditRecord 仅生成结构、不写盘、零 `internal/ledger` 引用 |
| 不改 Core `/shear` | ✅ `internal/api` / `internal/engine` 零修改 |
| 不改 `policies/manifest.json` | ✅ `policies/**` 整树零修改 |

补充：内核**不进入** `policies/`、**不出现在** `openapi.yaml` / `sdk/`、**不被** `README.md` / `README_CN.md` / `SUPPORT.md` 宣传为 Core 能力。

---

## 4. 测试与验收结果（在 `main @ 2ded4fb` 实测）

```
$ git status --short
(空)                                                            ✓ 工作树干净

$ go build ./internal/flagship/... ./cmd/flagship-server/...
(无输出)                                                        ✓

$ go vet ./internal/flagship/... ./cmd/flagship-server/... ./tests/flagship/...
(无输出)                                                        ✓

$ go test ./tests/flagship/...
ok  	entropy-shear/tests/flagship	1.168s                      ✓

$ go test ./...
ok  	entropy-shear/tests          (cached)   ← Core 回归全绿
ok  	entropy-shear/tests/flagship (cached)   ← Flagship 全绿
                                                                ✓

$ python3 -c "json.load 5 个 flagship JSON 文件"
flagship JSON: OK                                               ✓
```

flagship 子模块覆盖：18 testfunc / 28 子用例全 PASS。覆盖维度包括矩阵数值守门（`TestDefaultMatrixFixed`）、weights 回落（`TestResolveWeightsFallback`）、risk boost 归一化（`TestApplyRiskBoostNormalizes` / `TestComputeNormalizedWeightsSumOne`）、HOLD 区间（`TestComputeHoldBand`）、确定性 audit（`TestReasonAuditDeterminism`）、HTTP 往返（`TestHTTPRoundTrip`）、YES / HOLD / NO 三组端到端用例（`TestReasonExample{Yes,Hold,NoHardConstraint}`）。

JSON 校验范围：`schemas/flagship/reason-input.schema.json`、`schemas/flagship/reason-output.schema.json`、`examples/flagship/reason-{yes,hold,no}-request.json`。

---

## 5. Core 隔离结论

**完全隔离，未触碰任何 Core 资产。**

import 图扫描（`grep -rn` 搜索 `entropy-shear/internal/{api,engine,policy,schema,ledger,signature}` 在 `internal/flagship/` + `cmd/flagship-server/` + `tests/flagship/`）：**零命中**。flagship 代码仅依赖：

- Go 标准库
- `entropy-shear/internal/flagship/*`（自身子包）

Core 路径写入审计：`policies/**`、`internal/{api,engine,errors,ledger,policy,schema,signature}/**`、`cmd/{server,verify-ledger,validate-policy,hash-policy}/**`、`schemas/{ledger-record,policy,shear-request,shear-response}.schema.json`、`sdk/**`、`integrations/**`、`openapi.yaml`、顶层 `examples/*.json` / `tests/*.go`、`README.md` / `README_CN.md` / `SUPPORT.md`、`go.mod` / `go.sum`、`Dockerfile` / `docker-compose.yml`、`.gitignore`、`ledger/**` —— 全部**未触**。

锚点 `flagship_p0_dev_progress.round_1` 自记字段全部 `false`：`core_touched=false`、`policies_manifest_touched=false`、`openapi_or_sdk_touched=false`、`external_llm_called=false`、`new_dependency_introduced=false`、`go_module_touched=false`，与实测一致。

---

## 6. P1 HOLD 项（冻结期不动；进入 P1 须先做新一轮锚点滚动）

| # | HOLD 项 | 性质 | 触发位置 |
|---|---|---|---|
| H1 | HTTP handler 测试覆盖（405 / 400 / `DisallowUnknownFields` / `/health` 分支） | 覆盖空缺 | [`tests/flagship/reasoner_e2e_test.go`](../tests/flagship/reasoner_e2e_test.go) `TestHTTPRoundTrip` 重写了 mux，未复用 [`cmd/flagship-server/main.go`](../cmd/flagship-server/main.go) 的 `reasonHandler` |
| H2 | NO low-score 分支测试（`score < T2` 且无硬冲突 → `FLAGSHIP_REASONER_NO_LOW_SCORE`） | 覆盖空缺 | [`internal/flagship/reasoner/reasoner.go`](../internal/flagship/reasoner/reasoner.go) `defaultRemediation` 默认分支 |
| H3 | `Computation.NormalizedWeights` 应返回 map 副本而非引用 | 维护风险 | [`internal/flagship/state/matrix.go`](../internal/flagship/state/matrix.go) `Compute` |
| H4 | `RejectInstruction.ID` 应包含 `conflicting_items` 内容指纹（当前仅 `len(conflicts)`） | 同请求异内容碰撞 | [`internal/flagship/output/token.go`](../internal/flagship/output/token.go) `NewRejectInstruction` |
| H5 | 真正的 JSON Schema runtime 校验（不仅 `json.load` 语法解析） | 强契约 | 需新增依赖 — P0 禁 |
| H6 | `CanonicalJSON` error 应写 trace 而非静默忽略 | 反模式 | [`internal/flagship/reasoner/reasoner.go`](../internal/flagship/reasoner/reasoner.go) 第 51–56 行 `_, _` |
| H7 | 是否对接 Core ledger / signature 形成证据链 | 战略决策 | R6 v0 显式不接 |
| H8 | HTTP 是否加鉴权 / 限流 / TLS | 生产化决策 | P0 是裸闭环 |
| H9 | 是否进入 LLM Gateway 阶段（与本内核拼装） | 战略决策 | non_goals 显式不做 |

> 所有 H 项都属于 P1+。本冻结期不应动；若要解任一项，必须先新建 [`LONGMA_TASK_ANCHOR.json`](../LONGMA_TASK_ANCHOR.json) 的 P1 锚点（更新 `single_goal` / `scope_id` / `authorization_radius.allowed_files` / `forbidden_boundaries` / `flagship_p1_scope`），并同步配套软约束文档（`LONGMA_SOFT_GUARD.md` / `CLAUDE.md` / `AGENTS.md` / `.claude/settings.json`）。**在新锚点未完成前，本仓库不得对任何 Core / Pro / Flagship 代码做 Write / Edit / MultiEdit。**

---

## 7. 下一步建议

**最小下一步**：由用户决定是否启动**熵剪旗舰版 P1 锚点滚动**——更新 [`LONGMA_TASK_ANCHOR.json`](../LONGMA_TASK_ANCHOR.json) 的 `single_goal` / `scope_id` / `authorization_radius.allowed_files` / `flagship_p1_scope`，并同步 `LONGMA_SOFT_GUARD.md` / `CLAUDE.md` / `AGENTS.md` / `.claude/settings.json`。在新锚点未完成前，本仓库一律保持只读。
