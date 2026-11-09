from __future__ import annotations

import sys
from pathlib import Path
from typing import Any

import grpc
import pytest
from google.protobuf import json_format
from google.protobuf.struct_pb2 import ListValue, Struct

_PROTO_ROOT = Path(__file__).resolve().parents[1] / "proto"
if str(_PROTO_ROOT) not in sys.path:
    sys.path.insert(0, str(_PROTO_ROOT))

from ai_agent.v1 import common_pb2, model_gateway_pb2, model_gateway_pb2_grpc  # noqa: E402

from pulse_agent.clients.pulse_model_gateway import PulseModelGatewayClient, PulseModelGatewayError  # noqa: E402


class FakeModelGateway(model_gateway_pb2_grpc.ModelGatewayServiceServicer):
    def __init__(self) -> None:
        self.chat_requests: list[model_gateway_pb2.ModelChatRequest] = []
        self.metadata: list[dict[str, str]] = []

    async def Chat(
        self,
        request: model_gateway_pb2.ModelChatRequest,
        context: grpc.aio.ServicerContext,
    ) -> model_gateway_pb2.ModelChatResponse:
        self.chat_requests.append(request)
        self.metadata.append(dict(context.invocation_metadata()))
        if request.request_id == "fail":
            await context.abort(grpc.StatusCode.FAILED_PRECONDITION, "model_schema_invalid")
        return model_gateway_pb2.ModelChatResponse(
            content='{"answer":"ok"}',
            response_json=_struct({"answer": "ok"}),
            usage=common_pb2.Usage(
                provider_key="fake",
                model_id="fake-model",
                input_tokens=3,
                output_tokens=4,
                latency_ms=5,
                metadata_json=_struct({"provider": "fake"}),
            ),
        )

    async def StreamChat(
        self,
        request: model_gateway_pb2.ModelChatRequest,
        context: grpc.aio.ServicerContext,
    ) -> Any:
        del request, context
        yield model_gateway_pb2.ModelChatEvent(delta="hello")
        yield model_gateway_pb2.ModelChatEvent(done=True, usage=common_pb2.Usage(provider_key="fake"))

    async def Embed(
        self,
        request: model_gateway_pb2.EmbedRequest,
        context: grpc.aio.ServicerContext,
    ) -> model_gateway_pb2.EmbedResponse:
        del context
        return model_gateway_pb2.EmbedResponse(
            embeddings=[_list([float(len(item)), 0.5]) for item in request.input],
            usage=common_pb2.Usage(provider_key="fake", model_id="fake-embed", input_tokens=len(request.input)),
        )


async def test_chat_maps_request_response_and_reuses_channel() -> None:
    target, server, fake = await _start_fake_server()
    client = PulseModelGatewayClient(target, service_token="secret", timeout_seconds=2)
    try:
        first = await client.chat(
            [{"role": "system", "content": "Return JSON"}, {"role": "user", "content": "Ping"}],
            purpose="intent_router",
            company_id="company-1",
            project_id="project-1",
            user_id="user-1",
            session_id="session-1",
            thread_id="thread-1",
            request_id="req-1",
            response_format_json={"type": "json_object"},
            max_tokens=50,
            temperature=0.1,
            timeout_ms=1200,
            trace_context={"trace_id": "trace-1", "baggage": {"tenant": "test"}},
        )
        channel = client._channel
        second = await client.final_response([{"role": "user", "content": "Again"}], request_id="req-2")
    finally:
        await client.close()
        await server.stop(0)

    assert first["response_json"] == {"answer": "ok"}
    assert first["usage"]["provider_key"] == "fake"
    assert second["content"] == '{"answer":"ok"}'
    assert channel is not None and client._channel is None
    request = fake.chat_requests[0]
    assert request.purpose == model_gateway_pb2.MODEL_PURPOSE_INTENT_ROUTER
    assert request.company_id == "company-1"
    assert request.messages[1].content == "Ping"
    assert request.response_format_json.fields["type"].string_value == "json_object"
    assert request.trace_context.trace_id == "trace-1"
    assert fake.metadata[0]["x-pulse-service-token"] == "secret"


async def test_stream_chat_and_embed() -> None:
    target, server, _fake = await _start_fake_server()
    client = PulseModelGatewayClient(target, service_token="secret", timeout_seconds=2)
    try:
        events = [event async for event in client.stream_chat([{"role": "user", "content": "Hi"}], purpose="planner")]
        embedded = await client.embed(["one", "three"], request_id="embed-1")
    finally:
        await client.close()
        await server.stop(0)

    assert events == [
        {"delta": "hello", "done": False, "usage": {}, "error": {}},
        {
            "delta": "",
            "done": True,
            "usage": {
                "provider_key": "fake",
                "model_id": "",
                "input_tokens": 0,
                "output_tokens": 0,
                "latency_ms": 0,
                "metadata_json": {},
            },
            "error": {},
        },
    ]
    assert embedded["embeddings"] == [[3.0, 0.5], [5.0, 0.5]]
    assert embedded["usage"]["model_id"] == "fake-embed"


async def test_chat_maps_grpc_errors() -> None:
    target, server, _fake = await _start_fake_server()
    client = PulseModelGatewayClient(target, service_token="secret", timeout_seconds=2)
    try:
        with pytest.raises(PulseModelGatewayError) as exc:
            await client.chat([{"role": "user", "content": "bad"}], purpose="planner", request_id="fail")
    finally:
        await client.close()
        await server.stop(0)
    assert exc.value.code == "failed_precondition"
    assert "model_schema_invalid" in str(exc.value)


async def test_client_requires_service_token_by_default(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("AI_AGENT_INSECURE_DEV", raising=False)
    client = PulseModelGatewayClient("127.0.0.1:1", service_token="")
    with pytest.raises(RuntimeError, match="Pulse ModelGateway gRPC requires"):
        await client.chat([{"role": "user", "content": "hello"}], purpose="planner")


async def _start_fake_server() -> tuple[str, grpc.aio.Server, FakeModelGateway]:
    fake = FakeModelGateway()
    server = grpc.aio.server()
    model_gateway_pb2_grpc.add_ModelGatewayServiceServicer_to_server(fake, server)
    port = server.add_insecure_port("127.0.0.1:0")
    await server.start()
    return f"127.0.0.1:{port}", server, fake


def _struct(value: dict[str, Any]) -> Struct:
    out = Struct()
    json_format.ParseDict(value, out)
    return out


def _list(value: list[float]) -> ListValue:
    out = ListValue()
    for item in value:
        out.values.add().number_value = item
    return out
