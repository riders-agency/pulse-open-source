import { chromium, expect } from '@playwright/test';
import fs from 'node:fs/promises';
import path from 'node:path';

// Uses a new isolated guest, real SSE responses and persisted task data. No production credentials.
const base = process.env.PULSE_BASE_URL || 'https://demo.example.test';
if (!['demo.example.test', 'localhost', '127.0.0.1'].includes(new URL(base).hostname)) throw new Error('Use a dedicated demo instance');
const dir = path.resolve(process.env.PULSE_ARTIFACT_DIR || '../../artifacts/tech/task-story');
await fs.mkdir(dir, { recursive: true });
const browser = await chromium.launch({ channel: 'chrome', headless: true });
const context = await browser.newContext({ locale: 'ru-RU', viewport: { width: 1280, height: 900 }, recordVideo: { dir: path.join(dir, 'raw'), size: { width: 1280, height: 900 } } });
const started = Date.now();
const page = await context.newPage();
page.setDefaultTimeout(30000);
const evidence = { mode: 'live-task-conversation', checks: [], errors: [], timings: {} };
const sessions = [];
const responses = [];
page.on('pageerror', error => evidence.errors.push(error.message));
page.on('response', response => {
  if (response.status() >= 400 && response.url().includes('/api/')) evidence.errors.push(`${response.status()} ${new URL(response.url()).pathname}`);
  if (response.url().endsWith('/messages/stream')) responses.push((async () => {
    const body = await response.text();
    for (const frame of body.split('\n\n')) {
      const data = frame.split(/\r?\n/).filter(line => line.startsWith('data:')).map(line => line.slice(5).trimStart()).join('\n');
      if (!data) continue;
      const payload = JSON.parse(data);
      if (payload.session) sessions.push(payload.session);
    }
  })());
});
const mark = name => { evidence.timings[name] = (Date.now() - started) / 1000; console.log(JSON.stringify({ phase: name, seconds: evidence.timings[name] })); };
const pause = ms => page.waitForTimeout(ms);
const sanitize = trace => ({
  status: trace.status,
  nodes: trace.nodes.map(({ node, status, latency_ms }) => ({ node, status, latency_ms })),
  model_calls: trace.model_calls.map(({ node, provider, model, input_tokens, output_tokens }) => ({ node, provider, model, input_tokens, output_tokens })),
  tools: trace.tools.map(({ tool, status, event }) => ({ tool, status, event })),
  input_tokens: trace.input_tokens, output_tokens: trace.output_tokens,
});
try {
  await page.goto(`${base}/demo?start=1`);
  await expect(page).toHaveURL(/\/tasks\/all$/, { timeout: 30000 });
  await page.getByTestId('ai-employee-launcher').click();
  await page.getByLabel('Открыть на весь экран').click();
  const input = page.getByTestId('ai-composer-input');
  await expect(input).toBeEnabled();
  await pause(800);
  mark('start');
  await input.pressSequentially('Запланируй задачу подготовить презентацию Pulse на завтра.', { delay: 28 });
  await pause(700);
  await page.getByTestId('ai-composer-send').click();
  mark('request');
  const proposal = page.getByTestId('ai-proposal-card-task-create');
  await expect(proposal).toBeVisible({ timeout: 90000 });
  await expect(proposal.getByLabel('Название', { exact: true })).toHaveValue('Подготовить презентацию Pulse');
  await expect(input).toBeEnabled({ timeout: 90000 });
  await Promise.all(responses);
  const proposedSession = sessions.at(-1);
  const pendingTrace = proposedSession.messages.map(item => item.payload_json?.execution).find(item => item?.status === 'waiting_approval');
  expect(pendingTrace).toBeTruthy();
  expect(pendingTrace.model_calls.length).toBeGreaterThan(0);
  expect(pendingTrace.input_tokens + pendingTrace.output_tokens).toBeGreaterThan(0);
  expect(proposedSession.tool_calls.find(call => call.tool_key === 'task.create').status).toBe('pending_approval');
  mark('proposal');
  await pause(1800);
  await input.pressSequentially('Да, создавай.', { delay: 65 });
  await pause(700);
  await page.getByTestId('ai-composer-send').click();
  mark('confirmation');
  await expect(page.getByTestId('ai-tool-result-card')).toContainText('Задача создана', { timeout: 90000 });
  await expect(input).toBeEnabled({ timeout: 90000 });
  await Promise.all(responses);
  const finalSession = sessions.at(-1);
  const call = finalSession.tool_calls.find(item => item.tool_key === 'task.create');
  expect(call.status).toBe('succeeded');
  const confirmationIndex = finalSession.messages.findIndex(item => item.role === 'user' && item.body === 'Да, создавай.');
  const resultIndex = finalSession.messages.findIndex(item => item.role === 'tool' && item.payload_json?.tool_call_id === call.ai_tool_call_id);
  expect(confirmationIndex).toBeGreaterThan(-1);
  expect(resultIndex).toBeGreaterThan(confirmationIndex);
  expect(finalSession.messages.filter(item => item.role === 'assistant').at(-1).body).toContain('Задача создана');
  expect(call.output_json.url).toMatch(/^\/tasks\/[0-9a-f-]+$/);
  const completeTrace = finalSession.messages.map(item => item.payload_json?.execution).filter(item => item?.status === 'completed').at(-1);
  expect(completeTrace).toBeTruthy();
  mark('result');
  await pause(3000);
  await page.screenshot({ path: path.join(dir, 'result.png') });
  const trace = page.locator('.execution-trace').last();
  await trace.locator(':scope > summary').click();
  await trace.locator('.agent-graph').scrollIntoViewIfNeeded();
  await pause(700);
  mark('graph');
  await page.screenshot({ path: path.join(dir, 'graph.png') });
  await pause(4000);
  mark('end');
  evidence.recording = { request: 'Запланируй задачу подготовить презентацию Pulse на завтра.', confirmation: 'Да, создавай.', task_title: 'Подготовить презентацию Pulse', tool: { name: call.tool_key, status: call.status }, proposal: sanitize(pendingTrace), completion: sanitize(completeTrace) };
  await trace.locator(':scope > summary').click();
  // Reload and open the actual entity; this is acceptance evidence, outside the presentation clip.
  await page.reload();
  await page.getByTestId('ai-employee-launcher').click();
  await page.getByLabel('Открыть на весь экран').click();
  await page.locator(`.ai-entity-link[href="${call.output_json.url}"]`).first().click();
  await expect(page.getByRole('heading', { name: 'Подготовить презентацию Pulse', exact: true })).toBeVisible();
  evidence.checks.push('Model usage comes from provider response', 'Task remains pending until natural-language confirmation', 'Tool succeeded and task survived reload');
  expect(evidence.errors).toEqual([]);
} catch (error) {
  evidence.failure = error.message;
  evidence.lastMessages = sessions.at(-1)?.messages.map(({ body, payload_json }) => ({ body, execution: payload_json?.execution }));
  await page.screenshot({ path: path.join(dir, 'error.png'), fullPage: true });
  process.exitCode = 1;
} finally {
  await context.close();
  await fs.copyFile(await page.video().path(), path.join(dir, 'ai-task.webm'));
  await fs.writeFile(path.join(dir, 'evidence.json'), JSON.stringify(evidence, null, 2));
  console.log(JSON.stringify(evidence));
  await browser.close();
}
