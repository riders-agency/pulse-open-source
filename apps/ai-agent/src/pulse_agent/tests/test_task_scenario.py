from __future__ import annotations

from pulse_agent.graph.planning import plan_turn
from pulse_agent.graph.task_extraction import parse_task_draft
from pulse_agent.scenarios import SCENARIOS


def test_exact_russian_task_scenario() -> None:
    draft = parse_task_draft(
        "создай задачу мне для клиента Орион завтра сделать все правка для сеошников и сообщить им тоже",
        local_date="2026-06-15",
        timezone="Europe/Moscow",
    )
    assert draft.client_query == "Орион"
    assert draft.title == "Внести правки для SEO-специалистов"
    assert draft.description == "Сделать все правки для SEO-специалистов и сообщить им о результате."
    assert draft.due_date == "2026-06-16"
    assert draft.due_date_label == "завтра"
    assert draft.priority == "medium"
    assert draft.labels == ["ai-created", "seo"]
    assert draft.assignee_hint == "me"
    assert draft.notification_intent == "Сообщить SEO-специалистам после выполнения"
    assert draft.checklist == [
        "Собрать список правок",
        "Внести правки",
        "Проверить результат",
        "Сообщить SEO-специалистам",
    ]


def test_exact_russian_task_is_not_misrouted_as_client_create() -> None:
    plan = plan_turn(
        "создай задачу мне для клиента Орион завтра сделать все правка для сеошников и сообщить им тоже",
        local_date="2026-06-15",
        timezone="Europe/Moscow",
    )
    assert plan.intent == "task.create"
    assert plan.required_tools == ["client.search", "task.create"]
    assert plan.tool_input_preview["client_query"] == "Орион"


def test_compound_create_client_and_task_plan() -> None:
    plan = plan_turn(
        "Создай клиента Орион и поставь ему задачу подготовить договор на продвижение и запусти задачу тоже",
        local_date="2026-06-15",
        timezone="Europe/Moscow",
    )
    assert plan.intent == "compound.client_create_and_task_create"
    assert plan.requires_approval is True
    assert plan.required_tools == ["client.create", "task.create"]
    assert plan.tool_input_preview["client"]["name"] == "Орион"
    assert plan.tool_input_preview["task"]["client_query"] == "Орион"
    assert plan.tool_input_preview["task"]["title"] == "Подготовить договор на продвижение"
    assert plan.tool_input_preview["task"]["start_task_after_create"] is True


def test_weekday_temporal_expression_uses_supplied_local_date() -> None:
    draft = parse_task_draft(
        "Создай задачу созвониться с Орион в следующий понедельник",
        local_date="2026-06-15",
        timezone="Europe/Moscow",
    )
    assert draft.due_date == "2026-06-22"
    assert draft.due_date_label == "понедельник"


def test_thirty_acceptance_scenarios_are_registered() -> None:
    assert len(SCENARIOS) == 30
    ids = [str(item["id"]) for item in SCENARIOS]
    assert len(ids) == len(set(ids))
    assert "compound_create_client_and_task" in ids
    assert "interrupt_new_task_over_pending" in ids


def test_all_registered_intent_scenarios_route_deterministically() -> None:
    for scenario in SCENARIOS:
        expected = scenario.get("expected_intent")
        if not expected:
            continue
        plan = plan_turn(str(scenario["prompt"]), local_date="2026-06-15", timezone="Europe/Moscow")
        assert plan.intent == expected, scenario["id"]
