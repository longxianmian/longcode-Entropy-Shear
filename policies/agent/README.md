# Policy Pack · Agent Action Gate

| | |
|---|---|
| File           | `agent-action-policy.v1.json` |
| Policy ID      | `policy-agent-action-v1` |
| Version        | `1.0.0` |
| Scenario       | Agent Tool Gate — pre-execution verdict for any tool call an agent wants to run |
| Maintainer     | `longxianmian@gmail.com` |

## What this pack decides

Every action an agent emits is classified into one of three buckets:

| Action shape | Verdict | Rule |
|---|---|---|
| `action.name == "delete_production_data"` | **No**   | `forbid-delete-production-data` |
| `action.category == "payment"`             | **Hold** | `hold-payment-action` |
| `action.mode == "readonly"`                | **Yes**  | `allow-readonly-action` |
| anything else                              | **Hold** | `default_effect` |

`No` means the gate refuses; `Hold` means human confirmation is required; `Yes` means the agent may proceed.

## Required `facts` shape

```json
{
  "agent":  { "id": "agent-007", "team": "ops" },
  "action": {
    "name":     "list_active_users",
    "category": "data",
    "mode":     "readonly",
    "target":   "user_table"
  }
}
```

Only `action.name`, `action.category`, and `action.mode` are read by the policy. Everything else is recorded in the trace's input hash but not used in matching.

## Boundary

- This pack does **not** decide whether the action *should exist* — that is the agent's planner's responsibility.
- `Hold` is **not** an error. It means the gate cannot decide alone and the calling system must collect human confirmation before retry.
- The pack assumes the gate caller (Tool Gate sample under `integrations/agent-tool-gate/`) has already performed its own input sanitization.

## How to use

```bash
go run ./cmd/validate-policy --file policies/agent/agent-action-policy.v1.json
go run ./cmd/hash-policy --file policies/agent/agent-action-policy.v1.json
```

The hash listed in `policies/manifest.json` must match the output of `hash-policy` exactly. `tests/policy_pack_test.go` re-derives every hash on every test run.
