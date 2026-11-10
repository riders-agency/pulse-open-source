"use client";

import { useCallback } from "react";

import { logoutSession } from "../../auth/api/authApi";

export function useShellLogout({
  clearWorkspace,
  setToken,
  token,
}: {
  clearWorkspace: () => void;
  setToken: (token: string) => void;
  token: string | null;
}) {
  return useCallback(async () => {
    if (token) {
      try {
        await logoutSession(token);
      } catch {
        /* сеть / 401 игнорируем */
      }
    }
    setToken("");
    clearWorkspace();
  }, [clearWorkspace, setToken, token]);
}
