from __future__ import annotations

import time
from typing import Any


def add_event(
    events: list[dict[str, Any]],
    state: dict[str, Any],
    event_type: str,
    message: str,
    *,
    card: dict[str, Any] | None = None,
    tool_call: dict[str, Any] | None = None,
    interrupt: dict[str, Any] | None = None,
    error: dict[str, Any] | None = None,
) -> None:
    events.append(
        graph_event(
            {**state, "events": events},
            event_type,
            message,
            card=card,
            tool_call=tool_call,
            interrupt=interrupt,
            error=error,
        )
    )


def graph_event(
    state: dict[str, Any],
    event_type: str,
    message: str,
    *,
    card: dict[str, Any] | None = None,
    tool_call: dict[str, Any] | None = None,
    interrupt: dict[str, Any] | None = None,
    error: dict[str, Any] | None = None,
) -> dict[str, Any]:
    seq = len(state.get("events") or []) + 1
    return {
        "seq": seq,
        "run_id": str(state.get("run_id") or ""),
        "thread_id": str(state.get("thread_id") or ""),
        "session_id": str(state.get("session_id") or ""),
        "event_type": event_type,
        "message": message,
        "card": card or {},
        "tool_call": tool_call or {},
        "interrupt": interrupt or {},
        "error": error or {},
        "observability": event_observability(state, seq=seq),
    }


def event_observability(state: dict[str, Any], *, seq: int) -> dict[str, Any]:
    started_ms = int_value(state.get("observability_started_ms"))
    current_ms = time.monotonic_ns() // 1_000_000
    return {
        "schema": "riders-pulse-ai-event-observability-v1",
        "event_index": seq,
        "elapsed_ms": max(0, current_ms - started_ms) if started_ms > 0 else 0,
    }


def int_value(value: Any) -> int:
    if isinstance(value, int):
        return value
    if isinstance(value, float):
        return int(value)
    if isinstance(value, str) and value.isdigit():
        return int(value)
    return 0
