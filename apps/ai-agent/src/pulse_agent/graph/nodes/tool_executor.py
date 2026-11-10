from __future__ import annotations


def tool_executor(state: dict) -> dict:
    decision = state.get("approval_decision") or {}
    if decision.get("action") != "approve":
        return {"tool_results": [{"status": "skipped", "reason": "approval_required"}]}
    return {"tool_results": [{"status": "delegated_to_go_tool_gateway"}]}
