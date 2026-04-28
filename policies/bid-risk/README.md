# Policy Pack · Bid Risk Gate

| | |
|---|---|
| File           | `bid-risk-policy.v1.json` |
| Policy ID      | `policy-bid-risk-v1` |
| Version        | `1.0.0` |
| Scenario       | Bid / tender compliance — decides whether a submission can proceed |
| Maintainer     | `longxianmian@gmail.com` |

## What this pack decides

| Submission shape | Verdict | Rule |
|---|---|---|
| `bid.required_files_missing_count > 0` | **No**   | `missing-required-file` |
| `qualification.evidence_complete == false` | **Hold** | `uncertain-qualification` |
| `risk.critical_issue_count == 0` | **Yes**  | `all-critical-checks-pass` |
| anything else | **Hold** | `default_effect` |

## Required `facts` shape

```json
{
  "bid":           { "id": "BID-...", "required_files_missing_count": 0 },
  "qualification": { "evidence_complete": true },
  "risk":          { "critical_issue_count": 0 }
}
```

## Boundary

- The pack assumes counts are pre-computed by an upstream document-extraction service.
- `Hold` is not "we did not know what to do" — it is "the evidence is not yet complete; route to human review."
- It does not score or rank bids. Ranking is a separate scoring problem.

## How to use

```bash
go run ./cmd/validate-policy --file policies/bid-risk/bid-risk-policy.v1.json
go run ./cmd/hash-policy --file policies/bid-risk/bid-risk-policy.v1.json
```
