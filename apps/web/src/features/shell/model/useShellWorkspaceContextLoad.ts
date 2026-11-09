"use client";

import { useEffect, type MutableRefObject } from "react";

import { getWorkspaceContext } from "../api/shellApi";

export function useShellWorkspaceContextLoad({
  lastSavedWorkspace,
  setActiveCompany,
  setActiveProject,
  setWorkspaceError,
  token,
  workspaceContextLoaded,
  workspaceProjectFromDB,
}: {
  lastSavedWorkspace: MutableRefObject<string>;
  setActiveCompany: (companyId: string) => void;
  setActiveProject: (projectId: string) => void;
  setWorkspaceError: (message: string) => void;
  token: string | null;
  workspaceContextLoaded: MutableRefObject<boolean>;
  workspaceProjectFromDB: MutableRefObject<string>;
}) {
  useEffect(() => {
    workspaceContextLoaded.current = false;
    workspaceProjectFromDB.current = "";
    lastSavedWorkspace.current = "";
    if (!token) return;
    let cancelled = false;
    getWorkspaceContext(token)
      .then((ctx) => {
        if (cancelled) return;
        workspaceContextLoaded.current = true;
        workspaceProjectFromDB.current = ctx.active_project_id || "";
        if (ctx.active_company_id) setActiveCompany(ctx.active_company_id);
        if (ctx.active_project_id) setActiveProject(ctx.active_project_id);
      })
      .catch((err) => {
        if (cancelled) return;
        workspaceContextLoaded.current = true;
        setWorkspaceError(err instanceof Error ? err.message : "Не удалось загрузить рабочий контекст");
      });
    return () => {
      cancelled = true;
    };
  }, [lastSavedWorkspace, setActiveCompany, setActiveProject, setWorkspaceError, token, workspaceContextLoaded, workspaceProjectFromDB]);
}
