from __future__ import annotations

from typing import Any

from pulse_agent.graph import read_loop


def test_read_loop_replans_with_normalized_observation(monkeypatch: Any) -> None:
    calls: list[tuple[str, dict[str, Any]]] = []

    def fake_run_read_tool(_state: dict[str, Any], tool_key: str, input_json: dict[str, Any]) -> dict[str, Any]:
        calls.append((tool_key, input_json))
        if tool_key == "client.search":
            return {
                "ok": True,
                "status": "succeeded",
                "output_json": {"answer": "Клиент найден", "items": [{"id": "client-1", "name": "Орион"}]},
                "audit_event_id": "audit-client",
            }
        return {
            "ok": True,
            "status": "succeeded",
            "output_json": {"answer": "Оплата 30 июня"},
            "audit_event_id": "audit-payment",
        }

    monkeypatch.setattr(read_loop, "run_read_tool_if_enabled", fake_run_read_tool)

    result = read_loop.read_context(
        {
            "run_id": "run-1",
            "turn_id": "turn-1",
            "plan": {"intent": "payment.lookup_due", "required_tools": ["client.search", "payment.lookup_due"]},
            "body": "Когда оплата у Орион?",
            "events": [],
        }
    )

    assert [tool_key for tool_key, _input in calls] == ["client.search", "payment.lookup_due"]
    assert calls[1][1]["client_id"] == "client-1"
    assert result["final_response"] == "Оплата 30 июня"
    assert result["tool_observations"][0]["supporting_only"] is True
    assert result["tool_observations"][0]["run_id"] == "run-1"
