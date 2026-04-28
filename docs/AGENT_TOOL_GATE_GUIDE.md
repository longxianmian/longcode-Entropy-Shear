# Agent Tool Gate Guide

> How to wire Entropy Shear in front of an Agent's tool executor. Covers the architectural pattern, the recommended `action` facts shape, the three-state response handling, and a default high-risk-action taxonomy.

This guide pairs with the runnable samples in [`integrations/agent-tool-gate/{node,python}`](../integrations/agent-tool-gate/).

## 1. Architectural pattern

```
┌──────────────────────────────┐
│  Agent runtime               │
│  (LLM, planner, scratchpad)  │
└─────────────┬────────────────┘
              │  produces an action object
              ▼
┌──────────────────────────────┐
│  AgentToolGate.gate(action)  │   ← this layer
│                              │
│  build facts ──► POST /shear │
└─────────────┬────────────────┘
              │  Yes / No / Hold + trace + signature + shear_id
              ▼
┌──────────────────────────────┐
│  Tool executor               │
│  - Yes  → run tool           │
│  - No   → refuse, surface    │
│           reason             │
│  - Hold → queue for human    │
└──────────────────────────────┘
```

Three boundaries to keep clean:

1. **Agent runtime ↔ Tool Gate**: the agent passes a structured `action`, never raw text. If your agent emits free text "I want to delete production data", parse it into an `action` first.
2. **Tool Gate ↔ Entropy Shear**: nothing other than `{policy, facts}` flows over the wire. No session, no auth, no tenant.
3. **Tool Gate ↔ Tool executor**: the gate returns a verdict; the executor decides what to actually do. Don't let the gate run the tool.

## 2. Recommended `action` facts shape

The shipped agent pack reads three fields. A robust action facts envelope looks like:

```json
{
  "agent": {
    "id":   "agent-007",
    "team": "ops"
  },
  "action": {
    "name":     "list_active_users",
    "category": "data | payment | email | code-exec | external-api | …",
    "mode":     "readonly | write",
    "target":   "user_table",
    "context":  { "any": "extra fields you want recorded in input_hash" }
  }
}
```

Notes:

- **`name`** is the *logical* action name, not a function pointer.
- **`category`** is a small enum your team controls. Keep it stable — every change forces a manifest hash bump for any pack that matches on it.
- **`mode`** is the cheapest way to flip readonly into a default-allow path. Pair with `category` for finer-grained decisions.
- **`target`** is purely metadata. Useful for forensics, not for matching.
- **`context`** is your escape hatch. Anything in here goes into the trace's `input_hash` and is auditable, but is invisible to matching unless your pack reads `action.context.<x>`.

## 3. Handling Yes / No / Hold

| Verdict | Decision | Caller responsibility |
|---|---|---|
| `Yes`  | `allow` | Execute the tool. Surface `route` if your downstream needs it. |
| `No`   | `deny`  | Refuse the action. Bubble `reason` up to the operator UI / logs. Do **not** auto-rephrase the action and retry; that defeats the gate. |
| `Hold` | `hold`  | Queue for human confirmation **or** collect missing facts and retry. **Never auto-loop.** |

Anti-patterns:

- ❌ Looping on `Hold` until you accidentally land on `Yes`.
- ❌ Treating `No` as a temporary failure and retrying with backoff.
- ❌ Caching verdicts client-side. The ledger is the audit trail; bypassing it means losing the ability to prove what happened.

## 4. Default high-risk action taxonomy

A starter taxonomy your team can extend. Each row is a candidate rule (deny-list / hold-list) for the agent pack:

| Category | Example actions | Default verdict |
|---|---|---|
| `data:destruct` | `delete_production_data`, `drop_table`, `truncate_<x>` | **No** |
| `data:exfiltrate` | `dump_secrets`, `export_pii_<x>` | **No** |
| `payment` | `transfer_funds`, `refund_order`, `charge_card` | **Hold** |
| `external-api:high-impact` | `send_mass_email`, `post_to_socials`, `call_emergency_service` | **Hold** |
| `code-exec:remote` | `run_remote_shell`, `deploy_to_prod` | **Hold** |
| `code-exec:sandbox` | `python_eval_in_sandbox` | **Yes** (if mode=readonly) / **Hold** otherwise |
| `data:read` | `list_<x>`, `get_<x>`, `search_<x>` | **Yes** when `action.mode == "readonly"` |
| anything else | (default) | **Hold** (let humans triage new shapes) |

When you graduate a category from `Hold` to `Yes`, ship a new pack (`v2`) and update `policies/manifest.json`. Don't mutate `v1` in place.

## 5. Failure modes you should test

A robust Tool Gate integration handles each of:

| Scenario | Expected behavior |
|---|---|
| Entropy Shear is unreachable | Treat as `Hold` (do **not** silently allow). |
| `/shear` returns 5xx | Treat as `Hold`. |
| `/shear` returns `Hold` with no `applied_rule_id` | Default fall-through hit; queue for human. |
| Action object has unexpected fields | Pass-through; the engine ignores unmatched paths. Log for visibility. |
| Ledger reports `broken_at` | Page on-call. Do not trust new verdicts until the chain is restored. |

## 6. Why a separate gate at all?

Because the agent's job — **understanding intent and proposing actions** — is fundamentally non-deterministic. Putting that decision back inside the agent's LLM call (e.g. "you may also self-critique") yields the same model's confidence in its own action: not auditable, not reproducible.

A separate, deterministic gate gives you:

- a stable contract surface (`policy + facts`) you can version;
- a hash-chain audit trail that survives even if the agent restarts;
- the ability to run the same gate from Node, Python, Go, or curl;
- a Hold path that doesn't pretend the model knows what to do.

That is what Entropy Shear is for.
