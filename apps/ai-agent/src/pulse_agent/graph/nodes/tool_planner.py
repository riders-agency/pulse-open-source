from __future__ import annotations


def tool_planner(state: dict) -> dict:
    plan = state.get("plan") or {}
    return {"candidate_tools": plan.get("required_tools") or []}
