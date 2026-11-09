from __future__ import annotations

from dataclasses import dataclass


@dataclass
class PulseAuditClient:
    target: str

    async def record(self, event_type: str, metadata: dict) -> None:
        del event_type, metadata
