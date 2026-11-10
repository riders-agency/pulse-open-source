from __future__ import annotations

from pulse_agent.graph.task_extraction import parse_task_draft
from pulse_agent.schemas.cards import CardStatus, task_create_card


def test_task_create_card_hides_debug_json_and_requires_approval() -> None:
    draft = parse_task_draft("создай задачу для клиента Орион завтра правки seo", local_date="2026-06-15")
    card = task_create_card("card-1", draft.to_tool_input())
    payload = card.model_dump(mode="json", by_alias=True)
    assert card.card_type == "task_create_proposal"
    assert payload["schema"] == "riders-pulse-ai-card-v1"
    assert card.intent_key == "task.create"
    assert card.requires_approval is True
    assert card.status == CardStatus.NEEDS_APPROVAL
    assert card.debug_json["hidden_by_default"] is True
    assert any(action.action_key == "approve" and action.confirm_required for action in card.actions)
    assert any(field.key == "client" and field.editable for field in card.fields)
    assert any(field.key == "due_date" and field.value == "завтра · 16.06.2026" for field in card.fields)
    assert payload["display_json"]["due_date_display"] == "завтра · 16.06.2026"
    assert payload["input_json"]["due_date"] == "2026-06-16"
