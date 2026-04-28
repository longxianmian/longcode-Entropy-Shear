// Runnable demonstration of AgentToolGate.
// Run with `tsx example.ts` or `node --experimental-strip-types example.ts`
// (Node 22+) after `docker compose up -d`.

import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { EntropyShearClient } from "../../../sdk/js/src/index";
import { AgentToolGate, type AgentAction } from "./tool-gate";

const POLICY_PATH = resolve(
  __dirname,
  "../../../policies/agent/agent-action-policy.v1.json"
);

async function main(): Promise<void> {
  const policy = JSON.parse(readFileSync(POLICY_PATH, "utf-8"));
  const client = new EntropyShearClient({
    baseUrl: process.env.ENTROPY_SHEAR_URL ?? "http://127.0.0.1:8080",
  });
  const gate = new AgentToolGate({
    client,
    policy,
    agentId: "agent-007",
    agentTeam: "ops",
  });

  const cases: { label: string; action: AgentAction }[] = [
    {
      label: "readonly query (expect allow)",
      action: { name: "list_active_users", category: "data", mode: "readonly", target: "user_table" },
    },
    {
      label: "delete production (expect deny)",
      action: { name: "delete_production_data", category: "data", mode: "write", target: "orders_2025" },
    },
    {
      label: "transfer funds (expect hold)",
      action: { name: "transfer_funds", category: "payment", mode: "write", target: "wallet:U-9001" },
    },
  ];

  for (const c of cases) {
    const res = await gate.gate(c.action);
    console.log(
      `${c.label}: decision=${res.decision} verdict=${res.shear.verdict} ` +
        `rule=${res.shear.applied_rule_id ?? "<default>"} shear_id=${res.shear.shear_id}`
    );
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
