from __future__ import annotations


def idempotency_key(project_id: str, session_id: str, run_id: str, tool_call_id: str, revision: int) -> str:
    return f"{project_id}:{session_id}:{run_id}:{tool_call_id}:{revision}"


def test_write_idempotency_key_shape() -> None:
    key = idempotency_key("p", "s", "r", "t", 2)
    assert key == "p:s:r:t:2"
    assert len(key.split(":")) == 5
