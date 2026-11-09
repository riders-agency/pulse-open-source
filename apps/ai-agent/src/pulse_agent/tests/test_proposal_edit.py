from __future__ import annotations

from pulse_agent.graph.proposal_edit import resolve_task_proposal_edit


def test_proposal_edit_today_updates_same_proposal_input() -> None:
    current = {"tool_call_id": "tool-1", "card_id": "card-1", "due_date": "2026-06-16"}
    edit = resolve_task_proposal_edit("давай сегодня", current, local_date="2026-06-15")
    assert edit.changed is True
    assert edit.patch["due_date"] == "2026-06-15"
    assert edit.patch["due_date_label"] == "сегодня"
    assert "срок" in edit.human_summary


def test_unknown_followup_does_not_mutate_proposal() -> None:
    edit = resolve_task_proposal_edit("что там по оплате?", {"title": "x"}, local_date="2026-06-15")
    assert edit.changed is False
    assert edit.patch == {}


def test_task_search_with_active_proposal_routes_as_new_request() -> None:
    edit = resolve_task_proposal_edit("Найди задачи по Орион", {"title": "x"}, local_date="2026-06-15")
    assert edit.changed is False
    assert edit.patch == {}


def test_explicit_task_create_with_active_proposal_routes_as_new_request() -> None:
    edit = resolve_task_proposal_edit(
        "Создай задачу проверить оплату Орион завтра",
        {"title": "x", "due_date": "2026-06-15"},
        local_date="2026-06-15",
    )
    assert edit.changed is False
    assert edit.patch == {}


def test_client_edit_strips_preposition() -> None:
    edit = resolve_task_proposal_edit("поменяй клиента на Ромашка", {"title": "x"}, local_date="2026-06-15")
    assert edit.changed is True
    assert edit.patch["client_query"] == "Ромашка"
