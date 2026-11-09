"use client";

import type { ReactNode } from "react";

import { useAuthGateState } from "./model/useAuthGateState";
import { AuthenticatedAppShell } from "./ui/AuthenticatedAppShell";
import { AuthLoginScreen } from "./ui/AuthLoginScreen";

export function AuthGate({ children }: { children: ReactNode }) {
  const auth = useAuthGateState();

  if (auth.isLoginRequired) {
    return (
      <AuthLoginScreen
        ready={auth.hydrated}
        busy={auth.busy}
        displayName={auth.displayName}
        email={auth.email}
        error={auth.error}
        mode={auth.mode}
        password={auth.password}
        setDisplayName={auth.setDisplayName}
        setEmail={auth.setEmail}
        setMode={auth.setMode}
        setPassword={auth.setPassword}
        submit={auth.submit}
      />
    );
  }

  return <AuthenticatedAppShell sidebarCollapsed={auth.sidebarCollapsed}>{children}</AuthenticatedAppShell>;
}
