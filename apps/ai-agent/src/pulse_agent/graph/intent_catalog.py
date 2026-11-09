from __future__ import annotations

from dataclasses import dataclass
from typing import Literal

from pulse_agent.graph.intent_schema import IntentKey, TurnKind

RiskLevel = Literal["read", "medium", "high", "critical", "unsafe"]
TargetEntityType = Literal[
    "task",
    "client",
    "project",
    "workflow",
    "payment",
    "integration",
    "provider",
    "task_process",
    "none",
]


@dataclass(frozen=True)
class CatalogTool:
    key: str
    risk_level: RiskLevel = "read"
    requires_approval: bool = False
    reason: str = "model_understanding"


@dataclass(frozen=True)
class IntentCatalogRoute:
    key: IntentKey
    domain: str
    turn_kind: TurnKind
    target_entity_type: TargetEntityType
    tools: tuple[CatalogTool, ...]
    description: str
    examples: tuple[str, ...] = ()

    @property
    def risk_level(self) -> RiskLevel:
        order = {"read": 0, "medium": 1, "high": 2, "critical": 3, "unsafe": 4}
        return max((tool.risk_level for tool in self.tools), key=lambda value: order[value], default="read")

    @property
    def requires_approval(self) -> bool:
        return any(tool.requires_approval for tool in self.tools)


def read_tool(key: str, reason: str = "model_understanding") -> CatalogTool:
    return CatalogTool(key=key, reason=reason)


def write_tool(key: str, risk: RiskLevel = "high", reason: str = "model_write_intent") -> CatalogTool:
    return CatalogTool(key=key, risk_level=risk, requires_approval=True, reason=reason)


MODEL_INTENT_ROUTES: tuple[IntentCatalogRoute, ...] = (
    IntentCatalogRoute(
        IntentKey.PAGE_WHERE_AM_I,
        "project",
        TurnKind.PAGE_CONTEXT,
        "project",
        (read_tool("ui_context.read", "current_page"),),
        "Current screen, route or visible Pulse surface.",
        ("Где я сейчас?", "What screen am I on?"),
    ),
    IntentCatalogRoute(
        IntentKey.PAGE_VERIFY_CONTEXT,
        "project",
        TurnKind.PAGE_CONTEXT,
        "project",
        (read_tool("ui_context.read", "verify_page_context"),),
        "Verify whether current UI context matches the user's expectation.",
        ("Это точно экран задач?", "Проверь контекст страницы"),
    ),
    IntentCatalogRoute(
        IntentKey.PROJECT_EXPLAIN,
        "project",
        TurnKind.DOMAIN_QUESTION,
        "project",
        (read_tool("context.read", "project_explain"),),
        "Explain the current project, company, counts and visible project context.",
        ("Объясни текущий проект", "Что это за проект?"),
    ),
    IntentCatalogRoute(
        IntentKey.CLIENT_SEARCH,
        "client",
        TurnKind.READ_TOOL,
        "client",
        (read_tool("client.search", "client_read"),),
        "List, search or show project clients.",
        ("Выведи клиентов", "Какие у меня клиенты?", "Find the Adv account"),
    ),
    IntentCatalogRoute(
        IntentKey.CLIENT_READ,
        "client",
        TurnKind.READ_TOOL,
        "client",
        (read_tool("client.read", "client_read"),),
        "Read details of a specific client when a client is named or selected.",
        ("Покажи карточку клиента Орион", "Что известно по клиенту Adv?"),
    ),
    IntentCatalogRoute(
        IntentKey.CLIENT_COUNT,
        "client",
        TurnKind.READ_TOOL,
        "client",
        (read_tool("client.count", "client_count"),),
        "Count project clients.",
        ("Сколько у меня клиентов?",),
    ),
    IntentCatalogRoute(
        IntentKey.CLIENT_CREATE,
        "client",
        TurnKind.WRITE_INTENT,
        "client",
        (write_tool("client.create", "medium", "client_create"),),
        "Create a new client proposal.",
        ("Создай клиента Орион",),
    ),
    IntentCatalogRoute(
        IntentKey.CLIENT_UPDATE,
        "client",
        TurnKind.WRITE_INTENT,
        "client",
        (read_tool("client.search", "resolve_client"), write_tool("client.update", "medium", "client_update")),
        "Update an existing client after approval.",
        ("Поменяй телефон клиента Орион",),
    ),
    IntentCatalogRoute(
        IntentKey.TASK_SEARCH,
        "task",
        TurnKind.READ_TOOL,
        "task",
        (read_tool("task.search", "task_read"),),
        "Search, list or report on tasks.",
        ("Покажи задачи по Орион", "Какие задачи просрочены?"),
    ),
    IntentCatalogRoute(
        IntentKey.TASK_READ,
        "task",
        TurnKind.READ_TOOL,
        "task",
        (read_tool("task.read", "task_read"),),
        "Read a specific task when task identity is named or selected.",
        ("Покажи эту задачу", "Что в задаче RP-12?"),
    ),
    IntentCatalogRoute(
        IntentKey.TASK_CREATE,
        "task",
        TurnKind.WRITE_INTENT,
        "task",
        (read_tool("client.search", "resolve_client"), write_tool("task.create", "high", "task_create")),
        "Create a task proposal, optionally linked to a client.",
        ("Создай задачу проверить оплату Орион завтра",),
    ),
    IntentCatalogRoute(
        IntentKey.TASK_UPDATE,
        "task",
        TurnKind.WRITE_INTENT,
        "task",
        (write_tool("task.update", "high", "task_update"),),
        "Update a task after approval.",
        ("Поменяй дедлайн задачи на пятницу",),
    ),
    IntentCatalogRoute(
        IntentKey.TASK_MOVE,
        "task",
        TurnKind.WRITE_INTENT,
        "task",
        (write_tool("task.move", "high", "task_move"),),
        "Move or start a selected task after approval.",
        ("Запусти эту задачу", "Перенеси задачу в работу"),
    ),
    IntentCatalogRoute(
        IntentKey.PAYMENT_LOOKUP_DUE,
        "payment",
        TurnKind.READ_TOOL,
        "payment",
        (read_tool("client.search", "resolve_client"), read_tool("payment.lookup_due", "payment_read")),
        "Find nearest due payments for a named client or for the project.",
        ("Когда оплата у Орион?", "У кого ближайшие оплаты?"),
    ),
    IntentCatalogRoute(
        IntentKey.PAYMENT_SCHEDULE_READ,
        "payment",
        TurnKind.READ_TOOL,
        "payment",
        (read_tool("payment.schedule.read", "payment_schedule_read"),),
        "Read a payment schedule for a client or project.",
        ("Покажи график оплат Орион",),
    ),
    IntentCatalogRoute(
        IntentKey.PAYMENT_ACTION_CREATE,
        "payment",
        TurnKind.WRITE_INTENT,
        "payment",
        (
            read_tool("client.search", "resolve_client"),
            write_tool("payment.action.create", "medium", "payment_action_create"),
        ),
        "Create a payment action or reminder after approval.",
        ("Создай напоминание об оплате Орион",),
    ),
    IntentCatalogRoute(
        IntentKey.PAYMENT_CONFIRM,
        "payment",
        TurnKind.WRITE_INTENT,
        "payment",
        (write_tool("payment.action.confirm", "high", "payment_confirm"),),
        "Confirm that a payment was received after approval.",
        ("Подтверди оплату по Орион",),
    ),
    IntentCatalogRoute(
        IntentKey.WORKFLOW_ANSWER_QUESTION,
        "workflow",
        TurnKind.WORKFLOW_QUESTION,
        "workflow",
        (read_tool("workflow.read_draft", "workflow_question"),),
        "Answer questions about a workflow graph or draft.",
        ("Что делает этот workflow?", "Сколько здесь действий?"),
    ),
    IntentCatalogRoute(
        IntentKey.WORKFLOW_READ_DRAFT,
        "workflow",
        TurnKind.WORKFLOW_QUESTION,
        "workflow",
        (read_tool("workflow.read_draft", "workflow_read_draft"),),
        "Read the selected workflow draft.",
        ("Покажи draft workflow",),
    ),
    IntentCatalogRoute(
        IntentKey.WORKFLOW_READ_CONTEXT,
        "workflow",
        TurnKind.WORKFLOW_QUESTION,
        "workflow",
        (read_tool("workflow.read_context", "workflow_read_context"),),
        "Read workflow context around a selected workflow or run.",
        ("Покажи контекст workflow",),
    ),
    IntentCatalogRoute(
        IntentKey.WORKFLOW_COMPILE_PROMPT,
        "workflow",
        TurnKind.WORKFLOW_QUESTION,
        "workflow",
        (read_tool("workflow.compile_prompt", "workflow_compile_prompt"),),
        "Compile or inspect the prompt context used for workflow AI operations.",
        ("Собери prompt для этого workflow", "Покажи промпт workflow"),
    ),
    IntentCatalogRoute(
        IntentKey.PROMPT_COPY,
        "workflow",
        TurnKind.READ_TOOL,
        "workflow",
        (read_tool("workflow.compile_prompt", "prompt_copy"),),
        "Prepare a redacted prompt snapshot for copying.",
        ("Скопируй промпт", "Покажи redacted prompt"),
    ),
    IntentCatalogRoute(
        IntentKey.WORKFLOW_PATCH,
        "workflow",
        TurnKind.WORKFLOW_PATCH,
        "workflow",
        (
            read_tool("workflow.read_draft", "workflow_read_before_patch"),
            write_tool("workflow.patch", "critical", "workflow_patch"),
        ),
        "Propose and apply workflow graph changes only after approval.",
        ("Добавь MAX уведомление в workflow",),
    ),
    IntentCatalogRoute(
        IntentKey.WORKFLOW_RUN_TEST,
        "workflow",
        TurnKind.WRITE_INTENT,
        "workflow",
        (write_tool("workflow.run.test", "high", "workflow_run_test"),),
        "Run a workflow test after approval.",
        ("Запусти тест этого workflow",),
    ),
    IntentCatalogRoute(
        IntentKey.TASK_PROCESS_READ,
        "task_process",
        TurnKind.READ_TOOL,
        "task_process",
        (read_tool("task_process.read", "task_process_read"),),
        "Read task process statuses, transitions and board rules.",
        ("Покажи процесс задач", "Какие статусы на доске?"),
    ),
    IntentCatalogRoute(
        IntentKey.TASK_PROCESS_PATCH,
        "task_process",
        TurnKind.WORKFLOW_PATCH,
        "task_process",
        (
            read_tool("task_process.read", "task_process_read_before_patch"),
            write_tool("task_process.patch.apply", "critical", "task_process_patch"),
        ),
        "Patch task process rules only after approval.",
        ("Добавь статус review в процесс задач",),
    ),
    IntentCatalogRoute(
        IntentKey.INTEGRATION_CATALOG_LIST,
        "integration",
        TurnKind.READ_TOOL,
        "integration",
        (read_tool("integration.catalog.list", "integration_catalog_list"),),
        "List available integration providers and catalog entries.",
        ("Какие интеграции доступны?", "Покажи каталог интеграций"),
    ),
    IntentCatalogRoute(
        IntentKey.INTEGRATION_DEPENDENCY_READ,
        "integration",
        TurnKind.READ_TOOL,
        "integration",
        (read_tool("integration.dependency.read", "integration_dependency_read"),),
        "Read provider/dependency configuration status without exposing secrets.",
        ("Какая модель используется?", "Покажи AI provider", "Проверь зависимость Apinet"),
    ),
    IntentCatalogRoute(
        IntentKey.INTEGRATION_DIAGNOSTICS,
        "integration",
        TurnKind.WRITE_INTENT,
        "integration",
        (write_tool("integration.diagnostics.run", "high", "integration_diagnostics"),),
        "Run integration diagnostics after approval.",
        ("Проверь AI интеграцию проекта", "Протестируй интеграцию MAX"),
    ),
    IntentCatalogRoute(
        IntentKey.INTEGRATION_SECRET_ROTATE,
        "integration",
        TurnKind.WRITE_INTENT,
        "integration",
        (write_tool("integration.secret.rotate", "critical", "integration_secret_rotate"),),
        "Rotate an integration secret after explicit approval.",
        ("Ротируй ключ Apinet",),
    ),
)

MODEL_INTENT_ROUTE_BY_KEY = {route.key: route for route in MODEL_INTENT_ROUTES}
READ_MODEL_INTENTS = {
    route.key for route in MODEL_INTENT_ROUTES if route.risk_level == "read" and not route.requires_approval
}


def route_for_model_intent(intent_key: IntentKey) -> IntentCatalogRoute | None:
    return MODEL_INTENT_ROUTE_BY_KEY.get(intent_key)


def model_prompt_intent_catalog() -> str:
    rows = []
    for route in MODEL_INTENT_ROUTES:
        tools = ",".join(tool.key for tool in route.tools)
        examples = "; ".join(route.examples[:2])
        rows.append(
            f"- {route.key.value}: domain={route.domain}; risk={route.risk_level}; "
            f"approval={str(route.requires_approval).lower()}; tools={tools}; "
            f"use={route.description}; examples={examples}"
        )
    return "\n".join(rows)
