"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Table } from "../../components/ui/table";
import type { Project } from "../../lib/api";
import { listCompanyProjects } from "./api/companiesApi";
import { usePulseStore } from "../../store/pulse-store";

export default function CompanyProjectsPage() {
  const params = useParams<{ companyId?: string }>();
  const token = usePulseStore((s) => s.token);
  const activeCompanyId = usePulseStore((s) => s.activeCompanyId);
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const setActiveCompany = usePulseStore((s) => s.setActiveCompany);
  const setActiveProject = usePulseStore((s) => s.setActiveProject);
  const setProjects = usePulseStore((s) => s.setProjects);
  const [projects, setLocalProjects] = useState<Project[]>([]);
  const [error, setError] = useState("");
  const companyId = params.companyId || activeCompanyId;

  useEffect(() => {
    if (!token) return;
    if (!companyId) return;
    if (params.companyId) setActiveCompany(companyId);
    listCompanyProjects(companyId, token)
      .then((rows) => {
        const safe = rows;
        setLocalProjects(safe);
        setProjects(safe);
      })
      .catch((err) => setError(err.message));
  }, [companyId, params.companyId, setActiveCompany, setProjects, token]);

  return (
    <main className="page">
      <div className="page-title">
        <div>
          <h1>Проекты компании</h1>
          <p className="muted">Здесь только список проектов. Создание вынесено на отдельный экран.</p>
        </div>
        <Link href="/projects/new">
          <Button>Создать проект</Button>
        </Link>
      </div>
      {error ? <p className="status-failed">{error}</p> : null}
      <Card>
        <CardHeader>
          <CardTitle>Проекты</CardTitle>
            <CardDescription>Откройте проект, чтобы работать со сценариями, интеграциями и входящими запросами.</CardDescription>
        </CardHeader>
        <CardContent>
          {projects.length === 0 ? (
            <div className="empty-state">Проектов пока нет</div>
          ) : (
            <Table>
              <thead>
                <tr>
                  <th>Название</th>
                  <th>Короткий адрес</th>
                  <th>Ключ входа</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {projects.map((project) => (
                  <tr className={project.project_id === activeProjectId ? "is-active" : ""} key={project.project_id}>
                    <td>{project.name}</td>
                    <td>{project.slug}</td>
                    <td className="mono">{project.hook_key || "—"}</td>
                    <td>
                      <Link className="action-link" href="/project" onClick={() => setActiveProject(project.project_id)}>
                        {project.project_id === activeProjectId ? "Активный проект" : "Выбрать проект"}
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>
          )}
        </CardContent>
      </Card>
    </main>
  );
}
