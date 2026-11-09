from __future__ import annotations

from datetime import UTC, datetime, timedelta

from pulse_agent.memory_lifecycle import (
    MEMORY_SCHEMA,
    delete_memory_scope,
    export_memory_scope,
    prune_expired_memory,
    reduce_memory_state,
    scope_for_state,
)


def test_memory_lifecycle_scopes_records_and_redacts_secret_turns() -> None:
    state = {
        "company_id": "company-1",
        "project_id": "project-1",
        "user_id": "user-1",
        "session_id": "session-1",
        "thread_id": "thread-1",
        "turn_id": "turn-1",
        "run_id": "run-1",
        "body": "Найди клиента Орион",
        "plan": {"intent": "client.search"},
        "tool_observations": [
            {
                "entity_refs": [
                    {
                        "entity_type": "client",
                        "entity_id": "client-1",
                        "label": "Орион",
                        "source": "client.search",
                    }
                ]
            }
        ],
        "working_memory": {
            "recent_turns": [
                {
                    "scope": {
                        "company_id": "other",
                        "project_id": "project-1",
                        "user_id": "user-1",
                        "session_id": "session-1",
                        "thread_id": "thread-1",
                    },
                    "value": {"user_summary": "other tenant"},
                    "expires_at": (datetime.now(UTC) + timedelta(days=1)).isoformat(),
                }
            ]
        },
    }

    update = reduce_memory_state(state, final_response="Нашел клиента Орион")
    memory = update["working_memory"]

    assert memory["schema"] == MEMORY_SCHEMA
    assert memory["scope"] == scope_for_state(state)
    assert [item["value"]["user_summary"] for item in memory["recent_turns"]] == ["Найди клиента Орион"]
    assert memory["entities"][0]["value"]["entity_id"] == "client-1"
    assert memory["entities"][0]["provenance"]["source"] == "client.search"
    assert update["active_topic"]["entity"]["entity_id"] == "client-1"

    secret_update = reduce_memory_state({**state, "body": "мой api key secret"}, final_response="")
    assert secret_update["working_memory"]["recent_turns"] == []


def test_memory_export_delete_and_expiry_hooks() -> None:
    scope = {
        "company_id": "company-1",
        "project_id": "project-1",
        "user_id": "user-1",
        "session_id": "session-1",
        "thread_id": "thread-1",
    }
    now = datetime.now(UTC)
    memory = {
        "recent_turns": [
            {"scope": scope, "value": {"user_summary": "fresh"}, "expires_at": (now + timedelta(days=1)).isoformat()},
            {"scope": scope, "value": {"user_summary": "old"}, "expires_at": (now - timedelta(days=1)).isoformat()},
        ],
        "entities": [
            {"scope": scope, "value": {"entity_id": "client-1"}, "expires_at": (now + timedelta(days=1)).isoformat()}
        ],
    }

    exported = export_memory_scope(memory, scope)
    assert len(exported["recent_turns"]) == 2
    assert len(exported["entities"]) == 1

    pruned = prune_expired_memory(memory, now=now)
    assert [item["value"]["user_summary"] for item in pruned["recent_turns"]] == ["fresh"]

    deleted = delete_memory_scope(memory, scope)
    assert deleted["recent_turns"] == []
    assert deleted["entities"] == []
