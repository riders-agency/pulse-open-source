import type { AIPageContext } from "../../lib/api";
import {
  AI_EMPLOYEE_LEGACY_STORAGE_PREFIX,
  AI_EMPLOYEE_STORAGE_PREFIX,
  AI_QUICK_ACTIONS,
  PAGE_KIND_BY_ROUTE,
  TASK_COLLECTION_PATHS,
} from "./aiEmployeeConstants";

export const taskCollectionPaths = [...TASK_COLLECTION_PATHS];

export function isTaskDetailPath(pathname: string) {
  return /^\/tasks\/[^/]+/.test(pathname) && !taskCollectionPaths.some((path) => pathname.startsWith(path));
}

export function pageKind(pathname: string): AIPageContext["pageKind"] {
  if (isTaskDetailPath(pathname)) return "task";
  for (const route of PAGE_KIND_BY_ROUTE) {
    if (route.match(pathname)) return route.pageKind;
  }
  if (pathname.startsWith("/tasks") || pathname === "/work" || pathname.startsWith("/work/")) return "tasks";
  return "unknown";
}

export function matchID(pathname: string, pattern: RegExp) {
  const match = pathname.match(pattern);
  return match?.[1] ?? undefined;
}

export function taskIDFromPathname(pathname: string) {
  if (isTaskDetailPath(pathname)) return matchID(pathname, /^\/tasks\/([^/]+)/);
  return matchID(pathname, /\/work\/tasks\/([^/]+)/);
}

export function buildPageContext(pathname: string, activeProjectId?: string | null, query = ""): AIPageContext {
  const queryParams = new URLSearchParams(query);
  return {
    route: `${pathname}${query ? `?${query}` : ""}`,
    pathname,
    pageKind: pageKind(pathname),
    projectId: activeProjectId ?? undefined,
    workflowId: matchID(pathname, /\/workflows\/([^/]+)/),
    workflowRunId: matchID(pathname, /\/runs\/([^/]+)/),
    taskId: taskIDFromPathname(pathname),
    clientId: matchID(pathname, /\/clients\/([^/]+)/),
    selectedActionKey: queryParams.get("action") ?? (typeof window !== "undefined" ? new URLSearchParams(window.location.search).get("action") ?? undefined : undefined),
    visiblePanel: typeof window !== "undefined" ? document.title : undefined,
    timestamp: new Date().toISOString(),
  };
}

export function projectSessionStorageKey(projectId: string) {
  return `${AI_EMPLOYEE_STORAGE_PREFIX}${projectId}`;
}

export function legacyProjectSessionStorageKey(projectId: string) {
  return `${AI_EMPLOYEE_LEGACY_STORAGE_PREFIX}${projectId}`;
}

export function screenScope(context: AIPageContext) {
  switch (context.pageKind) {
    case "task_board":
    case "tasks":
      return "task_board";
    case "task":
      return "task_detail";
    case "workflow_editor":
      return "workflow_editor";
    case "workflow_run":
      return "workflow_run";
    case "integrations":
      return "integrations";
    case "site_chat":
      return "site_chat";
    case "system_status":
      return "system_status";
    case "clients":
    case "client":
      return "client_detail";
    default:
      return "project_dashboard";
  }
}

export function quickActionsFor(context: AIPageContext) {
  const common = [...AI_QUICK_ACTIONS.common];
  if (context.pathname.startsWith("/tasks/time")) {
    return [...common, "Покажи, по каким задачам сегодня шло время", "Создай задачу проверить оплаты завтра"];
  }
  if (context.pathname.startsWith("/tasks/timeline")) {
    return [...common, "Что горит по срокам на этой неделе?", "Создай задачу проверить оплаты завтра"];
  }
  if (context.pathname.startsWith("/tasks/processes")) {
    return [...common, "Добавь статус «На проверке» в процесс задач", "Объясни текущий процесс задач"];
  }
  switch (context.pageKind) {
    case "task_board":
    case "tasks":
      return [...common, ...AI_QUICK_ACTIONS.taskDefault];
    case "task":
      return [...common, "Объясни эту задачу", "Создай похожую задачу на завтра"];
    case "clients":
      return [...common, ...AI_QUICK_ACTIONS.clientDefault];
    case "client":
      return [...common, "Когда оплата у этого клиента?", "Создай задачу проверить оплату этого клиента завтра"];
    case "workflow_editor":
      return [...common, ...AI_QUICK_ACTIONS.workflowDefault];
    case "workflow_run":
      return [...common, "Объясни, почему этот запуск завершился ошибкой", "Предложи исправление"];
    case "integrations":
      return [...common, "Проверь подключение ИИ"];
    case "payments":
      return [...common, "Покажи ближайшие оплаты", "Создай задачу проверить просроченные оплаты"];
    case "site_chat":
      return [...common, "Покажи активные обращения из чата сайта", "Создай задачу по текущему обращению", "Подготовь ответ клиенту"];
    default:
      return [...common, "Объясни текущий проект", "Создай задачу проверить оплаты завтра", "Помоги добавить клиента"];
  }
}
