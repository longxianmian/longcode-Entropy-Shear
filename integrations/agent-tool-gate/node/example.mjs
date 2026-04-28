// Runnable demo: node example.mjs (Node 18+).
// Requires `docker compose up -d` first so /shear is reachable.

import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";
import { AgentToolGate } from "./tool-gate.mjs";

const __dirname = dirname(fileURLToPath(import.meta.url));
const POLICY_PATH = resolve(__dirname, "../../../policies/agent/agent-action-policy.v1.json");

const policy = JSON.parse(readFileSync(POLICY_PATH, "utf-8"));
const baseUrl = process.env.ENTROPY_SHEAR_URL ?? "http://127.0.0.1:8080";

const gate = new AgentToolGate({
  baseUrl,
  policy,
  agentId: "agent-007",
  agentTeam: "ops",
});

const cases = [
  {
    label: "readonly query        (expect allow)",
    action: { name: "list_active_users", category: "data", mode: "readonly", target: "user_table" },
  },
  {
    label: "delete production     (expect deny)",
    action: { name: "delete_production_data", category: "data", mode: "write", target: "orders_2025" },
  },
  {
    label: "transfer funds        (expect hold)",
    action: { name: "transfer_funds", category: "payment", mode: "write", target: "wallet:U-9001" },
  },
];

for (const c of cases) {
  const res = await gate.gate(c.action);
  console.log(
    `${c.label}  decision=${res.decision}  verdict=${res.shear.verdict}  ` +
      `rule=${res.shear.applied_rule_id ?? "<default>"}  shear_id=${res.shear.shear_id}`
  );
}
