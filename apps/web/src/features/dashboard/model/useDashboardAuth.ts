"use client";

import { useCallback, useState } from "react";

import { usePulseStore } from "../../../store/pulse-store";
import { loginDashboard } from "../api/dashboardApi";
import { messageFromError } from "./errors";

const DEFAULT_EMAIL = "";
const DEFAULT_PASSWORD = "";

export function useDashboardAuth() {
  const token = usePulseStore((s) => s.token);
  const setToken = usePulseStore((s) => s.setToken);
  const [email, setEmail] = useState(DEFAULT_EMAIL);
  const [password, setPassword] = useState(DEFAULT_PASSWORD);
  const [error, setError] = useState("");

  const login = useCallback(async () => {
    setError("");
    try {
      const out = await loginDashboard({ email, password });
      if (out.access_token) setToken(out.access_token);
    } catch (err) {
      setError(messageFromError(err, "Не удалось войти"));
    }
  }, [email, password, setToken]);

  return {
    email,
    error,
    login,
    password,
    setEmail,
    setPassword,
    token,
  };
}
