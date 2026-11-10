"""Bounded execution evidence. No prompts, credentials or model responses are retained."""
from __future__ import annotations

from collections.abc import Callable
from contextvars import ContextVar
from time import monotonic_ns
from typing import Any

_calls: ContextVar[list[dict[str, Any]] | None] = ContextVar("pulse_model_calls", default=None)


def record_model_call(purpose: str, usage: dict[str, Any]) -> None:
    calls = _calls.get()
    if calls is None:
        return
    calls.append({
        "purpose": purpose,
        "provider": str(usage.get("provider_key") or ""),
        "model": str(usage.get("model_id") or ""),
        "input_tokens": max(0, int(usage.get("input_tokens") or 0)),
        "output_tokens": max(0, int(usage.get("output_tokens") or 0)),
        "latency_ms": max(0, int(usage.get("latency_ms") or 0)),
    })


def traced_node(name: str, operation: Callable[[dict[str, Any]], dict[str, Any]]) -> Callable[..., Any]:
    def run(state: dict[str, Any]) -> dict[str, Any]:
        calls: list[dict[str, Any]] = []
        token = _calls.set(calls)
        started = monotonic_ns()
        try:
            result = operation(state)
        finally:
            _calls.reset(token)
        trace = list(state.get("node_trace") or [])
        trace.append({"node": name, "status": "completed", "latency_ms": (monotonic_ns() - started) // 1_000_000})
        recorded = list(state.get("model_calls") or [])
        recorded.extend({**call, "node": name} for call in calls)
        return {**result, "node_trace": trace[-64:], "model_calls": recorded[-32:]}
    return run


def execution_summary(result: dict[str, Any]) -> dict[str, Any]:
    nodes = list(result.get("node_trace") or [])
    waiting = bool(result.get("__interrupt__"))
    if waiting:
        nodes.append({"node": "await_approval", "status": "waiting", "latency_ms": 0})
    calls = list(result.get("model_calls") or [])
    tools = []
    start = max(0, int(result.get("execution_event_start") or 0))
    for event in (result.get("events") or [])[start:]:
        call = event.get("tool_call") or {}
        if call.get("tool_key"):
            tools.append({"tool": str(call["tool_key"]), "status": str(call.get("status") or ""),
                          "event": str(event.get("event_type") or "")})
    return {
        "schema": "pulse-execution-v1", "nodes": nodes[-64:], "model_calls": calls[-32:], "tools": tools[-64:],
        "status": "waiting_approval" if waiting else "completed",
        "input_tokens": sum(int(call.get("input_tokens") or 0) for call in calls),
        "output_tokens": sum(int(call.get("output_tokens") or 0) for call in calls),
        "model_call_count": len(calls), "usage_source": "provider_response",
    }
