"""Runnable demo of AgentToolGate.

Usage::

    docker compose up -d --build         # from the repo root
    python3 integrations/agent-tool-gate/python/example.py
"""

from __future__ import annotations

import json
import os
import sys
from pathlib import Path

# Make the SDK and gate importable regardless of cwd.
_REPO_ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(_REPO_ROOT / "sdk" / "python"))
sys.path.insert(0, str(Path(__file__).resolve().parent))

from entropy_shear_client import EntropyShearClient  # noqa: E402
from tool_gate import AgentAction, AgentToolGate  # noqa: E402

POLICY_PATH = _REPO_ROOT / "policies" / "agent" / "agent-action-policy.v1.json"


def main() -> None:
    policy = json.loads(POLICY_PATH.read_text("utf-8"))
    client = EntropyShearClient(
        base_url=os.environ.get("ENTROPY_SHEAR_URL", "http://127.0.0.1:8080")
    )
    gate = AgentToolGate(client=client, policy=policy, agent_id="agent-007", agent_team="ops")

    cases = [
        (
            "readonly query        (expect allow)",
            AgentAction(name="list_active_users", category="data", mode="readonly", target="user_table"),
        ),
        (
            "delete production     (expect deny)",
            AgentAction(name="delete_production_data", category="data", mode="write", target="orders_2025"),
        ),
        (
            "transfer funds        (expect hold)",
            AgentAction(name="transfer_funds", category="payment", mode="write", target="wallet:U-9001"),
        ),
    ]

    for label, action in cases:
        res = gate.gate(action)
        rule = res.shear.get("applied_rule_id") or "<default>"
        print(
            f"{label}  decision={res.decision}  verdict={res.shear['verdict']}  "
            f"rule={rule}  shear_id={res.shear['shear_id']}"
        )


if __name__ == "__main__":
    main()
