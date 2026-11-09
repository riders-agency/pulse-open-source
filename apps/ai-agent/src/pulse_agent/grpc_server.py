from __future__ import annotations

import logging
import os
import sys
from collections.abc import Iterator
from concurrent import futures
from contextlib import suppress
from datetime import UTC, datetime
from importlib import metadata
from ipaddress import ip_address
from pathlib import Path
from typing import Any

import grpc
from google.protobuf import json_format
from google.protobuf.struct_pb2 import Struct
from google.protobuf.timestamp_pb2 import Timestamp
from grpc_health.v1 import health, health_pb2, health_pb2_grpc

from pulse_agent.checkpointing import STATE_SCHEMA_VERSION, CheckpointRuntime, CheckpointRuntimeError
from pulse_agent.config import Settings, validate_settings
from pulse_agent.degraded_mode import dependency_capability_matrix
from pulse_agent.dialogue_state import migrate_dialogue_state
from pulse_agent.execution_trace import execution_summary
from pulse_agent.graph.builder import build_graph
from pulse_agent.observability import monotonic_ms
from pulse_agent.thread_policy import (
    duplicate_turn_result,
    expected_thread_revision,
    is_mutating_turn,
    latest_thread_state,
    remember_turn,
    stable_turn_id,
    turn_idempotency_key,
)

_PROTO_ROOT = Path(__file__).resolve().parent / "proto"
if str(_PROTO_ROOT) not in sys.path:
    sys.path.insert(0, str(_PROTO_ROOT))

from ai_agent.v1 import agent_runtime_pb2, agent_runtime_pb2_grpc  # noqa: E402

log = logging.getLogger(__name__)


class ServiceTokenInterceptor(grpc.ServerInterceptor):
    def __init__(self, token: str) -> None:
        self._token = token

    def intercept_service(
        self,
        continuation: Any,
        handler_call_details: grpc.HandlerCallDetails,
    ) -> grpc.RpcMethodHandler | None:
        if not self._token:
            return continuation(handler_call_details)
        metadata = {key.lower(): value for key, value in (handler_call_details.invocation_metadata or [])}
        if metadata.get("x-pulse-service-token") == self._token:
            return continuation(handler_call_details)
        handler = continuation(handler_call_details)
        if handler is None:
            return None
        return _unauthenticated_handler(handler)


class AgentRuntime(agent_runtime_pb2_grpc.AgentRuntimeServiceServicer):
    def __init__(self, settings: Settings | None = None, checkpoint_runtime: CheckpointRuntime | None = None) -> None:
        self._settings = settings or Settings()
        self._checkpoint_runtime = checkpoint_runtime or CheckpointRuntime(self._settings)
        graph_builder = build_graph()
        self._graph = self._checkpoint_runtime.compile(graph_builder) if graph_builder is not None else None

    def close(self) -> None:
        self._checkpoint_runtime.close()

    def StartTurn(
        self, request: agent_runtime_pb2.TurnRequest, context: grpc.ServicerContext
    ) -> Iterator[agent_runtime_pb2.AgentEvent]:
        del context
        run_id = request.idempotency_key or request.request_id or "dev-run"
        thread_id = request.thread_id or f"pulse:{request.project_id}:{request.session_id}"
        checkpoint_scope = self._checkpoint_runtime.scope(
            request.company_id,
            request.project_id,
            request.session_id,
            thread_id,
        )
        if self._graph is None:
            yield _event(
                1,
                run_id,
                thread_id,
                request.session_id,
                agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR,
                "LangGraph runtime is not available",
                error={"code": "langgraph_unavailable"},
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        page_context = json_format.MessageToDict(request.page_context_json)
        client_context = json_format.MessageToDict(request.client_context_json)
        ui_state = json_format.MessageToDict(request.ui_state_json)
        try:
            checkpoint_snapshot = self._checkpoint_runtime.get_state(self._graph, checkpoint_scope)
            checkpoint_values = _snapshot_values(checkpoint_snapshot)
            checkpoint_next = [str(value) for value in (getattr(checkpoint_snapshot, "next", ()) or [])]
            previous_event_count = len(_list_value(checkpoint_values.get("events")))
        except CheckpointRuntimeError as exc:
            yield _event(
                1,
                run_id,
                thread_id,
                request.session_id,
                agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR,
                "LangGraph checkpoint error",
                error={"code": "langgraph_checkpoint_error", "message": str(exc)},
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        current_revision = _int_value(checkpoint_values.get("thread_revision"))
        previous_dialogue_state = migrate_dialogue_state(checkpoint_values.get("dialogue_state"), checkpoint_values)
        expected_revision = expected_thread_revision(page_context, client_context, ui_state)
        turn_key = turn_idempotency_key(request.request_id, request.idempotency_key, thread_id, request.body)
        duplicate = duplicate_turn_result(checkpoint_values, turn_key)
        if duplicate is not None:
            yield _event(
                1,
                run_id,
                thread_id,
                request.session_id,
                agent_runtime_pb2.AGENT_EVENT_TYPE_STATUS,
                "Duplicate turn replayed",
                tool_call={
                    "idempotency_key": turn_key,
                    "thread_revision": _int_value(duplicate.get("thread_revision")),
                },
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        effective_active_tool_call_id = request.active_tool_call_id or str(
            checkpoint_values.get("active_tool_call_id") or ""
        )
        mutating_turn = is_mutating_turn(
            request.body,
            active_tool_call_id=effective_active_tool_call_id,
            page_context=page_context,
        )
        if expected_revision > 0 and current_revision > 0 and expected_revision != current_revision and mutating_turn:
            yield _thread_policy_error(
                1,
                run_id,
                thread_id,
                request.session_id,
                "thread_revision_conflict",
                "Thread revision is stale",
                checkpoint_values,
                checkpoint_next,
                expected_revision=expected_revision,
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        if checkpoint_next and mutating_turn and not effective_active_tool_call_id:
            yield _thread_policy_error(
                1,
                run_id,
                thread_id,
                request.session_id,
                "active_mutating_turn_conflict",
                "Thread already has an active mutating turn",
                checkpoint_values,
                checkpoint_next,
                expected_revision=expected_revision,
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        next_revision = current_revision + 1
        state = {
            "company_id": request.company_id,
            "project_id": request.project_id,
            "user_id": request.user_id,
            "session_id": request.session_id,
            "thread_id": thread_id,
            "checkpoint_thread_id": checkpoint_scope.checkpoint_thread_id,
            "checkpoint_namespace": checkpoint_scope.checkpoint_namespace,
            "checkpoint_enabled": self._checkpoint_runtime.enabled,
            "checkpoint_status": "running",
            "state_schema_version": STATE_SCHEMA_VERSION,
            "service_version": _package_version("pulse-ai-agent"),
            "thread_revision": next_revision,
            "expected_thread_revision": expected_revision,
            "run_id": run_id,
            "turn_id": stable_turn_id(thread_id, turn_key),
            "idempotency_key": turn_key,
            "body": request.body,
            "local_date": request.local_date,
            "timezone": request.timezone,
            "locale": request.locale or "ru-RU",
            "page_context": page_context,
            "client_context": client_context,
            "ui_state": ui_state,
            "working_memory": _dict_value(checkpoint_values.get("working_memory")),
            "long_term_memory_refs": _list_value(checkpoint_values.get("long_term_memory_refs")),
            "dialogue_state": previous_dialogue_state,
            "active_topic": _dict_value(
                previous_dialogue_state.get("active_topic") or checkpoint_values.get("active_topic")
            ),
            "suspended_topics": _list_value(
                previous_dialogue_state.get("suspended_topics") or checkpoint_values.get("suspended_topics")
            ),
            "resolved_entities": _list_value(
                previous_dialogue_state.get("resolved_entities") or checkpoint_values.get("resolved_entities")
            ),
            "unresolved_slots": _dict_value(
                previous_dialogue_state.get("unresolved_slots") or checkpoint_values.get("unresolved_slots")
            ),
            "active_card_id": request.active_card_id or str(checkpoint_values.get("active_card_id") or ""),
            "active_tool_call_id": request.active_tool_call_id
            or str(checkpoint_values.get("active_tool_call_id") or ""),
            "active_proposal_id": request.active_proposal_id
            or str(checkpoint_values.get("active_proposal_id") or ""),
            "active_card": _dict_value(checkpoint_values.get("active_card")),
            "active_tool_call": _dict_value(checkpoint_values.get("active_tool_call")),
            "active_proposal": _dict_value(checkpoint_values.get("active_proposal")),
            "pending_interrupt": _dict_value(checkpoint_values.get("pending_interrupt")) or None,
            "messages": [_chat_message_to_state(message) for message in request.history],
            "trace_context": _trace_context_to_state(request.trace_context),
            "observability_started_ms": monotonic_ms(),
            "execution_event_start": previous_event_count,
            "node_trace": [],
            "model_calls": [],
        }
        try:
            self._checkpoint_runtime.record_thread_scope(checkpoint_scope, status="running")
            result = self._graph.invoke(state, config=self._checkpoint_runtime.graph_config(checkpoint_scope))
            status = "interrupted" if _has_interrupt(result) else "succeeded"
            _set_ai_observability_checkpoint_status(result, status)
            self._remember_turn_result(checkpoint_scope, turn_key, next_revision, result)
            self._checkpoint_runtime.record_thread_scope(checkpoint_scope, status=status)
        except CheckpointRuntimeError as exc:
            log.exception("langgraph checkpoint failed", extra={"run_id": run_id, "thread_id": thread_id})
            yield _event(
                1,
                run_id,
                thread_id,
                request.session_id,
                agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR,
                "LangGraph checkpoint error",
                error={"code": "langgraph_checkpoint_error", "message": str(exc)},
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        except Exception as exc:
            with suppress(CheckpointRuntimeError):
                self._checkpoint_runtime.record_thread_scope(checkpoint_scope, status="failed")
            log.exception("langgraph turn failed", extra={"run_id": run_id, "thread_id": thread_id})
            yield _event(
                1,
                run_id,
                thread_id,
                request.session_id,
                agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR,
                "LangGraph runtime error",
                error={"code": "langgraph_runtime_error", "message": str(exc)},
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        for event in _events_with_execution(result, previous_event_count):
            yield _event_from_graph(event, run_id=run_id, thread_id=thread_id, session_id=request.session_id)
        yield from _interrupt_events(
            result,
            run_id=run_id,
            thread_id=thread_id,
            session_id=request.session_id,
            previous_event_count=previous_event_count,
        )

    def ResumeTurn(
        self, request: agent_runtime_pb2.ResumeRequest, context: grpc.ServicerContext
    ) -> Iterator[agent_runtime_pb2.AgentEvent]:
        del context
        run_id = request.run_id or request.idempotency_key or request.request_id or "dev-run"
        thread_id = request.thread_id or f"pulse:{request.project_id}:{request.session_id}"
        company_id = request.company_id or (request.auth_context.company_id if request.auth_context.company_id else "")
        checkpoint_scope = self._checkpoint_runtime.scope(company_id, request.project_id, request.session_id, thread_id)
        if self._graph is None:
            yield _event(
                1,
                run_id,
                thread_id,
                request.session_id,
                agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR,
                "LangGraph runtime is not available",
                error={"code": "langgraph_unavailable"},
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        try:
            checkpoint_snapshot = self._checkpoint_runtime.get_state(self._graph, checkpoint_scope)
            checkpoint_values = _snapshot_values(checkpoint_snapshot)
            checkpoint_next = [str(value) for value in (getattr(checkpoint_snapshot, "next", ()) or [])]
        except CheckpointRuntimeError as exc:
            yield _event(
                1,
                run_id,
                thread_id,
                request.session_id,
                agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR,
                "LangGraph checkpoint error",
                error={"code": "langgraph_checkpoint_error", "message": str(exc)},
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        if checkpoint_snapshot is None:
            yield _event(
                1,
                run_id,
                thread_id,
                request.session_id,
                agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR,
                "Checkpoint state not found",
                error={"code": "checkpoint_not_found"},
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        current_revision = _int_value(checkpoint_values.get("thread_revision"))
        expected_revision = _expected_revision_from_trace_context(request.trace_context)
        if expected_revision > 0 and current_revision > 0 and expected_revision != current_revision:
            yield _thread_policy_error(
                1,
                run_id,
                thread_id,
                request.session_id,
                "thread_revision_conflict",
                "Thread revision is stale",
                checkpoint_values,
                checkpoint_next,
                expected_revision=expected_revision,
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        if not checkpoint_next:
            yield _event(
                1,
                run_id,
                thread_id,
                request.session_id,
                agent_runtime_pb2.AGENT_EVENT_TYPE_STATUS,
                "Resume already processed",
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        active_card = (
            checkpoint_values.get("active_card") if isinstance(checkpoint_values.get("active_card"), dict) else {}
        )
        expected_revision = _int_value(active_card.get("revision") if isinstance(active_card, dict) else None)
        requested_revision = int(request.decision.card_revision)
        if requested_revision > 0 and expected_revision > 0 and requested_revision != expected_revision:
            yield _event(
                1,
                run_id,
                thread_id,
                request.session_id,
                agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR,
                "Card revision is stale",
                error={
                    "code": "stale_card_revision",
                    "expected_revision": expected_revision,
                    "actual_revision": requested_revision,
                },
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        try:
            from langgraph.types import Command

            resume_payload = _resume_payload(request, checkpoint_values)
            self._checkpoint_runtime.record_thread_scope(checkpoint_scope, status="resuming")
            previous_event_count = len(_list_value(checkpoint_values.get("events")))
            result = self._graph.invoke(
                Command(resume=resume_payload),
                config=self._checkpoint_runtime.graph_config(checkpoint_scope),
            )
            status = "interrupted" if _has_interrupt(result) else "succeeded"
            _set_ai_observability_checkpoint_status(result, status)
            resume_key = request.idempotency_key or request.request_id or request.run_id
            self._remember_turn_result(checkpoint_scope, resume_key, current_revision + 1, result)
            self._checkpoint_runtime.record_thread_scope(checkpoint_scope, status=status)
        except CheckpointRuntimeError as exc:
            log.exception("langgraph checkpoint resume failed", extra={"run_id": run_id, "thread_id": thread_id})
            yield _event(
                1,
                run_id,
                thread_id,
                request.session_id,
                agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR,
                "LangGraph checkpoint error",
                error={"code": "langgraph_checkpoint_error", "message": str(exc)},
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        except Exception as exc:
            with suppress(CheckpointRuntimeError):
                self._checkpoint_runtime.record_thread_scope(checkpoint_scope, status="failed")
            log.exception("langgraph resume failed", extra={"run_id": run_id, "thread_id": thread_id})
            yield _event(
                1,
                run_id,
                thread_id,
                request.session_id,
                agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR,
                "LangGraph resume error",
                error={"code": "langgraph_resume_error", "message": str(exc)},
            )
            yield _event(2, run_id, thread_id, request.session_id, agent_runtime_pb2.AGENT_EVENT_TYPE_DONE, "DONE")
            return
        emitted = 0
        for event in _events_with_execution(result, previous_event_count):
            emitted += 1
            yield _event_from_graph(event, run_id=run_id, thread_id=thread_id, session_id=request.session_id)
        if emitted == 0:
            yield _event(
                1,
                run_id,
                thread_id,
                request.session_id,
                agent_runtime_pb2.AGENT_EVENT_TYPE_STATUS,
                "Resume completed",
            )
        yield from _interrupt_events(
            result,
            run_id=run_id,
            thread_id=thread_id,
            session_id=request.session_id,
            previous_event_count=previous_event_count,
        )

    def _remember_turn_result(
        self,
        checkpoint_scope: Any,
        idempotency_key: str,
        thread_revision: int,
        result: object,
    ) -> None:
        if not idempotency_key:
            return
        if self._graph is None or not hasattr(self._graph, "update_state"):
            return
        events = _new_graph_events(result, 0)
        final_response = str(result.get("final_response") or "") if isinstance(result, dict) else ""
        try:
            snapshot = self._checkpoint_runtime.get_state(self._graph, checkpoint_scope)
            values = _snapshot_values(snapshot)
            turn_result = {
                "thread_revision": thread_revision,
                "event_count": len(events),
                "final_response": final_response,
            }
            observability = _dict_value(result.get("ai_observability_json")) if isinstance(result, dict) else {}
            self._checkpoint_runtime.update_state(
                self._graph,
                checkpoint_scope,
                {
                    "thread_revision": thread_revision,
                    "last_turn_idempotency_key": idempotency_key,
                    "last_turn_result": turn_result,
                    "ai_observability_json": observability,
                    "processed_turns": remember_turn(values, idempotency_key, turn_result),
                },
            )
        except CheckpointRuntimeError:
            raise

    def CancelRun(
        self, request: agent_runtime_pb2.CancelRunRequest, context: grpc.ServicerContext
    ) -> agent_runtime_pb2.CancelRunResponse:
        del context
        thread_id = request.thread_id or f"pulse:{request.project_id}:{request.session_id}"
        company_id = request.auth_context.company_id if request.auth_context.company_id else ""
        checkpoint_scope = self._checkpoint_runtime.scope(company_id, request.project_id, request.session_id, thread_id)
        if self._graph is None:
            return agent_runtime_pb2.CancelRunResponse(ok=False, status="langgraph_unavailable")
        try:
            checkpoint_snapshot = self._checkpoint_runtime.get_state(self._graph, checkpoint_scope)
            checkpoint_values = _snapshot_values(checkpoint_snapshot)
            checkpoint_next = [str(value) for value in (getattr(checkpoint_snapshot, "next", ()) or [])]
            if checkpoint_snapshot is None:
                return agent_runtime_pb2.CancelRunResponse(ok=False, status="checkpoint_not_found")
            if str(checkpoint_values.get("current_step") or "") == "canceled":
                return agent_runtime_pb2.CancelRunResponse(ok=True, status="already_canceled")
            if checkpoint_next:
                from langgraph.types import Command

                self._graph.invoke(
                    Command(resume=_cancel_resume_payload(request, checkpoint_values)),
                    config=self._checkpoint_runtime.graph_config(checkpoint_scope),
                )
                self._checkpoint_runtime.update_state(
                    self._graph,
                    checkpoint_scope,
                    _canceled_state_values(request, checkpoint_values),
                )
            else:
                self._checkpoint_runtime.update_state(
                    self._graph,
                    checkpoint_scope,
                    _canceled_state_values(request, checkpoint_values),
                )
            self._checkpoint_runtime.record_thread_scope(checkpoint_scope, status="canceled")
        except CheckpointRuntimeError:
            log.exception("langgraph checkpoint cancel failed", extra={"thread_id": thread_id})
            return agent_runtime_pb2.CancelRunResponse(ok=False, status="checkpoint_error")
        except Exception:
            log.exception("langgraph cancel failed", extra={"thread_id": thread_id})
            with suppress(CheckpointRuntimeError):
                self._checkpoint_runtime.record_thread_scope(checkpoint_scope, status="failed")
            return agent_runtime_pb2.CancelRunResponse(ok=False, status="cancel_error")
        return agent_runtime_pb2.CancelRunResponse(ok=True, status="canceled")

    def GetThreadState(
        self,
        request: agent_runtime_pb2.GetThreadStateRequest,
        context: grpc.ServicerContext,
    ) -> agent_runtime_pb2.ThreadStateResponse:
        del context
        thread_id = request.thread_id or f"pulse:{request.project_id}:{request.session_id}"
        company_id = request.auth_context.company_id if request.auth_context.company_id else ""
        checkpoint_scope = self._checkpoint_runtime.scope(company_id, request.project_id, request.session_id, thread_id)
        checkpoint_values: dict[str, Any] = {}
        checkpoint_next: list[str] = []
        checkpoint_found = False
        try:
            checkpoint_snapshot = self._checkpoint_runtime.get_state(self._graph, checkpoint_scope)
            if checkpoint_snapshot is not None:
                checkpoint_found = True
                raw_values = getattr(checkpoint_snapshot, "values", {}) or {}
                if isinstance(raw_values, dict):
                    checkpoint_values = raw_values
                checkpoint_next = [str(value) for value in (getattr(checkpoint_snapshot, "next", ()) or [])]
        except CheckpointRuntimeError as exc:
            checkpoint_values = {"last_error": str(exc)}
        dialogue_state = migrate_dialogue_state(checkpoint_values.get("dialogue_state"), checkpoint_values)
        state = _struct(
            {
                "ready": True,
                "runtime": "langgraph",
                "service_version": _package_version("pulse-ai-agent"),
                "python_version": sys.version.split()[0],
                "langgraph_version": _package_version("langgraph"),
                "grpcio_version": _package_version("grpcio"),
                "project_id": request.project_id,
                "session_id": request.session_id,
                "checkpoint_enabled": self._checkpoint_runtime.enabled,
                "capability_matrix": dependency_capability_matrix(
                    self._settings,
                    checkpoint_enabled=self._checkpoint_runtime.enabled,
                ),
                "checkpoint_namespace": checkpoint_scope.checkpoint_namespace,
                "checkpoint_thread_id": checkpoint_scope.checkpoint_thread_id,
                "checkpoint_found": checkpoint_found,
                "state_schema_version": STATE_SCHEMA_VERSION,
                "thread_revision": _int_value(checkpoint_values.get("thread_revision")),
                "last_turn_idempotency_key": str(checkpoint_values.get("last_turn_idempotency_key") or ""),
                "current_step": str(checkpoint_values.get("current_step") or ""),
                "active_mode": str(checkpoint_values.get("active_mode") or ""),
                "dialogue_state": _dict_value(dialogue_state),
                "active_topic": _dict_value(
                    dialogue_state.get("active_topic") or checkpoint_values.get("active_topic")
                ),
                "suspended_topics": _list_value(
                    dialogue_state.get("suspended_topics") or checkpoint_values.get("suspended_topics")
                ),
                "resolved_entities": _list_value(
                    dialogue_state.get("resolved_entities") or checkpoint_values.get("resolved_entities")
                ),
                "unresolved_slots": _dict_value(
                    dialogue_state.get("unresolved_slots") or checkpoint_values.get("unresolved_slots")
                ),
                "active_card_id": str(checkpoint_values.get("active_card_id") or ""),
                "active_tool_call_id": str(checkpoint_values.get("active_tool_call_id") or ""),
                "active_proposal_id": str(checkpoint_values.get("active_proposal_id") or ""),
                "pending_interrupt": bool(checkpoint_values.get("pending_interrupt")),
                "cancel_requested": bool(checkpoint_values.get("cancel_requested")),
                "canceled_at": str(checkpoint_values.get("canceled_at") or ""),
                "cancel_reason": str(checkpoint_values.get("cancel_reason") or ""),
                "ai_observability": _dict_value(checkpoint_values.get("ai_observability_json")),
                "next_nodes": checkpoint_next,
                "last_error": str(checkpoint_values.get("last_error") or ""),
            }
        )
        return agent_runtime_pb2.ThreadStateResponse(
            thread_id=thread_id,
            run_id="",
            state_json=state,
            cards=[],
        )


def serve(settings: Settings) -> None:
    validate_settings(settings)
    _validate_security(settings)
    bind_addr = _bind_addr(settings)
    interceptors = [ServiceTokenInterceptor(settings.service_token)] if settings.service_token else []
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=16), interceptors=interceptors)
    runtime = AgentRuntime(settings)
    agent_runtime_pb2_grpc.add_AgentRuntimeServiceServicer_to_server(runtime, server)
    health_servicer = health.HealthServicer()
    health_pb2_grpc.add_HealthServicer_to_server(health_servicer, server)
    health_servicer.set("", health_pb2.HealthCheckResponse.SERVING)
    health_servicer.set("ai_agent.v1.AgentRuntimeService", health_pb2.HealthCheckResponse.SERVING)
    server.add_insecure_port(bind_addr)
    server.start()
    log.info(
        "pulse ai agent grpc server started",
        extra={
            "grpc_addr": settings.grpc_addr,
            "tool_gateway": settings.pulse_tool_grpc_addr,
            "model_gateway": settings.pulse_model_grpc_addr,
            "insecure_dev": settings.insecure_dev,
            "service_token_required": bool(settings.service_token),
        },
    )
    try:
        server.wait_for_termination()
    finally:
        runtime.close()


def _validate_security(settings: Settings) -> None:
    if settings.insecure_dev:
        return
    if not settings.service_token:
        raise RuntimeError("AI_AGENT_SERVICE_TOKEN is required unless AI_AGENT_INSECURE_DEV=true")
    bind_addr = _bind_addr(settings)
    if not _is_loopback_or_unix(bind_addr) and not _bool_env("AI_AGENT_GRPC_ALLOW_REMOTE_TOKEN"):
        raise RuntimeError(
            "AI agent token gRPC bind must be loopback or unix socket unless AI_AGENT_GRPC_ALLOW_REMOTE_TOKEN=true"
        )


def _bind_addr(settings: Settings) -> str:
    bind_addr = settings.grpc_addr
    if bind_addr.startswith(":"):
        if settings.insecure_dev and not settings.insecure_bind_all:
            return "127.0.0.1" + bind_addr
        if settings.insecure_dev and settings.insecure_bind_all:
            return f"[::]{bind_addr}"
        return "127.0.0.1" + bind_addr
    return bind_addr


def _is_loopback_or_unix(bind_addr: str) -> bool:
    if bind_addr.startswith("unix:"):
        return True
    host = bind_addr
    if host.startswith("[") and "]" in host:
        host = host[1 : host.index("]")]
    elif ":" in host:
        host = host.rsplit(":", 1)[0]
    host = host.strip()
    if host in {"", "localhost"}:
        return True
    try:
        return ip_address(host).is_loopback
    except ValueError:
        return False


def _bool_env(key: str) -> bool:
    return os.environ.get(key, "").strip().lower() in {"1", "true", "yes", "on"}


def _unauthenticated_handler(handler: grpc.RpcMethodHandler) -> grpc.RpcMethodHandler:
    def abort_unary_unary(request: object, context: grpc.ServicerContext) -> object:
        del request
        context.abort(grpc.StatusCode.UNAUTHENTICATED, "invalid service token")
        raise RuntimeError("grpc abort returned unexpectedly")

    def abort_unary_stream(request: object, context: grpc.ServicerContext) -> Iterator[object]:
        del request
        context.abort(grpc.StatusCode.UNAUTHENTICATED, "invalid service token")
        yield from ()

    if handler.unary_stream:
        return grpc.unary_stream_rpc_method_handler(
            abort_unary_stream,
            request_deserializer=handler.request_deserializer,
            response_serializer=handler.response_serializer,
        )
    return grpc.unary_unary_rpc_method_handler(
        abort_unary_unary,
        request_deserializer=handler.request_deserializer,
        response_serializer=handler.response_serializer,
    )


def _events_with_execution(result: Any, previous_event_count: int) -> list[dict[str, object]]:
    events = _new_graph_events(result, previous_event_count)
    if events and isinstance(result, dict):
        last = events[-1]
        metadata = {**_dict_value(last.get("tool_call")), "execution": execution_summary(result)}
        events[-1] = {**last, "tool_call": metadata}
    return events


def _event(
    seq: int,
    run_id: str,
    thread_id: str,
    session_id: str,
    event_type: int,
    message: str,
    *,
    card: dict[str, object] | None = None,
    tool_call: dict[str, object] | None = None,
    interrupt: dict[str, object] | None = None,
    error: dict[str, object] | None = None,
) -> agent_runtime_pb2.AgentEvent:
    ts = Timestamp()
    ts.GetCurrentTime()
    return agent_runtime_pb2.AgentEvent(
        seq=seq,
        run_id=run_id,
        thread_id=thread_id,
        session_id=session_id,
        event_type=event_type,
        created_at=ts,
        message=message,
        card=_struct(card or {}),
        tool_call=_struct(tool_call or {}),
        interrupt=_struct(interrupt or {}),
        error=_struct(error or {}),
    )


def _event_from_graph(
    event: dict[str, object],
    *,
    run_id: str,
    thread_id: str,
    session_id: str,
) -> agent_runtime_pb2.AgentEvent:
    event_name = str(event.get("event_type") or "")
    event_type = _PROTO_EVENT_TYPES.get(event_name, agent_runtime_pb2.AGENT_EVENT_TYPE_STATUS)
    seq = _int_value(event.get("seq"))
    if seq <= 0:
        seq = 1
    return _event(
        seq,
        str(event.get("run_id") or run_id),
        str(event.get("thread_id") or thread_id),
        str(event.get("session_id") or session_id),
        event_type,
        str(event.get("message") or ""),
        card=_dict_value(event.get("card")),
        tool_call=_dict_value(event.get("tool_call")),
        interrupt=_dict_value(event.get("interrupt")),
        error=_dict_value(event.get("error")),
    )


def _chat_message_to_state(message: agent_runtime_pb2.ChatMessage) -> dict[str, object]:
    return {
        "message_id": message.message_id,
        "role": message.role,
        "body": message.body,
        "payload_json": json_format.MessageToDict(message.payload_json),
        "created_at": message.created_at.ToJsonString() if message.created_at else "",
    }


def _trace_context_to_state(trace_context: agent_runtime_pb2.TraceContext) -> dict[str, object]:
    return {
        "trace_id": trace_context.trace_id,
        "span_id": trace_context.span_id,
        "request_id": trace_context.request_id,
        "baggage": dict(trace_context.baggage),
    }


def _thread_policy_error(
    seq: int,
    run_id: str,
    thread_id: str,
    session_id: str,
    code: str,
    message: str,
    checkpoint_values: dict[str, Any],
    checkpoint_next: list[str],
    *,
    expected_revision: int,
) -> agent_runtime_pb2.AgentEvent:
    latest = latest_thread_state(checkpoint_values, checkpoint_next)
    return _event(
        seq,
        run_id,
        thread_id,
        session_id,
        agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR,
        message,
        error={
            "code": code,
            "expected_thread_revision": expected_revision,
            "actual_thread_revision": _int_value(checkpoint_values.get("thread_revision")),
            "latest_state": latest,
        },
    )


def _expected_revision_from_trace_context(trace_context: agent_runtime_pb2.TraceContext) -> int:
    for key, value in dict(trace_context.baggage).items():
        if key in {"expectedThreadRevision", "expected_thread_revision", "threadRevision", "thread_revision"}:
            return _int_value(value)
    return 0


def _dict_value(value: object) -> dict[str, object]:
    if isinstance(value, dict):
        return value
    return {}


def _list_value(value: object) -> list[object]:
    if isinstance(value, list):
        return value
    return []


def _int_value(value: object) -> int:
    if isinstance(value, int):
        return value
    if isinstance(value, str) and value.isdigit():
        return int(value)
    return 0


def _snapshot_values(snapshot: object | None) -> dict[str, Any]:
    values = getattr(snapshot, "values", {}) if snapshot is not None else {}
    return values if isinstance(values, dict) else {}


def _resume_payload(
    request: agent_runtime_pb2.ResumeRequest,
    checkpoint_values: dict[str, Any],
) -> dict[str, object]:
    raw_active_card = checkpoint_values.get("active_card")
    active_card: dict[str, Any] = raw_active_card if isinstance(raw_active_card, dict) else {}
    raw_pending = checkpoint_values.get("pending_interrupt")
    pending: dict[str, Any] = raw_pending if isinstance(raw_pending, dict) else {}
    edited = json_format.MessageToDict(request.edited_card_json) if request.HasField("edited_card_json") else {}
    action = request.decision.action.strip().lower() or "approve"
    if action in {"apply", "approved"}:
        action = "approve"
    if action in {"rejected"}:
        action = "reject"
    if action in {"canceled", "cancelled"}:
        action = "cancel"
    active_tool_call_id = request.interrupt_id or str(pending.get("active_tool_call_id") or "")
    return {
        "schema": "riders-pulse-resume-decision-v1",
        "request_id": request.request_id,
        "idempotency_key": request.idempotency_key,
        "action": action,
        "reason": request.decision.reason,
        "card_id": str(active_card.get("card_id") or pending.get("card_id") or ""),
        "card_revision": int(request.decision.card_revision),
        "interrupt_id": request.interrupt_id or str(pending.get("interrupt_id") or ""),
        "active_tool_call_id": active_tool_call_id,
        "edited_card_json": edited,
        "local_date": request.local_date,
        "timezone": request.timezone,
        "checkpoint_status": "resuming",
        "trace_context": _trace_context_to_state(request.trace_context),
        "observability_started_ms": monotonic_ms(),
    }


def _cancel_resume_payload(
    request: agent_runtime_pb2.CancelRunRequest,
    checkpoint_values: dict[str, Any],
) -> dict[str, object]:
    raw_pending = checkpoint_values.get("pending_interrupt")
    pending = raw_pending if isinstance(raw_pending, dict) else {}
    return {
        "schema": "riders-pulse-resume-decision-v1",
        "request_id": request.request_id,
        "idempotency_key": request.request_id or request.run_id,
        "action": "cancel",
        "reason": "user_cancel",
        "card_id": str(pending.get("card_id") or ""),
        "card_revision": _int_value(pending.get("card_revision")),
        "interrupt_id": str(pending.get("interrupt_id") or request.run_id or ""),
        "active_tool_call_id": str(pending.get("active_tool_call_id") or ""),
    }


def _canceled_state_values(
    request: agent_runtime_pb2.CancelRunRequest,
    checkpoint_values: dict[str, Any],
) -> dict[str, object]:
    active_card = (
        {**checkpoint_values["active_card"], "status": "canceled"}
        if isinstance(checkpoint_values.get("active_card"), dict)
        else None
    )
    values: dict[str, object] = {
        "cancel_requested": True,
        "cancel_reason": "user_cancel",
        "canceled_at": datetime.now(UTC).isoformat(),
        "current_step": "canceled",
        "active_mode": "canceled",
        "pending_interrupt": None,
        "awaiting_tool_observation": False,
        "run_id": request.run_id or checkpoint_values.get("run_id") or "",
    }
    if active_card is not None:
        values["active_card"] = active_card
    return values


def _has_interrupt(result: object) -> bool:
    return isinstance(result, dict) and bool(result.get("__interrupt__"))


def _set_ai_observability_checkpoint_status(result: object, status: str) -> None:
    if not isinstance(result, dict):
        return
    observability = result.get("ai_observability_json")
    if not isinstance(observability, dict):
        return
    checkpoint = observability.get("checkpoint")
    if isinstance(checkpoint, dict):
        checkpoint["status"] = status
    events = result.get("events")
    if not isinstance(events, list):
        return
    for event in events:
        if not isinstance(event, dict) or event.get("event_type") != "DONE":
            continue
        tool_call = event.get("tool_call")
        if not isinstance(tool_call, dict):
            continue
        event_observability = tool_call.get("ai_observability")
        if event_observability is observability:
            continue
        if isinstance(event_observability, dict):
            event_checkpoint = event_observability.get("checkpoint")
            if isinstance(event_checkpoint, dict):
                event_checkpoint["status"] = status


def _new_graph_events(result: object, previous_event_count: int) -> list[dict[str, object]]:
    if not isinstance(result, dict):
        return []
    events = result.get("events")
    if not isinstance(events, list):
        return []
    out: list[dict[str, object]] = []
    for event in events[max(previous_event_count, 0) :]:
        if isinstance(event, dict):
            out.append(event)
    return out


def _interrupt_events(
    result: object,
    *,
    run_id: str,
    thread_id: str,
    session_id: str,
    previous_event_count: int = 0,
) -> Iterator[agent_runtime_pb2.AgentEvent]:
    if not isinstance(result, dict):
        return
    interrupts = result.get("__interrupt__")
    if not isinstance(interrupts, list) or not interrupts:
        return
    raw_events = result.get("events")
    events = raw_events if isinstance(raw_events, list) else []
    new_events = events[max(previous_event_count, 0) :]
    if any(
        isinstance(event, dict) and event.get("event_type") in {"INTERRUPT", "APPROVAL_REQUIRED"}
        for event in new_events
    ):
        return
    max_seq = 0
    for event in new_events:
        if isinstance(event, dict):
            max_seq = max(max_seq, _int_value(event.get("seq")))
    for idx, item in enumerate(interrupts, start=1):
        value = getattr(item, "value", {})
        interrupt_id = str(getattr(item, "id", "") or "")
        payload = value if isinstance(value, dict) else {"value": str(value)}
        if interrupt_id:
            payload = {**payload, "langgraph_interrupt_id": interrupt_id}
        yield _event(
            max_seq + idx,
            run_id,
            thread_id,
            session_id,
            agent_runtime_pb2.AGENT_EVENT_TYPE_INTERRUPT,
            "Жду подтверждения",
            interrupt=payload,
        )


_PROTO_EVENT_TYPES = {
    "RUN_STARTED": agent_runtime_pb2.AGENT_EVENT_TYPE_RUN_STARTED,
    "STATUS": agent_runtime_pb2.AGENT_EVENT_TYPE_STATUS,
    "MESSAGE_DELTA": agent_runtime_pb2.AGENT_EVENT_TYPE_MESSAGE_DELTA,
    "MESSAGE_FINAL": agent_runtime_pb2.AGENT_EVENT_TYPE_MESSAGE_FINAL,
    "CARD_CREATED": agent_runtime_pb2.AGENT_EVENT_TYPE_CARD_CREATED,
    "CARD_UPDATED": agent_runtime_pb2.AGENT_EVENT_TYPE_CARD_UPDATED,
    "CARD_PATCHED": agent_runtime_pb2.AGENT_EVENT_TYPE_CARD_PATCHED,
    "TOOL_PROPOSED": agent_runtime_pb2.AGENT_EVENT_TYPE_TOOL_PROPOSED,
    "TOOL_STARTED": agent_runtime_pb2.AGENT_EVENT_TYPE_TOOL_STARTED,
    "TOOL_RESULT": agent_runtime_pb2.AGENT_EVENT_TYPE_TOOL_RESULT,
    "INTERRUPT": agent_runtime_pb2.AGENT_EVENT_TYPE_INTERRUPT,
    "APPROVAL_REQUIRED": agent_runtime_pb2.AGENT_EVENT_TYPE_APPROVAL_REQUIRED,
    "MEMORY_UPDATED": agent_runtime_pb2.AGENT_EVENT_TYPE_MEMORY_UPDATED,
    "ERROR": agent_runtime_pb2.AGENT_EVENT_TYPE_ERROR,
    "DONE": agent_runtime_pb2.AGENT_EVENT_TYPE_DONE,
}


def _struct(value: dict[str, object]) -> Struct:
    out = Struct()
    json_format.ParseDict(value, out)
    return out


def _package_version(name: str) -> str:
    try:
        return metadata.version(name)
    except metadata.PackageNotFoundError:
        return "unknown"
