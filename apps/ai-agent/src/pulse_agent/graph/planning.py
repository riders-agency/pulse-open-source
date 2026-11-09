from __future__ import annotations

from dataclasses import dataclass, field

from pulse_agent.graph.task_extraction import parse_task_draft


@dataclass
class CommandPlan:
    intent: str
    confidence: float
    required_tools: list[str] = field(default_factory=list)
    risk_level: str = "read"
    requires_approval: bool = False
    user_visible_summary: str = ""
    tool_input_preview: dict[str, object] = field(default_factory=dict)


def plan_turn(text: str, *, local_date: str = "", timezone: str = "") -> CommandPlan:
    lower = text.lower()
    if is_phatic(lower):
        return CommandPlan(
            intent="conversation",
            confidence=0.98,
            user_visible_summary=(
                "Я на связи. Могу читать контекст проекта, готовить карточки действий "
                "и выполнять изменения после подтверждения."
            ),
        )
    if is_navigation_query(lower):
        return CommandPlan(
            intent="navigation.query",
            confidence=0.96,
            required_tools=["context.read"],
            user_visible_summary="Определить текущий экран по UI context",
        )
    if "api key" in lower or "секрет" in lower or "парол" in lower:
        return CommandPlan(
            intent="unsafe",
            confidence=0.99,
            user_visible_summary="Отказать без раскрытия секретов",
        )
    if "prompt" in lower or "промпт" in lower:
        return CommandPlan(
            intent="prompt.inspect",
            confidence=0.95,
            required_tools=["prompt.snapshot"],
            user_visible_summary="Показать redacted prompt snapshot",
        )
    if "модель" in lower or "apinet" in lower:
        return CommandPlan(
            intent="integration.diagnose",
            confidence=0.9,
            required_tools=["integration.inspect"],
            user_visible_summary="Проверить модель и provider configuration",
        )
    if "тестовый workflow" in lower or ("запусти" in lower and "workflow" in lower):
        return CommandPlan(
            intent="workflow.run_test",
            confidence=0.9,
            required_tools=["workflow.test_run"],
            risk_level="high",
            requires_approval=True,
            user_visible_summary="Запустить тестовый workflow после подтверждения",
        )
    if "почему run" in lower or "run упал" in lower:
        return CommandPlan(
            intent="workflow.explain",
            confidence=0.9,
            required_tools=["context.read"],
            user_visible_summary="Объяснить ошибку run по scoped context",
        )
    if "workflow" in lower or "telegram" in lower or "email" in lower:
        return CommandPlan(
            intent="workflow.patch",
            confidence=0.88,
            required_tools=["workflow.patch.propose"],
            risk_level="critical",
            requires_approval=True,
            user_visible_summary="Подготовить workflow patch preview",
        )
    if "процесс" in lower and ("review" in lower or "статус" in lower):
        return CommandPlan(
            intent="task_process.patch",
            confidence=0.9,
            required_tools=["task_process.patch.propose"],
            risk_level="critical",
            requires_approval=True,
            user_visible_summary="Подготовить patch процесса задач",
        )
    if is_compound_client_task_request(lower):
        client_name = extract_client_name_for_plan(text)
        task_text = extract_compound_task_text(text, client_name)
        task_draft = parse_task_draft(task_text, local_date=local_date, timezone=timezone)
        if not task_draft.client_query:
            task_draft.client_query = client_name
        if not task_draft.client_name:
            task_draft.client_name = client_name
        if "запусти" in lower or "стартани" in lower:
            task_draft.start_task_after_create = True
        return CommandPlan(
            intent="compound.client_create_and_task_create",
            confidence=0.93,
            required_tools=["client.create", "task.create"],
            risk_level="high",
            requires_approval=True,
            user_visible_summary="Создать клиента и подготовить связанную задачу после подтверждения",
            tool_input_preview={
                "client": {"name": client_name},
                "task": task_draft.to_tool_input(),
            },
        )
    if "созда" in lower and "зада" in lower:
        draft = parse_task_draft(text, local_date=local_date, timezone=timezone)
        return CommandPlan(
            intent="task.create",
            confidence=0.96,
            required_tools=["client.search", "task.create"],
            risk_level="high",
            requires_approval=True,
            user_visible_summary="Создать задачу после подтверждения",
            tool_input_preview=draft.to_tool_input(),
        )
    if is_task_search(lower):
        return CommandPlan(
            intent="task.search",
            confidence=0.95,
            required_tools=["context.read"],
            user_visible_summary="Найти задачи через project-scoped read tools",
            tool_input_preview={"query": text, "entity_type": "task"},
        )
    if "когда" in lower and "оплат" in lower:
        return CommandPlan(
            intent="payment.lookup_due",
            confidence=0.95,
            required_tools=["client.search", "payment.lookup_due"],
            user_visible_summary="Найти клиента и дату оплаты через read-only tools",
        )
    if "созда" in lower and "клиент" in lower:
        name = extract_client_name_for_plan(text)
        return CommandPlan(
            intent="client.create",
            confidence=0.9 if name != "Новый клиент" else 0.65,
            required_tools=["client.create"],
            risk_level="high",
            requires_approval=True,
            user_visible_summary="Создать клиента после подтверждения",
            tool_input_preview={"name": name},
        )
    return CommandPlan(intent="question.answer", confidence=0.5, user_visible_summary="Ответить по текущему контексту")


def is_phatic(lower: str) -> bool:
    text = lower.strip(" ?!.")
    return text in {
        "ты тут",
        "ты здесь",
        "работаешь",
        "ау",
        "привет",
        "алло",
        "на связи",
        "ты на связи",
    }


def is_navigation_query(lower: str) -> bool:
    return any(marker in lower for marker in ("где я сейчас", "где мы сейчас", "на какой странице", "какой экран"))


def is_task_search(lower: str) -> bool:
    if any(
        marker in lower for marker in ("найди задач", "найти задач", "покажи задач", "что по задач", "какие задач")
    ) or (lower.startswith("найди ") and "задач" in lower):
        return True
    return "задач" in lower and any(
        marker in lower for marker in ("покажи", "отчет", "отчёт", "список", "осталось", "найди", "найти", "какие")
    )


def is_approval_apply(lower: str) -> bool:
    text = lower.strip(" ?!.")
    if any(marker in text for marker in ("да, созда", "да созда", "подтвержда", "применяй")):
        return True
    return text in {
        "да",
        "давай",
        "да создавай",
        "да, создавай",
        "создавай",
        "подтверждаю",
        "применить",
        "ок применяй",
        "ок, применяй",
    }


def is_approval_reject(lower: str) -> bool:
    text = lower.strip(" ?!.")
    if any(marker in text for marker in ("не надо", "отмен", "не создавай", "отклон")):
        return True
    return text in {
        "нет",
        "не надо",
        "отмена",
        "отмени",
        "сбрось",
        "отклонить",
        "не создавай",
    }


def extract_client_name_for_plan(text: str) -> str:
    lowered = text.lower()
    marker = "клиента"
    idx = lowered.find(marker)
    if idx < 0:
        return "Новый клиент"
    tail = text[idx + len(marker) :].strip()
    words = []
    for word in tail.split():
        if word.lower() in {"и", "поставь", "создай", "задачу"}:
            break
        words.append(word.strip(".,!?«»\"'"))
    return " ".join(words).strip() or "Новый клиент"


def is_compound_client_task_request(lower: str) -> bool:
    return ("создай клиента" in lower or "создать клиента" in lower) and "задач" in lower and "для клиента" not in lower


def extract_compound_task_text(text: str, client_name: str) -> str:
    lowered = text.lower()
    marker_idx = lowered.find("задач")
    if marker_idx < 0:
        return f"создай задачу для клиента {client_name}"
    tail = text[marker_idx:]
    for prefix in ("задачу", "задача", "задачи"):
        if tail.lower().startswith(prefix):
            tail = tail[len(prefix) :].strip()
            break
    tail = tail.replace("ему ", "").replace("ей ", "")
    tail = tail.replace("тоже", "").strip(" .,!?")
    if "договор" in tail.lower():
        return f"создай задачу для клиента {client_name} подготовить договор на продвижение"
    if "документ" in tail.lower():
        return f"создай задачу для клиента {client_name} подготовить документы для начала сотрудничества"
    if not tail:
        tail = "подготовить документы для начала сотрудничества"
    return f"создай задачу для клиента {client_name} {tail}"
