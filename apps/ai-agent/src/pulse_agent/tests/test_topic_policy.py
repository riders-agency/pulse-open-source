from __future__ import annotations

from pulse_agent.graph.intent_schema import IntentKey, classify_turn
from pulse_agent.topic_policy import apply_topic_policy, evaluate_topic_policy


def test_read_detour_resolves_pronoun_to_latest_client_and_keeps_pending() -> None:
    state = pending_task_state(
        referents=[{"entity_type": "client", "entity_id": "client-1", "label": "Орион"}],
        body="покажи оплаты по нему",
    )
    classification = classify_turn("покажи оплаты по нему", active_tool_call_id="pending-task-create")
    plan = {
        "intent": "payment.lookup_due",
        "required_tools": ["client.search", "payment.lookup_due"],
        "tool_input_preview": {"query": "покажи оплаты по нему"},
    }

    next_classification, next_plan, policy = apply_topic_policy(state, classification, plan)

    assert next_classification.intent_key == IntentKey.PAYMENT_LOOKUP_DUE
    assert policy["decision"] == "read_detour_keep_pending"
    assert policy["resolved_referent"]["entity_id"] == "client-1"
    assert policy["pending_action"]["active_tool_call_id"] == "pending-task-create"
    assert next_plan["required_tools"] == ["payment.lookup_due"]
    assert next_plan["tool_input_preview"]["client_id"] == "client-1"


def test_return_marker_restores_pending_action_without_applying_it() -> None:
    state = pending_task_state(body="вернемся к задаче")
    classification = classify_turn("вернемся к задаче", active_tool_call_id="pending-task-create")

    policy = evaluate_topic_policy(state, classification)

    assert policy["decision"] == "return_to_pending_action"
    assert policy["branch_override"] == "conversation_reply"
    assert policy["active_mode"] == "resume_pending_action"
    assert policy["pending_action"]["card_id"] == "card-task-create"


def test_correction_updates_same_pending_action() -> None:
    state = pending_task_state(body="измени срок на завтра")
    classification = classify_turn("измени срок на завтра", active_tool_call_id="pending-task-create")
    next_classification, next_plan, policy = apply_topic_policy(state, classification, {})

    assert policy["decision"] == "correction_to_pending_action"
    assert next_classification.intent_key == IntentKey.PROPOSAL_EDIT
    assert next_plan["intent"] == "proposal.edit"


def test_ambiguous_pronoun_requires_targeted_candidate_clarification() -> None:
    state = pending_task_state(
        referents=[
            {"entity_type": "client", "entity_id": "client-1", "label": "Орион"},
            {"entity_type": "client", "entity_id": "client-2", "label": "KSS"},
        ],
        body="покажи оплаты по нему",
    )
    classification = classify_turn("покажи оплаты по нему", active_tool_call_id="pending-task-create")

    policy = evaluate_topic_policy(state, classification)

    assert policy["decision"] == "clarify_referent"
    assert policy["requires_clarification"] is True
    assert "Орион" in policy["clarification_question"]
    assert "KSS" in policy["clarification_question"]


def pending_task_state(
    *,
    referents: list[dict[str, str]] | None = None,
    body: str = "",
) -> dict[str, object]:
    active_card = {
        "card_id": "card-task-create",
        "card_type": "task_create_proposal",
        "intent_key": "task.create",
        "risk_level": "high",
        "status": "needs_approval",
        "debug_json": {
            "input_json": {
                "title": "Проверить оплату",
                "client_query": "Орион",
                "due_date": "2026-06-26",
            }
        },
    }
    return {
        "body": body,
        "active_tool_call_id": "pending-task-create",
        "active_card": active_card,
        "pending_interrupt": {"active_tool_call_id": "pending-task-create", "card_id": "card-task-create"},
        "dialogue_state": {
            "schema": "riders-pulse-dialogue-state-v1",
            "active_topic": {"topic_id": "task.create:client:client-1", "intent": "task.create"},
            "last_referents": referents or [],
            "resolved_entities": referents or [],
            "pending_action": {
                "active_tool_call_id": "pending-task-create",
                "card_id": "card-task-create",
                "intent_key": "task.create",
            },
        },
    }
