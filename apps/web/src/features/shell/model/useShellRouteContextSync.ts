"use client";

import { useCallback, useEffect, useRef } from "react";

import { projectIdFromPath, runIdFromPath } from "../../../lib/workspace-routing";
import { getShellProject, getShellRun } from "../api/shellApi";

export function useShellRouteContextSync({
  pathname,
  setActiveCompany,
  setActiveProject,
  setWorkspaceError,
  token,
}: {
  pathname: string;
  setActiveCompany: (companyId: string) => void;
  setActiveProject: (projectId: string) => void;
  setWorkspaceError: (message: string) => void;
  token: string;
}) {
  const syncedRouteProjectId = useRef("");

  useEffect(() => {
    const routeProjectId = projectIdFromPath(pathname);
    const routeRunId = runIdFromPath(pathname);
    const syncKey = routeProjectId ? `project:${routeProjectId}` : routeRunId ? `run:${routeRunId}` : "";
    if (!token || !syncKey) {
      if (!syncKey) syncedRouteProjectId.current = "";
      return;
    }
    if (syncedRouteProjectId.current === syncKey) return;
    let cancelled = false;
    const loadContext = routeProjectId
      ? getShellProject(routeProjectId, token)
      : getShellRun(routeRunId, token).then((run) => ({
          company_id: run.company_id,
          project_id: run.project_id,
        }));
    loadContext
      .then((project) => {
        if (cancelled) return;
        syncedRouteProjectId.current = syncKey;
        if (project.company_id) setActiveCompany(project.company_id);
        setActiveProject(project.project_id);
        setWorkspaceError("");
      })
      .catch((err) => {
        if (!cancelled) setWorkspaceError(err instanceof Error ? err.message : "Не удалось загрузить профиль проекта");
      });
    return () => {
      cancelled = true;
    };
  }, [pathname, setActiveCompany, setActiveProject, setWorkspaceError, token]);

  return {
    markProjectRouteSynced: useCallback((projectId: string) => {
      syncedRouteProjectId.current = `project:${projectId}`;
    }, []),
    resetRouteContextSync: useCallback(() => {
      syncedRouteProjectId.current = "";
    }, []),
  };
}
