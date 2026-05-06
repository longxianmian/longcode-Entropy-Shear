# 龙码常数开发期软约束包 P0 (LONGMA_SOFT_GUARD)

> 本文件不是熵剪 Core 的产品功能。它只在 AI 编程代理（Claude Code / Codex 等）开发本仓库时生效，用于约束开发过程的目标边界与执行边界，治理目标漂移、行动越权、路径漂移、幻觉判断与嘴炮完成。

## 0. 边界声明（必读）

本软约束包：

- **不属于** Entropy Shear Core 的对外能力。
- **不进入** `policies/`、**不修改** `policies/manifest.json`、**不修改** `tests/policy_pack_test.go`。
- **不被** `openapi.yaml` 与 `sdk/` 暴露。
- **不被** `README.md` / `README_CN.md` / `SUPPORT.md` / `docs/` 宣传为 Core 能力。
- **不代表** Core 实现 Longma Constant。Longma Constant 仍然属于 Pro / Flagship 路线，详见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md) 与 [`README.md`](README.md) 的 Edition Boundary 段落。
- 与 [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json) 互为正副本：本文件给人读，锚点 JSON 给机器读，二者冲突时**以锚点 JSON 为准**。

## 1. 唯一目标（target constant）

> **在现有 Entropy Shear Core 之上准备开发熵剪旗舰版（Flagship），但本轮只建立开发期软约束包 P0，不开发任何核心代码。**

本轮交付物**只**包含以下五个文件：

1. `LONGMA_TASK_ANCHOR.json`
2. `LONGMA_SOFT_GUARD.md`（本文件）
3. `AGENTS.md`
4. `CLAUDE.md`
5. `.claude/settings.json`

非目标（本轮严禁触达）：

- 实现 Longma Constant 进 Core
- 实现五行冲突治理 / AI 规则编译器 / 影子模式 / 高级审计账本
- 修改三态判定逻辑或任何 Core 引擎代码
- 新增 Core 产品 policy pack 或更新 `policies/manifest.json`
- 改动 Core 对外契约（API / Schema / SDK / OpenAPI）
- 改动产品文档（README / README_CN / SUPPORT / docs/）
- 改动构建与发布产物（Dockerfile / docker-compose / go.mod / go.sum）

## 2. 授权半径

允许写入的文件 = 第 1 节五个文件，且**仅这五个**。
其它一切只允许 Read / ls / grep（用于验证），**不允许写**。

## 3. 禁止边界

下列路径**禁止**修改：

- `internal/**`
- `cmd/**`
- `schemas/**`
- `sdk/**`
- `policies/**`
- `integrations/**`
- `docs/**`
- `tests/**`
- `ledger/**`
- `README.md`、`README_CN.md`、`SUPPORT.md`
- `openapi.yaml`
- `go.mod`、`go.sum`
- `Dockerfile`、`docker-compose.yml`

下列动作**禁止**自行执行（必须用户当轮显式指令）：

- `git commit` / `git push` / `git tag` / 任何 release 动作
- 改动 CI、外部依赖、镜像仓库、域名、密钥
- 安装新包、升级 go mod、改动构建产物
- 删除或重写禁止路径下的任何已有文件

## 4. 治理的五种反模式

### AP1 目标漂移
**定义**：在执行过程中偷偷扩展、迁移或替换本轮目标。
**典型表现**：把"建立软约束包"扩张成"顺手新增一个 policy pack"或"顺手 commit 一下"。
**守则**：任何越出第 1 节单一目标的动作必须先停下，把"我想做 X，是否授权"显式问出来。

### AP2 行动越权
**定义**：未经用户当轮显式指令对禁止路径或外部系统执行写动作。
**典型表现**：自行 `git commit`、`git push`、`git tag`、改 CI、装包、改外部接口。
**守则**：写动作前先把目标路径与第 3 节比对一遍；越界则停。

### AP3 路径漂移
**定义**：对真实不存在的目录或文件进行操作；路径凭空臆造。
**典型表现**：写到 `policies/longma-constant/`（本仓库没有这个目录）；引用 `internal/longma/`（不存在）。
**守则**：写入前用 Read / `ls` / `find` 在真实仓库验证父目录存在；不依赖记忆。

### AP4 幻觉判断
**定义**：凭空假设接口签名、函数行为、模块结构、字段名、版本号、文档原文。
**典型表现**：声称"`internal/engine` 提供 `EvaluateLongma()`"而该函数从未存在；引用 README 中没有的句子。
**守则**：任何主张要能给出**文件路径 + 行号**；不得引用未读过的代码或文档；不得参照训练数据中的"熵剪 / 龙码常数"印象，一切以仓库当前文件为准。

### AP5 嘴炮完成
**定义**：声称完成而无可验证证据。
**典型表现**：只说"已完成"，不给 diff、不给文件清单、不报告 HOLD 项。
**守则**：完成必须给出第 6 节的全部证据；少一项视为未完成。

## 5. HOLD 条件

出现以下任一情况，立即 HOLD 并向用户报告，**不要**自行决策推进：

1. 拟动作目标超出第 1 节单一目标。
2. 拟修改文件不在第 2 节授权半径。
3. 拟修改文件落在第 3 节禁止路径。
4. 需要新增产品级能力、policy pack 或动 Core 引擎。
5. 需要 `git commit` / `push` / `tag` / release，但用户当轮没有显式指令。
6. 无法在真实仓库中验证某个被引用的文件 / 接口 / 字段 / 路径。
7. 完成证据缺任一项。

## 6. 完成证据（缺一不可）

完成时必须输出：

1. **changed files** — 本轮新增 / 修改的全部文件清单及用途。
2. **git diff --stat** — 实际触达哪些文件、增删行数（含 untracked 列举）。
3. **检查结果** — 执行过的校验命令及输出（本轮以仓库结构 / 路径校验为主，不要求跑 Go 测试）。
4. **边界检查** — 是否触碰禁止路径、是否修改 `policies/manifest.json`、是否动 Core 引擎代码。
5. **风险与 HOLD 项** — 已知风险、待澄清项、需要用户决策的事项。

## 7. 与 Core 产品边界的关系

本软约束包只在开发期生效，**不写入** `policies/`、**不出现在** `policies/manifest.json`、**不被** `tests/policy_pack_test.go` 校验、**不出现在** `openapi.yaml` 与 `sdk/`、**不在** README / docs 里被宣传为 Core 能力。

如果未来需要把 Longma Constant 升级为 Pro / Flagship 中真正的产品能力，那将是另一轮独立任务，并发生在 Core 范围之外（独立模块或独立分发）。本软约束包不构成对那一轮的承诺，也不锁死其设计。

## 8. 落地实施说明

- `.claude/settings.json` 通过 `permissions.deny` 把第 3 节的禁止路径机器化兜底，但这只是兜底，不是免责：守则依然以本文件与锚点 JSON 为准。
- 本仓库 `.gitignore` 将整个 `.claude/` 目录排除，因此 `.claude/settings.json` 的 deny 配置只在拥有该文件的工作副本上生效，不随 git 分发到其他副本。如果需要让兜底配置在团队 / CI 中生效，需要单独决定是否调整 `.gitignore`，本软约束包不擅自改 `.gitignore`。

## 9. 进入下一轮（熵剪旗舰版 P0）前的强制门槛

进入熵剪旗舰版 P0 开发前，**必须先更新** [`LONGMA_TASK_ANCHOR.json`](LONGMA_TASK_ANCHOR.json)，把授权半径切换为具体开发任务，例如：

- 将 `target_constant.single_goal` 从"建立软约束包 P0"替换为旗舰版 P0 当轮的真实开发目标。
- 将 `scope_id` / `scope_kind` 从 `longma-soft-guard-p0` / `development-time-soft-constraint` 替换为旗舰版 P0 对应的标识。
- 将 `authorization_radius.allowed_files` 从本轮 5 个文件替换为旗舰版 P0 真正需要触达的代码 / 模块路径。
- 同步更新 `current_round_deliverables_only`、`forbidden_boundaries`、`completion_evidence_required` 与 `core_product_boundary` 字段。
- 同步更新 `.claude/settings.json` 的 `permissions.deny` 列表，使机器兜底与新一轮授权半径一致。

**不得沿用本轮"只建立软约束包"的授权半径**去做旗舰版 P0 开发。沿用即视为 AP1（目标漂移）与 AP2（行动越权）同时触发，立即 HOLD。

锚点未完成切换前，AI 编程代理仍受本软约束包 P0 约束，只能 Read / ls / grep，**不得对 Core / Pro / Flagship 任何代码进行 Write / Edit / MultiEdit**。
