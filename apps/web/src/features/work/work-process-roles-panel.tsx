"use client";

import { useEffect, useState } from "react";
import { Plus, RefreshCcw, Save } from "lucide-react";

import { Button } from "../../components/ui/button";
import { Field, Input } from "../../components/ui/input";
import type { WorkProcessRole, WorkRolePermissions, WorkWorkflow } from "../../lib/api";
import { createProcessRole, resetProcessRoles, updateProcessRole } from "./api/workApi";

const permissionCatalog: Array<{ key: keyof WorkRolePermissions & string; label: string; description: string }> = [
  { key: "read_tasks", label: "Видит задачи", description: "Может открывать задачи этого процесса." },
  { key: "create_tasks", label: "Создает задачи", description: "Может заводить задачи в колонках процесса." },
  { key: "edit_tasks", label: "Редактирует задачи", description: "Может менять поля, чеклисты и кастомные поля." },
  { key: "move_tasks", label: "Двигает статусы", description: "Может переносить задачи между статусами workflow." },
  { key: "comment_tasks", label: "Комментирует", description: "Может писать комментарии в задачах." },
  { key: "log_time", label: "Ведет время", description: "Может запускать таймеры и добавлять time entries." },
  { key: "view_reports", label: "Видит отчеты", description: "Видит отчеты только по доступным процессам." },
  { key: "approve_time", label: "Утверждает время", description: "Может работать с согласованием времени." },
  { key: "invite_members", label: "Приглашает", description: "Может приглашать людей в этот процесс." },
  { key: "manage_roles", label: "Редактирует роли", description: "Может менять роли и разрешения процесса." },
  { key: "manage_workflow", label: "Редактирует процесс", description: "Может менять процесс, доски и workflow-настройки." },
  { key: "view_process", label: "Видит настройки", description: "Может открыть экран процесса без административных действий." },
];

export function ProcessRolesPanel({ token, workflow, onReload, onError }: { token?: string | null; workflow: WorkWorkflow; onReload: () => Promise<void>; onError: (value: string) => void }) {
  const [roles, setRoles] = useState<WorkProcessRole[]>(workflow.roles || []);
  const [saving, setSaving] = useState("");
  const [newRoleOpen, setNewRoleOpen] = useState(false);
  const [newName, setNewName] = useState("Новая роль процесса");
  const [newDescription, setNewDescription] = useState("");

  useEffect(() => {
    setRoles(workflow.roles || []);
  }, [workflow]);

  const updateRole = (roleId: string, patch: Partial<WorkProcessRole>) => {
    setRoles((current) => current.map((role) => (role.task_process_role_id === roleId ? { ...role, ...patch } : role)));
  };

  const togglePermission = (roleId: string, key: string, value: boolean) => {
    setRoles((current) => current.map((role) => {
      if (role.task_process_role_id !== roleId) return role;
      return { ...role, permissions: { ...role.permissions, [key]: value } };
    }));
  };

  const save = async (role: WorkProcessRole) => {
    if (!token) return;
    setSaving(role.task_process_role_id);
    try {
      await updateProcessRole(workflow.task_workflow_id, role.task_process_role_id, token, {
        name: role.name,
        description: role.description,
        permissions: role.permissions,
        position: role.position,
      });
      await onReload();
    } finally {
      setSaving("");
    }
  };

  const create = async () => {
    if (!token) return;
    await createProcessRole(workflow.task_workflow_id, token, {
      name: newName,
      description: newDescription,
      permissions: { read_tasks: true, comment_tasks: true },
      position: roles.length * 10 + 70,
    });
    setNewRoleOpen(false);
    setNewName("Новая роль процесса");
    setNewDescription("");
    await onReload();
  };

  const reset = async () => {
    if (!token) return;
    await resetProcessRoles(workflow.task_workflow_id, token);
    await onReload();
  };

  return (
    <section className="work-roles-panel">
      <header>
        <div>
          <h3>Роли и доступ</h3>
          <p>Роли живут только внутри этого процесса. Их можно редактировать, добавлять заново и пересоздавать из русских предустановок.</p>
        </div>
        <div className="work-role-actions">
          <Button variant="secondary" size="sm" onClick={() => setNewRoleOpen(!newRoleOpen)}><Plus size={15} /> Роль</Button>
          <Button variant="secondary" size="sm" onClick={() => void reset().catch((err) => onError(err.message))}><RefreshCcw size={15} /> Пересоздать</Button>
        </div>
      </header>
      {newRoleOpen ? (
        <form
          className="work-role-create"
          onSubmit={(event) => {
            event.preventDefault();
            void create().catch((err) => onError(err.message));
          }}
        >
          <Field label="Название"><Input value={newName} onChange={(event) => setNewName(event.target.value)} /></Field>
          <Field label="Описание"><Input value={newDescription} onChange={(event) => setNewDescription(event.target.value)} /></Field>
          <div className="form-actions"><Button type="submit">Добавить роль</Button></div>
        </form>
      ) : null}
      <div className="work-role-list">
        {roles.map((role) => (
          <article key={role.task_process_role_id} className="work-role-card">
            <header>
              <div>
                <Input value={role.name} onChange={(event) => updateRole(role.task_process_role_id, { name: event.target.value })} aria-label="Название роли" />
                <Input value={role.description} onChange={(event) => updateRole(role.task_process_role_id, { description: event.target.value })} aria-label="Описание роли" />
              </div>
              <Button size="sm" onClick={() => void save(role).catch((err) => onError(err.message))} disabled={saving === role.task_process_role_id}>
                <Save size={15} /> {saving === role.task_process_role_id ? "Сохранение" : "Сохранить"}
              </Button>
            </header>
            <div className="work-permission-grid">
              {permissionCatalog.map((permission) => (
                <label key={`${role.task_process_role_id}-${permission.key}`} className="work-permission">
                  <input
                    type="checkbox"
                    checked={Boolean(role.permissions?.[permission.key])}
                    onChange={(event) => togglePermission(role.task_process_role_id, permission.key, event.target.checked)}
                  />
                  <span>
                    <strong>{permission.label}</strong>
                    <small>{permission.description}</small>
                  </span>
                </label>
              ))}
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}
