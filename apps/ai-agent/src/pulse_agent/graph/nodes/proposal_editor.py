from __future__ import annotations

from pulse_agent.graph.proposal_edit import resolve_task_proposal_edit


def proposal_editor(state: dict) -> dict:
    current = ((state.get("active_card") or {}).get("debug_json") or {}).get("input_json") or {}
    edit = resolve_task_proposal_edit(
        str(state.get("body") or ""), current, local_date=str(state.get("local_date") or "")
    )
    return {"proposal_edit": edit.__dict__}
