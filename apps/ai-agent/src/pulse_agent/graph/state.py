from __future__ import annotations

from typing import Any, TypedDict


class AgentState(TypedDict, total=False):
    messages: list[dict[str, Any]]
    company_id: str
    project_id: str
    user_id: str
    session_id: str
    thread_id: str
    checkpoint_thread_id: str
    checkpoint_namespace: str
    checkpoint_enabled: bool
    checkpoint_status: str
    state_schema_version: int
    service_version: str
    thread_revision: int
    expected_thread_revision: int
    run_id: str
    turn_id: str
    idempotency_key: str
    body: str
    current_body: str
    locale: str
    timezone: str
    local_date: str
    page_context: dict[str, Any]
    client_context: dict[str, Any]
    ui_state: dict[str, Any]
    auth_context_summary: dict[str, Any]
    tool_registry: dict[str, Any]
    working_memory: dict[str, Any]
    long_term_memory_refs: list[dict[str, Any]]
    dialogue_state: dict[str, Any]
    active_topic: dict[str, Any]
    suspended_topics: list[dict[str, Any]]
    resolved_entities: list[dict[str, Any]]
    unresolved_slots: dict[str, Any]
    active_mode: str
    current_step: str
    active_card_id: str
    active_tool_call_id: str
    active_proposal_id: str
    active_card: dict[str, Any] | None
    active_tool_call: dict[str, Any] | None
    active_proposal: dict[str, Any] | None
    intent: dict[str, Any] | None
    plan: dict[str, Any] | None
    topic_policy: dict[str, Any] | None
    branch: str
    candidate_tools: list[str]
    proposed_cards: dict[str, dict[str, Any]]
    proposal_edit: dict[str, Any] | None
    proposal_gate: dict[str, Any] | None
    pending_interrupt: dict[str, Any] | None
    approval_decision: dict[str, Any] | None
    cancel_requested: bool
    canceled_at: str
    cancel_reason: str
    tool_results: list[dict[str, Any]]
    tool_observations: list[dict[str, Any]]
    replan_trace: list[dict[str, Any]]
    final_response: str
    message_payload_json: dict[str, Any]
    awaiting_tool_observation: bool
    events: list[dict[str, Any]]
    errors: list[dict[str, Any]]
    audit_refs: list[dict[str, Any]]
    usage: dict[str, Any]
    execution_event_start: int
    node_trace: list[dict[str, Any]]
    model_calls: list[dict[str, Any]]
    trace_context: dict[str, Any]
    observability_started_ms: int
    ai_observability_json: dict[str, Any]
    processed_turns: dict[str, dict[str, Any]]
    last_turn_idempotency_key: str
    last_turn_result: dict[str, Any]


def merge_working_memory(left: dict[str, Any] | None, right: dict[str, Any] | None) -> dict[str, Any]:
    merged = dict(left or {})
    merged.update(right or {})
    return merged


def upsert_card(cards: dict[str, dict[str, Any]] | None, card: dict[str, Any]) -> dict[str, dict[str, Any]]:
    result = dict(cards or {})
    card_id = str(card.get("card_id") or "")
    if card_id:
        current = result.get(card_id)
        if not current or int(card.get("revision", 0)) >= int(current.get("revision", 0)):
            result[card_id] = card
    return result
