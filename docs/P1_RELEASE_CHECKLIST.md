# Entropy Shear · P1 Release Checklist

> Confirm before tagging `v0.1.0-p1`.

## 1. Compatibility (P0 must remain green)

- [ ] `POST /shear` request shape unchanged (`{policy, facts}` only).
- [ ] `POST /shear` response shape unchanged (`verdict / applied_rule_id / route / reason / trace / signature / shear_id`).
- [ ] Same input + same `previous_shear_hash` → same `signature` (timestamp-free signature still holds).
- [ ] `GET /ledger/{shear_id}` and `GET /ledger/verify` unchanged.
- [ ] All P0 unit tests still pass (`go test ./tests`).

## 2. Artifacts in repo

- [ ] `openapi.yaml` — parses in Swagger Editor; covers `/health`, `/shear`, `/ledger/{shear_id}`, `/ledger/verify` only (no unimplemented endpoints).
- [ ] `schemas/policy.schema.json`
- [ ] `schemas/shear-request.schema.json`
- [ ] `schemas/shear-response.schema.json`
- [ ] `schemas/ledger-record.schema.json`
- [ ] Each schema declares `$schema` and `title`.
- [ ] `examples/agent-action-request.json` — canonical → `Yes` (`allow-readonly-action`).
- [ ] `examples/ai-customer-service-request.json` — canonical → `Hold` (`need-order-id-before-human`).
- [ ] `examples/bid-risk-request.json` — canonical → `No` (`missing-required-file`).
- [ ] `examples/permission-gate-request.json` — canonical → `Yes` (`vip-or-member-allow`).
- [ ] Existing `examples/cityone-*.json` unchanged.
- [ ] `sdk/js/{package.json, src/index.ts, README.md}` — TypeScript, fetch-only, no auth, no retry.
- [ ] `sdk/python/{entropy_shear_client.py, README.md}` — stdlib only, no third-party deps.
- [ ] `cmd/verify-ledger/main.go` — exits 0 / 1 / 2; never writes.
- [ ] `tests/examples_test.go` — every example asserted; every policy proven across Yes/No/Hold.
- [ ] `tests/schema_examples_test.go` — schema files parse, examples are valid JSON with `policy + facts`.
- [ ] `tests/ledger_test.go` — covers `VerifyFile` (offline) on both happy and missing-file paths.
- [ ] `README.md` §11 documents the P1 surface.
- [ ] `NOTICE.md` unchanged.

## 3. Acceptance commands (must all pass)

```bash
go test ./...

docker compose up -d --build
curl -s http://127.0.0.1:8080/health

curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/agent-action-request.json | jq

curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/ai-customer-service-request.json | jq

curl -s -X POST http://127.0.0.1:8080/shear \
  -H 'Content-Type: application/json' \
  -d @examples/bid-risk-request.json | jq

curl -s http://127.0.0.1:8080/ledger/verify | jq
go run ./cmd/verify-ledger
```

Expected:

- `go test ./...` → all green.
- `/health` → `{"ok":true,"service":"entropy-shear","version":"v1.0.0"}`.
- `agent-action-request.json` → `verdict: "Yes"`, `applied_rule_id: "allow-readonly-action"`.
- `ai-customer-service-request.json` → `verdict: "Hold"`, `applied_rule_id: "need-order-id-before-human"`.
- `bid-risk-request.json` → `verdict: "No"`, `applied_rule_id: "missing-required-file"`.
- `/ledger/verify` → `ok: true`, `total ≥ 3`, `broken_at: null`.
- `verify-ledger` exits 0 with the same `ok / total / latest_hash`.

## 4. Boundary audit (must remain absent)

- [ ] No LLM client / inference code anywhere.
- [ ] No rule-autogeneration code.
- [ ] No admin UI, no user system, no auth middleware.
- [ ] No database driver / ORM / migration tool.
- [ ] No multi-tenant scaffolding.
- [ ] No `lios` / `cityone` / external-system imports inside `internal/`.
- [ ] No business workflow hard-coded into engine or handler.
- [ ] No raw business data persisted in full — ledger only stores hashes.

## 5. Versioning

- [ ] `cmd/server/main.go::version` still `v1.0.0` (P0 baseline) **or** bumped intentionally.
- [ ] `sdk/js/package.json::version` matches release intent.
- [ ] No automatic git tag / push from this branch — release tag is a separate manual step.

## 6. Post-release (manual, not in this branch)

- Tag `v0.1.0-p1` after all checks above are green.
- Attach release notes citing P0/P1 diff.
- Optionally attach a hash manifest of `examples/` and `schemas/` for archival.
