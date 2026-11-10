from __future__ import annotations

import os
from dataclasses import dataclass
from ipaddress import ip_address
from re import fullmatch
from urllib.parse import parse_qs, urlparse


@dataclass(frozen=True)
class Settings:
    environment: str = os.getenv("RP_ENV", "development")
    grpc_addr: str = os.getenv("AI_AGENT_GRPC_ADDR", ":9097")
    pulse_tool_grpc_addr: str = os.getenv("PULSE_TOOL_GRPC_ADDR", "127.0.0.1:9098")
    pulse_model_grpc_addr: str = os.getenv("PULSE_MODEL_GRPC_ADDR", "127.0.0.1:9099")
    database_url: str = os.getenv("DATABASE_URL", "")
    checkpoint_database_url: str = os.getenv("AI_AGENT_CHECKPOINT_DATABASE_URL", "")
    checkpoint_schema: str = os.getenv("AI_AGENT_CHECKPOINT_SCHEMA", "ai_agent_checkpoint")
    checkpoint_retention_days: str = os.getenv("AI_AGENT_CHECKPOINT_RETENTION_DAYS", "30")
    db_tls_policy: str = os.getenv("PULSE_DB_TLS_POLICY", "")
    insecure_dev: bool = os.getenv("AI_AGENT_INSECURE_DEV", "false").lower() == "true"
    insecure_bind_all: bool = os.getenv("AI_AGENT_INSECURE_BIND_ALL", "false").lower() == "true"
    service_token: str = os.getenv("AI_AGENT_SERVICE_TOKEN", os.getenv("AI_AGENT_GRPC_SERVICE_TOKEN", ""))
    run_migrations: bool = os.getenv("AI_AGENT_RUN_MIGRATIONS", "false").lower() == "true"
    strict_msgpack: bool = os.getenv("LANGGRAPH_STRICT_MSGPACK", "true").lower() == "true"
    grpc_timeout_seconds: str = os.getenv("AI_AGENT_GRPC_TIMEOUT_SECONDS", "")
    tool_grpc_timeout_seconds: str = os.getenv("PULSE_TOOL_GRPC_TIMEOUT_SECONDS", "")
    model_grpc_timeout_seconds: str = os.getenv("PULSE_MODEL_GRPC_TIMEOUT_SECONDS", "")
    service_name: str = "pulse-ai-agent"


class ConfigValidationError(RuntimeError):
    schema = "pulse.config.v1"

    def __init__(self, problems: list[tuple[str, str]]) -> None:
        self.problems = problems
        rendered = "; ".join(f"{key}: {message}" for key, message in problems)
        super().__init__(f"production config invalid schema={self.schema} role=ai-agent: {rendered}")


def load_settings() -> Settings:
    return Settings()


def validate_settings(settings: Settings) -> None:
    if settings.environment.lower() != "production":
        return
    problems: list[tuple[str, str]] = []
    if settings.insecure_dev:
        problems.append(("AI_AGENT_INSECURE_DEV", "must be false in production"))
    if settings.insecure_bind_all:
        problems.append(("AI_AGENT_INSECURE_BIND_ALL", "must be false in production"))
    if _is_placeholder(settings.service_token) or len(settings.service_token) < 32:
        problems.append(("AI_AGENT_SERVICE_TOKEN", "required and must be at least 32 characters"))
    if not _is_loopback_or_unix(_bind_addr(settings.grpc_addr)):
        problems.append(("AI_AGENT_GRPC_ADDR", "must bind to loopback or unix socket in production"))
    if _is_placeholder(settings.checkpoint_database_url):
        problems.append(("AI_AGENT_CHECKPOINT_DATABASE_URL", "required in production"))
    else:
        _validate_database_tls(
            settings.checkpoint_database_url,
            settings.db_tls_policy,
            problems,
            "AI_AGENT_CHECKPOINT_DATABASE_URL",
        )
    if settings.database_url:
        _validate_database_tls(settings.database_url, settings.db_tls_policy, problems, "DATABASE_URL")
    if settings.checkpoint_schema in {"", "public"} or not _valid_schema_name(settings.checkpoint_schema):
        problems.append(("AI_AGENT_CHECKPOINT_SCHEMA", "must be a non-public PostgreSQL identifier"))
    if not _positive_int(settings.checkpoint_retention_days):
        problems.append(("AI_AGENT_CHECKPOINT_RETENTION_DAYS", "must be a positive integer"))
    for key, value in (
        ("AI_AGENT_GRPC_TIMEOUT_SECONDS", settings.grpc_timeout_seconds),
        ("PULSE_TOOL_GRPC_TIMEOUT_SECONDS", settings.tool_grpc_timeout_seconds),
        ("PULSE_MODEL_GRPC_TIMEOUT_SECONDS", settings.model_grpc_timeout_seconds),
    ):
        if not _positive_int(value):
            problems.append((key, "must be a positive integer"))
    if problems:
        raise ConfigValidationError(problems)


def _validate_database_tls(database_url: str, db_tls_policy: str, problems: list[tuple[str, str]], key: str) -> None:
    if db_tls_policy not in {"local-only", "verify-ca", "verify-full"}:
        problems.append(("PULSE_DB_TLS_POLICY", "required: local-only, verify-ca or verify-full"))
        return
    parsed = urlparse(database_url)
    query = parse_qs(parsed.query)
    sslmode = query.get("sslmode", [""])[0]
    if db_tls_policy == "local-only":
        if sslmode != "disable" or not _is_local_host(parsed.hostname or ""):
            problems.append((key, "local-only requires loopback host and sslmode=disable"))
        return
    if sslmode != db_tls_policy:
        problems.append((key, "TLS policy requires matching sslmode"))


def _bind_addr(grpc_addr: str) -> str:
    if grpc_addr.startswith(":"):
        return "127.0.0.1" + grpc_addr
    return grpc_addr


def _is_loopback_or_unix(bind_addr: str) -> bool:
    if bind_addr.startswith("unix:"):
        return True
    host = bind_addr
    if host.startswith("[") and "]" in host:
        host = host[1 : host.index("]")]
    elif ":" in host:
        host = host.rsplit(":", 1)[0]
    host = host.strip()
    return _is_local_host(host)


def _is_local_host(host: str) -> bool:
    if host in {"", "localhost"}:
        return True
    try:
        return ip_address(host).is_loopback
    except ValueError:
        return False


def _positive_int(value: str) -> bool:
    try:
        return int(value) > 0
    except ValueError:
        return False


def _valid_schema_name(value: str) -> bool:
    return fullmatch(r"[a-z_][a-z0-9_]{0,62}", value) is not None


def _is_placeholder(value: str) -> bool:
    normalized = value.strip().lower()
    return normalized.startswith("replace-with") or normalized in {
        "",
        "change-me",
        "changeme",
        "replace-me",
        "password",
        "secret",
        "demo",
        "dev",
    }
