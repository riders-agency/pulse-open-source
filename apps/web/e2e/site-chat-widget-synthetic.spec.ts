import { expect, test } from "@playwright/test";
import { readFile } from "node:fs/promises";
import path from "node:path";

test("public widget keeps one root and captures live webvisor fallback events", async ({ page }) => {
  const widgetPath = path.resolve(process.cwd(), "../api/internal/modules/sitechat/transport/http/widget.js");
  const widget = await readFile(widgetPath, "utf8");
  expect(widget).not.toContain('messagePayload("Запись');
  expect(widget).toContain("call.recording");
  const batches: unknown[] = [];
  const presenceBodies: { visit_session_id?: string; conversation_id?: string; metadata?: { page_session_id?: string } }[] = [];

  await page.addInitScript(() => {
    window.localStorage.setItem("rp_live_assist_k_visit_session_id", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa");
    window.localStorage.setItem("rp_live_assist_k_conversation_id", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb");
    Object.defineProperty(navigator, "mediaDevices", {
      configurable: true,
      value: {
        getUserMedia: async () => {
          const AudioContextCtor = window.AudioContext || (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
          if (!AudioContextCtor) return new MediaStream();
          const ctx = new AudioContextCtor();
          const oscillator = ctx.createOscillator();
          const destination = ctx.createMediaStreamDestination();
          oscillator.connect(destination);
          oscillator.start();
          return destination.stream;
        },
      },
    });
  });

  await page.route("https://visitor.test/**", async (route) => {
    await route.fulfill({
      contentType: "text/html",
      body: `<!doctype html>
        <html>
          <head><meta charset="utf-8"><title>Visitor test</title></head>
          <body>
            <button id="menu" onclick="document.body.classList.toggle('menu-open');document.querySelector('#panel').hidden=!document.querySelector('#panel').hidden">Menu</button>
            <div style="height:1800px;padding-top:700px"><div id="panel" hidden>Opened panel</div></div>
            <script async src="https://pulse.test/site-chat/widget.js" data-site-chat-key="k"></script>
            <script async src="https://pulse.test/site-chat/widget.js" data-site-chat-key="k"></script>
          </body>
        </html>`,
    });
  });
  await page.route("https://pulse.test/site-chat/widget.js", async (route) => {
    await route.fulfill({ contentType: "application/javascript", body: widget });
  });
  await page.route("https://pulse.test/api/v1/public/site-chat/widgets/k/config", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({ title: "Test", subtitle: "", accent_color: "#00B2FF", position: "right", settings_json: {} }),
    });
  });
  await page.route("https://pulse.test/api/v1/public/site-chat/presence", async (route) => {
    presenceBodies.push(route.request().postDataJSON());
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        visitor_id: "11111111-1111-1111-1111-111111111111",
        visit_session_id: "22222222-2222-2222-2222-222222222222",
        chat_widget_id: "33333333-3333-3333-3333-333333333333",
        project_id: "44444444-4444-4444-4444-444444444444",
        conversation_id: "55555555-5555-5555-5555-555555555555",
        monitoring_status: "active",
        recorder: { enabled: true, mode: "snapshot" },
        visitor_state_snapshot: {},
      }),
    });
  });
  await page.route("https://pulse.test/api/v1/public/site-chat/conversations", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({ conversation_id: "55555555-5555-5555-5555-555555555555" }),
    });
  });
  await page.route("https://pulse.test/api/v1/public/site-chat/conversations/**/messages?**", async (route) => {
    await route.fulfill({ contentType: "application/json", body: JSON.stringify([]) });
  });
  await page.route("https://pulse.test/api/v1/public/site-chat/calls/request", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        call_session_id: "66666666-6666-6666-6666-666666666666",
        conversation_id: "55555555-5555-5555-5555-555555555555",
        visitor_id: "11111111-1111-1111-1111-111111111111",
        visit_session_id: "22222222-2222-2222-2222-222222222222",
        media_kind: "audio",
        direction: "visitor_to_operator",
        status: "ringing",
        recording_enabled: true,
      }),
    });
  });
  await page.route("https://pulse.test/api/v1/public/site-chat/events", async (route) => {
    batches.push(route.request().postDataJSON());
    await route.fulfill({ status: 202, contentType: "application/json", body: JSON.stringify({ status: "accepted" }) });
  });

  await page.goto("https://visitor.test/?rp_e2e=synthetic", { waitUntil: "domcontentloaded" });
  await page.waitForFunction(() => document.querySelectorAll(".rp-site-chat").length === 1 && document.querySelectorAll('[id^="riders-site-chat-root-"]').length === 1);
  await page.waitForFunction(() => {
    const root = document.querySelector(".rp-site-chat") as HTMLElement | null;
    const monitor = root?.shadowRoot?.querySelector("[data-monitor]") as HTMLElement | null;
    return Boolean(monitor && !monitor.hidden);
  });

  await page.evaluate(() => window.scrollTo(0, 760));
  await page.mouse.move(240, 180);
  await page.mouse.click(240, 180);
  await page.evaluate(() => {
    const root = document.querySelector(".rp-site-chat") as HTMLElement | null;
    const button = root?.shadowRoot?.querySelector("[data-open]") as HTMLElement | null;
    button?.click();
  });
  await page.evaluate(() => {
    const root = document.querySelector(".rp-site-chat") as HTMLElement | null;
    const privacy = root?.shadowRoot?.querySelector("[data-privacy]") as HTMLInputElement | null;
    const call = root?.shadowRoot?.querySelector("[data-call-open]") as HTMLElement | null;
    if (privacy) {
      privacy.checked = true;
      privacy.dispatchEvent(new Event("change", { bubbles: true }));
    }
    call?.click();
  });
  await page.waitForFunction(() => {
    const root = document.querySelector(".rp-site-chat") as HTMLElement | null;
    const callPanel = root?.shadowRoot?.querySelector("[data-call]") as HTMLElement | null;
    const playAudio = root?.shadowRoot?.querySelector("[data-play-audio]") as HTMLElement | null;
    return Boolean(callPanel && !callPanel.hidden && playAudio);
  });
  await page.waitForTimeout(5_200);

  const pageState = await page.evaluate(() => {
    const root = document.querySelector(".rp-site-chat") as HTMLElement | null;
    const panel = root?.shadowRoot?.querySelector("[data-panel]") as HTMLElement | null;
    return {
      classRoots: document.querySelectorAll(".rp-site-chat").length,
      idRoots: document.querySelectorAll('[id^="riders-site-chat-root-"]').length,
      panelOpen: Boolean(panel && !panel.hidden),
      callHasAudioButton: Boolean(root?.shadowRoot?.querySelector("[data-play-audio]")),
      callMuteButtons: root?.shadowRoot?.querySelectorAll("[data-mute],[data-call-mute],.rp-call-mute").length ?? 0,
      callText: root?.shadowRoot?.querySelector("[data-call]")?.textContent ?? "",
      scrollY: Math.round(window.scrollY || 0),
    };
  });
  const events = batches.flatMap((batch) => {
    const value = batch as { events?: { event_type: string; payload?: Record<string, unknown> }[] };
    return Array.isArray(value.events) ? value.events : [];
  });
  const latestScroll = events.filter((event) => event.event_type === "scroll").at(-1);
  const latestClick = events.filter((event) => event.event_type === "click").at(-1);
  const widgetState = events.filter((event) => event.event_type === "widget_state").at(-1);
  const snapshot = events.find((event) => event.event_type === "monitoring.snapshot");
  const pageSessionIds = new Set(presenceBodies.map((body) => body.metadata?.page_session_id).filter(Boolean));

  expect(pageState.classRoots).toBe(1);
  expect(pageState.idRoots).toBe(1);
  expect(presenceBodies.length).toBeGreaterThanOrEqual(1);
  expect(presenceBodies[0]?.visit_session_id || "").not.toBe("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa");
  expect(presenceBodies[0]?.conversation_id || "").not.toBe("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb");
  expect(pageSessionIds.size).toBe(1);
  expect(Array.from(pageSessionIds)[0]).toMatch(/^page_/);
  expect(snapshot?.payload?.mirror).toBeTruthy();
  expect(Math.abs(Number(latestScroll?.payload?.scroll_y ?? 0) - pageState.scrollY)).toBeLessThanOrEqual(120);
  expect(Math.abs(Number(latestClick?.payload?.x ?? 0) - 240)).toBeLessThanOrEqual(80);
  expect(widgetState?.payload?.open).toBe(true);
  expect(pageState.panelOpen).toBe(true);
  expect(pageState.callHasAudioButton).toBe(true);
  expect(pageState.callMuteButtons).toBe(0);
  expect(pageState.callText).not.toContain("Без звука");
});
