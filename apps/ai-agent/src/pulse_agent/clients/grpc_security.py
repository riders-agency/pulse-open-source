from __future__ import annotations

import os
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager
from ipaddress import ip_address

import grpc


def validate_internal_grpc_target(
    target: str,
    service_token: str,
    *,
    service_name: str,
    insecure_dev_env: str = "AI_AGENT_INSECURE_DEV",
    insecure_allow_nonlocal_env: str = "PULSE_TOOL_GRPC_INSECURE_ALLOW_NONLOCAL",
    allow_remote_token_env: str = "PULSE_TOOL_GRPC_ALLOW_REMOTE_TOKEN",
) -> None:
    insecure_dev = _bool_env(insecure_dev_env) or _bool_env("PULSE_TOOL_GRPC_INSECURE_DEV")
    if insecure_dev:
        if not _is_loopback_or_unix(target) and not _bool_env(insecure_allow_nonlocal_env):
            raise RuntimeError(f"{service_name} insecure dev gRPC target must be loopback or unix socket: {target}")
        return
    if not service_token.strip():
        raise RuntimeError(f"{service_name} gRPC requires AI_AGENT_SERVICE_TOKEN or explicit {insecure_dev_env}=true")
    if not _is_loopback_or_unix(target) and not _bool_env(allow_remote_token_env):
        raise RuntimeError(
            f"{service_name} token gRPC target must be loopback or unix socket unless "
            f"{allow_remote_token_env}=true: {target}"
        )


@asynccontextmanager
async def internal_aio_channel(
    target: str,
    service_token: str,
    *,
    service_name: str,
    **policy: str,
) -> AsyncIterator[grpc.aio.Channel]:
    validate_internal_grpc_target(target, service_token, service_name=service_name, **policy)
    async with grpc.aio.insecure_channel(target) as channel:
        yield channel


def internal_aio_channel_unmanaged(
    target: str,
    service_token: str,
    *,
    service_name: str,
    **policy: str,
) -> grpc.aio.Channel:
    validate_internal_grpc_target(target, service_token, service_name=service_name, **policy)
    return grpc.aio.insecure_channel(target)


def _bool_env(key: str) -> bool:
    return os.environ.get(key, "").strip().lower() in {"1", "true", "yes", "on"}


def _is_loopback_or_unix(target: str) -> bool:
    if target.startswith("unix:"):
        return True
    if "://" in target:
        return False
    host = target.strip()
    if host.startswith("[") and "]" in host:
        host = host[1:host.index("]")]
    elif ":" in host:
        host = host.rsplit(":", 1)[0]
    host = host.strip("[]")
    if host in {"", "localhost"}:
        return True
    try:
        return ip_address(host).is_loopback
    except ValueError:
        return False
