import { apiDelete, apiGet, apiPatch, apiPost, apiPostVoid } from "../../../lib/api";

export type MeOut = {
  user_id: string;
  email: string;
  display_name: string;
  first_name: string;
};

export type SessionItem = {
  session_id: string;
  created_at: string;
  expires_at: string;
  user_agent?: string;
  ip?: string;
  current: boolean;
};

export type APIKeyGrant = {
  scope_kind: string;
  scope_id?: string;
  permissions: string[];
};

export type APIKeyItem = {
  api_key_id: string;
  name: string;
  key_prefix: string;
  status: string;
  created_at: string;
  last_used_at?: string;
  expires_at?: string;
  grants: APIKeyGrant[];
};

export async function getAccountSettings(token: string) {
  const [me, sessions, apiKeys] = await Promise.all([
    apiGet<MeOut>("/api/v1/auth/me", token),
    apiGet<{ sessions: SessionItem[] }>("/api/v1/auth/sessions", token),
    apiGet<{ api_keys: APIKeyItem[] }>("/api/v1/auth/api-keys", token),
  ]);
  return {
    me,
    sessions: Array.isArray(sessions.sessions) ? sessions.sessions : [],
    apiKeys: Array.isArray(apiKeys.api_keys) ? apiKeys.api_keys : [],
  };
}

export function updateAccountProfile(token: string, input: { display_name: string; first_name: string }) {
  return apiPatch<void>("/api/v1/auth/profile", token, input);
}

export function endAccountSession(token: string, sessionId: string) {
  return apiDelete(`/api/v1/auth/sessions/${sessionId}`, token);
}

export function logoutAccount(token: string) {
  return apiPostVoid("/api/v1/auth/logout", token);
}

export function createAccountAPIKey(token: string, input: { name: string; grants: APIKeyGrant[] }) {
  return apiPost<{ plaintext_api_key?: string }>("/api/v1/auth/api-keys", token, input);
}

export function revokeAccountAPIKey(token: string, apiKeyId: string) {
  return apiDelete(`/api/v1/auth/api-keys/${apiKeyId}`, token);
}
