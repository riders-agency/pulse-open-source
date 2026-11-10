from __future__ import annotations

from datetime import UTC, datetime
from enum import StrEnum
from typing import Any

from pydantic import BaseModel, Field


class AgentEventType(StrEnum):
    RUN_STARTED = "RUN_STARTED"
    STATUS = "STATUS"
    MESSAGE_DELTA = "MESSAGE_DELTA"
    MESSAGE_FINAL = "MESSAGE_FINAL"
    CARD_CREATED = "CARD_CREATED"
    CARD_UPDATED = "CARD_UPDATED"
    TOOL_PROPOSED = "TOOL_PROPOSED"
    TOOL_STARTED = "TOOL_STARTED"
    TOOL_RESULT = "TOOL_RESULT"
    INTERRUPT = "INTERRUPT"
    APPROVAL_REQUIRED = "APPROVAL_REQUIRED"
    ERROR = "ERROR"
    DONE = "DONE"


class AgentEvent(BaseModel):
    seq: int
    run_id: str
    thread_id: str
    session_id: str
    event_type: AgentEventType
    created_at: datetime = Field(default_factory=lambda: datetime.now(UTC))
    message: str = ""
    body_delta: str = ""
    card: dict[str, Any] | None = None
    tool_call: dict[str, Any] | None = None
    interrupt: dict[str, Any] | None = None
    error: dict[str, Any] | None = None
