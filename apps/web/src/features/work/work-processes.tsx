"use client";

import { useEffect, useMemo, useState } from "react";
import { Plus } from "lucide-react";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Field, Input } from "../../components/ui/input";
import type { WorkAccess, WorkWorkflow, WorkWorkflowTemplate } from "../../lib/api";
import { usePulseStore } from "../../store/pulse-store";
import { createTaskWorkflow, getWorkPresets, listTaskWorkflows } from "./api/workApi";
import { WorkFrame, WorkNoProject } from "./work-frame";
import { ProcessRolesPanel } from "./work-process-roles-panel";
import { statusLabel } from "./work-utils";

export function WorkProcessesPage() {
  const token = usePulseStore((s) => s.token);
  const projectId = usePulseStore((s) => s.activeProjectId);
  const [access, setAccess] = useState<WorkAccess>();
  const [items, setItems] = useState<WorkWorkflow[]>([]);
  const [templates, setTemplates] = useState<WorkWorkflowTemplate[]>([]);
  const [selectedWorkflowId, setSelectedWorkflowId] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [name, setName] = useState("");
  const [templateKey, setTemplateKey] = useState("client_request");
  const [statuses, setStatuses] = useState("Бэклог, Запланировано, В работе, Ревью, Готово");
  const [error, setError] = useState("");

  const load = async () => {
    if (!token || !projectId) return;
    const [payload, presets] = await Promise.all([
      listTaskWorkflows(projectId, token),
      getWorkPresets(token),
    ]);
    const workflows = payload.items || [];
    setAccess(payload.access);
    setItems(workflows);
    setTemplates(presets.workflow_templates || []);
    setSelectedWorkflowId((current) => current || workflows[0]?.task_workflow_id || "");
    setError("");
  };

  useEffect(() => {
    void load().catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить процессы"));
  }, [projectId, token]);

  const selected = useMemo(
    () => items.find((workflow) => workflow.task_workflow_id === selectedWorkflowId) || items[0],
    [items, selectedWorkflowId],
  );
  const template = templates.find((item) => item.key === templateKey);

  if (!projectId) return <WorkNoProject />;

  const create = async () => {
    if (!token || !projectId) return;
    const workflow = await createTaskWorkflow(token, {
      project_id: projectId,
      name,
      template_key: templateKey || undefined,
      status_names: templateKey ? [] : statuses.split(",").map((item) => item.trim()).filter(Boolean),
      create_default_boards: true,
    });
    setName("");
    setCreateOpen(false);
    setSelectedWorkflowId(workflow.task_workflow_id);
    await load();
  };

  return (
    <WorkFrame
      access={access}
      title="Процессы задач"
      description="Каждый процесс имеет собственные статусы, доски, роли и разрешения. Исполнитель процесса не видит другие процессы проекта."
      actions={access?.can_manage_processes ? <Button variant="secondary" onClick={() => setCreateOpen(!createOpen)}><Plus size={16} /> Создать процесс</Button> : null}
    >
      {error ? <div className="form-error">{error}</div> : null}
      {!access ? (
        <section className="work-empty"><h2>Загрузка процессов</h2><p>Проверяем роль и доступ к настройкам workflow.</p></section>
      ) : !access.can_manage_processes ? (
        <section className="work-empty"><h2>Процессы скрыты</h2><p>Ваша роль может работать с задачами, но не видеть и не редактировать процессы проекта.</p></section>
      ) : (
        <>
          {createOpen ? (
            <form
              className="work-process-form"
              onSubmit={(event) => {
                event.preventDefault();
                void create().catch((err) => setError(err.message));
              }}
            >
              <Field label="Предустановка">
                <select className="ui-input" value={templateKey} onChange={(event) => setTemplateKey(event.target.value)}>
                  {templates.map((item) => <option key={item.key} value={item.key}>{item.name}</option>)}
                  <option value="">Кастомный процесс</option>
                </select>
              </Field>
              <Field label="Название"><Input value={name} onChange={(event) => setName(event.target.value)} placeholder={template?.name || "Новый процесс"} /></Field>
              {!templateKey ? <Field label="Статусы"><Input value={statuses} onChange={(event) => setStatuses(event.target.value)} /></Field> : null}
              <div className="form-actions"><Button type="submit">Создать</Button></div>
              {template ? (
                <section className="work-template-preview">
                  <p>{template.description}</p>
                  <div className="work-status-chain">
                    {template.statuses.map((status) => <span key={status.name}>{statusLabel(status.name)}</span>)}
                  </div>
                </section>
              ) : null}
            </form>
          ) : null}
          <section className="work-process-shell">
            <nav className="work-process-nav" aria-label="Процессы проекта">
              {items.map((workflow) => (
                <button
                  key={workflow.task_workflow_id}
                  type="button"
                  className={workflow.task_workflow_id === selected?.task_workflow_id ? "is-active" : ""}
                  onClick={() => setSelectedWorkflowId(workflow.task_workflow_id)}
                >
                  <strong>{workflow.name}</strong>
                  <span>{workflow.statuses?.length || 0} статусов · {workflow.roles?.length || 0} ролей</span>
                </button>
              ))}
            </nav>
            {selected ? (
              <article className="work-process-detail">
                <header>
                  <div>
                    <h2>{selected.name}</h2>
                    <p>{selected.description || "Процесс проекта с собственными ролями, досками и правилами переходов."}</p>
                  </div>
                  {selected.is_default ? <Badge tone="success">Основной</Badge> : null}
                </header>
                <div className="work-status-chain">
                  {(selected.statuses || []).map((status) => <span key={status.task_workflow_status_id}>{statusLabel(status.name)}</span>)}
                </div>
                <section className="work-integration-note">
                  <strong>Связка с automation</strong>
                  <span>События задач `task.created`, `task.updated`, `task.status_changed`, `task.comment_added`, `time.logged` можно подключать в `/project-workflows` как действие процесса и отправлять в Telegram, CRM или другие каналы.</span>
                </section>
                <ProcessRolesPanel token={token} workflow={selected} onReload={load} onError={setError} />
              </article>
            ) : (
              <section className="work-empty"><h2>Нет процессов</h2><p>Создайте первый процесс по предустановке или вручную.</p></section>
            )}
          </section>
        </>
      )}
    </WorkFrame>
  );
}
