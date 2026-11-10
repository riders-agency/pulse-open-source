import { expect, test } from "@playwright/test";
const base = process.env.PULSE_BASE_URL || "http://127.0.0.1:19821";

test("public showcase starts with the task recording and reveals details on demand", async ({ page }) => {
  const response = await page.goto(`${base}/tech`);
  expect(response?.status()).toBe(200);
  await expect(page).toHaveTitle(/Pulse.*платформа для ИИ-агентов/);
  await expect(page.getByRole("heading", { level: 1 })).toContainText("Попросили в чате.");
  const video = page.getByLabel("Запись: создание задачи в чате");
  await expect(video).toBeVisible();
  expect(await video.getAttribute("autoplay")).toBeNull();
  expect(await video.getAttribute("controls")).toBeNull();
  expect(await page.locator("#technical-details").getAttribute("open")).toBeNull();
  const langGraphLogo = page.getByRole("img", { name: "LangGraph", exact: true });
  await expect(langGraphLogo).toBeVisible();
  await expect.poll(() => langGraphLogo.evaluate((element: HTMLImageElement) => element.naturalWidth)).toBeGreaterThan(0);
  await expect(page.locator("#agent-runtime .tech-graph-engine")).toContainText("Исполняет граф ИИ-агента");
  await expect(page.locator("#task-demo .tech-graph-engine")).toHaveCount(0);
  await expect(page.locator("#task-demo")).toContainText("платформа для ИИ-агентов");
  await expect(page.locator("#features article")).toHaveCount(12);
  await expect(page.getByRole("heading", { name: "SDK и фреймворки", exact: true })).toBeVisible();
  await expect(page.getByRole("tab", { name: /Сценарий/ })).toHaveAttribute("aria-selected", "true");
  expect(await page.locator("#agent-runtime").evaluate((element) => element.previousElementSibling?.id)).toBe("task-demo");
  await page.getByRole("button", { name: "Смотреть запись", exact: true }).click();
  await expect.poll(() => video.evaluate((element: HTMLVideoElement) => element.currentTime)).toBeGreaterThan(0.1);
  await expect(video).toHaveAttribute("controls", "");
  await video.evaluate((element: HTMLVideoElement) => element.pause());
  await expect(page.getByRole("button", { name: "Смотреть запись", exact: true })).toHaveCount(0);
  for (const width of [1440, 768, 390, 320]) {
    await page.setViewportSize({ width, height: 900 });
    await page.evaluate(() => scrollTo(0, 0));
    expect(await video.evaluate((element) => element.getBoundingClientRect().top)).toBeLessThan(460);
    expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(2);
  }
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page.reload();
  await expect.poll(() => video.evaluate((element: HTMLVideoElement) => element.readyState)).toBeGreaterThanOrEqual(2);
  await expect(page.getByRole("button", { name: "Смотреть запись", exact: true })).toBeVisible();
  await page.screenshot({ path: "../../artifacts/tech/showcase-desktop.png" });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.screenshot({ path: "../../artifacts/tech/showcase-mobile.png" });
  await page.locator("#technical-details > summary").click();
  await expect(page.locator("#engineering article")).toHaveCount(4);
  const decision = page.locator("#engineering article").first();
  await decision.locator("summary").focus();
  await page.keyboard.press("Enter");
  await expect(decision.getByRole("link", { name: "Открыть реализацию" })).toBeVisible();
  await expect(page.getByRole("tab")).toHaveCount(4);
  await page.getByRole("tab", { name: /ИИ → задача/ }).focus();
  await page.keyboard.press("ArrowRight");
  await expect(page.getByRole("tab", { name: /Сценарий/ })).toBeFocused();
  await page.keyboard.press("End");
  await expect(page.getByRole("tabpanel")).toContainText("Банки ещё не подключены");
  for (const id of ["ai-task", "workflow", "chat", "invoice", "office"]) {
    for (const extension of ["mp4", "webp", "vtt"]) expect((await page.request.get(`${base}/tech/recordings/${id}.${extension}`)).status()).toBe(200);
  }
  expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(2);
});

test("recorded graph distinguishes the approval pause from successful execution", async ({ page }) => {
  await page.goto(`${base}/tech#agent-runtime`);
  const graph = page.locator("#agent-runtime");
  await expect(graph.getByRole("heading", { name: "А вот как это выполнено" })).toBeVisible();
  await expect(graph.locator(".tech-run-receipt")).toContainText("Задача создана");
  await expect(graph.locator('[data-id="approval_decision"]')).toBeVisible();
  await expect(graph.getByLabel("Путь в графе")).toHaveValue("actual");
  await graph.getByRole("button", { name: "1. Подготовка и ожидание" }).click();
  await expect(graph.locator('[data-id="await_approval"]')).toHaveClass(/is-waiting/);
  await expect(graph.locator('[data-id="approval_decision"]')).toHaveCount(0);
  await expect(graph.locator('[data-id="finalizer"]')).toHaveCount(0);
  await graph.locator('[data-id="await_approval"]').click();
  await expect(graph.locator(".agent-node-detail")).toContainText("LangGraph interrupt");
  await graph.getByLabel("Путь в графе").selectOption("all");
  await expect(graph.locator(".react-flow__node")).toHaveCount(20);
  await expect(graph).toContainText("Показан возможный маршрут");
  await graph.getByLabel("Путь в графе").selectOption("actual");
  // A direct proposal_gate -> approval_decision shortcut is not part of the recorded proposal.
  await expect(graph.locator('[data-id="proposal_gate-approval_decision"]')).toHaveCount(0);
  await expect(graph.locator(".tech-run-usage")).toContainText("Входные токены");
  for (const width of [1440, 768, 390, 320]) {
    await page.setViewportSize({ width, height: 900 });
    expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(2);
  }
  await page.setViewportSize({ width: 1440, height: 1000 });
  await graph.screenshot({ path: "../../artifacts/tech/agent-runtime-desktop.png" });
});

test("old office link opens the company workflow section", async ({ request }) => {
  const response = await request.get(`${base}/tech/office`, { maxRedirects: 0 });
  expect(response.status()).toBe(307);
  expect(response.headers().location).toContain("/project-workflows?demo=1");
});
