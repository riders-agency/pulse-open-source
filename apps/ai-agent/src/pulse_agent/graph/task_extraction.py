from __future__ import annotations

import re
from contextlib import suppress
from dataclasses import dataclass, field
from datetime import date, timedelta

STOP_WORDS = {
    "сегодня",
    "завтра",
    "послезавтра",
    "сделать",
    "поправить",
    "подготовить",
    "подготовь",
    "подготовка",
    "правк",
    "правка",
    "правки",
    "и",
    "сообщить",
    "запусти",
    "запустить",
    "стартани",
    "таймер",
    "на",
    "по",
    "ней",
    "нему",
}

WEEKDAY_MARKERS = {
    "понедельник": 0,
    "понедельника": 0,
    "вторник": 1,
    "вторника": 1,
    "среду": 2,
    "среда": 2,
    "четверг": 3,
    "четверга": 3,
    "пятницу": 4,
    "пятница": 4,
    "субботу": 5,
    "суббота": 5,
    "воскресенье": 6,
    "воскресенья": 6,
}


@dataclass
class TaskDraft:
    title: str
    description: str
    client_query: str = ""
    client_id: str = ""
    client_name: str = ""
    due_date: str = ""
    due_date_label: str = ""
    priority: str = "medium"
    labels: list[str] = field(default_factory=lambda: ["ai-created"])
    assignee_hint: str = ""
    notification_intent: str = ""
    checklist: list[str] = field(default_factory=list)
    start_timer_after_create: bool = False
    start_task_after_create: bool = False
    missing_fields: list[dict[str, object]] = field(default_factory=list)

    def to_tool_input(self) -> dict[str, object]:
        return {
            "title": self.title,
            "description": self.description,
            "client_query": self.client_query,
            "client_id": self.client_id,
            "client_name": self.client_name,
            "due_date": self.due_date,
            "due_date_label": self.due_date_label,
            "priority": self.priority,
            "labels": self.labels,
            "assignee_hint": self.assignee_hint,
            "notification_intent": self.notification_intent,
            "checklist": self.checklist,
            "start_timer_after_create": self.start_timer_after_create,
            "start_task_after_create": self.start_task_after_create,
            "missing_fields": self.missing_fields,
        }


def parse_task_draft(text: str, *, local_date: str = "", timezone: str = "") -> TaskDraft:
    del timezone
    source = text.strip()
    lower = source.lower()
    client_query = extract_client_query(source)
    due, due_label = resolve_due_date(lower, local_date)
    seo = contains_seo(lower)
    labels = ["ai-created"]
    if seo:
        labels.append("seo")
    title = "Внести правки для SEO-специалистов" if seo else clean_title(source, client_query)
    if not title:
        title = "Проверить оплату" if "оплат" in lower else "Новая задача"
    description = (
        "Сделать все правки для SEO-специалистов и сообщить им о результате."
        if seo
        else f"Создано AI-сотрудником из запроса: {source}"
    )
    checklist = []
    if seo:
        checklist = [
            "Собрать список правок",
            "Внести правки",
            "Проверить результат",
            "Сообщить SEO-специалистам",
        ]
    notification = "Сообщить SEO-специалистам после выполнения" if "сообщ" in lower else ""
    draft = TaskDraft(
        title=title,
        description=description,
        client_query=client_query,
        due_date=due,
        due_date_label=due_label,
        labels=labels,
        assignee_hint="me" if "мне" in lower or "на меня" in lower else "",
        notification_intent=notification,
        checklist=checklist,
        start_timer_after_create=("таймер" in lower or "трекер времени" in lower),
        start_task_after_create=(("запусти" in lower or "стартани" in lower) and "задач" in lower),
    )
    if client_query and not draft.client_id:
        draft.missing_fields.append({"field": "client_id", "label": "Клиент", "query": client_query})
    return draft


def contains_seo(lower: str) -> bool:
    return "seo" in lower or "сео" in lower or "сеош" in lower or "сешо" in lower


def extract_client_query(text: str) -> str:
    patterns = [
        r"(?i)(?:для|по)\s+кли[а-яёe]*\s+([\wА-Яа-яЁё\-\"«»\. ]{2,80})",
        r"(?i)кли[а-яёe]*\s+([\wА-Яа-яЁё\-\"«»\. ]{2,80})",
        r"(?i)оплат[а-яё]*\s+(?:у\s+|по\s+)?([\wА-Яа-яЁё\-\"«»\. ]{2,80})",
        r"(?i)(?:для|по)\s+([\wА-Яа-яЁё\-\"«»\. ]{2,80})",
        r"(?i)у\s+([\wА-Яа-яЁё\-\"«»\. ]{2,80})",
    ]
    for pattern in patterns:
        match = re.search(r"\b" + pattern.removeprefix("(?i)"), text, re.IGNORECASE)
        if match:
            return trim_entity_name(match.group(1))
    return ""


def trim_entity_name(value: str) -> str:
    parts = re.split(r"\s+", value.strip(" \t\r\n.,!?\"'«»"))
    out: list[str] = []
    for part in parts:
        clean = part.strip(".,!?\"'«»")
        lower = clean.lower()
        if lower in STOP_WORDS or lower.startswith("правк"):
            if out:
                break
            continue
        out.append(clean)
    return " ".join(out).strip()


def resolve_due_date(lower: str, local_date: str) -> tuple[str, str]:
    base = date.today()
    if local_date:
        with suppress(ValueError):
            base = date.fromisoformat(local_date)
    if "послезавтра" in lower:
        return (base + timedelta(days=2)).isoformat(), "послезавтра"
    if "завтра" in lower:
        return (base + timedelta(days=1)).isoformat(), "завтра"
    if "сегодня" in lower:
        return base.isoformat(), "сегодня"
    for marker, weekday in WEEKDAY_MARKERS.items():
        if marker in lower:
            days = (weekday - base.weekday()) % 7
            if days == 0 or "следующ" in lower:
                days = days or 7
            return (base + timedelta(days=days)).isoformat(), marker
    return "", ""


def clean_title(text: str, client_query: str) -> str:
    command = (
        r"(?i)\b(?:создай(?:те)?|создать|добавь(?:те)?|добавить|поставь(?:те)?|поставить|"
        r"заведи(?:те)?|завести|запланируй(?:те)?|запланировать)\s+задач[ауи]?\b"
    )
    cleaned = re.sub(command, " ", text)
    cleaned = re.sub(r"(?i)\bмне\b", " ", cleaned)
    cleaned = re.sub(r"(?i)\b(?:на\s+)?(сегодня|завтра|послезавтра)\b", " ", cleaned)
    if client_query:
        cleaned = cleaned.replace(client_query, " ")
    cleaned = re.sub(r"(?i)(?:для|по)\s+кли[а-яёe]*", " ", cleaned)
    cleaned = re.sub(r"\s+", " ", cleaned).strip(" .,!?")
    if cleaned:
        return cleaned[:1].upper() + cleaned[1:]
    return ""
