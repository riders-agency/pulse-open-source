import { apiGet, apiPost, apiPut, ApiError, BROWSER_SESSION_TOKEN } from "../../lib/api";
import { registerAccount } from "../auth/api/authApi";
import { usePulseStore } from "../../store/pulse-store";

type Workflow = { statuses: { task_workflow_status_id: string }[] };
const token = BROWSER_SESSION_TOKEN;
let pending: Promise<void> | null = null;

export function startDemo() {
  pending ??= createDemo().finally(() => { pending = null; });
  return pending;
}

async function createDemo() {
  try {
    const me = await apiGet<{ email?: string }>("/api/v1/auth/me", token);
    if (me.email?.endsWith("@demo.example.test")) {
      const current = await apiGet<{ active_company_id: string; active_project_id: string }>("/api/v1/auth/workspace-context", token);
      if (current.active_project_id) {
        await apiPost(`/api/v1/projects/${current.active_project_id}/office-demo`, token);
        const store = usePulseStore.getState();
        store.setToken(token); store.setActiveCompany(current.active_company_id); store.setActiveProject(current.active_project_id);
        sessionStorage.setItem("pulse-demo-context", current.active_project_id);
        return;
      }
    }
  } catch (error) {
    if (!(error instanceof ApiError) || ![401, 403, 404].includes(error.status)) throw error;
    sessionStorage.removeItem("pulse-demo-context");
  }
  const id = crypto.randomUUID();
  await registerAccount({ email: `guest-${id}@demo.example.test`, password: crypto.randomUUID() + crypto.randomUUID(), display_name: "Гость демо" });
  const state = usePulseStore.getState();
  state.clearWorkspace(); state.setToken(token);
  const company = await apiPost<{ company_id: string }>("/api/v1/companies", token, { name: "Демо-команда", slug: `demo-${id}` });
  const project = await apiPost<{ project_id: string }>(`/api/v1/companies/${company.company_id}/projects`, token, { name: "Запуск нового продукта", slug: `launch-${id}` });
  await apiPut("/api/v1/auth/workspace-context", token, { active_company_id: company.company_id, active_project_id: project.project_id });
  state.setActiveCompany(company.company_id); state.setActiveProject(project.project_id);
  await seedTasks(project.project_id);
  await apiPost(`/api/v1/projects/${project.project_id}/office-demo`, token);
  sessionStorage.setItem("pulse-demo-context", project.project_id);
}

async function seedTasks(projectId: string) {
  const workflow = await apiPost<Workflow>("/api/v1/work/task-workflows", token, { project_id: projectId, name: "Запуск продукта", status_names: ["Новые", "В работе", "Готово"], create_default_boards: true });
  const status = workflow.statuses[0]?.task_workflow_status_id;
  if (!status) throw new Error("Не удалось подготовить доску задач. Повторите вход в демо.");
  for (const title of ["Подготовить прототип сайта", "Согласовать план запуска", "Проверить чат поддержки"]) {
    await apiPost("/api/v1/work/tasks", token, { project_id: projectId, task_workflow_status_id: status, title, type: "task", priority: "medium" });
  }
}
