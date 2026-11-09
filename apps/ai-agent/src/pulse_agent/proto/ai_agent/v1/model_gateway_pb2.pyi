from ai_agent.v1 import common_pb2 as _common_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ModelPurpose(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    MODEL_PURPOSE_UNSPECIFIED: _ClassVar[ModelPurpose]
    MODEL_PURPOSE_INTENT_ROUTER: _ClassVar[ModelPurpose]
    MODEL_PURPOSE_PLANNER: _ClassVar[ModelPurpose]
    MODEL_PURPOSE_TOOL_ARGS: _ClassVar[ModelPurpose]
    MODEL_PURPOSE_FINAL_RESPONSE: _ClassVar[ModelPurpose]
    MODEL_PURPOSE_MEMORY: _ClassVar[ModelPurpose]
MODEL_PURPOSE_UNSPECIFIED: ModelPurpose
MODEL_PURPOSE_INTENT_ROUTER: ModelPurpose
MODEL_PURPOSE_PLANNER: ModelPurpose
MODEL_PURPOSE_TOOL_ARGS: ModelPurpose
MODEL_PURPOSE_FINAL_RESPONSE: ModelPurpose
MODEL_PURPOSE_MEMORY: ModelPurpose

class ModelMessage(_message.Message):
    __slots__ = ("role", "content", "metadata_json")
    ROLE_FIELD_NUMBER: _ClassVar[int]
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    METADATA_JSON_FIELD_NUMBER: _ClassVar[int]
    role: str
    content: str
    metadata_json: _struct_pb2.Struct
    def __init__(self, role: _Optional[str] = ..., content: _Optional[str] = ..., metadata_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class ModelChatRequest(_message.Message):
    __slots__ = ("request_id", "company_id", "project_id", "user_id", "session_id", "thread_id", "purpose", "messages", "response_format_json", "tool_schemas_json", "max_tokens", "temperature", "timeout_ms", "trace_context")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    COMPANY_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    THREAD_ID_FIELD_NUMBER: _ClassVar[int]
    PURPOSE_FIELD_NUMBER: _ClassVar[int]
    MESSAGES_FIELD_NUMBER: _ClassVar[int]
    RESPONSE_FORMAT_JSON_FIELD_NUMBER: _ClassVar[int]
    TOOL_SCHEMAS_JSON_FIELD_NUMBER: _ClassVar[int]
    MAX_TOKENS_FIELD_NUMBER: _ClassVar[int]
    TEMPERATURE_FIELD_NUMBER: _ClassVar[int]
    TIMEOUT_MS_FIELD_NUMBER: _ClassVar[int]
    TRACE_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    company_id: str
    project_id: str
    user_id: str
    session_id: str
    thread_id: str
    purpose: ModelPurpose
    messages: _containers.RepeatedCompositeFieldContainer[ModelMessage]
    response_format_json: _struct_pb2.Struct
    tool_schemas_json: _struct_pb2.Struct
    max_tokens: int
    temperature: float
    timeout_ms: int
    trace_context: _common_pb2.TraceContext
    def __init__(self, request_id: _Optional[str] = ..., company_id: _Optional[str] = ..., project_id: _Optional[str] = ..., user_id: _Optional[str] = ..., session_id: _Optional[str] = ..., thread_id: _Optional[str] = ..., purpose: _Optional[_Union[ModelPurpose, str]] = ..., messages: _Optional[_Iterable[_Union[ModelMessage, _Mapping]]] = ..., response_format_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., tool_schemas_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., max_tokens: _Optional[int] = ..., temperature: _Optional[float] = ..., timeout_ms: _Optional[int] = ..., trace_context: _Optional[_Union[_common_pb2.TraceContext, _Mapping]] = ...) -> None: ...

class ModelChatResponse(_message.Message):
    __slots__ = ("content", "response_json", "usage")
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    RESPONSE_JSON_FIELD_NUMBER: _ClassVar[int]
    USAGE_FIELD_NUMBER: _ClassVar[int]
    content: str
    response_json: _struct_pb2.Struct
    usage: _common_pb2.Usage
    def __init__(self, content: _Optional[str] = ..., response_json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., usage: _Optional[_Union[_common_pb2.Usage, _Mapping]] = ...) -> None: ...

class ModelChatEvent(_message.Message):
    __slots__ = ("delta", "done", "usage", "error")
    DELTA_FIELD_NUMBER: _ClassVar[int]
    DONE_FIELD_NUMBER: _ClassVar[int]
    USAGE_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    delta: str
    done: bool
    usage: _common_pb2.Usage
    error: _struct_pb2.Struct
    def __init__(self, delta: _Optional[str] = ..., done: _Optional[bool] = ..., usage: _Optional[_Union[_common_pb2.Usage, _Mapping]] = ..., error: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class EmbedRequest(_message.Message):
    __slots__ = ("request_id", "company_id", "project_id", "user_id", "input", "trace_context")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    COMPANY_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    TRACE_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    company_id: str
    project_id: str
    user_id: str
    input: _containers.RepeatedScalarFieldContainer[str]
    trace_context: _common_pb2.TraceContext
    def __init__(self, request_id: _Optional[str] = ..., company_id: _Optional[str] = ..., project_id: _Optional[str] = ..., user_id: _Optional[str] = ..., input: _Optional[_Iterable[str]] = ..., trace_context: _Optional[_Union[_common_pb2.TraceContext, _Mapping]] = ...) -> None: ...

class EmbedResponse(_message.Message):
    __slots__ = ("embeddings", "usage")
    EMBEDDINGS_FIELD_NUMBER: _ClassVar[int]
    USAGE_FIELD_NUMBER: _ClassVar[int]
    embeddings: _containers.RepeatedCompositeFieldContainer[_struct_pb2.ListValue]
    usage: _common_pb2.Usage
    def __init__(self, embeddings: _Optional[_Iterable[_Union[_struct_pb2.ListValue, _Mapping]]] = ..., usage: _Optional[_Union[_common_pb2.Usage, _Mapping]] = ...) -> None: ...
