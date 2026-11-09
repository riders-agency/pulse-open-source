import { expect, test, type Page } from "@playwright/test";

const baseURL = (process.env.PULSE_BASE_URL || "https://pulse.riders.agency").replace(/\/+$/, "");

test("expired saved session returns to a centered usable login and demo entry", async ({ page }) => {
  await page.addInitScript(() => {
    localStorage.setItem(
      "riders-pulse-ui",
      JSON.stringify({
        state: {
          token: "__riders_pulse_browser_session__",
          hydrated: true,
          activeCompanyId: "",
          activeProjectId: "",
          theme: "light",
          sidebarCollapsed: false,
        },
        version: 0,
      }),
    );
  });
  await page.route("**/api/v1/auth/me", async (route) => {
    await new Promise((resolve) => setTimeout(resolve, 500));
    await route.fulfill({
      status: 401,
      contentType: "application/problem+json",
      body: JSON.stringify({ title: "Нет доступа" }),
    });
  });

  await page.goto(`${baseURL}/dashboard`, { waitUntil: "domcontentloaded" });
  await expect(page.getByRole("button", { name: "Войти", exact: true })).toBeVisible({ timeout: 5_000 });
  await expect(page.getByRole("link", { name: "Демо-вход", exact: true })).toBeVisible();
  await expect(page.getByLabel("Электронная почта", { exact: true })).toBeEditable();
  await expectAuthCardCentered(page);
});

test("auth gate keeps invalid-login card centered in the viewport", async ({ page }) => {
  await page.route("**/api/v1/auth/login", async (route) => {
    await route.fulfill({
      status: 401,
      contentType: "application/problem+json",
      body: JSON.stringify({ title: "Неверный email или пароль" }),
    });
  });

  await page.goto(`${baseURL}/dashboard`, { waitUntil: "domcontentloaded" });
  await expect(page.getByRole("link", { name: "Демо-вход", exact: true })).toBeVisible();
  await page.screenshot({ path: "../../artifacts/tech/auth-demo-desktop.png", fullPage: true });
  await page.setViewportSize({ width: 390, height: 844 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(2);
  await expect(page.getByRole("link", { name: "Демо-вход", exact: true })).toBeInViewport();
  await page.screenshot({ path: "../../artifacts/tech/auth-demo-mobile.png", fullPage: true });
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.locator("input[autocomplete='email']").fill("wrong@example.test");
  await page.locator("input[type='password']").fill("bad-password");
  await page.getByRole("button", { name: "Войти" }).click();
  await expect(page.getByText("Неверный email или пароль.")).toBeVisible({ timeout: 5_000 });
  await expectAuthCardCentered(page);
});

test("login accepts input only after its handlers are ready", async ({ page }) => {
  let releaseScripts = () => {};
  const scriptsReady = new Promise<void>((resolve) => { releaseScripts = resolve; });
  await page.route("**/_next/static/**/*.js*", async (route) => { await scriptsReady; await route.continue(); });
  await page.route("**/api/v1/auth/login", (route) => route.fulfill({ status: 401, contentType: "application/problem+json", body: JSON.stringify({ title: "Неверный email или пароль" }) }));
  try {
    await page.goto(`${baseURL}/dashboard`, { waitUntil: "commit" });
    const email = page.locator("input[autocomplete='email']");
    await expect(email).toBeVisible();
    await expect(email).toBeDisabled();
    releaseScripts();
    await expect(email).toBeEnabled();
    await email.fill("wrong@example.test");
    await page.locator("input[type='password']").fill("bad-password");
    await page.getByRole("button", { name: "Войти", exact: true }).click();
    await expect(page.getByText("Неверный email или пароль.")).toBeVisible();
  } finally { releaseScripts(); }
});

async function expectAuthCardCentered(page: Page) {
  const box = await page.getByTestId("auth-card").boundingBox();
  const viewport = page.viewportSize();
  expect(box, "auth card box").not.toBeNull();
  expect(viewport, "viewport").not.toBeNull();
  if (!box || !viewport) return;

  const centerX = box.x + box.width / 2;
  const centerY = box.y + box.height / 2;
  expect(Math.abs(centerX - viewport.width / 2), "auth card horizontal center").toBeLessThanOrEqual(16);
  expect(Math.abs(centerY - viewport.height / 2), "auth card vertical center").toBeLessThanOrEqual(16);
}
