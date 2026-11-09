from __future__ import annotations

from typing import Any

from pulse_agent.graph.intent_schema import IntentKey, classify_turn, proposal_gate_decision
from pulse_agent.graph.turn_understanding import deterministic_understanding, model_prompt


def test_explicit_task_create_is_new_command_even_with_active_proposal() -> None:
    for text in (
        "Создай задачу проверить оплату Орион завтра",
        "Нет, лучше создай отдельную задачу позвонить Орион в пятницу",
    ):
        classification = classify_turn(
            text,
            active_tool_call_id="pending-task-create",
        )

        assert classification.intent_key == IntentKey.TASK_CREATE
        assert classification.is_new_command is True

        gate = proposal_gate_decision(classification, "pending-task-create")
        assert gate.decision == "run_new_keep_pending"


def test_read_command_keeps_pending_proposal_but_routes_new_read() -> None:
    for text in ("Найди задачи по Орион", "Покажи большой отчет по задачам"):
        classification = classify_turn(text, active_tool_call_id="pending-task-create")
        gate = proposal_gate_decision(classification, "pending-task-create")

        assert classification.intent_key == IntentKey.TASK_SEARCH
        assert gate.decision == "run_new_keep_pending"


def test_payment_lookup_handles_context_switch_pronoun_phrase() -> None:
    classification = classify_turn("покажи оплаты по нему", active_tool_call_id="pending-task-create")

    assert classification.intent_key == IntentKey.PAYMENT_LOOKUP_DUE
    assert classification.target_entity_type == "payment"


def test_project_payment_lookup_does_not_require_client_search() -> None:
    for text in ("у кого ближайшие оплаты", "какие ближайшие оплаты по проекту"):
        classification = classify_turn(text)

        assert classification.intent_key == IntentKey.PAYMENT_LOOKUP_DUE
        assert classification.reason == "payment_project_lookup_due"
        assert [tool.tool_key for tool in classification.tool_candidates] == ["payment.lookup_due"]


def test_named_client_payment_lookup_keeps_client_resolution() -> None:
    classification = classify_turn("когда следующая оплата у Орион")

    assert classification.intent_key == IntentKey.PAYMENT_LOOKUP_DUE
    assert classification.reason == "payment_lookup_due"
    assert [tool.tool_key for tool in classification.tool_candidates] == ["client.search", "payment.lookup_due"]


def test_project_explain_and_ai_integration_are_read_context() -> None:
    for text in ("Объясни текущий проект", "Проверь AI интеграцию проекта"):
        classification = classify_turn(text)

        assert classification.intent_key == IntentKey.PROJECT_EXPLAIN
        assert classification.is_new_command is True
        assert classification.tool_candidates[0].tool_key == "context.read"


def test_client_list_and_count_are_distinct_read_intents() -> None:
    listing = classify_turn("какие у меня клиенты")
    count = classify_turn("сколько у меня клиентов")

    assert listing.intent_key == IntentKey.CLIENT_SEARCH
    assert listing.tool_candidates[0].tool_key == "client.search"
    assert count.intent_key == IntentKey.CLIENT_COUNT
    assert count.tool_candidates[0].tool_key == "client.count"


def test_proposal_edit_phrase_patches_current_proposal() -> None:
    classification = classify_turn("на сегодня", active_tool_call_id="pending-task-create")
    gate = proposal_gate_decision(classification, "pending-task-create")

    assert classification.intent_key == IntentKey.PROPOSAL_EDIT
    assert gate.decision == "patch_current"


def test_workflow_question_is_read_not_patch() -> None:
    classification = classify_turn(
        "сколько здесь действий",
        page_context={"pageKind": "workflow_editor", "workflowId": "wf-1"},
    )

    assert classification.intent_key == IntentKey.WORKFLOW_ANSWER_QUESTION
    assert classification.tool_candidates[0].tool_key == "workflow.read_draft"


def test_workflow_patch_requires_critical_proposal() -> None:
    classification = classify_turn(
        "сделай чтобы после оплаты отправлялся email",
        page_context={"pageKind": "workflow_editor", "workflowId": "wf-1"},
    )

    assert classification.intent_key == IntentKey.WORKFLOW_PATCH
    assert classification.tool_candidates[0].risk_level == "critical"


def test_deictic_task_move_requires_selected_task() -> None:
    missing = classify_turn("Перенеси ее на завтра", page_context={"pageKind": "task_board"})
    assert missing.intent_key == IntentKey.UNKNOWN
    assert missing.requires_clarification is True

    selected = classify_turn("Запусти эту задачу", page_context={"pageKind": "task", "taskId": "task-1"})
    assert selected.intent_key == IntentKey.TASK_MOVE
    assert selected.target_entity_id == "task-1"
    assert selected.deictic_reference is True


def test_deterministic_understanding_exposes_codes_not_reasoning() -> None:
    classification = classify_turn("Найди клиента Орион")
    understanding = deterministic_understanding(classification)
    payload = understanding.model_dump(by_alias=True)

    assert payload["schema"] == "riders-pulse-turn-understanding-v1"
    assert payload["source"] == "deterministic"
    assert payload["intent_key"] == IntentKey.CLIENT_SEARCH
    assert payload["rationale_code"] == "client_search"
    assert "rationale" not in payload
    assert "chain_of_thought" not in payload


def test_model_prompt_exposes_expanded_intent_catalog() -> None:
    prompt = model_prompt()

    for key in (
        "client.search",
        "client.read",
        "task.read",
        "payment.schedule.read",
        "payment.action.create",
        "integration.catalog.list",
        "integration.dependency.read",
        "integration.diagnostics",
        "integration.secret.rotate",
        "prompt.copy",
    ):
        assert key in prompt


def test_model_turn_understanding_routes_client_list_without_regex(monkeypatch: Any) -> None:
    monkeypatch.setenv("PULSE_AI_MODEL_TURN_UNDERSTANDING", "true")

    class FakeModelGateway:
        def __init__(self, *_args: Any, **_kwargs: Any) -> None:
            pass

        async def intent_router(self, *_args: Any, **_kwargs: Any) -> dict[str, Any]:
            return {
                "response_json": {
                    "schema": "riders-pulse-turn-understanding-v1",
                    "intent_key": "client.search",
                    "entities": [],
                    "rationale_code": "client_list",
                }
            }

    monkeypatch.setattr("pulse_agent.graph.turn_understanding.PulseModelGatewayClient", FakeModelGateway)

    classification = classify_turn("выведи клиентов")

    assert classification.intent_key == IntentKey.CLIENT_SEARCH
    assert classification.confidence == 0.82
    assert classification.reason == "client_list"
    assert classification.entities == []
    assert classification.tool_candidates[0].tool_key == "client.search"


def test_model_turn_understanding_routes_named_client_payment_lookup(monkeypatch: Any) -> None:
    monkeypatch.setenv("PULSE_AI_MODEL_TURN_UNDERSTANDING", "true")

    class FakeModelGateway:
        def __init__(self, *_args: Any, **_kwargs: Any) -> None:
            pass

        async def intent_router(self, *_args: Any, **_kwargs: Any) -> dict[str, Any]:
            return {
                "response_json": {
                    "schema": "riders-pulse-turn-understanding-v1",
                    "discourse_act": "question",
                    "intent_key": "payment.lookup_due",
                    "domain": "payment",
                    "confidence": 0.9,
                    "is_new_command": True,
                    "entities": [
                        {
                            "entity_type": "client",
                            "query": "Орион",
                            "confidence": 0.82,
                            "source": "user",
                        }
                    ],
                    "rationale_code": "payment_lookup_due",
                }
            }

    monkeypatch.setattr("pulse_agent.graph.turn_understanding.PulseModelGatewayClient", FakeModelGateway)

    classification = classify_turn("billing date for advanche?")

    assert classification.intent_key == IntentKey.PAYMENT_LOOKUP_DUE
    assert classification.entities[0].query == "Орион"
    assert [tool.tool_key for tool in classification.tool_candidates] == ["client.search", "payment.lookup_due"]


def test_model_turn_understanding_routes_project_payment_lookup(monkeypatch: Any) -> None:
    monkeypatch.setenv("PULSE_AI_MODEL_TURN_UNDERSTANDING", "true")

    class FakeModelGateway:
        def __init__(self, *_args: Any, **_kwargs: Any) -> None:
            pass

        async def intent_router(self, *_args: Any, **_kwargs: Any) -> dict[str, Any]:
            return {
                "response_json": {
                    "schema": "riders-pulse-turn-understanding-v1",
                    "discourse_act": "question",
                    "intent_key": "unknown",
                    "domain": "payment",
                    "is_new_command": True,
                    "slot_updates": {"scope": "project"},
                    "rationale_code": "payment_project_lookup_due",
                }
            }

    monkeypatch.setattr("pulse_agent.graph.turn_understanding.PulseModelGatewayClient", FakeModelGateway)

    classification = classify_turn("who has nearest upcoming payments?")

    assert classification.intent_key == IntentKey.PAYMENT_LOOKUP_DUE
    assert classification.reason == "payment_project_lookup_due"
    assert [tool.tool_key for tool in classification.tool_candidates] == ["payment.lookup_due"]


def test_model_turn_understanding_routes_paraphrased_read(monkeypatch: Any) -> None:
    monkeypatch.setenv("PULSE_AI_MODEL_TURN_UNDERSTANDING", "true")

    class FakeModelGateway:
        def __init__(self, *_args: Any, **_kwargs: Any) -> None:
            pass

        async def intent_router(self, *_args: Any, **_kwargs: Any) -> dict[str, Any]:
            return {
                "response_json": {
                    "schema": "riders-pulse-turn-understanding-v1",
                    "discourse_act": "question",
                    "intent_key": "client.search",
                    "domain": "client",
                    "confidence": 0.86,
                    "is_new_command": True,
                    "entities": [
                        {
                            "entity_type": "client",
                            "query": "Adv",
                            "confidence": 0.8,
                            "source": "user",
                        }
                    ],
                    "rationale_code": "model_client_lookup",
                }
            }

    monkeypatch.setattr("pulse_agent.graph.turn_understanding.PulseModelGatewayClient", FakeModelGateway)

    classification = classify_turn("could you surface the Adv account?")

    assert classification.intent_key == IntentKey.CLIENT_SEARCH
    assert classification.tool_candidates[0].tool_key == "client.search"
    assert classification.entities[0].query == "Adv"
    assert classification.reason == "model_client_lookup"


def test_model_turn_understanding_blocks_high_risk_write_without_flag(monkeypatch: Any) -> None:
    monkeypatch.setenv("PULSE_AI_MODEL_TURN_UNDERSTANDING", "true")
    monkeypatch.delenv("PULSE_AI_MODEL_WRITE_INTENTS", raising=False)

    class FakeModelGateway:
        def __init__(self, *_args: Any, **_kwargs: Any) -> None:
            pass

        async def intent_router(self, *_args: Any, **_kwargs: Any) -> dict[str, Any]:
            return {
                "response_json": {
                    "schema": "riders-pulse-turn-understanding-v1",
                    "discourse_act": "command",
                    "intent_key": "integration.secret.rotate",
                    "domain": "integration",
                    "confidence": 0.96,
                    "is_new_command": True,
                    "clarification_question": "Подтвердите ротацию секрета.",
                    "rationale_code": "model_secret_rotate",
                }
            }

    monkeypatch.setattr("pulse_agent.graph.turn_understanding.PulseModelGatewayClient", FakeModelGateway)

    classification = classify_turn("please rotate the apinet credential material")

    assert classification.intent_key == IntentKey.UNKNOWN
    assert classification.requires_clarification is True
    assert classification.reason == "model_write_intent_blocked"


def test_model_turn_understanding_low_confidence_requires_clarification(monkeypatch: Any) -> None:
    monkeypatch.setenv("PULSE_AI_MODEL_TURN_UNDERSTANDING", "true")

    class FakeModelGateway:
        def __init__(self, *_args: Any, **_kwargs: Any) -> None:
            pass

        async def intent_router(self, *_args: Any, **_kwargs: Any) -> dict[str, Any]:
            return {
                "response_json": {
                    "schema": "riders-pulse-turn-understanding-v1",
                    "discourse_act": "unknown",
                    "intent_key": "client.search",
                    "domain": "client",
                    "confidence": 0.41,
                    "clarification_question": "Уточните, какого клиента искать.",
                    "rationale_code": "model_ambiguous",
                }
            }

    monkeypatch.setattr("pulse_agent.graph.turn_understanding.PulseModelGatewayClient", FakeModelGateway)

    classification = classify_turn("that one from yesterday")

    assert classification.intent_key == IntentKey.UNKNOWN
    assert classification.confidence == 0.41
    assert classification.requires_clarification is True
    assert classification.reason == "model_low_confidence"
