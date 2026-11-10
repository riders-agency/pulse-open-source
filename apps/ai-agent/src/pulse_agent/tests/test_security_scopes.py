from __future__ import annotations

import pytest

from pulse_agent.security import assert_project_scope, redact


def test_cross_project_context_without_scope_is_rejected() -> None:
    with pytest.raises(PermissionError):
        assert_project_scope("", "project", "user")


def test_redacts_tokens_and_authorization() -> None:
    payload = {"authorization": "Bearer secret", "nested": {"api_key": "secret", "ok": "value"}}
    assert redact(payload) == {"authorization": "[redacted]", "nested": {"api_key": "[redacted]", "ok": "value"}}
