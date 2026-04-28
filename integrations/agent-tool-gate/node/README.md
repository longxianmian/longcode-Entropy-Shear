# Agent Tool Gate · Node

> **Scope**: this is the *gate layer* between an agent's tool selector and its tool executor. It does **not** implement an agent. It does not call an LLM. It does not generate plans. It does one thing: take an action the agent has chosen, ask Entropy Shear for a verdict, and translate that verdict into `allow / deny / hold`.

```
agent.plan() → action object
              ↓
         AgentToolGate.gate(action)
              ↓
        POST /shear          (Entropy Shear)
              ↓
        Yes  → "allow"   → agent executes the action
        No   → "deny"    → agent refuses, surfaces reason
        Hold → "hold"    → agent queues for human confirmation
```

## Files

| File | Purpose |
|---|---|
| `tool-gate.ts`  | TypeScript reference implementation (uses the SDK in `sdk/js`). |
| `example.ts`    | Three-action demo (readonly / delete-prod / payment). |
| `tool-gate.mjs` | Plain JS variant of the gate, no TypeScript toolchain needed. |
| `example.mjs`   | Plain JS demo runnable with `node example.mjs`. |
| `package.json`  | `npm run demo` → JS · `npm run demo:ts` → TypeScript via `tsx`. |

## Run the demo

```bash
docker compose up -d --build           # from the repo root
node integrations/agent-tool-gate/node/example.mjs
```

Expected output (one line per action):

```
readonly query        (expect allow)  decision=allow  verdict=Yes   rule=allow-readonly-action       shear_id=entropy-shear-...
delete production     (expect deny)   decision=deny   verdict=No    rule=forbid-delete-production-data shear_id=entropy-shear-...
transfer funds        (expect hold)   decision=hold   verdict=Hold  rule=hold-payment-action         shear_id=entropy-shear-...
```

After the demo, `curl /ledger/verify` should report `total` increased by 3 and `ok: true`.

## What the gate does **not** do

- It does not retry on `Hold`. `Hold` is a request for human input, not a transient failure.
- It does not cache verdicts. Every action hits the engine so the ledger is complete.
- It does not authenticate the agent — wire that into your agent runtime, not the gate.
- It does not enrich `facts` from external systems. If your policy needs richer signals, fetch them in your tool selector and pass them via `action.context`.

## Customizing the policy

Replace `POLICY_PATH` in `example.mjs` (or pass a different policy to `AgentToolGate`) with any pack under `policies/`. Every shipped pack is hash-pinned in `policies/manifest.json` and verified in CI by `tests/policy_pack_test.go`.
