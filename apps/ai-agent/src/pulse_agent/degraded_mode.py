from __future__ import annotations

import os
import time
from dataclasses import dataclass
from typing import Any

from pulse_agent.config import Settings


@dataclass
class CircuitBreaker:
    failure_threshold: int = 2
    recovery_seconds: int = 30
    failures: int = 0
    opened_at: float = 0.0

    def allow_request(self, now: float | None = None) -> bool:
        current = time.time() if now is None else now
        if self.opened_at <= 0:
            return True
        return current - self.opened_at >= self.recovery_seconds

    def record_success(self) -> None:
        self.failures = 0
        self.opened_at = 0.0

    def record_failure(self, now: float | None = None) -> None:
        self.failures += 1
        if self.failures >= self.failure_threshold:
            self.opened_at = time.time() if now is None else now

    def snapshot(self) -> dict[str, Any]:
        state = "open" if self.opened_at > 0 and not self.allow_request() else "closed"
        return {"state": state, "failures": self.failures, "recovery_seconds": self.recovery_seconds}


MODEL_GATEWAY_BREAKER = CircuitBreaker()
TOOL_GATEWAY_BREAKER = CircuitBreaker()


def dependency_capability_matrix(settings: Settings, *, checkpoint_enabled: bool) -> dict[str, Any]:
    model_ready = bool_env("PULSE_MODEL_GATEWAY_READY") or not bool_env("PULSE_AI_MODEL_TURN_UNDERSTANDING")
    tool_loop_enabled = bool_env("PULSE_AI_TOOL_GATEWAY_GRAPH_LOOP")
    tool_ready = tool_loop_enabled and bool(settings.service_token)
    provider_ready = bool_env("PULSE_MODEL_PROVIDER_READY") or bool_env("PULSE_MODEL_GATEWAY_READY")
    dependencies = {
        "checkpointer": dependency_state(checkpoint_enabled, "checkpoint_disabled_memory_only"),
        "model_gateway": dependency_state(model_ready, "model_gateway_not_ready"),
        "tool_gateway": dependency_state(tool_ready, "tool_gateway_graph_loop_not_ready"),
        "provider": dependency_state(
            provider_ready or not bool_env("PULSE_AI_MODEL_TURN_UNDERSTANDING"),
            "provider_not_ready",
        ),
    }
    ready = all(item["status"] == "ready" for item in dependencies.values() if item["required"])
    return {
        "schema": "riders-pulse-ai-capability-matrix-v1",
        "status": "ready" if ready else "degraded",
        "dependencies": dependencies,
        "circuit_breakers": {
            "model_gateway": MODEL_GATEWAY_BREAKER.snapshot(),
            "tool_gateway": TOOL_GATEWAY_BREAKER.snapshot(),
        },
        "recovery_action": "Check AI runtime readiness and retry after dependency recovery.",
    }


def dependency_state(ready: bool, reason: str) -> dict[str, Any]:
    return {"required": True, "status": "ready" if ready else "degraded", "reason": "" if ready else reason}


def tool_gateway_failure_observation(tool_key: str, error_code: str, message: str) -> dict[str, Any]:
    return {
        "ok": False,
        "status": "failed",
        "tool_key": tool_key,
        "output_json": {},
        "error_code": error_code,
        "error_message": message,
        "user_message": "AI runtime работает в ограниченном режиме. Повторите запрос после восстановления ToolGateway.",
        "audit_event_id": "",
    }


def bool_env(key: str) -> bool:
    return os.environ.get(key, "").strip().lower() in {"1", "true", "yes", "on"}
