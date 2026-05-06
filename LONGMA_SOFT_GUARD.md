# 熵剪旗舰版 P1 第一轮小修加固软约束（LONGMA_SOFT_GUARD）

> 本文件不是熵剪 Core 的产品功能。它只在 AI 编程代理（Claude Code / Codex 等）对熵剪旗舰版 P0 MVR 做 P1 第一轮小修加固时生效，用于约束开发过程的目标边界与执行边界，治理目标漂移、行动越权、路径漂移、幻觉判断与嘴炮完成。

## 0. 边界声明（必读）

熵剪旗舰版 P1 第一轮小修加固：

- **不属于** Entropy Shear Core 的对外能力。
- **不进入** `policies/`、**不修改** `policies/manifest.json`、**不修改** `tests/policy_pack_test.go`。
- **不被** `openapi.yaml` 与 `sdk/` 暴露。
- **不被** `README.md` / `README_CN.md` / `SUPPORT.md` / `docs/PRODUCT_MATRIX.md` 等产品文档宣传为 Core 能力。
- **不接入** 任何真实 LLM；本轮内核全部走确定性逻辑。
- **不改变 P0 主契约**：请求 / 响应 JSON 形状、verdict 大写 `YES` / `HOLD` / `NO`、矩阵数值、`λ μ T1 T2`、风险因子、`AlignmentTask` / `PermitToken` / `RejectInstruction` / `AuditRecord` 字段集，全部以 `v0.4.0-flagship-p0` 实际仓库为准，本轮不得改、不得猜。
- 与 [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json) 互为正副本：本文件给人读，锚点 JSON 给机器读，二者冲突时**以锚点 JSON 为准**。

Longma Constant 仍仅属于 Pro / Flagship 路线，详见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md) 与 [`README.md`](README.md) 的 Edition Boundary 段落。

## 1. 唯一目标（target constant）

> **在不破坏 Core、不改变 P0 主契约、不引入新依赖的前提下，对熵剪旗舰版 P0 MVR 做第一轮小修加固（FLAGSHIP_P1_HARDENING_ROUND_1，限 H1–H5 共 5 项）。**

本轮允许触达的路径**只**有：

1. `internal/flagship/**`
2. `cmd/flagship-server/**`
3. `tests/flagship/**`
4. `docs/FLAGSHIP_P0_FREEZE_REPORT.md`
5. `docs/FLAGSHIP_P0_DEV.md`
6. `LONGMA_TASK_ANCHOR.json`

> 注：相比 P0 第一轮，本轮**不再允许**写入 `schemas/flagship/**` 与 `examples/flagship/**`（已随 `v0.4.0-flagship-p0` 冻结）。这是 P1 第一轮相对 P0 第一轮的关键收紧。

非目标（本轮严禁触达）：

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
- 不改变 P0 主契约
- 不把龙码常数作为 Core 版正式能力发布

## 2. 授权半径

允许写入的文件 = 第 1 节六条路径，且**仅这六条**。其它一切只允许 Read / ls / grep（用于验证），**不允许写**。

落入 `allowed_files` 内的合规子能力实现范围（限 P1 第一轮 5 项 H1–H5）：

- **H1** HTTP handler 测试覆盖（405、bad JSON、`DisallowUnknownFields`、`/health`）
- **H2** NO low-score 分支测试（Score < T2 且无硬冲突 → NO，`reason_code = FLAGSHIP_REASONER_NO_LOW_SCORE`）
- **H3** `state.Compute` 返回的 `NormalizedWeights` 必须是 map 副本，避免共享传入 map 的引用
- **H4** `output.NewRejectInstruction` 生成的 ID 必须纳入 `conflicting_items` 内容摘要，不得只依赖 `len(conflicts)`
- **H5** reasoner 中 `CanonicalJSON` 返回 error 不得静默忽略；出错时必须写入 trace 或 fallback digest reason

H1–H5 之外的任何加固 / 扩展 / 重构都属于 AP1（目标漂移），立即 HOLD。

## 3. 禁止边界

下列路径**禁止**修改：

- `internal/{api,engine,errors,ledger,policy,schema,signature}/**`（Core 引擎）
- `cmd/{hash-policy,server,validate-policy,verify-ledger}/**`（Core 命令行）
- `schemas/*.json`（顶层 schema）
- `schemas/flagship/**`（**P1 新增收紧**：P0 已冻结）
- `sdk/**`
- `policies/**`（含 `policies/manifest.json`）
- `integrations/**`
- `docs/{AGENT_TOOL_GATE_GUIDE,INTEGRATION_GUIDE,P1_RELEASE_CHECKLIST,POLICY_PACK_GUIDE,PRODUCT_MATRIX,WHITEPAPER}.md`（既有 Core 产品文档；`docs/FLAGSHIP_P0_DEV.md` 与 `docs/FLAGSHIP_P0_FREEZE_REPORT.md` 除外）
- `tests/*.go`（顶层 Core 测试；`tests/flagship/**` 除外）
- `ledger/**`
- `examples/*.json`（顶层既有示例）
- `examples/flagship/**`（**P1 新增收紧**：P0 已冻结）
- `README.md`、`README_CN.md`、`SUPPORT.md`
- `openapi.yaml`
- `go.mod`、`go.sum`
- `Dockerfile`、`docker-compose.yml`
- `AGENTS.md`、`CLAUDE.md`、`LONGMA_SOFT_GUARD.md`、`.gitignore`

下列动作**禁止**自行执行（必须用户当轮显式指令）：

- `git commit` / `git push` / `git tag` / 任何 release 动作
- 改动 CI、外部依赖、镜像仓库、域名、密钥
- 安装新包、`go get`、`go mod tidy` 引入新包、改动构建产物、引入 jsonschema 等运行期校验库
- 接入真实 LLM、调用任何外部模型 API
- 对接 Core ledger / Core signature
- 加鉴权 / 限流 / TLS
- 改变 P0 主契约
- 在本轮纳入 H1–H5 之外的加固
- 删除或重写禁止路径下的任何已有文件

## 4. 治理的五种反模式

### AP1 目标漂移
**定义**：在执行过程中悄悄扩展、迁移或替换本轮目标。
**典型表现**：把"P1 第一轮小修加固（H1–H5）"扩张成"实现完整龙码 AIOS / LLM Gateway / 多租户 SaaS / 规则自动生成 / 知识资产化操作系统 / 能力工具系统 / 后台管理"；或在 P1 第一轮里顺手加 H1–H5 之外的加固（如鉴权 / 限流 / TLS / 接入 ledger / 改 schema / 改 examples / 调矩阵数值）；或顺手把内核能力以 policy pack 形式塞进 Core。
**守则**：任何越出第 1 节单一目标或 H1–H5 范围的动作必须先停下，把"我想做 X，是否授权"显式问出来。

### AP2 行动越权
**定义**：未经用户当轮显式指令对禁止路径或外部系统执行写动作。
**典型表现**：自行 `git commit` / `git push` / `git tag`、改 CI、装包、改 Core `/shear` 接口、改 `policies/manifest.json`、接入真实 LLM、对 `.env` / `secrets/` 读取、写入 `schemas/flagship/**` 或 `examples/flagship/**`、改 `go.mod` / `go.sum`。
**守则**：写动作前先把目标路径与第 3 节比对一遍；越界则停。

### AP3 路径漂移
**定义**：对真实不存在的目录或文件进行操作；路径凭空臆造。
**典型表现**：把代码写到 `internal/longma/`（不存在；正确路径是 `internal/flagship/`）；引用 `cmd/longma-server/`（不存在；正确路径是 `cmd/flagship-server/`）；把 freeze 文档写到 `docs/FLAGSHIP-P0-FREEZE-REPORT.md`（连字符错误；正确文件名是 `docs/FLAGSHIP_P0_FREEZE_REPORT.md`）。
**守则**：写入前用 Read / `ls` / `find` 在真实仓库验证父目录存在；不依赖记忆。本轮 `internal/flagship`、`cmd/flagship-server`、`tests/flagship` 在 P0 第一轮已落地，不得自行新建错误路径。

### AP4 幻觉判断
**定义**：凭空假设接口签名、函数行为、模块结构、字段名、版本号、文档原文。
**典型表现**：声称 `internal/flagship/state` 提供 `EvaluateLongma()` 而该函数从未存在；引用 `internal/flagship/output/audit.go` 不存在的字段；参照训练数据中的"五行 / 龙码常数 / 三态推理"印象写代码或文档；自行改矩阵数值、阈值或 verdict 大小写。
**守则**：任何主张要能给出**文件路径 + 行号**；不得引用未读过的代码或文档；P0 主契约（矩阵、`λ μ T1 T2`、风险因子、verdict 大写、字段名、reason_code）以 `v0.4.0-flagship-p0` 实际仓库为准。

### AP5 嘴炮完成
**定义**：声称完成而无可验证证据。
**典型表现**：只说"已完成"，不给 diff、不给文件清单、不给 `go build` / `go vet` / `go test` 输出，不报告 H1–H5 各项状态，不报告 HOLD 项。
**守则**：完成必须给出第 6 节的全部证据；少一项视为未完成。

## 5. HOLD 条件

出现以下任一情况，立即 HOLD 并向用户报告，**不要**自行决策推进：

1. 拟动作目标超出第 1 节单一目标（例如要做 LLM Gateway / 多租户 / 规则自动生成 / 完整龙码 AIOS）。
2. 拟动作要在 P1 第一轮纳入 H1–H5 之外的任何加固 / 扩展 / 重构。
3. 拟修改文件不在第 2 节授权半径。
4. 拟修改文件落在第 3 节禁止路径（含 `schemas/flagship/**` 与 `examples/flagship/**`）。
5. 拟改动会改变 P0 主契约（请求 / 响应 JSON 形状、verdict 大写、矩阵数值、`λ μ T1 T2`、风险因子、字段名）。
6. 需要新增产品级能力、policy pack 或动 Core 引擎 / Core 对外契约。
7. 需要 `git commit` / `push` / `tag` / release，但用户当轮没有显式指令。
8. 需要接入真实 LLM 或调用外部模型 API。
9. 需要新增依赖（`go get` / `go mod tidy` 引入新包，含 jsonschema 等）或升级既有依赖。
10. 需要对接 Core ledger / Core signature。
11. 需要做鉴权 / 限流 / TLS。
12. 无法在真实仓库中验证某个被引用的文件 / 接口 / 字段 / 路径。
13. 完成证据缺任一项。

## 6. 完成证据（缺一不可）

完成时必须输出：

1. **changed files** — 本轮新增 / 修改的全部文件清单及用途。
2. **git diff --stat** — 实际触达哪些文件、增删行数（含 untracked 列举）。
3. **检查结果** — 至少 `go build ./internal/flagship/... ./cmd/flagship-server/...`、`go vet ./internal/flagship/... ./cmd/flagship-server/... ./tests/flagship/...`、`go test ./tests/flagship/...` 与 `go test ./...` 各跑一次的输出（含通过 / 失败列表）。
4. **P0 主契约检查** — 实测 `TestDefaultMatrixFixed` 仍通过；verdict 仍为 `YES` / `HOLD` / `NO` 大写；`AlignmentTask` / `PermitToken` / `RejectInstruction` / `AuditRecord` 字段未减。
5. **边界检查** — 是否触碰禁止路径、是否修改 `policies/manifest.json`、是否动 Core 引擎代码、是否动 Core `/shear` 接口、是否接入真实 LLM、是否引入新依赖、是否动 `go.mod` / `go.sum`、是否在 H1–H5 之外做了加固。
6. **H1–H5 状态** — 每项的实现状态（已实现 / 未实现 / 部分实现）+ 触达文件 + 对应测试。
7. **风险与 HOLD 项** — 已知风险、待澄清项、需要用户决策的事项。

## 7. 与 Core 产品边界的关系

熵剪旗舰版 P1 第一轮小修加固只在闭源开发分支上演化，**不写入** `policies/`、**不出现在** `policies/manifest.json`、**不被** `tests/policy_pack_test.go` 校验、**不出现在** `openapi.yaml` 与 `sdk/`、**不在** README / README_CN / SUPPORT / 既有 `docs/` 里被宣传为 Core 能力。

未来若把本内核与 Core 拼装为完整 Pro / Flagship 商用产品，将是另一轮独立任务，并发生在 Core 范围之外（独立模块或独立分发）。本软约束包不构成对那一轮的承诺，也不锁死其设计。

## 8. 落地实施说明

- `.claude/settings.json` 通过 `permissions.deny` 把第 3 节的禁止路径机器化兜底（其中既有 Core 顶层文件已具名 deny，`internal/**`、`cmd/**`、`schemas/**`、`examples/**`、`tests/**`、`docs/**` 用碎片化子路径 deny，专门为 `internal/flagship/**`、`cmd/flagship-server/**`、`tests/flagship/**`、`docs/FLAGSHIP_P0_FREEZE_REPORT.md`、`docs/FLAGSHIP_P0_DEV.md` 让出可写空间；`schemas/flagship/**` 与 `examples/flagship/**` 自本锚点起被显式 deny），但这只是兜底，不是免责：守则依然以本文件与锚点 JSON 为准。
- 本仓库 `.gitignore` 将整个 `.claude/` 目录排除，因此 `.claude/settings.json` 的 deny 配置只在拥有该文件的工作副本上生效，不随 git 分发到其他副本。如果需要让兜底配置在团队 / CI 中生效，需要单独决定是否调整 `.gitignore`，本软约束包不擅自改 `.gitignore`。
- 本轮 `internal/flagship`、`cmd/flagship-server`、`tests/flagship` 已在 `v0.4.0-flagship-p0` 落地，不得自行新建错误路径或重命名既有目录。

## 9. 进入 P1 第二轮 / 旗舰版下一阶段前的强制门槛

进入熵剪旗舰版 P1 第二轮或更后阶段前，**必须先更新** [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json)：

- 将 `target_constant.single_goal` 从"P1 第一轮小修加固（H1–H5）"替换为下一阶段当轮的真实开发目标。
- 将 `scope_id` / `scope_kind` 替换为下一阶段对应的标识（例如 `flagship-p1-round-2-…` 或 `flagship-p2-…`）。
- 将 `authorization_radius.allowed_files` 替换为下一阶段真正需要触达的路径。
- 同步更新 `target_constant.current_round_deliverables_only`、`forbidden_boundaries`、`completion_evidence_required`、`flagship_p1_scope`（或新增 `flagship_p1_round_2_scope` / `flagship_p2_scope` 等）、`core_product_boundary` 字段。
- 同步更新 `.claude/settings.json` 的 `permissions.deny` 列表，使机器兜底与新一轮授权半径一致。
- 同步更新本文件、`AGENTS.md`、`CLAUDE.md`，使配套文档的"当前轮唯一目标"措辞与新锚点保持一致。

**不得沿用本轮"P1 第一轮小修加固（H1–H5）"的授权半径**去做下一阶段开发。沿用即视为 AP1（目标漂移）与 AP2（行动越权）同时触发，立即 HOLD。

锚点未完成切换前，AI 编程代理仍受本软约束包约束，只能在本轮 `allowed_files` 内写入；超出该范围只能 Read / ls / grep，**不得对 Core / Pro / Flagship 任何代码进行 Write / Edit / MultiEdit**。
