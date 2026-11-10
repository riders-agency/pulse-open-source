import { expect, test, type Page } from "@playwright/test";

const base = process.env.PULSE_BASE_URL || "http://127.0.0.1:19821";
const project = process.env.PULSE_WORK_PROJECT_ID || "11111111-1111-4111-8111-111111111103";
const company = process.env.PULSE_WORK_COMPANY_ID || "11111111-1111-4111-8111-111111111102";
const email = process.env.PULSE_OWNER_EMAIL || "demo@riders-pulse.local";
const password = process.env.PULSE_OWNER_PASSWORD;
if (!password) throw new Error("Set PULSE_OWNER_PASSWORD for the isolated test account");
if (!["127.0.0.1", "localhost"].includes(new URL(base).hostname)) throw new Error("These mutation tests require a local isolated instance");

async function csrf(page: Page) { return { "X-CSRF-Token": (await page.context().cookies()).find((cookie) => cookie.name === "rp_csrf")?.value || "" }; }

async function login(page: Page) {
  await page.goto(`${base}/dashboard`);
  await page.getByLabel("Электронная почта", { exact: true }).fill(email);
  await page.getByLabel("Пароль", { exact: true }).fill(password!);
  await page.getByRole("button", { name: "Войти", exact: true }).click();
  await expect(page.getByLabel("Активная компания")).toBeVisible();
  const context = await page.request.put(`${base}/api/v1/auth/workspace-context`, { headers: await csrf(page), data: { active_company_id: company, active_project_id: project } });
  expect(context.ok(), await context.text()).toBeTruthy();
}

function visitor(name: string, suffix: string) {
  return { visit_session_id: `22222222-2222-4222-8222-${suffix}`, conversation_id: `33333333-3333-4333-8333-${suffix}`, visitor_id: suffix, visitor_name: name, current_url: "https://example.test/services", current_title: "Услуги", status: "active", started_at: new Date().toISOString(), last_activity_at: new Date().toISOString(), duration_seconds: 125, visitor_message_count: 1, visitor_state_snapshot: { activity: { connection_status: "online" }, signals: { chat_open: true } } };
}
const ksenia = visitor("Ксения", "000000000001");
const denis = visitor("Денис", "000000000002");

async function chatFixtures(page: Page) {
  await page.route("**/site-visitors/active", (route) => route.fulfill({ json: [ksenia, denis] }));
  await page.route("**/site-chat-widgets", (route) => route.fulfill({ json: [] }));
  await page.route("**/site-visit-sessions/*/events?*", (route) => route.fulfill({ json: [] }));
  const messages = new Map([[ksenia.conversation_id, [{ message_id: "a", author_kind: "visitor", author_name: "Ксения", body: "Здравствуйте! Подскажите, как начать работу?", created_at: new Date().toISOString() }]], [denis.conversation_id, [{ message_id: "b", author_kind: "visitor", author_name: "Денис", body: "Расскажите о возможностях сервиса", created_at: new Date().toISOString() }]]]);
  let failures = 1;
  let posts = 0;
  await page.route("**/site-chat-conversations/*/messages", async (route) => {
    const cid = route.request().url().split("/").at(-2)!;
    if (route.request().method() === "POST") {
      posts++;
      if (failures-- > 0) return route.fulfill({ status: 503, contentType: "application/problem+json", body: JSON.stringify({ title: "Временная ошибка", detail: "Не удалось отправить сообщение. Повторите попытку." }) });
      const row = { message_id: `sent-${posts}`, author_kind: "operator", author_name: "Оператор", body: route.request().postDataJSON().body, created_at: new Date().toISOString() };
      messages.set(cid, [...messages.get(cid)!, row]);
      return route.fulfill({ json: row });
    }
    return route.fulfill({ json: messages.get(cid) || [] });
  });
  return { posts: () => posts };
}

async function noHorizontalOverflow(page: Page) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(2);
}

test("chat preserves failed drafts and keeps drafts separate between conversations", async ({ page }) => {
  await login(page);
  const fixture = await chatFixtures(page);
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page.goto(`${base}/site-chat`);
  const composer = page.getByRole("textbox", { name: "Сообщение посетителю" });
  await expect(page.getByRole("log")).toContainText("как начать работу");
  await composer.fill("Здравствуйте!\nПоможем с настройкой.");
  await page.getByRole("button", { name: "Отправить", exact: true }).click();
  await expect(composer).toHaveValue("Здравствуйте!\nПоможем с настройкой.");
  await expect(page.getByText("Не удалось отправить сообщение. Повторите попытку.", { exact: true }).first()).toBeVisible();
  await page.getByRole("button", { name: /^Денис/ }).click();
  await expect(composer).toHaveValue("");
  await composer.fill("Черновик для Дениса");
  await page.getByRole("button", { name: /^Ксения/ }).click();
  await expect(composer).toHaveValue("Здравствуйте!\nПоможем с настройкой.");
  await composer.press("Enter");
  await expect(composer).toHaveValue("");
  await expect(page.getByRole("log")).toContainText("Поможем с настройкой.");
  expect(fixture.posts()).toBe(2);
  await noHorizontalOverflow(page);
  await page.screenshot({ path: "../../artifacts/product-polish/chat-desktop.png" });
  await page.getByRole("button", { name: /^Денис/ }).click();
  await expect(composer).toHaveValue("Черновик для Дениса");
});

test("mobile chat has a usable composer and returns to the conversation list", async ({ page }) => {
  await login(page); await chatFixtures(page);
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto(`${base}/site-chat`);
  await page.getByRole("button", { name: /^Ксения/ }).click();
  await expect(page.getByRole("textbox", { name: "Сообщение посетителю" })).toBeVisible();
  await expect(page.getByRole("tab", { name: "Переписка" })).toBeVisible();
  await noHorizontalOverflow(page);
  await page.screenshot({ path: "../../artifacts/product-polish/chat-mobile.png" });
  const composer = await page.getByRole("textbox", { name: "Сообщение посетителю" }).boundingBox();
  expect(composer!.height).toBeGreaterThan(40);
  expect(composer!.y + composer!.height).toBeLessThanOrEqual(844);
  await page.getByRole("button", { name: "Открыть меню" }).click();
  await expect(page.getByRole("dialog")).toContainText("Рабочее пространство");
  await page.keyboard.press("Escape");
  await expect(page.getByRole("button", { name: "Открыть меню" })).toBeFocused();
  await page.getByRole("button", { name: "Все диалоги", exact: true }).click();
  await expect(page.getByRole("button", { name: /^Денис/ })).toBeVisible();
});

test("tasks can be deleted, hidden from the active list and restored", async ({ page }) => {
  await login(page);
  const title = `Проверить удаление ${Date.now()}`;
  const workflows = await (await page.request.get(`${base}/api/v1/work/task-workflows?project_id=${project}`)).json();
  let workflow = workflows.items[0];
  if (!workflow) {
    const result = await page.request.post(`${base}/api/v1/work/task-workflows`, { headers: await csrf(page), data: { project_id: project, name: "Рабочий процесс", status_names: ["Новые", "В работе", "Готово"], create_default_boards: true } });
    expect(result.ok(), await result.text()).toBeTruthy();
    workflow = await result.json();
  }
  const status = workflow.statuses[0].task_workflow_status_id;
  const created = await page.request.post(`${base}/api/v1/work/tasks`, { headers: await csrf(page), data: { project_id: project, task_workflow_status_id: status, title, type: "task", priority: "medium" } });
  expect(created.ok(), await created.text()).toBeTruthy();
  const task = await created.json();
  await page.goto(`${base}/tasks/${task.task_id}`);
  await page.getByRole("button", { name: "Удалить задачу", exact: true }).click();
  await page.getByRole("button", { name: "Удалить", exact: true }).click();
  await expect(page.getByRole("status")).toContainText("Задача удалена");
  await expect(page.getByRole("button", { name: "Удалить задачу", exact: true })).toHaveCount(0);
  await page.goto(`${base}/tasks/all`);
  await expect(page.getByRole("link", { name: title, exact: true })).toHaveCount(0);
  await page.getByLabel("Фильтр задач").selectOption("archived");
  const row = page.getByRole("region", { name: "Удалённые задачи" }).locator(".between").filter({ hasText: title });
  await expect(row).toBeVisible();
  await row.getByRole("button", { name: "Восстановить" }).click();
  await expect(row).toHaveCount(0);
  await page.getByLabel("Фильтр задач").selectOption("all");
  await expect(page.getByRole("link", { name: title, exact: true })).toBeVisible();
});

test("disabled action survives saving and reopening the workflow editor", async ({ page }) => {
  await login(page);
  const created = await page.request.post(`${base}/api/v1/projects/${project}/workflows`, { headers: await csrf(page), data: { name: "Проверка отключения шага", slug: `qa-disable-${Date.now()}` } });
  expect(created.ok(), await created.text()).toBeTruthy();
  const workflow = await created.json();
  const nodes = [
    { action_key: "start", action_type_key: "manual.trigger", label: "Ручной запуск", config: {} },
    { action_key: "mail", action_type_key: "email.send", label: "Письмо клиенту", config: { disabled: true } },
    { action_key: "finish", action_type_key: "stop", label: "Завершение", config: {} },
  ];
  const edges = [{ from: "start", to: "mail" }, { from: "mail", to: "finish" }];
  const saved = await page.request.put(`${base}/api/v1/workflows/${workflow.workflow_id}/draft`, { headers: await csrf(page), data: { draft_definition_json: { nodes, edges }, draft_graph_json: { nodes: nodes.map((node, index) => ({ id: node.action_key, type: "action", position: { x: 50 + index * 320, y: 150 }, data: { label: node.label, action_type_key: node.action_type_key, config: node.config } })), edges: edges.map((edge, index) => ({ id: `edge-${index}`, source: edge.from, target: edge.to })) }, draft_dsl_yaml: "", test_input_json: {} } });
  expect(saved.ok(), await saved.text()).toBeTruthy();
  await page.goto(`${base}/workflows/${workflow.workflow_id}/edit?action=mail`);
  await page.locator('.react-flow__node[data-id="mail"]').click();
  await page.getByRole("button", { name: "Включить действие", exact: true }).click();
  await expect(page.getByRole("button", { name: "Отключить действие", exact: true })).toBeVisible();
  await page.getByRole("button", { name: "Отключить действие", exact: true }).click();
  const response = page.waitForResponse((res) => res.url().endsWith(`/workflows/${workflow.workflow_id}/draft`) && res.request().method() === "PUT");
  await page.getByRole("button", { name: "Сохранить", exact: true }).click();
  expect((await response).ok()).toBeTruthy();
  await page.reload();
  await page.locator('.react-flow__node[data-id="mail"]').click();
  await expect(page.getByRole("button", { name: "Включить действие", exact: true })).toBeVisible();
  await expect(page.locator('.react-flow__node[data-id="mail"]')).toContainText("Отключено");
  const persisted = await (await page.request.get(`${base}/api/v1/workflows/${workflow.workflow_id}`)).json();
  expect(persisted.draft_definition_json.nodes.find((node: { action_key: string }) => node.action_key === "mail").config.disabled).toBe(true);
  await page.screenshot({ path: "../../artifacts/product-polish/workflow-disabled.png" });
});

test("AI chat supports keyboard focus and keeps a failed message for retry", async ({ page }) => {
  await login(page);
  await page.route("**/ai/employee", (route) => route.fulfill({ json: { employee: {}, configured_dependencies: [], setup_required: false, capability: { state: "ready" } } }));
  await page.route("**/ai/employee/sessions", (route) => route.fulfill({ json: { ai_chat_session_id: "44444444-4444-4444-8444-444444444444", project_id: project, messages: [], tool_calls: [] } }));
  await page.route("**/messages/stream", (route) => route.fulfill({ contentType: "text/event-stream", body: 'data: {"type":"ERROR","message":"Не удалось получить ответ. Повторите попытку."}\n\n' }));
  await page.goto(`${base}/dashboard`);
  const launcher = page.getByRole("button", { name: "ИИ сотрудник", exact: true });
  await launcher.click();
  const composer = page.getByRole("textbox", { name: "Сообщение ИИ-сотруднику" });
  await expect(composer).toBeFocused();
  await composer.fill("Помоги спланировать день");
  await composer.press("Shift+Enter");
  await expect(composer).toHaveValue("Помоги спланировать день\n");
  await composer.press("Enter");
  await expect(page.getByTestId("ai-error-banner")).toContainText("Повторите попытку");
  await expect(composer).toHaveValue("Помоги спланировать день");
  await page.setViewportSize({ width: 390, height: 844 });
  await noHorizontalOverflow(page);
  const bounds = await composer.boundingBox();
  expect(bounds!.y + bounds!.height).toBeLessThanOrEqual(844);
  await page.screenshot({ path: "../../artifacts/product-polish/ai-chat-mobile.png" });
  await composer.press("Escape");
  await expect(page.getByRole("dialog", { name: "ИИ сотрудник проекта" })).toHaveCount(0);
  await expect(launcher).toBeFocused();
});

test("workflow proposals show graph changes and never task creation fields", async ({ page }) => {
  await login(page);
  const callId = "55555555-5555-4555-8555-555555555555";
  const input = { workflow_id: "66666666-6666-4666-8666-666666666666", summary: "Добавить письмо после заявки", patch_json: { operations: [{ op: "add_node", action_key: "mail", action_type_key: "email.send" }] } };
  const session = { ai_chat_session_id: "44444444-4444-4444-8444-444444444444", project_id: project, messages: [{ role: "assistant", body: "Подготовил изменения сценария", payload_json: { kind: "proposal", ai_tool_call_id: callId } }], tool_calls: [{ ai_tool_call_id: callId, tool_key: "workflow.patch.apply", status: "pending_approval", input_json: input }] };
  await page.route("**/ai/employee", (route) => route.fulfill({ json: { employee: {}, setup_required: false, capability: { state: "ready" } } }));
  await page.route("**/ai/employee/sessions", (route) => route.fulfill({ json: session }));
  await page.route("**/messages/stream", (route) => route.fulfill({ contentType: "text/event-stream", body: `data: ${JSON.stringify({ type: "MESSAGE_FINAL", session })}\n\ndata: {"type":"DONE","ok":true}\n\n` }));
  await page.route("**/ai/tool-calls/*", async (route) => {
    expect(route.request().postDataJSON().input_json).toEqual(input);
    return route.fulfill({ json: session });
  });
  await page.route("**/ai/tool-calls/*/approve", (route) => route.fulfill({ json: { ...session, tool_calls: [{ ...session.tool_calls[0], status: "succeeded" }] } }));
  await page.goto(`${base}/dashboard`);
  await page.getByRole("button", { name: "ИИ сотрудник", exact: true }).click();
  const composer = page.getByRole("textbox", { name: "Сообщение ИИ-сотруднику" });
  await composer.fill("Добавь письмо в сценарий"); await composer.press("Enter");
  const card = page.getByTestId("ai-proposal-card");
  await expect(card).toContainText("Добавить действие");
  await expect(card).toContainText("email.send");
  await expect(card.getByLabel("Название", { exact: true })).toHaveCount(0);
  await expect(card).toContainText("сохранятся в черновик");
  await card.getByRole("button", { name: "Применить изменения", exact: true }).click();
  await expect(page.locator(".ai-notice")).toContainText("Действие выполнено");
});
