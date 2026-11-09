import type { AIPageContext } from "../../lib/api";

export const AI_EMPLOYEE_STORAGE_PREFIX = "pulse.aiEmployee.activeSession.";
export const AI_EMPLOYEE_LEGACY_STORAGE_PREFIX = "pulse.aiEmployee.session.";

export const AI_EMPLOYEE_TEST_IDS = {
  widget: "ai-employee-widget",
  drawer: "ai-employee-drawer",
  launcher: "ai-employee-launcher",
  chatLog: "ai-chat-log",
  chatEnd: "ai-chat-end",
  userMessage: "ai-message-user",
  assistantMessage: "ai-message-assistant",
  genericMessage: "ai-message",
  progressLane: "ai-progress-lane",
  progressStep: "ai-progress-step",
  composerInput: "ai-composer-input",
  composerSend: "ai-composer-send",
  composerStop: "ai-composer-stop",
  approvalApply: "ai-approval-apply",
  approvalReject: "ai-approval-reject",
} as const;

export const AI_EMPLOYEE_COPY_LABELS = {
  user: "Пользователь",
  assistant: "ИИ-сотрудник",
} as const;

export const AI_EMPLOYEE_THEME = {
  color: {
    background: "var(--background)",
    foreground: "var(--foreground)",
    card: "var(--card)",
    border: "var(--border)",
    primary: "var(--primary)",
    success: "var(--success)",
    warning: "var(--warning)",
    danger: "var(--danger)",
    info: "var(--info)",
  },
  radius: {
    control: "8px",
    panel: "12px",
  },
} as const;

export const TASK_COLLECTION_PATHS = [
  "/tasks/boards",
  "/tasks/all",
  "/tasks/calendar",
  "/tasks/processes",
  "/tasks/time",
  "/tasks/timeline",
  "/tasks/reports",
  "/tasks/workload",
  "/tasks/employees",
] as const;

export const PAGE_KIND_BY_ROUTE: Array<{
  pageKind: AIPageContext["pageKind"];
  match: (pathname: string) => boolean;
}> = [
  { pageKind: "site_chat", match: (pathname) => pathname === "/site-chat" || pathname.startsWith("/site-chat/") },
  { pageKind: "system_status", match: (pathname) => pathname === "/status" || pathname.startsWith("/status/") },
  {
    pageKind: "workflow_editor",
    match: (pathname) => pathname.includes("/project-workflows") || pathname.includes("/workflows/"),
  },
  { pageKind: "workflow_run", match: (pathname) => pathname.includes("/runs/") },
  { pageKind: "task_board", match: (pathname) => pathname.startsWith("/tasks/boards") },
  { pageKind: "integrations", match: (pathname) => pathname.includes("/project-integrations") || /\/projects\/[^/]+\/integrations/.test(pathname) },
  { pageKind: "client", match: (pathname) => pathname.includes("/clients/") },
  { pageKind: "clients", match: (pathname) => pathname.includes("/clients") },
  { pageKind: "payments", match: (pathname) => pathname.includes("/payments") },
  { pageKind: "settings", match: (pathname) => pathname.includes("/settings") },
  { pageKind: "project", match: (pathname) => pathname.includes("/projects/") },
  { pageKind: "dashboard", match: (pathname) => pathname.includes("/dashboard") },
];

export const AI_QUICK_ACTIONS = {
  common: ["Где я сейчас?"],
  taskDefault: ["Создай задачу проверить оплаты завтра", "Покажи открытые задачи", "Добавь статус «На проверке» в процесс задач"],
  clientDefault: ["Помоги добавить клиента", "Найди клиента", "Покажи ближайшие оплаты"],
  workflowDefault: ["Объясни этот сценарий", "Измени этот сценарий, чтобы после оплаты отправлялся Telegram и email", "Проверь сценарий"],
} as const;
