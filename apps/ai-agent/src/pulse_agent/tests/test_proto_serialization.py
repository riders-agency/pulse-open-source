from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path
from typing import Any

from google.protobuf import json_format
from google.protobuf.message import Message

_PROTO_ROOT = Path(__file__).resolve().parents[1] / "proto"
if str(_PROTO_ROOT) not in sys.path:
    sys.path.insert(0, str(_PROTO_ROOT))

from ai_agent.v1 import agent_runtime_pb2, model_gateway_pb2, pulse_tools_pb2  # noqa: E402


def test_ai_proto_serialization_fixture() -> None:
    fixture = _load_fixture()
    cases: list[tuple[str, dict[str, Any], Message]] = [
        ("turnRequest", fixture["turnRequest"], agent_runtime_pb2.TurnRequest()),
        ("toolSpec", fixture["toolSpec"], pulse_tools_pb2.ToolSpec()),
        ("modelChatRequest", fixture["modelChatRequest"], model_gateway_pb2.ModelChatRequest()),
    ]

    expected = fixture["expectedDeterministicSha256"]
    for name, payload, message in cases:
        json_format.ParseDict(payload, message)
        deterministic = message.SerializeToString(deterministic=True)
        assert hashlib.sha256(deterministic).hexdigest() == expected[name]
        json_format.MessageToDict(message, preserving_proto_field_name=False)


def _load_fixture() -> dict[str, Any]:
    path = _repo_root() / "contracts" / "proto" / "fixtures" / "ai_agent_v1_serialization.json"
    return json.loads(path.read_text())


def _repo_root() -> Path:
    for parent in Path(__file__).resolve().parents:
        if (parent / "contracts" / "proto" / "fixtures" / "ai_agent_v1_serialization.json").exists():
            return parent
    raise RuntimeError("repository root not found")
