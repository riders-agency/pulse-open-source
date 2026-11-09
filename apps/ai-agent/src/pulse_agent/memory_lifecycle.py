from __future__ import annotations

from datetime import UTC, datetime, timedelta
from typing import Any

from pulse_agent.graph.observations import sanitize_json

MEMORY_SCHEMA = "riders-pulse-working-memory-v1"
RECENT_TURN_TTL_DAYS = 7
ENTITY_TTL_DAYS = 30
MAX_RECENT_TURNS = 8
MAX_ENTITY_MEMORIES = 20

SECRET_MARKERS = {"api key", "authorization", "cookie", "password", "secret", "token", "парол", "секрет", "токен"}


def reduce_memory_state(state: dict[str, Any], *, final_response: str = "") -> dict[str, Any]:
    now = datetime.now(UTC)
    previous = working_memory_dict(state.get("working_memory"))
    scope = scope_for_state(state)
    recent_turns = scoped_records(previous.get("recent_turns"), scope, now=now)
    entity_records = scoped_records(previous.get("entities"), scope, now=now)
    recent_turns = append_recent_turn(recent_turns, state, final_response, scope, now)
    entity_records = merge_entity_records(entity_records, observations_from_state(state), scope, now)
    working_memory = {
        "schema": MEMORY_SCHEMA,
        "scope": scope,
        "classes": ["recent_turn", "working_entity"],
        "recent_turns": recent_turns[-MAX_RECENT_TURNS:],
        "entities": entity_records[-MAX_ENTITY_MEMORIES:],
        "retrieval_budget": {"recent_turns": MAX_RECENT_TURNS, "entities": MAX_ENTITY_MEMORIES},
        "retention": {
            "recent_turn_ttl_days": RECENT_TURN_TTL_DAYS,
            "entity_ttl_days": ENTITY_TTL_DAYS,
        },
        "model_extraction_enabled": False,
        "updated_at": now.isoformat(),
    }
    active_topic = active_topic_from_state(state, entity_records, scope, now)
    resolved_entities = [record["value"] for record in entity_records[-5:] if isinstance(record.get("value"), dict)]
    return {
        "working_memory": working_memory,
        "long_term_memory_refs": scoped_records(state.get("long_term_memory_refs"), scope, now=now),
        "active_topic": active_topic,
        "resolved_entities": resolved_entities,
        "suspended_topics": scoped_records(state.get("suspended_topics"), scope, now=now),
        "unresolved_slots": {},
    }


def export_memory_scope(memory: dict[str, Any], scope: dict[str, str]) -> dict[str, Any]:
    return {
        "schema": MEMORY_SCHEMA,
        "scope": scope,
        "recent_turns": scoped_records(memory.get("recent_turns"), scope),
        "entities": scoped_records(memory.get("entities"), scope),
    }


def delete_memory_scope(memory: dict[str, Any], scope: dict[str, str]) -> dict[str, Any]:
    cleaned = dict(memory)
    cleaned["recent_turns"] = [
        item for item in record_list(memory.get("recent_turns")) if not scope_matches(item, scope)
    ]
    cleaned["entities"] = [item for item in record_list(memory.get("entities")) if not scope_matches(item, scope)]
    return cleaned


def prune_expired_memory(memory: dict[str, Any], *, now: datetime) -> dict[str, Any]:
    cleaned = dict(memory)
    cleaned["recent_turns"] = scoped_records(memory.get("recent_turns"), {}, now=now, match_scope=False)
    cleaned["entities"] = scoped_records(memory.get("entities"), {}, now=now, match_scope=False)
    return cleaned


def scope_for_state(state: dict[str, Any]) -> dict[str, str]:
    return {
        "company_id": str(state.get("company_id") or ""),
        "project_id": str(state.get("project_id") or ""),
        "user_id": str(state.get("user_id") or ""),
        "session_id": str(state.get("session_id") or ""),
        "thread_id": str(state.get("thread_id") or ""),
    }


def append_recent_turn(
    records: list[dict[str, Any]],
    state: dict[str, Any],
    final_response: str,
    scope: dict[str, str],
    now: datetime,
) -> list[dict[str, Any]]:
    body = str(state.get("body") or state.get("current_body") or "").strip()
    if not body or contains_secret(body):
        return records
    record = {
        "schema": "riders-pulse-memory-record-v1",
        "memory_class": "recent_turn",
        "scope": scope,
        "value": {
            "turn_id": str(state.get("turn_id") or ""),
            "intent": str((state.get("plan") or {}).get("intent") if isinstance(state.get("plan"), dict) else ""),
            "user_summary": body[:500],
            "assistant_summary": final_response[:500],
        },
        "provenance": {"source": "user_turn", "run_id": str(state.get("run_id") or "")},
        "confidence": 1.0,
        "sensitivity": "normal",
        "created_at": now.isoformat(),
        "expires_at": (now + timedelta(days=RECENT_TURN_TTL_DAYS)).isoformat(),
    }
    return records + [record]


def merge_entity_records(
    records: list[dict[str, Any]],
    observations: list[dict[str, Any]],
    scope: dict[str, str],
    now: datetime,
) -> list[dict[str, Any]]:
    by_key = {memory_key(record): record for record in records if memory_key(record)}
    for ref in entity_refs(observations):
        if contains_secret(str(ref)):
            continue
        record = {
            "schema": "riders-pulse-memory-record-v1",
            "memory_class": "working_entity",
            "scope": scope,
            "value": sanitize_json(ref),
            "provenance": {"source": str(ref.get("source") or "tool_observation")},
            "confidence": float(ref.get("confidence") or 0.8),
            "sensitivity": "normal",
            "created_at": now.isoformat(),
            "expires_at": (now + timedelta(days=ENTITY_TTL_DAYS)).isoformat(),
        }
        by_key[memory_key(record)] = record
    return list(by_key.values())


def active_topic_from_state(
    state: dict[str, Any],
    entity_records: list[dict[str, Any]],
    scope: dict[str, str],
    now: datetime,
) -> dict[str, Any]:
    intent = str((state.get("plan") or {}).get("intent") if isinstance(state.get("plan"), dict) else "")
    latest_entity = (
        entity_records[-1]["value"] if entity_records and isinstance(entity_records[-1].get("value"), dict) else {}
    )
    return {
        "schema": "riders-pulse-topic-v1",
        "scope": scope,
        "intent": intent,
        "entity": latest_entity,
        "updated_at": now.isoformat(),
    }


def scoped_records(
    value: Any,
    scope: dict[str, str],
    *,
    now: datetime | None = None,
    match_scope: bool = True,
) -> list[dict[str, Any]]:
    records = []
    for item in record_list(value):
        if match_scope and not scope_matches(item, scope):
            continue
        if now is not None and is_expired(item, now):
            continue
        records.append(item)
    return records


def observations_from_state(state: dict[str, Any]) -> list[dict[str, Any]]:
    value = state.get("tool_observations")
    return [item for item in value if isinstance(item, dict)] if isinstance(value, list) else []


def entity_refs(observations: list[dict[str, Any]]) -> list[dict[str, Any]]:
    refs: list[dict[str, Any]] = []
    for observation in observations:
        raw_refs = observation.get("entity_refs")
        if isinstance(raw_refs, list):
            refs.extend(item for item in raw_refs if isinstance(item, dict))
    return refs


def working_memory_dict(value: Any) -> dict[str, Any]:
    return value if isinstance(value, dict) else {}


def record_list(value: Any) -> list[dict[str, Any]]:
    return [item for item in value if isinstance(item, dict)] if isinstance(value, list) else []


def scope_matches(record: dict[str, Any], scope: dict[str, str]) -> bool:
    record_scope = record.get("scope")
    if not isinstance(record_scope, dict):
        return False
    return all(str(record_scope.get(key) or "") == expected for key, expected in scope.items())


def is_expired(record: dict[str, Any], now: datetime) -> bool:
    expires_at = str(record.get("expires_at") or "")
    if not expires_at:
        return False
    try:
        return datetime.fromisoformat(expires_at) <= now
    except ValueError:
        return True


def memory_key(record: dict[str, Any]) -> str:
    value = record.get("value")
    if not isinstance(value, dict):
        return ""
    entity_type = str(value.get("entity_type") or "")
    entity_id = str(value.get("entity_id") or "")
    label = str(value.get("label") or value.get("query") or "")
    return ":".join((entity_type, entity_id, label)).strip(":")


def contains_secret(value: str) -> bool:
    lower = value.lower()
    return any(marker in lower for marker in SECRET_MARKERS)
