from __future__ import annotations

from typing import Any

from pulse_agent.execution_trace import traced_node
from pulse_agent.graph.approval_lifecycle import (
    approval_decision,
    await_approval,
    pending_approval_for_card,
    route_after_resume,
)
from pulse_agent.graph.events import add_event
from pulse_agent.graph.intent_routes import validate_intent_routes
from pulse_agent.graph.observations import (
    append_tool_observation,
    normalize_tool_observation,
    record_tool_result_event,
    replan_trace_entry,
)
from pulse_agent.graph.read_loop import read_context, workflow_question
from pulse_agent.graph.subgraphs.workflow_agent import workflow_patch_workflow
from pulse_agent.graph.tool_gateway import run_read_tool_if_enabled
from pulse_agent.graph.turn_nodes import (
    calculator_reply,
    classify_intent,
    conflict_choice,
    conversation_reply,
    finalizer,
    ingest_turn,
    proposal_editor,
    proposal_gate,
    recover_active_intent,
    route_branch,
)
from pulse_agent.schemas.cards import client_create_card, task_create_card


def build_graph() -> Any:
    validate_intent_routes()
    try:
        from langgraph.graph import END, START, StateGraph

        from pulse_agent.graph.state import AgentState
    except Exception:
        return None

    builder = StateGraph(AgentState)
    builder.add_node("ingest_turn", traced_node("ingest_turn", ingest_turn))
    builder.add_node("recover_active_intent", traced_node("recover_active_intent", recover_active_intent))
    builder.add_node("classify_intent", traced_node("classify_intent", classify_intent))
    builder.add_node("proposal_gate", traced_node("proposal_gate", proposal_gate))
    builder.add_node("proposal_editor", traced_node("proposal_editor", proposal_editor))
    builder.add_node("approval_decision", traced_node("approval_decision", approval_decision))
    builder.add_node("conversation_reply", traced_node("conversation_reply", conversation_reply))
    builder.add_node("calculator_reply", traced_node("calculator_reply", calculator_reply))
    builder.add_node("conflict_choice", traced_node("conflict_choice", conflict_choice))
    builder.add_node("read_context", traced_node("read_context", read_context))
    builder.add_node("workflow_question", traced_node("workflow_question", workflow_question))
    builder.add_node("workflow_patch_workflow", traced_node("workflow_patch_workflow", workflow_patch_workflow))
    builder.add_node("task_create_workflow", traced_node("task_create_workflow", task_create_workflow))
    builder.add_node("task_move_workflow", traced_node("task_move_workflow", task_move_workflow))
    builder.add_node("client_create_workflow", traced_node("client_create_workflow", client_create_workflow))
    builder.add_node("compound_workflow", traced_node("compound_workflow", compound_workflow))
    builder.add_node("await_approval", traced_node("await_approval", await_approval))
    builder.add_node("finalizer", traced_node("finalizer", finalizer))

    builder.add_edge(START, "ingest_turn")
    builder.add_edge("ingest_turn", "recover_active_intent")
    builder.add_edge("recover_active_intent", "classify_intent")
    builder.add_edge("classify_intent", "proposal_gate")
    builder.add_conditional_edges(
        "proposal_gate",
        route_branch,
        {
            "proposal_editor": "proposal_editor",
            "approval_decision": "approval_decision",
            "conversation_reply": "conversation_reply",
            "calculator_reply": "calculator_reply",
            "conflict_choice": "conflict_choice",
            "read_context": "read_context",
            "workflow_question": "workflow_question",
            "workflow_patch_workflow": "workflow_patch_workflow",
            "task_create_workflow": "task_create_workflow",
            "task_move_workflow": "task_move_workflow",
            "client_create_workflow": "client_create_workflow",
            "compound_workflow": "compound_workflow",
        },
    )
    for node in (
        "approval_decision",
        "conversation_reply",
        "calculator_reply",
        "conflict_choice",
        "read_context",
        "workflow_question",
    ):
        builder.add_edge(node, "finalizer")
    for node in (
        "proposal_editor",
        "task_create_workflow",
        "task_move_workflow",
        "client_create_workflow",
        "compound_workflow",
        "workflow_patch_workflow",
    ):
        builder.add_edge(node, "await_approval")
    builder.add_conditional_edges(
        "await_approval",
        route_after_resume,
        {
            "approval_decision": "approval_decision",
            "proposal_editor": "proposal_editor",
            "conversation_reply": "conversation_reply",
        },
    )
    builder.add_edge("finalizer", END)
    return builder


def task_create_workflow(state: dict[str, Any]) -> dict[str, Any]:
    plan = state.get("plan") or {}
    draft = plan.get("tool_input_preview")
    if not isinstance(draft, dict):
        draft = {}
    events = list(state.get("events") or [])
    add_event(events, state, "STATUS", "Шаг: build_draft")
    turn_observations: list[dict[str, Any]] = []
    tool_observations = [item for item in list(state.get("tool_observations") or []) if isinstance(item, dict)]
    replan_trace = [item for item in list(state.get("replan_trace") or []) if isinstance(item, dict)]
    client_query = str(draft.get("client_query") or "").strip()
    if client_query:
        client_input = {"query": client_query, "intent": "task.create"}
        add_event(
            events,
            state,
            "TOOL_PROPOSED",
            "Запланировать client.search",
            tool_call={
                "tool_key": "client.search",
                "risk_level": "read",
                "status": "proposed",
                "input_json": client_input,
            },
        )
        add_event(
            events,
            state,
            "TOOL_STARTED",
            "Ищу клиента через Go Tool Gateway",
            tool_call={
                "tool_key": "client.search",
                "risk_level": "read",
                "status": "running",
                "input_json": client_input,
            },
        )
        observation = run_read_tool_if_enabled(state, "client.search", client_input)
        if observation is not None:
            normalized = normalize_tool_observation(
                "client.search",
                observation,
                client_input,
                step_index=1,
                run_id=str(state.get("run_id") or ""),
                turn_id=str(state.get("turn_id") or ""),
            )
            turn_observations = append_tool_observation(turn_observations, normalized)
            tool_observations = append_tool_observation(tool_observations, normalized)
            replan_trace.append(replan_trace_entry("client.search", turn_observations, "task.create"))
            record_tool_result_event(events, state, normalized, "Client search выполнен")
            draft = resolve_task_client_from_observation(draft, normalized)
            plan = {**plan, "tool_input_preview": draft}
    card = task_create_card("card-task-create-draft", draft)
    card_json = card.model_dump(mode="json", by_alias=True)
    if turn_observations:
        debug_json = card_json.setdefault("debug_json", {})
        if isinstance(debug_json, dict):
            debug_json["tool_observations"] = turn_observations
    pending = pending_approval_for_card(card_json)
    add_event(
        events,
        state,
        "TOOL_PROPOSED",
        "Подготовить task.create proposal",
        tool_call={"tool_key": "task.create", "risk_level": "high", "status": "proposed", "input_json": draft},
    )
    add_event(
        events,
        state,
        "CARD_CREATED",
        "Подготовил карточку задачи",
        card=card_json,
    )
    add_event(
        events,
        state,
        "APPROVAL_REQUIRED",
        "Жду подтверждения",
        interrupt=pending,
    )
    return {
        "events": events,
        "active_card": card_json,
        "active_card_id": str(card_json.get("card_id") or ""),
        "active_tool_call_id": str(pending.get("active_tool_call_id") or ""),
        "pending_interrupt": pending,
        "active_mode": "task_create",
        "current_step": "await_approval",
        "plan": plan,
        "tool_observations": tool_observations,
        "replan_trace": replan_trace,
    }


def resolve_task_client_from_observation(draft: dict[str, Any], observation: dict[str, Any]) -> dict[str, Any]:
    if not isinstance(draft, dict) or str(draft.get("client_id") or "").strip():
        return draft
    client = selected_client_from_observation(observation)
    client_id = str(client.get("client_id") or client.get("entity_id") or "").strip()
    if not client_id:
        return draft
    client_name = str(
        client.get("name") or client.get("label") or client.get("query") or draft.get("client_query") or ""
    ).strip()
    resolved = dict(draft)
    resolved["client_id"] = client_id
    resolved["client_name"] = client_name
    resolved["client_resolved"] = True
    missing_fields = []
    for item in list(resolved.get("missing_fields") or []):
        if not isinstance(item, dict) or item.get("field") != "client_id":
            missing_fields.append(item)
    resolved["missing_fields"] = missing_fields
    return resolved


def selected_client_from_observation(observation: dict[str, Any]) -> dict[str, Any]:
    output = observation.get("normalized_output")
    if isinstance(output, dict):
        selected = output.get("selected_client")
        if isinstance(selected, dict) and (selected.get("client_id") or selected.get("entity_id")):
            return selected
        selected_id = str(output.get("selected_client_id") or "").strip()
        if selected_id:
            return {"client_id": selected_id, "name": str(output.get("selected_client_name") or "")}
    refs = observation.get("entity_refs")
    if isinstance(refs, list):
        for ref in refs:
            if isinstance(ref, dict) and ref.get("entity_type") == "client" and ref.get("entity_id"):
                return ref
    return {}


def task_move_workflow(state: dict[str, Any]) -> dict[str, Any]:
    plan = state.get("plan") or {}
    draft = plan.get("tool_input_preview")
    if not isinstance(draft, dict):
        draft = {}
    events = list(state.get("events") or [])
    card = {
        "schema": "riders-pulse-ai-card-v1",
        "card_id": "card-task-move-draft",
        "card_type": "task_move_proposal",
        "revision": 1,
        "status": "needs_approval",
        "risk_level": "high",
        "requires_approval": True,
        "tool_call_id": "pending-task-move",
        "intent_key": "task.move",
        "title": "Изменение выбранной задачи",
        "subtitle": "Проверьте изменение перед применением",
        "fields": [
            {"key": "task_id", "label": "Задача", "value": draft.get("task_id") or "", "value_type": "text"},
            {
                "key": "due_date",
                "label": "Срок",
                "value": draft.get("due_date_label") or draft.get("due_date") or "",
                "value_type": "date",
            },
            {"key": "comment", "label": "Комментарий", "value": draft.get("comment") or "", "value_type": "text"},
        ],
        "actions": [
            {"action_key": "approve", "label": "Применить", "style": "primary"},
            {"action_key": "reject", "label": "Отклонить", "style": "secondary"},
        ],
        "display_json": draft,
        "input_json": draft,
        "debug_json": {"hidden_by_default": True, "input_json": draft},
    }
    pending = pending_approval_for_card(card)
    add_event(events, state, "STATUS", "Шаг: build_draft")
    add_event(
        events,
        state,
        "TOOL_PROPOSED",
        "Подготовить task.move proposal",
        tool_call={"tool_key": "task.move", "risk_level": "high", "status": "proposed", "input_json": draft},
    )
    add_event(events, state, "CARD_CREATED", "Подготовил карточку изменения задачи", card=card)
    add_event(
        events,
        state,
        "APPROVAL_REQUIRED",
        "Жду подтверждения",
        interrupt=pending,
    )
    return {
        "events": events,
        "active_card": card,
        "active_card_id": str(card.get("card_id") or ""),
        "active_tool_call_id": str(pending.get("active_tool_call_id") or ""),
        "pending_interrupt": pending,
        "active_mode": "task_update",
        "current_step": "await_approval",
    }


def client_create_workflow(state: dict[str, Any]) -> dict[str, Any]:
    plan = state.get("plan") or {}
    draft = plan.get("tool_input_preview")
    if not isinstance(draft, dict):
        draft = {}
    events = list(state.get("events") or [])
    card = client_create_card("card-client-create-draft", draft)
    card_json = card.model_dump(mode="json", by_alias=True)
    pending = pending_approval_for_card(card_json)
    add_event(events, state, "STATUS", "Шаг: build_draft")
    add_event(
        events,
        state,
        "TOOL_PROPOSED",
        "Подготовить client.create proposal",
        tool_call={"tool_key": "client.create", "risk_level": "high", "status": "proposed", "input_json": draft},
    )
    add_event(
        events,
        state,
        "CARD_CREATED",
        "Подготовил карточку клиента",
        card=card_json,
    )
    add_event(
        events,
        state,
        "APPROVAL_REQUIRED",
        "Жду подтверждения",
        interrupt=pending,
    )
    return {
        "events": events,
        "active_card": card_json,
        "active_card_id": str(card_json.get("card_id") or ""),
        "active_tool_call_id": str(pending.get("active_tool_call_id") or ""),
        "pending_interrupt": pending,
        "active_mode": "client_create",
        "current_step": "await_approval",
    }


def compound_workflow(state: dict[str, Any]) -> dict[str, Any]:
    plan = state.get("plan") or {}
    preview = plan.get("tool_input_preview")
    if not isinstance(preview, dict):
        preview = {}
    task_preview = preview.get("task")
    task_input = task_preview if isinstance(task_preview, dict) else {}
    card = task_create_card("card-compound-task-create-draft", task_input)
    compound = card.model_dump(mode="json", by_alias=True)
    compound["card_type"] = "multi_step_plan"
    compound["intent_key"] = "multi_step.plan"
    compound["title"] = "Создание клиента и задачи"
    compound["subtitle"] = "Проверьте клиента, задачу и запуск после создания"
    compound["debug_json"]["compound_input_json"] = preview
    pending = pending_approval_for_card(compound)
    events = list(state.get("events") or [])
    add_event(events, state, "STATUS", "Шаг: build_draft")
    add_event(
        events,
        state,
        "TOOL_PROPOSED",
        "Подготовить multi-step write plan",
        tool_call={"tool_key": "multi_step.plan", "risk_level": "high", "status": "proposed", "input_json": preview},
    )
    add_event(events, state, "CARD_CREATED", "Подготовил создание клиента и задачи", card=compound)
    add_event(
        events,
        state,
        "APPROVAL_REQUIRED",
        "Жду подтверждения составного действия",
        interrupt=pending,
    )
    return {
        "events": events,
        "active_card": compound,
        "active_card_id": str(compound.get("card_id") or ""),
        "active_tool_call_id": str(pending.get("active_tool_call_id") or ""),
        "pending_interrupt": pending,
        "active_mode": "client_create",
        "current_step": "await_approval",
    }
