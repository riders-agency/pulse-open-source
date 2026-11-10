from ai_agent.v1 import common_pb2 as _common_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ListToolsRequest(_message.Message):
    __slots__ = ("request_id", "company_id", "project_id", "user_id", "auth_context", "trace_context")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    COMPANY_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    AUTH_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    TRACE_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    company_id: str
    project_id: str
    user_id: str
    auth_context: _common_pb2.AuthContext
    trace_context: _common_pb2.TraceContext
    def __init__(self, request_id: _Optional[str] = ..., company_id: _Optional[str] = ..., project_id: _Optional[str] = ..., user_id: _Optional[str] = ..., auth_context: _Optional[_Union[_common_pb2.AuthContext, _Mapping]] = ..., trace_context: _Optional[_Union[_common_pb2.TraceContext, _Mapping]] = ...) -> None: ...

class ToolSpec(_message.Message):
    __slots__ = ("tool_key", "display_name", "description", "risk_level", "input_schema_json", "output_schema_json", "required_permissions", "requires_approval", "supports_dry_run", "idempotency_required", "ui_card_type", "timeout_ms", "retry_policy")
    TOOL_KEY_FIELD_NUMBER: _ClassVar[int]
    DISPLAY_NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    RISK_LEVEL_FIELD_NUMBER: _ClassVar[int]
    INPUT_SCHEMA_JSON_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_SCHEMA_JSON_FIELD_NUMBER: _ClassVar[int]
    REQUIRED_PERMISSIONS_FIELD_NUMBER: _ClassVar[int]
    REQUIRES_APPROVAL_FIELD_NUMBER: _ClassVar[int]
    SUPPORTS_DRY_RUN_FIELD_NUMBER: _ClassVar[int]
    IDEMPOTENCY_REQUIRED_FIELD_NUMBER: _ClassVar[int]
    UI_CARD_TYPE_FIELD_NUMBER: _ClassVar[int]
    TIMEOUT_MS_FIELD_NUMBER: _ClassVar[int]
    RETRY_POLICY_FIELD_NUMBER: _ClassVar[int]
    tool_key: str
    display_name: str
    description: str
    risk_level: _common_pb2.RiskLevel
    input_schema_json: _struct_pb2.Struct
    output_schema_json: _struct_pb2.Struct
    required_permissions: _containers.RepeatedScalarFieldContainer[str]
    requires_approval: bool
    supports_dry_run: bool
    idempotency_required: bool
    ui_card_type: str
    timeout_ms: int
    retry_policy: _common_pb2.ToolRetryPolicy
    def __init__(self, tool_key: _Optional[str] = ..., display_name: _Optional[str] = ..., description: _Optional[str] = ..., risk_level: _Optional[_Union[_common_pb2.RiskLevel, str]] = ..., input_schema_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., output_schema_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., required_permissions: _Optional[_Iterable[str]] = ..., requires_approval: _Optional[bool] = ..., supports_dry_run: _Optional[bool] = ..., idempotency_required: _Optional[bool] = ..., ui_card_type: _Optional[str] = ..., timeout_ms: _Optional[int] = ..., retry_policy: _Optional[_Union[_common_pb2.ToolRetryPolicy, _Mapping]] = ...) -> None: ...

class ListToolsResponse(_message.Message):
    __slots__ = ("tools",)
    TOOLS_FIELD_NUMBER: _ClassVar[int]
    tools: _containers.RepeatedCompositeFieldContainer[ToolSpec]
    def __init__(self, tools: _Optional[_Iterable[_Union[ToolSpec, _Mapping]]] = ...) -> None: ...

class ReadContextRequest(_message.Message):
    __slots__ = ("request_id", "company_id", "project_id", "user_id", "session_id", "page_context_json", "auth_context", "trace_context")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    COMPANY_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    PAGE_CONTEXT_JSON_FIELD_NUMBER: _ClassVar[int]
    AUTH_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    TRACE_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    company_id: str
    project_id: str
    user_id: str
    session_id: str
    page_context_json: _struct_pb2.Struct
    auth_context: _common_pb2.AuthContext
    trace_context: _common_pb2.TraceContext
    def __init__(self, request_id: _Optional[str] = ..., company_id: _Optional[str] = ..., project_id: _Optional[str] = ..., user_id: _Optional[str] = ..., session_id: _Optional[str] = ..., page_context_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., auth_context: _Optional[_Union[_common_pb2.AuthContext, _Mapping]] = ..., trace_context: _Optional[_Union[_common_pb2.TraceContext, _Mapping]] = ...) -> None: ...

class ReadContextResponse(_message.Message):
    __slots__ = ("context_json", "allowed_tools")
    CONTEXT_JSON_FIELD_NUMBER: _ClassVar[int]
    ALLOWED_TOOLS_FIELD_NUMBER: _ClassVar[int]
    context_json: _struct_pb2.Struct
    allowed_tools: _containers.RepeatedCompositeFieldContainer[ToolSpec]
    def __init__(self, context_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., allowed_tools: _Optional[_Iterable[_Union[ToolSpec, _Mapping]]] = ...) -> None: ...

class ValidateToolCallRequest(_message.Message):
    __slots__ = ("request_id", "company_id", "project_id", "user_id", "session_id", "thread_id", "tool_call_id", "tool_key", "input_json", "dry_run", "auth_context", "trace_context")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    COMPANY_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    THREAD_ID_FIELD_NUMBER: _ClassVar[int]
    TOOL_CALL_ID_FIELD_NUMBER: _ClassVar[int]
    TOOL_KEY_FIELD_NUMBER: _ClassVar[int]
    INPUT_JSON_FIELD_NUMBER: _ClassVar[int]
    DRY_RUN_FIELD_NUMBER: _ClassVar[int]
    AUTH_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    TRACE_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    company_id: str
    project_id: str
    user_id: str
    session_id: str
    thread_id: str
    tool_call_id: str
    tool_key: str
    input_json: _struct_pb2.Struct
    dry_run: bool
    auth_context: _common_pb2.AuthContext
    trace_context: _common_pb2.TraceContext
    def __init__(self, request_id: _Optional[str] = ..., company_id: _Optional[str] = ..., project_id: _Optional[str] = ..., user_id: _Optional[str] = ..., session_id: _Optional[str] = ..., thread_id: _Optional[str] = ..., tool_call_id: _Optional[str] = ..., tool_key: _Optional[str] = ..., input_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., dry_run: _Optional[bool] = ..., auth_context: _Optional[_Union[_common_pb2.AuthContext, _Mapping]] = ..., trace_context: _Optional[_Union[_common_pb2.TraceContext, _Mapping]] = ...) -> None: ...

class ValidateToolCallResponse(_message.Message):
    __slots__ = ("ok", "normalized_input_json", "missing_fields", "user_message", "error_code", "error_message")
    OK_FIELD_NUMBER: _ClassVar[int]
    NORMALIZED_INPUT_JSON_FIELD_NUMBER: _ClassVar[int]
    MISSING_FIELDS_FIELD_NUMBER: _ClassVar[int]
    USER_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    ERROR_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    ok: bool
    normalized_input_json: _struct_pb2.Struct
    missing_fields: _containers.RepeatedCompositeFieldContainer[_struct_pb2.Struct]
    user_message: str
    error_code: str
    error_message: str
    def __init__(self, ok: _Optional[bool] = ..., normalized_input_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., missing_fields: _Optional[_Iterable[_Union[_struct_pb2.Struct, _Mapping]]] = ..., user_message: _Optional[str] = ..., error_code: _Optional[str] = ..., error_message: _Optional[str] = ...) -> None: ...

class CreateProposalRequest(_message.Message):
    __slots__ = ("request_id", "company_id", "project_id", "user_id", "session_id", "thread_id", "tool_call_id", "tool_key", "intent_key", "risk_level", "requires_approval", "input_json", "page_context_json", "idempotency_key", "expected_card_type", "parent_plan_id", "step_id", "auth_context", "trace_context")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    COMPANY_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    THREAD_ID_FIELD_NUMBER: _ClassVar[int]
    TOOL_CALL_ID_FIELD_NUMBER: _ClassVar[int]
    TOOL_KEY_FIELD_NUMBER: _ClassVar[int]
    INTENT_KEY_FIELD_NUMBER: _ClassVar[int]
    RISK_LEVEL_FIELD_NUMBER: _ClassVar[int]
    REQUIRES_APPROVAL_FIELD_NUMBER: _ClassVar[int]
    INPUT_JSON_FIELD_NUMBER: _ClassVar[int]
    PAGE_CONTEXT_JSON_FIELD_NUMBER: _ClassVar[int]
    IDEMPOTENCY_KEY_FIELD_NUMBER: _ClassVar[int]
    EXPECTED_CARD_TYPE_FIELD_NUMBER: _ClassVar[int]
    PARENT_PLAN_ID_FIELD_NUMBER: _ClassVar[int]
    STEP_ID_FIELD_NUMBER: _ClassVar[int]
    AUTH_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    TRACE_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    company_id: str
    project_id: str
    user_id: str
    session_id: str
    thread_id: str
    tool_call_id: str
    tool_key: str
    intent_key: str
    risk_level: _common_pb2.RiskLevel
    requires_approval: bool
    input_json: _struct_pb2.Struct
    page_context_json: _struct_pb2.Struct
    idempotency_key: str
    expected_card_type: str
    parent_plan_id: str
    step_id: str
    auth_context: _common_pb2.AuthContext
    trace_context: _common_pb2.TraceContext
    def __init__(self, request_id: _Optional[str] = ..., company_id: _Optional[str] = ..., project_id: _Optional[str] = ..., user_id: _Optional[str] = ..., session_id: _Optional[str] = ..., thread_id: _Optional[str] = ..., tool_call_id: _Optional[str] = ..., tool_key: _Optional[str] = ..., intent_key: _Optional[str] = ..., risk_level: _Optional[_Union[_common_pb2.RiskLevel, str]] = ..., requires_approval: _Optional[bool] = ..., input_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., page_context_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., idempotency_key: _Optional[str] = ..., expected_card_type: _Optional[str] = ..., parent_plan_id: _Optional[str] = ..., step_id: _Optional[str] = ..., auth_context: _Optional[_Union[_common_pb2.AuthContext, _Mapping]] = ..., trace_context: _Optional[_Union[_common_pb2.TraceContext, _Mapping]] = ...) -> None: ...

class PatchProposalRequest(_message.Message):
    __slots__ = ("request_id", "company_id", "project_id", "user_id", "session_id", "thread_id", "tool_call_id", "card_id", "base_revision", "patch_json", "page_context_json", "auth_context", "trace_context")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    COMPANY_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    THREAD_ID_FIELD_NUMBER: _ClassVar[int]
    TOOL_CALL_ID_FIELD_NUMBER: _ClassVar[int]
    CARD_ID_FIELD_NUMBER: _ClassVar[int]
    BASE_REVISION_FIELD_NUMBER: _ClassVar[int]
    PATCH_JSON_FIELD_NUMBER: _ClassVar[int]
    PAGE_CONTEXT_JSON_FIELD_NUMBER: _ClassVar[int]
    AUTH_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    TRACE_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    company_id: str
    project_id: str
    user_id: str
    session_id: str
    thread_id: str
    tool_call_id: str
    card_id: str
    base_revision: int
    patch_json: _struct_pb2.Struct
    page_context_json: _struct_pb2.Struct
    auth_context: _common_pb2.AuthContext
    trace_context: _common_pb2.TraceContext
    def __init__(self, request_id: _Optional[str] = ..., company_id: _Optional[str] = ..., project_id: _Optional[str] = ..., user_id: _Optional[str] = ..., session_id: _Optional[str] = ..., thread_id: _Optional[str] = ..., tool_call_id: _Optional[str] = ..., card_id: _Optional[str] = ..., base_revision: _Optional[int] = ..., patch_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., page_context_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., auth_context: _Optional[_Union[_common_pb2.AuthContext, _Mapping]] = ..., trace_context: _Optional[_Union[_common_pb2.TraceContext, _Mapping]] = ...) -> None: ...

class ApprovalRequest(_message.Message):
    __slots__ = ("request_id", "company_id", "project_id", "user_id", "session_id", "thread_id", "tool_call_id", "card_id", "card_revision", "decision", "idempotency_key", "auth_context", "trace_context")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    COMPANY_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    THREAD_ID_FIELD_NUMBER: _ClassVar[int]
    TOOL_CALL_ID_FIELD_NUMBER: _ClassVar[int]
    CARD_ID_FIELD_NUMBER: _ClassVar[int]
    CARD_REVISION_FIELD_NUMBER: _ClassVar[int]
    DECISION_FIELD_NUMBER: _ClassVar[int]
    IDEMPOTENCY_KEY_FIELD_NUMBER: _ClassVar[int]
    AUTH_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    TRACE_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    company_id: str
    project_id: str
    user_id: str
    session_id: str
    thread_id: str
    tool_call_id: str
    card_id: str
    card_revision: int
    decision: str
    idempotency_key: str
    auth_context: _common_pb2.AuthContext
    trace_context: _common_pb2.TraceContext
    def __init__(self, request_id: _Optional[str] = ..., company_id: _Optional[str] = ..., project_id: _Optional[str] = ..., user_id: _Optional[str] = ..., session_id: _Optional[str] = ..., thread_id: _Optional[str] = ..., tool_call_id: _Optional[str] = ..., card_id: _Optional[str] = ..., card_revision: _Optional[int] = ..., decision: _Optional[str] = ..., idempotency_key: _Optional[str] = ..., auth_context: _Optional[_Union[_common_pb2.AuthContext, _Mapping]] = ..., trace_context: _Optional[_Union[_common_pb2.TraceContext, _Mapping]] = ...) -> None: ...

class ToolProposalResponse(_message.Message):
    __slots__ = ("tool_call_id", "tool_key", "card_id", "card_type", "revision", "status", "input_json", "card_json", "user_message", "error_code", "error_message")
    TOOL_CALL_ID_FIELD_NUMBER: _ClassVar[int]
    TOOL_KEY_FIELD_NUMBER: _ClassVar[int]
    CARD_ID_FIELD_NUMBER: _ClassVar[int]
    CARD_TYPE_FIELD_NUMBER: _ClassVar[int]
    REVISION_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    INPUT_JSON_FIELD_NUMBER: _ClassVar[int]
    CARD_JSON_FIELD_NUMBER: _ClassVar[int]
    USER_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    ERROR_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    tool_call_id: str
    tool_key: str
    card_id: str
    card_type: str
    revision: int
    status: str
    input_json: _struct_pb2.Struct
    card_json: _struct_pb2.Struct
    user_message: str
    error_code: str
    error_message: str
    def __init__(self, tool_call_id: _Optional[str] = ..., tool_key: _Optional[str] = ..., card_id: _Optional[str] = ..., card_type: _Optional[str] = ..., revision: _Optional[int] = ..., status: _Optional[str] = ..., input_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., card_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., user_message: _Optional[str] = ..., error_code: _Optional[str] = ..., error_message: _Optional[str] = ...) -> None: ...

class ExecuteToolRequest(_message.Message):
    __slots__ = ("request_id", "idempotency_key", "company_id", "project_id", "user_id", "session_id", "thread_id", "tool_call_id", "tool_key", "input_json", "approval_token", "dry_run", "auth_context", "trace_context")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    IDEMPOTENCY_KEY_FIELD_NUMBER: _ClassVar[int]
    COMPANY_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    THREAD_ID_FIELD_NUMBER: _ClassVar[int]
    TOOL_CALL_ID_FIELD_NUMBER: _ClassVar[int]
    TOOL_KEY_FIELD_NUMBER: _ClassVar[int]
    INPUT_JSON_FIELD_NUMBER: _ClassVar[int]
    APPROVAL_TOKEN_FIELD_NUMBER: _ClassVar[int]
    DRY_RUN_FIELD_NUMBER: _ClassVar[int]
    AUTH_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    TRACE_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    idempotency_key: str
    company_id: str
    project_id: str
    user_id: str
    session_id: str
    thread_id: str
    tool_call_id: str
    tool_key: str
    input_json: _struct_pb2.Struct
    approval_token: str
    dry_run: bool
    auth_context: _common_pb2.AuthContext
    trace_context: _common_pb2.TraceContext
    def __init__(self, request_id: _Optional[str] = ..., idempotency_key: _Optional[str] = ..., company_id: _Optional[str] = ..., project_id: _Optional[str] = ..., user_id: _Optional[str] = ..., session_id: _Optional[str] = ..., thread_id: _Optional[str] = ..., tool_call_id: _Optional[str] = ..., tool_key: _Optional[str] = ..., input_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., approval_token: _Optional[str] = ..., dry_run: _Optional[bool] = ..., auth_context: _Optional[_Union[_common_pb2.AuthContext, _Mapping]] = ..., trace_context: _Optional[_Union[_common_pb2.TraceContext, _Mapping]] = ...) -> None: ...

class ExecuteToolResponse(_message.Message):
    __slots__ = ("ok", "status", "output_json", "error_code", "error_message", "user_message", "audit_event_id", "entity_links")
    OK_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_JSON_FIELD_NUMBER: _ClassVar[int]
    ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    ERROR_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    USER_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    AUDIT_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    ENTITY_LINKS_FIELD_NUMBER: _ClassVar[int]
    ok: bool
    status: _common_pb2.ToolExecutionStatus
    output_json: _struct_pb2.Struct
    error_code: str
    error_message: str
    user_message: str
    audit_event_id: str
    entity_links: _containers.RepeatedCompositeFieldContainer[_common_pb2.EntityLink]
    def __init__(self, ok: _Optional[bool] = ..., status: _Optional[_Union[_common_pb2.ToolExecutionStatus, str]] = ..., output_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., error_code: _Optional[str] = ..., error_message: _Optional[str] = ..., user_message: _Optional[str] = ..., audit_event_id: _Optional[str] = ..., entity_links: _Optional[_Iterable[_Union[_common_pb2.EntityLink, _Mapping]]] = ...) -> None: ...

class ResolveEntityRequest(_message.Message):
    __slots__ = ("request_id", "company_id", "project_id", "user_id", "entity_type", "query", "limit", "auth_context", "trace_context")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    COMPANY_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    ENTITY_TYPE_FIELD_NUMBER: _ClassVar[int]
    QUERY_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    AUTH_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    TRACE_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    company_id: str
    project_id: str
    user_id: str
    entity_type: str
    query: str
    limit: int
    auth_context: _common_pb2.AuthContext
    trace_context: _common_pb2.TraceContext
    def __init__(self, request_id: _Optional[str] = ..., company_id: _Optional[str] = ..., project_id: _Optional[str] = ..., user_id: _Optional[str] = ..., entity_type: _Optional[str] = ..., query: _Optional[str] = ..., limit: _Optional[int] = ..., auth_context: _Optional[_Union[_common_pb2.AuthContext, _Mapping]] = ..., trace_context: _Optional[_Union[_common_pb2.TraceContext, _Mapping]] = ...) -> None: ...

class ResolveEntityResponse(_message.Message):
    __slots__ = ("matches", "status")
    MATCHES_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    matches: _containers.RepeatedCompositeFieldContainer[_struct_pb2.Struct]
    status: str
    def __init__(self, matches: _Optional[_Iterable[_Union[_struct_pb2.Struct, _Mapping]]] = ..., status: _Optional[str] = ...) -> None: ...
