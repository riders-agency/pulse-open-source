from __future__ import annotations


def human_approval(state: dict) -> dict:
    card = state.get("active_card") or {}
    return {
        "pending_interrupt": {
            "interrupt_type": "tool_approval",
            "card": card,
            "allowed_actions": ["approve", "reject", "edit", "cancel"],
        }
    }
