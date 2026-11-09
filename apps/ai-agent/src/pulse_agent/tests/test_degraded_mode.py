from __future__ import annotations

from typing import Any

from pulse_agent.config import Settings
from pulse_agent.degraded_mode import (
    MODEL_GATEWAY_BREAKER,
    TOOL_GATEWAY_BREAKER,
    CircuitBreaker,
    dependency_capability_matrix,
)
from pulse_agent.graph import tool_gateway
from pulse_agent.graph.builder import build_graph
from pulse_agent.graph.intent_schema import IntentKey, classify_turn


def test_dependency_capability_matrix_reports_degraded_without_required_gateways(monkeypatch: Any) -> None:
    monkeypatch.delenv("PULSE_MODEL_GATEWAY_READY", raising=False)
    monkeypatch.delenv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP", raising=False)

    matrix = dependency_capability_matrix(Settings(service_token=""), checkpoint_enabled=False)

    assert matrix["schema"] == "riders-pulse-ai-capability-matrix-v1"
    assert matrix["status"] == "degraded"
    assert matrix["dependencies"]["checkpointer"]["status"] == "degraded"
    assert matrix["dependencies"]["tool_gateway"]["reason"] == "tool_gateway_graph_loop_not_ready"
    assert matrix["recovery_action"]


def test_circuit_breaker_opens_after_retry_budget() -> None:
    breaker = CircuitBreaker(failure_threshold=2, recovery_seconds=60)

    assert breaker.allow_request(now=1.0) is True
    breaker.record_failure(now=1.0)
    assert breaker.allow_request(now=2.0) is True
    breaker.record_failure(now=2.0)

    assert breaker.allow_request(now=3.0) is False
    assert breaker.snapshot()["failures"] == 2


def test_tool_gateway_outage_returns_failed_observation_not_graph_error(monkeypatch: Any) -> None:
    TOOL_GATEWAY_BREAKER.record_success()
    calls: list[dict[str, Any]] = []

    class FailingPulseToolsClient:
        def __init__(self, *_args: Any, **_kwargs: Any) -> None:
            pass

        async def run_read_tool(self, **kwargs: Any) -> dict[str, Any]:
            calls.append(kwargs)
            raise RuntimeError("tool gateway down")

    monkeypatch.setenv("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP", "true")
    monkeypatch.setattr(tool_gateway, "PulseToolsClient", FailingPulseToolsClient)
    graph_builder = build_graph()
    assert graph_builder is not None
    graph = graph_builder.compile()

    state = graph.invoke(
        {
            "company_id": "company-1",
            "project_id": "project-1",
            "user_id": "user-1",
            "session_id": "session-1",
            "thread_id": "thread-1",
            "run_id": "run-1",
            "body": "Найди клиента Орион",
        }
    )

    assert calls and calls[0]["tool_key"] == "client.search"
    assert state["tool_observations"][-1]["status"] == "failed"
    assert state["tool_observations"][-1]["error_code"] == "tool_gateway_error"
    assert "ограниченном режиме" in state["final_response"]


def test_model_gateway_outage_keeps_ambiguous_turn_unknown(monkeypatch: Any) -> None:
    MODEL_GATEWAY_BREAKER.record_success()
    monkeypatch.setenv("PULSE_AI_MODEL_TURN_UNDERSTANDING", "true")

    class FailingModelGateway:
        def __init__(self, *_args: Any, **_kwargs: Any) -> None:
            pass

        async def intent_router(self, *_args: Any, **_kwargs: Any) -> dict[str, Any]:
            raise RuntimeError("model gateway down")

    monkeypatch.setattr("pulse_agent.graph.turn_understanding.PulseModelGatewayClient", FailingModelGateway)

    classification = classify_turn("could you surface the Adv account?")

    assert classification.intent_key == IntentKey.UNKNOWN
    assert classification.requires_clarification is True
