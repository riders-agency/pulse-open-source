"use client";

import { useEffect, useState } from "react";

import type { CompanyLegalEntity, Project } from "../../../lib/api";
import { getCompanyClientContext } from "../api/clientDetailApi";

type ProjectOption = Pick<Project, "name" | "project_id">;

export function useCompanyClientContext({
  activeCompanyId,
  projects,
  token,
}: {
  activeCompanyId: string;
  projects: ProjectOption[];
  token: string | null;
}) {
  const [allProjects, setAllProjects] = useState<ProjectOption[]>([]);
  const [legalEntities, setLegalEntities] = useState<CompanyLegalEntity[]>([]);

  useEffect(() => {
    if (!token || !activeCompanyId) return;
    getCompanyClientContext(activeCompanyId, token)
      .then(({ projects: projectRows, legalEntities: entityRows }) => {
        setAllProjects(projectRows);
        setLegalEntities(entityRows);
      })
      .catch(() => {
        setAllProjects(projects);
        setLegalEntities([]);
      });
  }, [activeCompanyId, projects, token]);

  return { allProjects, legalEntities };
}
