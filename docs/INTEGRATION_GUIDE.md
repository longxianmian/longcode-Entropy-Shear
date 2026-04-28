# Integration Guide

> Wire your system into Entropy Shear in under an hour. This guide covers: starting the service, calling `/shear`, designing a policy, constructing facts, handling each verdict, and verifying the ledger.

## 1. Start the service

```bash
git clone https://github.com/longxianmian/longcode-Entropy-Shear.git
cd longcode-Entropy-Shear
docker compose up -d --build
curl -s http://127.0.0.1:8080/health | jq
```

You should see:

```json
{"ok": true, "service": "entropy-shear", "version": "v1.0.0"}
```

The ledger is persisted at `./ledger/shear-chain.jsonl` (host bind mount).

---

## 2. Call `POST /shear`

The contract is exactly `{policy, facts}` in, `ShearResult` out. There is no auth, no tenant, no session.

```bash
curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/agent-action-yes-request.json | jq
```

Full schema: [`schemas/shear-request.schema.json`](../schemas/shear-request.schema.json) and [`schemas/shear-response.schema.json`](../schemas/shear-response.schema.json).

For any production use, prefer one of the SDKs: [`sdk/js`](../sdk/js) or [`sdk/python`](../sdk/python).

---

## 3. Design a policy

A policy is a **versioned, priority-ordered ruleset with a default verdict**:

```json
{
  "id": "policy-my-domain-v1",
  "version": "1.0.0",
  "rules": [
    { "id": "deny-…", "priority": 1, "condition": {…}, "effect": "No",   "reason": "…" },
    { "id": "hold-…", "priority": 2, "condition": {…}, "effect": "Hold", "reason": "…" },
    { "id": "allow-…","priority": 3, "condition": {…}, "effect": "Yes",  "reason": "…" }
  ],
  "default_effect": "Hold",
  "default_reason": "未命中明确条件，需人工复核"
}
```

Design rules in this order:

1. **Hard denies first** (priority = 1). These should be your "never under any circumstances" cases.
2. **Holds next**. Anything that means "we cannot decide alone" goes here, before any allow.
3. **Allows last**. They only fire when no deny / hold matched.
4. **`default_effect` is your safety net**. Make it `Hold` unless you have a good reason for `Yes`.

Operator cheat sheet (full table in [`README.md` §3](../README.md)):

| Operator | Use when |
|---|---|
| `==` / `!=` | Compare scalars (string / number / bool). |
| `>` / `<` / `>=` / `<=` | Numeric thresholds. |
| `in` | Left-hand value is one of a set. |
| `contains` | Left-hand array or string contains a value. |

Field paths are dot-separated: `user.profile.level`. **Missing paths and type mismatches yield `false`, not errors.**

---

## 4. Construct `facts`

Facts is a free-form JSON object — the engine reads from it via dot-paths. Two principles:

1. **Shape the facts to mirror your domain.** `user.level`, `intent.type`, `bid.required_files_missing_count` are all chosen to be self-explanatory.
2. **Keep facts minimal.** What is not in the policy's conditions only adds noise (and ledger size). Pre-compute scores / counts upstream.

The shipped policy packs document exactly which fact paths they read. See `policies/<pack>/README.md`.

---

## 5. Handle each verdict

| Verdict | Caller's job |
|---|---|
| `Yes`  | Proceed with the action. The optional `route` may suggest a downstream redirect. |
| `No`   | Refuse the action. Surface `reason` to the operator / end user. |
| `Hold` | Do **not** retry blindly. Either collect the missing facts and retry, or escalate to a human queue. |

A useful pattern in code:

```ts
const r = await client.shear({ policy, facts });
switch (r.verdict) {
  case "Yes":  return execute(action);
  case "No":   throw new Refused(r.reason);
  case "Hold": return queueForHuman(r);
}
```

For a full Tool Gate sample see [`integrations/agent-tool-gate/`](../integrations/agent-tool-gate/).

---

## 6. Audit & ledger

Every `/shear` call appends a hash-chained record. Two endpoints + one CLI:

```bash
curl -s http://127.0.0.1:8080/ledger/verify | jq        # online
go run ./cmd/verify-ledger                              # offline, exit 1 on tamper
curl -s http://127.0.0.1:8080/ledger/<shear_id> | jq    # single record
```

Ledger record shape: [`schemas/ledger-record.schema.json`](../schemas/ledger-record.schema.json).

Tampering with any line in `ledger/shear-chain.jsonl` makes both `/ledger/verify` and `cmd/verify-ledger` report `broken_at`. Treat this as an integrity alarm.

---

## 7. Versioning and pack hashes

Every shipped policy is hash-pinned. Before deploying a policy update:

```bash
go run ./cmd/validate-policy --file policies/<scenario>/<name>.v<n>.json
go run ./cmd/hash-policy     --file policies/<scenario>/<name>.v<n>.json
```

Update the corresponding entry in `policies/manifest.json`. `tests/policy_pack_test.go` re-derives every hash on each test run, so any drift fails CI.

---

## 8. Error handling

| HTTP | error code | When |
|---|---|---|
| 400 | `invalid_json` | Body is not valid JSON. |
| 422 | `policy_schema_violation` | Missing fields / illegal values in policy. |
| 422 | `facts_schema_violation`  | Facts is not an object. |
| 422 | `unsupported_operator`    | Operator outside the eight supported. |
| 503 | `ledger_unavailable`      | Ledger append failed (disk / fs). |
| 503 | `service_unavailable`     | Generic engine error. |
| 404 | `not_found`               | Unknown `shear_id`. |

Single-rule evaluation errors **never** return 5xx. They fall through to the next rule (or the `default_effect`) and are recorded in the trace's `detail`. The engine cannot be brought down by a single bad rule.

---

## 9. Common integration shapes

| Integration | Recommended starting point |
|---|---|
| Agent Tool Gate | [`integrations/agent-tool-gate/`](../integrations/agent-tool-gate/) (Node + Python) |
| AI customer service | [`integrations/ai-customer-service-gate/`](../integrations/ai-customer-service-gate/) |
| Bid / tender risk | `policies/bid-risk/` + custom upstream document extractor |
| Portal / app entry | `policies/permission-gate/` |

Each has its own README pinning the expected verdicts and required facts shape.
