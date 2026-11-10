from __future__ import annotations

import json
from typing import Any

from pulse_agent.observability import AI_OBSERVABILITY_SCHEMA, build_ai_observability


def test_ai_observability_summarizes_turn_without_raw_content() -> None:
    events = [
        event("STATUS", "Шаг: triage", 3),
        event("STATUS", "Шаг: run_read_tools", 15),
        event("TOOL_RESULT", "Нашел клиента Орион", 30),
    ]
    payload = build_ai_observability(
        {
            "run_id": "run-1",
            "turn_id": "turn-1",
            "thread_revision": 3,
            "active_mode": "project_read",
            "current_step": "run_read_tools",
            "checkpoint_enabled": True,
            "checkpoint_namespace": "ai_agent_checkpoint",
            "checkpoint_status": "succeeded",
            "trace_context": {
                "trace_id": "trace-1",
                "span_id": "span-1",
                "request_id": "request-1",
                "baggage": {"tenant": "secret tenant value"},
            },
            "intent": {
                "intent_key": "client.search",
                "confidence": 0.93,
                "understanding_source": "model",
                "reason": "model_understanding",
                "model_usage": {
                    "provider_key": "ai.apinet",
                    "model_id": "gpt-test",
                    "input_tokens": 17,
                    "output_tokens": 9,
                    "latency_ms": 110,
                    "metadata_json": {"quota_status": "ok"},
                },
            },
            "tool_results": [{"ok": True}],
            "tool_observations": [
                {
                    "schema": "riders-pulse-tool-observation-v1",
                    "tool_key": "client.search",
                    "run_id": "run-1",
                    "status": "succeeded",
                    "ok": True,
                    "latency_ms": 42,
                    "attempt_count": 1,
                    "normalized_output": {"answer": "Нашел клиента Орион"},
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
        },
        events,
        response_payload={"response_status": "answered", "grounding": {"fallback_discarded": True}},
    )

    assert payload["schema"] == AI_OBSERVABILITY_SCHEMA
    assert payload["turn"]["duration_ms"] == 30
    assert payload["turn"]["time_to_first_event_ms"] == 3
    assert payload["intent"]["key"] == "client.search"
    assert payload["intent"]["source"] == "model"
    assert payload["tools"]["items"] == [
        {
            "tool_key": "client.search",
            "status": "succeeded",
            "ok": True,
            "latency_ms": 42,
            "attempt_count": 1,
            "error_code": "",
            "audit_ref_present": True,
            "entity_ref_count": 1,
        }
    ]
    assert payload["model"]["provider_key"] == "ai.apinet"
    assert payload["model"]["input_tokens"] == 17
    assert payload["model"]["cost_status"] == "pricing_not_configured"
    assert payload["trace"]["baggage_keys"] == ["tenant"]
    assert payload["node_spans"][0]["node"] == "triage"

    rendered = json.dumps(payload, ensure_ascii=False)
    assert "Орион" not in rendered
    assert "Нашел клиента" not in rendered
    assert "secret tenant value" not in rendered


def event(event_type: str, message: str, elapsed_ms: int) -> dict[str, Any]:
    return {
        "event_type": event_type,
        "message": message,
        "observability": {
            "schema": "riders-pulse-ai-event-observability-v1",
            "event_index": 1,
            "elapsed_ms": elapsed_ms,
        },
    }
