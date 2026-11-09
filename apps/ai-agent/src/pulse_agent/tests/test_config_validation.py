from __future__ import annotations

from dataclasses import replace

import pytest

from pulse_agent.config import ConfigValidationError, Settings, validate_settings


def test_validate_settings_skips_development() -> None:
    validate_settings(Settings(environment="development", insecure_dev=True))


def test_validate_settings_accepts_production_fixture() -> None:
    validate_settings(valid_settings())


def valid_settings() -> Settings:
    return Settings(
        environment="production",
        grpc_addr="127.0.0.1:9097",
        pulse_tool_grpc_addr="127.0.0.1:9098",
        pulse_model_grpc_addr="127.0.0.1:9099",
        database_url="postgres://pulse@127.0.0.1:5432/pulse?sslmode=disable",
        checkpoint_database_url="postgres://pulse_ai@127.0.0.1:5432/pulse_ai?sslmode=disable",
        checkpoint_schema="ai_agent_checkpoint",
        checkpoint_retention_days="30",
        db_tls_policy="local-only",
        insecure_dev=False,
        insecure_bind_all=False,
        service_token="a" * 40,
        grpc_timeout_seconds="10",
        tool_grpc_timeout_seconds="10",
        model_grpc_timeout_seconds="20",
    )


@pytest.mark.parametrize(
    ("settings", "key"),
    (
        (replace(valid_settings(), service_token="change-me"), "AI_AGENT_SERVICE_TOKEN"),
        (replace(valid_settings(), insecure_dev=True), "AI_AGENT_INSECURE_DEV"),
        (replace(valid_settings(), insecure_bind_all=True), "AI_AGENT_INSECURE_BIND_ALL"),
        (replace(valid_settings(), grpc_addr="0.0.0.0:9097"), "AI_AGENT_GRPC_ADDR"),
        (replace(valid_settings(), db_tls_policy="verify-full"), "DATABASE_URL"),
        (replace(valid_settings(), checkpoint_database_url=""), "AI_AGENT_CHECKPOINT_DATABASE_URL"),
        (replace(valid_settings(), checkpoint_schema="public"), "AI_AGENT_CHECKPOINT_SCHEMA"),
        (replace(valid_settings(), checkpoint_retention_days="0"), "AI_AGENT_CHECKPOINT_RETENTION_DAYS"),
        (replace(valid_settings(), grpc_timeout_seconds=""), "AI_AGENT_GRPC_TIMEOUT_SECONDS"),
        (replace(valid_settings(), tool_grpc_timeout_seconds="0"), "PULSE_TOOL_GRPC_TIMEOUT_SECONDS"),
        (replace(valid_settings(), model_grpc_timeout_seconds="nope"), "PULSE_MODEL_GRPC_TIMEOUT_SECONDS"),
    ),
)
def test_validate_settings_rejects_negative_fixtures(settings: Settings, key: str) -> None:
    with pytest.raises(ConfigValidationError) as exc:
        validate_settings(settings)

    message = str(exc.value)
    assert key in message
    assert settings.service_token not in message
