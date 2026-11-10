from __future__ import annotations

import sys
from concurrent import futures
from pathlib import Path
from typing import Any

import grpc
import pytest
from google.protobuf import json_format
from google.protobuf.struct_pb2 import Struct
from google.protobuf.timestamp_pb2 import Timestamp
from langgraph.checkpoint.memory import MemorySaver
from langgraph.types import Command

import pulse_agent.grpc_server as grpc_server
from pulse_agent.checkpointing import CheckpointScope, normalize_thread_scope_status
from pulse_agent.config import Settings

_PROTO_ROOT = Path(__file__).resolve().parents[1] / "proto"
if str(_PROTO_ROOT) not in sys.path:
    sys.path.insert(0, str(_PROTO_ROOT))

from ai_agent.v1 import agent_runtime_pb2, agent_runtime_pb2_grpc, common_pb2  # noqa: E402


def test_agent_runtime_grpc_contract(monkeypatch: pytest.MonkeyPatch) -> None:
    graph = FakeGraph()
    monkeypatch.setattr(grpc_server, "build_graph", lambda: FakeGraphBuilder(graph))
    runtime = grpc_server.AgentRuntime(checkpoint_runtime=FakeResumeCheckpointRuntime())  # type: ignore[arg-type]
    server, addr = _serve_runtime(runtime, token="contract-token")
    try:
        channel = grpc.insecure_channel(addr)
        stub = agent_runtime_pb2_grpc.AgentRuntimeServiceStub(channel)
        metadata = (("x-pulse-service-token", "contract-token"),)
        request = agent_runtime_pb2.TurnRequest(
            request_id="req-1",
            idempotency_key="idem-1",
            project_id="project-1",
            user_id="user-1",
            session_id="session-1",
            thread_id="thread-1",
            body="Проверь контракт",
            local_date="2026-06-25",
            timezone="Europe/Moscow",
            locale="ru-RU",
            page_context_json=_struct({"route": "/x", "nested": {"items": [1, True, "тест"]}}),
            client_context_json=_struct({"surface": "dashboard"}),
            ui_state_json=_struct({"panel": "ai"}),
            history=[
                common_pb2.ChatMessage(
                    message_id="m-1",
                    role="user",
                    body="история",
                    payload_json=_struct({"source": "fixture"}),
                    created_at=_timestamp(),
                )
            ],
        )

        events = list(stub.StartTurn(request, metadata=metadata, timeout=5))
        assert [event.seq for event in events] == [1, 2, 3, 4, 5]
        assert events[-1].event_type == agent_runtime_pb2.AGENT_EVENT_TYPE_DONE
        assert events[3].event_type == agent_runtime_pb2.AGENT_EVENT_TYPE_STATUS
        assert graph.state["page_context"]["nested"]["items"] == [1.0, True, "тест"]
        assert graph.state["messages"][0]["message_id"] == "m-1"
        assert graph.state["messages"][0]["payload_json"] == {"source": "fixture"}

        resumed = list(
            stub.ResumeTurn(
                agent_runtime_pb2.ResumeRequest(
                    request_id="resume-1",
                    project_id="project-1",
                    session_id="session-1",
                    thread_id="thread-1",
                    run_id="run-1",
                    decision=agent_runtime_pb2.ApprovalDecision(action="approve", reason="ok", card_revision=2),
                ),
                metadata=metadata,
                timeout=5,
            )
        )
        assert [event.event_type for event in resumed] == [
            agent_runtime_pb2.AGENT_EVENT_TYPE_RUN_STARTED,
            agent_runtime_pb2.AGENT_EVENT_TYPE_STATUS,
            agent_runtime_pb2.AGENT_EVENT_TYPE_DONE,
        ]
        assert graph.resume_payload["action"] == "approve"
        assert graph.resume_payload["active_tool_call_id"] == "tool-call-1"

        cancel = stub.CancelRun(
            agent_runtime_pb2.CancelRunRequest(
                request_id="cancel-1",
                project_id="project-1",
                session_id="session-1",
                thread_id="thread-1",
                run_id="run-1",
            ),
            metadata=metadata,
            timeout=5,
        )
        assert cancel.ok is True
        assert cancel.status == "canceled"

        state = stub.GetThreadState(
            agent_runtime_pb2.GetThreadStateRequest(
                request_id="state-1",
                project_id="project-1",
                session_id="session-1",
                thread_id="thread-1",
            ),
            metadata=metadata,
            timeout=5,
        )
        assert state.thread_id == "thread-1"
        state_payload = json_format.MessageToDict(state.state_json)
        assert state_payload["runtime"] == "langgraph"
        assert state_payload["project_id"] == "project-1"

        with pytest.raises(grpc.RpcError) as exc:
            list(stub.StartTurn(request, metadata=(("x-pulse-service-token", "wrong"),), timeout=5))
        assert exc.value.code() == grpc.StatusCode.UNAUTHENTICATED
    finally:
        server.stop(grace=None)


def test_checkpoint_scope_status_matches_database_contract() -> None:
    assert normalize_thread_scope_status("interrupted") == "waiting_approval"
    assert normalize_thread_scope_status("resuming") == "running"
    assert normalize_thread_scope_status("succeeded") == "succeeded"
    assert normalize_thread_scope_status("unexpected") == "failed"


def test_agent_runtime_uses_checkpoint_scoped_config(monkeypatch: pytest.MonkeyPatch) -> None:
    graph = ScopedConfigGraph()
    checkpoint_runtime = FakeCheckpointRuntime()
    monkeypatch.setattr(grpc_server, "build_graph", lambda: FakeGraphBuilder(graph))
    runtime = grpc_server.AgentRuntime(Settings(), checkpoint_runtime=checkpoint_runtime)  # type: ignore[arg-type]

    events = list(
        runtime.StartTurn(
            agent_runtime_pb2.TurnRequest(
                request_id="req-1",
                company_id="company-1",
                project_id="project-1",
                session_id="session-1",
                thread_id="public-thread",
                body="проверь",
            ),
            context=None,  # type: ignore[arg-type]
        )
    )

    assert events[-1].event_type == agent_runtime_pb2.AGENT_EVENT_TYPE_DONE
    assert graph.config["configurable"] == {"thread_id": "checkpoint-thread"}
    assert graph.state["thread_id"] == "public-thread"
    assert graph.state["checkpoint_thread_id"] == "checkpoint-thread"
    assert checkpoint_runtime.recorded_statuses == ["running", "succeeded"]


def test_agent_runtime_resumes_langgraph_interrupt_without_duplicate_write() -> None:
    checkpoint_runtime = MemoryCheckpointRuntime()
    runtime = grpc_server.AgentRuntime(Settings(), checkpoint_runtime=checkpoint_runtime)  # type: ignore[arg-type]
    try:
        start_events = list(
            runtime.StartTurn(
                agent_runtime_pb2.TurnRequest(
                    request_id="req-approval",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-approval",
                    thread_id="thread-approval",
                    body="Создай задачу проверить оплату Орион завтра",
                    local_date="2026-06-25",
                    timezone="Europe/Moscow",
                ),
                context=None,  # type: ignore[arg-type]
            )
        )

        assert any(event.event_type == agent_runtime_pb2.AGENT_EVENT_TYPE_CARD_CREATED for event in start_events)
        assert any(event.event_type == agent_runtime_pb2.AGENT_EVENT_TYPE_APPROVAL_REQUIRED for event in start_events)
        start_tool_events = [
            json_format.MessageToDict(event.tool_call)
            for event in start_events
            if event.event_type == agent_runtime_pb2.AGENT_EVENT_TYPE_TOOL_STARTED
        ]
        assert not [event for event in start_tool_events if _json_key(event, "tool_key") == "approval.apply"]
        snapshot = checkpoint_runtime.snapshot("thread-approval")
        assert snapshot.next == ("await_approval",)
        assert snapshot.values["active_card"]["revision"] == 1

        resume_events = list(
            runtime.ResumeTurn(
                agent_runtime_pb2.ResumeRequest(
                    request_id="resume-approval",
                    idempotency_key="resume-key-1",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-approval",
                    thread_id="thread-approval",
                    run_id="run-resume",
                    interrupt_id="11111111-1111-1111-1111-111111111111",
                    decision=agent_runtime_pb2.ApprovalDecision(action="approve", reason="ok", card_revision=1),
                ),
                context=None,  # type: ignore[arg-type]
            )
        )

        tool_events = [
            json_format.MessageToDict(event.tool_call)
            for event in resume_events
            if event.event_type == agent_runtime_pb2.AGENT_EVENT_TYPE_TOOL_STARTED
        ]
        assert [event.event_type for event in resume_events] == [
            agent_runtime_pb2.AGENT_EVENT_TYPE_STATUS,
            agent_runtime_pb2.AGENT_EVENT_TYPE_STATUS,
            agent_runtime_pb2.AGENT_EVENT_TYPE_TOOL_STARTED,
            agent_runtime_pb2.AGENT_EVENT_TYPE_DONE,
        ]
        assert _json_key(tool_events[-1], "tool_key") == "approval.apply"
        input_json = _json_key(tool_events[-1], "input_json")
        assert isinstance(input_json, dict)
        assert _json_key(input_json, "active_tool_call_id") == "11111111-1111-1111-1111-111111111111"
        assert checkpoint_runtime.snapshot("thread-approval").next == ()
        assert checkpoint_runtime.snapshot("thread-approval").values["thread_revision"] == 2

        repeated = list(
            runtime.ResumeTurn(
                agent_runtime_pb2.ResumeRequest(
                    request_id="resume-approval-again",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-approval",
                    thread_id="thread-approval",
                    run_id="run-resume",
                    interrupt_id="11111111-1111-1111-1111-111111111111",
                    decision=agent_runtime_pb2.ApprovalDecision(action="approve", reason="ok", card_revision=1),
                ),
                context=None,  # type: ignore[arg-type]
            )
        )
        assert [event.event_type for event in repeated] == [
            agent_runtime_pb2.AGENT_EVENT_TYPE_STATUS,
            agent_runtime_pb2.AGENT_EVENT_TYPE_DONE,
        ]
        assert "already processed" in repeated[0].message
    finally:
        runtime.close()


def test_agent_runtime_start_turn_revision_and_duplicate_idempotency() -> None:
    checkpoint_runtime = MemoryCheckpointRuntime()
    runtime = grpc_server.AgentRuntime(Settings(), checkpoint_runtime=checkpoint_runtime)  # type: ignore[arg-type]
    try:
        first = list(
            runtime.StartTurn(
                agent_runtime_pb2.TurnRequest(
                    request_id="req-idem",
                    idempotency_key="idem-key-1",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-idem",
                    thread_id="thread-idem",
                    body="ты тут?",
                ),
                context=None,  # type: ignore[arg-type]
            )
        )
        assert first[-1].event_type == agent_runtime_pb2.AGENT_EVENT_TYPE_DONE
        snapshot = checkpoint_runtime.snapshot("thread-idem")
        assert snapshot.values["thread_revision"] == 1
        assert snapshot.values["last_turn_idempotency_key"] == "idem-key-1"

        duplicate = list(
            runtime.StartTurn(
                agent_runtime_pb2.TurnRequest(
                    request_id="req-idem-duplicate",
                    idempotency_key="idem-key-1",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-idem",
                    thread_id="thread-idem",
                    body="ты тут?",
                ),
                context=None,  # type: ignore[arg-type]
            )
        )
        assert [event.event_type for event in duplicate] == [
            agent_runtime_pb2.AGENT_EVENT_TYPE_STATUS,
            agent_runtime_pb2.AGENT_EVENT_TYPE_DONE,
        ]
        assert "Duplicate turn replayed" in duplicate[0].message
        assert checkpoint_runtime.snapshot("thread-idem").values["thread_revision"] == 1

        state = runtime.GetThreadState(
            agent_runtime_pb2.GetThreadStateRequest(
                request_id="state-idem",
                project_id="project-1",
                session_id="session-idem",
                thread_id="thread-idem",
                auth_context=common_pb2.AuthContext(company_id="company-1"),
            ),
            context=None,  # type: ignore[arg-type]
        )
        state_json = json_format.MessageToDict(state.state_json)
        assert _json_key(state_json, "thread_revision") == 1
    finally:
        runtime.close()


def test_agent_runtime_rejects_stale_mutating_start_turn() -> None:
    checkpoint_runtime = MemoryCheckpointRuntime()
    runtime = grpc_server.AgentRuntime(Settings(), checkpoint_runtime=checkpoint_runtime)  # type: ignore[arg-type]
    try:
        list(
            runtime.StartTurn(
                agent_runtime_pb2.TurnRequest(
                    request_id="req-concurrent",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-concurrent",
                    thread_id="thread-concurrent",
                    body="Создай задачу проверить оплату Орион завтра",
                    local_date="2026-06-25",
                    timezone="Europe/Moscow",
                ),
                context=None,  # type: ignore[arg-type]
            )
        )
        assert checkpoint_runtime.snapshot("thread-concurrent").values["thread_revision"] == 1

        stale = list(
            runtime.StartTurn(
                agent_runtime_pb2.TurnRequest(
                    request_id="req-concurrent-stale",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-concurrent",
                    thread_id="thread-concurrent",
                    body="Создай задачу позвонить Орион завтра",
                    client_context_json=_struct({"expectedThreadRevision": 99}),
                    local_date="2026-06-25",
                    timezone="Europe/Moscow",
                ),
                context=None,  # type: ignore[arg-type]
            )
        )
        assert [event.event_type for event in stale] == [
            agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR,
            agent_runtime_pb2.AGENT_EVENT_TYPE_DONE,
        ]
        error = json_format.MessageToDict(stale[0].error)
        assert error["code"] == "thread_revision_conflict"
        assert _json_key(error, "actual_thread_revision") == 1
        latest_state = _json_key(error, "latest_state")
        assert isinstance(latest_state, dict)
        assert _json_key(latest_state, "pending_interrupt") is True
    finally:
        runtime.close()


def test_agent_runtime_uses_checkpoint_pending_action_for_new_write_switch() -> None:
    checkpoint_runtime = MemoryCheckpointRuntime()
    runtime = grpc_server.AgentRuntime(Settings(), checkpoint_runtime=checkpoint_runtime)  # type: ignore[arg-type]
    try:
        first = list(
            runtime.StartTurn(
                agent_runtime_pb2.TurnRequest(
                    request_id="req-client-create",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-switch",
                    thread_id="thread-switch",
                    body="Создай клиента Орион",
                    local_date="2026-06-25",
                    timezone="Europe/Moscow",
                ),
                context=None,  # type: ignore[arg-type]
            )
        )
        assert first[-1].event_type == agent_runtime_pb2.AGENT_EVENT_TYPE_APPROVAL_REQUIRED
        active_card = checkpoint_runtime.snapshot("thread-switch").values["active_card"]
        assert active_card["card_type"] == "client_create_proposal"

        second = list(
            runtime.StartTurn(
                agent_runtime_pb2.TurnRequest(
                    request_id="req-task-create-after-client",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-switch",
                    thread_id="thread-switch",
                    body="Создай задачу проверить оплату Орион завтра",
                    local_date="2026-06-25",
                    timezone="Europe/Moscow",
                ),
                context=None,  # type: ignore[arg-type]
            )
        )
        assert not [event for event in second if event.event_type == agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR]
        cards = [json_format.MessageToDict(event.card) for event in second if event.card]
        assert len(cards) == 1
        assert cards[-1]["card_type"] == "task_create_proposal"
        snapshot = checkpoint_runtime.snapshot("thread-switch")
        assert snapshot.values["active_card"]["card_type"] == "task_create_proposal"
        assert snapshot.values["active_tool_call_id"] == "pending-task-create"
    finally:
        runtime.close()


def test_agent_runtime_resume_rejects_stale_card_revision() -> None:
    checkpoint_runtime = MemoryCheckpointRuntime()
    runtime = grpc_server.AgentRuntime(Settings(), checkpoint_runtime=checkpoint_runtime)  # type: ignore[arg-type]
    try:
        list(
            runtime.StartTurn(
                agent_runtime_pb2.TurnRequest(
                    request_id="req-stale",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-stale",
                    thread_id="thread-stale",
                    body="Создай задачу проверить оплату Орион завтра",
                    local_date="2026-06-25",
                    timezone="Europe/Moscow",
                ),
                context=None,  # type: ignore[arg-type]
            )
        )

        stale = list(
            runtime.ResumeTurn(
                agent_runtime_pb2.ResumeRequest(
                    request_id="resume-stale",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-stale",
                    thread_id="thread-stale",
                    run_id="run-stale",
                    interrupt_id="11111111-1111-1111-1111-111111111111",
                    decision=agent_runtime_pb2.ApprovalDecision(action="approve", reason="ok", card_revision=2),
                ),
                context=None,  # type: ignore[arg-type]
            )
        )
        assert [event.event_type for event in stale] == [
            agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR,
            agent_runtime_pb2.AGENT_EVENT_TYPE_DONE,
        ]
        assert json_format.MessageToDict(stale[0].error)["code"] == "stale_card_revision"
        assert checkpoint_runtime.snapshot("thread-stale").next == ("await_approval",)
    finally:
        runtime.close()


def test_agent_runtime_resume_edit_patches_card_and_keeps_pending_interrupt() -> None:
    checkpoint_runtime = MemoryCheckpointRuntime()
    runtime = grpc_server.AgentRuntime(Settings(), checkpoint_runtime=checkpoint_runtime)  # type: ignore[arg-type]
    try:
        list(
            runtime.StartTurn(
                agent_runtime_pb2.TurnRequest(
                    request_id="req-edit",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-edit",
                    thread_id="thread-edit",
                    body="Создай задачу проверить оплату Орион завтра",
                    local_date="2026-06-25",
                    timezone="Europe/Moscow",
                ),
                context=None,  # type: ignore[arg-type]
            )
        )

        edit_events = list(
            runtime.ResumeTurn(
                agent_runtime_pb2.ResumeRequest(
                    request_id="resume-edit",
                    idempotency_key="resume-edit-key",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-edit",
                    thread_id="thread-edit",
                    run_id="run-edit",
                    interrupt_id="pending-task-create",
                    decision=agent_runtime_pb2.ApprovalDecision(action="edit", reason="fix card", card_revision=1),
                    edited_card_json=_struct(
                        {
                            "input_json": {
                                "priority": "low",
                                "due_date": "2026-06-26",
                                "due_date_label": "пятницу",
                                "provider_payload": {"secret": "drop"},
                            },
                            "debug_json": {"token": "drop"},
                        }
                    ),
                ),
                context=None,  # type: ignore[arg-type]
            )
        )

        assert any(event.event_type == agent_runtime_pb2.AGENT_EVENT_TYPE_CARD_PATCHED for event in edit_events)
        assert any(event.event_type == agent_runtime_pb2.AGENT_EVENT_TYPE_APPROVAL_REQUIRED for event in edit_events)
        snapshot = checkpoint_runtime.snapshot("thread-edit")
        assert snapshot.next == ("await_approval",)
        active_card = snapshot.values["active_card"]
        assert active_card["revision"] == 2
        assert active_card["input_json"]["priority"] == "low"
        assert active_card["input_json"]["due_date"] == "2026-06-26"
        assert "provider_payload" not in active_card["input_json"]
        assert active_card["debug_json"]["base_revision"] == 1
        assert active_card["debug_json"]["safe_patch_keys"] == ["due_date", "due_date_label", "priority"]
        assert snapshot.values["pending_interrupt"]["card_revision"] == 2

        stale = list(
            runtime.ResumeTurn(
                agent_runtime_pb2.ResumeRequest(
                    request_id="resume-edit-stale",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-edit",
                    thread_id="thread-edit",
                    run_id="run-edit",
                    interrupt_id="pending-task-create",
                    decision=agent_runtime_pb2.ApprovalDecision(action="approve", reason="old tab", card_revision=1),
                ),
                context=None,  # type: ignore[arg-type]
            )
        )
        assert stale[0].event_type == agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR
        assert json_format.MessageToDict(stale[0].error)["code"] == "stale_card_revision"
    finally:
        runtime.close()


def test_agent_runtime_cancel_run_updates_checkpoint_state() -> None:
    checkpoint_runtime = MemoryCheckpointRuntime()
    runtime = grpc_server.AgentRuntime(Settings(), checkpoint_runtime=checkpoint_runtime)  # type: ignore[arg-type]
    try:
        list(
            runtime.StartTurn(
                agent_runtime_pb2.TurnRequest(
                    request_id="req-cancel",
                    company_id="company-1",
                    project_id="project-1",
                    session_id="session-cancel",
                    thread_id="thread-cancel",
                    body="Создай задачу проверить оплату Орион завтра",
                    local_date="2026-06-25",
                    timezone="Europe/Moscow",
                ),
                context=None,  # type: ignore[arg-type]
            )
        )
        assert checkpoint_runtime.snapshot("thread-cancel").next == ("await_approval",)

        cancel = runtime.CancelRun(
            agent_runtime_pb2.CancelRunRequest(
                request_id="cancel-1",
                project_id="project-1",
                session_id="session-cancel",
                thread_id="thread-cancel",
                run_id="run-cancel",
                auth_context=common_pb2.AuthContext(
                    company_id="company-1",
                    project_id="project-1",
                    user_id="user-1",
                ),
            ),
            context=None,  # type: ignore[arg-type]
        )

        assert cancel.ok is True
        assert cancel.status == "canceled"
        snapshot = checkpoint_runtime.snapshot("thread-cancel")
        assert snapshot.next == ()
        assert snapshot.values["current_step"] == "canceled"
        assert snapshot.values["active_mode"] == "canceled"
        assert snapshot.values["cancel_requested"] is True
        assert snapshot.values["cancel_reason"] == "user_cancel"
        assert snapshot.values["pending_interrupt"] is None
        assert snapshot.values["active_card"]["status"] == "canceled"

        repeated = runtime.CancelRun(
            agent_runtime_pb2.CancelRunRequest(
                request_id="cancel-2",
                project_id="project-1",
                session_id="session-cancel",
                thread_id="thread-cancel",
                run_id="run-cancel",
                auth_context=common_pb2.AuthContext(
                    company_id="company-1",
                    project_id="project-1",
                    user_id="user-1",
                ),
            ),
            context=None,  # type: ignore[arg-type]
        )
        assert repeated.ok is True
        assert repeated.status == "already_canceled"
    finally:
        runtime.close()


class FakeGraphBuilder:
    def __init__(self, graph: Any) -> None:
        self._graph = graph

    def compile(self) -> Any:
        return self._graph


class FakeGraph:
    def __init__(self) -> None:
        self.state: dict[str, Any] = {}
        self.resume_payload: dict[str, Any] = {}

    def invoke(self, state: dict[str, Any] | Command, config: dict[str, Any]) -> dict[str, Any]:
        assert config["configurable"]["thread_id"] == "thread-1"
        if isinstance(state, Command):
            self.resume_payload = state.resume
            return {
                "events": [
                    {"seq": 1, "event_type": "RUN_STARTED", "message": "resumed"},
                    {"seq": 2, "event_type": "STATUS", "message": "approval_resume=approve"},
                    {"seq": 3, "event_type": "DONE", "message": "done"},
                ]
            }
        self.state = state
        return {
            "events": [
                {"seq": 1, "event_type": "RUN_STARTED", "message": "started"},
                {"seq": 2, "event_type": "STATUS", "message": "planning"},
                {"seq": 3, "event_type": "MESSAGE_FINAL", "message": "готово"},
                {"seq": 4, "event_type": "FUTURE_EVENT", "message": "future"},
                {"seq": 5, "event_type": "DONE", "message": "done"},
            ]
        }


class ScopedConfigGraph:
    def __init__(self) -> None:
        self.state: dict[str, Any] = {}
        self.config: dict[str, Any] = {}

    def invoke(self, state: dict[str, Any], config: dict[str, Any]) -> dict[str, Any]:
        self.state = state
        self.config = config
        return {"events": [{"seq": 1, "event_type": "DONE", "message": "done"}]}


class FakeSnapshot:
    def __init__(self) -> None:
        self.values = {
            "events": [],
            "active_card": {"card_id": "card-1", "revision": 2},
            "pending_interrupt": {"active_tool_call_id": "tool-call-1", "interrupt_id": "interrupt-1"},
        }
        self.next = ("await_approval",)


class FakeResumeCheckpointRuntime:
    enabled = True

    def compile(self, graph_builder: FakeGraphBuilder) -> Any:
        return graph_builder.compile()

    def close(self) -> None:
        pass

    def scope(self, company_id: str, project_id: str, session_id: str, thread_id: str) -> CheckpointScope:
        return CheckpointScope(
            company_id=company_id,
            project_id=project_id,
            session_id=session_id,
            public_thread_id=thread_id,
            checkpoint_thread_id=thread_id,
            checkpoint_namespace="checkpoint-namespace",
        )

    def graph_config(self, scope: CheckpointScope) -> dict[str, dict[str, str]]:
        return {"configurable": {"thread_id": scope.checkpoint_thread_id}}

    def record_thread_scope(self, scope: CheckpointScope, *, status: str) -> None:
        assert scope.checkpoint_thread_id == "thread-1"
        assert status in {"running", "succeeded", "interrupted", "resuming", "failed", "canceled"}

    def get_state(self, graph: Any, scope: CheckpointScope) -> FakeSnapshot:
        assert scope.checkpoint_thread_id == "thread-1"
        return FakeSnapshot()

    def update_state(self, graph: Any, scope: CheckpointScope, values: dict[str, Any]) -> bool:
        assert scope.checkpoint_thread_id == "thread-1"
        assert values["current_step"] == "canceled"
        return True


class FakeCheckpointRuntime:
    enabled = True

    def __init__(self) -> None:
        self.recorded_statuses: list[str] = []

    def compile(self, graph_builder: FakeGraphBuilder) -> Any:
        return graph_builder.compile()

    def close(self) -> None:
        pass

    def scope(self, company_id: str, project_id: str, session_id: str, thread_id: str) -> CheckpointScope:
        assert (company_id, project_id, session_id, thread_id) == (
            "company-1",
            "project-1",
            "session-1",
            "public-thread",
        )
        return CheckpointScope(
            company_id=company_id,
            project_id=project_id,
            session_id=session_id,
            public_thread_id=thread_id,
            checkpoint_thread_id="checkpoint-thread",
            checkpoint_namespace="checkpoint-namespace",
        )

    def graph_config(self, scope: CheckpointScope) -> dict[str, dict[str, str]]:
        return {"configurable": {"thread_id": scope.checkpoint_thread_id}}

    def record_thread_scope(self, scope: CheckpointScope, *, status: str) -> None:
        assert scope.checkpoint_thread_id == "checkpoint-thread"
        self.recorded_statuses.append(status)

    def get_state(self, graph: Any, scope: CheckpointScope) -> None:
        return None


class MemoryCheckpointRuntime:
    enabled = True

    def __init__(self) -> None:
        self._checkpointer = MemorySaver()
        self._graph: Any | None = None

    def compile(self, graph_builder: FakeGraphBuilder) -> Any:
        self._graph = graph_builder.compile(checkpointer=self._checkpointer)
        return self._graph

    def close(self) -> None:
        pass

    def scope(self, company_id: str, project_id: str, session_id: str, thread_id: str) -> CheckpointScope:
        return CheckpointScope(
            company_id=company_id,
            project_id=project_id,
            session_id=session_id,
            public_thread_id=thread_id,
            checkpoint_thread_id=thread_id,
            checkpoint_namespace=f"{company_id}:{project_id}:{session_id}",
        )

    def graph_config(self, scope: CheckpointScope) -> dict[str, dict[str, str]]:
        return {"configurable": {"thread_id": scope.checkpoint_thread_id}}

    def record_thread_scope(self, scope: CheckpointScope, *, status: str) -> None:
        assert scope.checkpoint_thread_id
        assert status in {"running", "succeeded", "interrupted", "resuming", "failed", "canceled"}

    def get_state(self, graph: Any, scope: CheckpointScope) -> Any:
        return graph.get_state(config=self.graph_config(scope))

    def update_state(self, graph: Any, scope: CheckpointScope, values: dict[str, Any]) -> bool:
        graph.update_state(config=self.graph_config(scope), values=values)
        return True

    def snapshot(self, thread_id: str) -> Any:
        assert self._graph is not None
        return self._graph.get_state(config={"configurable": {"thread_id": thread_id}})


def _serve_runtime(
    runtime: agent_runtime_pb2_grpc.AgentRuntimeServiceServicer,
    *,
    token: str,
) -> tuple[grpc.Server, str]:
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=4),
        interceptors=[grpc_server.ServiceTokenInterceptor(token)],
    )
    agent_runtime_pb2_grpc.add_AgentRuntimeServiceServicer_to_server(runtime, server)
    port = server.add_insecure_port("127.0.0.1:0")
    server.start()
    return server, f"127.0.0.1:{port}"


def _struct(value: dict[str, Any]) -> Struct:
    out = Struct()
    json_format.ParseDict(value, out)
    return out


def _timestamp() -> Timestamp:
    out = Timestamp()
    out.FromJsonString("2026-06-25T04:00:00Z")
    return out


def _json_key(value: dict[str, Any], snake_key: str) -> Any:
    camel_key = snake_key.split("_")[0] + "".join(part.title() for part in snake_key.split("_")[1:])
    return value.get(snake_key, value.get(camel_key))
