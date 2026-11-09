from __future__ import annotations

from typing import Any

from pulse_agent.graph.intent_schema import (
    IntentKey,
    IntentPolicy,
    ToolCandidate,
    TurnClassification,
    TurnKind,
)
from pulse_agent.graph.observations import sanitize_json

TOPIC_POLICY_SCHEMA = "riders-pulse-topic-policy-v1"

RETURN_MARKERS = (
    "вернемся",
    "вернёмся",
    "вернуться",
    "возвращаемся",
    "назад к",
    "продолжим",
    "back to",
    "return to",
    "resume",
)
SWITCH_MARKERS = (
    "теперь",
    "а теперь",
    "переключись",
    "давай про",
    "посмотри",
    "покажи",
    "найди",
    "создай",
    "create",
    "show",
    "find",
)
CORRECTION_MARKERS = (
    "исправь",
    "измени",
    "поменяй",
    "не так",
    "лучше",
    "замени",
    "добавь",
    "убери",
)
DEICTIC_MARKERS = (
    " его",
    " ему",
    " ее",
    " её",
    " ей",
    " это",
    " этот",
    " эту",
    " этим",
    " там",
    " туда",
    " по нему",
    " по ней",
    " для него",
    " для нее",
    " it",
    " this",
    " that",
)
READ_ONLY_KINDS = {
    TurnKind.DOMAIN_QUESTION,
    TurnKind.PAGE_CONTEXT,
    TurnKind.READ_TOOL,
    TurnKind.SMALLTALK,
    TurnKind.UTILITY,
    TurnKind.WORKFLOW_QUESTION,
}


def apply_topic_policy(
    state: dict[str, Any],
    classification: TurnClassification,
    plan: dict[str, Any],
) -> tuple[TurnClassification, dict[str, Any], dict[str, Any]]:
    policy = evaluate_topic_policy(state, classification)
    next_classification = classification
    next_plan = dict(plan)
    if policy.get("decision") == "correction_to_pending_action":
        next_classification = _proposal_edit_classification(classification, str(policy.get("reason") or "correction"))
        next_plan = _proposal_edit_plan()
    elif policy.get("requires_clarification"):
        question = str(policy.get("clarification_question") or "Уточните, к какому объекту относится запрос.")
        next_classification = classification.model_copy(
            update={
                "turn_kind": TurnKind.UNKNOWN,
                "intent_key": IntentKey.UNKNOWN,
                "confidence": min(classification.confidence, 0.71),
                "requires_clarification": True,
                "clarification_question": question,
                "reason": str(policy.get("reason") or "topic_policy_clarification"),
            }
        )
        next_plan = _clarification_plan(question)
    next_plan = _apply_resolved_referent(next_plan, next_classification, policy)
    return next_classification, next_plan, policy


def evaluate_topic_policy(state: dict[str, Any], classification: TurnClassification) -> dict[str, Any]:
    body = str(state.get("body") or state.get("current_body") or "")
    lower = f" {body.lower().strip()} "
    pending = pending_action(state)
    resolved, ambiguous = resolve_referent(state, classification, lower)
    if classification.intent_key == IntentKey.UNSAFE:
        return base_policy("security_guard", 100, "unsafe_intent")
    if classification.intent_key == IntentKey.APPROVAL_REJECT:
        return base_policy("cancel_or_reject", 90, "approval_reject", pending=pending)
    if classification.intent_key == IntentKey.APPROVAL_APPLY:
        return base_policy("approval_apply", 85, "approval_apply", pending=pending)
    if pending and is_correction(lower, classification):
        return base_policy("correction_to_pending_action", 80, "pending_action_correction", pending=pending)
    if pending and has_marker(lower, RETURN_MARKERS):
        return base_policy(
            "return_to_pending_action",
            70,
            "explicit_return_marker",
            pending=pending,
            branch_override="conversation_reply",
            active_mode="resume_pending_action",
            current_step="await_approval",
            suspend_active_topic=False,
        )
    if ambiguous:
        policy = base_policy("clarify_referent", 65, "ambiguous_deictic_referent", pending=pending)
        policy.update(
            {
                "requires_clarification": True,
                "clarification_question": clarification_question(ambiguous),
                "candidate_referents": ambiguous,
                "branch_override": "conversation_reply",
                "active_mode": "error_recovery",
                "current_step": "request_clarification",
            }
        )
        return policy
    if pending and is_read_only(classification):
        policy = base_policy(
            "read_detour_keep_pending",
            55,
            "read_only_detour_while_pending",
            pending=pending,
            suspend_active_topic=True,
        )
        if resolved:
            policy["resolved_referent"] = resolved
        return policy
    if resolved:
        policy = base_policy("followup_resolved_referent", 45, "deictic_referent_resolved", pending=pending)
        policy["resolved_referent"] = resolved
        return policy
    if has_marker(lower, SWITCH_MARKERS) and classification.is_new_command:
        return base_policy("explicit_switch", 35, "explicit_switch_marker", pending=pending)
    if classification.is_followup or has_deictic(lower):
        return base_policy("followup_without_referent", 25, "followup_no_compatible_referent", pending=pending)
    return base_policy("new_intent", 10, classification.reason or "route_normally", pending=pending)


def resolve_referent(
    state: dict[str, Any],
    classification: TurnClassification,
    lower: str,
) -> tuple[dict[str, Any], list[dict[str, Any]]]:
    if not should_resolve_referent(classification, lower):
        return {}, []
    entity_type = compatible_entity_type(classification)
    candidates = compatible_referents(state, entity_type)
    if len(candidates) == 1:
        return candidates[0], []
    if len(candidates) > 1:
        return {}, candidates[:3]
    return {}, []


def should_resolve_referent(classification: TurnClassification, lower: str) -> bool:
    if classification.entities and any(entity.entity_id or entity.query for entity in classification.entities):
        return False
    return classification.deictic_reference or classification.is_followup or has_deictic(lower)


def compatible_entity_type(classification: TurnClassification) -> str:
    if classification.intent_key in {IntentKey.PAYMENT_LOOKUP_DUE, IntentKey.TASK_SEARCH}:
        return "client"
    if classification.intent_key in {IntentKey.TASK_MOVE, IntentKey.TASK_UPDATE, IntentKey.PROPOSAL_EDIT}:
        return "task"
    if classification.intent_key in {IntentKey.CLIENT_SEARCH, IntentKey.CLIENT_UPDATE}:
        return "client"
    if classification.intent_key in {IntentKey.WORKFLOW_ANSWER_QUESTION, IntentKey.WORKFLOW_PATCH}:
        return "workflow"
    return str(classification.target_entity_type or "none")


def compatible_referents(state: dict[str, Any], entity_type: str) -> list[dict[str, Any]]:
    if entity_type in {"", "none", "payment"}:
        return []
    dialogue = dict_value(state.get("dialogue_state"))
    raw_candidates = []
    raw_candidates.extend(list_value(dialogue.get("last_referents")))
    raw_candidates.extend(list_value(dialogue.get("resolved_entities")))
    raw_candidates.extend(list_value(state.get("resolved_entities")))
    active_entity = dict_value(dict_value(dialogue.get("active_topic")).get("entity")) or dict_value(
        dict_value(state.get("active_topic")).get("entity")
    )
    if active_entity:
        raw_candidates.append(active_entity)
    deduped: dict[str, dict[str, Any]] = {}
    for item in raw_candidates:
        normalized = sanitize_json(item)
        if not isinstance(normalized, dict) or normalized.get("entity_type") != entity_type:
            continue
        key = entity_key(normalized)
        if key:
            deduped[key] = normalized
    return list(deduped.values())


def pending_action(state: dict[str, Any]) -> dict[str, Any]:
    dialogue_pending = dict_value(dict_value(state.get("dialogue_state")).get("pending_action"))
    pending = dict_value(state.get("pending_interrupt"))
    active_card = dict_value(state.get("active_card"))
    active_tool_call_id = str(
        state.get("active_tool_call_id")
        or pending.get("active_tool_call_id")
        or dialogue_pending.get("active_tool_call_id")
        or ""
    )
    if not active_tool_call_id and not active_card and not dialogue_pending:
        return {}
    return {
        "schema": "riders-pulse-topic-policy-pending-action-v1",
        "active_tool_call_id": active_tool_call_id,
        "active_proposal_id": str(state.get("active_proposal_id") or dialogue_pending.get("active_proposal_id") or ""),
        "card_id": str(active_card.get("card_id") or pending.get("card_id") or dialogue_pending.get("card_id") or ""),
        "card_type": str(active_card.get("card_type") or dialogue_pending.get("card_type") or ""),
        "intent_key": str(active_card.get("intent_key") or dialogue_pending.get("intent_key") or ""),
        "status": str(active_card.get("status") or dialogue_pending.get("status") or "pending"),
    }


def base_policy(
    decision: str,
    priority: int,
    reason: str,
    *,
    pending: dict[str, Any] | None = None,
    branch_override: str = "",
    active_mode: str = "",
    current_step: str = "",
    suspend_active_topic: bool = False,
) -> dict[str, Any]:
    return {
        "schema": TOPIC_POLICY_SCHEMA,
        "version": 1,
        "decision": decision,
        "priority": priority,
        "reason": reason,
        "pending_action": pending or {},
        "branch_override": branch_override,
        "active_mode": active_mode,
        "current_step": current_step,
        "suspend_active_topic": suspend_active_topic,
        "requires_clarification": False,
    }


def _apply_resolved_referent(
    plan: dict[str, Any],
    classification: TurnClassification,
    policy: dict[str, Any],
) -> dict[str, Any]:
    resolved = dict_value(policy.get("resolved_referent"))
    if not resolved:
        return plan
    next_plan = dict(plan)
    preview = dict_value(next_plan.get("tool_input_preview")).copy()
    preview["resolved_entities"] = [resolved]
    if resolved.get("entity_type") == "client":
        preview["client_id"] = str(resolved.get("entity_id") or "")
        preview["client_name"] = str(resolved.get("label") or resolved.get("query") or "")
        if classification.intent_key == IntentKey.PAYMENT_LOOKUP_DUE:
            next_plan["required_tools"] = ["payment.lookup_due"]
    next_plan["tool_input_preview"] = preview
    return next_plan


def _proposal_edit_classification(classification: TurnClassification, reason: str) -> TurnClassification:
    return classification.model_copy(
        update={
            "turn_kind": TurnKind.PROPOSAL_EDIT,
            "intent_key": IntentKey.PROPOSAL_EDIT,
            "confidence": max(classification.confidence, 0.9),
            "is_followup": True,
            "is_proposal_edit": True,
            "tool_candidates": [ToolCandidate(tool_key="proposal.edit", risk_level="medium")],
            "policy": IntentPolicy(allowed_tools=["proposal.edit"], write_serial_required=True),
            "reason": reason,
        }
    )


def _proposal_edit_plan() -> dict[str, Any]:
    return {
        "intent": IntentKey.PROPOSAL_EDIT.value,
        "confidence": 0.9,
        "required_tools": ["proposal.edit"],
        "risk_level": "medium",
        "requires_approval": False,
        "user_visible_summary": "Обновить текущую карточку.",
        "tool_input_preview": {},
    }


def _clarification_plan(question: str) -> dict[str, Any]:
    return {
        "intent": IntentKey.UNKNOWN.value,
        "confidence": 0.7,
        "required_tools": [],
        "risk_level": "read",
        "requires_approval": False,
        "user_visible_summary": question,
        "tool_input_preview": {},
    }


def clarification_question(candidates: list[dict[str, Any]]) -> str:
    labels = [candidate_label(item) for item in candidates if candidate_label(item)]
    if not labels:
        return "Уточните, к какому объекту относится запрос."
    return "Уточните, к кому относится запрос: " + ", ".join(labels) + "."


def is_read_only(classification: TurnClassification) -> bool:
    if classification.turn_kind in READ_ONLY_KINDS:
        return True
    return bool(classification.tool_candidates) and all(
        is_read_candidate(candidate) for candidate in classification.tool_candidates
    )


def is_read_candidate(candidate: ToolCandidate) -> bool:
    return candidate.risk_level == "read" and not candidate.requires_approval


def is_correction(lower: str, classification: TurnClassification) -> bool:
    return classification.intent_key == IntentKey.PROPOSAL_EDIT or has_marker(lower, CORRECTION_MARKERS)


def has_deictic(lower: str) -> bool:
    return has_marker(lower, DEICTIC_MARKERS)


def has_marker(lower: str, markers: tuple[str, ...]) -> bool:
    return any(marker in lower for marker in markers)


def candidate_label(item: dict[str, Any]) -> str:
    return str(item.get("label") or item.get("query") or item.get("entity_id") or "").strip()[:120]


def entity_key(entity: dict[str, Any]) -> str:
    return ":".join(
        str(entity.get(key) or "") for key in ("entity_type", "entity_id", "label", "query") if entity.get(key)
    )


def dict_value(value: Any) -> dict[str, Any]:
    return value if isinstance(value, dict) else {}


def list_value(value: Any) -> list[dict[str, Any]]:
    return [item for item in value if isinstance(item, dict)] if isinstance(value, list) else []
