from __future__ import annotations


def load_context(state: dict) -> dict:
    return {"context_loaded": True, "page_context": state.get("page_context") or {}}
