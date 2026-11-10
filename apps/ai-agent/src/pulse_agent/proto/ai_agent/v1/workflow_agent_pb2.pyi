from ai_agent.v1 import common_pb2 as _common_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class WorkflowAgentRequest(_message.Message):
    __slots__ = ("request_id", "company_id", "project_id", "user_id", "action", "mail")
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    COMPANY_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    MAIL_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    company_id: str
    project_id: str
    user_id: str
    action: str
    mail: _struct_pb2.Struct
    def __init__(self, request_id: _Optional[str] = ..., company_id: _Optional[str] = ..., project_id: _Optional[str] = ..., user_id: _Optional[str] = ..., action: _Optional[str] = ..., mail: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class WorkflowAgentResponse(_message.Message):
    __slots__ = ("output", "usage", "executed_nodes")
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    USAGE_FIELD_NUMBER: _ClassVar[int]
    EXECUTED_NODES_FIELD_NUMBER: _ClassVar[int]
    output: _struct_pb2.Struct
    usage: _common_pb2.Usage
    executed_nodes: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, output: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., usage: _Optional[_Union[_common_pb2.Usage, _Mapping]] = ..., executed_nodes: _Optional[_Iterable[str]] = ...) -> None: ...
