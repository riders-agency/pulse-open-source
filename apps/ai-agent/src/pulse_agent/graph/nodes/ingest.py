from __future__ import annotations

from typing import Any


def ingest_turn(state: dict[str, Any]) -> dict[str, Any]:
    body = str(state.get("body") or "")
    return {"messages": [*list(state.get("messages") or []), {"role": "user", "body": body}]}
