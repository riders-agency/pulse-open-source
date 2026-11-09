from __future__ import annotations

from typing import Any

from pulse_agent.graph.events import add_event
from pulse_agent.graph.tool_gateway import output_json

OBSERVATION_SCHEMA = "riders-pulse-tool-observation-v1"
MAX_OBSERVATIONS_PER_TURN = 4
MAX_OBSERVATIONS_IN_STATE = 20
SENSITIVE_KEYS = {
    "authorization",
    "cookie",
    "debug",
    "password",
    "provider_payload",
    "raw",
    "raw_json",
    "secret",
    "token",
}


def normalize_tool_observation(
    tool_key: str,
    raw_observation: dict[str, Any],
    input_json: dict[str, Any],
    *,
    step_index: int,
    run_id: str = "",
    turn_id: str = "",
) -> dict[str, Any]:
    normalized_output = sanitize_json(output_json(raw_observation))
    status = str(raw_observation.get("status") or "").strip().lower()
    ok = bool(raw_observation.get("ok", status in {"succeeded", "success"}))
    if not status:
        status = "succeeded" if ok else "failed"
    return {
        "schema": OBSERVATION_SCHEMA,
        "tool_call_id": str(raw_observation.get("tool_call_id") or ""),
        "tool_key": tool_key,
        "step_index": step_index,
        "run_id": run_id,
        "turn_id": turn_id,
        "status": status,
        "ok": ok,
        "latency_ms": int_value(raw_observation.get("latency_ms")),
        "attempt_count": max(1, int_value(raw_observation.get("attempt_count"))),
        "input_summary": summarize_input(input_json),
        "normalized_output": normalized_output,
        "entity_refs": extract_entity_refs(tool_key, normalized_output),
        "error_code": str(raw_observation.get("error_code") or ""),
        "error_message": str(raw_observation.get("error_message") or ""),
        "user_message": str(raw_observation.get("user_message") or ""),
        "audit_ref": str(raw_observation.get("audit_event_id") or raw_observation.get("audit_ref") or ""),
    }


def append_tool_observation(
    observations: list[dict[str, Any]],
    observation: dict[str, Any],
) -> list[dict[str, Any]]:
    next_observations = [item for item in observations if isinstance(item, dict)]
    next_observations.append(observation)
    return next_observations[-MAX_OBSERVATIONS_IN_STATE:]


def next_tool_input(
    base_input: dict[str, Any],
    tool_key: str,
    observations: list[dict[str, Any]],
) -> dict[str, Any]:
    enriched = dict(base_input)
    if tool_key == "payment.lookup_due":
        client = latest_entity_ref(observations, "client")
        if client:
            enriched["client_id"] = client.get("entity_id") or ""
            enriched["client_name"] = client.get("label") or client.get("query") or ""
            enriched["resolved_entities"] = [client]
    return enriched


def record_tool_result_event(
    events: list[dict[str, Any]],
    state: dict[str, Any],
    observation: dict[str, Any],
    default_message: str,
    *,
    risk_level: str = "read",
) -> None:
    add_event(
        events,
        state,
        "TOOL_RESULT",
        tool_observation_message(observation, default_message),
        tool_call=tool_event_call(observation, risk_level=risk_level),
    )


def tool_event_call(observation: dict[str, Any], *, risk_level: str = "read") -> dict[str, Any]:
    normalized_output = observation.get("normalized_output")
    output = normalized_output if isinstance(normalized_output, dict) else {}
    return {
        "tool_key": str(observation.get("tool_key") or ""),
        "risk_level": risk_level,
        "status": str(observation.get("status") or "succeeded"),
        "output_json": output,
        "entity_refs": observation.get("entity_refs") if isinstance(observation.get("entity_refs"), list) else [],
        "audit_event_id": str(observation.get("audit_ref") or ""),
        "observation_schema": OBSERVATION_SCHEMA,
    }


def summarize_observations(observations: list[dict[str, Any]], fallback: str) -> str:
    current_turn = [item for item in observations if item.get("schema") == OBSERVATION_SCHEMA]
    if not current_turn:
        return fallback
    failures = [item for item in current_turn if not item.get("ok")]
    if failures:
        failure = failures[-1]
        return str(
            failure.get("user_message")
            or failure.get("error_message")
            or failure.get("error_code")
            or "Не удалось получить данные через tool."
        )
    answers: list[str] = []
    for item in current_turn:
        if item.get("supporting_only"):
            continue
        output = item.get("normalized_output")
        if not isinstance(output, dict):
            continue
        answer = str(output.get("answer") or output.get("message") or "").strip()
        if answer:
            answers.append(answer)
    if not answers:
        user_message = str(current_turn[-1].get("user_message") or "").strip()
        return user_message or fallback
    if len(answers) == 1:
        return answers[0]
    return "\n".join(f"- {answer}" for answer in answers)


def replan_trace_entry(tool_key: str, observations: list[dict[str, Any]], next_key: str = "") -> dict[str, Any]:
    return {
        "schema": "riders-pulse-replan-trace-v1",
        "after_tool": tool_key,
        "observation_count": len(observations),
        "decision": "continue" if next_key else "finalize",
        "next_tool": next_key,
    }


def tool_observation_message(observation: dict[str, Any], default_message: str) -> str:
    output = observation.get("normalized_output")
    if isinstance(output, dict):
        message = str(output.get("answer") or output.get("message") or "").strip()
        if message:
            return message
    return str(observation.get("user_message") or default_message)


def latest_entity_ref(observations: list[dict[str, Any]], entity_type: str) -> dict[str, Any] | None:
    for observation in reversed(observations):
        refs = observation.get("entity_refs")
        if not isinstance(refs, list):
            continue
        for ref in refs:
            if isinstance(ref, dict) and ref.get("entity_type") == entity_type:
                return ref
    return None


def extract_entity_refs(tool_key: str, output: dict[str, Any]) -> list[dict[str, Any]]:
    refs: list[dict[str, Any]] = []
    items = output.get("items")
    if isinstance(items, list):
        for item in items[:5]:
            if not isinstance(item, dict):
                continue
            ref = entity_ref_from_item(tool_key, item)
            if ref:
                refs.append(ref)
    if not refs:
        ref = entity_ref_from_item(tool_key, output)
        if ref:
            refs.append(ref)
    return refs


def entity_ref_from_item(tool_key: str, item: dict[str, Any]) -> dict[str, Any] | None:
    entity_type = entity_type_for_tool(tool_key)
    entity_id = first_string(item, "id", "client_id", "task_id", "entity_id")
    label = first_string(item, "name", "title", "label", "client_name")
    query = first_string(item, "query")
    if not entity_id and not label and not query:
        return None
    return {
        "entity_type": entity_type,
        "entity_id": entity_id,
        "label": label,
        "query": query,
        "source": tool_key,
    }


def entity_type_for_tool(tool_key: str) -> str:
    if tool_key.startswith("client."):
        return "client"
    if tool_key.startswith("task."):
        return "task"
    if tool_key.startswith("payment."):
        return "payment"
    if tool_key.startswith("workflow."):
        return "workflow"
    return "entity"


def summarize_input(input_json: dict[str, Any]) -> dict[str, Any]:
    summary: dict[str, Any] = {}
    for key in ("query", "intent", "client_id", "client_name", "workflow_id"):
        value = input_json.get(key)
        if isinstance(value, str) and value:
            summary[key] = value[:240]
    resolved = input_json.get("resolved_entities")
    if isinstance(resolved, list):
        summary["resolved_entities"] = sanitize_json(resolved[:3])
    return summary


def sanitize_json(value: Any, *, depth: int = 0) -> Any:
    if depth > 4:
        return "[truncated]"
    if isinstance(value, dict):
        sanitized: dict[str, Any] = {}
        for key, item in list(value.items())[:40]:
            key_str = str(key)
            if key_str.lower() in SENSITIVE_KEYS:
                sanitized[key_str] = "[redacted]"
                continue
            sanitized[key_str] = sanitize_json(item, depth=depth + 1)
        return sanitized
    if isinstance(value, list):
        return [sanitize_json(item, depth=depth + 1) for item in value[:25]]
    if isinstance(value, str):
        return value[:2000]
    if isinstance(value, int | float | bool) or value is None:
        return value
    return str(value)[:500]


def first_string(item: dict[str, Any], *keys: str) -> str:
    for key in keys:
        value = item.get(key)
        if isinstance(value, str) and value:
            return value
    return ""


def int_value(value: Any) -> int:
    if isinstance(value, int):
        return value
    if isinstance(value, float):
        return int(value)
    if isinstance(value, str) and value.isdigit():
        return int(value)
    return 0
