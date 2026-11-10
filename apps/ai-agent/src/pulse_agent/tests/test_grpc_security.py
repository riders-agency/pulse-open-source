from __future__ import annotations

import pytest

from pulse_agent.clients.grpc_security import validate_internal_grpc_target
from pulse_agent.config import Settings
from pulse_agent.grpc_server import _bind_addr, _validate_security


def test_production_port_bind_defaults_to_loopback() -> None:
    settings = Settings(grpc_addr=":9097", service_token="service-token")

    assert _bind_addr(settings) == "127.0.0.1:9097"
    _validate_security(settings)


def test_production_remote_bind_requires_explicit_override(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("AI_AGENT_GRPC_ALLOW_REMOTE_TOKEN", raising=False)
    settings = Settings(grpc_addr="0.0.0.0:9097", service_token="service-token")

    with pytest.raises(RuntimeError, match="loopback or unix socket"):
        _validate_security(settings)


def test_insecure_dev_bind_all_is_explicit() -> None:
    settings = Settings(grpc_addr=":9097", insecure_dev=True, insecure_bind_all=True)

    assert _bind_addr(settings) == "[::]:9097"
    _validate_security(settings)


def test_tool_gateway_requires_token_without_insecure_dev(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("AI_AGENT_INSECURE_DEV", raising=False)
    monkeypatch.delenv("PULSE_TOOL_GRPC_INSECURE_DEV", raising=False)

    with pytest.raises(RuntimeError, match="requires AI_AGENT_SERVICE_TOKEN"):
        validate_internal_grpc_target("127.0.0.1:9098", "", service_name="Pulse ToolGateway")


def test_tool_gateway_token_allows_loopback(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("AI_AGENT_INSECURE_DEV", raising=False)
    monkeypatch.delenv("PULSE_TOOL_GRPC_INSECURE_DEV", raising=False)

    validate_internal_grpc_target("127.0.0.1:9098", "service-token", service_name="Pulse ToolGateway")


def test_tool_gateway_remote_token_requires_explicit_override(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("AI_AGENT_INSECURE_DEV", raising=False)
    monkeypatch.delenv("PULSE_TOOL_GRPC_ALLOW_REMOTE_TOKEN", raising=False)

    with pytest.raises(RuntimeError, match="loopback or unix socket"):
        validate_internal_grpc_target("10.0.0.10:9098", "service-token", service_name="Pulse ToolGateway")
