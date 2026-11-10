from __future__ import annotations

from typing import Any

from pulse_agent.graph import tool_gateway
from pulse_agent.graph.builder import build_graph
from pulse_agent.graph.planning import plan_turn


def test_compiled_graph_emits_task_create_card_and_approval() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()

    state = graph.invoke(
        {
            "project_id": "project-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "Создай задачу проверить оплату Орион завтра",
            "local_date": "2026-06-15",
            "timezone": "Europe/Moscow",
        }
    )

    events = state["events"]
    assert [event["seq"] for event in events] == sorted(event["seq"] for event in events)
    assert any(event["event_type"] == "CARD_CREATED" for event in events)
    assert any(event["event_type"] == "APPROVAL_REQUIRED" for event in events)
    assert state["active_mode"] == "task_create"


def test_task_search_routes_to_read_tool_not_pending_edit() -> None:
    plan = plan_turn("Найди задачи по Орион", local_date="2026-06-15", timezone="Europe/Moscow")
    assert plan.intent == "task.search"

    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()
    state = graph.invoke(
        {
            "project_id": "project-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "Найди задачи по Орион",
            "active_tool_call_id": "pending-task-create",
            "local_date": "2026-06-15",
            "timezone": "Europe/Moscow",
        }
    )

    tool_events = [event for event in state["events"] if event["event_type"] == "TOOL_STARTED"]
    assert tool_events
    assert tool_events[-1]["tool_call"]["tool_key"] == "task.search"
    assert state["active_mode"] == "project_read"


def test_graph_loop_executes_read_tool_inside_python_graph(monkeypatch) -> None:
    calls: list[dict[str, Any]] = []

    class FakePulseToolsClient:
        def __init__(self, target: str, service_token: str = "") -> None:
            self.target = target
            self.service_token = service_token

        async def run_read_tool(self, **kwargs: Any) -> dict[str, Any]:
            calls.append(kwargs)
            return {
                "ok": True,
                "status": "succeeded",
                "output_json": {"answer": "Нашел 2 задачи", "items": [{"id": "task-1"}]},
                "user_message": "",
                "audit_event_id": "audit-1",
            }

    monkeypatch.setenv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP", "true")
    monkeypatch.setattr(tool_gateway, "PulseToolsClient", FakePulseToolsClient)
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()

    state = graph.invoke(
        {
            "company_id": "company-1",
            "project_id": "project-1",
            "user_id": "user-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "Найди задачи по Орион",
        }
    )

    assert calls and calls[-1]["tool_key"] == "task.search"
    tool_results = [event for event in state["events"] if event["event_type"] == "TOOL_RESULT"]
    assert tool_results[-1]["tool_call"]["output_json"]["answer"] == "Нашел 2 задачи"
    assert state["awaiting_tool_observation"] is False
    assert state["final_response"] == "Нашел 2 задачи"
    assert state["tool_observations"][-1]["schema"] == "riders-pulse-tool-observation-v1"
    assert state["tool_observations"][-1]["normalized_output"]["items"] == [{"id": "task-1"}]
    final_events = [event for event in state["events"] if event["event_type"] == "MESSAGE_FINAL"]
    grounded_response = final_events[-1]["tool_call"]["grounded_response"]
    assert grounded_response["schema"] == "riders-pulse-grounded-final-response-v1"
    assert grounded_response["response_status"] == "answered"
    assert grounded_response["evidence_refs"][0]["audit_ref"] == "audit-1"
    assert grounded_response["entity_refs"][0]["entity_type"] == "task"
    assert state["message_payload_json"] == grounded_response
    done_events = [event for event in state["events"] if event["event_type"] == "DONE"]
    observability = done_events[-1]["tool_call"]["ai_observability"]
    assert observability["schema"] == "riders-pulse-ai-observability-v1"
    assert observability["intent"]["key"] == "task.search"
    assert observability["tools"]["items"][0]["tool_key"] == "task.search"
    assert observability["tools"]["items"][0]["latency_ms"] >= 0
    assert "Нашел 2 задачи" not in str(observability)
    assert state["ai_observability_json"] == observability


def test_graph_loop_replans_next_read_tool_from_observation(monkeypatch) -> None:
    calls: list[dict[str, Any]] = []

    class FakePulseToolsClient:
        def __init__(self, target: str, service_token: str = "") -> None:
            self.target = target
            self.service_token = service_token

        async def run_read_tool(self, **kwargs: Any) -> dict[str, Any]:
            calls.append(kwargs)
            if kwargs["tool_key"] == "client.search":
                return {
                    "ok": True,
                    "status": "succeeded",
                    "output_json": {
                        "answer": "Нашел клиента Орион",
                        "items": [{"id": "client-1", "name": "Орион", "raw": {"token": "hidden"}}],
                    },
                    "user_message": "",
                    "audit_event_id": "audit-client",
                }
            return {
                "ok": True,
                "status": "succeeded",
                "output_json": {"answer": "Следующая оплата Орион 30 июня", "provider_payload": {"secret": "x"}},
                "user_message": "",
                "audit_event_id": "audit-payment",
            }

    monkeypatch.setenv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP", "true")
    monkeypatch.setattr(tool_gateway, "PulseToolsClient", FakePulseToolsClient)
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()

    state = graph.invoke(
        {
            "company_id": "company-1",
            "project_id": "project-1",
            "user_id": "user-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "Когда оплата у Орион?",
        }
    )

    assert [call["tool_key"] for call in calls] == ["client.search", "payment.lookup_due"]
    assert calls[1]["input_json"]["client_id"] == "client-1"
    assert calls[1]["input_json"]["client_name"] == "Орион"
    observations = state["tool_observations"]
    assert [item["tool_key"] for item in observations] == ["client.search", "payment.lookup_due"]
    assert observations[0]["supporting_only"] is True
    assert observations[0]["normalized_output"]["items"][0]["raw"] == "[redacted]"
    assert observations[1]["normalized_output"]["provider_payload"] == "[redacted]"
    assert state["replan_trace"][-1]["decision"] == "finalize"
    assert state["final_response"] == "Следующая оплата Орион 30 июня"


def test_project_payment_lookup_skips_client_resolution(monkeypatch) -> None:
    calls: list[dict[str, Any]] = []

    class FakePulseToolsClient:
        def __init__(self, target: str, service_token: str = "") -> None:
            self.target = target
            self.service_token = service_token

        async def run_read_tool(self, **kwargs: Any) -> dict[str, Any]:
            calls.append(kwargs)
            return {
                "ok": True,
                "status": "succeeded",
                "output_json": {"answer": "Ближайшие оплаты: Орион — 2026-06-30, 1000.00 RUB."},
                "user_message": "",
                "audit_event_id": "audit-payments",
            }

    monkeypatch.setenv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP", "true")
    monkeypatch.setattr(tool_gateway, "PulseToolsClient", FakePulseToolsClient)
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()

    state = graph.invoke(
        {
            "company_id": "company-1",
            "project_id": "project-1",
            "user_id": "user-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "у кого ближайшие оплаты",
        }
    )

    assert [call["tool_key"] for call in calls] == ["payment.lookup_due"]
    assert calls[0]["input_json"]["scope"] == "project"
    assert calls[0]["input_json"]["query"] == ""
    assert state["intent"]["intent_key"] == "payment.lookup_due"
    assert state["final_response"].startswith("Ближайшие оплаты:")


def test_graph_loop_read_before_write_keeps_write_approval(monkeypatch) -> None:
    calls: list[dict[str, Any]] = []

    class FakePulseToolsClient:
        def __init__(self, target: str, service_token: str = "") -> None:
            self.target = target
            self.service_token = service_token

        async def run_read_tool(self, **kwargs: Any) -> dict[str, Any]:
            calls.append(kwargs)
            return {
                "ok": True,
                "status": "succeeded",
                "output_json": {"answer": "Клиент найден", "items": [{"id": "client-1", "name": "Орион"}]},
                "user_message": "",
                "audit_event_id": "audit-client",
            }

    monkeypatch.setenv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP", "true")
    monkeypatch.setattr(tool_gateway, "PulseToolsClient", FakePulseToolsClient)
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()

    state = graph.invoke(
        {
            "company_id": "company-1",
            "project_id": "project-1",
            "user_id": "user-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "Создай задачу для клиента Орион проверить оплату завтра",
            "local_date": "2026-06-15",
            "timezone": "Europe/Moscow",
        }
    )

    assert [call["tool_key"] for call in calls] == ["client.search"]
    cards = [event["card"] for event in state["events"] if event["event_type"] == "CARD_CREATED"]
    assert cards[-1]["card_type"] == "task_create_proposal"
    assert cards[-1]["input_json"]["client_id"] == "client-1"
    assert cards[-1]["input_json"]["client_name"] == "Орион"
    assert cards[-1]["input_json"]["client_resolved"] is True
    assert cards[-1]["input_json"]["missing_fields"] == []
    assert [field for field in cards[-1]["fields"] if field["key"] == "client"][-1]["state"] == "normal"
    assert cards[-1]["debug_json"]["tool_observations"][0]["tool_key"] == "client.search"
    assert state["active_mode"] == "task_create"
    assert state["current_step"] == "await_approval"


def test_new_task_create_while_pending_switches_to_new_task_card() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()
    state = graph.invoke(
        {
            "project_id": "project-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "Создай задачу позвонить Орион завтра",
            "active_tool_call_id": "pending-task-create",
            "local_date": "2026-06-15",
            "timezone": "Europe/Moscow",
        }
    )

    cards = [event["card"] for event in state["events"] if event["event_type"] == "CARD_CREATED"]
    assert cards[-1]["card_type"] == "task_create_proposal"
    assert state["proposal_gate"]["decision"] == "run_new_keep_pending"
    assert state["active_mode"] == "task_create"
    assert state["current_step"] == "await_approval"


def test_task_create_after_pending_client_card_switches_context() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()
    state = graph.invoke(
        {
            "project_id": "project-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "Создай задачу проверить оплату Орион завтра",
            "active_tool_call_id": "pending-client-create",
            "active_card": {
                "card_id": "card-client-create",
                "card_type": "client_create_proposal",
                "intent_key": "client.create",
                "risk_level": "high",
                "status": "needs_approval",
                "debug_json": {"input_json": {"name": "Орион"}},
            },
            "pending_interrupt": {"active_tool_call_id": "pending-client-create", "card_id": "card-client-create"},
            "local_date": "2026-06-15",
            "timezone": "Europe/Moscow",
        }
    )

    cards = [event["card"] for event in state["events"] if event["event_type"] == "CARD_CREATED"]
    assert cards[-1]["card_type"] == "task_create_proposal"
    assert cards[-1]["intent_key"] == "task.create"
    assert state["active_card"]["card_type"] == "task_create_proposal"
    assert state["active_tool_call_id"] == "pending-task-create"
    assert state["proposal_gate"]["decision"] == "run_new_keep_pending"
    assert "Сначала завершите" not in state.get("final_response", "")


def test_workflow_question_routes_to_workflow_read_tool() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()
    state = graph.invoke(
        {
            "project_id": "project-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "сколько здесь действий",
            "page_context": {"pageKind": "workflow_editor", "workflowId": "workflow-1"},
        }
    )

    tool_events = [event for event in state["events"] if event["event_type"] == "TOOL_STARTED"]
    assert tool_events[-1]["tool_call"]["tool_key"] == "workflow.read_draft"
    assert state["active_mode"] == "project_read"


def test_workflow_question_with_url_routes_to_workflow_read_tool() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()
    state = graph.invoke(
        {
            "project_id": "project-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "сколько здесь действий https://pulse.riders.agency/workflows/wf-1/edit?action=manual_payment_trigger",
            "page_context": {"pageKind": "workflow_editor", "workflowId": "wf-1"},
        }
    )

    tool_events = [event for event in state["events"] if event["event_type"] == "TOOL_STARTED"]
    assert tool_events[-1]["tool_call"]["tool_key"] == "workflow.read_draft"
    assert not [
        event
        for event in state["events"]
        if event.get("card", {}).get("card_type") == "workflow_patch_preview"
    ]


def test_workflow_patch_routes_to_critical_preview_before_approval() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()
    state = graph.invoke(
        {
            "project_id": "project-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "сделай чтобы после оплаты отправлялся email",
            "page_context": {
                "pageKind": "workflow_editor",
                "workflowId": "wf-1",
                "workflowRevision": 7,
            },
        }
    )

    tool_events = [event["tool_call"] for event in state["events"] if event["event_type"] == "TOOL_STARTED"]
    assert [event["tool_key"] for event in tool_events] == ["workflow.read_draft"]
    proposed = [event["tool_call"] for event in state["events"] if event["event_type"] == "TOOL_PROPOSED"]
    assert proposed[-1]["tool_key"] == "workflow.patch"
    assert proposed[-1]["risk_level"] == "critical"
    cards = [event["card"] for event in state["events"] if event["event_type"] == "CARD_CREATED"]
    card = cards[-1]
    assert card["card_type"] == "workflow_patch_preview"
    assert card["requires_approval"] is True
    assert card["risk_level"] == "critical"
    assert card["input_json"]["workflow_id"] == "wf-1"
    assert card["input_json"]["base_revision"] == 7
    assert state["active_mode"] == "workflow_patch"
    assert state["pending_interrupt"]["active_tool_call_id"] == "pending-workflow-patch"


def test_calculator_turn_does_not_call_project_tools() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()
    state = graph.invoke(
        {
            "project_id": "project-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "1+2",
        }
    )

    assert state["final_response"] == "3"
    assert not [event for event in state["events"] if event["event_type"].startswith("TOOL_")]


def test_client_search_routes_to_client_search_tool_even_with_pending_proposal() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()
    state = graph.invoke(
        {
            "project_id": "project-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "Найди клиента Орион",
            "active_tool_call_id": "pending-task-create",
        }
    )

    tool_events = [event for event in state["events"] if event["event_type"] == "TOOL_STARTED"]
    assert tool_events[-1]["tool_call"]["tool_key"] == "client.search"
    assert state["proposal_gate"]["decision"] == "run_new_keep_pending"


def test_project_explain_routes_to_context_read() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()

    for body in ("Объясни текущий проект", "Проверь AI интеграцию проекта"):
        state = graph.invoke(
            {
                "project_id": "project-1",
                "session_id": "session-1",
                "thread_id": "pulse:project-1:session-1",
                "run_id": "run-1",
                "body": body,
                "page_context": {"route": "/", "pageKind": "dashboard"},
            }
        )
        tool_events = [event for event in state["events"] if event["event_type"] == "TOOL_STARTED"]
        assert tool_events[-1]["tool_call"]["tool_key"] == "context.read"
        assert tool_events[-1]["tool_call"]["input_json"]["entity_type"] == "project"
        assert state["active_mode"] == "project_read"
        assert state["intent"]["intent_key"] == "project.explain"


def test_client_list_routes_to_client_search_tool() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()
    state = graph.invoke(
        {
            "project_id": "project-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "какие у меня клиенты",
        }
    )

    tool_events = [event for event in state["events"] if event["event_type"] == "TOOL_STARTED"]
    assert tool_events[-1]["tool_call"]["tool_key"] == "client.search"
    assert tool_events[-1]["tool_call"]["input_json"]["query"] == ""
    assert state["active_mode"] == "project_read"
    assert state["intent"]["intent_key"] == "client.search"


def test_tasks_and_payments_query_runs_multiple_read_tools() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()
    state = graph.invoke(
        {
            "project_id": "project-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "Покажи что у Орион по задачам и оплатам",
        }
    )

    tool_keys = [
        event["tool_call"]["tool_key"]
        for event in state["events"]
        if event["event_type"] == "TOOL_STARTED"
    ]
    assert tool_keys == ["task.search", "payment.lookup_due"]
    assert state["active_mode"] == "project_read"


def test_selected_task_deictic_move_creates_task_move_proposal() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()
    state = graph.invoke(
        {
            "project_id": "project-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "Запусти эту задачу",
            "page_context": {"pageKind": "task", "taskId": "task-1"},
        }
    )

    cards = [event["card"] for event in state["events"] if event["event_type"] == "CARD_CREATED"]
    assert cards[-1]["card_type"] == "task_move_proposal"
    assert state["active_mode"] == "task_update"


def test_phatic_graph_turn_uses_conversation_mode_without_tools() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()
    state = graph.invoke(
        {
            "project_id": "project-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "ты тут?",
        }
    )

    assert state["active_mode"] == "conversation"
    assert not [event for event in state["events"] if event["event_type"].startswith("TOOL_")]
    assert any(event["event_type"] == "MESSAGE_FINAL" for event in state["events"])


def test_pending_task_priority_and_weekday_edit_uses_same_proposal() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()
    state = graph.invoke(
        {
            "project_id": "project-1",
            "session_id": "session-1",
            "thread_id": "pulse:project-1:session-1",
            "run_id": "run-1",
            "body": "срок поставь на пятницу, приоритет низкий",
            "active_tool_call_id": "pending-task-create",
            "local_date": "2026-06-15",
            "timezone": "Europe/Moscow",
        }
    )

    edits = [
        event["tool_call"]["input_json"]["patch"]
        for event in state["events"]
        if event.get("tool_call", {}).get("tool_key") == "proposal.edit"
    ]
    assert edits == [{"due_date": "2026-06-19", "due_date_label": "пятницу", "priority": "low"}]
    assert state["active_mode"] == "resume_pending_action"


def test_pending_task_natural_language_approval_and_reject_are_approval_mode() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()

    for body, expected_tool in (("да, создавай", "approval.apply"), ("не надо, отмена", "approval.reject")):
        state = graph.invoke(
            {
                "project_id": "project-1",
                "session_id": "session-1",
                "thread_id": "pulse:project-1:session-1",
                "run_id": "run-1",
                "body": body,
                "active_tool_call_id": "pending-task-create",
            }
        )
        tool_events = [event for event in state["events"] if event["event_type"] == "TOOL_STARTED"]
        assert tool_events[-1]["tool_call"]["tool_key"] == expected_tool
        assert state["active_mode"] == "approval"


def test_context_switch_read_detour_and_return_preserve_task_proposal() -> None:
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()
    pending_card = {
        "card_id": "card-task-create",
        "card_type": "task_create_proposal",
        "intent_key": "task.create",
        "risk_level": "high",
        "status": "needs_approval",
        "debug_json": {"input_json": {"title": "Проверить оплату", "client_query": "Орион"}},
    }
    base_state = {
        "project_id": "project-1",
        "session_id": "session-1",
        "thread_id": "pulse:project-1:session-1",
        "run_id": "run-1",
        "active_tool_call_id": "pending-task-create",
        "active_card": pending_card,
        "pending_interrupt": {"active_tool_call_id": "pending-task-create", "card_id": "card-task-create"},
        "dialogue_state": {
            "schema": "riders-pulse-dialogue-state-v1",
            "active_topic": {"topic_id": "task.create:client:client-1", "intent": "task.create"},
            "last_referents": [{"entity_type": "client", "entity_id": "client-1", "label": "Орион"}],
            "resolved_entities": [{"entity_type": "client", "entity_id": "client-1", "label": "Орион"}],
            "pending_action": {"active_tool_call_id": "pending-task-create", "card_id": "card-task-create"},
        },
    }

    read_state = graph.invoke({**base_state, "body": "покажи оплаты по нему"})
    tool_events = [event for event in read_state["events"] if event["event_type"] == "TOOL_STARTED"]
    assert tool_events[-1]["tool_call"]["tool_key"] == "payment.lookup_due"
    assert tool_events[-1]["tool_call"]["input_json"]["client_id"] == "client-1"
    assert read_state["active_tool_call_id"] == "pending-task-create"
    assert read_state["dialogue_state"]["pending_action"]["active_tool_call_id"] == "pending-task-create"

    return_state = graph.invoke({**read_state, "events": [], "body": "вернемся к задаче", "run_id": "run-2"})
    assert return_state["active_mode"] == "resume_pending_action"
    assert return_state["current_step"] == "await_approval"
    assert return_state["active_tool_call_id"] == "pending-task-create"
    assert "Возвращаюсь к текущей карточке" in return_state["final_response"]
