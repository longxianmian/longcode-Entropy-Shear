# CLAUDE.md — Claude Code 入口（熵剪旗舰版 P1 第一轮小修加固软约束）

> 本文件给 Claude Code 阅读。其他 AI 编程代理请读 [`AGENTS.md`](AGENTS.md)。

## 在做任何动作前，先读这两份文件

1. [`LONGMA_SOFT_GUARD.md`](LONGMA_SOFT_GUARD.md) — 人类可读总章（详细约束）。
2. [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json) — 机器可读锚点；与 MD 冲突时**以 JSON 为准**。

## 当前轮唯一目标

> 在不破坏 Core、不改变 P0 主契约、不引入新依赖的前提下，对熵剪旗舰版 P0 MVR 做第一轮小修加固（FLAGSHIP_P1_HARDENING_ROUND_1，限 H1–H5 共 5 项）。

## 本轮允许写入的代码 / 文档路径（且仅这些）

- `internal/flagship/**`
- `cmd/flagship-server/**`
- `tests/flagship/**`
- `docs/FLAGSHIP_P0_FREEZE_REPORT.md`
- `docs/FLAGSHIP_P0_DEV.md`
- `LONGMA_TASK_ANCHOR.json`

> 相比 P0 第一轮，本轮**不再允许**写入 `schemas/flagship/**` 与 `examples/flagship/**`（已随 `v0.4.0-flagship-p0` 冻结）。

其它一切只允许 Read / ls / grep。Core 引擎、对外契约、policies、ledger、产品 README / docs / SUPPORT、构建产物、`.gitignore`、AGENTS.md / CLAUDE.md / LONGMA_SOFT_GUARD.md、`schemas/flagship/**`、`examples/flagship/**` 全部禁止改动。

## 本轮允许做的（H1–H5，5 项；缺一不可，多一也不可）

- **H1** HTTP handler 测试覆盖（405、bad JSON、`DisallowUnknownFields`、`/health` GET / 非 GET）
- **H2** NO low-score 分支测试（`Score < T2` 且无硬冲突 → NO，`reason_code = FLAGSHIP_REASONER_NO_LOW_SCORE`）
- **H3** `state.Compute` 返回的 `NormalizedWeights` 必须是 map 副本，避免共享传入 map 的引用
- **H4** `output.NewRejectInstruction` 生成的 ID 必须纳入 `conflicting_items` 内容摘要，不得只依赖 `len(conflicts)`
- **H5** reasoner 中 `CanonicalJSON` 返回 error 不得静默忽略；出错时必须写入 trace 或 fallback digest reason

## 非目标（本轮严禁触达）

- 不接入真实 LLM
- 不开发 LLM Gateway
- 不开发完整龙码 AIOS
- 不开发知识资产化操作系统
- 不开发能力工具系统
- 不开发后台管理系统
- 不做多租户 SaaS
- 不做规则自动生成
- 不接入 Core ledger（含 `internal/ledger`、`ledger/`）
- 不接入 Core signature（含 `internal/signature`）
- 不做鉴权 / 限流 / TLS
- 不引入 jsonschema 等任何新依赖
- 不修改 `go.mod` / `go.sum`
- 不修改现有 Core /shear 接口
- 不修改 `policies/manifest.json`
- 不修改 `openapi.yaml`
- 不修改 `sdk/**`
- 不修改 `Dockerfile` / `docker-compose.yml`
- 不在本轮纳入 H1–H5 之外的任何加固 / 扩展 / 重构
- 不改变 P0 主契约（请求 / 响应 JSON 形状、verdict 大写 `YES` / `HOLD` / `NO`、矩阵数值、`λ μ T1 T2`、风险因子、`AlignmentTask` / `PermitToken` / `RejectInstruction` / `AuditRecord` 字段集）
- 不把龙码常数作为 Core 版正式能力发布

## Claude Code 专属操作守则

1. **写动作前先读真实文件**：用 `Read` 读取目标文件，用 `Bash: ls` / `Bash: find` 验证父目录真实存在。绝不依赖记忆中的目录结构。本轮 `internal/flagship`、`cmd/flagship-server`、`tests/flagship` 已在 `v0.4.0-flagship-p0` 落地，不得自行新建错误路径。
2. **引用接口前先 grep**：在引用 `internal/flagship/**` 下的函数签名、字段名、错误码、`schemas/` 中的字段、`openapi.yaml` 中的端点之前，用 `Bash: grep -rn` 在真实代码中验证。**严禁参照训练数据中的"熵剪 / 龙码常数 / 三态推理"印象**，一切以当前仓库文件为准；P0 主契约（矩阵数值、`λ μ T1 T2`、风险因子、verdict 大写、字段名、reason_code）以 `v0.4.0-flagship-p0` 实际仓库为准。
3. **不改 Core 引擎与对外契约**：禁止修改 `internal/{api,engine,errors,ledger,policy,schema,signature}/**`、`cmd/{server,verify-ledger,validate-policy,hash-policy}/**`、`schemas/*.json`（顶层）、`schemas/flagship/**`、`sdk/**`、`policies/**`、`integrations/**`、`openapi.yaml`、`tests/*.go`（顶层）、`ledger/**`、`examples/*.json`（顶层）、`examples/flagship/**`、`README.md` / `README_CN.md` / `SUPPORT.md`、`Dockerfile` / `docker-compose.yml`、`go.mod` / `go.sum`、`.gitignore`、`AGENTS.md` / `CLAUDE.md` / `LONGMA_SOFT_GUARD.md`。所有 P1 第一轮代码 / 测试 / docs 只能落入 `allowed_files`。
4. **不接入真实 LLM、不引入新依赖、不部署生产**：本轮全部走确定性逻辑，不接入任何 LLM；不 `go get` / 不升级 `go.mod` / 不 `go mod tidy` 引入新包（含 jsonschema 等运行期校验库）；不动 docker compose / 不 docker push / 不 `git tag`。
5. **不改 P0 主契约**：矩阵数值、`λ μ T1 T2`、风险因子、verdict 大写、`AlignmentTask` / `PermitToken` / `RejectInstruction` / `AuditRecord` 字段一律不动；`TestDefaultMatrixFixed` 的守门必须继续通过。
6. **越界即 HOLD**：写动作目标若落在 [`LONGMA_SOFT_GUARD.md`](LONGMA_SOFT_GUARD.md) 第 3 节的禁止路径内（含 `schemas/flagship/**` 与 `examples/flagship/**`），或要在 H1–H5 之外做加固，立即停下并把决策权交还给用户。`.claude/settings.json` 已对常见禁止路径配置了 `permissions.deny` 兜底，但这**只是兜底**，不是免责。
7. **不要自行 git 提交**：不自行 `git commit` / `git push` / `git tag` / 任何 release 动作；本轮所有产物都应停留在 working tree，由用户决定是否落库。
8. **完成必须给证据**：交付时必须输出 changed files、`git diff --stat`、`go build` / `go vet` / `go test ./tests/flagship/...` / `go test ./...` 的输出、P0 主契约检查、边界检查、H1–H5 各项状态、风险与 HOLD 项。少一项视为未完成。
9. **进入 P1 第二轮 / 旗舰版下一阶段前必须先换锚点**：进入 P1 第二轮或更后阶段前，必须先更新 [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json)，把 `target_constant.single_goal`、`scope_id`、`scope_kind`、`authorization_radius.allowed_files`、`forbidden_boundaries`、`flagship_p1_scope` 等字段切换为下一阶段的真实开发目标，并同步更新 `.claude/settings.json` 的 `permissions.deny`、本文件、`AGENTS.md`、`LONGMA_SOFT_GUARD.md`。**不得沿用本轮"P1 第一轮小修加固（H1–H5）"的授权半径**去做下一阶段开发；沿用即触发 AP1（目标漂移）+ AP2（行动越权），立即 HOLD。锚点未完成切换前，仅可在本轮 `allowed_files` 内写入；超出范围只允许 Read / ls / grep，不得对 Core / Pro / Flagship 任何代码进行 Write / Edit / MultiEdit。

## 五种反模式（速查）

| # | 反模式 | 触发即 HOLD |
|---|---|---|
| AP1 | 目标漂移 | 偏离上文唯一目标（含把内核扩张为 LLM Gateway / 多租户 SaaS / 规则自动生成 / 知识资产化 / 能力工具系统 / 后台管理 / 完整龙码 AIOS；或在 P1 第一轮中加入 H1–H5 之外的任何加固 / 扩展 / 重构；或回头改 P0 主契约） |
| AP2 | 行动越权 | 写禁止路径（含 `schemas/flagship/**` 与 `examples/flagship/**`）/ 自行 `git commit` / `push` / `tag` / 改外部依赖 / 接入真实 LLM / 改 Core /shear / 改 `policies/manifest.json` / 引入 jsonschema 等新依赖 / 对接 Core ledger / signature / 加鉴权 / 限流 / TLS |
| AP3 | 路径漂移 | 对真实不存在的目录或文件操作 |
| AP4 | 幻觉判断 | 引用未读过的代码、字段、文档；引用训练数据中的"熵剪 / 龙码常数 / 三态推理"印象；自行改矩阵数值或阈值 |
| AP5 | 嘴炮完成 | 完成无证据 |

## 边界声明

熵剪旗舰版 P1 第一轮小修加固**不是** Entropy Shear Core 的对外能力。它**不进入** `policies/`、**不修改** `policies/manifest.json`、**不被** `openapi.yaml` 与 `sdk/` 暴露、**不被** README / README_CN / SUPPORT 宣传为 Core 能力。Longma Constant 仍仅属于 Pro / Flagship 路线，详见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md) 与 [`README.md`](README.md) 的 Edition Boundary 段落。
