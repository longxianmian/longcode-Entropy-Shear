# Commercial Support · 商业支持

> The core Entropy Shear engine is released under [Apache 2.0](LICENSE) and is free for any use, including commercial. This page is for teams that want a deployment partner, an SLA, custom policy packs, or licensing for environments where Apache 2.0 alone isn't enough.
>
> 熵剪核心引擎采用 [Apache 2.0](LICENSE) 开源，包括商业用途在内全部免费。本页面面向需要部署伙伴、SLA 保障、定制策略包，或 Apache 2.0 之外的额外许可的团队。

> **Edition scope:** The open-source repository contains **Entropy Shear Core** only. **Pro** and **Flagship** capabilities are commercial editions and are not part of the Apache 2.0 open-source core. See [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md) for the full edition boundary.
>
> **版本范围**：本开源仓库**仅包含 Entropy Shear Core（标准版 / 开源版）**。**Pro** 与 **Flagship** 属于商业版本，不在 Apache 2.0 开源核心之内。完整边界见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md)。

---

## Contact · 联系方式

**📧 Email**: [longxianmian@gmail.com](mailto:longxianmian@gmail.com)

Languages: 中文 / English. We reply within 2 business days for inquiries that include a project description, team size, and target deployment environment.

---

## What we offer · 我们提供什么

### 1. Entropy Shear Cloud — 托管 API

Pay-as-you-go hosted `/shear` endpoint. Suitable when you want the verdict engine without running it yourself.

- 99.9% SLA (paid tiers)
- Hash-pinned policy packs maintained for you
- Ledger archival to your S3 / OSS bucket
- Independent third-party time-stamping for the ledger chain

按量计费的托管 `/shear` 接口，适合需要熵剪能力但不想自己运维的团队。99.9% SLA、策略包维护、账本归档到你自己的对象存储、链式时间戳第三方存证。

### 2. Entropy Shear On-Prem — 私有化部署

Annual license with SLA, designed for finance, healthcare, public-sector, and compliance environments.

- Single-binary install or Kubernetes Helm chart
- On-site or remote pack authoring assistance
- 24×7 support escalation path (paid tiers)
- Quarterly review of your active policy packs

年付授权 + SLA，面向金融、医疗、政企、合规场景。单二进制安装或 K8s Helm chart、私有化策略包开发协助、24×7 支持响应、季度策略审阅。

### 3. Custom policy pack development — 定制策略包

Domain-specific verdict packs authored under NDA. Existing focus areas:

- **Bid & tender risk control** — required-files compliance, qualification evidence completeness, critical-risk gating. 标书风控：必需文件合规、资质证据完整性、关键风险拦截。
- **AI customer service safety** — refund / complaint triage, medical / legal / financial advice gating, knowledge-base confidence routing. AI 客服：退款投诉分流、高风险建议拦截、知识库置信度路由。
- **Agent tool governance** — production-action allow/deny/hold lists, payment confirmation gates, destructive-action prevention. Agent 工具治理：生产动作放行 / 拦截 / 暂缓、支付动作人工确认、破坏性动作前置拦截。
- **Education AI safety (K-12 + university)** — primary-school content safety, exam-period eligibility, university AIGC compliance. 教育 AI 安全（K-12 + 高校）：小学内容安全、中学考试期资格、大学 AIGC 合规。
- **Compliance & KYC gating** — for fintech, public sector, and cross-border regulated workflows. 合规 / KYC 闸门：金融科技、政企、跨境合规流程。

Custom-pack engagements include policy authoring, hash registration, integration sample, and one round of post-deploy revision.

定制 pack 工作包含：策略起草、hash 登记、接入样板、上线后一轮修订。

### 4. Integration consulting — 接入咨询

For teams already running Entropy Shear who want a senior pair-programming session on:

- AI Agent action-gate architecture
- Multi-pack composition without coupling
- Ledger archival, audit, and time-stamping pipelines
- Migration from ad-hoc rule code to pack-based governance

按小时 / 项目计费的高级咨询：Agent 动作闸门架构、多 pack 组合、账本归档审计与时间戳管道、从硬编码规则迁移到 pack 治理。

---

## What's NOT in any commercial tier · 商业服务也不会做的事

The same hard limits that govern the open-source engine apply to every commercial engagement. We will never:

- Inject LLM calls into the verdict engine.
- Generate rules automatically from data — rule authority always stays with you.
- Build a multi-tenant SaaS that holds raw business data on our servers.
- Couple the engine to any specific business platform.

This isn't a marketing position; it's a deliberate engineering boundary. The verdict path stays deterministic, auditable, and replayable, no matter how the commercial wrapper evolves.

商业服务同样保持核心边界：永远不引入 LLM 调用、不自动生成规则、不在我方服务器持久化原始业务数据、不耦合特定平台。这不是市场口号，是工程约束 — 让裁决路径在任何商业包装下都保持确定性、可审计、可复算。

---

## Commercial editions roadmap · 商业版本路线

The commercial tiers extend Core; they never replace its deterministic kernel. The full edition matrix lives in [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md).

| Edition | Adds on top of | Headline capability | Status |
|---|---|---|---|
| **Entropy Shear Pro** | Core | **Longma Constant** — goal-locked AI governance | Commercial; not implemented in this repository |
| **Entropy Shear Flagship** | Pro | five-element conflict governance + AI rule compiler + shadow mode + advanced audit ledger | Commercial; not implemented in this repository |

商业版本在 Core 之上叠加，永远不替换 Core 的确定性内核。完整边界见 [`docs/PRODUCT_MATRIX.md`](docs/PRODUCT_MATRIX.md)。

| 版本 | 在哪一档之上叠加 | 核心增量 | 状态 |
|---|---|---|---|
| **Entropy Shear Pro 版** | Core | **龙码常数（Longma Constant）** — 目标锁定型 AI 治理 | 商业版；本仓库未实现 |
| **Entropy Shear Flagship 旗舰版** | Pro 版 | 五元素冲突治理 + AI 规则编译器 + 影子模式 + 高级审计账本 | 商业版；本仓库未实现 |

For inquiries about Pro / Flagship licensing, deployment, or pre-release access: [longxianmian@gmail.com](mailto:longxianmian@gmail.com).

---

## Open-source contribution · 开源协作

For technical discussion, bug reports, and feature requests:

- [GitHub Issues](https://github.com/longxianmian/longcode-Entropy-Shear/issues) — defect reports, technical questions
- [GitHub Discussions](https://github.com/longxianmian/longcode-Entropy-Shear/discussions) — design proposals, integration patterns

We accept PRs against the core engine under the Apache 2.0 contribution model. By submitting a PR you agree to license your contribution under Apache 2.0.

技术讨论、缺陷报告、功能请求请走 GitHub Issues / Discussions。核心引擎接受 PR — 提交 PR 即视为你同意以 Apache 2.0 许可你的贡献。

---

## Looking for partners · 寻求合作伙伴

We are actively looking for partners in:

- ISVs (independent software vendors) in finance, healthcare, and legal compliance
- LLM / AI Agent platforms looking for a deterministic governance layer
- System integrators serving banks, insurers, and public-sector clouds
- Educational technology vendors building AI-safety-aware tutors and exam systems

我们正在寻找以下领域的合作伙伴：金融 / 医疗 / 法律合规领域的 ISV；面向大模型 / AI Agent 平台的确定性治理层；银行、保险、政务云的私有化部署系统集成商；建设 AI 安全意识辅导 / 考试系统的教育科技厂商。

If any of this matches you, please reach out: [longxianmian@gmail.com](mailto:longxianmian@gmail.com).

---

## Rights holder · 权利主体

Copyright (c) 2026 龙码（广州）数字科技有限公司. All rights reserved on commercial trademarks and brand assets. Engine source code is licensed under Apache 2.0 — see [`LICENSE`](LICENSE) and [`NOTICE.md`](NOTICE.md).

Author / 作者: 龙先冕 / 龙行天下
