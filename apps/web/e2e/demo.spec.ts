import { expect, test, type Frame, type Page } from "@playwright/test";
const base = process.env.PULSE_BASE_URL || "http://127.0.0.1:19821";

async function demoSurface(page: Page): Promise<Page | Frame> {
  await expect.poll(async () => /\/tasks\/all$/.test(page.url()) || await page.locator('iframe[title="Демонстрационное пространство Pulse"]').count() > 0 || await page.getByRole("button", { name: "Открыть демо", exact: true }).count() > 0).toBe(true);
  const embed = page.locator('iframe[title="Демонстрационное пространство Pulse"]');
  if (await embed.count()) {
    await expect.poll(async () => {
      const box = await embed.boundingBox();
      return Math.abs((box?.width ?? 0) - (page.viewportSize()?.width ?? 0));
    }).toBeLessThanOrEqual(2);
    const handle = await embed.elementHandle();
    const frame = await handle?.contentFrame();
    if (!frame) throw new Error("Demo frame not attached");
    expect(new URL(page.url()).origin).toBe(new URL(base).origin);
    expect(new URL(page.url()).pathname).toBe("/demo");
    return frame;
  }
  return page;
}

test("demo creates isolated guest workspaces and blocks external operations", async ({ browser }) => {
  const first = await browser.newContext();
  const second = await browser.newContext();
  try {
    const page = await first.newPage();
    let registrations = 0;
    page.on("request", (request) => { if (request.method() === "POST" && request.url().endsWith("/api/v1/auth/register")) registrations++; });
    await page.goto(`${base}/dashboard`);
    await expect(page.getByRole("link", { name: "Сценарии и архитектура проекта" })).toBeVisible();
    await page.getByRole("link", { name: "Демо-вход", exact: true }).click();
    let app = await demoSurface(page);
    await expect(app.getByText("Демо-пространство", { exact: true })).toBeVisible();
    await expect(app.getByRole("link", { name: "Подготовить прототип сайта", exact: true })).toBeVisible();
    const project = await app.evaluate(() => sessionStorage.getItem("pulse-demo-context"));
    expect(project).toBeTruthy();
    expect(registrations).toBe(1);
    await page.goto(`${base}/demo?start=1`);
    app = await demoSurface(page);
    await expect(app.getByText("Демо-пространство", { exact: true })).toBeVisible();
    expect(await app.evaluate(() => sessionStorage.getItem("pulse-demo-context"))).toBe(project);
    expect(registrations).toBe(1);
    const task = await app.getByRole("link", { name: "Подготовить прототип сайта", exact: true }).getAttribute("href");
    const demoOrigin = new URL(app.url()).origin;
    const csrf = (await first.cookies(demoOrigin)).find((cookie) => cookie.name === "rp_csrf")?.value || "";
    const denied = await page.request.post(`${demoOrigin}/api/v1/projects/${project}/action-dependencies`, { headers: { "X-CSRF-Token": csrf }, data: {} });
    expect(denied.status()).toBe(403);
    const visitor = await second.newPage();
    await visitor.goto(`${base}/demo`);
    const visitorApp = await demoSurface(visitor);
    await visitorApp.getByRole("button", { name: "Открыть демо", exact: true }).click();
    await expect(visitorApp.getByText("Демо-пространство", { exact: true })).toBeVisible();
    expect(await visitorApp.evaluate(() => sessionStorage.getItem("pulse-demo-context"))).not.toBe(project);
    const forbidden = await visitor.request.get(`${demoOrigin}/api/v1/work/tasks/${task?.split("/").at(-1)}`);
    expect([403, 404]).toContain(forbidden.status());
    await page.setViewportSize({ width: 390, height: 844 });
    if (app !== page) expect((await page.locator('iframe').boundingBox())?.width).toBe(390);
    expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(2);
    expect(await app.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(2);
    await page.screenshot({ path: "../../artifacts/tech/demo-workspace-mobile.png" });
    if (app !== page) {
      page.once("dialog", (dialog) => void dialog.accept());
      expect(await app.evaluate(() => window.confirm("Подтвердить действие в демонстрационном пространстве?"))).toBe(true);
      expect(await page.evaluate(() => sessionStorage.getItem("pulse-demo-context"))).toBeNull();
      expect((await first.cookies(base)).filter((cookie) => ["rp_access", "rp_refresh"].includes(cookie.name))).toHaveLength(0);
      await expect(page).toHaveURL(`${base}/demo?start=1`);
    }
  } finally { await first.close(); await second.close(); }
});
