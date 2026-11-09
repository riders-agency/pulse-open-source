from __future__ import annotations

from dataclasses import dataclass

from pulse_agent.graph.intent_schema import IntentKey


@dataclass(frozen=True)
class IntentRoute:
    branch: str
    active_mode: str
    status_label: str


INTENT_ROUTES: dict[str, IntentRoute] = {
    IntentKey.CONVERSATION_SMALLTALK.value: IntentRoute("conversation_reply", "conversation", "conversation"),
    IntentKey.UTILITY_CALCULATOR.value: IntentRoute("calculator_reply", "conversation", "conversation"),
    IntentKey.PAGE_WHERE_AM_I.value: IntentRoute("read_context", "navigation", "navigation"),
    IntentKey.PAGE_VERIFY_CONTEXT.value: IntentRoute("read_context", "navigation", "navigation"),
    IntentKey.PROJECT_EXPLAIN.value: IntentRoute("read_context", "project_read", "project_read"),
    IntentKey.CLIENT_SEARCH.value: IntentRoute("read_context", "project_read", "project_read"),
    IntentKey.CLIENT_READ.value: IntentRoute("read_context", "project_read", "project_read"),
    IntentKey.CLIENT_COUNT.value: IntentRoute("read_context", "project_read", "project_read"),
    IntentKey.CLIENT_CREATE.value: IntentRoute("client_create_workflow", "client_create", "client_create"),
    IntentKey.CLIENT_UPDATE.value: IntentRoute("conversation_reply", "error_recovery", "error_recovery"),
    IntentKey.TASK_SEARCH.value: IntentRoute("read_context", "project_read", "project_read"),
    IntentKey.TASK_READ.value: IntentRoute("read_context", "project_read", "project_read"),
    IntentKey.TASK_CREATE.value: IntentRoute("task_create_workflow", "task_create", "task_create"),
    IntentKey.TASK_UPDATE.value: IntentRoute("conversation_reply", "error_recovery", "error_recovery"),
    IntentKey.TASK_MOVE.value: IntentRoute("task_move_workflow", "task_update", "task_update"),
    IntentKey.PAYMENT_LOOKUP_DUE.value: IntentRoute("read_context", "project_read", "project_read"),
    IntentKey.PAYMENT_SCHEDULE_READ.value: IntentRoute("read_context", "project_read", "project_read"),
    IntentKey.PAYMENT_ACTION_CREATE.value: IntentRoute("conversation_reply", "error_recovery", "error_recovery"),
    IntentKey.PAYMENT_CONFIRM.value: IntentRoute("conversation_reply", "error_recovery", "error_recovery"),
    IntentKey.WORKFLOW_ANSWER_QUESTION.value: IntentRoute("workflow_question", "project_read", "project_read"),
    IntentKey.WORKFLOW_READ_DRAFT.value: IntentRoute("workflow_question", "project_read", "project_read"),
    IntentKey.WORKFLOW_READ_CONTEXT.value: IntentRoute("read_context", "project_read", "project_read"),
    IntentKey.WORKFLOW_COMPILE_PROMPT.value: IntentRoute("read_context", "project_read", "project_read"),
    IntentKey.WORKFLOW_PATCH.value: IntentRoute("workflow_patch_workflow", "workflow_patch", "workflow_patch"),
    IntentKey.WORKFLOW_RUN_TEST.value: IntentRoute("conversation_reply", "error_recovery", "error_recovery"),
    IntentKey.TASK_PROCESS_READ.value: IntentRoute("read_context", "project_read", "project_read"),
    IntentKey.TASK_PROCESS_PATCH.value: IntentRoute("conversation_reply", "error_recovery", "error_recovery"),
    IntentKey.PROMPT_COPY.value: IntentRoute("read_context", "project_read", "project_read"),
    IntentKey.INTEGRATION_CATALOG_LIST.value: IntentRoute("read_context", "project_read", "project_read"),
    IntentKey.INTEGRATION_DEPENDENCY_READ.value: IntentRoute("read_context", "project_read", "project_read"),
    IntentKey.INTEGRATION_DIAGNOSTICS.value: IntentRoute("conversation_reply", "error_recovery", "error_recovery"),
    IntentKey.INTEGRATION_SECRET_ROTATE.value: IntentRoute("conversation_reply", "error_recovery", "error_recovery"),
    IntentKey.PROPOSAL_EDIT.value: IntentRoute("proposal_editor", "resume_pending_action", "resume_pending_action"),
    IntentKey.APPROVAL_APPLY.value: IntentRoute("approval_decision", "approval", "approval"),
    IntentKey.APPROVAL_REJECT.value: IntentRoute("approval_decision", "approval", "approval"),
    IntentKey.MULTI_STEP_PLAN.value: IntentRoute("compound_workflow", "client_create", "client_create"),
    IntentKey.UNSAFE.value: IntentRoute("conversation_reply", "error_recovery", "error_recovery"),
    IntentKey.UNKNOWN.value: IntentRoute("conversation_reply", "error_recovery", "error_recovery"),
}

LEGACY_PLAN_ROUTES: dict[str, IntentRoute] = {
    "conversation": INTENT_ROUTES[IntentKey.CONVERSATION_SMALLTALK.value],
    "navigation.query": INTENT_ROUTES[IntentKey.PAGE_WHERE_AM_I.value],
    "compound.client_create_and_task_create": INTENT_ROUTES[IntentKey.MULTI_STEP_PLAN.value],
    "integration.diagnose": IntentRoute("read_context", "project_read", "project_read"),
    "prompt.inspect": IntentRoute("read_context", "project_read", "project_read"),
    "question.answer": IntentRoute("read_context", "project_read", "project_read"),
    "task_process.patch": INTENT_ROUTES[IntentKey.TASK_PROCESS_PATCH.value],
    "unsafe": INTENT_ROUTES[IntentKey.UNSAFE.value],
    "workflow.explain": IntentRoute("read_context", "project_read", "project_read"),
    "workflow.run_test": INTENT_ROUTES[IntentKey.WORKFLOW_RUN_TEST.value],
}

STATUS_LABELS = {
    "approval": "Режим: approval",
    "client_create": "Режим: client_create",
    "conversation": "Режим: conversation",
    "error_recovery": "Режим: error_recovery",
    "navigation": "Режим: navigation",
    "project_read": "Режим: project_read",
    "resume_pending_action": "Режим: resume_pending_action",
    "task_create": "Режим: task_create",
    "task_update": "Режим: task_update",
    "workflow_patch": "Режим: workflow_patch",
}


def validate_intent_routes() -> None:
    missing = [intent.value for intent in IntentKey if intent.value not in INTENT_ROUTES]
    if missing:
        raise ValueError(f"Intent route registry is missing routes: {', '.join(missing)}")


def route_for_intent(intent: str) -> IntentRoute:
    validate_intent_routes()
    route = INTENT_ROUTES.get(intent) or LEGACY_PLAN_ROUTES.get(intent)
    if route is None:
        raise ValueError(f"Unhandled intent route: {intent}")
    return route


def branch_for_intent(intent: str) -> str:
    return route_for_intent(intent).branch


def active_mode_for_intent(intent: str) -> str:
    return route_for_intent(intent).active_mode


def mode_status(intent: str) -> str:
    route = route_for_intent(intent)
    return STATUS_LABELS.get(route.status_label, "Режим: error_recovery")
