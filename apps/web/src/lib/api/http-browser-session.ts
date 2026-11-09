import { AUTH_API_ROUTES, isAuthBootstrapPathname } from "./auth-routes";
import { apiBase } from "./base";
import { transportFetch } from "./http-transport";

type RefreshResponse = { access_token?: string };
type CSRFResponse = { csrf_token?: string };

const CSRF_COOKIE = "rp_csrf";
const CSRF_HEADER = "X-CSRF-Token";
const PUBLIC_API_PATH_PREFIX = ["/api", "v1", "public", ""].join("/");
const HOOKS_PATH_PREFIX = "/hooks/";
const REFRESH_TIMEOUT_MS = 8_000;

let refreshInFlight: Promise<string | undefined> | null = null;
let csrfInFlight: Promise<boolean> | null = null;

export async function prepareBrowserRequest(path: string, init: RequestInit) {
  let next = withBrowserDefaults(init);
  if (shouldBootstrapCSRFBeforeSend(path, next)) {
    await bootstrapBrowserCSRF();
    next = withBrowserDefaults(init);
  }
  return next;
}

export async function refreshBrowserSession() {
  refreshInFlight ??= transportFetch(
    apiURL(AUTH_API_ROUTES.refresh),
    withBrowserDefaults({ method: "POST" }),
    { retries: 0, timeoutMs: REFRESH_TIMEOUT_MS },
  )
    .then(async (response): Promise<string | undefined> => {
      if (!response.ok) return undefined;
      const data = await response.json().catch(() => ({})) as RefreshResponse;
      return typeof data.access_token === "string" && data.access_token ? data.access_token : undefined;
    })
    .catch(() => undefined)
    .finally(() => {
      refreshInFlight = null;
    });
  return refreshInFlight;
}

export async function bootstrapBrowserCSRF() {
  csrfInFlight ??= requestBrowserCSRF(AUTH_API_ROUTES.csrf)
    .then(async (ok) => ok || requestBrowserCSRF(AUTH_API_ROUTES.refreshCSRF))
    .catch(() => false)
    .finally(() => {
      csrfInFlight = null;
    });
  return csrfInFlight;
}

async function requestBrowserCSRF(path: string) {
  const response = await transportFetch(
    apiURL(path),
    withBrowserDefaults({ method: "GET" }),
    { retries: 0, timeoutMs: REFRESH_TIMEOUT_MS },
  );
  if (!response.ok) return false;
  const data = await response.json().catch(() => ({})) as CSRFResponse;
  return Boolean(data.csrf_token || browserCSRFToken());
}

export function isAuthBootstrapPath(path: string) {
  return isAuthBootstrapPathname(apiPathname(path));
}

export function apiURL(path: string) {
  return /^https?:\/\//i.test(path) ? path : `${apiBase}${path}`;
}

export function shouldSkipBrowserRefresh(headers: HeadersInit | undefined) {
  const authorization = new Headers(headers).get("Authorization") ?? "";
  return authorization.replace(/^Bearer\s+/i, "").trim().startsWith("rp_live_");
}

export async function shouldRetryAfterCSRF(path: string, init: RequestInit, response: Response) {
  if (isAuthBootstrapPath(path) || isPublicAPIPath(path) || shouldSkipBrowserRefresh(init.headers)) return false;
  if (!requiresCSRF(init.method) || hasAuthorizationHeader(init.headers)) return false;
  return isCSRFRequiredResponse(response);
}

export async function isCSRFRequiredResponse(response: Response) {
  if (response.status !== 403 || !(response.headers.get("Content-Type") ?? "").includes("json")) return false;
  const data = await response.clone().json().catch(() => ({}));
  if (!data || typeof data !== "object") return false;
  const problem = data as { code?: unknown; title?: unknown };
  return problem.code === "csrf_required" || problem.title === "csrf required";
}

export function withRefreshedAuthorization(init: RequestInit, accessToken: string): RequestInit {
  const headers = new Headers(init.headers);
  if (headers.has("Authorization")) headers.set("Authorization", `Bearer ${accessToken}`);
  return { ...init, headers };
}

export function withBrowserDefaults(init: RequestInit): RequestInit {
  const headers = new Headers(init.headers);
  addBrowserCSRF(headers, init.method);
  return { ...init, credentials: "include", headers };
}

function shouldBootstrapCSRFBeforeSend(path: string, init: RequestInit) {
  if (isAuthBootstrapPath(path) || isPublicAPIPath(path) || shouldSkipBrowserRefresh(init.headers)) return false;
  if (!requiresCSRF(init.method) || hasAuthorizationHeader(init.headers)) return false;
  return !new Headers(init.headers).has(CSRF_HEADER);
}

function addBrowserCSRF(headers: Headers, method: string | undefined) {
  if (!requiresCSRF(method) || headers.has(CSRF_HEADER)) return;
  const token = browserCSRFToken();
  if (token) headers.set(CSRF_HEADER, token);
}

function apiPathname(path: string) {
  if (!/^https?:\/\//i.test(path)) return path.split("?")[0] || "/";
  try { return new URL(path).pathname; } catch { return path; }
}

function isPublicAPIPath(path: string) {
  const pathname = apiPathname(path);
  return pathname.startsWith(PUBLIC_API_PATH_PREFIX) || pathname.startsWith(HOOKS_PATH_PREFIX);
}

function hasAuthorizationHeader(headers: HeadersInit | undefined) {
  return Boolean(new Headers(headers).get("Authorization")?.trim());
}

function browserCSRFToken() {
  if (typeof document === "undefined") return "";
  const match = document.cookie.split(";").map((part) => part.trim()).find((part) => part.startsWith(`${CSRF_COOKIE}=`));
  return match ? decodeURIComponent(match.slice(CSRF_COOKIE.length + 1)) : "";
}

function requiresCSRF(method: string | undefined) {
  return !["GET", "HEAD", "OPTIONS"].includes((method ?? "GET").toUpperCase());
}
