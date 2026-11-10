"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

import { asArray, type Project as ProjectRow } from "../../../lib/api";
import { usePulseStore } from "../../../store/pulse-store";
import { listDashboardProjects } from "../api/dashboardApi";
import { messageFromError } from "./errors";

export function useDashboardProjects(token: string | undefined, activeCompanyId?: string | null) {
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const setActiveProject = usePulseStore((s) => s.setActiveProject);
  const projectsRaw = usePulseStore((s) => s.projects);
  const setProjects = usePulseStore((s) => s.setProjects);
  const [error, setError] = useState("");
  const projects = useMemo(() => asArray<ProjectRow>(projectsRaw), [projectsRaw]);
  const activeProject = useMemo(
    () => projects.find((project) => project.project_id === activeProjectId),
    [activeProjectId, projects],
  );

  const loadProjects = useCallback(async () => {
    if (!token || !activeCompanyId) return;
    const safe = await listDashboardProjects(activeCompanyId, token);
    setProjects(safe);
    if (!activeProjectId && safe[0]?.project_id) setActiveProject(safe[0].project_id);
  }, [activeCompanyId, activeProjectId, setActiveProject, setProjects, token]);

  useEffect(() => {
    void loadProjects().catch((err) => setError(messageFromError(err, "Не удалось загрузить проекты")));
  }, [loadProjects]);

  return {
    activeProject,
    activeProjectId,
    error,
  };
}
