from __future__ import annotations

import asyncio
import hashlib
import json
import os
import time
from typing import Any, cast

from pulse_agent.clients.pulse_tools import PulseToolsClient
from pulse_agent.config import Settings
from pulse_agent.degraded_mode import TOOL_GATEWAY_BREAKER, tool_gateway_failure_observation


def graph_tool_gateway_enabled() -> bool:
    return os.environ.get("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP", "").strip().lower() in {"1", "true", "yes", "on"}


def run_read_tool_if_enabled(
    state: dict[str, Any],
    tool_key: str,
    input_json: dict[str, Any],
) -> dict[str, Any] | None:
    if not graph_tool_gateway_enabled():
        return None
    started = time.monotonic()
    if not TOOL_GATEWAY_BREAKER.allow_request():
        observation = tool_gateway_failure_observation(
            tool_key,
            "tool_gateway_circuit_open",
            "ToolGateway circuit is open",
        )
        observation["latency_ms"] = elapsed_ms(started)
        observation["attempt_count"] = 1
        return observation
    settings = Settings()
    tool_call_id = stable_tool_call_id(state, tool_key, input_json)
    try:
        result = _run(
            PulseToolsClient(settings.pulse_tool_grpc_addr, settings.service_token).run_read_tool(
                company_id=str(state.get("company_id") or ""),
                project_id=str(state.get("project_id") or ""),
                user_id=str(state.get("user_id") or ""),
                session_id=str(state.get("session_id") or ""),
                thread_id=str(state.get("thread_id") or ""),
                tool_call_id=tool_call_id,
                tool_key=tool_key,
                input_json=input_json,
                idempotency_key=tool_call_id,
            )
        )
    except Exception as exc:
        TOOL_GATEWAY_BREAKER.record_failure()
        observation = tool_gateway_failure_observation(tool_key, "tool_gateway_error", type(exc).__name__)
        observation["latency_ms"] = elapsed_ms(started)
        observation["attempt_count"] = 1
        return observation
    TOOL_GATEWAY_BREAKER.record_success()
    result["latency_ms"] = elapsed_ms(started)
    result["attempt_count"] = 1
    return cast(dict[str, Any], result)


def execute_approval_if_enabled(
    state: dict[str, Any],
    decision: dict[str, Any],
    *,
    approve: bool,
) -> dict[str, Any] | None:
    if not graph_tool_gateway_enabled():
        return None
    started = time.monotonic()
    if not TOOL_GATEWAY_BREAKER.allow_request():
        observation = tool_gateway_failure_observation(
            "approval.apply" if approve else "approval.reject",
            "tool_gateway_circuit_open",
            "ToolGateway circuit is open",
        )
        observation["latency_ms"] = elapsed_ms(started)
        observation["attempt_count"] = 1
        return observation
    settings = Settings()
    raw_active_card = state.get("active_card")
    active_card: dict[str, Any] = raw_active_card if isinstance(raw_active_card, dict) else {}
    tool_call_id = str(decision.get("active_tool_call_id") or state.get("active_tool_call_id") or "")
    card_id = str(decision.get("card_id") or active_card.get("card_id") or "")
    card_revision = int_value(decision.get("card_revision") or active_card.get("revision"))
    client = PulseToolsClient(settings.pulse_tool_grpc_addr, settings.service_token)
    company_id = str(state.get("company_id") or "")
    project_id = str(state.get("project_id") or "")
    user_id = str(state.get("user_id") or "")
    session_id = str(state.get("session_id") or "")
    thread_id = str(state.get("thread_id") or "")
    idempotency_key = str(decision.get("idempotency_key") or decision.get("request_id") or tool_call_id)
    try:
        if approve:
            result = _run(
                client.approve_proposal(
                    company_id=company_id,
                    project_id=project_id,
                    user_id=user_id,
                    session_id=session_id,
                    thread_id=thread_id,
                    tool_call_id=tool_call_id,
                    card_id=card_id,
                    card_revision=card_revision,
                    idempotency_key=idempotency_key,
                )
            )
        else:
            result = _run(
                client.reject_proposal(
                    company_id=company_id,
                    project_id=project_id,
                    user_id=user_id,
                    session_id=session_id,
                    thread_id=thread_id,
                    tool_call_id=tool_call_id,
                    card_id=card_id,
                    card_revision=card_revision,
                    idempotency_key=idempotency_key,
                )
            )
    except Exception as exc:
        TOOL_GATEWAY_BREAKER.record_failure()
        observation = tool_gateway_failure_observation(
            "approval.apply" if approve else "approval.reject",
            "tool_gateway_error",
            type(exc).__name__,
        )
        observation["latency_ms"] = elapsed_ms(started)
        observation["attempt_count"] = 1
        return observation
    TOOL_GATEWAY_BREAKER.record_success()
    result["latency_ms"] = elapsed_ms(started)
    result["attempt_count"] = 1
    return cast(dict[str, Any], result)


def stable_tool_call_id(state: dict[str, Any], tool_key: str, input_json: dict[str, Any]) -> str:
    raw = json.dumps(input_json, sort_keys=True, ensure_ascii=False, separators=(",", ":"))
    digest = hashlib.sha256(f"{state.get('run_id')}:{tool_key}:{raw}".encode()).hexdigest()[:24]
    return f"langgraph-read:{tool_key}:{digest}"


def output_json(observation: dict[str, Any] | None) -> dict[str, Any]:
    if not isinstance(observation, dict):
        return {}
    value = observation.get("output_json")
    return value if isinstance(value, dict) else {}


def int_value(value: Any) -> int:
    if isinstance(value, int):
        return value
    if isinstance(value, float):
        return int(value)
    if isinstance(value, str) and value.isdigit():
        return int(value)
    return 0


def _run(awaitable: Any) -> Any:
    try:
        asyncio.get_running_loop()
    except RuntimeError:
        return asyncio.run(awaitable)
    raise RuntimeError("Pulse ToolGateway graph loop cannot run inside an active asyncio loop")


def elapsed_ms(started: float) -> int:
    return max(0, int((time.monotonic() - started) * 1000))
