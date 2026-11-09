"use client";

import { Select } from "../../../components/ui/select";
import type { useShellWorkspace } from "../model/useShellWorkspace";

type ShellWorkspace = ReturnType<typeof useShellWorkspace>;

export function ShellWorkspaceSwitcher({ workspace }: { workspace: ShellWorkspace }) {
  const { activeCompanyId, activeProjectId, companies, projects, switchCompany, switchProject, workspaceError } = workspace;

  return (
    <div className="sidebar-workspace" aria-label="Рабочий контекст">
      <label>
        <span>Компания</span>
        <Select
          aria-label="Активная компания"
          value={activeCompanyId}
          onChange={(event) => switchCompany(event.target.value)}
          disabled={companies.length === 0}
        >
          {companies.length === 0 ? <option value="">Компания</option> : null}
          {companies.map((company) => (
            <option value={company.company_id} key={company.company_id}>
              {company.name}
            </option>
          ))}
        </Select>
      </label>
      <label>
        <span>Проект</span>
        <Select
          aria-label="Активный проект"
          value={activeProjectId}
          onChange={(event) => switchProject(event.target.value)}
          disabled={projects.length === 0}
        >
          {projects.length === 0 ? <option value="">Проект</option> : null}
          {projects.map((project) => (
            <option value={project.project_id} key={project.project_id}>
              {project.name}
            </option>
          ))}
        </Select>
      </label>
      {workspaceError ? <div className="sidebar-error">{workspaceError}</div> : null}
    </div>
  );
}
