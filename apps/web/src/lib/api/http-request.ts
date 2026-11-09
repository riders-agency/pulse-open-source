import { BROWSER_SESSION_TOKEN, notifyUnauthorized } from "./http-errors";
import {
  apiURL,
  bootstrapBrowserCSRF,
  isAuthBootstrapPath,
  isCSRFRequiredResponse,
  prepareBrowserRequest,
  refreshBrowserSession,
  shouldRetryAfterCSRF,
  shouldSkipBrowserRefresh,
  withBrowserDefaults,
  withRefreshedAuthorization,
} from "./http-browser-session";
import { retryPolicy, transportFetch, type TransportPolicy } from "./http-transport";

type APIBody = BodyInit | null | undefined;

export type APIFetchOptions = Omit<RequestInit, "body" | "headers" | "method"> & {
  method?: string;
  token?: string;
  headers?: HeadersInit;
  body?: APIBody;
  json?: unknown;
  form?: FormData;
  retries?: number;
  timeoutMs?: number;
};

export function authHeaders(token: string, json = true) {
  const headers: Record<string, string> = json ? { "Content-Type": "application/json" } : {};
  applyAuthHeader(headers, token);
  return headers;
}

export async function apiFetch(path: string, options: APIFetchOptions = {}) {
  const { retries, timeoutMs, ...requestOptions } = options;
  return sendWithAuth(path, buildRequestInit(requestOptions), { retries, timeoutMs });
}

export async function fetchWithAuth(path: string, init: RequestInit = {}) {
  return sendWithAuth(path, init);
}

function buildRequestInit(options: APIFetchOptions): RequestInit {
  const { token = "", headers: sourceHeaders, json, form, body, method, ...rest } = options;
  const headers = new Headers(sourceHeaders);
  applyAuthHeader(headers, token);

  let requestBody: APIBody = body;
  if (form !== undefined) {
    requestBody = form;
    headers.delete("Content-Type");
  } else if (json !== undefined) {
    requestBody = JSON.stringify(json);
    if (!headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  }

  return {
    ...rest,
    method: normalizeMethod(method, requestBody),
    headers,
    body: requestBody,
  };
}

async function sendWithAuth(path: string, init: RequestInit, policy: TransportPolicy = {}) {
  const firstInit = await prepareBrowserRequest(path, init);
  const res = await transportFetch(apiURL(path), firstInit, retryPolicy(firstInit, policy));
  if (await shouldRetryAfterCSRF(path, firstInit, res)) {
    const csrfReady = await bootstrapBrowserCSRF();
    if (csrfReady) {
      const retryInit = withBrowserDefaults(init);
      const retry = await transportFetch(apiURL(path), retryInit, retryPolicy(retryInit, policy));
      if (!(await isCSRFRequiredResponse(retry))) return retry;
    }
    notifyUnauthorized();
    return res;
  }
  if (res.status !== 401 || isAuthBootstrapPath(path) || shouldSkipBrowserRefresh(firstInit.headers)) {
    return res;
  }
  const accessToken = await refreshBrowserSession();
  if (!accessToken) {
    notifyUnauthorized();
    return res;
  }
  const retryInit = withBrowserDefaults(withRefreshedAuthorization(init, accessToken));
  return transportFetch(apiURL(path), retryInit, retryPolicy(retryInit, policy));
}

function normalizeMethod(method: string | undefined, body: APIBody) {
  if (method) return method.toUpperCase();
  return body === undefined || body === null ? "GET" : "POST";
}

function applyAuthHeader(headers: Headers | Record<string, string>, token: string) {
  if (!token || token === BROWSER_SESSION_TOKEN) return;
  if (headers instanceof Headers) {
    if (!headers.has("Authorization")) headers.set("Authorization", `Bearer ${token}`);
    return;
  }
  if (!headers.Authorization) headers.Authorization = `Bearer ${token}`;
}
