from __future__ import annotations

from datetime import UTC, datetime
from typing import Any

from pulse_agent.graph.observations import sanitize_json

DIALOGUE_STATE_SCHEMA = "riders-pulse-dialogue-state-v1"
MAX_SUSPENDED_TOPICS = 5
MAX_RESOLVED_ENTITIES = 20
MAX_LAST_REFERENTS = 5
MAX_LAST_TOOL_OBSERVATIONS = 8


def reduce_dialogue_state(
    state: dict[str, Any],
    *,
    final_response: str = "",
    now: datetime | None = None,
) -> dict[str, Any]:
    current_time = now or datetime.now(UTC)
    previous = migrate_dialogue_state(state.get("dialogue_state"), state, now=current_time)
    scope = scope_for_state(state)
    resolved_entities = merge_resolved_entities(previous.get("resolved_entities"), state)
    active_topic = active_topic_for_state(state, resolved_entities, scope, current_time)
    suspended_topics = suspended_topics_for_state(previous, state)
    return {
        "dialogue_state": {
            "schema": DIALOGUE_STATE_SCHEMA,
            "version": 1,
            "scope": scope,
            "active_topic": active_topic,
            "suspended_topics": suspended_topics,
            "resolved_entities": resolved_entities,
            "last_referents": resolved_entities[-MAX_LAST_REFERENTS:],
            "unresolved_slots": unresolved_slots_for_state(state),
            "pending_action": pending_action_for_state(state),
            "last_tool_observations": last_tool_observations(state),
            "conversation_summary": conversation_summary_for_state(state, final_response),
            "locale": str(state.get("locale") or "ru-RU"),
            "timezone": str(state.get("timezone") or ""),
            "local_date": str(state.get("local_date") or ""),
            "page_frame": page_frame_for_state(state),
            "field_policies": field_policies(),
            "updated_at": current_time.isoformat(),
        },
        "active_topic": active_topic,
        "suspended_topics": suspended_topics,
        "resolved_entities": resolved_entities,
        "unresolved_slots": unresolved_slots_for_state(state),
    }


def migrate_dialogue_state(
    value: Any,
    legacy_state: dict[str, Any] | None = None,
    *,
    now: datetime | None = None,
) -> dict[str, Any]:
    legacy_state = legacy_state or {}
    current_time = now or datetime.now(UTC)
    if isinstance(value, dict) and value.get("schema") == DIALOGUE_STATE_SCHEMA:
        migrated = sanitize_json(value)
        if not isinstance(migrated, dict):
            return {}
        migrated["version"] = int_value(migrated.get("version")) or 1
        migrated["active_topic"] = dict_value(migrated.get("active_topic"))
        migrated["suspended_topics"] = bounded_topics(migrated.get("suspended_topics"))
        migrated["resolved_entities"] = bounded_entities(migrated.get("resolved_entities"))
        migrated["last_referents"] = bounded_entities(migrated.get("last_referents"))[-MAX_LAST_REFERENTS:]
        migrated["unresolved_slots"] = dict_value(migrated.get("unresolved_slots"))
        migrated["pending_action"] = dict_value(migrated.get("pending_action"))
        migrated["last_tool_observations"] = bounded_observations(migrated.get("last_tool_observations"))
        return migrated
    scope = scope_for_state(legacy_state)
    resolved = bounded_entities(legacy_state.get("resolved_entities"))
    active_topic = dict_value(legacy_state.get("active_topic"))
    if not active_topic:
        active_topic = active_topic_for_state(legacy_state, resolved, scope, current_time)
    return {
        "schema": DIALOGUE_STATE_SCHEMA,
        "version": 1,
        "scope": scope,
        "active_topic": active_topic,
        "suspended_topics": bounded_topics(legacy_state.get("suspended_topics")),
        "resolved_entities": resolved,
        "last_referents": resolved[-MAX_LAST_REFERENTS:],
        "unresolved_slots": dict_value(legacy_state.get("unresolved_slots")),
        "pending_action": pending_action_for_state(legacy_state),
        "last_tool_observations": last_tool_observations(legacy_state),
        "conversation_summary": conversation_summary_for_state(
            legacy_state, str(legacy_state.get("final_response") or "")
        ),
        "locale": str(legacy_state.get("locale") or "ru-RU"),
        "timezone": str(legacy_state.get("timezone") or ""),
        "local_date": str(legacy_state.get("local_date") or ""),
        "page_frame": page_frame_for_state(legacy_state),
        "field_policies": field_policies(),
        "updated_at": current_time.isoformat(),
    }


def active_topic_for_state(
    state: dict[str, Any],
    resolved_entities: list[dict[str, Any]],
    scope: dict[str, str],
    now: datetime,
) -> dict[str, Any]:
    plan = dict_value(state.get("plan"))
    intent = str(plan.get("intent") or dict_value(state.get("intent")).get("intent_key") or "")
    entity = (
        resolved_entities[-1] if resolved_entities else dict_value(dict_value(state.get("active_topic")).get("entity"))
    )
    topic_id = ":".join(
        item for item in (intent, str(entity.get("entity_type") or ""), str(entity.get("entity_id") or "")) if item
    )
    return {
        "schema": "riders-pulse-dialogue-topic-v1",
        "topic_id": topic_id or "conversation",
        "intent": intent or "conversation",
        "branch": str(state.get("branch") or ""),
        "active_mode": str(state.get("active_mode") or ""),
        "entity": entity,
        "scope": scope,
        "updated_at": now.isoformat(),
    }


def merge_resolved_entities(previous: Any, state: dict[str, Any]) -> list[dict[str, Any]]:
    by_key: dict[str, dict[str, Any]] = {entity_key(item): item for item in bounded_entities(previous)}
    for ref in entity_refs_from_observations(state.get("tool_observations")):
        key = entity_key(ref)
        if key:
            by_key[key] = ref
    for ref in bounded_entities(state.get("resolved_entities")):
        key = entity_key(ref)
        if key:
            by_key[key] = ref
    return list(by_key.values())[-MAX_RESOLVED_ENTITIES:]


def suspended_topics_for_state(previous: dict[str, Any], state: dict[str, Any]) -> list[dict[str, Any]]:
    topics = bounded_topics(previous.get("suspended_topics"))
    policy = dict_value(state.get("topic_policy"))
    if policy.get("decision") == "return_to_pending_action":
        return topics
    if not policy.get("suspend_active_topic"):
        return topics
    active_topic = dict_value(previous.get("active_topic")) or dict_value(state.get("active_topic"))
    if not active_topic:
        return topics
    pending = dict_value(policy.get("pending_action")) or pending_action_for_state(state)
    suspended = sanitize_json({**active_topic, "pending_action": pending, "suspended_reason": policy.get("reason")})
    if not isinstance(suspended, dict):
        return topics
    key = str(suspended.get("topic_id") or "")
    topics = [item for item in topics if str(item.get("topic_id") or "") != key]
    return bounded_topics([*topics, suspended])


def unresolved_slots_for_state(state: dict[str, Any]) -> dict[str, Any]:
    plan = dict_value(state.get("plan"))
    preview = dict_value(plan.get("tool_input_preview"))
    intent = str(plan.get("intent") or "")
    slots: dict[str, Any] = {}
    if intent == "task.create":
        if preview.get("client_query") and not preview.get("client_id"):
            slots["client"] = {
                "reason": "client_query_unresolved",
                "query": str(preview.get("client_query") or "")[:240],
            }
        if not str(preview.get("title") or "").strip():
            slots["title"] = {"reason": "missing_task_title"}
    if intent == "client.create" and not str(preview.get("name") or preview.get("client_name") or "").strip():
        slots["client_name"] = {"reason": "missing_client_name"}
    if intent == "workflow.patch" and not str(preview.get("workflow_id") or "").strip():
        slots["workflow_id"] = {"reason": "missing_workflow_id"}
    return sanitize_json(slots) if isinstance(sanitize_json(slots), dict) else {}


def pending_action_for_state(state: dict[str, Any]) -> dict[str, Any]:
    pending = dict_value(state.get("pending_interrupt"))
    active_card = dict_value(state.get("active_card"))
    active_tool_call_id = str(pending.get("active_tool_call_id") or state.get("active_tool_call_id") or "")
    if not active_tool_call_id and not active_card:
        return {}
    return {
        "schema": "riders-pulse-dialogue-pending-action-v1",
        "active_tool_call_id": active_tool_call_id,
        "active_proposal_id": str(state.get("active_proposal_id") or ""),
        "card_id": str(active_card.get("card_id") or pending.get("card_id") or ""),
        "card_type": str(active_card.get("card_type") or ""),
        "intent_key": str(active_card.get("intent_key") or dict_value(state.get("plan")).get("intent") or ""),
        "risk_level": str(active_card.get("risk_level") or ""),
        "status": str(active_card.get("status") or "pending"),
    }


def last_tool_observations(state: dict[str, Any]) -> list[dict[str, Any]]:
    observations = []
    for item in list_value(state.get("tool_observations"))[-MAX_LAST_TOOL_OBSERVATIONS:]:
        observations.append(
            sanitize_json(
                {
                    "schema": item.get("schema"),
                    "tool_key": item.get("tool_key"),
                    "status": item.get("status"),
                    "ok": item.get("ok"),
                    "entity_refs": item.get("entity_refs") if isinstance(item.get("entity_refs"), list) else [],
                    "audit_ref": item.get("audit_ref"),
                }
            )
        )
    return [item for item in observations if isinstance(item, dict)]


def conversation_summary_for_state(state: dict[str, Any], final_response: str) -> dict[str, Any]:
    return {
        "schema": "riders-pulse-dialogue-summary-v1",
        "last_user_summary": str(state.get("body") or state.get("current_body") or "")[:500],
        "last_assistant_summary": final_response[:500],
        "last_intent": str(dict_value(state.get("plan")).get("intent") or ""),
        "thread_revision": int_value(state.get("thread_revision")),
    }


def page_frame_for_state(state: dict[str, Any]) -> dict[str, Any]:
    page = dict_value(state.get("page_context"))
    sanitized = sanitize_json(
        {
            "route": page.get("route") or page.get("pathname") or "",
            "page_kind": page.get("page_kind") or page.get("pageKind") or "",
            "visible_panel": page.get("visible_panel") or page.get("visiblePanel") or "",
            "selected_node_id": page.get("selected_node_id") or page.get("selectedNodeId") or "",
            "selected_action_key": page.get("selected_action_key") or page.get("selectedActionKey") or "",
        }
    )
    return sanitized if isinstance(sanitized, dict) else {}


def field_policies() -> dict[str, dict[str, str]]:
    return {
        "active_topic": {"producer": "dialogue_state.reduce", "persistence": "checkpoint", "redaction": "safe_ids"},
        "suspended_topics": {"producer": "topic_policy", "persistence": "checkpoint", "redaction": "safe_ids"},
        "resolved_entities": {"producer": "tool_observations", "persistence": "checkpoint", "redaction": "safe_ids"},
        "unresolved_slots": {"producer": "slot_resolver", "persistence": "checkpoint", "redaction": "safe_summary"},
        "pending_action": {"producer": "approval_lifecycle", "persistence": "checkpoint", "redaction": "safe_ids"},
        "last_tool_observations": {"producer": "ToolGateway", "persistence": "checkpoint", "redaction": "sanitized"},
        "conversation_summary": {"producer": "finalizer", "persistence": "checkpoint", "redaction": "summary_only"},
    }


def bounded_topics(value: Any) -> list[dict[str, Any]]:
    return [
        sanitize_json(item)
        for item in list_value(value)[-MAX_SUSPENDED_TOPICS:]
        if isinstance(sanitize_json(item), dict)
    ]


def bounded_entities(value: Any) -> list[dict[str, Any]]:
    return [
        sanitize_json(item)
        for item in list_value(value)[-MAX_RESOLVED_ENTITIES:]
        if isinstance(sanitize_json(item), dict)
    ]


def bounded_observations(value: Any) -> list[dict[str, Any]]:
    return [
        sanitize_json(item)
        for item in list_value(value)[-MAX_LAST_TOOL_OBSERVATIONS:]
        if isinstance(sanitize_json(item), dict)
    ]


def entity_refs_from_observations(value: Any) -> list[dict[str, Any]]:
    refs: list[dict[str, Any]] = []
    for observation in list_value(value):
        for ref in list_value(observation.get("entity_refs")):
            normalized = sanitize_json(ref)
            if isinstance(normalized, dict):
                refs.append(normalized)
    return refs


def scope_for_state(state: dict[str, Any]) -> dict[str, str]:
    return {
        "company_id": str(state.get("company_id") or ""),
        "project_id": str(state.get("project_id") or ""),
        "user_id": str(state.get("user_id") or ""),
        "session_id": str(state.get("session_id") or ""),
        "thread_id": str(state.get("thread_id") or ""),
    }


def entity_key(entity: dict[str, Any]) -> str:
    return ":".join(str(entity.get(key) or "") for key in ("entity_type", "entity_id", "label", "query")).strip(":")


def dict_value(value: Any) -> dict[str, Any]:
    return value if isinstance(value, dict) else {}


def list_value(value: Any) -> list[dict[str, Any]]:
    return [item for item in value if isinstance(item, dict)] if isinstance(value, list) else []


def int_value(value: Any) -> int:
    if isinstance(value, int):
        return value
    if isinstance(value, float):
        return int(value)
    if isinstance(value, str) and value.isdigit():
        return int(value)
    return 0
