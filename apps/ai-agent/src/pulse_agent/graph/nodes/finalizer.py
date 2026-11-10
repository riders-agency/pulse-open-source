from __future__ import annotations


def finalizer(state: dict) -> dict:
    if state.get("final_response"):
        return {}
    plan = state.get("plan") or {}
    return {"final_response": plan.get("user_visible_summary") or "Готово."}
