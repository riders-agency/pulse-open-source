from __future__ import annotations

import hashlib
import json
from typing import Any, cast

from pulse_agent.graph.events import add_event
from pulse_agent.graph.observations import (
    append_tool_observation,
    normalize_tool_observation,
    record_tool_result_event,
    summarize_observations,
)
from pulse_agent.graph.tool_gateway import execute_approval_if_enabled

ALLOWED_EDIT_FIELDS = {
    "checklist",
    "client_id",
    "client_name",
    "client_query",
    "comment",
    "description",
    "due_date",
    "due_date_label",
    "labels",
    "notification_intent",
    "priority",
    "start_timer_after_create",
    "task_id",
    "title",
}


def approval_decision(state: dict[str, Any]) -> dict[str, Any]:
    plan = state.get("plan") or {}
    intent = str(plan.get("intent") or "approval.apply")
    raw_decision = state.get("approval_decision")
    decision = cast(dict[str, Any], raw_decision) if isinstance(raw_decision, dict) else {}
    action = str(decision.get("action") or "").lower()
    if action in {"approve", "apply"}:
        action = "apply"
    elif action in {"reject", "cancel"}:
        action = "reject"
    else:
        action = "reject" if intent == "approval.reject" else "apply"
    tool_key = "approval.reject" if action == "reject" else "approval.apply"
    events = list(state.get("events") or [])
    add_event(events, state, "STATUS", "Шаг: commit_write")
    add_event(
        events,
        state,
        "TOOL_STARTED",
        "Передаю решение approval в Go control plane",
        tool_call={
            "tool_key": tool_key,
            "risk_level": "high",
            "status": "running",
            "input_json": {
                "active_tool_call_id": str(
                    decision.get("active_tool_call_id") or state.get("active_tool_call_id") or ""
                ),
                "card_id": str(decision.get("card_id") or ""),
                "card_revision": int_value(decision.get("card_revision")),
                "decision": action,
                "resume_request_id": str(decision.get("request_id") or ""),
                "resume_idempotency_key": str(decision.get("idempotency_key") or ""),
            },
        },
    )
    observation = execute_approval_if_enabled(state, decision, approve=tool_key == "approval.apply")
    if observation is not None:
        normalized = normalize_tool_observation(
            tool_key,
            observation,
            {
                "active_tool_call_id": str(
                    decision.get("active_tool_call_id") or state.get("active_tool_call_id") or ""
                ),
                "decision": action,
            },
            step_index=1,
            run_id=str(state.get("run_id") or ""),
            turn_id=str(state.get("turn_id") or ""),
        )
        tool_observations = append_tool_observation(list(state.get("tool_observations") or []), normalized)
        record_tool_result_event(events, state, normalized, "Approval processed", risk_level="high")
        return {
            "events": events,
            "active_mode": "approval",
            "current_step": "summarize_result",
            "awaiting_tool_observation": False,
            "tool_results": [observation],
            "tool_observations": tool_observations,
            "final_response": summarize_observations([normalized], "Действие обработано."),
        }
    return {
        "events": events,
        "active_mode": "approval",
        "current_step": "commit_write",
        "awaiting_tool_observation": True,
    }


def await_approval(state: dict[str, Any]) -> dict[str, Any]:
    try:
        from langgraph.types import interrupt
    except Exception:
        events = list(state.get("events") or [])
        add_event(
            events,
            state,
            "ERROR",
            "LangGraph interrupt runtime is not available",
            error={"code": "langgraph_interrupt_unavailable"},
        )
        return {
            "events": events,
            "current_step": "failed",
            "active_mode": "error_recovery",
            "final_response": "Не удалось поставить действие на подтверждение.",
            "branch": "conversation_reply",
        }

    pending = pending_approval_from_state(state)
    resume_value = interrupt(pending)
    decision = normalize_resume_decision(resume_value)
    action = str(decision.get("action") or "")
    events = list(state.get("events") or [])
    add_event(events, state, "STATUS", f"approval_resume={action}")
    if action == "edit":
        patch = proposal_patch_from_resume(decision, state)
        return {
            "events": events,
            "approval_decision": decision,
            "proposal_edit": {
                "patch": patch,
                "human_summary": "Обновил карточку по вашему решению.",
            },
            "active_mode": "resume_pending_action",
            "current_step": "resume_pending_action",
            "pending_interrupt": None,
            "branch": "proposal_editor",
            **resume_observability_state(decision),
        }
    if action in {"approve", "reject", "cancel"}:
        intent = "approval.apply" if action == "approve" else "approval.reject"
        return {
            "events": events,
            "approval_decision": decision,
            "plan": {
                "intent": intent,
                "confidence": 1.0,
                "required_tools": [intent],
                "risk_level": "high",
                "requires_approval": False,
                "user_visible_summary": "Продолжить подтвержденное действие",
            },
            "active_tool_call_id": str(decision.get("active_tool_call_id") or state.get("active_tool_call_id") or ""),
            "active_mode": "approval",
            "current_step": "commit_write",
            "pending_interrupt": None,
            "branch": "approval_decision",
            **resume_observability_state(decision),
        }
    return {
        "events": events,
        "approval_decision": decision,
        "active_mode": "resume_pending_action",
        "current_step": "await_approval",
        "final_response": "Не понял решение по карточке. Подтвердите, отклоните или отредактируйте действие.",
        "branch": "conversation_reply",
        **resume_observability_state(decision),
    }


def route_after_resume(state: dict[str, Any]) -> str:
    branch = str(state.get("branch") or "")
    if branch in {"approval_decision", "proposal_editor", "conversation_reply"}:
        return branch
    return "conversation_reply"


def pending_approval_from_state(state: dict[str, Any]) -> dict[str, Any]:
    pending = state.get("pending_interrupt")
    if isinstance(pending, dict) and pending:
        return pending
    active_card = state.get("active_card")
    if isinstance(active_card, dict):
        return pending_approval_for_card(active_card)
    return {
        "interrupt_type": "tool_approval",
        "interrupt_id": str(state.get("active_tool_call_id") or ""),
        "active_tool_call_id": str(state.get("active_tool_call_id") or ""),
        "card_id": str(state.get("active_card_id") or ""),
        "card_revision": 0,
        "risk_level": "high",
        "allowed_actions": ["approve", "reject", "edit", "cancel"],
    }


def resume_observability_state(decision: dict[str, Any]) -> dict[str, Any]:
    update: dict[str, Any] = {}
    started_ms = int_value(decision.get("observability_started_ms"))
    if started_ms > 0:
        update["observability_started_ms"] = started_ms
    checkpoint_status = str(decision.get("checkpoint_status") or "")
    if checkpoint_status:
        update["checkpoint_status"] = checkpoint_status
    trace_context = decision.get("trace_context")
    if isinstance(trace_context, dict):
        update["trace_context"] = trace_context
    return update


def pending_approval_for_card(card: dict[str, Any]) -> dict[str, Any]:
    tool_call_id = str(card.get("tool_call_id") or "")
    card_id = str(card.get("card_id") or "")
    revision = int_value(card.get("revision"))
    return {
        "interrupt_type": "tool_approval",
        "interrupt_id": tool_call_id or card_id,
        "active_tool_call_id": tool_call_id,
        "card_id": card_id,
        "card_revision": revision,
        "base_revision": revision,
        "proposal_id": str(card.get("proposal_id") or f"{card_id}:{tool_call_id}"),
        "idempotency_key": proposal_idempotency_key(card),
        "risk_level": str(card.get("risk_level") or "high"),
        "allowed_actions": ["approve", "reject", "edit", "cancel"],
    }


def apply_proposal_edit(state: dict[str, Any], patch: dict[str, Any]) -> tuple[dict[str, Any], dict[str, Any]]:
    active_card = state.get("active_card")
    card = dict(active_card) if isinstance(active_card, dict) else {}
    if not card:
        return {}, {}
    safe_patch = sanitize_proposal_patch(patch, card)
    if not safe_patch:
        return card, pending_approval_for_card(card)
    current_input = current_card_input(card)
    updated_input = {**current_input, **safe_patch}
    revision = int_value(card.get("revision")) + 1
    card["revision"] = revision
    card["status"] = "needs_approval"
    card["input_json"] = updated_input
    card["display_json"] = {**dict_value(card.get("display_json")), **safe_patch}
    card["fields"] = update_card_fields(card.get("fields"), updated_input)
    debug_json = dict_value(card.get("debug_json"))
    debug_json["input_json"] = updated_input
    debug_json["base_revision"] = revision - 1
    debug_json["safe_patch_keys"] = sorted(safe_patch)
    card["debug_json"] = debug_json
    return card, pending_approval_for_card(card)


def proposal_patch_from_resume(decision: dict[str, Any], state: dict[str, Any]) -> dict[str, Any]:
    edited = decision.get("edited_card_json")
    if not isinstance(edited, dict):
        return {}
    active_card = state.get("active_card")
    card = active_card if isinstance(active_card, dict) else {}
    raw_input = edited.get("input_json")
    if isinstance(raw_input, dict):
        return sanitize_proposal_patch(raw_input, card)
    return sanitize_proposal_patch(edited, card)


def sanitize_proposal_patch(patch: dict[str, Any], card: dict[str, Any]) -> dict[str, Any]:
    allowed = allowed_edit_fields(card)
    safe: dict[str, Any] = {}
    for key, value in patch.items():
        if key not in allowed:
            continue
        if isinstance(value, dict):
            continue
        if isinstance(value, list) and key not in {"checklist", "labels"}:
            continue
        safe[key] = value
    return safe


def allowed_edit_fields(card: dict[str, Any]) -> set[str]:
    allowed = set(ALLOWED_EDIT_FIELDS)
    for field in card.get("fields") or []:
        if isinstance(field, dict):
            key = str(field.get("key") or "")
            if key:
                allowed.add(key)
    return allowed


def current_card_input(card: dict[str, Any]) -> dict[str, Any]:
    input_json = card.get("input_json")
    if isinstance(input_json, dict):
        return dict(input_json)
    debug_json = card.get("debug_json")
    if isinstance(debug_json, dict) and isinstance(debug_json.get("input_json"), dict):
        return dict(debug_json["input_json"])
    display_json = card.get("display_json")
    return dict(display_json) if isinstance(display_json, dict) else {}


def update_card_fields(raw_fields: Any, values: dict[str, Any]) -> list[dict[str, Any]]:
    fields = [dict(field) for field in raw_fields or [] if isinstance(field, dict)]
    for field in fields:
        key = str(field.get("key") or "")
        if key == "client":
            field["value"] = values.get("client_name") or values.get("client_query") or field.get("value") or ""
        elif key == "due_date":
            field["value"] = values.get("due_date_label") or values.get("due_date") or field.get("value") or ""
        elif key in values:
            field["value"] = values[key]
    return fields


def proposal_idempotency_key(card: dict[str, Any]) -> str:
    raw = json.dumps(
        {
            "card_id": str(card.get("card_id") or ""),
            "revision": int_value(card.get("revision")),
            "tool_call_id": str(card.get("tool_call_id") or ""),
        },
        sort_keys=True,
        ensure_ascii=False,
        separators=(",", ":"),
    )
    return "proposal:" + hashlib.sha256(raw.encode()).hexdigest()[:24]


def dict_value(value: Any) -> dict[str, Any]:
    return value if isinstance(value, dict) else {}


def normalize_resume_decision(value: Any) -> dict[str, Any]:
    raw = value if isinstance(value, dict) else {}
    action = str(raw.get("action") or raw.get("decision") or "").strip().lower()
    if action in {"apply", "approved", "approve"}:
        action = "approve"
    elif action in {"rejected", "reject"}:
        action = "reject"
    elif action in {"cancelled", "canceled", "cancel"}:
        action = "cancel"
    elif action == "edit":
        action = "edit"
    else:
        action = "unknown"
    edited = raw.get("edited_card_json") or raw.get("edited_card") or raw.get("patch")
    return {
        "schema": "riders-pulse-resume-decision-v1",
        "request_id": str(raw.get("request_id") or ""),
        "idempotency_key": str(raw.get("idempotency_key") or ""),
        "action": action,
        "reason": str(raw.get("reason") or ""),
        "card_id": str(raw.get("card_id") or ""),
        "card_revision": int_value(raw.get("card_revision")),
        "interrupt_id": str(raw.get("interrupt_id") or ""),
        "active_tool_call_id": str(raw.get("active_tool_call_id") or raw.get("interrupt_id") or ""),
        "edited_card_json": edited if isinstance(edited, dict) else {},
    }


def int_value(value: Any) -> int:
    if isinstance(value, int):
        return value
    if isinstance(value, float):
        return int(value)
    if isinstance(value, str) and value.isdigit():
        return int(value)
    return 0
