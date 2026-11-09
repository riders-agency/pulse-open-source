import { expect, test } from "@playwright/test";
import { readFile } from "node:fs/promises";
import path from "node:path";

test("visitor widget keeps failed text, supports Shift+Enter and fits a phone", async ({ page }) => {
  const widget = await readFile(path.resolve(process.cwd(), "../api/internal/modules/sitechat/transport/http/widget.js"), "utf8");
  let sends = 0;
  const messages: object[] = [];
  const identity = { visitor_id: "11111111-1111-4111-8111-111111111111", visit_session_id: "22222222-2222-4222-8222-222222222222", conversation_id: "33333333-3333-4333-8333-333333333333", project_id: "44444444-4444-4444-8444-444444444444", monitoring_status: "off", recorder: { enabled: false }, visitor_state_snapshot: {} };
  await page.route("https://visitor.test/**", (route) => route.fulfill({ contentType: "text/html", body: '<!doctype html><html lang="ru"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Демонстрация чата</title><body><h1>Демо-проект</h1><script src="https://pulse.test/site-chat/widget.js" data-site-chat-key="retry-demo"></script></body></html>' }));
  await page.route("https://pulse.test/site-chat/widget.js", (route) => route.fulfill({ contentType: "application/javascript; charset=utf-8", body: widget }));
  await page.route("https://pulse.test/api/v1/public/site-chat/**", (route) => {
    const url = new URL(route.request().url());
    if (url.pathname.endsWith("/config")) return route.fulfill({ json: { title: "Команда проекта", subtitle: "Поможем с вашим вопросом", accent_color: "#ef6500", position: "right", settings_json: {} } });
    if (url.pathname.endsWith("/messages") && route.request().method() === "POST") {
      sends++;
      if (sends === 1) return route.fulfill({ status: 503, json: { detail: "temporary failure" } });
      const body = route.request().postDataJSON().body;
      messages.push({ message_id: "saved", author_kind: "visitor", body, created_at: new Date().toISOString() });
      return route.fulfill({ json: { message_id: "saved" } });
    }
    if (url.pathname.endsWith("/messages")) return route.fulfill({ json: messages });
    return route.fulfill({ json: identity });
  });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto("https://visitor.test");
  await page.locator("[data-open]").click();
  await page.locator("[data-privacy]").check();
  const composer = page.getByRole("textbox", { name: "Сообщение", exact: true });
  await composer.fill("Здравствуйте!");
  await composer.press("Shift+Enter");
  await composer.press("A");
  await expect(composer).toHaveValue("Здравствуйте!\nA");
  await composer.fill("Здравствуйте!\nХочу подключить сервис.");
  await page.getByRole("button", { name: "Отправить", exact: true }).click();
  await expect(page.getByRole("alert")).toContainText("Текст сохранён");
  await expect(composer).toHaveValue("Здравствуйте!\nХочу подключить сервис.");
  await page.getByRole("button", { name: "Отправить", exact: true }).click();
  await expect(composer).toHaveValue("");
  await expect(page.getByRole("log")).toContainText("Хочу подключить сервис.");
  expect(sends).toBe(2);
  const box = await composer.boundingBox();
  expect(box!.x).toBeGreaterThanOrEqual(0);
  expect(box!.y + box!.height).toBeLessThanOrEqual(844);
  await page.screenshot({ path: "../../artifacts/product-polish/visitor-chat-mobile.png" });
  await page.getByRole("button", { name: "Закрыть", exact: true }).click();
  await expect(page.locator("[data-open]")).toBeFocused();
});
