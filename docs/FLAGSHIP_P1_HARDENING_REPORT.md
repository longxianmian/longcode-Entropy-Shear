# 熵剪旗舰版 P1 第一轮加固报告

> 本文件是熵剪旗舰版 P1 第一轮小修加固（H1–H5）的正式冻结归档。
> 本报告**只描述闭源开发分支上的 P1 第一轮冻结事实**，不属于 Entropy Shear
> Core 对外能力，**不进入** `policies/`、**不修改** `policies/manifest.json`、
> **不被** `openapi.yaml` / `sdk/` 暴露、**不被** `README.md` / `README_CN.md` /
> `SUPPORT.md` 宣传为 Core 能力。详见
> [`../LONGMA_TASK_ANCHOR.json`](../LONGMA_TASK_ANCHOR.json) 与
> [`../LONGMA_SOFT_GUARD.md`](../LONGMA_SOFT_GUARD.md)。

签发日期：2026-05-06 · 主分支：`main` · 仓库：`entropy-shear`

---

## 1. 版本信息

| 项 | 值 |
|---|---|
| **main commit** | `a6d6ed0469588a00c111888366f514c3740c9bca` |
| **commit subject** | `fix: harden flagship p1 reasoner edges (#3)` |
| **tag** | `v0.4.1-flagship-p1-hardening`（annotated tag，指向 `a6d6ed0`，message: `flagship p1 hardening round 1`） |
| **PR** | `#3`，源分支 `feat/flagship-p1-hardening` → `main` |
| **合并方式** | squash merge |
| **基线** | `v0.4.0-flagship-p0`（commit `2ded4fb`，PR `#1`） |
| **冻结状态** | ✅ **已冻结**：working tree clean、`origin/main = HEAD`、tag 已落、anchor `flagship_p1_hardening_freeze.status = "frozen"` |

`git log --oneline --decorate -5` 实测：

```
a6d6ed0 (HEAD -> main, tag: v0.4.1-flagship-p1-hardening, origin/main) fix: harden flagship p1 reasoner edges (#3)
a184b0a docs: add flagship p0 freeze report (#2)
2ded4fb (tag: v0.4.0-flagship-p0) feat: add flagship p0 tristate reasoner (#1)
399a7c9 chore: add longma soft guard for flagship development
78c093c docs: define core/pro/flagship edition boundaries
```

> P0 冻结点 `v0.4.0-flagship-p0` → P1 第一轮冻结点 `v0.4.1-flagship-p1-hardening`，纯小修加固，未引入新主能力。

---

## 2. P1 第一轮已完成加固项（H1–H5，5/5）

| ID | 加固项 | 落地位置 | 状态 |
|---|---|---|---|
| **H1** | HTTP handler 测试覆盖（405、bad JSON、`DisallowUnknownFields`、`/health` GET / 非 GET） | [`cmd/flagship-server/main_test.go`](../cmd/flagship-server/main_test.go) — `TestReasonHandlerRejectsGET` / `TestReasonHandlerBadJSON` / `TestReasonHandlerDisallowsUnknownFields` / `TestReasonHandlerHappyMinimal` / `TestHealthHandlerGET` / `TestHealthHandlerRejectsPOST`，复用真实 `reasonHandler` / `healthHandler` | ✅ |
| **H2** | NO low-score 分支测试（`Score < T2` 且无硬冲突 → NO，`reason_code = FLAGSHIP_REASONER_NO_LOW_SCORE`） | [`tests/flagship/p1_hardening_test.go`](../tests/flagship/p1_hardening_test.go) `TestReasonNoLowScoreReasonCode`（最小 Input：状态 = (0,0,0,1.0,0)，Score=0.20<T2=0.35） | ✅ |
| **H3** | `state.Compute` 返回的 `NormalizedWeights` 必须是副本，不得共享传入 map 引用 | [`internal/flagship/state/matrix.go`](../internal/flagship/state/matrix.go) `Compute` 改 `NormalizedWeights: copyWeights(weights)`；[`internal/flagship/reasoner/reasoner.go`](../internal/flagship/reasoner/reasoner.go) `Output.NormalizedWeights` 改用 `comp.NormalizedWeights`；测试 `TestComputeReturnsWeightsCopy` + `TestReasonOutputNormalizedWeightsIndependent` | ✅ |
| **H4** | `output.NewRejectInstruction` 生成 ID 必须纳入 `conflicting_items` 内容指纹，不得只依赖 `len(conflicts)` | [`internal/flagship/output/token.go`](../internal/flagship/output/token.go) `NewRejectInstruction` 用 `strings.Join(conflicting, "\x1f")` 替换 `strconv.Itoa(len(conflicting))`；测试 `TestRejectInstructionIDIncludesConflictContent` 覆盖（a）不同内容→不同 ID（b）顺序变化→不同 ID（c）相同输入→相同 ID | ✅ |
| **H5** | reasoner 中 `CanonicalJSON` 返回 error 不得静默忽略；出错时必须写入 trace 或 fallback digest reason | [`internal/flagship/reasoner/reasoner.go`](../internal/flagship/reasoner/reasoner.go) 新增 `canonicalOrFallback` 助函数，error 时 trace 追加 `canonical_json_error[<kind>]: <err>; using fallback digest`，并返回 per-kind 哨兵字节 `FLAGSHIP_CANONICAL_JSON_FALLBACK:<kind>` 作 digest 输入；测试 `TestReasonCanonicalJSONErrorTrace` 用 `chan int` 在 `Metadata` 触发 input-kind 失败 | ✅ |

**P0 主契约守门**：本轮**未改变** P0 主契约任一字段、矩阵数值、`λ μ T1 T2`、风险因子、verdict 大写、`AlignmentTask` / `PermitToken` / `RejectInstruction` / `AuditRecord` 字段集；`TestDefaultMatrixFixed` 继续通过。

---

## 3. 明确未做边界（P1 第一轮严守的不变量）

| 边界 | 状态 |
|---|---|
| 不接入真实 LLM | ✅ 全确定性逻辑 |
| 不做 LLM Gateway | ✅ |
| 不做完整龙码 AIOS | ✅ |
| 不接 Core ledger（`internal/ledger`、`ledger/`） | ✅ |
| 不接 Core signature（`internal/signature`） | ✅ |
| 不做鉴权 | ✅ HTTP 仍是裸闭环 |
| 不做限流 | ✅ |
| 不做 TLS | ✅ |
| 不引入新依赖（含 `jsonschema` 等） | ✅ 仅标准库；`go.mod` / `go.sum` 零修改 |
| 不修改 Core `/shear`（`internal/api`、`internal/engine`、`internal/schema`、`internal/policy`、`internal/errors`） | ✅ |
| 不修改 `policies/manifest.json` | ✅ `policies/**` 整树未触 |
| 不修改 `go.mod` / `go.sum` | ✅ |
| 不修改 `Dockerfile` / `docker-compose.yml` | ✅ |
| 不修改 `openapi.yaml` | ✅ |
| 不修改 `sdk/**` | ✅ |
| 不修改 `schemas/flagship/**` 与 `examples/flagship/**`（P0 已冻结） | ✅ |
| 不在 H1–H5 之外做加固 / 扩展 / 重构 | ✅ |
| 不改变 P0 主契约 | ✅ |

---

## 4. 测试与验收结果（在 `main @ a6d6ed0` 实测）

```
$ git status --short
(空)                                                            ✓ 工作树干净

$ go build ./internal/flagship/... ./cmd/flagship-server/...
(无输出)                                                        ✓

$ go vet ./internal/flagship/... ./cmd/flagship-server/... ./tests/flagship/...
(无输出)                                                        ✓

$ go test ./tests/flagship/...
ok  	entropy-shear/tests/flagship	2.255s                      ✓
（含原 P0 18 testfunc + 新增 P1 5 testfunc）

$ go test ./...
ok  	entropy-shear/cmd/flagship-server	0.793s   ← H1 6 项 PASS
ok  	entropy-shear/tests          (cached)        ← Core 回归全绿
ok  	entropy-shear/tests/flagship (cached)        ← Flagship 全绿
                                                                ✓
```

新增测试（11 个 testfunc 全 PASS）：

```
=== H1（cmd/flagship-server/main_test.go）
--- PASS: TestReasonHandlerRejectsGET
--- PASS: TestReasonHandlerBadJSON
--- PASS: TestReasonHandlerDisallowsUnknownFields
--- PASS: TestReasonHandlerHappyMinimal
--- PASS: TestHealthHandlerGET
--- PASS: TestHealthHandlerRejectsPOST

=== H2/H3/H4/H5（tests/flagship/p1_hardening_test.go）
--- PASS: TestReasonNoLowScoreReasonCode
--- PASS: TestComputeReturnsWeightsCopy
--- PASS: TestReasonOutputNormalizedWeightsIndependent
--- PASS: TestRejectInstructionIDIncludesConflictContent
--- PASS: TestReasonCanonicalJSONErrorTrace
```

JSON 冻结复检：`schemas/flagship/**` 与 `examples/flagship/**` 完全未触；`flagship JSON: OK`。

---

## 5. Core 隔离结论

**完全隔离，未触碰任何 Core 资产。**

import 图扫描（`grep -rn` `entropy-shear/internal/{api,engine,policy,schema,ledger,signature}` against `internal/flagship/` + `cmd/flagship-server/` + `tests/flagship/`）：**零命中**。flagship 代码仅依赖：

- Go 标准库
- `entropy-shear/internal/flagship/*`（自身子包）

Core 路径写入审计：`policies/**`、`internal/{api,engine,errors,ledger,policy,schema,signature}/**`、`cmd/{server,verify-ledger,validate-policy,hash-policy}/**`、`schemas/{ledger-record,policy,shear-request,shear-response}.schema.json`、`schemas/flagship/**`、`sdk/**`、`integrations/**`、`openapi.yaml`、顶层 `examples/*.json` 与 `examples/flagship/**`、顶层 `tests/*.go`、`README.md` / `README_CN.md` / `SUPPORT.md`、`go.mod` / `go.sum`、`Dockerfile` / `docker-compose.yml`、`.gitignore`、`ledger/**`、`AGENTS.md` / `CLAUDE.md` / `LONGMA_SOFT_GUARD.md` —— 全部**未触**。

锚点 `flagship_p0_freeze` 与 `flagship_p0_dev_progress` 历史指针保留不动，本轮 P1 加固未回头改 P0 任何业务逻辑。

---

## 6. P1 后续 HOLD 项（冻结期不动；进入下一轮须先做新一轮锚点滚动）

| # | HOLD 项 | 性质 | 备注 |
|---|---|---|---|
| H6 | 真正的 JSON Schema runtime 校验（不仅 `json.load` 语法解析） | 强契约 | 需新增 `santhosh-tekuri/jsonschema` 类依赖 — P1 第一轮禁；P1 第二轮或 P2 决策 |
| H7 | 是否对接 Core ledger（`internal/ledger`） / Core signature（`internal/signature`）形成证据链 | 战略决策 | R6 v0 显式不接；进入需 P2 锚点滚动 |
| H8 | HTTP 鉴权 / 限流 / TLS（生产化） | 生产化决策 | P1 第一轮显式不做；进入生产前必做 |
| H9 | 是否进入 LLM Gateway / 编程大脑网关阶段 | 战略决策 | non_goals 显式不做；属于完整龙码 AIOS 路线 |
| H10 | 是否进入 P2 锚点滚动（替换 single_goal、scope_id、allowed_files、forbidden_paths、`flagship_p2_scope`） | 流程门槛 | 任何 H6–H9 之一启动前都必须先做 |

> 所有 H 项都属于 P1 第二轮或 P2+。本冻结期不应动；若要解任一项，必须先新建 [`../LONGMA_TASK_ANCHOR.json`](../LONGMA_TASK_ANCHOR.json) 的下一轮锚点（更新 `single_goal` / `scope_id` / `authorization_radius.allowed_files` / `forbidden_boundaries` / `flagship_p2_scope`），并同步配套软约束文档（`LONGMA_SOFT_GUARD.md` / `CLAUDE.md` / `AGENTS.md` / `.claude/settings.json`）。**在新锚点未完成前，本仓库不得对任何 Core / Pro / Flagship 代码做 Write / Edit / MultiEdit。**

---

## 7. 下一步建议

**最小下一步**：由用户决定是否启动**熵剪旗舰版 P1 第二轮 / 或 P2 锚点滚动**——只更新 [`../LONGMA_TASK_ANCHOR.json`](../LONGMA_TASK_ANCHOR.json) 的 `single_goal` / `scope_id` / `authorization_radius.allowed_files` / `flagship_p2_scope`（或 `flagship_p1_round_2_scope`），并同步 `LONGMA_SOFT_GUARD.md` / `CLAUDE.md` / `AGENTS.md` / `.claude/settings.json`。在新锚点未完成前，本仓库一律保持只读。
