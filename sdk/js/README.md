# @longcode/entropy-shear-client

Minimal HTTP client for the [Entropy Shear](https://github.com/longxianmian/longcode-Entropy-Shear) tri-state decision engine.

The SDK only wraps `fetch`. It does **not** evaluate rules locally, cache results, retry, or implement auth.

## Install (local)

```bash
# from the entropy-shear repo root
ln -s "$(pwd)/sdk/js" node_modules/@longcode/entropy-shear-client
# or just copy sdk/js into your project
```

## Requirements

- Node.js 18+ (uses global `fetch` and `AbortController`).
- TypeScript optional. The package exposes `src/index.ts` directly.

## Usage

```ts
import { EntropyShearClient } from "@longcode/entropy-shear-client";

const client = new EntropyShearClient({ baseUrl: "http://127.0.0.1:8080" });

const result = await client.shear({
  policy: {
    id: "policy-permission-gate-v1",
    version: "1.0.0",
    rules: [
      {
        id: "vip-or-member-allow",
        priority: 1,
        condition: { field: "user.level", operator: "in", value: ["member", "vip"] },
        effect: "Yes",
        route: "/portal/home",
        reason: "高等级用户允许进入",
      },
    ],
    default_effect: "Hold",
    default_reason: "身份信息不足，需补充验证",
  },
  facts: { user: { id: "U-9001", level: "vip", tags: [] } },
});

console.log(result.verdict);   // "Yes"
console.log(result.shear_id);  // "entropy-shear-20260428-000001"
console.log(result.signature); // "sha256:..."
```

## API surface

| Method | Returns | HTTP |
|---|---|---|
| `client.health()`              | `HealthResult`  | `GET /health` |
| `client.shear({policy,facts})` | `ShearResult`   | `POST /shear` |
| `client.getLedgerRecord(id)`   | `LedgerRecord`  | `GET /ledger/{id}` |
| `client.verifyLedger()`        | `VerifyResult`  | `GET /ledger/verify` |

Non-2xx responses throw `EntropyShearError` with `.status`, `.code`, `.detail`.

## Options

```ts
new EntropyShearClient({
  baseUrl: "http://127.0.0.1:8080",
  // override fetch (e.g. node-fetch, undici, msw)
  fetch: customFetch,
  // per-request timeout, default 10000
  timeoutMs: 5000,
  // shared abort signal
  signal: abortController.signal,
});
```

## Boundary

This SDK matches the P1 boundary of Entropy Shear itself: no local rule evaluation, no LLM, no DSL, no schema autogen. If your service needs offline evaluation, run the Entropy Shear binary as a sidecar — that is the supported integration shape.
