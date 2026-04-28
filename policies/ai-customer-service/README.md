# Policy Pack · AI Customer Service Gate

| | |
|---|---|
| File           | `ai-customer-service-policy.v1.json` |
| Policy ID      | `policy-ai-customer-service-v1` |
| Version        | `1.0.0` |
| Scenario       | AI customer service answer gate — decides "answer / refuse / escalate / collect more facts" |
| Maintainer     | `longxianmian@gmail.com` |

## What this pack decides

| Conversation shape | Verdict | Rule | Suggested business action |
|---|---|---|---|
| `intent.type ∈ {refund, complaint, after_sale}` | **Hold** | `need-order-id-before-human` | Ask for order ID, then escalate |
| `risk.category == "medical_advice"`             | **No**   | `forbidden-medical-advice` | Refuse, redirect to a licensed professional |
| `knowledge.hit_confidence >= 0.85`              | **Yes**  | `faq-answerable` | Render the AI answer |
| anything else                                    | **Hold** | `default_effect` | Escalate to human or collect more context |

## Required `facts` shape

```json
{
  "session":   { "id": "sess-...", "channel": "web" },
  "intent":    { "type": "refund", "raw_text": "..." },
  "risk":      { "category": "general" },
  "knowledge": { "hit_confidence": 0.42 }
}
```

Three fields are load-bearing for matching: `intent.type`, `risk.category`, `knowledge.hit_confidence`.

## Boundary

- The pack does **not** generate answer text — that is the LLM's job.
- The pack does **not** classify intent or risk — feed it numbers/labels your upstream NLU has already produced.
- `Hold` is the **safe default** for ambiguity. Engineering teams that route every Hold to a human inbox will get the spec's intended behavior.

## Companion facts examples

`integrations/ai-customer-service-gate/facts-examples/` ships three vetted facts payloads — one per verdict. They pair directly with this pack.

## How to use

```bash
go run ./cmd/validate-policy --file policies/ai-customer-service/ai-customer-service-policy.v1.json
go run ./cmd/hash-policy --file policies/ai-customer-service/ai-customer-service-policy.v1.json
```
