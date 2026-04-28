"""Agent Tool Gate (Python reference).

This module is the *gate layer* between an Agent's tool selector and
its tool executor. It does not implement an Agent. It does not call an
LLM. It does one thing: take an action the agent has chosen, ask
Entropy Shear for a verdict, and return one of allow / deny / hold.

The caller is responsible for actually executing, refusing, or
queueing-for-human the action.
"""

from __future__ import annotations

import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Optional

# Make the SDK importable when this file is run directly from a checkout.
_REPO_ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(_REPO_ROOT / "sdk" / "python"))

from entropy_shear_client import EntropyShearClient  # noqa: E402

__all__ = ["AgentAction", "AgentToolGate", "GateResponse", "map_verdict"]


@dataclass
class AgentAction:
    name: str
    category: str
    mode: str  # "readonly" | "write"
    target: Optional[str] = None
    context: dict[str, Any] = field(default_factory=dict)


@dataclass
class GateResponse:
    decision: str  # "allow" | "deny" | "hold"
    shear: dict[str, Any]


def map_verdict(verdict: str) -> str:
    if verdict == "Yes":
        return "allow"
    if verdict == "No":
        return "deny"
    return "hold"


class AgentToolGate:
    """Thin adapter: agent action -> /shear -> decision."""

    def __init__(
        self,
        *,
        client: EntropyShearClient,
        policy: dict[str, Any],
        agent_id: str = "agent-unknown",
        agent_team: str = "default",
    ) -> None:
        self.client = client
        self.policy = policy
        self.agent_id = agent_id
        self.agent_team = agent_team

    def gate(self, action: AgentAction) -> GateResponse:
        action_payload: dict[str, Any] = {
            "name": action.name,
            "category": action.category,
            "mode": action.mode,
        }
        if action.target is not None:
            action_payload["target"] = action.target
        if action.context:
            action_payload["context"] = action.context

        facts = {
            "agent": {"id": self.agent_id, "team": self.agent_team},
            "action": action_payload,
        }
        shear = self.client.shear(policy=self.policy, facts=facts)
        return GateResponse(decision=map_verdict(shear["verdict"]), shear=shear)
