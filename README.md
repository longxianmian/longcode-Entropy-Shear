# Entropy Shear

> **LLMs give you answers. Entropy Shear tells you which ones to trust.**

**Entropy Shear is not an AI. It is a deterministic rule engine.**

Feed it a `policy` and `facts`. It returns one of three indisputable verdicts: **`Yes`**, **`No`**, or **`Hold`**.

No guessing. No hallucination. No LLM calls.
When logic reaches a deadlock, it admits *"I can't judge this"* and emits `Hold` instead of inventing an answer.

Read this in **English** · [中文](README_CN.md)

---

## Why you need this

| Your problem | How Entropy Shear solves it |
|---|---|
| **AI Agent safety** | Gate every Agent action through `/shear`. `Yes` → execute. `No` → block. `Hold` → queue for human review. |
| **Business rules buried in code** | Move rules into Policy JSON. Change them without touching code, redeploying, or restarting. |
| **Audit needs proof, not screenshots** | Every verdict is hashed and appended to a tamper-evident JSONL ledger. The chain is independently verifiable, online or offline. |
| **Uncertain edge cases** | `Hold` is a first-class verdict. It marks ambiguity instead of silently passing or falsely rejecting. |
| **Compliance / KYC / IP-rights gating** | Standard versioned Policy Packs ship with content-addressed hashes — pin a pack, prove which version made which call. |

---

## 5-minute quick start

```bash
# Boot the engine
docker compose up -d --build

# Health
curl -s http://127.0.0.1:8080/health

# Your first verdict
curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/cityone-request.json | jq

# Verify the audit chain
curl -s http://127.0.0.1:8080/ledger/verify | jq
```

That's it. Single binary, JSONL ledger, no database, no auth scaffolding.

Local Go dev:

```bash
go run ./cmd/server   # listens on :8080, ledger at ./ledger/shear-chain.jsonl
```

| Env var | Default |
|---|---|
| `ENTROPY_SHEAR_ADDR`   | `:8080` |
| `ENTROPY_SHEAR_LEDGER` | `ledger/shear-chain.jsonl` |

---

## Core API

| Method | Path | Purpose |
|---|---|---|
| `GET`  | `/health`            | Service health probe |
| `POST` | `/shear`             | Submit a single verdict request |
| `GET`  | `/ledger/{shear_id}` | Fetch a single ledger record |
| `GET`  | `/ledger/verify`     | Re-derive every chained hash from genesis |

### Request

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
        "reason": "Member or VIP users allowed"
      }
    ],
    "default_effect": "Hold",
    "default_reason": "No rule matched, requires human review"
  },
  "facts": {
    "user": { "id": "U-10001", "level": "member", "tags": [] }
  }
}
```

### Response

```json
{
  "verdict": "Yes",
  "applied_rule_id": "rule-001",
  "route": "/campaign/member-landing",
  "reason": "Member or VIP users allowed",
  "trace": [
    { "rule_id": "rule-001", "evaluated": true, "matched": true,
      "detail": "user.level in [member vip], actual member, matched" }
  ],
  "signature": "sha256:...",
  "shear_id": "entropy-shear-20260428-000001"
}
```

`signature` and `shear_id` are your audit trail. Every call appends a record to `ledger/shear-chain.jsonl`.

### Three verdicts

| Verdict | Meaning |
|---|---|
| `Yes`  | Allow / pass / executable |
| `No`   | Refuse / block / not executable |
| `Hold` | Insufficient evidence — needs human review or more facts |

`Hold` is a legal state, not an error. When no rule matches, the engine emits `policy.default_effect` (typically `Hold`).

### Eight operators

`==`  `!=`  `>`  `<`  `>=`  `<=`  `in`  `contains`

Field paths are dot-separated: `user.profile.level`. **Missing paths and type mismatches yield `false`, never an exception.**

### Errors

| HTTP | error code | When |
|---|---:|---|
| 400 | `invalid_json` | Body is not JSON |
| 422 | `policy_schema_violation` / `facts_schema_violation` | Required fields missing or illegal |
| 422 | `unsupported_operator` | Operator outside the eight supported |
| 200 | *(falls through to `Hold`)* | Single-rule evaluation error — recorded in trace `detail` |
| 404 | `not_found` | Unknown `shear_id` |
| 503 | `ledger_unavailable` / `service_unavailable` | Disk / engine fault |

A single bad rule **cannot** bring the service down — the engine falls through to the next rule (or `default_effect`) and records the cause in `trace`.

### Ledger

- File: `ledger/shear-chain.jsonl`, one JSON record per line.
- Append-only. No update, no delete.
- Each record's `previous_shear_hash` equals the previous record's `current_shear_hash`. The first record uses `sha256:genesis`.
- `current_shear_hash` includes `timestamp`; the response `signature` is timestamp-free (idempotent).
- Tamper any line and `GET /ledger/verify` (and `cmd/verify-ledger`) report `broken_at` (1-indexed).

Offline verification:

```bash
go run ./cmd/verify-ledger
go run ./cmd/verify-ledger -ledger /path/to/shear-chain.jsonl
```

Exit 0 on intact, 1 on tamper, 2 on I/O error. **Never writes, never creates directories.**

---

## Real-world use cases

### Production AI Agent Tool Gate

Every action your Agent emits gets gated through `/shear` before execution.

| Action | Verdict | What the gate caller does |
|---|---|---|
| `list_active_users` (readonly) | **Yes**  | execute |
| `delete_production_data`       | **No**   | refuse, surface reason |
| `transfer_funds`               | **Hold** | queue for human confirmation |

Runnable samples (Node + Python) live under [`integrations/agent-tool-gate/`](integrations/agent-tool-gate/).

### AI customer service triage

Customer asks for a refund? Gate it through `/shear`. Missing order ID → `Hold` and ask for it. Medical advice request → `No` and redirect to a professional. High-confidence FAQ → `Yes` and answer.

Companion facts examples and policy pack: [`integrations/ai-customer-service-gate/`](integrations/ai-customer-service-gate/) + [`policies/ai-customer-service/`](policies/ai-customer-service/).

### 标书风控 / Bid & tender risk control

| Submission state | Verdict |
|---|---|
| Required filings missing | **No** — block, list missing items |
| Qualification evidence incomplete | **Hold** — escalate to compliance review |
| All critical checks pass | **Yes** — approve with full audit trail |

Pack: [`policies/bid-risk/`](policies/bid-risk/). Companion request example: [`examples/bid-risk-request.json`](examples/bid-risk-request.json).

### Education — three tiers

Three deterministic verdicts map naturally to age-appropriate AI safety policies. Sample policy directions only — write your own pack and pin its hash.

| Tier | Concrete gate | Yes | No | Hold |
|---|---|---|---|---|
| **Primary school**  | AI homework tutor safety   | In-grade-range and age-appropriate question | Adult / graphic / out-of-policy content | Out-of-syllabus or exam-period — route to teacher / parent |
| **Middle / high school** | Exam-period answer eligibility | Practice mode + permitted topic | Live exam window + assessed subject | Borderline topic during exam — proctor review |
| **University**      | AIGC / plagiarism compliance | AIGC ratio under threshold, similarity OK | AIGC ratio above hard limit, or known plagiarism | Borderline — route to advisor for manual review |

### Identity routing

Used in CityOne-style marketing flows: LINE, WeChat, app users hit a single `/shear`, the verdict picks the landing experience. Policy lives in JSON, changes ship instantly without redeploy.

[`examples/cityone-request.json`](examples/cityone-request.json) is the canonical demo.

### Permission / portal entry

VIP / member → `Yes` + route. Blacklist tag → `No`. Identity facts thin → `Hold` and ask for verification.

Pack: [`policies/permission-gate/`](policies/permission-gate/).

---

## SDKs and policy tooling

```ts
// Node 18+ / TypeScript — no build step, uses global fetch
import { EntropyShearClient } from "@longcode/entropy-shear-client";
const r = await new EntropyShearClient({ baseUrl: "http://127.0.0.1:8080" })
  .shear({ policy, facts });
```

```python
# Python 3.9+ — stdlib only
from entropy_shear_client import EntropyShearClient
r = EntropyShearClient(base_url="http://127.0.0.1:8080").shear(policy=policy, facts=facts)
```

Full client: [`sdk/js/`](sdk/js/) · [`sdk/python/`](sdk/python/).

Validate and hash a policy pack offline:

```bash
go run ./cmd/validate-policy --file policies/agent/agent-action-policy.v1.json
go run ./cmd/hash-policy     --file policies/agent/agent-action-policy.v1.json
```

`policies/manifest.json` records the canonical SHA-256 of every shipped pack; `tests/policy_pack_test.go` re-derives every hash on every test run, so drift fails CI.

---

## Development status

Current version: **`v0.2.0-p2`** (P0 + P1 + P2)

| Stage | Tag | Delivered |
|---|---|---|
| **P0** *(frozen)* | `p0-freeze-20260428` | Verdict kernel — `/shear`, three-state, trace, signature, JSONL ledger, Docker, tests |
| **P1**            | `v0.1.0-p1`          | OpenAPI 3.0.3 spec, JSON Schemas, JS / Python SDKs, offline `cmd/verify-ledger`, Agent tri-state examples |
| **P2** *(current)*| `v0.2.0-p2`          | Versioned policy packs + manifest, `cmd/validate-policy` + `cmd/hash-policy`, Agent Tool Gate samples (Node + Python), AI Customer Service Gate, four developer guides |

Full P2 release notes and acceptance commands: [`docs/P1_RELEASE_CHECKLIST.md`](docs/P1_RELEASE_CHECKLIST.md), [`docs/INTEGRATION_GUIDE.md`](docs/INTEGRATION_GUIDE.md).

### What Entropy Shear will never do

- ❌ Call any LLM — no hallucinations, no prompt injection surface
- ❌ Auto-generate rules — rule authority stays with you
- ❌ Depend on external databases — one binary, one JSONL file
- ❌ Implement user / auth / tenant systems — sits *behind* your auth, not in front
- ❌ Couple with any specific platform (no LIOS submodule, no proprietary host)
- ❌ Hard-code business workflow into the engine
- ❌ Persist raw business data — only hashes go to the ledger

### Repository layout

```
entropy-shear/
  cmd/                         server, verify-ledger, validate-policy, hash-policy
  internal/                    api, engine, schema, ledger, signature, policy, errors
  examples/                    POST /shear request bodies (cityone, agent ×4, ai, bid, gate)
  schemas/                     JSON Schema (draft 2020-12) for Policy / Request / Response / Ledger
  sdk/                         js (TypeScript, fetch) + python (urllib stdlib)
  openapi.yaml                 OpenAPI 3.0.3 spec
  policies/                    versioned policy packs + manifest.json (hash-pinned)
    agent/  ai-customer-service/  bid-risk/  permission-gate/
  integrations/                runnable adapter samples (no LLM, no Agent core)
    agent-tool-gate/{node,python}   ai-customer-service-gate/
  docs/                        WHITEPAPER, INTEGRATION_GUIDE, POLICY_PACK_GUIDE,
                               AGENT_TOOL_GATE_GUIDE, P1_RELEASE_CHECKLIST
  ledger/shear-chain.jsonl     append-only ledger (runtime artifact)
  tests/                       go test ./... (engine + handler + examples + schemas + packs)
  Dockerfile  docker-compose.yml  NOTICE.md  LICENSE
```

### Acceptance commands

```bash
go test ./...
docker compose up -d --build
curl -s http://127.0.0.1:8080/health | jq
curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/agent-action-yes-request.json | jq    # → Yes
curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/agent-action-no-request.json | jq     # → No
curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/agent-action-hold-request.json | jq   # → Hold
curl -s http://127.0.0.1:8080/ledger/verify | jq
go run ./cmd/verify-ledger
go run ./cmd/validate-policy --file policies/agent/agent-action-policy.v1.json
go run ./cmd/hash-policy     --file policies/agent/agent-action-policy.v1.json
```

---

## License

The core engine is released under the **Apache License 2.0** — see [`LICENSE`](LICENSE) and [`NOTICE.md`](NOTICE.md).

For commercial deployment support, on-prem licensing, SLAs, or custom integrations, see [`SUPPORT.md`](SUPPORT.md).

---

## Commercial support

The core engine is free under Apache 2.0. For teams that need more:

- **Entropy Shear Cloud** — pay-as-you-go hosted API, no server maintenance.
- **Entropy Shear On-Prem** — annual license + SLA, designed for finance, healthcare, and compliance environments.
- **Custom policy pack development** — domain-specific packs (bid risk, compliance, education AI safety, agent governance) authored under NDA.

👉 [Commercial licensing & support →](SUPPORT.md)

---

Read this in **English** · [中文](README_CN.md)
