import {
  AUTH_UNAUTHORIZED_EVENT,
  AUTH_API_ROUTES,
  BROWSER_SESSION_TOKEN,
  apiGet,
  apiPost,
  apiPostVoid,
  isUnauthorizedError,
} from "../../../lib/api";

export { AUTH_UNAUTHORIZED_EVENT, BROWSER_SESSION_TOKEN, isUnauthorizedError };

export function verifyBrowserSession() {
  return apiGet<{ user_id: string }>(AUTH_API_ROUTES.me, BROWSER_SESSION_TOKEN);
}

export function loginWithPassword(input: { email: string; password: string }) {
  return apiPost<{ access_token?: string }>(AUTH_API_ROUTES.login, "", input);
}

export function registerAccount(input: { email: string; password: string; display_name: string }) {
  return apiPost<{ access_token?: string }>(AUTH_API_ROUTES.register, "", input);
}

export function logoutSession(token: string) {
  return apiPostVoid(AUTH_API_ROUTES.logout, token);
}
