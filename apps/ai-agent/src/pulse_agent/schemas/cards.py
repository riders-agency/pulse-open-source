from __future__ import annotations

from datetime import UTC, date, datetime
from enum import StrEnum
from typing import Any, Literal

from pydantic import BaseModel, Field


class CardStatus(StrEnum):
    DRAFT = "draft"
    PROPOSED = "proposed"
    NEEDS_APPROVAL = "needs_approval"
    APPROVED = "approved"
    RUNNING = "running"
    SUCCEEDED = "succeeded"
    FAILED = "failed"
    REJECTED = "rejected"
    CANCELED = "canceled"


class CardActionStyle(StrEnum):
    PRIMARY = "primary"
    SECONDARY = "secondary"
    DANGER = "danger"
    GHOST = "ghost"


class CardField(BaseModel):
    key: str
    label: str
    value: Any = None
    value_type: str = "text"
    state: Literal["normal", "missing", "warning", "error", "success"] = "normal"
    editable: bool = False
    edit_schema_json: dict[str, Any] | None = None


class CardSection(BaseModel):
    key: str
    title: str
    fields: list[CardField] = Field(default_factory=list)
    items: list[str] = Field(default_factory=list)


class EntityLink(BaseModel):
    entity_type: str
    entity_id: str
    label: str
    url: str


class CardAction(BaseModel):
    action_key: str
    label: str
    style: CardActionStyle = CardActionStyle.SECONDARY
    disabled: bool = False
    confirm_required: bool = False
    payload_json: dict[str, Any] = Field(default_factory=dict)


class BaseCard(BaseModel):
    schema_: str = Field(default="riders-pulse-ai-card-v1", alias="schema")
    card_id: str
    card_type: str
    version: int = 1
    revision: int = 1
    tool_call_id: str = ""
    intent_key: str = ""
    status: CardStatus = CardStatus.DRAFT
    title: str
    subtitle: str = ""
    risk_level: str = "read"
    requires_approval: bool = False
    fields: list[CardField] = Field(default_factory=list)
    sections: list[CardSection] = Field(default_factory=list)
    actions: list[CardAction] = Field(default_factory=list)
    display_json: dict[str, Any] = Field(default_factory=dict)
    input_json: dict[str, Any] = Field(default_factory=dict)
    entity_links: list[EntityLink] = Field(default_factory=list)
    debug_json: dict[str, Any] = Field(default_factory=dict)
    created_at: datetime = Field(default_factory=lambda: datetime.now(UTC))
    updated_at: datetime = Field(default_factory=lambda: datetime.now(UTC))


def task_create_card(card_id: str, draft: dict[str, Any], revision: int = 1) -> BaseCard:
    client_name = draft.get("client_name") or draft.get("client_query") or ""
    missing_client = bool(draft.get("client_query") and not draft.get("client_id"))
    due_date_display = task_due_date_display(draft)
    fields = [
        CardField(
            key="client",
            label="Клиент",
            value=client_name,
            state="missing" if missing_client else "normal",
            editable=True,
        ),
        CardField(key="title", label="Название", value=draft.get("title"), editable=True),
        CardField(
            key="description", label="Описание", value=draft.get("description"), value_type="text", editable=True
        ),
        CardField(key="due_date", label="Срок", value=due_date_display, value_type="date", editable=True),
        CardField(
            key="priority", label="Приоритет", value=draft.get("priority", "medium"), value_type="badge", editable=True
        ),
        CardField(key="labels", label="Метки", value=draft.get("labels", []), value_type="list", editable=True),
        CardField(
            key="notification_intent", label="Уведомление", value=draft.get("notification_intent", ""), editable=True
        ),
        CardField(
            key="start_timer_after_create",
            label="Таймер",
            value=bool(draft.get("start_timer_after_create")),
            value_type="bool",
            editable=True,
        ),
    ]
    sections = [
        CardSection(key="checklist", title="Чеклист", items=list(draft.get("checklist") or [])),
    ]
    return BaseCard(
        card_id=card_id,
        card_type="task_create_proposal",
        tool_call_id="pending-task-create",
        intent_key="task.create",
        revision=revision,
        status=CardStatus.NEEDS_APPROVAL,
        title="Создание задачи",
        subtitle="Проверьте поля и подтвердите создание",
        risk_level="high",
        requires_approval=True,
        fields=fields,
        sections=sections,
        actions=[
            CardAction(
                action_key="approve", label="Создать задачу", style=CardActionStyle.PRIMARY, confirm_required=True
            ),
            CardAction(action_key="reject", label="Отклонить", style=CardActionStyle.SECONDARY),
            CardAction(action_key="edit", label="Изменить", style=CardActionStyle.GHOST),
        ],
        display_json={**draft, "due_date_display": due_date_display},
        input_json=draft,
        debug_json={"input_json": draft, "hidden_by_default": True},
    )


def task_due_date_display(draft: dict[str, Any]) -> str:
    raw_due = str(draft.get("due_date") or "").strip()
    label = str(draft.get("due_date_label") or "").strip()
    formatted = format_iso_date_ru(raw_due)
    if label and formatted:
        return f"{label} · {formatted}"
    return formatted or label or raw_due


def format_iso_date_ru(value: str) -> str:
    try:
        parsed = date.fromisoformat(value)
    except ValueError:
        return ""
    return parsed.strftime("%d.%m.%Y")


def client_create_card(card_id: str, draft: dict[str, Any], revision: int = 1) -> BaseCard:
    name = str(draft.get("name") or draft.get("client_name") or "").strip()
    return BaseCard(
        card_id=card_id,
        card_type="client_create_proposal",
        tool_call_id="pending-client-create",
        intent_key="client.create",
        revision=revision,
        status=CardStatus.NEEDS_APPROVAL,
        title="Создание клиента",
        subtitle="Проверьте карточку и подтвердите создание клиента",
        risk_level="high",
        requires_approval=True,
        fields=[
            CardField(
                key="name",
                label="Название",
                value=name,
                state="missing" if not name else "normal",
                editable=True,
            ),
            CardField(
                key="client_type",
                label="Тип",
                value=draft.get("client_type", "company"),
                value_type="badge",
                editable=True,
            ),
        ],
        actions=[
            CardAction(
                action_key="approve",
                label="Создать клиента",
                style=CardActionStyle.PRIMARY,
                confirm_required=True,
            ),
            CardAction(action_key="reject", label="Отклонить", style=CardActionStyle.SECONDARY),
            CardAction(action_key="edit", label="Изменить", style=CardActionStyle.GHOST),
        ],
        display_json={"name": name, "client_type": draft.get("client_type", "company")},
        input_json={"name": name, "client_type": draft.get("client_type", "company")},
        debug_json={
            "input_json": {"name": name, "client_type": draft.get("client_type", "company")},
            "hidden_by_default": True,
        },
    )
