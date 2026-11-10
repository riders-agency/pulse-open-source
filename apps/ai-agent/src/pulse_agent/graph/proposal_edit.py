from __future__ import annotations

from dataclasses import dataclass

from pulse_agent.graph.task_extraction import extract_client_query, resolve_due_date


@dataclass
class ProposalEditResult:
    changed: bool
    patch: dict[str, object]
    human_summary: str


def resolve_task_proposal_edit(
    text: str, current_input: dict[str, object], *, local_date: str = ""
) -> ProposalEditResult:
    lower = text.lower().strip()
    if is_new_request(lower) or is_read_question(lower):
        return ProposalEditResult(False, {}, "")
    patch: dict[str, object] = {}
    summaries: list[str] = []
    due, due_label = resolve_due_date(lower, local_date)
    if due:
        patch["due_date"] = due
        patch["due_date_label"] = due_label
        summaries.append(f"срок — {due_label}")
    if "срочно" in lower or "urgent" in lower:
        patch["priority"] = "urgent"
        summaries.append("приоритет — срочный")
    elif "высок" in lower or "high" in lower:
        patch["priority"] = "high"
        summaries.append("приоритет — высокий")
    elif "низк" in lower or "low" in lower:
        patch["priority"] = "low"
        summaries.append("приоритет — низкий")
    elif "средн" in lower or "medium" in lower:
        patch["priority"] = "medium"
        summaries.append("приоритет — средний")
    if "без клиента" in lower or "клиента не надо" in lower:
        patch.update({"client_id": "", "client_name": "", "client_query": ""})
        summaries.append("клиент очищен")
    else:
        query = extract_client_query(text)
        if query and is_explicit_client_edit(lower):
            patch.update({"client_query": query, "client_id": "", "client_name": ""})
            summaries.append(f"клиент — {query}")
    if lower.startswith("назови ") or "название " in lower:
        title = text.split(" ", 1)[1].strip() if " " in text else ""
        if title:
            patch["title"] = title
            summaries.append("название обновлено")
    if not patch:
        return ProposalEditResult(False, {}, "")
    merged = {**current_input, **patch}
    del merged
    return ProposalEditResult(True, patch, "Обновил текущую карточку: " + ", ".join(summaries) + ".")


def is_read_question(lower: str) -> bool:
    if "?" in lower and any(marker in lower for marker in ("оплат", "когда", "сколько", "статус")):
        return True
    return lower.startswith(("что там", "когда ", "сколько ", "а когда", "найди ", "найти ", "покажи "))


def is_new_request(lower: str) -> bool:
    if lower.startswith(("/", "новый запрос", "новая команда")):
        return True
    create_markers = ("создай ", "создать ", "добавь ", "поставь ")
    if lower.startswith(create_markers) and any(
        entity in lower for entity in ("задач", "клиент", "workflow", "процесс")
    ):
        return True
    return lower.startswith(("объясни ", "расскажи ", "проверь ")) and any(
        entity in lower for entity in ("задач", "клиент", "оплат", "workflow", "проект")
    )


def is_explicit_client_edit(lower: str) -> bool:
    return any(
        marker in lower
        for marker in (
            "клиент ",
            "клиента ",
            "клиент:",
            "для клиента",
            "по клиенту",
        )
    )
