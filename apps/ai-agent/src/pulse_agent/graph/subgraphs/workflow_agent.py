from __future__ import annotations

from typing import Any

from pulse_agent.graph.approval_lifecycle import pending_approval_for_card
from pulse_agent.graph.events import add_event


def workflow_patch_preview_card(patch: dict) -> dict:
    return {"card_type": "workflow_patch_preview", "patch": patch}


def workflow_patch_workflow(state: dict[str, Any]) -> dict[str, Any]:
    raw_plan = state.get("plan")
    plan: dict[str, Any] = raw_plan if isinstance(raw_plan, dict) else {}
    raw_page_context = state.get("page_context")
    page_context: dict[str, Any] = raw_page_context if isinstance(raw_page_context, dict) else {}
    raw_preview = plan.get("tool_input_preview")
    preview: dict[str, Any] = raw_preview if isinstance(raw_preview, dict) else {}
    query = str(preview.get("query") if isinstance(preview, dict) else state.get("body") or "")
    workflow_id = str(page_context.get("workflowId") or page_context.get("workflow_id") or "")
    base_revision = int_value(page_context.get("workflowRevision") or page_context.get("workflow_revision"))
    patch = {
        "operation": "propose_workflow_patch",
        "instruction": query,
        "workflow_id": workflow_id,
        "base_revision": base_revision,
    }
    card = {
        "schema": "riders-pulse-ai-card-v1",
        "card_id": f"card-workflow-patch-{workflow_id or 'draft'}",
        "card_type": "workflow_patch_preview",
        "version": 1,
        "revision": 1,
        "tool_call_id": "pending-workflow-patch",
        "intent_key": "workflow.patch",
        "status": "needs_approval",
        "risk_level": "critical",
        "requires_approval": True,
        "title": "Изменение сценария",
        "subtitle": "Проверьте diff перед применением",
        "fields": [
            {"key": "workflow_id", "label": "Workflow", "value": workflow_id, "value_type": "text"},
            {"key": "base_revision", "label": "Revision", "value": base_revision, "value_type": "number"},
            {"key": "instruction", "label": "Изменение", "value": query, "value_type": "text"},
        ],
        "actions": [
            {"action_key": "approve", "label": "Применить", "style": "primary", "confirm_required": True},
            {"action_key": "reject", "label": "Отклонить", "style": "secondary"},
            {"action_key": "edit", "label": "Исправить", "style": "secondary"},
        ],
        "display_json": {"diff": patch, "requires_revision_validation": True},
        "input_json": patch,
        "debug_json": {"hidden_by_default": True, "base_revision": base_revision},
    }
    pending = pending_approval_for_card(card)
    events = list(state.get("events") or [])
    add_event(events, state, "STATUS", "Шаг: read_draft")
    add_event(
        events,
        state,
        "TOOL_STARTED",
        "Читаю workflow draft перед patch preview",
        tool_call={
            "tool_key": "workflow.read_draft",
            "risk_level": "read",
            "status": "running",
            "input_json": {"workflow_id": workflow_id, "base_revision": base_revision},
        },
    )
    add_event(events, state, "STATUS", "Шаг: build_diff")
    add_event(
        events,
        state,
        "TOOL_PROPOSED",
        "Подготовить workflow.patch proposal",
        tool_call={"tool_key": "workflow.patch", "risk_level": "critical", "status": "proposed", "input_json": patch},
    )
    add_event(events, state, "CARD_CREATED", "Подготовил workflow patch preview", card=card)
    add_event(events, state, "APPROVAL_REQUIRED", "Жду подтверждения изменения сценария", interrupt=pending)
    return {
        "events": events,
        "active_card": card,
        "active_card_id": str(card.get("card_id") or ""),
        "active_tool_call_id": str(pending.get("active_tool_call_id") or ""),
        "pending_interrupt": pending,
        "active_mode": "workflow_patch",
        "current_step": "await_approval",
    }


def int_value(value: Any) -> int:
    if isinstance(value, int):
        return value
    if isinstance(value, float):
        return int(value)
    if isinstance(value, str) and value.isdigit():
        return int(value)
    return 0
