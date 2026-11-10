import {
  ApiError,
  notifyApiError,
  notifyUnauthorized,
  type ApiErrorContext,
} from "./http-errors";
import { AUTH_API_ROUTES } from "./auth-routes";
import { normalizeProblem, titleForStatus } from "./http-problem";
import { apiFetch } from "./http-request";

export {
  API_ERROR_EVENT,
  AUTH_UNAUTHORIZED_EVENT,
  ApiError,
  BROWSER_SESSION_TOKEN,
  isUnauthorizedError,
  type ApiErrorContext,
  type ApiErrorPayload,
  type ApiProblem,
} from "./http-errors";
export { apiFetch, authHeaders, fetchWithAuth, type APIFetchOptions } from "./http-request";

export async function parseResponse<T>(res: Response, context: ApiErrorContext = {}): Promise<T> {
  const data = await readResponseBody(res);
  if (!res.ok) {
    throwApiError(res, data, context);
  }
  return data as T;
}

/** PATCH/DELETE with optional empty body (204); does not consume body on success (so caller may read JSON). */
export async function parseResponseVoid(res: Response, context: ApiErrorContext = {}): Promise<void> {
  if (!res.ok) {
    const data = await readResponseBody(res);
    throwApiError(res, data, context);
  }
}

export async function apiGet<T>(path: string, token: string) {
  return parseResponse<T>(await apiFetch(path, { method: "GET", token }), { method: "GET", path });
}

export async function apiPost<T>(path: string, token: string, body?: unknown, context: ApiErrorContext = {}) {
  return parseResponse<T>(await apiFetch(path, { method: "POST", token, json: body ?? {} }), { ...context, method: "POST", path });
}

export async function apiPostVoid(path: string, token: string, body?: unknown) {
  const res = await apiFetch(path, { method: "POST", token, json: body ?? {} });
  await parseResponseVoid(res, { method: "POST", path });
}

export async function apiUpload<T>(path: string, token: string, form: FormData) {
  return parseResponse<T>(await apiFetch(path, { method: "POST", token, form }), { method: "POST", path });
}

export async function apiPut<T>(path: string, token: string, body?: unknown) {
  return parseResponse<T>(await apiFetch(path, { method: "PUT", token, json: body ?? {} }), { method: "PUT", path });
}

export async function apiPatch<T>(path: string, token: string, body?: unknown) {
  const res = await apiFetch(path, { method: "PATCH", token, json: body ?? {} });
  await parseResponseVoid(res, { method: "PATCH", path });
  if (res.status === 204) return undefined as T;
  const ct = res.headers.get("Content-Type") ?? "";
  if (!ct.includes("json")) return undefined as T;
  return (await res.json()) as T;
}

export async function apiDelete(path: string, token: string) {
  await parseResponseVoid(
    await apiFetch(path, { method: "DELETE", token }),
    { method: "DELETE", path },
  );
}

async function readResponseBody(res: Response): Promise<unknown> {
  const contentType = res.headers.get("Content-Type") ?? "";
  if (contentType.includes("json")) {
    return res.json().catch(() => ({}));
  }
  const text = await res.text().catch(() => "");
  if (!text) return {};
  return { detail: text };
}

function throwApiError(res: Response, data: unknown, context: ApiErrorContext): never {
  const problem = normalizeProblem(res, data, context);
  if (res.status === 401 || isCSRFRequiredProblem(res.status, problem)) notifyUnauthorized();
  const error = new ApiError(res.status, problem.title || titleForStatus(res.status), problem.detail, problem, context);
  if (!context.inlineError && !shouldSuppressApiError(context.path, res.status, problem.code)) notifyApiError(error);
  throw error;
}

function shouldSuppressApiError(path: string | undefined, status: number, code?: string) {
  const pathname = (path ?? "").split("?")[0];
  if (status === 403 && code === "csrf_required") return true;
  if (status === 401 && (pathname === AUTH_API_ROUTES.me || pathname === AUTH_API_ROUTES.refresh)) return true;
  if (status === 404 && /^\/api\/v1\/clients\/[^/]+\/payment-actions$/.test(pathname)) return true;
  if (status === 404 && /^\/api\/v1\/clients\/[^/]+\/payments\/manual$/.test(pathname)) return true;
  if (status === 404 && /^\/api\/v1\/projects\/[^/]+\/client-service-types(\/[^/]+)?$/.test(pathname)) return true;
  return false;
}

function isCSRFRequiredProblem(status: number, problem: { code?: string; title?: string }) {
  return status === 403 && (problem.code === "csrf_required" || problem.title === "csrf required");
}
