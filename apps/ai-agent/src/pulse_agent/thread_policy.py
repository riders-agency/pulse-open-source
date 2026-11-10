from __future__ import annotations

import hashlib
from typing import Any, cast

from pulse_agent.graph.intent_schema import TurnKind, classify_turn

MUTATING_TURN_KINDS = {
    TurnKind.APPROVAL_DECISION,
    TurnKind.MULTI_STEP,
    TurnKind.PROPOSAL_EDIT,
    TurnKind.WORKFLOW_PATCH,
    TurnKind.WRITE_INTENT,
}


def expected_thread_revision(*payloads: dict[str, Any]) -> int:
    for payload in payloads:
        for key in ("expectedThreadRevision", "expected_thread_revision", "threadRevision", "thread_revision"):
            value = int_value(payload.get(key))
            if value > 0:
                return value
    return 0


def is_mutating_turn(text: str, *, active_tool_call_id: str = "", page_context: dict[str, Any] | None = None) -> bool:
    classification = classify_turn(text, active_tool_call_id=active_tool_call_id, page_context=page_context or {})
    return classification.turn_kind in MUTATING_TURN_KINDS


def turn_idempotency_key(request_id: str, idempotency_key: str, thread_id: str, body: str) -> str:
    if idempotency_key:
        return idempotency_key
    if request_id:
        return request_id
    digest = hashlib.sha256(f"{thread_id}\0{body}".encode()).hexdigest()[:24]
    return f"turn:{digest}"


def stable_turn_id(thread_id: str, key: str) -> str:
    digest = hashlib.sha256(f"{thread_id}\0{key}".encode()).hexdigest()[:24]
    return f"turn:{digest}"


def duplicate_turn_result(values: dict[str, Any], key: str) -> dict[str, Any] | None:
    processed = values.get("processed_turns")
    if isinstance(processed, dict) and isinstance(processed.get(key), dict):
        return cast(dict[str, Any], processed[key])
    if str(values.get("last_turn_idempotency_key") or "") == key:
        result = values.get("last_turn_result")
        return result if isinstance(result, dict) else {"thread_revision": int_value(values.get("thread_revision"))}
    return None


def remember_turn(values: dict[str, Any], key: str, result: dict[str, Any]) -> dict[str, Any]:
    processed = values.get("processed_turns")
    out = dict(processed) if isinstance(processed, dict) else {}
    out[key] = result
    if len(out) > 20:
        for stale_key in list(out)[: len(out) - 20]:
            out.pop(stale_key, None)
    return out


def latest_thread_state(values: dict[str, Any], next_nodes: list[str]) -> dict[str, Any]:
    return {
        "thread_revision": int_value(values.get("thread_revision")),
        "current_step": str(values.get("current_step") or ""),
        "active_mode": str(values.get("active_mode") or ""),
        "active_card_id": str(values.get("active_card_id") or ""),
        "active_tool_call_id": str(values.get("active_tool_call_id") or ""),
        "pending_interrupt": bool(values.get("pending_interrupt")),
        "next_nodes": next_nodes,
    }


def int_value(value: Any) -> int:
    if isinstance(value, int):
        return value
    if isinstance(value, float):
        return int(value)
    if isinstance(value, str) and value.isdigit():
        return int(value)
    return 0
