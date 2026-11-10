from __future__ import annotations

from pulse_agent.security import redact


def prompt_inspector_card(prompt_json: dict) -> dict:
    return {"card_type": "prompt_inspector", "prompt_json": redact(prompt_json)}
