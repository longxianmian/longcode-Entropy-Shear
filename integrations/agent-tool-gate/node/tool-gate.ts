// Agent Tool Gate (TypeScript reference).
//
// This module is the *gate layer* between an Agent's tool selector and
// its tool executor. It does NOT implement an Agent. It does NOT
// generate plans, summarize, or call an LLM. It only:
//
//   1. accepts an action the Agent has already chosen,
//   2. constructs `facts` from that action,
//   3. POSTs (policy, facts) to Entropy Shear,
//   4. returns one of allow / deny / hold to the caller.
//
// The caller is responsible for actually executing, refusing, or
// queueing-for-human the action.

import type {
  EntropyShearClient,
  Policy,
  ShearResult,
  Verdict,
} from "../../../sdk/js/src/index";

export type GateDecision = "allow" | "deny" | "hold";

export interface AgentAction {
  /** Logical action name, e.g. `delete_production_data` or `transfer_funds`. */
  name: string;
  /** Coarse classification, e.g. `data` / `payment` / `email`. */
  category: string;
  /** Whether the action only reads or also writes / mutates. */
  mode: "readonly" | "write";
  /** Optional resource the action targets. Pure metadata. */
  target?: string;
  /** Optional caller-provided context blob, recorded into the trace's input hash. */
  context?: Record<string, unknown>;
}

export interface AgentToolGateOptions {
  client: EntropyShearClient;
  policy: Policy;
  /** Identifier of the agent issuing the action. */
  agentId?: string;
  /** Logical team the agent belongs to. */
  agentTeam?: string;
}

export interface GateResponse {
  decision: GateDecision;
  shear: ShearResult;
}

/**
 * AgentToolGate is a thin adapter: agent action → /shear → decision.
 * It deliberately keeps no local state; every call hits the engine so
 * the ledger reflects the full audit trail.
 */
export class AgentToolGate {
  private readonly client: EntropyShearClient;
  private readonly policy: Policy;
  private readonly agentId: string;
  private readonly agentTeam: string;

  constructor(opts: AgentToolGateOptions) {
    this.client = opts.client;
    this.policy = opts.policy;
    this.agentId = opts.agentId ?? "agent-unknown";
    this.agentTeam = opts.agentTeam ?? "default";
  }

  async gate(action: AgentAction): Promise<GateResponse> {
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
    const shear = await this.client.shear({ policy: this.policy, facts });
    return { decision: mapVerdict(shear.verdict), shear };
  }
}

export function mapVerdict(v: Verdict): GateDecision {
  switch (v) {
    case "Yes":
      return "allow";
    case "No":
      return "deny";
    case "Hold":
    default:
      return "hold";
  }
}
