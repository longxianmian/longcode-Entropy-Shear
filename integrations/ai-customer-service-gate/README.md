# AI Customer Service Gate

> **Scope**: this directory shows how to wire Entropy Shear into an AI customer service product as the *answer / refuse / escalate* decision layer. It does not generate text. It does not classify intent. Both of those remain the LLM / NLU layer's job; the gate consumes their output.

```
user message
    ↓
NLU + KB retriever  ← (your stack)
    ↓               produces:
                    intent.type, risk.category, knowledge.hit_confidence
    ↓
POST /shear  with policy = ai-customer-service + facts = above
    ↓
Yes  → AI may answer using KB
No   → refuse, redirect to a professional
Hold → ask for missing info (e.g. order ID) or escalate to human
    ↓
business action
```

## Companion policy

Use [`policies/ai-customer-service/ai-customer-service-policy.v1.json`](../../policies/ai-customer-service/ai-customer-service-policy.v1.json). It is hash-pinned in `policies/manifest.json` and shipped read-only.

## Facts examples

Each file under `facts-examples/` is a `facts` object only — pair it with the policy above to form a `/shear` request body.

| File | Expected verdict | Rule | Suggested business action |
|---|---|---|---|
| `refund-missing-order.json`  | `Hold` | `need-order-id-before-human` | Ask the user for an order ID, then escalate. |
| `high-risk-medical.json`     | `No`   | `forbidden-medical-advice`   | Refuse, redirect to a licensed professional. |
| `faq-high-confidence.json`   | `Yes`  | `faq-answerable`             | Render the AI answer (with the cited KB doc). |

## Try it

```bash
# from the repo root, with the service running:
POLICY=$(jq -c . policies/ai-customer-service/ai-customer-service-policy.v1.json)
FACTS=$(jq  -c . integrations/ai-customer-service-gate/facts-examples/refund-missing-order.json)
curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d "{\"policy\":$POLICY,\"facts\":$FACTS}" | jq '{verdict, applied_rule_id, reason, shear_id}'
```

Repeat for `high-risk-medical.json` and `faq-high-confidence.json`. `tests/policy_pack_test.go` exercises this same pairing on every `go test` run, so the expected verdicts above stay honest.

## What the gate does **not** do

- It does **not** rewrite the AI's answer. Use it as an upstream gate, not a post-filter.
- It does **not** decide *which* human to escalate to — that's a routing concern.
- It does **not** retain conversation context. Pass each turn's facts independently.
- `Hold` is a request for *more information*, not an error.
