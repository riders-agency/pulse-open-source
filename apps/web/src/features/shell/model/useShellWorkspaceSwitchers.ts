"use client";

import { useCallback } from "react";

import { companySwitchPath, projectSwitchPath } from "../../../lib/workspace-routing";

type ShellRouter = { push: (href: string) => void };

export function useShellWorkspaceSwitchers({
  activeCompanyId,
  activeProjectId,
  markProjectRouteSynced,
  pathname,
  resetRouteContextSync,
  router,
  setActiveCompany,
  setActiveProject,
  setWorkspaceError,
}: {
  activeCompanyId: string;
  activeProjectId: string;
  markProjectRouteSynced: (projectId: string) => void;
  pathname: string;
  resetRouteContextSync: () => void;
  router: ShellRouter;
  setActiveCompany: (companyId: string) => void;
  setActiveProject: (projectId: string) => void;
  setWorkspaceError: (message: string) => void;
}) {
  const switchCompany = useCallback(
    (nextCompanyId: string) => {
      if (!nextCompanyId || nextCompanyId === activeCompanyId) return;
      setWorkspaceError("");
      resetRouteContextSync();
      setActiveCompany(nextCompanyId);
      router.push(companySwitchPath(pathname, activeCompanyId, nextCompanyId));
    },
    [activeCompanyId, pathname, resetRouteContextSync, router, setActiveCompany, setWorkspaceError],
  );

  const switchProject = useCallback(
    (nextProjectId: string) => {
      if (!nextProjectId || nextProjectId === activeProjectId) return;
      setWorkspaceError("");
      markProjectRouteSynced(nextProjectId);
      setActiveProject(nextProjectId);
      router.push(projectSwitchPath(pathname, activeProjectId, nextProjectId));
    },
    [activeProjectId, markProjectRouteSynced, pathname, router, setActiveProject, setWorkspaceError],
  );

  return { switchCompany, switchProject };
}
