import type { ApiErrorContext, ApiProblem } from "./http-errors";

export function normalizeProblem(res: Response, data: unknown, context: ApiErrorContext): ApiProblem {
  const raw = isRecord(data) ? data : {};
  const status = typeof raw.status === "number" ? raw.status : res.status;
  const moduleName = typeof raw.module === "string" ? raw.module : moduleFromPath(context.path);
  const code = typeof raw.code === "string" ? raw.code : codeForStatus(status);
  const mapped = mappedProblem(moduleName, code, status);
  const detail = typeof raw.detail === "string" && raw.detail.trim() ? raw.detail : mapped.detail;
  return {
    type: typeof raw.type === "string" ? raw.type : "about:blank",
    module: moduleName,
    code,
    title: typeof raw.title === "string" && raw.title.trim() ? raw.title : mapped.title,
    status,
    detail,
    action: typeof raw.action === "string" && raw.action.trim() ? raw.action : mapped.action,
    validate: isRecord(raw.validate) ? raw.validate : undefined,
    request_id:
      typeof raw.request_id === "string" && raw.request_id.trim()
        ? raw.request_id
        : res.headers.get("X-Request-Id") || undefined,
    extra: {
      ...(isRecord(raw.extra) ? raw.extra : {}),
      path: context.path,
      method: context.method,
    },
  };
}

function mappedProblem(moduleName: string, code: string, status: number) {
  const key = `${moduleName}.${code}`;
  const map: Record<string, { title: string; detail?: string; action: string }> = {
    "routing.route_not_found": {
      title: "API маршрут не найден",
      detail: "Фронт вызвал endpoint, которого нет на сервере.",
      action: "Проверьте контракт endpoint и активный рабочий контекст.",
    },
    "routing.method_not_allowed": {
      title: "Метод API не поддерживается",
      detail: "Endpoint найден, но HTTP-метод для него не подключен.",
      action: "Проверьте действие формы или API-клиент.",
    },
    "clients.dadata_not_configured": {
      title: "DaData не подключена",
      detail: "Для автозаполнения реквизитов нужен ключ DaData.",
      action: "Подключите DaData в компании или выбранном проекте и повторите поиск.",
    },
    "clients.identifiers": {
      title: "Некорректные реквизиты клиента",
      action: "Проверьте ИНН, КПП, ОГРН или ОГРНИП.",
    },
    "clients.payment_closed": {
      title: "Оплата уже закрыта",
      action: "Обновите карточку оплаты и проверьте текущий статус.",
    },
  };
  return map[key] ?? { title: titleForStatus(status), action: actionForStatus(status) };
}

export function titleForStatus(status: number) {
  if (status === 400) return "Запрос не прошел проверку";
  if (status === 401) return "Нужна авторизация";
  if (status === 403) return "Нет доступа";
  if (status === 404) return "Не найдено";
  if (status === 405) return "Метод не поддерживается";
  if (status === 409) return "Конфликт состояния";
  if (status >= 500) return "Ошибка сервера";
  return `HTTP ${status}`;
}

export function actionForStatus(status: number) {
  if (status === 400) return "Проверьте данные формы и повторите действие.";
  if (status === 401) return "Войдите снова и повторите действие.";
  if (status === 403) return "Проверьте права доступа к компании, проекту или ресурсу.";
  if (status === 404) return "Проверьте выбранный раздел, компанию и проект.";
  if (status === 405) return "Используйте доступное действие для этого endpoint.";
  if (status === 409) return "Обновите данные и повторите действие.";
  if (status >= 500) return "Повторите позже или передайте request_id разработчику.";
  return "Проверьте запрос и повторите действие.";
}

function codeForStatus(status: number) {
  if (status === 400) return "bad_request";
  if (status === 401) return "unauthorized";
  if (status === 403) return "forbidden";
  if (status === 404) return "not_found";
  if (status === 405) return "method_not_allowed";
  if (status === 409) return "conflict";
  if (status >= 500) return "server_error";
  return "http_error";
}

function moduleFromPath(path?: string) {
  if (!path) return "api";
  if (path.includes("/clients") || path.includes("/client-services") || path.includes("/payment-actions")) return "clients";
  if (path.includes("/companies")) return "company";
  if (path.includes("/projects")) return "project";
  if (path.includes("/workflows")) return "workflow";
  if (path.includes("/runs")) return "runs";
  if (path.includes("/webhook")) return "webhooks";
  if (path.includes("/action-dependencies")) return "integrations";
  if (path.includes("/auth")) return "auth";
  return "api";
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
