from __future__ import annotations

from pulse_agent.security import assert_project_scope


def policy_guard(state: dict) -> dict:
    assert_project_scope(
        str(state.get("company_id") or ""), str(state.get("project_id") or ""), str(state.get("user_id") or "")
    )
    return {"policy_ok": True}
