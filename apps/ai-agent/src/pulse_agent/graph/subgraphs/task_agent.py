from __future__ import annotations

from pulse_agent.graph.task_extraction import parse_task_draft
from pulse_agent.schemas.cards import task_create_card


def build_task_create_proposal(
    text: str, *, local_date: str = "", timezone: str = "", card_id: str = "task-create"
) -> dict:
    draft = parse_task_draft(text, local_date=local_date, timezone=timezone)
    return task_create_card(card_id, draft.to_tool_input()).model_dump(mode="json", by_alias=True)
