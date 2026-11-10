from __future__ import annotations

from datetime import UTC, datetime

from pulse_agent.dialogue_state import DIALOGUE_STATE_SCHEMA, migrate_dialogue_state, reduce_dialogue_state


def test_dialogue_state_reducer_builds_versioned_state_with_pending_action() -> None:
    now = datetime(2026, 6, 25, 10, 0, tzinfo=UTC)
    state = {
        "company_id": "company-1",
        "project_id": "project-1",
        "user_id": "user-1",
        "session_id": "session-1",
        "thread_id": "thread-1",
        "thread_revision": 7,
        "locale": "ru-RU",
        "timezone": "Europe/Moscow",
        "local_date": "2026-06-25",
        "body": "Создай задачу для Орион",
        "plan": {
            "intent": "task.create",
            "tool_input_preview": {"client_query": "Орион", "title": ""},
        },
        "page_context": {"route": "/work", "visiblePanel": "board"},
        "tool_observations": [
            {
                "schema": "riders-pulse-tool-observation-v1",
                "tool_key": "client.search",
                "status": "succeeded",
                "ok": True,
                "entity_refs": [
                    {
                        "entity_type": "client",
                        "entity_id": "client-1",
                        "label": "Орион",
                        "source": "client.search",
                    }
                ],
                "audit_ref": "audit-1",
            }
        ],
        "active_card": {
            "card_id": "card-1",
            "card_type": "task_create_proposal",
            "intent_key": "task.create",
            "risk_level": "high",
            "status": "needs_approval",
        },
        "pending_interrupt": {"active_tool_call_id": "tool-1", "card_id": "card-1"},
        "dialogue_state": {
            "schema": DIALOGUE_STATE_SCHEMA,
            "suspended_topics": [{"topic_id": f"old-{idx}"} for idx in range(8)],
        },
    }

    update = reduce_dialogue_state(state, final_response="Подготовил карточку", now=now)
    dialogue = update["dialogue_state"]

    assert dialogue["schema"] == DIALOGUE_STATE_SCHEMA
    assert dialogue["active_topic"]["topic_id"] == "task.create:client:client-1"
    assert dialogue["resolved_entities"][0]["entity_id"] == "client-1"
    assert dialogue["last_referents"][0]["label"] == "Орион"
    assert dialogue["unresolved_slots"]["client"]["reason"] == "client_query_unresolved"
    assert dialogue["unresolved_slots"]["title"]["reason"] == "missing_task_title"
    assert dialogue["pending_action"]["active_tool_call_id"] == "tool-1"
    assert dialogue["last_tool_observations"][0]["tool_key"] == "client.search"
    assert dialogue["conversation_summary"]["thread_revision"] == 7
    assert dialogue["page_frame"]["route"] == "/work"
    assert len(dialogue["suspended_topics"]) == 5
    assert dialogue["field_policies"]["resolved_entities"]["redaction"] == "safe_ids"
    assert update["active_topic"] == dialogue["active_topic"]


def test_dialogue_state_migrates_legacy_checkpoint_fields() -> None:
    now = datetime(2026, 6, 25, 10, 5, tzinfo=UTC)
    migrated = migrate_dialogue_state(
        None,
        {
            "company_id": "company-1",
            "project_id": "project-1",
            "user_id": "user-1",
            "session_id": "session-1",
            "thread_id": "thread-1",
            "active_topic": {"topic_id": "legacy-topic", "intent": "client.search"},
            "suspended_topics": [{"topic_id": "old"}],
            "resolved_entities": [{"entity_type": "client", "entity_id": "client-1", "label": "Орион"}],
            "unresolved_slots": {"client": {"reason": "ambiguous"}},
            "pending_interrupt": {"active_tool_call_id": "tool-1"},
            "body": "найди клиента",
            "final_response": "Нашел",
        },
        now=now,
    )

    assert migrated["schema"] == DIALOGUE_STATE_SCHEMA
    assert migrated["active_topic"]["topic_id"] == "legacy-topic"
    assert migrated["resolved_entities"][0]["entity_id"] == "client-1"
    assert migrated["last_referents"][0]["label"] == "Орион"
    assert migrated["unresolved_slots"]["client"]["reason"] == "ambiguous"
    assert migrated["pending_action"]["active_tool_call_id"] == "tool-1"
    assert migrated["conversation_summary"]["last_assistant_summary"] == "Нашел"


def test_dialogue_state_migration_normalizes_existing_schema_and_redacts() -> None:
    migrated = migrate_dialogue_state(
        {
            "schema": DIALOGUE_STATE_SCHEMA,
            "version": "1",
            "active_topic": {"topic_id": "topic"},
            "suspended_topics": [{"topic_id": f"old-{idx}"} for idx in range(7)],
            "resolved_entities": [{"entity_type": "client", "entity_id": str(idx)} for idx in range(25)],
            "last_tool_observations": [{"tool_key": "client.search", "status": "succeeded", "token": "must-redact"}],
        },
        {},
    )

    assert migrated["version"] == 1
    assert len(migrated["suspended_topics"]) == 5
    assert len(migrated["resolved_entities"]) == 20
    assert migrated["last_tool_observations"][0]["token"] == "[redacted]"
