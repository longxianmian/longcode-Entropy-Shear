// Plain ESM JavaScript variant of tool-gate.ts. Behavior is identical.
// Use this when you want to run without a TypeScript toolchain.

export function mapVerdict(verdict) {
  switch (verdict) {
    case "Yes":
      return "allow";
    case "No":
      return "deny";
    case "Hold":
    default:
      return "hold";
  }
}

export class AgentToolGate {
  constructor({ baseUrl, policy, agentId = "agent-unknown", agentTeam = "default", fetchImpl = globalThis.fetch }) {
    if (!baseUrl) throw new Error("baseUrl is required");
    if (!policy) throw new Error("policy is required");
    if (typeof fetchImpl !== "function") {
      throw new Error("fetch is not available; Node 18+ required or pass fetchImpl");
    }
    this.baseUrl = baseUrl.replace(/\/+$/, "");
    this.policy = policy;
    this.agentId = agentId;
    this.agentTeam = agentTeam;
    this.fetch = fetchImpl;
  }

  async gate(action) {
    const facts = {
      agent: { id: this.agentId, team: this.agentTeam },
      action: {
        name: action.name,
        category: action.category,
        mode: action.mode,
        ...(action.target !== undefined ? { target: action.target } : {}),
        ...(action.context !== undefined ? { context: action.context } : {}),
      },
    };
    const res = await this.fetch(this.baseUrl + "/shear", {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ policy: this.policy, facts }),
    });
    if (!res.ok) {
      const body = await res.text();
      throw new Error(`entropy-shear ${res.status}: ${body}`);
    }
    const shear = await res.json();
    return { decision: mapVerdict(shear.verdict), shear };
  }
}
