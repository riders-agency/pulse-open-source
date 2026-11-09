from __future__ import annotations

from enum import StrEnum
from typing import Any

from pydantic import BaseModel, Field


class RiskLevel(StrEnum):
    READ = "read"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class ToolSpec(BaseModel):
    tool_key: str
    display_name: str
    description: str
    risk_level: RiskLevel
    input_schema_json: dict[str, Any] = Field(default_factory=dict)
    output_schema_json: dict[str, Any] = Field(default_factory=dict)
    required_permissions: list[str] = Field(default_factory=list)
    requires_approval: bool = False
    supports_dry_run: bool = False
    idempotency_required: bool = False
    ui_card_type: str = ""
    timeout_ms: int = 10_000
