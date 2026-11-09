import datetime

from ai_agent.v1 import common_pb2 as _common_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AgentEventType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    AGENT_EVENT_TYPE_UNSPECIFIED: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_RUN_STARTED: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_STATUS: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_MESSAGE_DELTA: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_MESSAGE_FINAL: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_CARD_CREATED: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_CARD_UPDATED: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_CARD_PATCHED: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_TOOL_PROPOSED: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_TOOL_STARTED: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_TOOL_RESULT: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_INTERRUPT: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_APPROVAL_REQUIRED: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_MEMORY_UPDATED: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_ERROR: _ClassVar[AgentEventType]
    AGENT_EVENT_TYPE_DONE: _ClassVar[AgentEventType]
AGENT_EVENT_TYPE_UNSPECIFIED: AgentEventType
AGENT_EVENT_TYPE_RUN_STARTED: AgentEventType
AGENT_EVENT_TYPE_STATUS: AgentEventType
AGENT_EVENT_TYPE_MESSAGE_DELTA: AgentEventType
AGENT_EVENT_TYPE_MESSAGE_FINAL: AgentEventType
AGENT_EVENT_TYPE_CARD_CREATED: AgentEventType
AGENT_EVENT_TYPE_CARD_UPDATED: AgentEventType
AGENT_EVENT_TYPE_CARD_PATCHED: AgentEventType
AGENT_EVENT_TYPE_TOOL_PROPOSED: AgentEventType
AGENT_EVENT_TYPE_TOOL_STARTED: AgentEventType
AGENT_EVENT_TYPE_TOOL_RESULT: AgentEventType
AGENT_EVENT_TYPE_INTERRUPT: AgentEventType
AGENT_EVENT_TYPE_APPROVAL_REQUIRED: AgentEventType
AGENT_EVENT_TYPE_MEMORY_UPDATED: AgentEventType
AGENT_EVENT_TYPE_ERROR: AgentEventType
AGENT_EVENT_TYPE_DONE: AgentEventType

class TurnRequest(_message.Message):
    __slots__ = ("request_id", "idempotency_key", "company_id", "project_id", "user_id", "session_id", "thread_id", "message_id", "body", "local_date", "timezone", "locale", "page_context_json", "client_context_json", "ui_state_json", "active_card_id", "active_tool_call_id", "active_proposal_id", "history", "runtime_policy", "auth_context", "trace_context")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    IDEMPOTENCY_KEY_FIELD_NUMBER: _ClassVar[int]
    COMPANY_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    THREAD_ID_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_ID_FIELD_NUMBER: _ClassVar[int]
    BODY_FIELD_NUMBER: _ClassVar[int]
    LOCAL_DATE_FIELD_NUMBER: _ClassVar[int]
    TIMEZONE_FIELD_NUMBER: _ClassVar[int]
    LOCALE_FIELD_NUMBER: _ClassVar[int]
    PAGE_CONTEXT_JSON_FIELD_NUMBER: _ClassVar[int]
    CLIENT_CONTEXT_JSON_FIELD_NUMBER: _ClassVar[int]
    UI_STATE_JSON_FIELD_NUMBER: _ClassVar[int]
    ACTIVE_CARD_ID_FIELD_NUMBER: _ClassVar[int]
    ACTIVE_TOOL_CALL_ID_FIELD_NUMBER: _ClassVar[int]
    ACTIVE_PROPOSAL_ID_FIELD_NUMBER: _ClassVar[int]
    HISTORY_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_POLICY_FIELD_NUMBER: _ClassVar[int]
    AUTH_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    TRACE_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    idempotency_key: str
    company_id: str
    project_id: str
    user_id: str
    session_id: str
    thread_id: str
    message_id: str
    body: str
    local_date: str
    timezone: str
    locale: str
    page_context_json: _struct_pb2.Struct
    client_context_json: _struct_pb2.Struct
    ui_state_json: _struct_pb2.Struct
    active_card_id: str
    active_tool_call_id: str
    active_proposal_id: str
    history: _containers.RepeatedCompositeFieldContainer[_common_pb2.ChatMessage]
    runtime_policy: _common_pb2.RuntimePolicy
    auth_context: _common_pb2.AuthContext
    trace_context: _common_pb2.TraceContext
    def __init__(self, request_id: _Optional[str] = ..., idempotency_key: _Optional[str] = ..., company_id: _Optional[str] = ..., project_id: _Optional[str] = ..., user_id: _Optional[str] = ..., session_id: _Optional[str] = ..., thread_id: _Optional[str] = ..., message_id: _Optional[str] = ..., body: _Optional[str] = ..., local_date: _Optional[str] = ..., timezone: _Optional[str] = ..., locale: _Optional[str] = ..., page_context_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., client_context_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., ui_state_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., active_card_id: _Optional[str] = ..., active_tool_call_id: _Optional[str] = ..., active_proposal_id: _Optional[str] = ..., history: _Optional[_Iterable[_Union[_common_pb2.ChatMessage, _Mapping]]] = ..., runtime_policy: _Optional[_Union[_common_pb2.RuntimePolicy, _Mapping]] = ..., auth_context: _Optional[_Union[_common_pb2.AuthContext, _Mapping]] = ..., trace_context: _Optional[_Union[_common_pb2.TraceContext, _Mapping]] = ...) -> None: ...

class ApprovalDecision(_message.Message):
    __slots__ = ("action", "reason", "card_revision")
    ACTION_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    CARD_REVISION_FIELD_NUMBER: _ClassVar[int]
    action: str
    reason: str
    card_revision: int
    def __init__(self, action: _Optional[str] = ..., reason: _Optional[str] = ..., card_revision: _Optional[int] = ...) -> None: ...

class ResumeRequest(_message.Message):
    __slots__ = ("request_id", "idempotency_key", "company_id", "project_id", "user_id", "session_id", "thread_id", "run_id", "interrupt_id", "decision", "edited_card_json", "local_date", "timezone", "auth_context", "trace_context")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    IDEMPOTENCY_KEY_FIELD_NUMBER: _ClassVar[int]
    COMPANY_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    THREAD_ID_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    INTERRUPT_ID_FIELD_NUMBER: _ClassVar[int]
    DECISION_FIELD_NUMBER: _ClassVar[int]
    EDITED_CARD_JSON_FIELD_NUMBER: _ClassVar[int]
    LOCAL_DATE_FIELD_NUMBER: _ClassVar[int]
    TIMEZONE_FIELD_NUMBER: _ClassVar[int]
    AUTH_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    TRACE_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    idempotency_key: str
    company_id: str
    project_id: str
    user_id: str
    session_id: str
    thread_id: str
    run_id: str
    interrupt_id: str
    decision: ApprovalDecision
    edited_card_json: _struct_pb2.Struct
    local_date: str
    timezone: str
    auth_context: _common_pb2.AuthContext
    trace_context: _common_pb2.TraceContext
    def __init__(self, request_id: _Optional[str] = ..., idempotency_key: _Optional[str] = ..., company_id: _Optional[str] = ..., project_id: _Optional[str] = ..., user_id: _Optional[str] = ..., session_id: _Optional[str] = ..., thread_id: _Optional[str] = ..., run_id: _Optional[str] = ..., interrupt_id: _Optional[str] = ..., decision: _Optional[_Union[ApprovalDecision, _Mapping]] = ..., edited_card_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., local_date: _Optional[str] = ..., timezone: _Optional[str] = ..., auth_context: _Optional[_Union[_common_pb2.AuthContext, _Mapping]] = ..., trace_context: _Optional[_Union[_common_pb2.TraceContext, _Mapping]] = ...) -> None: ...

class CancelRunRequest(_message.Message):
    __slots__ = ("request_id", "project_id", "session_id", "thread_id", "run_id", "auth_context", "trace_context")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    THREAD_ID_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    AUTH_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    TRACE_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    project_id: str
    session_id: str
    thread_id: str
    run_id: str
    auth_context: _common_pb2.AuthContext
    trace_context: _common_pb2.TraceContext
    def __init__(self, request_id: _Optional[str] = ..., project_id: _Optional[str] = ..., session_id: _Optional[str] = ..., thread_id: _Optional[str] = ..., run_id: _Optional[str] = ..., auth_context: _Optional[_Union[_common_pb2.AuthContext, _Mapping]] = ..., trace_context: _Optional[_Union[_common_pb2.TraceContext, _Mapping]] = ...) -> None: ...

class CancelRunResponse(_message.Message):
    __slots__ = ("ok", "status")
    OK_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ok: bool
    status: str
    def __init__(self, ok: _Optional[bool] = ..., status: _Optional[str] = ...) -> None: ...

class GetThreadStateRequest(_message.Message):
    __slots__ = ("request_id", "project_id", "session_id", "thread_id", "auth_context", "trace_context")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    THREAD_ID_FIELD_NUMBER: _ClassVar[int]
    AUTH_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    TRACE_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    project_id: str
    session_id: str
    thread_id: str
    auth_context: _common_pb2.AuthContext
    trace_context: _common_pb2.TraceContext
    def __init__(self, request_id: _Optional[str] = ..., project_id: _Optional[str] = ..., session_id: _Optional[str] = ..., thread_id: _Optional[str] = ..., auth_context: _Optional[_Union[_common_pb2.AuthContext, _Mapping]] = ..., trace_context: _Optional[_Union[_common_pb2.TraceContext, _Mapping]] = ...) -> None: ...

class ThreadStateResponse(_message.Message):
    __slots__ = ("thread_id", "run_id", "state_json", "cards")
    THREAD_ID_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    STATE_JSON_FIELD_NUMBER: _ClassVar[int]
    CARDS_FIELD_NUMBER: _ClassVar[int]
    thread_id: str
    run_id: str
    state_json: _struct_pb2.Struct
    cards: _containers.RepeatedCompositeFieldContainer[_struct_pb2.Struct]
    def __init__(self, thread_id: _Optional[str] = ..., run_id: _Optional[str] = ..., state_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., cards: _Optional[_Iterable[_Union[_struct_pb2.Struct, _Mapping]]] = ...) -> None: ...

class AgentEvent(_message.Message):
    __slots__ = ("seq", "run_id", "thread_id", "session_id", "event_type", "created_at", "body_delta", "message", "card", "tool_call", "interrupt", "error", "usage", "debug")
    SEQ_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    THREAD_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    BODY_DELTA_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    CARD_FIELD_NUMBER: _ClassVar[int]
    TOOL_CALL_FIELD_NUMBER: _ClassVar[int]
    INTERRUPT_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    USAGE_FIELD_NUMBER: _ClassVar[int]
    DEBUG_FIELD_NUMBER: _ClassVar[int]
    seq: int
    run_id: str
    thread_id: str
    session_id: str
    event_type: AgentEventType
    created_at: _timestamp_pb2.Timestamp
    body_delta: str
    message: str
    card: _struct_pb2.Struct
    tool_call: _struct_pb2.Struct
    interrupt: _struct_pb2.Struct
    error: _struct_pb2.Struct
    usage: _common_pb2.Usage
    debug: _struct_pb2.Struct
    def __init__(self, seq: _Optional[int] = ..., run_id: _Optional[str] = ..., thread_id: _Optional[str] = ..., session_id: _Optional[str] = ..., event_type: _Optional[_Union[AgentEventType, str]] = ..., created_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., body_delta: _Optional[str] = ..., message: _Optional[str] = ..., card: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., tool_call: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., interrupt: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., error: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., usage: _Optional[_Union[_common_pb2.Usage, _Mapping]] = ..., debug: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...
