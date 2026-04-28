# Agent Tool Gate · Python

> **Scope**: this is the *gate layer* between an agent's tool selector and its tool executor. It does **not** implement an agent, it does not call an LLM, it does not generate plans. It does one thing: take an action the agent has chosen, ask Entropy Shear for a verdict, and translate that verdict into `allow / deny / hold`.

## Files

| File | Purpose |
|---|---|
| `tool_gate.py` | Reusable `AgentToolGate` class wrapping `sdk/python/entropy_shear_client.py`. |
| `example.py`   | Three-action demo (readonly / delete-prod / payment). |

## Run the demo

```bash
docker compose up -d --build           # from the repo root
python3 integrations/agent-tool-gate/python/example.py
```

Expected output (one line per action):

```
readonly query        (expect allow)  decision=allow  verdict=Yes   rule=allow-readonly-action       shear_id=entropy-shear-...
delete production     (expect deny)   decision=deny   verdict=No    rule=forbid-delete-production-data shear_id=entropy-shear-...
transfer funds        (expect hold)   decision=hold   verdict=Hold  rule=hold-payment-action         shear_id=entropy-shear-...
```

After the demo, `curl /ledger/verify` should report `total` increased by 3 and `ok: true`.

## Requirements

- Python 3.9+
- No third-party packages — only stdlib `urllib`. The example wires `sys.path` so it works directly from a checkout.

## What the gate does **not** do

- It does not retry on `Hold`. `Hold` is a request for human input, not a transient failure.
- It does not cache verdicts. Every action hits the engine so the ledger is complete.
- It does not authenticate the agent — wire that into your agent runtime.
- It does not enrich `facts` from external systems. Pass the richer signals via `AgentAction.context`.

## Customizing the policy

Replace `POLICY_PATH` in `example.py` (or pass a different policy to `AgentToolGate`) with any pack under `policies/`. Every shipped pack is hash-pinned in `policies/manifest.json` and verified in CI by `tests/policy_pack_test.go`.
