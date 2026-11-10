"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import type { ClientServiceType, Integration, Project, Run, Workflow } from "../../lib/api";
import {
  archiveProjectServiceType,
  createProjectServiceType,
  getProjectDashboard,
  listProjectServiceTypes,
  updateProjectServiceType,
} from "./api/projectDetailApi";
import { usePulseStore } from "../../store/pulse-store";
import { ProjectServiceTypesPanel } from "./ui/ProjectServiceTypesPanel";

export default function ProjectPage() {
  const params = useParams<{ projectId?: string }>();
  const token = usePulseStore((s) => s.token);
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const setActiveProject = usePulseStore((s) => s.setActiveProject);
  const [project, setProject] = useState<Project | null>(null);
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [integrations, setIntegrations] = useState<Integration[]>([]);
  const [runs, setRuns] = useState<Run[]>([]);
  const [serviceTypes, setServiceTypes] = useState<ClientServiceType[]>([]);
  const [newServiceType, setNewServiceType] = useState({ display_name: "", service_type_key: "", description: "" });
  const [error, setError] = useState("");
  const [busyTypeId, setBusyTypeId] = useState("");
  const projectId = params.projectId || activeProjectId;

  const loadServiceTypes = async (id: string) => {
    if (!token) return;
    setServiceTypes(await listProjectServiceTypes(id, token));
  };

  useEffect(() => {
    if (!token) return;
    if (!projectId) return;
    if (params.projectId) setActiveProject(projectId);
    getProjectDashboard(projectId, token)
      .then(({ project: projectRow, workflows: safeWorkflows, integrations: integrationRows, serviceTypes: typeRows, runs: runRows }) => {
        setProject(projectRow);
        setWorkflows(safeWorkflows);
        setIntegrations(integrationRows);
        setServiceTypes(typeRows);
        setRuns(runRows);
      })
      .catch((err) => setError(err.message));
  }, [projectId, params.projectId, setActiveProject, token]);

  const patchServiceType = async (type: ClientServiceType) => {
    if (!token || !projectId) return;
    setBusyTypeId(type.client_service_type_id);
    setError("");
    try {
      await updateProjectServiceType(projectId, type.client_service_type_id, token, {
        service_type_key: type.service_type_key,
        display_name: type.display_name,
        description: type.description ?? "",
        sort_order: type.sort_order,
      });
      await loadServiceTypes(projectId);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить тип услуги");
    } finally {
      setBusyTypeId("");
    }
  };

  const archiveServiceType = async (type: ClientServiceType) => {
    if (!token || !projectId) return;
    setBusyTypeId(type.client_service_type_id);
    setError("");
    try {
      await archiveProjectServiceType(projectId, type.client_service_type_id, token);
      await loadServiceTypes(projectId);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось архивировать тип услуги");
    } finally {
      setBusyTypeId("");
    }
  };

  const createServiceType = async () => {
    if (!token || !projectId || !newServiceType.display_name.trim()) return;
    setBusyTypeId("new");
    setError("");
    try {
      await createProjectServiceType(projectId, token, {
        display_name: newServiceType.display_name,
        service_type_key: newServiceType.service_type_key,
        description: newServiceType.description,
        sort_order: serviceTypes.length ? Math.max(...serviceTypes.map((type) => type.sort_order || 100)) + 10 : 100,
      });
      setNewServiceType({ display_name: "", service_type_key: "", description: "" });
      await loadServiceTypes(projectId);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось добавить тип услуги");
    } finally {
      setBusyTypeId("");
    }
  };

  return (
    <main className="page">
      <div className="page-title">
        <div>
          <h1>{project?.name ?? "Проект"}</h1>
          <p className="muted">Обзор проекта. Редактирование сущностей находится на отдельных экранах.</p>
        </div>
        <div className="row">
          <Link href="/project-workflows/new"><Button>Создать сценарий</Button></Link>
          <Link href="/project-integrations/new"><Button variant="secondary">Добавить интеграцию</Button></Link>
        </div>
      </div>
      {error ? <Badge tone="danger">{error}</Badge> : null}
      <section className="grid-3">
        <div className="metric"><span>Сценарии</span><strong>{workflows.length}</strong></div>
        <div className="metric"><span>Интеграции</span><strong>{integrations.length}</strong></div>
        <div className="metric"><span>Последние запуски</span><strong>{runs.length}</strong></div>
      </section>
      <section className="grid-2 section-panel-spaced">
        <Card>
          <CardHeader>
            <CardTitle>Основные разделы</CardTitle>
            <CardDescription>Путь пользователя: вебхук пришел → действия выполнились → результат виден.</CardDescription>
          </CardHeader>
          <CardContent className="stack">
            <Link className="between" href="/project-workflows"><span>Сценарии</span><Badge>Открыть</Badge></Link>
            <Link className="between" href="/project-integrations"><span>Интеграции</span><Badge>Открыть</Badge></Link>
            <Link className="between" href="/project-payments"><span>Оплаты</span><Badge>Открыть</Badge></Link>
            <Link className="between" href="/project#service-types"><span>Типы услуг</span><Badge>Открыть</Badge></Link>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Вебхук проекта</CardTitle>
            <CardDescription>Ключ проекта используется в URL входящих вебхуков.</CardDescription>
          </CardHeader>
          <CardContent className="stack">
            <div className="between"><span className="muted">Ключ входа</span><span className="mono">{project?.hook_key || "—"}</span></div>
            <pre className="code-block">{`POST /hooks/${project?.hook_key || "{project_key}"}/{endpoint_key}`}</pre>
          </CardContent>
        </Card>
      </section>
      <ProjectServiceTypesPanel
        archiveServiceType={archiveServiceType}
        busyTypeId={busyTypeId}
        createServiceType={createServiceType}
        newServiceType={newServiceType}
        patchServiceType={patchServiceType}
        serviceTypes={serviceTypes}
        setNewServiceType={setNewServiceType}
        setServiceTypes={setServiceTypes}
      />
    </main>
  );
}
