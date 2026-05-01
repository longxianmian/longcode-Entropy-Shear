# Entropy Shear Product Matrix

> The single source of truth for what each Entropy Shear edition does — and, just as importantly, what each one **does not** do.
>
> 各版本能力边界的唯一权威来源 — 同样重要的是：明确**不做什么**。

This document scopes three editions: **Core** (the contents of this open-source repository), **Pro** (a future commercial edition), and **Flagship** (a future enterprise edition). Pro and Flagship are **not implemented in this repository** and are **not** part of the Apache 2.0 open-source core. They are documented here to set expectations, not to imply availability.

本文界定三个版本：**Core**（即本开源仓库）、**Pro**（未来的商业版本）、**Flagship**（未来的企业旗舰版本）。Pro 与 Flagship **不在本仓库实现**，也**不在** Apache 2.0 开源核心之内。此处仅为划清边界、管理预期，不代表已可购买或交付。

---

## 1. Edition matrix

| Edition | Positioning | Included | Not Included |
|---|---|---|---|
| **Core / Open Source** | Deterministic tri-state decision engine | Yes/Hold/No verdicts, `/shear`, policy + facts contract, JSONL ledger, policy packs, SDKs, integration samples | Longma Constant, AI compiler, five-element governance, shadow mode |
| **Pro / Commercial** | Goal-locked AI governance engine | Core + Longma Constant | Flagship conflict governance, AI compiler |
| **Flagship / Enterprise** | Enterprise conflict governance engine | Pro + five-element governance + AI compiler + shadow mode + advanced audit ledger | Full LIOS operating system |

---

## 2. Entropy Shear Core — what this repository delivers

**License:** Apache 2.0 (see [`../LICENSE`](../LICENSE) and [`../NOTICE.md`](../NOTICE.md))
**Latest tag:** `v0.3.0-p3`
**Stage history:** P0 → P1 → P2 → P3

Entropy Shear Core is the deterministic tri-state decision engine. Its surface is intentionally small and **stable** — once tagged, the `/shear` request/response shape and ledger format do not change in incompatible ways.

| Capability | Where it lives |
|---|---|
| `Yes` / `Hold` / `No` verdict | `internal/engine/` |
| Eight operators (`==` `!=` `>` `<` `>=` `<=` `in` `contains`) | `internal/engine/evaluator.go` |
| Policy + Facts input contract | `internal/schema/`, `schemas/*.json`, `openapi.yaml` |
| `/shear` HTTP API | `internal/api/handler.go` |
| Trace per rule | embedded in every response |
| Tamper-evident JSONL ledger | `internal/ledger/`, `cmd/verify-ledger/` |
| Hash-pinned policy packs | `policies/` + `cmd/{validate,hash}-policy` |
| JS / Python SDKs (HTTP only) | `sdk/js/`, `sdk/python/` |
| Integration samples (Agent Tool Gate, AI Customer Service Gate) | `integrations/` |

Core's hard limits — these will not change as long as Core remains Core:

- ❌ No LLM calls
- ❌ No automatic rule generation
- ❌ No external database dependency
- ❌ No user / auth / tenant system
- ❌ No multi-tenant SaaS scaffolding
- ❌ No platform coupling (no LIOS submodule, no proprietary host)
- ❌ No business-workflow hard-coding

If a feature crosses any of these lines, it does not belong in Core — it belongs in Pro or Flagship.

---

## 3. Pro — Goal-locked AI governance (commercial, not implemented)

Pro adds a single named capability on top of Core: **Longma Constant** (龙码常数).

| | |
|---|---|
| **Positioning** | Goal-locked AI governance engine. Used when a deployment must keep AI decisions aligned to a pre-set, hash-pinned target — not just rule-compliant in the moment. |
| **Built on top of** | Core (unchanged) |
| **Includes** | Core + Longma Constant |
| **Does not include** | Flagship-only conflict governance; AI rule compiler; shadow mode; advanced audit ledger |
| **License** | Commercial (terms via [`../SUPPORT.md`](../SUPPORT.md)) |
| **Status** | Not implemented in this repository. Reserved namespace. |

What Longma Constant *is not*: an LLM, a rule generator, a runtime modifier of Core's verdict logic. Pro extends Core; it does not replace any part of it. Calls to Pro that bypass Core's deterministic verdict path are by definition not Pro.

---

## 4. Flagship — Enterprise conflict governance (commercial, not implemented)

Flagship adds the heavy governance machinery to Pro: cross-rule conflict detection, an AI rule compiler, a shadow mode for safe rule rollouts, and an advanced audit ledger.

| | |
|---|---|
| **Positioning** | Enterprise conflict governance engine. For organizations whose policy surface area outgrows manual stewardship — typically banks, insurers, regulated public-sector clouds. |
| **Built on top of** | Pro (which is built on Core) |
| **Includes** | Pro + five-element conflict governance + AI rule compiler + shadow mode + advanced audit ledger |
| **Does not include** | A full LIOS operating system. Flagship is a governance edition, not a platform OS. |
| **License** | Commercial (terms via [`../SUPPORT.md`](../SUPPORT.md)) |
| **Status** | Not implemented in this repository. Reserved namespace. |

The hard line between Flagship and any "platform" product: Flagship governs *decisions*. It does not govern *workflows*, *data pipelines*, or *user systems*. Those remain outside Entropy Shear in any edition.

---

## 5. Stability promise across editions

| Promise | Core | Pro | Flagship |
|---|:-:|:-:|:-:|
| `/shear` request/response shape never breaks within a major version | ✓ | ✓ | ✓ |
| Ledger format is append-only and chain-verifiable | ✓ | ✓ | ✓ |
| Policy pack manifest hash semantics never silently change | ✓ | ✓ | ✓ |
| Verdict path stays deterministic (no probabilistic detours) | ✓ | ✓ | ✓ |
| Source of the verdict path remains auditable | ✓ | ✓ | ✓ |

These promises are the contract. Any edition that breaks one of them is not Entropy Shear.

---

## 6. What this matrix is not

- It is not a release schedule. Pro and Flagship may ship, may ship later, or may not ship at all in this exact form.
- It is not a feature checklist customers can rely on for procurement before commercial terms are signed.
- It is not an invitation to graft Pro / Flagship features into Core. Contributions that implement Longma Constant, an AI rule compiler, conflict governance, or shadow mode in this repository will be declined — those belong in commercial editions.

For commercial inquiries see [`../SUPPORT.md`](../SUPPORT.md). For the technical contract of Core see [`./WHITEPAPER.md`](./WHITEPAPER.md), [`./INTEGRATION_GUIDE.md`](./INTEGRATION_GUIDE.md), [`./POLICY_PACK_GUIDE.md`](./POLICY_PACK_GUIDE.md), and [`./AGENT_TOOL_GATE_GUIDE.md`](./AGENT_TOOL_GATE_GUIDE.md).
