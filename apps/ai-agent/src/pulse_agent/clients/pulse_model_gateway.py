from __future__ import annotations

import os
import sys
from collections.abc import AsyncIterator
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

import grpc
from google.protobuf import json_format
from google.protobuf.struct_pb2 import Struct

from pulse_agent.clients.grpc_security import internal_aio_channel_unmanaged
from pulse_agent.execution_trace import record_model_call

_PROTO_ROOT = Path(__file__).resolve().parents[1] / "proto"
if str(_PROTO_ROOT) not in sys.path:
    sys.path.insert(0, str(_PROTO_ROOT))

from ai_agent.v1 import common_pb2, model_gateway_pb2, model_gateway_pb2_grpc  # noqa: E402


class PulseModelGatewayError(RuntimeError):
    def __init__(self, code: str, message: str) -> None:
        self.code = code
        super().__init__(message)


def _struct(value: dict[str, Any] | None) -> Struct:
    out = Struct()
    json_format.ParseDict(value or {}, out)
    return out


def _dict(value: Struct | None) -> dict[str, Any]:
    if value is None:
        return {}
    return dict(json_format.MessageToDict(value))


def _metadata(token: str) -> tuple[tuple[str, str], ...]:
    if not token:
        return ()
    return (
        ("x-pulse-service", "pulse-ai-agent"),
        ("x-pulse-service-token", token),
    )


def _purpose(value: str) -> int:
    normalized = value.strip().lower().replace("-", "_")
    mapping = {
        "intent": model_gateway_pb2.MODEL_PURPOSE_INTENT_ROUTER,
        "intent_router": model_gateway_pb2.MODEL_PURPOSE_INTENT_ROUTER,
        "planner": model_gateway_pb2.MODEL_PURPOSE_PLANNER,
        "tool_args": model_gateway_pb2.MODEL_PURPOSE_TOOL_ARGS,
        "final": model_gateway_pb2.MODEL_PURPOSE_FINAL_RESPONSE,
        "final_response": model_gateway_pb2.MODEL_PURPOSE_FINAL_RESPONSE,
        "memory": model_gateway_pb2.MODEL_PURPOSE_MEMORY,
    }
    return int(mapping.get(normalized, model_gateway_pb2.MODEL_PURPOSE_UNSPECIFIED))


def _usage(value: Any) -> dict[str, Any]:
    if value is None:
        return {}
    return {
        "provider_key": value.provider_key,
        "model_id": value.model_id,
        "input_tokens": value.input_tokens,
        "output_tokens": value.output_tokens,
        "latency_ms": value.latency_ms,
        "metadata_json": _dict(value.metadata_json),
    }


@dataclass
class PulseModelGatewayClient:
    target: str
    service_token: str = ""
    timeout_seconds: float | None = None
    _channel: grpc.aio.Channel | None = field(default=None, init=False, repr=False)

    async def close(self) -> None:
        if self._channel is not None:
            await self._channel.close()
            self._channel = None

    async def chat(
        self,
        messages: list[dict[str, Any]],
        *,
        purpose: str,
        company_id: str = "",
        project_id: str = "",
        user_id: str = "",
        session_id: str = "",
        thread_id: str = "",
        request_id: str = "",
        response_format_json: dict[str, Any] | None = None,
        tool_schemas_json: dict[str, Any] | None = None,
        max_tokens: int = 0,
        temperature: float = 0.0,
        timeout_ms: int = 0,
        trace_context: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        request = model_gateway_pb2.ModelChatRequest(
            request_id=request_id,
            company_id=company_id,
            project_id=project_id,
            user_id=user_id,
            session_id=session_id,
            thread_id=thread_id,
            purpose=_purpose(purpose),
            messages=[_message(item) for item in messages],
            response_format_json=_struct(response_format_json),
            tool_schemas_json=_struct(tool_schemas_json),
            max_tokens=max_tokens,
            temperature=temperature,
            timeout_ms=timeout_ms,
            trace_context=_trace_context(trace_context),
        )
        try:
            response = await self._stub().Chat(
                request,
                metadata=_metadata(self.service_token),
                timeout=self._timeout(timeout_ms),
            )
        except grpc.aio.AioRpcError as exc:
            raise PulseModelGatewayError(exc.code().name.lower(), exc.details() or "model gateway error") from exc
        record_model_call(purpose, _usage(response.usage))
        return {
            "content": response.content,
            "response_json": _dict(response.response_json),
            "usage": _usage(response.usage),
        }

    async def stream_chat(
        self,
        messages: list[dict[str, Any]],
        *,
        purpose: str,
        **kwargs: Any,
    ) -> AsyncIterator[dict[str, Any]]:
        request = model_gateway_pb2.ModelChatRequest(
            purpose=_purpose(purpose),
            messages=[_message(item) for item in messages],
            request_id=str(kwargs.get("request_id") or ""),
            company_id=str(kwargs.get("company_id") or ""),
            project_id=str(kwargs.get("project_id") or ""),
            user_id=str(kwargs.get("user_id") or ""),
            session_id=str(kwargs.get("session_id") or ""),
            thread_id=str(kwargs.get("thread_id") or ""),
            response_format_json=_struct(kwargs.get("response_format_json")),
            tool_schemas_json=_struct(kwargs.get("tool_schemas_json")),
            max_tokens=int(kwargs.get("max_tokens") or 0),
            temperature=float(kwargs.get("temperature") or 0.0),
            timeout_ms=int(kwargs.get("timeout_ms") or 0),
            trace_context=_trace_context(kwargs.get("trace_context")),
        )
        try:
            stream = self._stub().StreamChat(
                request,
                metadata=_metadata(self.service_token),
                timeout=self._timeout(int(kwargs.get("timeout_ms") or 0)),
            )
            async for event in stream:
                yield {
                    "delta": event.delta,
                    "done": event.done,
                    "usage": _usage(event.usage) if event.HasField("usage") else {},
                    "error": _dict(event.error) if event.HasField("error") else {},
                }
        except grpc.aio.AioRpcError as exc:
            raise PulseModelGatewayError(exc.code().name.lower(), exc.details() or "model gateway error") from exc

    async def embed(
        self,
        input_texts: list[str],
        *,
        company_id: str = "",
        project_id: str = "",
        user_id: str = "",
        request_id: str = "",
        trace_context: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        request = model_gateway_pb2.EmbedRequest(
            request_id=request_id,
            company_id=company_id,
            project_id=project_id,
            user_id=user_id,
            input=input_texts,
            trace_context=_trace_context(trace_context),
        )
        try:
            response = await self._stub().Embed(
                request,
                metadata=_metadata(self.service_token),
                timeout=self._timeout(0),
            )
        except grpc.aio.AioRpcError as exc:
            raise PulseModelGatewayError(exc.code().name.lower(), exc.details() or "model gateway error") from exc
        return {
            "embeddings": [[value.number_value for value in item.values] for item in response.embeddings],
            "usage": _usage(response.usage),
        }

    async def intent_router(self, messages: list[dict[str, Any]], **kwargs: Any) -> dict[str, Any]:
        return await self.chat(messages, purpose="intent_router", **kwargs)

    async def planner(self, messages: list[dict[str, Any]], **kwargs: Any) -> dict[str, Any]:
        return await self.chat(messages, purpose="planner", **kwargs)

    async def tool_args(self, messages: list[dict[str, Any]], **kwargs: Any) -> dict[str, Any]:
        return await self.chat(messages, purpose="tool_args", **kwargs)

    async def final_response(self, messages: list[dict[str, Any]], **kwargs: Any) -> dict[str, Any]:
        return await self.chat(messages, purpose="final_response", **kwargs)

    async def memory(self, messages: list[dict[str, Any]], **kwargs: Any) -> dict[str, Any]:
        return await self.chat(messages, purpose="memory", **kwargs)

    def _stub(self) -> Any:
        if self._channel is None:
            self._channel = internal_aio_channel_unmanaged(
                self.target,
                self.service_token,
                service_name="Pulse ModelGateway",
                insecure_allow_nonlocal_env="PULSE_MODEL_GRPC_INSECURE_ALLOW_NONLOCAL",
                allow_remote_token_env="PULSE_MODEL_GRPC_ALLOW_REMOTE_TOKEN",
            )
        return model_gateway_pb2_grpc.ModelGatewayServiceStub(self._channel)

    def _timeout(self, timeout_ms: int) -> float | None:
        if timeout_ms > 0:
            return timeout_ms / 1000
        if self.timeout_seconds is not None:
            return self.timeout_seconds
        raw = os.environ.get("PULSE_MODEL_GRPC_TIMEOUT_SECONDS", "").strip()
        if raw.isdigit() and int(raw) > 0:
            return float(raw)
        return None


def _message(value: dict[str, Any]) -> Any:
    return model_gateway_pb2.ModelMessage(
        role=str(value.get("role") or "user"),
        content=str(value.get("content") or value.get("body") or ""),
        metadata_json=_struct(value.get("metadata_json") if isinstance(value.get("metadata_json"), dict) else {}),
    )


def _trace_context(value: dict[str, Any] | None) -> Any:
    raw = value or {}
    baggage = raw.get("baggage")
    return common_pb2.TraceContext(
        trace_id=str(raw.get("trace_id") or ""),
        span_id=str(raw.get("span_id") or ""),
        request_id=str(raw.get("request_id") or ""),
        baggage={str(k): str(v) for k, v in baggage.items()} if isinstance(baggage, dict) else {},
    )
