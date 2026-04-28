# Policy Pack Guide

> A *policy pack* is a versioned, hash-pinned policy file shipped under `policies/`, paired with a README that documents the exact facts it consumes and the verdicts it produces. This guide covers the structure, the manifest contract, hashing, and how to ship a new pack.

## 1. Directory layout

```
policies/
  <scenario>/
    <name>-policy.v<n>.json     # the policy file
    README.md                   # facts shape, verdict table, boundaries
  manifest.json                 # the index — hash-pinned listing of every pack
```

Current packs (P2):

- `policies/agent/agent-action-policy.v1.json`
- `policies/ai-customer-service/ai-customer-service-policy.v1.json`
- `policies/bid-risk/bid-risk-policy.v1.json`
- `policies/permission-gate/permission-gate-policy.v1.json`

## 2. Naming convention

```
<scenario-folder>/<name>-policy.v<major>.json
```

- `<scenario-folder>` matches the integration target (`agent`, `ai-customer-service`, …).
- `<name>` is human-readable, hyphenated.
- `v<major>` is monotonically increasing. **Bump `v<major>` whenever the canonical hash would change.** Inside the file, also bump `policy.version` to the new SemVer (e.g. `1.0.0` → `2.0.0` for a breaking change, `1.0.0` → `1.1.0` for an additive rule).

A pack at `v1.json` and another at `v2.json` may legally coexist — runtime callers pick which one to load.

## 3. The policy file itself

Each `*.v<n>.json` file is exactly the JSON form that `POST /shear` expects under the `policy` key. Validated against [`schemas/policy.schema.json`](../schemas/policy.schema.json):

```json
{
  "id": "policy-<scenario>-v<n>",
  "version": "<semver>",
  "rules":   [ … priority-ordered rules … ],
  "default_effect": "Hold | Yes | No",
  "default_reason": "..."
}
```

Operators allowed: `== != > < >= <= in contains`. See the engine's [README §3](../README.md).

## 4. The README contract

Every pack ships a README with exactly these sections:

| Section | Why |
|---|---|
| Header table (file, policy id, version, scenario, maintainer) | one-glance metadata |
| **What this pack decides** — verdict table per fact shape | how the rules behave |
| **Required `facts` shape** — JSON template | what the caller must supply |
| **Boundary** — what the pack does **not** decide | sets correct expectations |
| **How to use** — `validate-policy` + `hash-policy` invocations | one-command verification |

See `policies/agent/README.md` for a reference template.

## 5. The manifest

`policies/manifest.json` is a flat list of every pack the repo ships:

```json
{
  "manifest_version": "1.0.0",
  "generated_at": "2026-04-28",
  "policies": [
    {
      "path": "policies/agent/agent-action-policy.v1.json",
      "policy_id": "policy-agent-action-v1",
      "policy_version": "1.0.0",
      "hash": "sha256:…",
      "scenario": "agent-tool-gate",
      "maintainer": "longxianmian@gmail.com",
      "created_at": "2026-04-28"
    }
  ]
}
```

Field rules:

- `path` — relative to the repo root.
- `policy_id` / `policy_version` — exactly as written inside the policy file.
- `hash` — the canonical SHA-256 of the policy. **Recomputed on every test run via `cmd/hash-policy`; drift fails CI.**
- `scenario` — short slug matching the integration target.
- `maintainer` — single-owner contact, not a team alias. The owner is responsible for triaging breaking changes.
- `created_at` — ISO date the pack was first introduced. Bumping version does not reset this.

## 6. Hash discipline

The canonical hash is computed over the **decoded + re-canonicalized** JSON, not the raw file bytes. So whitespace and key order in the source file do not affect the hash — but adding, removing, or changing any field does.

```bash
go run ./cmd/hash-policy --file policies/agent/agent-action-policy.v1.json
# {
#   "policy_id": "policy-agent-action-v1",
#   "version": "1.0.0",
#   "hash": "sha256:d6c00d46…"
# }
```

Use this hash when:

- registering a pack with a third-party timestamp service;
- pinning a deployed pack inside a release manifest;
- making a software-rights / copyright filing;
- referring to a specific pack version in incident reports.

## 7. Lifecycle of a new pack

```
1. Decide the scenario folder and pack name.
2. Write the policy file. Run cmd/validate-policy until it returns ok=true.
3. Write the README following §4 of this guide.
4. Run cmd/hash-policy. Copy the hash into manifest.json.
5. Add a manifest entry (path / id / version / hash / scenario / maintainer / created_at).
6. Add a test row in tests/policy_pack_test.go that asserts the canonical
   verdict for the pack's primary use case.
7. Run go test ./.... All packs re-derive their hashes; any drift fails.
8. Open a PR. The hash, manifest, README, and test row must all land together.
```

Never edit a published `v<n>.json` in place. Bump to `v<n+1>.json` and update the manifest accordingly.

## 8. Boundary — what packs cannot do

A policy pack is a ruleset, not a program. It cannot:

- call out to external services (no HTTP / DNS / DB inside facts evaluation);
- transform facts before matching (the engine just does dot-path lookups);
- compose multiple policies (use one pack per `/shear` call; chain at the application layer);
- rely on randomness or current time (rules must be deterministic).

If you need any of the above, do it in your tool selector / facts builder before calling `/shear`. The pack stays the deterministic, auditable kernel.
