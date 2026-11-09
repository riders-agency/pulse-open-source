from __future__ import annotations


def load_memories(state: dict) -> dict:
    return {"long_term_memory_refs": state.get("long_term_memory_refs") or []}
