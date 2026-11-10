from __future__ import annotations

import pytest

from pulse_agent.graph.intent_schema import IntentKey, TurnClassification, TurnKind
from pulse_agent.graph.turn_nodes import calculate_expression, plan_for_classification, route_branch


def test_turn_nodes_plan_for_classification_uses_structured_plan_dict() -> None:
    plan = plan_for_classification(
        TurnClassification(
            turn_kind=TurnKind.WRITE_INTENT,
            intent_key=IntentKey.TASK_CREATE,
            confidence=0.95,
            is_new_command=True,
        ),
        "Создай задачу проверить оплату завтра",
        local_date="2026-06-25",
        timezone="Europe/Moscow",
    )

    assert plan["intent"] == "task.create"
    assert plan["requires_approval"] is True
    assert isinstance(plan["tool_input_preview"], dict)


def test_turn_nodes_calculator_and_route_branch_are_isolated() -> None:
    assert calculate_expression("2 + 3 * 4") == "14"
    assert route_branch({"branch": "read_context"}) == "read_context"
    assert route_branch({"plan": {"intent": "client.search"}}) == "read_context"


@pytest.mark.parametrize("body", [
    "Запланируй задачу подготовить презентацию Pulse на завтра.",
    "Добавь задачу проверить модель APInet завтра",
    "Поставь задачу настроить email завтра",
])
def test_classified_creation_cannot_be_reinterpreted_by_keywords(body: str) -> None:
    plan = plan_for_classification(
        TurnClassification(turn_kind=TurnKind.WRITE_INTENT, intent_key=IntentKey.TASK_CREATE,
                           confidence=0.91, is_new_command=True, understanding_source="model"),
        body, local_date="2026-11-09", timezone="Europe/Moscow",
    )
    assert plan["intent"] == "task.create"
    assert plan["risk_level"] == "high"
    assert plan["requires_approval"] is True
    assert plan["required_tools"] == ["client.search", "task.create"]
    assert plan["tool_input_preview"]["due_date"] == "2026-11-10"
