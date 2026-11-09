"use client";

import { usePathname, useRouter } from "next/navigation";
import { useMemo, useRef, useState } from "react";

import { usePulseStore } from "../../../store/pulse-store";
import { useShellAccess } from "./useShellAccess";
import { useShellChromeState } from "./useShellChromeState";
import { useShellWorkspaceData } from "./useShellWorkspaceData";
import { useShellLogout } from "./useShellLogout";
import { useShellWorkspaceContextLoad } from "./useShellWorkspaceContextLoad";
import { useShellRouteContextSync } from "./useShellRouteContextSync";
import { useShellWorkspaceSwitchers } from "./useShellWorkspaceSwitchers";

export type { ShellTaskNavIcon, ShellTaskNavItem } from "./shellTaskNav";

export function useShellWorkspace() {
  const pathname = usePathname();
  const router = useRouter();
  const rawCompanies = usePulseStore((s) => s.companies);
  const rawProjects = usePulseStore((s) => s.projects);
  const activeCompanyId = usePulseStore((s) => s.activeCompanyId);
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const setActiveCompany = usePulseStore((s) => s.setActiveCompany);
  const setActiveProject = usePulseStore((s) => s.setActiveProject);
  const setCompanies = usePulseStore((s) => s.setCompanies);
  const setProjects = usePulseStore((s) => s.setProjects);
  const theme = usePulseStore((s) => s.theme);
  const setTheme = usePulseStore((s) => s.setTheme);
  const sidebarCollapsed = usePulseStore((s) => s.sidebarCollapsed);
  const setSidebarCollapsed = usePulseStore((s) => s.setSidebarCollapsed);
  const token = usePulseStore((s) => s.token);
  const setToken = usePulseStore((s) => s.setToken);
  const clearWorkspace = usePulseStore((s) => s.clearWorkspace);
  const workspaceContextLoaded = useRef(false);
  const workspaceProjectFromDB = useRef("");
  const lastSavedWorkspace = useRef("");
  const projectsLoadedForCompany = useRef("");
  const [workspaceError, setWorkspaceError] = useState("");
  const companies = Array.isArray(rawCompanies) ? rawCompanies : [];
  const projects = Array.isArray(rawProjects) ? rawProjects : [];
  const { tasksNavOpen, toggleTasksNav } = useShellChromeState(pathname, theme);
  const { markProjectRouteSynced, resetRouteContextSync } = useShellRouteContextSync({
    pathname,
    setActiveCompany,
    setActiveProject,
    setWorkspaceError,
    token,
  });
  const { switchCompany, switchProject } = useShellWorkspaceSwitchers({
    activeCompanyId,
    activeProjectId,
    markProjectRouteSynced,
    pathname,
    resetRouteContextSync,
    router,
    setActiveCompany,
    setActiveProject,
    setWorkspaceError,
  });
  const { canManageProcesses, canUse, canUseDocs, canUseWorkspaceModules, taskNavItems, tasksActive } = useShellAccess({
    activeProjectId,
    pathname,
    token,
  });
  useShellWorkspaceContextLoad({ lastSavedWorkspace, setActiveCompany, setActiveProject, setWorkspaceError, token, workspaceContextLoaded, workspaceProjectFromDB });
  useShellWorkspaceData({
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
  });

  const activeCompany = useMemo(() => companies.find((company) => company.company_id === activeCompanyId), [activeCompanyId, companies]);
  const activeProject = useMemo(() => projects.find((project) => project.project_id === activeProjectId), [activeProjectId, projects]);

  const logout = useShellLogout({ clearWorkspace, setToken, token });

  return {
    activeCompany,
    activeCompanyId,
    activeProject,
    activeProjectId,
    canManageProcesses,
    canUse,
    canUseDocs,
    canUseWorkspaceModules,
    companies,
    logout,
    pathname,
    projects,
    sidebarCollapsed,
    switchCompany,
    switchProject,
    taskNavItems,
    tasksActive,
    tasksNavOpen,
    theme,
    toggleSidebar: () => setSidebarCollapsed(!sidebarCollapsed),
    toggleTasksNav,
    toggleTheme: () => setTheme(theme === "dark" ? "light" : "dark"),
    workspaceError,
  };
}
