# 熵剪旗舰版 P0 开发软约束（LONGMA_SOFT_GUARD）

> 本文件不是熵剪 Core 的产品功能。它只在 AI 编程代理（Claude Code / Codex 等）开发熵剪旗舰版 P0 龙码三态逻辑推理内核时生效，用于约束开发过程的目标边界与执行边界，治理目标漂移、行动越权、路径漂移、幻觉判断与嘴炮完成。

## 0. 边界声明（必读）

熵剪旗舰版 P0 龙码三态逻辑推理内核：

- **不属于** Entropy Shear Core 的对外能力。
- **不进入** `policies/`、**不修改** `policies/manifest.json`、**不修改** `tests/policy_pack_test.go`。
- **不被** `openapi.yaml` 与 `sdk/` 暴露。
- **不被** `README.md` / `README_CN.md` / `SUPPORT.md` / `docs/PRODUCT_MATRIX.md` 等产品文档宣传为 Core 能力。
- **不接入** 任何真实 LLM；本轮内核全部走确定性逻辑。
- 与 [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json) 互为正副本：本文件给人读，锚点 JSON 给机器读，二者冲突时**以锚点 JSON 为准**。

Longma Constant 仍仅属于 Pro / Flagship 路线，详见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md) 与 [`README.md`](README.md) 的 Edition Boundary 段落。

## 1. 唯一目标（target constant）

> **在不破坏现有 Entropy Shear Core 的前提下，开发熵剪旗舰版 P0 的龙码三态逻辑推理内核最小可运行版本（Minimum Verifiable Reasoner）。**

本轮允许触达的路径**只**有：

1. `internal/flagship/**`
2. `cmd/flagship-server/**`
3. `schemas/flagship/**`
4. `examples/flagship/**`
5. `tests/flagship/**`
6. `docs/FLAGSHIP_P0_DEV.md`
7. `LONGMA_TASK_ANCHOR.json`

非目标（本轮严禁触达）：

- 不开发完整龙码 AIOS
- 不开发 LLM Gateway
- 不接入真实 LLM
- 不开发知识资产化操作系统
- 不开发能力工具系统
- 不开发后台管理系统
- 不做多租户 SaaS
- 不做规则自动生成
- 不改现有 Core /shear 接口
- 不改现有 policy + facts 裁决逻辑
- 不改现有 ledger / signature / policy pack 行为
- 不把龙码常数作为 Core 版正式能力发布

## 2. 授权半径

允许写入的文件 = 第 1 节七条路径，且**仅这七条**。其它一切只允许 Read / ls / grep（用于验证），**不允许写**。

落入 `allowed_files` 内的合规子能力实现范围（仅作目标边界声明，不强制全部一次实现）：

- 多源输入标准结构
- 五元映射器：Goal / Fact / Evidence / Constraint / Action
- 原子校验规则执行器
- 内部状态计算
- 5×5 状态评估矩阵
- 五元干涉模型 Lite
- 加权冲突消解
- 三态状态机：YES / HOLD / NO
- HOLD 动态对齐任务生成
- permit_token / reject_instruction
- 审计记录结构
- 最小测试用例

## 3. 禁止边界

下列路径**禁止**修改：

- `internal/{api,engine,errors,ledger,policy,schema,signature}/**`（Core 引擎）
- `cmd/{hash-policy,server,validate-policy,verify-ledger}/**`（Core 命令行）
- `schemas/*.json`（顶层 schema；`schemas/flagship/**` 除外）
- `sdk/**`
- `policies/**`（含 `policies/manifest.json`）
- `integrations/**`
- `docs/{AGENT_TOOL_GATE_GUIDE,INTEGRATION_GUIDE,P1_RELEASE_CHECKLIST,POLICY_PACK_GUIDE,PRODUCT_MATRIX,WHITEPAPER}.md`（既有产品文档；`docs/FLAGSHIP_P0_DEV.md` 除外）
- `tests/*.go`（顶层 Core 测试；`tests/flagship/**` 除外）
- `ledger/**`
- `examples/*.json`（顶层既有示例；`examples/flagship/**` 除外）
- `README.md`、`README_CN.md`、`SUPPORT.md`
- `openapi.yaml`
- `go.mod`、`go.sum`
- `Dockerfile`、`docker-compose.yml`
- `AGENTS.md`、`CLAUDE.md`、`LONGMA_SOFT_GUARD.md`、`.gitignore`

下列动作**禁止**自行执行（必须用户当轮显式指令）：

- `git commit` / `git push` / `git tag` / 任何 release 动作
- 改动 CI、外部依赖、镜像仓库、域名、密钥
- 安装新包、升级 `go mod`、`go mod tidy` 引入新包、改动构建产物
- 接入真实 LLM、调用任何外部模型 API
- 删除或重写禁止路径下的任何已有文件

## 4. 治理的五种反模式

### AP1 目标漂移
**定义**：在执行过程中悄悄扩展、迁移或替换本轮目标。
**典型表现**：把"P0 龙码三态逻辑推理内核 MVR"扩张成"实现完整龙码 AIOS / LLM Gateway / 多租户 SaaS / 规则自动生成 / 知识资产化操作系统 / 能力工具系统 / 后台管理"；或顺手把内核能力以 policy pack 形式塞进 Core。
**守则**：任何越出第 1 节单一目标的动作必须先停下，把"我想做 X，是否授权"显式问出来。

### AP2 行动越权
**定义**：未经用户当轮显式指令对禁止路径或外部系统执行写动作。
**典型表现**：自行 `git commit` / `git push` / `git tag`、改 CI、装包、改 Core `/shear` 接口、改 `policies/manifest.json`、接入真实 LLM、对 `.env` / `secrets/` 读取。
**守则**：写动作前先把目标路径与第 3 节比对一遍；越界则停。

### AP3 路径漂移
**定义**：对真实不存在的目录或文件进行操作；路径凭空臆造。
**典型表现**：把代码写到 `internal/longma/`（不存在；正确路径是 `internal/flagship/`）；引用 `cmd/longma-server/`（不存在；正确路径是 `cmd/flagship-server/`）；把 P0 dev 文档写到 `docs/FLAGSHIP-P0-DEV.md`（连字符错误；正确文件名是 `docs/FLAGSHIP_P0_DEV.md`）。
**守则**：写入前用 Read / `ls` / `find` 在真实仓库验证父目录存在；不依赖记忆。本轮 `internal/flagship`、`cmd/flagship-server`、`schemas/flagship`、`examples/flagship`、`tests/flagship` 等目录初始为空，第一次写入需要显式 `mkdir -p` 并对用户透明。

### AP4 幻觉判断
**定义**：凭空假设接口签名、函数行为、模块结构、字段名、版本号、文档原文。
**典型表现**：声称 `internal/engine` 提供 `EvaluateLongma()` 而该函数从未存在；引用 `schemas/policy.schema.json` 不存在的字段；参照训练数据中的"五行 / 龙码常数 / 三态推理"印象写代码或文档。
**守则**：任何主张要能给出**文件路径 + 行号**；不得引用未读过的代码或文档；不得参照训练数据中的"熵剪 / 龙码常数 / 三态推理"印象，一切以仓库当前文件为准。

### AP5 嘴炮完成
**定义**：声称完成而无可验证证据。
**典型表现**：只说"已完成"，不给 diff、不给文件清单、不给 `go build` / `go test` 输出，不报告 HOLD 项。
**守则**：完成必须给出第 6 节的全部证据；少一项视为未完成。

## 5. HOLD 条件

出现以下任一情况，立即 HOLD 并向用户报告，**不要**自行决策推进：

1. 拟动作目标超出第 1 节单一目标（例如要做 LLM Gateway / 多租户 / 规则自动生成 / 完整龙码 AIOS）。
2. 拟修改文件不在第 2 节授权半径。
3. 拟修改文件落在第 3 节禁止路径。
4. 需要新增产品级能力、policy pack 或动 Core 引擎 / Core 对外契约。
5. 需要 `git commit` / `push` / `tag` / release，但用户当轮没有显式指令。
6. 需要接入真实 LLM 或调用外部模型 API。
7. 需要新增依赖（`go get` / `go mod tidy` 引入新包）或升级既有依赖。
8. 无法在真实仓库中验证某个被引用的文件 / 接口 / 字段 / 路径。
9. 完成证据缺任一项。

## 6. 完成证据（缺一不可）

完成时必须输出：

1. **changed files** — 本轮新增 / 修改的全部文件清单及用途（含 `mkdir -p` 创建的新目录）。
2. **git diff --stat** — 实际触达哪些文件、增删行数（含 untracked 列举）。
3. **检查结果** — 至少 `go build ./internal/flagship/... ./cmd/flagship-server/...` 与 `go test ./tests/flagship/...` 各跑一次的输出（含通过 / 失败列表）。本轮不要求改 Core 测试，但 flagship 子模块自身必须可编译并通过自带测试。
4. **边界检查** — 是否触碰禁止路径、是否修改 `policies/manifest.json`、是否动 Core 引擎代码、是否动 Core `/shear` 接口、是否接入真实 LLM、是否引入新依赖。
5. **风险与 HOLD 项** — 已知风险、待澄清项、需要用户决策的事项。

## 7. 与 Core 产品边界的关系

熵剪旗舰版 P0 龙码三态逻辑推理内核只在闭源开发分支上演化，**不写入** `policies/`、**不出现在** `policies/manifest.json`、**不被** `tests/policy_pack_test.go` 校验、**不出现在** `openapi.yaml` 与 `sdk/`、**不在** README / README_CN / SUPPORT / 既有 `docs/` 里被宣传为 Core 能力。

未来若把本内核与 Core 拼装为完整 Pro / Flagship 商用产品，将是另一轮独立任务，并发生在 Core 范围之外（独立模块或独立分发）。本软约束包不构成对那一轮的承诺，也不锁死其设计。

## 8. 落地实施说明

- `.claude/settings.json` 通过 `permissions.deny` 把第 3 节的禁止路径机器化兜底（其中既有 Core 顶层文件已具名 deny，`internal/**`、`cmd/**`、`schemas/**`、`examples/**`、`tests/**`、`docs/**` 用碎片化子路径 deny，专门为 `*/flagship/**` 与 `docs/FLAGSHIP_P0_DEV.md` 让出可写空间），但这只是兜底，不是免责：守则依然以本文件与锚点 JSON 为准。
- 本仓库 `.gitignore` 将整个 `.claude/` 目录排除，因此 `.claude/settings.json` 的 deny 配置只在拥有该文件的工作副本上生效，不随 git 分发到其他副本。如果需要让兜底配置在团队 / CI 中生效，需要单独决定是否调整 `.gitignore`，本软约束包不擅自改 `.gitignore`。
- `internal/flagship`、`cmd/flagship-server`、`schemas/flagship`、`examples/flagship`、`tests/flagship` 等目录在本轮锚点同步发布时尚未创建；P0 第一次代码写入时由代理自身用 `mkdir -p`（或等价工具）创建，并在交付证据中明示所建目录。

## 9. 进入 P1 / 旗舰版下一阶段前的强制门槛

进入熵剪旗舰版 P1 或更后阶段前，**必须先更新** [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json)：

- 将 `target_constant.single_goal` 从"P0 龙码三态逻辑推理内核 MVR"替换为下一阶段当轮的真实开发目标。
- 将 `scope_id` / `scope_kind` 替换为下一阶段对应的标识（例如 `flagship-p1-…`）。
- 将 `authorization_radius.allowed_files` 替换为下一阶段真正需要触达的路径。
- 同步更新 `target_constant.current_round_deliverables_only`、`forbidden_boundaries`、`completion_evidence_required`、`flagship_p0_scope`（或新增 `flagship_p1_scope` 等）、`core_product_boundary` 字段。
- 同步更新 `.claude/settings.json` 的 `permissions.deny` 列表，使机器兜底与新一轮授权半径一致。
- 同步更新本文件、`AGENTS.md`、`CLAUDE.md`，使配套文档的"当前轮唯一目标"措辞与新锚点保持一致。

**不得沿用本轮"P0 龙码三态逻辑推理内核 MVR"的授权半径**去做下一阶段开发。沿用即视为 AP1（目标漂移）与 AP2（行动越权）同时触发，立即 HOLD。

锚点未完成切换前，AI 编程代理仍受本软约束包约束，只能在本轮 `allowed_files` 内写入；超出该范围只能 Read / ls / grep，**不得对 Core / Pro / Flagship 任何代码进行 Write / Edit / MultiEdit**。
