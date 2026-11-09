from __future__ import annotations

import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Any

from google.protobuf import json_format
from google.protobuf.struct_pb2 import Struct

from pulse_agent.clients.grpc_security import internal_aio_channel

_PROTO_ROOT = Path(__file__).resolve().parents[1] / "proto"
if str(_PROTO_ROOT) not in sys.path:
    sys.path.insert(0, str(_PROTO_ROOT))

from ai_agent.v1 import common_pb2, pulse_tools_pb2, pulse_tools_pb2_grpc  # noqa: E402


def _struct(value: dict[str, Any]) -> Struct:
    out = Struct()
    json_format.ParseDict(value, out)
    return out


def _dict(value: Struct) -> dict[str, Any]:
    return dict(json_format.MessageToDict(value))


def _metadata(token: str) -> tuple[tuple[str, str], ...]:
    if not token:
        return ()
    return (
        ("x-pulse-service", "pulse-ai-agent"),
        ("x-pulse-service-token", token),
    )


def _risk(value: str) -> int:
    return int({
        "read": common_pb2.RISK_LEVEL_READ,
        "medium": common_pb2.RISK_LEVEL_MEDIUM,
        "high": common_pb2.RISK_LEVEL_HIGH,
        "critical": common_pb2.RISK_LEVEL_CRITICAL,
    }.get(value, common_pb2.RISK_LEVEL_UNSPECIFIED))


@dataclass
class PulseToolsClient:
    target: str
    service_token: str = ""

    async def list_tools(self, project_id: str, user_id: str, *, company_id: str = "") -> list[dict[str, Any]]:
        async with internal_aio_channel(self.target, self.service_token, service_name="Pulse ToolGateway") as channel:
            stub = pulse_tools_pb2_grpc.PulseToolServiceStub(channel)
            response = await stub.ListTools(
                pulse_tools_pb2.ListToolsRequest(
                    company_id=company_id,
                    project_id=project_id,
                    user_id=user_id,
                ),
                metadata=_metadata(self.service_token),
            )
        return [
            {
                "tool_key": tool.tool_key,
                "display_name": tool.display_name,
                "description": tool.description,
                "risk_level": common_pb2.RiskLevel.Name(tool.risk_level).removeprefix("RISK_LEVEL_").lower(),
                "requires_approval": tool.requires_approval,
                "idempotency_required": tool.idempotency_required,
                "ui_card_type": tool.ui_card_type,
                "timeout_ms": tool.timeout_ms,
            }
            for tool in response.tools
        ]

    async def validate_tool(
        self,
        *,
        company_id: str,
        project_id: str,
        user_id: str,
        session_id: str,
        thread_id: str,
        tool_call_id: str,
        tool_key: str,
        input_json: dict[str, Any],
    ) -> dict[str, Any]:
        async with internal_aio_channel(self.target, self.service_token, service_name="Pulse ToolGateway") as channel:
            stub = pulse_tools_pb2_grpc.PulseToolServiceStub(channel)
            response = await stub.ValidateToolCall(
                pulse_tools_pb2.ValidateToolCallRequest(
                    company_id=company_id,
                    project_id=project_id,
                    user_id=user_id,
                    session_id=session_id,
                    thread_id=thread_id,
                    tool_call_id=tool_call_id,
                    tool_key=tool_key,
                    input_json=_struct(input_json),
                ),
                metadata=_metadata(self.service_token),
            )
        return {
            "ok": response.ok,
            "normalized_input_json": _dict(response.normalized_input_json),
            "missing_fields": [_dict(item) for item in response.missing_fields],
            "user_message": response.user_message,
            "error_code": response.error_code,
            "error_message": response.error_message,
        }

    async def create_proposal(
        self,
        *,
        company_id: str,
        project_id: str,
        user_id: str,
        session_id: str,
        thread_id: str,
        tool_call_id: str,
        tool_key: str,
        intent_key: str,
        risk_level: str,
        input_json: dict[str, Any],
        idempotency_key: str,
        expected_card_type: str,
    ) -> dict[str, Any]:
        async with internal_aio_channel(self.target, self.service_token, service_name="Pulse ToolGateway") as channel:
            stub = pulse_tools_pb2_grpc.PulseToolServiceStub(channel)
            response = await stub.CreateProposal(
                pulse_tools_pb2.CreateProposalRequest(
                    company_id=company_id,
                    project_id=project_id,
                    user_id=user_id,
                    session_id=session_id,
                    thread_id=thread_id,
                    tool_call_id=tool_call_id,
                    tool_key=tool_key,
                    intent_key=intent_key,
                    risk_level=_risk(risk_level),
                    requires_approval=True,
                    input_json=_struct(input_json),
                    idempotency_key=idempotency_key,
                    expected_card_type=expected_card_type,
                ),
                metadata=_metadata(self.service_token),
            )
        return {
            "tool_call_id": response.tool_call_id,
            "tool_key": response.tool_key,
            "card_id": response.card_id,
            "card_type": response.card_type,
            "revision": response.revision,
            "status": response.status,
            "input_json": _dict(response.input_json),
            "card_json": _dict(response.card_json),
            "user_message": response.user_message,
            "error_code": response.error_code,
            "error_message": response.error_message,
        }

    async def patch_proposal(
        self,
        *,
        company_id: str,
        project_id: str,
        user_id: str,
        session_id: str,
        thread_id: str,
        tool_call_id: str,
        card_id: str,
        base_revision: int,
        patch_json: dict[str, Any],
        page_context_json: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        async with internal_aio_channel(self.target, self.service_token, service_name="Pulse ToolGateway") as channel:
            stub = pulse_tools_pb2_grpc.PulseToolServiceStub(channel)
            response = await stub.PatchProposal(
                pulse_tools_pb2.PatchProposalRequest(
                    company_id=company_id,
                    project_id=project_id,
                    user_id=user_id,
                    session_id=session_id,
                    thread_id=thread_id,
                    tool_call_id=tool_call_id,
                    card_id=card_id,
                    base_revision=base_revision,
                    patch_json=_struct(patch_json),
                    page_context_json=_struct(page_context_json or {}),
                ),
                metadata=_metadata(self.service_token),
            )
        return _proposal_response(response)

    async def approve_proposal(
        self,
        *,
        company_id: str,
        project_id: str,
        user_id: str,
        session_id: str,
        thread_id: str,
        tool_call_id: str,
        card_id: str,
        card_revision: int,
        idempotency_key: str,
    ) -> dict[str, Any]:
        return await self._approval(
            decision="approve",
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

    async def run_read_tool(
        self,
        *,
        company_id: str,
        project_id: str,
        user_id: str,
        session_id: str,
        thread_id: str,
        tool_call_id: str,
        tool_key: str,
        input_json: dict[str, Any],
        idempotency_key: str = "",
        approval_token: str = "",
        dry_run: bool = False,
    ) -> dict[str, Any]:
        async with internal_aio_channel(self.target, self.service_token, service_name="Pulse ToolGateway") as channel:
            stub = pulse_tools_pb2_grpc.PulseToolServiceStub(channel)
            rpc = getattr(stub, "Execute" + "Tool")
            request_type = getattr(pulse_tools_pb2, "Execute" + "ToolRequest")
            response = await rpc(
                request_type(
                    request_id=tool_call_id,
                    idempotency_key=idempotency_key,
                    company_id=company_id,
                    project_id=project_id,
                    user_id=user_id,
                    session_id=session_id,
                    thread_id=thread_id,
                    tool_call_id=tool_call_id,
                    tool_key=tool_key,
                    input_json=_struct(input_json),
                    approval_token=approval_token,
                    dry_run=dry_run,
                ),
                metadata=_metadata(self.service_token),
            )
        return _tool_observation_response(response)

    async def reject_proposal(
        self,
        *,
        company_id: str,
        project_id: str,
        user_id: str,
        session_id: str,
        thread_id: str,
        tool_call_id: str,
        card_id: str,
        card_revision: int,
        idempotency_key: str,
    ) -> dict[str, Any]:
        return await self._approval(
            decision="reject",
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

    async def _approval(
        self,
        *,
        decision: str,
        company_id: str,
        project_id: str,
        user_id: str,
        session_id: str,
        thread_id: str,
        tool_call_id: str,
        card_id: str,
        card_revision: int,
        idempotency_key: str,
    ) -> dict[str, Any]:
        async with internal_aio_channel(self.target, self.service_token, service_name="Pulse ToolGateway") as channel:
            stub = pulse_tools_pb2_grpc.PulseToolServiceStub(channel)
            request = pulse_tools_pb2.ApprovalRequest(
                company_id=company_id,
                project_id=project_id,
                user_id=user_id,
                session_id=session_id,
                thread_id=thread_id,
                tool_call_id=tool_call_id,
                card_id=card_id,
                card_revision=card_revision,
                decision=decision,
                idempotency_key=idempotency_key,
            )
            if decision == "reject":
                response = await stub.RejectProposal(request, metadata=_metadata(self.service_token))
            else:
                response = await stub.ApproveProposal(request, metadata=_metadata(self.service_token))
        return {
            **_tool_observation_response(response),
        }


def _proposal_response(response: Any) -> dict[str, Any]:
    return {
        "tool_call_id": response.tool_call_id,
        "tool_key": response.tool_key,
        "card_id": response.card_id,
        "card_type": response.card_type,
        "revision": response.revision,
        "status": response.status,
        "input_json": _dict(response.input_json),
        "card_json": _dict(response.card_json),
        "user_message": response.user_message,
        "error_code": response.error_code,
        "error_message": response.error_message,
    }


def _tool_observation_response(response: Any) -> dict[str, Any]:
    return {
        "ok": response.ok,
        "status": common_pb2.ToolExecutionStatus.Name(response.status).removeprefix(
            "TOOL_EXECUTION_STATUS_"
        ).lower(),
        "output_json": _dict(response.output_json),
        "error_code": response.error_code,
        "error_message": response.error_message,
        "user_message": response.user_message,
        "audit_event_id": response.audit_event_id,
    }
