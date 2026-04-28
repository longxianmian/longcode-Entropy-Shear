# Policy Pack · Permission Gate

| | |
|---|---|
| File           | `permission-gate-policy.v1.json` |
| Policy ID      | `policy-permission-gate-v1` |
| Version        | `1.0.0` |
| Scenario       | Enterprise / portal entry gate — decides whether a user may enter a resource |
| Maintainer     | `longxianmian@gmail.com` |

## What this pack decides

| User shape | Verdict | Rule |
|---|---|---|
| `user.level ∈ {member, vip}` | **Yes**  | `vip-or-member-allow` (route `/portal/home`) |
| `user.tags contains "blocked"` | **No**   | `blocked-tag-deny` |
| anything else | **Hold** | `default_effect` |

## Required `facts` shape

```json
{
  "user":    { "id": "U-9001", "level": "vip", "tags": ["paid_2025"] },
  "context": { "channel": "web", "action": "open_portal" }
}
```

## Boundary

- The pack does not authenticate users — feed it the result of authentication.
- `Hold` means identity facts are insufficient; ask for additional verification rather than denying outright.
- The pack does not make per-action authorization decisions — for that, compose with a richer policy.

## How to use

```bash
go run ./cmd/validate-policy --file policies/permission-gate/permission-gate-policy.v1.json
go run ./cmd/hash-policy --file policies/permission-gate/permission-gate-policy.v1.json
```
