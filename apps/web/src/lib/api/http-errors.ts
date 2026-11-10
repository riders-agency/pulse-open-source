import { actionForStatus, titleForStatus } from "./http-problem";

export const AUTH_UNAUTHORIZED_EVENT = "riders-pulse:unauthorized";
export const API_ERROR_EVENT = "riders-pulse:api-error";
export const BROWSER_SESSION_TOKEN = "__riders_pulse_browser_session__";

export type ApiProblem = {
  type?: string;
  module?: string;
  code?: string;
  title?: string;
  status?: number;
  detail?: string;
  action?: string;
  validate?: Record<string, unknown>;
  request_id?: string;
  extra?: Record<string, unknown>;
};

export type ApiErrorContext = {
  inlineError?: boolean;
  method?: string;
  path?: string;
};

export type ApiErrorPayload = Required<Pick<ApiProblem, "title" | "status">> &
  Pick<ApiProblem, "type" | "module" | "code" | "detail" | "action" | "validate" | "request_id" | "extra"> &
  ApiErrorContext;

export class ApiError extends Error {
  status: number;
  module?: string;
  code?: string;
  title?: string;
  detail?: string;
  action?: string;
  validate?: Record<string, unknown>;
  requestId?: string;
  path?: string;
  method?: string;
  extra?: Record<string, unknown>;

  constructor(status: number, title: string, detail?: string, problem: ApiProblem = {}, context: ApiErrorContext = {}) {
    super(detail || title || `HTTP ${status}`);
    this.name = "ApiError";
    this.status = status;
    this.module = problem.module;
    this.code = problem.code;
    this.title = title;
    this.detail = detail;
    this.action = problem.action;
    this.validate = problem.validate;
    this.requestId = problem.request_id;
    this.path = context.path;
    this.method = context.method;
    this.extra = problem.extra;
  }

  toPayload(): ApiErrorPayload {
    return {
      status: this.status,
      module: this.module,
      code: this.code,
      title: this.title || titleForStatus(this.status),
      detail: this.detail,
      action: this.action,
      validate: this.validate,
      request_id: this.requestId,
      path: this.path,
      method: this.method,
      extra: this.extra,
    };
  }
}

export function isUnauthorizedError(err: unknown) {
  return err instanceof ApiError && err.status === 401;
}

export function notifyUnauthorized() {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new CustomEvent(AUTH_UNAUTHORIZED_EVENT));
  }
}

export function notifyApiError(error: ApiError) {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new CustomEvent<ApiErrorPayload>(API_ERROR_EVENT, { detail: error.toPayload() }));
  }
}

export function defaultProblem(status: number) {
  return { title: titleForStatus(status), action: actionForStatus(status) };
}
