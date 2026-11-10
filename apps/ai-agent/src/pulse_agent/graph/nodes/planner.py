from __future__ import annotations

from pulse_agent.graph.planning import plan_turn


def planner(state: dict) -> dict:
    plan = plan_turn(
        str(state.get("body") or ""),
        local_date=str(state.get("local_date") or ""),
        timezone=str(state.get("timezone") or ""),
    )
    return {"plan": plan.__dict__}
