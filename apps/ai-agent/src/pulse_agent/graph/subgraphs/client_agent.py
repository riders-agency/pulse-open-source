from __future__ import annotations


def client_choice_card(matches: list[dict]) -> dict:
    return {"card_type": "client_choice", "matches": matches}
