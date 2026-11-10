from __future__ import annotations

import asyncio
import json
import os
from typing import Any, Literal

from pydantic import BaseModel, Field

from pulse_agent.clients.pulse_model_gateway import PulseModelGatewayClient, PulseModelGatewayError
from pulse_agent.config import Settings
from pulse_agent.degraded_mode import MODEL_GATEWAY_BREAKER
from pulse_agent.graph.intent_catalog import (
    READ_MODEL_INTENTS,
    CatalogTool,
    model_prompt_intent_catalog,
    route_for_model_intent,
)
from pulse_agent.graph.intent_schema import (
    EntityRef,
    IntentKey,
    IntentPolicy,
    ToolCandidate,
    TurnClassification,
    TurnKind,
)

DiscourseAct = Literal["command", "question", "approval", "correction", "smalltalk", "unknown"]
UnderstandingDomain = Literal[
    "task",
    "client",
    "workflow",
    "payment",
    "project",
    "approval",
    "conversation",
    "integration",
    "task_process",
    "utility",
    "unknown",
]


class TurnUnderstanding(BaseModel):
    schema_: Literal["riders-pulse-turn-understanding-v1"] = Field(
        default="riders-pulse-turn-understanding-v1",
        alias="schema",
    )
    discourse_act: DiscourseAct = "unknown"
    intent_key: IntentKey = IntentKey.UNKNOWN
    domain: UnderstandingDomain = "unknown"
    confidence: float = Field(default=0.0, ge=0.0, le=1.0)
    is_new_command: bool = False
    is_followup: bool = False
    switch_requested: bool = False
    entities: list[EntityRef] = Field(default_factory=list)
    slot_updates: dict[str, Any] = Field(default_factory=dict)
    clarification_question: str = ""
    multi_intent: list[IntentKey] = Field(default_factory=list)
    rationale_code: str = ""
    model_usage: dict[str, Any] = Field(default_factory=dict)
    source: Literal["deterministic", "model", "fallback"] = "fallback"

    model_config = {"populate_by_name": True}


def model_turn_understanding_enabled() -> bool:
    return os.environ.get("PULSE_AI_MODEL_TURN_UNDERSTANDING", "").strip().lower() in {"1", "true", "yes", "on"}


def classify_with_model_gateway(
    text: str,
    *,
    active_tool_call_id: str = "",
    page_context: dict[str, Any] | None = None,
) -> TurnClassification | None:
    if not model_turn_understanding_enabled():
        return None
    try:
        understanding = _run_model_understanding(text, active_tool_call_id, page_context or {})
    except (PulseModelGatewayError, RuntimeError, ValueError):
        return None
    return classification_from_understanding(understanding)


def deterministic_understanding(classification: TurnClassification) -> TurnUnderstanding:
    return TurnUnderstanding(
        discourse_act=discourse_act_for(classification),
        intent_key=classification.intent_key,
        domain=domain_for(classification.intent_key),
        confidence=classification.confidence,
        is_new_command=classification.is_new_command,
        is_followup=classification.is_followup,
        entities=classification.entities,
        clarification_question=classification.clarification_question,
        rationale_code=classification.reason,
        source="deterministic",
    )


def classification_from_understanding(understanding: TurnUnderstanding) -> TurnClassification:
    if understanding.confidence < 0.72:
        return unknown_from_understanding(understanding, "model_low_confidence")
    route = route_for_model_intent(understanding.intent_key)
    if route is None:
        return unknown_from_understanding(understanding, "model_intent_not_routable")
    if route.key not in READ_MODEL_INTENTS and not model_write_intents_enabled():
        return unknown_from_understanding(understanding, "model_write_intent_blocked")
    return catalog_classification(understanding)


def catalog_classification(
    understanding: TurnUnderstanding,
) -> TurnClassification:
    route = route_for_model_intent(understanding.intent_key)
    if route is None:
        return unknown_from_understanding(understanding, "model_intent_not_routable")
    tools = catalog_tools_for_understanding(understanding)
    allowed_tools = [tool.key for tool in tools]
    return TurnClassification(
        turn_kind=route.turn_kind,
        intent_key=understanding.intent_key,
        confidence=understanding.confidence,
        is_new_command=True,
        is_followup=understanding.is_followup,
        target_entity_type=route.target_entity_type,
        entities=understanding.entities,
        tool_candidates=[
            ToolCandidate(
                tool_key=tool.key,
                risk_level=tool.risk_level,
                requires_approval=tool.requires_approval,
                reason=tool.reason,
            )
            for tool in tools
        ],
        policy=IntentPolicy(
            allowed_tools=allowed_tools,
            write_serial_required=any(tool.requires_approval for tool in tools),
            idempotency_required=any(tool.requires_approval for tool in tools),
        ),
        understanding_source=understanding.source,
        model_usage=understanding.model_usage,
        requires_clarification=bool(understanding.clarification_question),
        clarification_question=understanding.clarification_question,
        reason=reason_for_understanding(understanding),
    )


def catalog_tools_for_understanding(understanding: TurnUnderstanding) -> tuple[CatalogTool, ...]:
    route = route_for_model_intent(understanding.intent_key)
    if route is None:
        return ()
    if route.key == IntentKey.PAYMENT_LOOKUP_DUE and payment_scope_is_project(understanding):
        return (CatalogTool(key="payment.lookup_due", reason="project_payment_read"),)
    return route.tools


def payment_scope_is_project(understanding: TurnUnderstanding) -> bool:
    scope = str(understanding.slot_updates.get("scope") or "").strip().lower()
    if scope in {"project", "all", "company"}:
        return True
    return understanding.rationale_code == "payment_project_lookup_due"


def reason_for_understanding(understanding: TurnUnderstanding) -> str:
    if understanding.rationale_code:
        return understanding.rationale_code
    if understanding.intent_key == IntentKey.CLIENT_SEARCH and not understanding.entities:
        return "client_list"
    if understanding.intent_key == IntentKey.PAYMENT_LOOKUP_DUE and payment_scope_is_project(understanding):
        return "payment_project_lookup_due"
    return "model_understanding"


def unknown_from_understanding(understanding: TurnUnderstanding, reason: str) -> TurnClassification:
    return TurnClassification(
        turn_kind=TurnKind.UNKNOWN,
        intent_key=IntentKey.UNKNOWN,
        confidence=min(understanding.confidence, 0.71),
        understanding_source=understanding.source,
        model_usage=understanding.model_usage,
        requires_clarification=True,
        clarification_question=understanding.clarification_question
        or "Уточните, нужно прочитать данные, создать карточку или изменить текущий объект?",
        reason=reason,
    )


def discourse_act_for(classification: TurnClassification) -> DiscourseAct:
    if classification.turn_kind == TurnKind.APPROVAL_DECISION:
        return "approval"
    if classification.turn_kind == TurnKind.PROPOSAL_EDIT:
        return "correction"
    if classification.turn_kind == TurnKind.SMALLTALK:
        return "smalltalk"
    if classification.turn_kind in {TurnKind.READ_TOOL, TurnKind.DOMAIN_QUESTION, TurnKind.WORKFLOW_QUESTION}:
        return "question"
    if classification.is_new_command:
        return "command"
    return "unknown"


def domain_for(intent: IntentKey) -> UnderstandingDomain:
    value = intent.value
    if value.startswith("task"):
        return "task"
    if value.startswith("client"):
        return "client"
    if value.startswith("workflow"):
        return "workflow"
    if value.startswith("prompt"):
        return "workflow"
    if value.startswith("payment"):
        return "payment"
    if value.startswith("integration"):
        return "integration"
    if value.startswith("task_process"):
        return "task_process"
    if value.startswith("approval"):
        return "approval"
    if value.startswith("conversation"):
        return "conversation"
    if value.startswith("utility"):
        return "utility"
    if value.startswith("page") or value.startswith("project"):
        return "project"
    return "unknown"


def model_write_intents_enabled() -> bool:
    return os.environ.get("PULSE_AI_MODEL_WRITE_INTENTS", "").strip().lower() in {"1", "true", "yes", "on"}


def _run_model_understanding(text: str, active_tool_call_id: str, page_context: dict[str, Any]) -> TurnUnderstanding:
    if not MODEL_GATEWAY_BREAKER.allow_request():
        raise PulseModelGatewayError("model_gateway_circuit_open", "ModelGateway circuit is open")
    settings = Settings()
    try:
        response = _run(
            PulseModelGatewayClient(settings.pulse_model_grpc_addr, settings.service_token).intent_router(
                [
                    {"role": "system", "content": model_prompt()},
                    {
                        "role": "user",
                        "content": json.dumps(
                            {
                                "body": text,
                                "active_tool_call_id": active_tool_call_id,
                                "page_context": page_context,
                            },
                            ensure_ascii=False,
                        ),
                    },
                ],
                response_format_json={"type": "json_object"},
                max_tokens=450,
                temperature=0.0,
                company_id=context_id(page_context, "company_id", "companyId"),
                project_id=context_id(page_context, "project_id", "projectId"),
                user_id=context_id(page_context, "user_id", "userId"),
                session_id=context_id(page_context, "session_id", "sessionId"),
            )
        )
    except Exception:
        MODEL_GATEWAY_BREAKER.record_failure()
        raise
    MODEL_GATEWAY_BREAKER.record_success()
    payload = response.get("response_json") if isinstance(response, dict) else {}
    if not isinstance(payload, dict) or not payload:
        payload = json.loads(str(response.get("content") or "{}")) if isinstance(response, dict) else {}
    payload["source"] = "model"
    payload["model_usage"] = response.get("usage") if isinstance(response.get("usage"), dict) else {}
    payload = normalize_model_payload(payload)
    return TurnUnderstanding.model_validate(payload)


def normalize_model_payload(payload: dict[str, Any]) -> dict[str, Any]:
    raw_intent = str(payload.get("intent_key") or payload.get("intent") or payload.get("key") or "").strip()
    if raw_intent and "intent_key" not in payload:
        payload["intent_key"] = raw_intent
    if str(payload.get("intent_key") or "").strip() in {"", "unknown"}:
        inferred = infer_intent_from_model_payload(payload)
        if inferred != IntentKey.UNKNOWN:
            payload["intent_key"] = inferred.value
    try:
        intent_key = IntentKey(str(payload.get("intent_key") or "unknown"))
    except ValueError:
        return payload
    if intent_key != IntentKey.UNKNOWN and float_value(payload.get("confidence")) <= 0:
        payload["confidence"] = 0.82
    if not str(payload.get("domain") or "").strip() or str(payload.get("domain") or "").strip() == "unknown":
        payload["domain"] = domain_for(intent_key)
    if not str(payload.get("discourse_act") or "").strip() or str(payload.get("discourse_act") or "") == "unknown":
        payload["discourse_act"] = "command" if intent_key not in READ_MODEL_INTENTS else "question"
    if "is_new_command" not in payload and intent_key not in {IntentKey.UNKNOWN, IntentKey.CONVERSATION_SMALLTALK}:
        payload["is_new_command"] = True
    return payload


def infer_intent_from_model_payload(payload: dict[str, Any]) -> IntentKey:
    rationale = str(payload.get("rationale_code") or "").strip()
    raw_slot_updates = payload.get("slot_updates")
    slot_updates = raw_slot_updates if isinstance(raw_slot_updates, dict) else {}
    scope = str(slot_updates.get("scope") or "").strip().lower()
    if rationale == "client_list":
        return IntentKey.CLIENT_SEARCH
    if rationale in {"payment_project_lookup_due", "payment_lookup_due"}:
        return IntentKey.PAYMENT_LOOKUP_DUE
    if scope in {"project", "all", "company"} and rationale.startswith("payment_"):
        return IntentKey.PAYMENT_LOOKUP_DUE
    return IntentKey.UNKNOWN


def float_value(value: Any) -> float:
    try:
        return float(value)
    except (TypeError, ValueError):
        return 0.0


def context_id(context: dict[str, Any], *keys: str) -> str:
    for key in keys:
        value = str(context.get(key) or "").strip()
        if value:
            return value
    return ""


def model_prompt() -> str:
    return (
        "Return compact JSON matching riders-pulse-turn-understanding-v1. "
        "Use only intent_key values from this catalog, or unknown. Do not execute actions. "
        "Return rationale_code, not chain-of-thought. Use unknown below 0.72 confidence. "
        "Distinguish creating a NEW task from searching EXISTING tasks: requests to schedule, "
        "add or record work as a task (e.g. 'заведи задачу', 'запланируй задачу') use task.create. "
        "A task.create classification only prepares a proposal; the server still requires human approval. "
        "For list-all clients like 'выведи клиентов', use client.search, no entities, "
        "rationale_code=client_list. For named client payment like 'когда оплата у Орион', "
        "use payment.lookup_due with a client entity query. For project-wide nearest payments "
        "like 'у кого ближайшие оплаты', use payment.lookup_due with slot_updates.scope=project "
        "and rationale_code=payment_project_lookup_due. Intent catalog:\n"
        f"{model_prompt_intent_catalog()}"
    )


def _run(awaitable: Any) -> Any:
    try:
        asyncio.get_running_loop()
    except RuntimeError:
        return asyncio.run(awaitable)
    raise RuntimeError("Pulse ModelGateway turn understanding cannot run inside an active asyncio loop")
