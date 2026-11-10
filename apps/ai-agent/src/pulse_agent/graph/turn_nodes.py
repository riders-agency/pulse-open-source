from __future__ import annotations

from dataclasses import asdict
from typing import Any, cast

from pulse_agent.dialogue_state import reduce_dialogue_state
from pulse_agent.graph.advice import advice_for_turn
from pulse_agent.graph.approval_lifecycle import apply_proposal_edit
from pulse_agent.graph.events import add_event
from pulse_agent.graph.intent_routes import active_mode_for_intent, branch_for_intent, mode_status
from pulse_agent.graph.intent_schema import IntentKey, TurnClassification, classify_turn, proposal_gate_decision
from pulse_agent.graph.planning import plan_turn
from pulse_agent.graph.proposal_edit import resolve_task_proposal_edit
from pulse_agent.graph.task_extraction import parse_task_draft, resolve_due_date
from pulse_agent.grounded_response import build_grounded_response
from pulse_agent.memory_lifecycle import reduce_memory_state
from pulse_agent.observability import build_ai_observability
from pulse_agent.topic_policy import apply_topic_policy


def ingest_turn(state: dict[str, Any]) -> dict[str, Any]:
    body = str(state.get("body") or state.get("current_body") or "")
    messages = list(state.get("messages") or [])
    if body:
        messages.append({"role": "user", "body": body})
    events = list(state.get("events") or [])
    if not events:
        add_event(events, state, "RUN_STARTED", "Запускаю граф AI-сотрудника")
    add_event(events, state, "STATUS", "Шаг: triage")
    return {
        "messages": messages,
        "events": events,
        "current_step": "triage",
        "active_mode": str(state.get("active_mode") or "conversation"),
    }


def recover_active_intent(state: dict[str, Any]) -> dict[str, Any]:
    active = state.get("active_card") or {}
    active_tool_call_id = str(state.get("active_tool_call_id") or "")
    events = list(state.get("events") or [])
    if not active_tool_call_id and active.get("card_type") != "task_create_proposal":
        return {"events": events, "current_step": "hydrate_context"}

    add_event(events, state, "STATUS", "Шаг: hydrate_context")
    return {"events": events, "current_step": "classify_intent"}


def classify_intent(state: dict[str, Any]) -> dict[str, Any]:
    body = str(state.get("body") or state.get("current_body") or "")
    active_tool_call_id = str(state.get("active_tool_call_id") or "")
    events = list(state.get("events") or [])
    page_context = turn_page_context(state)
    classification = classify_turn(
        body,
        active_tool_call_id=active_tool_call_id,
        page_context=page_context,
    )
    plan = plan_for_classification(
        classification,
        body,
        local_date=str(state.get("local_date") or ""),
        timezone=str(state.get("timezone") or ""),
    )
    classification, plan, topic_policy = apply_topic_policy(state, classification, plan)
    add_event(events, state, "STATUS", f"intent_key={classification.intent_key.value}")
    add_event(events, state, "STATUS", mode_status(str(plan.get("intent") or "")))
    add_event(events, state, "STATUS", f"topic_policy={topic_policy.get('decision')}")
    return {
        "events": events,
        "intent": classification.model_dump(mode="json"),
        "plan": plan,
        "topic_policy": topic_policy,
        "final_response": str(plan.get("user_visible_summary") or ""),
        "active_mode": active_mode_for_intent(str(plan.get("intent") or "")),
        "current_step": "classify_intent",
    }


def proposal_gate(state: dict[str, Any]) -> dict[str, Any]:
    active_tool_call_id = str(state.get("active_tool_call_id") or "")
    raw_classification = state.get("intent") if isinstance(state.get("intent"), dict) else {}
    if raw_classification:
        classification = TurnClassification.model_validate(raw_classification)
    else:
        classification = classify_turn(
            str(state.get("body") or state.get("current_body") or ""),
            active_tool_call_id=active_tool_call_id,
            page_context=turn_page_context(state),
        )
    gate = proposal_gate_decision(classification, active_tool_call_id)
    events = list(state.get("events") or [])
    raw_topic_policy = state.get("topic_policy")
    topic_policy: dict[str, Any] = raw_topic_policy if isinstance(raw_topic_policy, dict) else {}
    if topic_policy.get("requires_clarification"):
        add_event(events, state, "STATUS", "proposal_gate=topic_policy_clarification")
        return {
            "events": events,
            "proposal_gate": gate.model_dump(mode="json"),
            "branch": "conversation_reply",
            "active_mode": str(topic_policy.get("active_mode") or "error_recovery"),
            "current_step": str(topic_policy.get("current_step") or "request_clarification"),
            "final_response": str(topic_policy.get("clarification_question") or state.get("final_response") or ""),
        }
    if topic_policy.get("decision") == "return_to_pending_action":
        add_event(events, state, "STATUS", "proposal_gate=return_to_pending_action")
        return {
            "events": events,
            "proposal_gate": gate.model_dump(mode="json"),
            "branch": "conversation_reply",
            "active_mode": "resume_pending_action",
            "current_step": "await_approval",
            "final_response": "Возвращаюсь к текущей карточке. Проверьте и подтвердите действие.",
        }
    add_event(events, state, "STATUS", f"proposal_gate={gate.decision}")
    if gate.decision == "approval_bridge":
        return {
            "events": events,
            "proposal_gate": gate.model_dump(mode="json"),
            "branch": "approval_decision",
            "active_mode": "approval",
            "current_step": "commit_write",
            "plan": approval_plan_for(classification.intent_key),
        }
    if gate.decision == "patch_current":
        active = state.get("active_card") or {}
        current_input = {}
        if isinstance(active, dict):
            raw_debug = active.get("debug_json")
            if isinstance(raw_debug, dict) and isinstance(raw_debug.get("input_json"), dict):
                current_input = raw_debug["input_json"]
        edit = resolve_task_proposal_edit(
            str(state.get("body") or state.get("current_body") or ""),
            current_input,
            local_date=str(state.get("local_date") or ""),
        )
        if edit.changed:
            return {
                "events": events,
                "proposal_gate": gate.model_dump(mode="json"),
                "branch": "proposal_editor",
                "active_mode": "resume_pending_action",
                "current_step": "resume_pending_action",
                "proposal_edit": {"patch": edit.patch, "human_summary": edit.human_summary},
            }
    if gate.decision == "conflict_choice":
        return {
            "events": events,
            "proposal_gate": gate.model_dump(mode="json"),
            "branch": "conflict_choice",
            "active_mode": "resume_pending_action",
            "current_step": "request_clarification",
        }
    return {
        "events": events,
        "proposal_gate": gate.model_dump(mode="json"),
        "branch": branch_for_intent(str((state.get("plan") or {}).get("intent") or "question.answer")),
    }


def proposal_editor(state: dict[str, Any]) -> dict[str, Any]:
    body = str(state.get("body") or state.get("current_body") or "")
    edit = state.get("proposal_edit") or {}
    patch = edit.get("patch")
    if not isinstance(patch, dict):
        patch = {}
    human_summary = str(edit.get("human_summary") or "Обновил текущую карточку.")
    raw_active_card = state.get("active_card")
    active_card: dict[str, Any] = raw_active_card if isinstance(raw_active_card, dict) else {}
    base_revision = int(active_card.get("revision") or 0) if isinstance(active_card, dict) else 0
    card, pending = apply_proposal_edit(state, patch)
    events = list(state.get("events") or [])
    add_event(events, state, "STATUS", "Шаг: build_draft")
    add_event(
        events,
        state,
        "TOOL_STARTED",
        "Обновляю текущий proposal",
        tool_call={
            "tool_key": "proposal.edit",
            "risk_level": "medium",
            "status": "running",
            "input_json": {
                "body": body,
                "active_tool_call_id": str(state.get("active_tool_call_id") or ""),
                "active_proposal_id": str(state.get("active_proposal_id") or ""),
                "card_id": str(active_card.get("card_id") or "") if isinstance(active_card, dict) else "",
                "base_revision": base_revision,
                "patch": patch,
            },
        },
    )
    if card and pending:
        add_event(events, state, "CARD_PATCHED", human_summary, card=card)
        add_event(events, state, "APPROVAL_REQUIRED", "Жду подтверждения", interrupt=pending)
    updated_card = card if card else active_card
    return {
        "events": events,
        "final_response": human_summary,
        "active_card": updated_card,
        "active_card_id": str(updated_card.get("card_id") or ""),
        "active_tool_call_id": str(pending.get("active_tool_call_id") or state.get("active_tool_call_id") or ""),
        "pending_interrupt": pending or state.get("pending_interrupt"),
        "active_mode": "resume_pending_action",
        "current_step": "await_approval",
    }


def turn_page_context(state: dict[str, Any]) -> dict[str, Any]:
    raw = state.get("page_context")
    page_context = dict(raw) if isinstance(raw, dict) else {}
    for key, state_key in (
        ("company_id", "company_id"),
        ("project_id", "project_id"),
        ("projectId", "project_id"),
        ("user_id", "user_id"),
        ("session_id", "session_id"),
    ):
        if state.get(state_key):
            page_context[key] = str(state.get(state_key) or "")
    return page_context


def conflict_choice(state: dict[str, Any]) -> dict[str, Any]:
    events = list(state.get("events") or [])
    raw_gate = state.get("proposal_gate")
    gate = cast(dict[str, Any], raw_gate) if isinstance(raw_gate, dict) else {}
    add_event(
        events,
        state,
        "CARD_CREATED",
        "Нужно выбрать действие",
        card={
            "schema": "riders-pulse-ai-card-v1",
            "card_id": "card-conflict-choice",
            "card_type": "conflict_choice",
            "revision": 1,
            "status": "needs_approval",
            "risk_level": "high",
            "requires_approval": True,
            "title": "Есть незавершенная карточка",
            "subtitle": "Новая write-команда не будет смешана с текущим pending action.",
            "fields": [
                {"key": "active_tool_call_id", "label": "Активный tool call", "value": gate.get("active_tool_call_id")},
                {"key": "new_intent", "label": "Новая команда", "value": gate.get("target_intent_key")},
            ],
            "actions": [
                {"action_key": "keep_pending", "label": "Оставить текущую", "style": "secondary"},
                {"action_key": "cancel_pending", "label": "Отменить текущую", "style": "danger"},
            ],
            "debug_json": {"hidden_by_default": True, "proposal_gate": gate},
        },
    )
    return {
        "events": events,
        "final_response": "Сначала завершите или отмените текущую карточку, затем я создам новую.",
        "active_mode": "resume_pending_action",
        "current_step": "request_clarification",
    }


def conversation_reply(state: dict[str, Any]) -> dict[str, Any]:
    plan = state.get("plan") or {}
    raw_topic_policy = state.get("topic_policy")
    topic_policy: dict[str, Any] = raw_topic_policy if isinstance(raw_topic_policy, dict) else {}
    final_response = str(state.get("final_response") or plan.get("user_visible_summary") or "На связи.")
    if topic_policy.get("decision") == "return_to_pending_action":
        return {
            "final_response": final_response,
            "active_mode": "resume_pending_action",
            "current_step": "await_approval",
        }
    advice = advice_for_turn(state)
    if advice is not None:
        return {**advice, "active_mode": "conversation", "current_step": "summarize_result"}
    if topic_policy.get("requires_clarification"):
        return {
            "final_response": final_response,
            "active_mode": str(state.get("active_mode") or "error_recovery"),
            "current_step": "request_clarification",
        }
    return {
        "final_response": final_response,
        "active_mode": "conversation",
        "current_step": "summarize_result",
    }


def calculator_reply(state: dict[str, Any]) -> dict[str, Any]:
    body = str(state.get("body") or state.get("current_body") or "")
    result = calculate_expression(body)
    final_response = result if result else "Не удалось вычислить выражение."
    events = list(state.get("events") or [])
    add_event(
        events,
        state,
        "CARD_CREATED",
        "Результат вычисления",
        card={
            "schema": "riders-pulse-ai-card-v1",
            "card_id": "card-calculator-result",
            "card_type": "calculator_result",
            "revision": 1,
            "status": "succeeded",
            "risk_level": "read",
            "requires_approval": False,
            "title": "Калькулятор",
            "fields": [
                {"key": "expression", "label": "Выражение", "value": body},
                {"key": "result", "label": "Результат", "value": final_response},
            ],
            "display_json": {"expression": body, "result": final_response},
            "debug_json": {"hidden_by_default": True},
        },
    )
    return {
        "events": events,
        "final_response": final_response,
        "active_mode": "conversation",
        "current_step": "summarize_result",
    }


def finalizer(state: dict[str, Any]) -> dict[str, Any]:
    events = list(state.get("events") or [])
    fallback_response = str(state.get("final_response") or "")
    grounded_response = build_grounded_response(state, fallback=fallback_response or "Готово.")
    final = str(grounded_response.get("final_response") or fallback_response or "Готово.")
    next_state = {**state, "final_response": final, "message_payload_json": grounded_response}
    if final and not state.get("awaiting_tool_observation"):
        add_event(events, next_state, "MESSAGE_FINAL", final, tool_call={"grounded_response": grounded_response})
    ai_observability = build_ai_observability(next_state, events, response_payload=grounded_response)
    next_state = {**next_state, "ai_observability_json": ai_observability}
    add_event(events, next_state, "DONE", "DONE", tool_call={"ai_observability": ai_observability})
    memory_update = reduce_memory_state(next_state, final_response=final)
    dialogue_update = reduce_dialogue_state({**next_state, **memory_update}, final_response=final)
    return {
        "events": events,
        "final_response": final,
        "message_payload_json": grounded_response,
        "ai_observability_json": ai_observability,
        **memory_update,
        **dialogue_update,
    }


def route_branch(state: dict[str, Any]) -> str:
    branch = str(state.get("branch") or "")
    if branch:
        return branch
    plan = state.get("plan") or {}
    return branch_for_intent(str(plan.get("intent") or "question.answer"))


def plan_for_classification(
    classification: TurnClassification,
    body: str,
    *,
    local_date: str,
    timezone: str,
) -> dict[str, Any]:
    key = classification.intent_key
    if key == IntentKey.CONVERSATION_SMALLTALK:
        return {
            "intent": key.value,
            "confidence": classification.confidence,
            "required_tools": [],
            "risk_level": "read",
            "requires_approval": False,
            "user_visible_summary": (
                "На связи. Могу читать проектный контекст, готовить карточки и выполнять изменения после подтверждения."
            ),
            "tool_input_preview": {},
        }
    if key == IntentKey.UTILITY_CALCULATOR:
        return {
            "intent": key.value,
            "confidence": classification.confidence,
            "required_tools": [],
            "risk_level": "read",
            "requires_approval": False,
            "user_visible_summary": "Вычислить выражение локально, без project tools.",
            "tool_input_preview": {"expression": body},
        }
    if key in {IntentKey.PAGE_WHERE_AM_I, IntentKey.PAGE_VERIFY_CONTEXT}:
        return {
            "intent": key.value,
            "confidence": classification.confidence,
            "required_tools": ["ui_context.read"],
            "risk_level": "read",
            "requires_approval": False,
            "user_visible_summary": "Проверить текущий UI context.",
            "tool_input_preview": {},
        }
    if key == IntentKey.PROJECT_EXPLAIN and classification.reason == "task_payment_read":
        return {
            "intent": key.value,
            "confidence": classification.confidence,
            "required_tools": [tool.tool_key for tool in classification.tool_candidates],
            "risk_level": "read",
            "requires_approval": False,
            "user_visible_summary": "Прочитать задачи и оплаты через project-scoped tools.",
            "tool_input_preview": {"query": body, "entity_type": "project"},
        }
    if key == IntentKey.PROJECT_EXPLAIN:
        return {
            "intent": key.value,
            "confidence": classification.confidence,
            "required_tools": ["context.read"],
            "risk_level": "read",
            "requires_approval": False,
            "user_visible_summary": "Прочитать текущий проектный контекст.",
            "tool_input_preview": {"query": body, "entity_type": "project"},
        }
    if key == IntentKey.TASK_SEARCH:
        return {
            "intent": key.value,
            "confidence": classification.confidence,
            "required_tools": ["task.search"],
            "risk_level": "read",
            "requires_approval": False,
            "user_visible_summary": "Найти задачи через project-scoped tools.",
            "tool_input_preview": {"query": body, "entity_type": "task"},
        }
    if key == IntentKey.CLIENT_SEARCH:
        query = (
            ""
            if classification.reason == "client_list"
            else first_entity_query(classification, "client", default=body)
        )
        return {
            "intent": key.value,
            "confidence": classification.confidence,
            "required_tools": ["client.search"],
            "risk_level": "read",
            "requires_approval": False,
            "user_visible_summary": "Прочитать клиентов проекта через project-scoped tools.",
            "tool_input_preview": {"query": query, "entity_type": "client"},
        }
    if key == IntentKey.CLIENT_COUNT:
        return {
            "intent": key.value,
            "confidence": classification.confidence,
            "required_tools": ["client.count"],
            "risk_level": "read",
            "requires_approval": False,
            "user_visible_summary": "Посчитать клиентов проекта через read tool.",
            "tool_input_preview": {},
        }
    if key == IntentKey.PAYMENT_LOOKUP_DUE:
        if classification.reason == "payment_project_lookup_due":
            return {
                "intent": key.value,
                "confidence": classification.confidence,
                "required_tools": ["payment.lookup_due"],
                "risk_level": "read",
                "requires_approval": False,
                "user_visible_summary": "Найти ближайшие оплаты проекта.",
                "tool_input_preview": {"query": "", "scope": "project", "limit": 8},
            }
        query = first_entity_query(classification, "client", default=body)
        return {
            "intent": key.value,
            "confidence": classification.confidence,
            "required_tools": ["client.search", "payment.lookup_due"],
            "risk_level": "read",
            "requires_approval": False,
            "user_visible_summary": "Найти ближайшую оплату клиента.",
            "tool_input_preview": {"query": query},
        }
    if key == IntentKey.WORKFLOW_ANSWER_QUESTION:
        return {
            "intent": key.value,
            "confidence": classification.confidence,
            "required_tools": ["workflow.read_draft"],
            "risk_level": "read",
            "requires_approval": False,
            "user_visible_summary": "Прочитать workflow draft и ответить по действиям.",
            "tool_input_preview": {"query": body},
        }
    if key == IntentKey.CLIENT_CREATE:
        return asdict(plan_turn(body, local_date=local_date, timezone=timezone))
    if key == IntentKey.TASK_CREATE:
        # Classification has already selected the allowed intent. Re-running a
        # keyword router here loses model-recognized synonyms and changes the route.
        draft = parse_task_draft(body, local_date=local_date, timezone=timezone)
        return {
            "intent": key.value,
            "confidence": classification.confidence,
            "required_tools": ["client.search", "task.create"],
            "risk_level": "high",
            "requires_approval": True,
            "user_visible_summary": "Создать задачу после подтверждения",
            "tool_input_preview": draft.to_tool_input(),
        }
    if key == IntentKey.TASK_MOVE:
        selected_task_id = ""
        for entity in classification.entities:
            if entity.entity_type == "task" and entity.entity_id:
                selected_task_id = entity.entity_id
                break
        due, due_label = resolve_due_date(body, local_date)
        return {
            "intent": key.value,
            "confidence": classification.confidence,
            "required_tools": ["task.move"],
            "risk_level": "high",
            "requires_approval": True,
            "user_visible_summary": "Подготовить изменение выбранной задачи после подтверждения",
            "tool_input_preview": {
                "task_id": selected_task_id,
                "due_date": due,
                "due_date_label": due_label,
                "comment": body,
                "human_action_label": "Изменить выбранную задачу",
            },
        }
    if key == IntentKey.MULTI_STEP_PLAN:
        return asdict(plan_turn(body, local_date=local_date, timezone=timezone))
    if key == IntentKey.WORKFLOW_PATCH:
        return {
            "intent": key.value,
            "confidence": classification.confidence,
            "required_tools": ["workflow.patch"],
            "risk_level": "critical",
            "requires_approval": True,
            "user_visible_summary": "Подготовить workflow patch preview.",
            "tool_input_preview": {"query": body},
        }
    if key == IntentKey.UNKNOWN:
        return {
            "intent": key.value,
            "confidence": classification.confidence,
            "required_tools": [],
            "risk_level": "read",
            "requires_approval": False,
            "user_visible_summary": classification.clarification_question,
            "tool_input_preview": {},
        }
    return {
        "intent": key.value,
        "confidence": classification.confidence,
        "required_tools": [tool.tool_key for tool in classification.tool_candidates],
        "risk_level": risk_level_for_classification(classification),
        "requires_approval": requires_approval_for_classification(classification),
        "user_visible_summary": "Ответить по текущему контексту.",
        "tool_input_preview": {
            "query": first_entity_query(classification, default=body),
            "entity_type": classification.target_entity_type,
        },
    }


def first_entity_query(classification: TurnClassification, *entity_types: str, default: str = "") -> str:
    allowed = set(entity_types)
    for entity in classification.entities:
        if allowed and entity.entity_type not in allowed:
            continue
        if entity.query:
            return entity.query
    return default


def risk_level_for_classification(classification: TurnClassification) -> str:
    rank = {"read": 0, "medium": 1, "high": 2, "critical": 3, "unsafe": 4}
    risk = "read"
    for tool in classification.tool_candidates:
        if rank.get(tool.risk_level, 0) > rank.get(risk, 0):
            risk = tool.risk_level
    return risk


def requires_approval_for_classification(classification: TurnClassification) -> bool:
    return classification.policy.idempotency_required or any(
        tool.requires_approval for tool in classification.tool_candidates
    )


def calculate_expression(value: str) -> str:
    import ast
    import operator

    allowed_binary = {
        ast.Add: operator.add,
        ast.Sub: operator.sub,
        ast.Mult: operator.mul,
        ast.Div: operator.truediv,
    }

    def eval_node(node: ast.AST) -> float:
        if isinstance(node, ast.Expression):
            return eval_node(node.body)
        if isinstance(node, ast.Constant) and isinstance(node.value, int | float):
            return float(node.value)
        if isinstance(node, ast.UnaryOp) and isinstance(node.op, ast.USub):
            return -eval_node(node.operand)
        if isinstance(node, ast.BinOp) and type(node.op) in allowed_binary:
            return float(allowed_binary[type(node.op)](eval_node(node.left), eval_node(node.right)))
        raise ValueError("unsupported expression")

    expression = value.strip().replace("х", "*").replace("x", "*").rstrip("=")
    try:
        parsed = ast.parse(expression, mode="eval")
        result = eval_node(parsed)
    except Exception:
        return ""
    if result.is_integer():
        return str(int(result))
    return str(result)


def approval_plan_for(intent_key: IntentKey) -> dict[str, Any]:
    if intent_key == IntentKey.APPROVAL_REJECT:
        return {
            "intent": "approval.reject",
            "confidence": 0.99,
            "required_tools": ["approval.reject"],
            "risk_level": "high",
            "requires_approval": False,
            "user_visible_summary": "Отклонить активное действие",
        }
    return {
        "intent": "approval.apply",
        "confidence": 0.99,
        "required_tools": ["approval.apply"],
        "risk_level": "high",
        "requires_approval": False,
        "user_visible_summary": "Применить активное подтвержденное действие",
    }
