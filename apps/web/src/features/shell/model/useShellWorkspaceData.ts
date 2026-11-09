import { useEffect } from "react";

import type { Company } from "../../../lib/api";
import type { ProjectMini } from "../../../store/pulse-store";
import {
  listShellCompanies,
  listShellCompanyProjects,
  saveWorkspaceContext,
  switchCompanyContext,
} from "../api/shellApi";

type WorkspaceRef<T> = { current: T };

export function useShellWorkspaceData({
  activeCompanyId,
  activeProjectId,
  clearWorkspace,
  lastSavedWorkspace,
  projects,
  projectsLoadedForCompany,
  setActiveCompany,
  setActiveProject,
  setCompanies,
  setProjects,
  setWorkspaceError,
  token,
  workspaceContextLoaded,
  workspaceProjectFromDB,
}: {
  activeCompanyId: string;
  activeProjectId: string;
  clearWorkspace: () => void;
  lastSavedWorkspace: WorkspaceRef<string>;
  projects: ProjectMini[];
  projectsLoadedForCompany: WorkspaceRef<string>;
  setActiveCompany: (id: string) => void;
  setActiveProject: (id: string) => void;
  setCompanies: (rows: Company[]) => void;
  setProjects: (rows: ProjectMini[]) => void;
  setWorkspaceError: (value: string) => void;
  token?: string | null;
  workspaceContextLoaded: WorkspaceRef<boolean>;
  workspaceProjectFromDB: WorkspaceRef<string>;
}) {
  useEffect(() => {
    if (!token) return;
    let cancelled = false;
    listShellCompanies(token)
      .then((safe) => {
        if (cancelled) return;
        setCompanies(safe);
        if (safe.length === 0) {
          clearWorkspace();
          return;
        }
        if (!activeCompanyId || !safe.some((company) => company.company_id === activeCompanyId)) {
          setActiveCompany(safe[0].company_id);
        }
        setWorkspaceError("");
      })
      .catch((err) => {
        if (!cancelled) setWorkspaceError(err instanceof Error ? err.message : "Не удалось загрузить компании");
      });
    return () => {
      cancelled = true;
    };
  }, [activeCompanyId, clearWorkspace, setActiveCompany, setCompanies, setWorkspaceError, token]);

  useEffect(() => {
    if (!token || !activeCompanyId) return;
    let cancelled = false;
    switchCompanyContext(activeCompanyId, token).catch((err) => {
      if (!cancelled) setWorkspaceError(err instanceof Error ? err.message : "Не удалось переключить компанию");
    });
    return () => {
      cancelled = true;
    };
  }, [activeCompanyId, setWorkspaceError, token]);

  useEffect(() => {
    if (!token || !workspaceContextLoaded.current) return;
    if (activeCompanyId && !activeProjectId && workspaceProjectFromDB.current) return;
    if (activeCompanyId && projectsLoadedForCompany.current !== activeCompanyId) return;
    if (activeProjectId && !projects.some((project) => project.project_id === activeProjectId)) return;
    const key = `${activeCompanyId || ""}:${activeProjectId || ""}`;
    if (lastSavedWorkspace.current === key) return;
    lastSavedWorkspace.current = key;
    saveWorkspaceContext(token, {
      active_company_id: activeCompanyId || null,
      active_project_id: activeProjectId || null,
    }).catch((err) => {
      setWorkspaceError(err instanceof Error ? err.message : "Не удалось сохранить рабочий контекст");
    });
  }, [activeCompanyId, activeProjectId, lastSavedWorkspace, projects, projectsLoadedForCompany, setWorkspaceError, token, workspaceContextLoaded, workspaceProjectFromDB]);

  useEffect(() => {
    if (!token || !activeCompanyId) {
      projectsLoadedForCompany.current = "";
      setProjects([]);
      return;
    }
    projectsLoadedForCompany.current = "";
    let cancelled = false;
    listShellCompanyProjects(activeCompanyId, token)
      .then((safe) => {
        if (cancelled) return;
        setProjects(safe);
        projectsLoadedForCompany.current = activeCompanyId;
        const fromDB = workspaceProjectFromDB.current;
        const fallbackProject = safe.find((project) => project.project_id === fromDB) ?? safe[0];
        if (workspaceContextLoaded.current && activeProjectId && !safe.some((project) => project.project_id === activeProjectId)) {
          setActiveProject(fallbackProject?.project_id || "");
          return;
        }
        if (workspaceContextLoaded.current && !activeProjectId && fallbackProject) setActiveProject(fallbackProject.project_id);
        setWorkspaceError("");
      })
      .catch((err) => {
        if (!cancelled) {
          setProjects([]);
          setWorkspaceError(err instanceof Error ? err.message : "Не удалось загрузить проекты компании");
        }
      });
    return () => {
      cancelled = true;
    };
  }, [activeCompanyId, activeProjectId, projectsLoadedForCompany, setActiveProject, setProjects, setWorkspaceError, token, workspaceContextLoaded, workspaceProjectFromDB]);
}
