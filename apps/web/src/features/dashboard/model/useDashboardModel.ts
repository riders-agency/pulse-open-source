"use client";

import { useDashboardAuth } from "./useDashboardAuth";
import { useDashboardCompanies } from "./useDashboardCompanies";
import { useDashboardProjects } from "./useDashboardProjects";
import { useDashboardWorkflowRuns } from "./useDashboardWorkflowRuns";

export function useDashboardModel() {
  const auth = useDashboardAuth();
  const companies = useDashboardCompanies(auth.token);
  const projects = useDashboardProjects(auth.token, companies.activeCompanyId);
  const workflowRuns = useDashboardWorkflowRuns(auth.token, projects.activeProjectId);

  return {
    ...auth,
    ...companies,
    ...projects,
    ...workflowRuns,
    error: auth.error || companies.error || projects.error || workflowRuns.error,
  };
}

export type DashboardModel = ReturnType<typeof useDashboardModel>;
