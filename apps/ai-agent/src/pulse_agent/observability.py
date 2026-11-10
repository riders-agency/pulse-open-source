from __future__ import annotations

import os
import time
from typing import Any

from pulse_agent.degraded_mode import MODEL_GATEWAY_BREAKER, TOOL_GATEWAY_BREAKER
from pulse_agent.grounded_response import current_turn_observations

AI_OBSERVABILITY_SCHEMA = "riders-pulse-ai-observability-v1"


def monotonic_ms() -> int:
    return time.monotonic_ns() // 1_000_000


def event_observability(state: dict[str, Any], *, seq: int) -> dict[str, Any]:
    started_ms = int_value(state.get("observability_started_ms"))
    current_ms = monotonic_ms()
    elapsed_ms = max(0, current_ms - started_ms) if started_ms > 0 else 0
    return {
        "schema": "riders-pulse-ai-event-observability-v1",
        "event_index": seq,
        "elapsed_ms": elapsed_ms,
    }


def build_ai_observability(
    state: dict[str, Any],
    events: list[dict[str, Any]],
    *,
    response_payload: dict[str, Any] | None = None,
) -> dict[str, Any]:
    response_payload = response_payload or {}
    observations = current_turn_observations(state)
    intent = dict_value(state.get("intent"))
    trace_context = dict_value(state.get("trace_context"))
    duration_ms = max_elapsed_ms(events)
    return {
        "schema": AI_OBSERVABILITY_SCHEMA,
        "release": {
            "service": "pulse-ai-agent",
            "version": str(state.get("service_version") or os.environ.get("PULSE_RELEASE_ID") or "unknown"),
        },
        "turn": {
            "run_id": str(state.get("run_id") or ""),
            "turn_id": str(state.get("turn_id") or ""),
            "thread_revision": int_value(state.get("thread_revision")),
            "active_mode": str(state.get("active_mode") or ""),
            "current_step": str(state.get("current_step") or ""),
            "duration_ms": duration_ms,
            "time_to_first_event_ms": time_to_first_event_ms(events),
            "event_count": len(events),
        },
        "intent": {
            "key": str(intent.get("intent_key") or intent.get("intent") or ""),
            "confidence": float_value(intent.get("confidence")),
            "source": str(intent.get("understanding_source") or "deterministic"),
            "reason_code": str(intent.get("reason") or ""),
            "requires_clarification": bool(intent.get("requires_clarification")),
        },
        "quality": {
            "response_status": str(response_payload.get("response_status") or ""),
            "clarification_required": bool(intent.get("requires_clarification")),
            "fallback_discarded": bool(dict_value(response_payload.get("grounding")).get("fallback_discarded")),
            "tool_error_count": len([item for item in observations if not item.get("ok")]),
            "degraded": any(not item.get("ok") for item in observations)
            or MODEL_GATEWAY_BREAKER.snapshot()["state"] == "open"
            or TOOL_GATEWAY_BREAKER.snapshot()["state"] == "open",
        },
        "tools": tool_observability(observations),
        "model": model_observability(intent),
        "checkpoint": {
            "enabled": bool(state.get("checkpoint_enabled")),
            "namespace": str(state.get("checkpoint_namespace") or ""),
            "status": str(state.get("checkpoint_status") or "unknown"),
        },
        "trace": trace_observability(trace_context),
        "node_spans": node_spans(events),
        "circuit_breakers": {
            "model_gateway": MODEL_GATEWAY_BREAKER.snapshot(),
            "tool_gateway": TOOL_GATEWAY_BREAKER.snapshot(),
        },
    }


def tool_observability(observations: list[dict[str, Any]]) -> dict[str, Any]:
    items: list[dict[str, Any]] = []
    for item in observations:
        items.append(
            {
                "tool_key": str(item.get("tool_key") or ""),
                "status": str(item.get("status") or ""),
                "ok": bool(item.get("ok")),
                "latency_ms": int_value(item.get("latency_ms")),
                "attempt_count": max(1, int_value(item.get("attempt_count"))),
                "error_code": str(item.get("error_code") or ""),
                "audit_ref_present": bool(item.get("audit_ref")),
                "entity_ref_count": list_len(item.get("entity_refs")),
            }
        )
    return {
        "total": len(items),
        "succeeded": len([item for item in items if item["ok"]]),
        "failed": len([item for item in items if not item["ok"]]),
        "items": items,
    }


def model_observability(intent: dict[str, Any]) -> dict[str, Any]:
    usage = dict_value(intent.get("model_usage"))
    input_tokens = int_value(usage.get("input_tokens"))
    output_tokens = int_value(usage.get("output_tokens"))
    return {
        "provider_key": str(usage.get("provider_key") or ""),
        "model_id": str(usage.get("model_id") or ""),
        "input_tokens": input_tokens,
        "output_tokens": output_tokens,
        "latency_ms": int_value(usage.get("latency_ms")),
        "quota_status": str(dict_value(usage.get("metadata_json")).get("quota_status") or "unknown"),
        "estimated_cost_microusd": 0,
        "cost_status": "pricing_not_configured" if input_tokens or output_tokens else "not_applicable",
    }


def trace_observability(trace_context: dict[str, Any]) -> dict[str, Any]:
    baggage = trace_context.get("baggage")
    baggage_keys = sorted(str(key) for key in baggage) if isinstance(baggage, dict) else []
    return {
        "trace_id": str(trace_context.get("trace_id") or ""),
        "span_id": str(trace_context.get("span_id") or ""),
        "request_id": str(trace_context.get("request_id") or ""),
        "baggage_keys": baggage_keys[:20],
    }


def node_spans(events: list[dict[str, Any]]) -> list[dict[str, Any]]:
    points: list[tuple[str, int]] = []
    for event in events:
        if event.get("event_type") != "STATUS":
            continue
        message = str(event.get("message") or "")
        if not message.startswith("Шаг: "):
            continue
        points.append((message.removeprefix("Шаг: ").strip(), event_elapsed_ms(event)))
    if not points:
        return []
    end_ms = max_elapsed_ms(events)
    spans: list[dict[str, Any]] = []
    for index, (node, started_ms) in enumerate(points):
        next_ms = points[index + 1][1] if index + 1 < len(points) else end_ms
        spans.append({"node": node, "latency_ms": max(0, next_ms - started_ms)})
    return spans


def max_elapsed_ms(events: list[dict[str, Any]]) -> int:
    return max((event_elapsed_ms(event) for event in events), default=0)


def time_to_first_event_ms(events: list[dict[str, Any]]) -> int:
    for event in events:
        return event_elapsed_ms(event)
    return 0


def event_elapsed_ms(event: dict[str, Any]) -> int:
    observability = event.get("observability")
    if not isinstance(observability, dict):
        return 0
    return int_value(observability.get("elapsed_ms"))


def dict_value(value: Any) -> dict[str, Any]:
    return value if isinstance(value, dict) else {}


def list_len(value: Any) -> int:
    return len(value) if isinstance(value, list) else 0


def int_value(value: Any) -> int:
    if isinstance(value, int):
        return value
    if isinstance(value, float):
        return int(value)
    if isinstance(value, str) and value.isdigit():
        return int(value)
    return 0


def float_value(value: Any) -> float:
    if isinstance(value, int | float):
        return float(value)
    if isinstance(value, str):
        try:
            return float(value)
        except ValueError:
            return 0.0
    return 0.0
