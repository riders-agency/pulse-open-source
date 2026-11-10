import datetime

from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RiskLevel(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    RISK_LEVEL_UNSPECIFIED: _ClassVar[RiskLevel]
    RISK_LEVEL_READ: _ClassVar[RiskLevel]
    RISK_LEVEL_MEDIUM: _ClassVar[RiskLevel]
    RISK_LEVEL_HIGH: _ClassVar[RiskLevel]
    RISK_LEVEL_CRITICAL: _ClassVar[RiskLevel]

class ToolExecutionStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    TOOL_EXECUTION_STATUS_UNSPECIFIED: _ClassVar[ToolExecutionStatus]
    TOOL_EXECUTION_STATUS_SUCCEEDED: _ClassVar[ToolExecutionStatus]
    TOOL_EXECUTION_STATUS_FAILED: _ClassVar[ToolExecutionStatus]
    TOOL_EXECUTION_STATUS_DENIED: _ClassVar[ToolExecutionStatus]
    TOOL_EXECUTION_STATUS_NEEDS_APPROVAL: _ClassVar[ToolExecutionStatus]
    TOOL_EXECUTION_STATUS_CONFLICT: _ClassVar[ToolExecutionStatus]
RISK_LEVEL_UNSPECIFIED: RiskLevel
RISK_LEVEL_READ: RiskLevel
RISK_LEVEL_MEDIUM: RiskLevel
RISK_LEVEL_HIGH: RiskLevel
RISK_LEVEL_CRITICAL: RiskLevel
TOOL_EXECUTION_STATUS_UNSPECIFIED: ToolExecutionStatus
TOOL_EXECUTION_STATUS_SUCCEEDED: ToolExecutionStatus
TOOL_EXECUTION_STATUS_FAILED: ToolExecutionStatus
TOOL_EXECUTION_STATUS_DENIED: ToolExecutionStatus
TOOL_EXECUTION_STATUS_NEEDS_APPROVAL: ToolExecutionStatus
TOOL_EXECUTION_STATUS_CONFLICT: ToolExecutionStatus

class TraceContext(_message.Message):
    __slots__ = ("trace_id", "span_id", "request_id", "baggage")
    class BaggageEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    TRACE_ID_FIELD_NUMBER: _ClassVar[int]
    SPAN_ID_FIELD_NUMBER: _ClassVar[int]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    BAGGAGE_FIELD_NUMBER: _ClassVar[int]
    trace_id: str
    span_id: str
    request_id: str
    baggage: _containers.ScalarMap[str, str]
    def __init__(self, trace_id: _Optional[str] = ..., span_id: _Optional[str] = ..., request_id: _Optional[str] = ..., baggage: _Optional[_Mapping[str, str]] = ...) -> None: ...

class AuthContext(_message.Message):
    __slots__ = ("company_id", "project_id", "user_id", "permissions", "role", "policy_json")
    COMPANY_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    PERMISSIONS_FIELD_NUMBER: _ClassVar[int]
    ROLE_FIELD_NUMBER: _ClassVar[int]
    POLICY_JSON_FIELD_NUMBER: _ClassVar[int]
    company_id: str
    project_id: str
    user_id: str
    permissions: _containers.RepeatedScalarFieldContainer[str]
    role: str
    policy_json: _struct_pb2.Struct
    def __init__(self, company_id: _Optional[str] = ..., project_id: _Optional[str] = ..., user_id: _Optional[str] = ..., permissions: _Optional[_Iterable[str]] = ..., role: _Optional[str] = ..., policy_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class RuntimePolicy(_message.Message):
    __slots__ = ("runtime", "max_graph_steps", "max_model_calls", "max_tool_calls", "max_write_proposals", "timeout_ms", "allow_read_tools_without_approval")
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    MAX_GRAPH_STEPS_FIELD_NUMBER: _ClassVar[int]
    MAX_MODEL_CALLS_FIELD_NUMBER: _ClassVar[int]
    MAX_TOOL_CALLS_FIELD_NUMBER: _ClassVar[int]
    MAX_WRITE_PROPOSALS_FIELD_NUMBER: _ClassVar[int]
    TIMEOUT_MS_FIELD_NUMBER: _ClassVar[int]
    ALLOW_READ_TOOLS_WITHOUT_APPROVAL_FIELD_NUMBER: _ClassVar[int]
    runtime: str
    max_graph_steps: int
    max_model_calls: int
    max_tool_calls: int
    max_write_proposals: int
    timeout_ms: int
    allow_read_tools_without_approval: bool
    def __init__(self, runtime: _Optional[str] = ..., max_graph_steps: _Optional[int] = ..., max_model_calls: _Optional[int] = ..., max_tool_calls: _Optional[int] = ..., max_write_proposals: _Optional[int] = ..., timeout_ms: _Optional[int] = ..., allow_read_tools_without_approval: _Optional[bool] = ...) -> None: ...

class ChatMessage(_message.Message):
    __slots__ = ("message_id", "role", "body", "payload_json", "created_at")
    MESSAGE_ID_FIELD_NUMBER: _ClassVar[int]
    ROLE_FIELD_NUMBER: _ClassVar[int]
    BODY_FIELD_NUMBER: _ClassVar[int]
    PAYLOAD_JSON_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    message_id: str
    role: str
    body: str
    payload_json: _struct_pb2.Struct
    created_at: _timestamp_pb2.Timestamp
    def __init__(self, message_id: _Optional[str] = ..., role: _Optional[str] = ..., body: _Optional[str] = ..., payload_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., created_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class EntityLink(_message.Message):
    __slots__ = ("entity_type", "entity_id", "label", "url")
    ENTITY_TYPE_FIELD_NUMBER: _ClassVar[int]
    ENTITY_ID_FIELD_NUMBER: _ClassVar[int]
    LABEL_FIELD_NUMBER: _ClassVar[int]
    URL_FIELD_NUMBER: _ClassVar[int]
    entity_type: str
    entity_id: str
    label: str
    url: str
    def __init__(self, entity_type: _Optional[str] = ..., entity_id: _Optional[str] = ..., label: _Optional[str] = ..., url: _Optional[str] = ...) -> None: ...

class Usage(_message.Message):
    __slots__ = ("provider_key", "model_id", "input_tokens", "output_tokens", "latency_ms", "metadata_json")
    PROVIDER_KEY_FIELD_NUMBER: _ClassVar[int]
    MODEL_ID_FIELD_NUMBER: _ClassVar[int]
    INPUT_TOKENS_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_TOKENS_FIELD_NUMBER: _ClassVar[int]
    LATENCY_MS_FIELD_NUMBER: _ClassVar[int]
    METADATA_JSON_FIELD_NUMBER: _ClassVar[int]
    provider_key: str
    model_id: str
    input_tokens: int
    output_tokens: int
    latency_ms: int
    metadata_json: _struct_pb2.Struct
    def __init__(self, provider_key: _Optional[str] = ..., model_id: _Optional[str] = ..., input_tokens: _Optional[int] = ..., output_tokens: _Optional[int] = ..., latency_ms: _Optional[int] = ..., metadata_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class ToolRetryPolicy(_message.Message):
    __slots__ = ("max_attempts", "initial_backoff_ms", "max_backoff_ms")
    MAX_ATTEMPTS_FIELD_NUMBER: _ClassVar[int]
    INITIAL_BACKOFF_MS_FIELD_NUMBER: _ClassVar[int]
    MAX_BACKOFF_MS_FIELD_NUMBER: _ClassVar[int]
    max_attempts: int
    initial_backoff_ms: int
    max_backoff_ms: int
    def __init__(self, max_attempts: _Optional[int] = ..., initial_backoff_ms: _Optional[int] = ..., max_backoff_ms: _Optional[int] = ...) -> None: ...
