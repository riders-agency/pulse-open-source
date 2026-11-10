"use client";

import { useEffect, useMemo, useState } from "react";
import { UserPlus } from "lucide-react";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Field, Input } from "../../components/ui/input";
import { Select } from "../../components/ui/select";
import { Table } from "../../components/ui/table";
import type { Client, WorkAccess, WorkEmployee, WorkWorkflow } from "../../lib/api";
import { usePulseStore } from "../../store/pulse-store";
import { inviteWorkProjectMember, listCompanyClients, listTaskWorkflows, listWorkEmployees } from "./api/workApi";
import { WorkFrame, WorkNoProject } from "./work-frame";
import { minutesLabel } from "./work-utils";

export function WorkEmployeesPage() {
  const token = usePulseStore((s) => s.token);
  const projectId = usePulseStore((s) => s.activeProjectId);
  const companyId = usePulseStore((s) => s.activeCompanyId);
  const [access, setAccess] = useState<WorkAccess>();
  const [rows, setRows] = useState<WorkEmployee[]>([]);
  const [workflows, setWorkflows] = useState<WorkWorkflow[]>([]);
  const [clients, setClients] = useState<Client[]>([]);
  const [inviteOpen, setInviteOpen] = useState(false);
  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [workflowId, setWorkflowId] = useState("");
  const [roleId, setRoleId] = useState("");
  const [clientId, setClientId] = useState("");
  const [inviteResult, setInviteResult] = useState("");
  const [error, setError] = useState("");

  const load = async () => {
    if (!token || !projectId) return;
    const [employeesPayload, workflowsPayload] = await Promise.all([
      listWorkEmployees(projectId, token),
      listTaskWorkflows(projectId, token),
    ]);
    const workflowRows = workflowsPayload.items || [];
    setAccess(employeesPayload.access);
    setRows(employeesPayload.items || []);
    setWorkflows(workflowRows);
    setWorkflowId((current) => current || workflowRows[0]?.task_workflow_id || "");
    if (companyId) {
      setClients(await listCompanyClients(companyId, token));
    }
    setError("");
  };

  useEffect(() => {
    void load().catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить сотрудников"));
  }, [projectId, token, companyId]);

  const selectedWorkflow = useMemo(
    () => workflows.find((workflow) => workflow.task_workflow_id === workflowId) || workflows[0],
    [workflowId, workflows],
  );
  const roles = selectedWorkflow?.roles || [];
  const selectedRole = roles.find((role) => role.task_process_role_id === roleId) || roles[0];

  useEffect(() => {
    if (selectedWorkflow && !roles.some((role) => role.task_process_role_id === roleId)) {
      setRoleId(roles[0]?.task_process_role_id || "");
    }
  }, [roleId, roles, selectedWorkflow]);

  if (!projectId) return <WorkNoProject />;

  const invite = async () => {
    if (!token || !projectId || !selectedWorkflow || !selectedRole) return;
    const payload = await inviteWorkProjectMember(projectId, token, {
      email,
      display_name: displayName,
      task_workflow_id: selectedWorkflow.task_workflow_id,
      task_process_role_id: selectedRole.task_process_role_id,
      role: selectedRole.role_key,
      client_id: selectedRole.role_key === "client" ? clientId : "",
    });
    setInviteResult(payload.temporary_password ? `Временный пароль: ${payload.temporary_password}` : "Доступ обновлен для существующего пользователя");
    setEmail("");
    setDisplayName("");
    setClientId("");
    await load();
  };

  return (
    <WorkFrame
      access={access}
      title="Сотрудники"
      description="Приглашайте людей в конкретный процесс. Участник процесса не получает доступ к другим задачам и модулям проекта."
      actions={access?.can_manage_members ? <Button variant="secondary" onClick={() => setInviteOpen(!inviteOpen)}><UserPlus size={16} /> Пригласить</Button> : null}
    >
      {error ? <div className="form-error">{error}</div> : null}
      {inviteResult ? <div className="work-invite-result">{inviteResult}</div> : null}
      {inviteOpen ? (
        <form
          className="work-invite"
          onSubmit={(event) => {
            event.preventDefault();
            void invite().catch((err) => setError(err.message));
          }}
        >
          <Field label="Электронная почта"><Input value={email} onChange={(event) => setEmail(event.target.value)} /></Field>
          <Field label="Имя"><Input value={displayName} onChange={(event) => setDisplayName(event.target.value)} /></Field>
          <Field label="Процесс">
            <Select value={selectedWorkflow?.task_workflow_id || ""} onChange={(event) => setWorkflowId(event.target.value)}>
              {workflows.map((workflow) => <option key={workflow.task_workflow_id} value={workflow.task_workflow_id}>{workflow.name}</option>)}
            </Select>
          </Field>
          <Field label="Роль процесса">
            <Select value={selectedRole?.task_process_role_id || ""} onChange={(event) => setRoleId(event.target.value)}>
              {roles.map((role) => <option key={role.task_process_role_id} value={role.task_process_role_id}>{role.name}</option>)}
            </Select>
          </Field>
          {selectedRole?.role_key === "client" ? (
            <Field label="Клиент">
              <Select value={clientId} onChange={(event) => setClientId(event.target.value)}>
                <option value="">Выберите клиента</option>
                {clients.map((client) => <option key={client.client_id} value={client.client_id}>{client.name}</option>)}
              </Select>
            </Field>
          ) : null}
          <div className="form-actions"><Button type="submit">Отправить доступ</Button></div>
        </form>
      ) : null}
      <Table>
        <thead>
          <tr><th>Сотрудник</th><th>Доступ к процессам</th><th>Команда</th><th>Загрузка</th><th>В работе</th><th>Просрочено</th><th>Время</th></tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.employee_profile_id}>
              <td><strong>{row.full_name}</strong><br /><span className="muted">{row.email}</span></td>
              <td>
                <div className="work-access-chips">
                  {(row.process_accesses || []).length ? row.process_accesses?.map((item) => (
                    <Badge key={`${row.user_id}-${item.task_workflow_id}-${item.task_process_role_id}`} tone="muted">
                      {item.workflow_name}: {item.role_name}
                    </Badge>
                  )) : <Badge tone="muted">{projectRoleLabel(row.role)}</Badge>}
                </div>
              </td>
              <td>{row.department || "Без команды"}</td>
              <td>{Math.round((row.logged_week_minutes / 60 / Math.max(row.capacity_hours_per_week, 1)) * 100)}%</td>
              <td>{row.tasks_in_progress}</td>
              <td>{row.overdue_tasks}</td>
              <td>{minutesLabel(row.logged_week_minutes)}</td>
            </tr>
          ))}
        </tbody>
      </Table>
    </WorkFrame>
  );
}

function projectRoleLabel(role: string) {
  const labels: Record<string, string> = {
    owner: "Владелец",
    admin: "Администратор",
    operations_manager: "Операционный менеджер",
    project_manager: "Руководитель проекта",
    account_manager: "Аккаунт-менеджер",
    team_lead: "Тимлид",
    employee: "Сотрудник",
    contractor: "Подрядчик",
    client_guest: "Клиент",
    finance: "Финансы",
    hr: "HR",
    read_only_auditor: "Аудитор",
  };
  return labels[role] || "Участник";
}
