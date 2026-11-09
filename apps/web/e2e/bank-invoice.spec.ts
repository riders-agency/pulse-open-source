import { expect, test, type Page } from "@playwright/test";

const base = process.env.PULSE_BASE_URL || "http://127.0.0.1:19821";
const action = process.env.PULSE_PAYMENT_ACTION_ID;
const password = process.env.PULSE_OWNER_PASSWORD;
if (!["127.0.0.1", "localhost"].includes(new URL(base).hostname)) throw new Error("Invoice mutation tests require an isolated local instance");
test.skip(!action || !password, "Set PULSE_PAYMENT_ACTION_ID and PULSE_OWNER_PASSWORD for synthetic data");
test.use({ locale: "ru-RU" });
async function csrf(page: Page) { return { "X-CSRF-Token": (await page.context().cookies()).find((cookie) => cookie.name === "rp_csrf")?.value || "" }; }
async function login(page: Page) {
  await page.goto(`${base}/dashboard`);
  await page.getByLabel("Электронная почта", { exact: true }).fill(process.env.PULSE_OWNER_EMAIL || "demo@riders-pulse.local");
  await page.getByLabel("Пароль", { exact: true }).fill(password!);
  await page.getByRole("button", { name: "Войти", exact: true }).click();
  await expect(page.getByLabel("Активная компания")).toBeVisible();
  const result = await page.request.put(`${base}/api/v1/auth/workspace-context`, { headers: await csrf(page), data: {
    active_company_id: process.env.PULSE_WORK_COMPANY_ID || "11111111-1111-4111-8111-111111111102",
    active_project_id: process.env.PULSE_WORK_PROJECT_ID || "11111111-1111-4111-8111-111111111103",
  } });
  expect(result.ok(), await result.text()).toBeTruthy();
}

test("invoice draft persists, survives a lost response, and never issues in a bank", async ({ page }) => {
  await login(page);
  const endpoint = `${base}/api/v1/payment-actions/${action}/bank-invoice`;
  const before = await (await page.request.get(endpoint)).json();
  await page.goto(`${base}/payment-actions/${action}`);
  await expect(page.getByText("Банк ещё не подключён.", { exact: false })).toBeVisible();
  const amount = (Number(before.amount_minor) / 100).toLocaleString("ru-RU", { style: "currency", currency: "RUB", minimumFractionDigits: 0, maximumFractionDigits: 2 });
  await expect(page.locator(".payment-focus-card--invoice strong")).toHaveText(amount);
  const purpose = `Услуги дизайна ${Date.now()}`;
  await page.getByLabel("Плательщик", { exact: true }).fill("ООО Демо");
  await page.getByLabel("Назначение платежа", { exact: true }).fill(purpose);
  await page.getByLabel("Что подготовить", { exact: true }).selectOption("payment_link");
  await page.getByLabel("Оплатить до", { exact: true }).fill("31.12.2026");
  let writes = 0; const keys: string[] = [];
  await page.route("**/payment-actions/*/bank-invoice", async (route) => {
    if (route.request().method() !== "PUT") return route.continue();
    writes++; keys.push(route.request().headers()["idempotency-key"]);
    const response = await route.fetch();
    if (writes === 1) return route.fulfill({ status: 503, contentType: "application/problem+json", body: JSON.stringify({ title: "Ответ потерян", detail: "Повторите сохранение" }) });
    return route.fulfill({ response });
  });
  await page.getByRole("button", { name: "Сохранить черновик", exact: true }).click();
  await expect(page.getByRole("alert").filter({ hasText: "Повторите сохранение" })).toBeVisible();
  await expect(page.getByLabel("Назначение платежа", { exact: true })).toHaveValue(purpose);
  await page.getByRole("button", { name: "Сохранить черновик", exact: true }).click();
  await expect(page.getByRole("status")).toContainText("Черновик сохранён");
  expect(keys).toHaveLength(2); expect(keys[0]).toBe(keys[1]);
  const after = await (await page.request.get(endpoint)).json();
  expect(after.draft.revision).toBe((before.draft?.revision || 0) + 1);
  expect(after.payment_status).toBe(before.payment_status);
  expect(after.draft.amount_minor).toBe(before.amount_minor);
  expect(after.draft.due_date).toBe("2026-12-31");
  expect(after.providers).toEqual([]);
  await page.reload();
  await expect(page.getByLabel("Назначение платежа", { exact: true })).toHaveValue(purpose);
  await expect(page.getByLabel("Оплатить до", { exact: true })).toHaveValue("31.12.2026");
  await expect(page.getByRole("button", { name: "Выпустить в банке", exact: true })).toBeDisabled();
  await page.screenshot({ path: "../../artifacts/product-polish/bank-invoice-desktop.png", fullPage: true });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.getByLabel("Назначение платежа", { exact: true }).scrollIntoViewIfNeeded();
  expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(2);
  await page.screenshot({ path: "../../artifacts/product-polish/bank-invoice-mobile.png", fullPage: true });
});
