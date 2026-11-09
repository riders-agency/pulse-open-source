import { useCallback, useEffect, useState } from "react";

import {
  AUTH_UNAUTHORIZED_EVENT,
  BROWSER_SESSION_TOKEN,
  isUnauthorizedError,
  loginWithPassword,
  registerAccount,
} from "../api/authApi";
import { usePulseStore } from "../../../store/pulse-store";
import { useAuthGateEnvironment } from "./useAuthGateEnvironment";

export type AuthMode = "login" | "register";
export type AuthSubmitInput = {
  email: string;
  password: string;
  displayName?: string;
};

export function useAuthGateState() {
  const hydrated = usePulseStore((s) => s.hydrated);
  const token = usePulseStore((s) => s.token);
  const setToken = usePulseStore((s) => s.setToken);
  const setHydrated = usePulseStore((s) => s.setHydrated);
  const clearWorkspace = usePulseStore((s) => s.clearWorkspace);
  const theme = usePulseStore((s) => s.theme);
  const sidebarCollapsed = usePulseStore((s) => s.sidebarCollapsed);
  const [mode, setMode] = useState<AuthMode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const returnToLogin = useCallback(() => {
    setToken("");
    clearWorkspace();
  }, [clearWorkspace, setToken]);

  useAuthGateEnvironment({ hydrated, setHydrated, theme });

  useEffect(() => {
    const resetUnauthorized = () => {
      setToken("");
      clearWorkspace();
    };
    window.addEventListener(AUTH_UNAUTHORIZED_EVENT, resetUnauthorized);
    return () => window.removeEventListener(AUTH_UNAUTHORIZED_EVENT, resetUnauthorized);
  }, [clearWorkspace, setToken]);

  useEffect(() => {
    if (!hydrated) return;
    if (token && token !== BROWSER_SESSION_TOKEN) {
      returnToLogin();
    }
  }, [hydrated, returnToLogin, token]);

  const login = async (input?: AuthSubmitInput) => {
    const nextEmail = (input?.email ?? email).trim();
    const nextPassword = input?.password ?? password;
    if (!nextEmail || !nextPassword) {
      setError("Введите email и пароль.");
      return;
    }
    setError("");
    setEmail(nextEmail);
    setPassword(nextPassword);
    setBusy(true);
    try {
      const out = await loginWithPassword({ email: nextEmail, password: nextPassword });
      if (out.access_token) {
        setToken(BROWSER_SESSION_TOKEN);
      } else setError("API не вернул токен доступа");
    } catch (err) {
      setError(isUnauthorizedError(err) ? "Неверный email или пароль." : err instanceof Error ? err.message : "Не удалось проверить логин и пароль");
    } finally {
      setBusy(false);
    }
  };

  const register = async (input?: AuthSubmitInput) => {
    const nextEmail = (input?.email ?? email).trim();
    const nextPassword = input?.password ?? password;
    const nextDisplayName = (input?.displayName ?? displayName).trim();
    if (!nextEmail || !nextPassword) {
      setError("Введите email и пароль.");
      return;
    }
    setError("");
    setEmail(nextEmail);
    setPassword(nextPassword);
    setDisplayName(nextDisplayName);
    setBusy(true);
    try {
      const out = await registerAccount({
        email: nextEmail,
        password: nextPassword,
        display_name: nextDisplayName || nextEmail.split("@")[0] || "User",
      });
      if (out.access_token) {
        setToken(BROWSER_SESSION_TOKEN);
      } else setError("API не вернул токен доступа");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось зарегистрироваться");
    } finally {
      setBusy(false);
    }
  };

  const submit = (input?: AuthSubmitInput) => {
    if (mode === "login") void login(input);
    else void register(input);
  };

  return {
    busy,
    displayName,
    email,
    error,
    hydrated,
    isLoginRequired: !hydrated || token !== BROWSER_SESSION_TOKEN,
    mode,
    password,
    returnToLogin,
    setDisplayName,
    setEmail,
    setMode,
    setPassword,
    sidebarCollapsed,
    submit,
  };
}
