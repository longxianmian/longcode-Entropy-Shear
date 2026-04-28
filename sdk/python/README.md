# entropy-shear-client (Python)

Minimal HTTP client for the [Entropy Shear](https://github.com/longxianmian/longcode-Entropy-Shear) tri-state decision engine. Stdlib only — no `pip install` required.

The SDK only wraps `urllib.request`. It does **not** evaluate rules locally, cache results, retry, or implement auth.

## Requirements

- Python 3.9+
- No third-party packages

## Usage

```python
from entropy_shear_client import EntropyShearClient

client = EntropyShearClient(base_url="http://127.0.0.1:8080")

policy = {
    "id": "policy-permission-gate-v1",
    "version": "1.0.0",
    "rules": [
        {
            "id": "vip-or-member-allow",
            "priority": 1,
            "condition": {"field": "user.level", "operator": "in", "value": ["member", "vip"]},
            "effect": "Yes",
            "route": "/portal/home",
            "reason": "高等级用户允许进入",
        },
    ],
    "default_effect": "Hold",
    "default_reason": "身份信息不足，需补充验证",
}
facts = {"user": {"id": "U-9001", "level": "vip", "tags": []}}

result = client.shear(policy=policy, facts=facts)
print(result["verdict"])    # "Yes"
print(result["shear_id"])   # "entropy-shear-20260428-000001"
print(result["signature"])  # "sha256:..."

# Audit
print(client.verify_ledger())  # {"ok": True, "total": 1, ...}
record = client.get_ledger_record(result["shear_id"])
```

## API surface

| Method | Returns | HTTP |
|---|---|---|
| `client.health()`                    | `dict` | `GET /health` |
| `client.shear(policy=..., facts=...)`| `dict` | `POST /shear` |
| `client.get_ledger_record(shear_id)` | `dict` | `GET /ledger/{shear_id}` |
| `client.verify_ledger()`             | `dict` | `GET /ledger/verify` |

Non-2xx responses raise `EntropyShearError(status, code, detail)`.

## Options

```python
EntropyShearClient(base_url="http://127.0.0.1:8080", timeout=5.0)
```

## Boundary

This SDK matches the P1 boundary of Entropy Shear: no local rule evaluation, no LLM, no DSL, no schema autogen. For offline evaluation, run the Entropy Shear binary as a sidecar.
