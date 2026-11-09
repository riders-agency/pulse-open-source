"use client";

import { useCallback, useEffect, useState } from "react";

import { usePulseStore } from "../../../store/pulse-store";
import {
  createAccountAPIKey,
  endAccountSession,
  getAccountSettings,
  logoutAccount,
  revokeAccountAPIKey,
  updateAccountProfile,
  type APIKeyItem,
  type MeOut,
  type SessionItem,
} from "../api/accountApi";

type AccountScopeKind = "company" | "project" | "workflow" | "system";
type AccountPermission = "read" | "write" | "admin";

function accountScopeKind(value: string): AccountScopeKind {
  return value === "company" || value === "project" || value === "workflow" || value === "system" ? value : "project";
}

function accountPermission(value: string): AccountPermission {
  return value === "write" || value === "admin" ? value : "read";
}

export function useAccountSettings() {
  const token = usePulseStore((s) => s.token);
  const setToken = usePulseStore((s) => s.setToken);
  const clearWorkspace = usePulseStore((s) => s.clearWorkspace);
  const activeCompanyId = usePulseStore((s) => s.activeCompanyId);
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const [me, setMe] = useState<MeOut | null>(null);
  const [sessions, setSessions] = useState<SessionItem[]>([]);
  const [apiKeys, setAPIKeys] = useState<APIKeyItem[]>([]);
  const [newAPIKey, setNewAPIKey] = useState("");
  const [keyName, setKeyName] = useState("Automation key");
  const [scopeKind, setScopeKind] = useState<AccountScopeKind>("project");
  const [scopeID, setScopeID] = useState(activeProjectId || activeCompanyId || "");
  const [permission, setPermission] = useState<AccountPermission>("read");
  const [displayName, setDisplayName] = useState("");
  const [firstName, setFirstName] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    if (!token) return;
    setError("");
    try {
      const { me: m, sessions: sessionRows, apiKeys: keyRows } = await getAccountSettings(token);
      setMe(m);
      setDisplayName(m.display_name ?? "");
      setFirstName(m.first_name ?? "");
      setSessions(sessionRows);
      setAPIKeys(keyRows);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка загрузки");
    }
  }, [token]);

  useEffect(() => {
    void load();
  }, [load]);

  const saveProfile = async () => {
    if (!token) return;
    setBusy(true);
    setError("");
    try {
      await updateAccountProfile(token, {
        display_name: displayName,
        first_name: firstName,
      });
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Не удалось сохранить");
    } finally {
      setBusy(false);
    }
  };

  const endSession = async (sessionId: string, current: boolean) => {
    if (!token || current) return;
    try {
      await endAccountSession(token, sessionId);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Не удалось завершить сессию");
    }
  };

  const logoutEverywhereCurrent = async () => {
    if (!token) return;
    try {
      await logoutAccount(token);
    } catch {
      /* ignore network error */
    }
    setToken("");
    clearWorkspace();
  };

  const createAPIKey = async () => {
    if (!token) return;
    setBusy(true);
    setError("");
    setNewAPIKey("");
    try {
      const out = await createAccountAPIKey(token, {
        name: keyName,
        grants: [
          {
            scope_kind: scopeKind,
            scope_id: scopeKind === "system" ? "" : scopeID,
            permissions: [permission],
          },
        ],
      });
      setNewAPIKey(out.plaintext_api_key ?? "");
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Не удалось выпустить API key");
    } finally {
      setBusy(false);
    }
  };

  const revokeAPIKey = async (id: string) => {
    if (!token) return;
    try {
      await revokeAccountAPIKey(token, id);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Не удалось отозвать API key");
    }
  };

  const updateScopeKind = (value: string) => {
    const nextScopeKind = accountScopeKind(value);
    setScopeKind(nextScopeKind);
    if (nextScopeKind === "company") setScopeID(activeCompanyId);
    if (nextScopeKind === "project") setScopeID(activeProjectId);
    if (nextScopeKind === "system") setScopeID("");
  };

  return {
    apiKeys,
    busy,
    createAPIKey,
    displayName,
    endSession,
    error,
    firstName,
    keyName,
    logoutEverywhereCurrent,
    me,
    newAPIKey,
    permission,
    revokeAPIKey,
    saveProfile,
    scopeID,
    scopeKind,
    sessions,
    setDisplayName,
    setFirstName,
    setKeyName,
    setScopeID,
    updatePermission: (value: string) => setPermission(accountPermission(value)),
    updateScopeKind,
  };
}
