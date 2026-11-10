from __future__ import annotations

import re
from enum import StrEnum
from typing import Any, Literal
from urllib.parse import urlparse

from pydantic import BaseModel, Field

from pulse_agent.graph.advice import is_advice_request


class TurnKind(StrEnum):
    SMALLTALK = "smalltalk"
    UTILITY = "utility"
    PAGE_CONTEXT = "page_context"
    DOMAIN_QUESTION = "domain_question"
    READ_TOOL = "read_tool"
    WRITE_INTENT = "write_intent"
    PROPOSAL_EDIT = "proposal_edit"
    APPROVAL_DECISION = "approval_decision"
    WORKFLOW_QUESTION = "workflow_question"
    WORKFLOW_PATCH = "workflow_patch"
    MULTI_STEP = "multi_step"
    UNSAFE = "unsafe"
    UNKNOWN = "unknown"


class IntentKey(StrEnum):
    CONVERSATION_SMALLTALK = "conversation.smalltalk"
    UTILITY_CALCULATOR = "utility.calculator"
    PAGE_WHERE_AM_I = "page.where_am_i"
    PAGE_VERIFY_CONTEXT = "page.verify_context"
    PROJECT_EXPLAIN = "project.explain"
    CLIENT_SEARCH = "client.search"
    CLIENT_READ = "client.read"
    CLIENT_COUNT = "client.count"
    CLIENT_CREATE = "client.create"
    CLIENT_UPDATE = "client.update"
    TASK_SEARCH = "task.search"
    TASK_READ = "task.read"
    TASK_CREATE = "task.create"
    TASK_UPDATE = "task.update"
    TASK_MOVE = "task.move"
    PAYMENT_LOOKUP_DUE = "payment.lookup_due"
    PAYMENT_SCHEDULE_READ = "payment.schedule.read"
    PAYMENT_ACTION_CREATE = "payment.action.create"
    PAYMENT_CONFIRM = "payment.confirm"
    WORKFLOW_ANSWER_QUESTION = "workflow.answer_question"
    WORKFLOW_READ_DRAFT = "workflow.read_draft"
    WORKFLOW_READ_CONTEXT = "workflow.read_context"
    WORKFLOW_COMPILE_PROMPT = "workflow.compile_prompt"
    WORKFLOW_PATCH = "workflow.patch"
    WORKFLOW_RUN_TEST = "workflow.run_test"
    TASK_PROCESS_READ = "task_process.read"
    TASK_PROCESS_PATCH = "task_process.patch"
    PROMPT_COPY = "prompt.copy"
    INTEGRATION_CATALOG_LIST = "integration.catalog.list"
    INTEGRATION_DEPENDENCY_READ = "integration.dependency.read"
    INTEGRATION_DIAGNOSTICS = "integration.diagnostics"
    INTEGRATION_SECRET_ROTATE = "integration.secret.rotate"
    PROPOSAL_EDIT = "proposal.edit"
    APPROVAL_APPLY = "approval.apply"
    APPROVAL_REJECT = "approval.reject"
    MULTI_STEP_PLAN = "multi_step.plan"
    UNSAFE = "unsafe"
    UNKNOWN = "unknown"


class EntityRef(BaseModel):
    entity_type: Literal[
        "task",
        "client",
        "project",
        "workflow",
        "payment",
        "integration",
        "provider",
        "task_process",
        "none",
    ] = "none"
    entity_id: str = ""
    query: str = ""
    confidence: float = 0.0
    source: str = "user"


class ToolCandidate(BaseModel):
    tool_key: str
    risk_level: Literal["read", "medium", "high", "critical", "unsafe"] = "read"
    requires_approval: bool = False
    reason: str = ""


class IntentPolicy(BaseModel):
    allowed_tools: list[str] = Field(default_factory=list)
    denied_tools: list[str] = Field(default_factory=list)
    read_parallel_allowed: bool = True
    write_serial_required: bool = True
    idempotency_required: bool = False


class ProposalGateDecision(BaseModel):
    decision: Literal[
        "route_normally",
        "patch_current",
        "approval_bridge",
        "run_new_keep_pending",
        "conflict_choice",
    ] = "route_normally"
    reason: str = ""
    active_tool_call_id: str = ""
    target_intent_key: str = ""


class TurnClassification(BaseModel):
    turn_kind: TurnKind
    intent_key: IntentKey
    confidence: float
    is_new_command: bool = False
    is_followup: bool = False
    is_proposal_edit: bool = False
    deictic_reference: bool = False
    target_entity_type: Literal[
        "task",
        "client",
        "project",
        "workflow",
        "payment",
        "integration",
        "provider",
        "task_process",
        "none",
    ] = "none"
    target_entity_id: str = ""
    explicit_url: str = ""
    entities: list[EntityRef] = Field(default_factory=list)
    tool_candidates: list[ToolCandidate] = Field(default_factory=list)
    policy: IntentPolicy = Field(default_factory=IntentPolicy)
    understanding_source: Literal["deterministic", "model", "fallback"] = "deterministic"
    model_usage: dict[str, Any] = Field(default_factory=dict)
    requires_clarification: bool = False
    clarification_question: str = ""
    reason: str = ""


def classify_turn(
    text: str,
    *,
    active_tool_call_id: str = "",
    page_context: dict[str, Any] | None = None,
) -> TurnClassification:
    body = text.strip()
    lower = body.lower()
    page_context = page_context or {}
    explicit_url = extract_url(body)

    if contains_secret_request(lower):
        return TurnClassification(
            turn_kind=TurnKind.UNSAFE,
            intent_key=IntentKey.UNSAFE,
            confidence=0.99,
            explicit_url=explicit_url,
            reason="secret_request",
        )

    if is_phatic(lower):
        return TurnClassification(
            turn_kind=TurnKind.SMALLTALK,
            intent_key=IntentKey.CONVERSATION_SMALLTALK,
            confidence=0.98,
            reason="phatic",
        )

    if is_calculator(lower):
        return TurnClassification(
            turn_kind=TurnKind.UTILITY,
            intent_key=IntentKey.UTILITY_CALCULATOR,
            confidence=0.92,
            reason="calculator",
        )

    if is_context_verification(lower):
        return TurnClassification(
            turn_kind=TurnKind.PAGE_CONTEXT,
            intent_key=IntentKey.PAGE_VERIFY_CONTEXT,
            confidence=0.96,
            is_new_command=True,
            target_entity_type="project",
            tool_candidates=[ToolCandidate(tool_key="ui_context.read", reason="verify_page_context")],
            policy=IntentPolicy(allowed_tools=["ui_context.read", "route.explain"]),
            reason="page_context_verification",
        )

    if is_navigation_query(lower):
        return TurnClassification(
            turn_kind=TurnKind.PAGE_CONTEXT,
            intent_key=IntentKey.PAGE_WHERE_AM_I,
            confidence=0.97,
            is_new_command=True,
            target_entity_type="project",
            tool_candidates=[ToolCandidate(tool_key="ui_context.read", reason="current_page")],
            policy=IntentPolicy(allowed_tools=["ui_context.read", "route.explain"]),
            reason="navigation_query",
        )

    if active_tool_call_id and is_approval_apply(lower):
        return TurnClassification(
            turn_kind=TurnKind.APPROVAL_DECISION,
            intent_key=IntentKey.APPROVAL_APPLY,
            confidence=0.99,
            is_followup=True,
            tool_candidates=[ToolCandidate(tool_key="approval.apply", risk_level="high", reason="active_proposal")],
            policy=IntentPolicy(
                allowed_tools=["approval.apply"],
                write_serial_required=True,
                idempotency_required=True,
            ),
            reason="approval_apply",
        )

    if active_tool_call_id and is_approval_reject(lower):
        return TurnClassification(
            turn_kind=TurnKind.APPROVAL_DECISION,
            intent_key=IntentKey.APPROVAL_REJECT,
            confidence=0.99,
            is_followup=True,
            tool_candidates=[ToolCandidate(tool_key="approval.reject", risk_level="high", reason="active_proposal")],
            policy=IntentPolicy(allowed_tools=["approval.reject"], write_serial_required=True),
            reason="approval_reject",
        )

    workflow_question = classify_workflow(lower, page_context, explicit_url)
    if workflow_question is not None:
        return workflow_question

    if explicit_url:
        return TurnClassification(
            turn_kind=TurnKind.DOMAIN_QUESTION,
            intent_key=IntentKey.PROJECT_EXPLAIN,
            confidence=0.78,
            is_new_command=True,
            explicit_url=explicit_url,
            tool_candidates=[ToolCandidate(tool_key="context.read", reason="url_context")],
            reason="explicit_url",
        )

    if is_project_explain(lower):
        return TurnClassification(
            turn_kind=TurnKind.DOMAIN_QUESTION,
            intent_key=IntentKey.PROJECT_EXPLAIN,
            confidence=0.93,
            is_new_command=True,
            target_entity_type="project",
            tool_candidates=[ToolCandidate(tool_key="context.read", reason="project_explain")],
            policy=IntentPolicy(allowed_tools=["context.read", "ui_context.read"]),
            reason="project_explain",
        )

    if is_task_payment_read(lower):
        return TurnClassification(
            turn_kind=TurnKind.READ_TOOL,
            intent_key=IntentKey.PROJECT_EXPLAIN,
            confidence=0.9,
            is_new_command=True,
            target_entity_type="client",
            entities=[EntityRef(entity_type="client", query=extract_after_markers(body, ("у", "по")), confidence=0.65)],
            tool_candidates=[
                ToolCandidate(tool_key="task.search", reason="task_read"),
                ToolCandidate(tool_key="payment.lookup_due", reason="payment_read"),
            ],
            policy=IntentPolicy(allowed_tools=["task.search", "payment.lookup_due"]),
            reason="task_payment_read",
        )

    if is_client_count(lower):
        return TurnClassification(
            turn_kind=TurnKind.READ_TOOL,
            intent_key=IntentKey.CLIENT_COUNT,
            confidence=0.94,
            is_new_command=True,
            target_entity_type="client",
            tool_candidates=[ToolCandidate(tool_key="client.count", reason="client_count")],
            policy=IntentPolicy(allowed_tools=["client.count"]),
            reason="client_count",
        )

    if is_client_search(lower):
        list_request = is_client_list(lower)
        query = "" if list_request else extract_after_markers(body, ("клиента", "клиент", "по", "про"))
        return TurnClassification(
            turn_kind=TurnKind.READ_TOOL,
            intent_key=IntentKey.CLIENT_SEARCH,
            confidence=0.95,
            is_new_command=True,
            target_entity_type="client",
            entities=[EntityRef(entity_type="client", query=query, confidence=0.75)] if query else [],
            tool_candidates=[ToolCandidate(tool_key="client.search", reason="client_read")],
            policy=IntentPolicy(allowed_tools=["client.search"]),
            reason="client_list" if list_request else "client_search",
        )

    if is_task_search(lower):
        query = extract_after_markers(body, ("по", "про", "клиенту", "клиента"))
        return TurnClassification(
            turn_kind=TurnKind.READ_TOOL,
            intent_key=IntentKey.TASK_SEARCH,
            confidence=0.95,
            is_new_command=True,
            target_entity_type="task",
            entities=[EntityRef(entity_type="client", query=query, confidence=0.7)] if query else [],
            tool_candidates=[ToolCandidate(tool_key="task.search", reason="task_read")],
            policy=IntentPolicy(allowed_tools=["task.search", "client.search"]),
            reason="task_search",
        )

    if is_project_payment_lookup(lower):
        return TurnClassification(
            turn_kind=TurnKind.READ_TOOL,
            intent_key=IntentKey.PAYMENT_LOOKUP_DUE,
            confidence=0.95,
            is_new_command=True,
            target_entity_type="payment",
            tool_candidates=[ToolCandidate(tool_key="payment.lookup_due", reason="project_payment_read")],
            policy=IntentPolicy(allowed_tools=["payment.lookup_due"]),
            reason="payment_project_lookup_due",
        )

    if is_payment_lookup(lower):
        return TurnClassification(
            turn_kind=TurnKind.READ_TOOL,
            intent_key=IntentKey.PAYMENT_LOOKUP_DUE,
            confidence=0.94,
            is_new_command=True,
            target_entity_type="payment",
            tool_candidates=[
                ToolCandidate(tool_key="client.search", reason="resolve_client"),
                ToolCandidate(tool_key="payment.lookup_due", reason="payment_read"),
            ],
            policy=IntentPolicy(allowed_tools=["client.search", "payment.lookup_due"]),
            reason="payment_lookup_due",
        )

    if is_compound_client_task(lower):
        return TurnClassification(
            turn_kind=TurnKind.MULTI_STEP,
            intent_key=IntentKey.MULTI_STEP_PLAN,
            confidence=0.92,
            is_new_command=True,
            target_entity_type="client",
            tool_candidates=[
                ToolCandidate(tool_key="client.create", risk_level="high", requires_approval=True),
                ToolCandidate(tool_key="task.create", risk_level="high", requires_approval=True),
            ],
            policy=IntentPolicy(
                allowed_tools=["client.create", "task.create", "task.move"],
                write_serial_required=True,
                idempotency_required=True,
            ),
            reason="compound_client_task",
        )

    if is_task_create(lower):
        return TurnClassification(
            turn_kind=TurnKind.WRITE_INTENT,
            intent_key=IntentKey.TASK_CREATE,
            confidence=0.96,
            is_new_command=True,
            target_entity_type="task",
            tool_candidates=[
                ToolCandidate(tool_key="client.search", reason="resolve_client"),
                ToolCandidate(tool_key="task.create", risk_level="high", requires_approval=True),
            ],
            policy=IntentPolicy(
                allowed_tools=["client.search", "task.create"], write_serial_required=True, idempotency_required=True
            ),
            reason="task_create",
        )

    if is_task_move(lower):
        selected_task_id = str(page_context.get("taskId") or page_context.get("task_id") or "")
        if not selected_task_id:
            return TurnClassification(
                turn_kind=TurnKind.UNKNOWN,
                intent_key=IntentKey.UNKNOWN,
                confidence=0.72,
                requires_clarification=True,
                clarification_question="Уточните, какую задачу нужно изменить.",
                reason="deictic_task_without_target",
            )
        return TurnClassification(
            turn_kind=TurnKind.WRITE_INTENT,
            intent_key=IntentKey.TASK_MOVE,
            confidence=0.92,
            is_new_command=True,
            deictic_reference=True,
            target_entity_type="task",
            target_entity_id=selected_task_id,
            entities=[EntityRef(entity_type="task", entity_id=selected_task_id, confidence=0.9, source="ui_context")],
            tool_candidates=[ToolCandidate(tool_key="task.move", risk_level="high", requires_approval=True)],
            policy=IntentPolicy(allowed_tools=["task.move"], write_serial_required=True, idempotency_required=True),
            reason="deictic_task_move",
        )

    if is_client_create(lower):
        return TurnClassification(
            turn_kind=TurnKind.WRITE_INTENT,
            intent_key=IntentKey.CLIENT_CREATE,
            confidence=0.9,
            is_new_command=True,
            target_entity_type="client",
            tool_candidates=[ToolCandidate(tool_key="client.create", risk_level="high", requires_approval=True)],
            policy=IntentPolicy(allowed_tools=["client.create"], write_serial_required=True, idempotency_required=True),
            reason="client_create",
        )

    if active_tool_call_id and is_proposal_edit_phrase(lower):
        return TurnClassification(
            turn_kind=TurnKind.PROPOSAL_EDIT,
            intent_key=IntentKey.PROPOSAL_EDIT,
            confidence=0.9,
            is_followup=True,
            is_proposal_edit=True,
            tool_candidates=[ToolCandidate(tool_key="proposal.edit", risk_level="medium")],
            policy=IntentPolicy(allowed_tools=["proposal.edit"], write_serial_required=True),
            reason="proposal_edit_phrase",
        )

    # Explicit advice stays text-only after all known business-command rules.
    # A model classifier must not silently reinterpret it as a database search.
    if not active_tool_call_id and is_advice_request(body):
        return TurnClassification(
            turn_kind=TurnKind.UNKNOWN, intent_key=IntentKey.UNKNOWN,
            confidence=0.95, policy=IntentPolicy(allowed_tools=[]), reason="explicit_advice",
        )

    model_classification = model_fallback_classification(
        body,
        active_tool_call_id=active_tool_call_id,
        page_context=page_context,
    )
    if model_classification is not None:
        return model_classification

    return TurnClassification(
        turn_kind=TurnKind.UNKNOWN,
        intent_key=IntentKey.UNKNOWN,
        confidence=0.4,
        requires_clarification=True,
        clarification_question="Уточните, нужно прочитать данные, создать карточку или изменить текущий объект?",
        reason="no_rule_match",
    )


def model_fallback_classification(
    text: str,
    *,
    active_tool_call_id: str = "",
    page_context: dict[str, Any] | None = None,
) -> TurnClassification | None:
    from pulse_agent.graph.turn_understanding import classify_with_model_gateway

    return classify_with_model_gateway(text, active_tool_call_id=active_tool_call_id, page_context=page_context or {})


def proposal_gate_decision(classification: TurnClassification, active_tool_call_id: str) -> ProposalGateDecision:
    if not active_tool_call_id:
        return ProposalGateDecision(decision="route_normally", reason="no_active_proposal")
    if classification.intent_key in {IntentKey.APPROVAL_APPLY, IntentKey.APPROVAL_REJECT}:
        return ProposalGateDecision(
            decision="approval_bridge",
            reason=classification.reason,
            active_tool_call_id=active_tool_call_id,
            target_intent_key=classification.intent_key.value,
        )
    if classification.intent_key == IntentKey.PROPOSAL_EDIT:
        return ProposalGateDecision(
            decision="patch_current",
            reason=classification.reason,
            active_tool_call_id=active_tool_call_id,
            target_intent_key=classification.intent_key.value,
        )
    if classification.is_new_command:
        read_turn_kinds = {
            TurnKind.READ_TOOL,
            TurnKind.PAGE_CONTEXT,
            TurnKind.DOMAIN_QUESTION,
            TurnKind.WORKFLOW_QUESTION,
        }
        if classification.turn_kind in read_turn_kinds:
            return ProposalGateDecision(
                decision="run_new_keep_pending",
                reason="new_read_command_while_proposal_pending",
                active_tool_call_id=active_tool_call_id,
                target_intent_key=classification.intent_key.value,
            )
        if classification.turn_kind in {TurnKind.WRITE_INTENT, TurnKind.MULTI_STEP, TurnKind.WORKFLOW_PATCH}:
            return ProposalGateDecision(
                decision="run_new_keep_pending",
                reason="new_write_command_while_proposal_pending",
                active_tool_call_id=active_tool_call_id,
                target_intent_key=classification.intent_key.value,
            )
    return ProposalGateDecision(decision="route_normally", reason="not_proposal_related")


def contains_secret_request(lower: str) -> bool:
    return any(marker in lower for marker in ("api key", "секрет", "парол", "токен", "secret"))


def is_phatic(lower: str) -> bool:
    text = lower.strip(" ?!.")
    return text in {"ты тут", "ты здесь", "работаешь", "ау", "привет", "алло", "на связи", "ты на связи"}


def is_calculator(lower: str) -> bool:
    return bool(re.fullmatch(r"\s*\d+([\s+\-*/xх]\s*\d+)+\s*=?\s*", lower))


def is_navigation_query(lower: str) -> bool:
    return any(marker in lower for marker in ("где я сейчас", "где мы сейчас", "на какой странице", "какой экран"))


def is_context_verification(lower: str) -> bool:
    return any(
        marker in lower for marker in ("разве это экран", "проверь контекст", "верный ли экран", "это точно экран")
    )


def is_project_explain(lower: str) -> bool:
    if "проект" not in lower and "интеграц" not in lower:
        return False
    return any(
        marker in lower
        for marker in (
            "объясни",
            "обьясни",
            "расскажи",
            "что за проект",
            "текущий проект",
            "проверь ai интеграц",
            "проверь аи интеграц",
            "проверь интеграц",
            "ai интеграц",
            "аи интеграц",
        )
    )


def is_approval_apply(lower: str) -> bool:
    text = lower.strip(" ?!.")
    return text in {"да", "давай", "да создавай", "да, создавай", "создавай", "подтверждаю", "применить"} or any(
        marker in text for marker in ("да, созда", "да созда", "подтвержда", "применяй")
    )


def is_approval_reject(lower: str) -> bool:
    text = lower.strip(" ?!.")
    return text in {"нет", "не надо", "отмена", "отмени", "сбрось", "отклонить", "не создавай"} or any(
        marker in text for marker in ("не надо", "отмен", "не создавай", "отклон")
    )


def classify_workflow(lower: str, page_context: dict[str, Any], explicit_url: str = "") -> TurnClassification | None:
    url_path = urlparse(explicit_url).path if explicit_url else ""
    mentions_workflow = (
        "workflow" in lower or str(page_context.get("pageKind") or "") == "workflow_editor" or "/workflows/" in url_path
    )
    if not mentions_workflow:
        return None
    if any(marker in lower for marker in ("добавь", "удали", "переделай", "замени", "сделай чтобы", "измени")):
        return TurnClassification(
            turn_kind=TurnKind.WORKFLOW_PATCH,
            intent_key=IntentKey.WORKFLOW_PATCH,
            confidence=0.9,
            is_new_command=True,
            target_entity_type="workflow",
            tool_candidates=[ToolCandidate(tool_key="workflow.patch", risk_level="critical", requires_approval=True)],
            policy=IntentPolicy(allowed_tools=["workflow.read_draft", "workflow.patch"], idempotency_required=True),
            reason="workflow_patch",
        )
    if any(marker in lower for marker in ("сколько", "какие действия", "есть дубликаты", "что делает", "объясни")):
        return TurnClassification(
            turn_kind=TurnKind.WORKFLOW_QUESTION,
            intent_key=IntentKey.WORKFLOW_ANSWER_QUESTION,
            confidence=0.9,
            is_new_command=True,
            target_entity_type="workflow",
            tool_candidates=[ToolCandidate(tool_key="workflow.read_draft", reason="workflow_question")],
            policy=IntentPolicy(allowed_tools=["workflow.read_draft", "workflow.answer_question"]),
            reason="workflow_question",
        )
    return None


def is_task_search(lower: str) -> bool:
    if any(
        marker in lower for marker in ("найди задач", "найти задач", "покажи задач", "что по задач", "какие задач")
    ) or (lower.startswith("найди ") and "задач" in lower):
        return True
    return "задач" in lower and any(
        marker in lower for marker in ("покажи", "отчет", "отчёт", "список", "осталось", "найди", "найти", "какие")
    )


def is_client_count(lower: str) -> bool:
    return "сколько" in lower and "клиент" in lower


def is_client_search(lower: str) -> bool:
    return is_client_list(lower) or any(
        marker in lower
        for marker in (
            "найди клиент",
            "найти клиент",
            "покажи клиент",
        )
    ) or (
        lower.startswith("найди ") and "клиент" in lower
    )


def is_client_list(lower: str) -> bool:
    return any(
        marker in lower
        for marker in (
            "какие клиент",
            "какие у меня клиент",
            "мои клиент",
            "список клиент",
            "кто клиент",
        )
    )


def is_payment_lookup(lower: str) -> bool:
    if "оплат" not in lower and "платеж" not in lower and "платёж" not in lower:
        return False
    return any(
        marker in lower
        for marker in (
            "когда",
            "покажи",
            "какая",
            "какие",
            "что по",
            "по нему",
            "по ней",
            "ближайш",
            "следующ",
            "предстоящ",
            "у кого",
        )
    )


def is_project_payment_lookup(lower: str) -> bool:
    if "оплат" not in lower and "платеж" not in lower and "платёж" not in lower:
        return False
    if any(marker in lower for marker in ("у кого", "кто должен")):
        return True
    if any(marker in lower for marker in ("по проекту", "в проекте", "по всем клиент", "у всех клиент")):
        return True
    if any(marker in lower for marker in ("по нему", "по ней", "у него", "у нее", "у неё")):
        return False
    if " у " in f" {lower} ":
        return False
    return any(marker in lower for marker in ("ближайш", "предстоящ", "следующ"))


def is_task_payment_read(lower: str) -> bool:
    return any(marker in lower for marker in ("покажи", "что", "какие")) and "задач" in lower and "оплат" in lower


def is_compound_client_task(lower: str) -> bool:
    return ("создай клиента" in lower or "создать клиента" in lower) and "задач" in lower and "для клиента" not in lower


def is_task_create(lower: str) -> bool:
    if any(marker in lower for marker in ("создай зада", "создать зада", "добавь зада", "поставь зада")):
        return True
    return "задач" in lower and any(verb in lower for verb in ("создай", "создать", "добавь", "поставь"))


def is_task_move(lower: str) -> bool:
    has_deictic = any(marker in lower for marker in ("эту задач", "эту ", "ее ", "её ", "текущ"))
    has_action = any(marker in lower for marker in ("запусти", "стартани", "перенеси", "передвинь", "перемести"))
    return has_deictic and has_action and ("зада" in lower or "ее" in lower or "её" in lower)


def is_client_create(lower: str) -> bool:
    return "создай клиент" in lower or "создать клиент" in lower


def is_proposal_edit_phrase(lower: str) -> bool:
    return any(
        marker in lower
        for marker in (
            "на сегодня",
            "сегодня",
            "на завтра",
            "завтра",
            "послезавтра",
            "срочно",
            "без клиента",
            "поменяй клиента",
            "назови ",
            "таймер не запускай",
            "приоритет",
        )
    )


def extract_url(text: str) -> str:
    match = re.search(r"https?://\S+", text)
    return match.group(0).rstrip(".,)") if match else ""


def extract_after_markers(text: str, markers: tuple[str, ...]) -> str:
    words = text.split()
    lowered = [word.lower().strip(".,!?") for word in words]
    for marker in markers:
        if marker in lowered:
            idx = lowered.index(marker)
            return " ".join(word.strip(".,!?") for word in words[idx + 1 : idx + 3]).strip()
    return ""
