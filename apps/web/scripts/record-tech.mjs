import { chromium, expect } from '@playwright/test';
import fs from 'node:fs/promises';
import path from 'node:path';
import { recordWorkflow, recordChat, recordInvoice } from './tech-scenarios.mjs';
const base = process.env.PULSE_BASE_URL || 'http://127.0.0.1:19821';
if (!['127.0.0.1','localhost'].includes(new URL(base).hostname)) throw new Error('Recordings require an isolated local instance');
const project=process.env.PULSE_WORK_PROJECT_ID || '11111111-1111-4111-8111-111111111103';
const artifacts=path.resolve('../../artifacts/tech');
await fs.mkdir(artifacts,{recursive:true});
const browser=await chromium.launch({channel:'chrome',headless:true});
const auth=await browser.newContext({locale:'ru-RU'}), login=await auth.newPage();
await login.goto(`${base}/dashboard`);
await login.getByLabel('Электронная почта',{exact:true}).fill(process.env.PULSE_OWNER_EMAIL||'demo@riders-pulse.local');
await login.getByLabel('Пароль',{exact:true}).fill(process.env.PULSE_OWNER_PASSWORD);
await login.getByRole('button',{name:'Войти',exact:true}).click();
await login.getByLabel('Активная компания').waitFor();
const csrf=async(context)=>({'X-CSRF-Token':(await context.cookies()).find(c=>c.name==='rp_csrf')?.value||''});
await login.request.put(`${base}/api/v1/auth/workspace-context`,{headers:await csrf(auth),data:{active_company_id:process.env.PULSE_WORK_COMPANY_ID||'11111111-1111-4111-8111-111111111102',active_project_id:project}});
const state=await auth.storageState();await auth.close();
const context=await browser.newContext({storageState:state,viewport:{width:1280,height:800},locale:'ru-RU',recordVideo:{dir:path.join(artifacts,'raw'),size:{width:1280,height:800}}});
const page=await context.newPage();
const pause=(ms=1500)=>page.waitForTimeout(ms);
const scenario=process.argv[2]||'ai-task';
try {
  if(scenario==='ai-task') {
    await page.goto(`${base}/tasks/all`);
    await page.getByRole('button',{name:'ИИ сотрудник',exact:true}).click();
    await page.getByLabel('Открыть на весь экран',{exact:true}).click();
    await pause();
    const input=page.getByRole('textbox',{name:'Сообщение ИИ-сотруднику'});
    await input.pressSequentially('Создай задачу подготовить презентацию проекта на завтра',{delay:45});
    await pause(700); await input.press('Enter');
    const proposal=page.getByTestId('ai-proposal-card-task-create');
    await expect(proposal).toBeVisible({timeout:30000});
    await expect(proposal.getByLabel('Название',{exact:true})).toHaveValue('Подготовить презентацию проекта');
    await pause(1800);
    await proposal.locator('label').filter({hasText:'Приоритет'}).locator('select').selectOption('high');
    await proposal.getByText('Технические детали',{exact:true}).click(); await pause(2200);
    await proposal.getByText('Технические детали',{exact:true}).click();
    await page.screenshot({path:path.join(artifacts,'ai-task.png')});
    const approved=page.waitForResponse(r=>r.url().endsWith('/approve')&&r.request().method()==='POST');
    await proposal.getByRole('button',{name:'Создать задачу',exact:true}).click();
    const res=await approved, data=await res.json();
    const call=data.tool_calls.find(c=>c.tool_key==='task.create');
    expect(call.status,JSON.stringify(call)).toBe('succeeded');
    await pause(2000);
    await fs.writeFile(path.join(artifacts,'ai-task-evidence.json'),JSON.stringify({session:data.ai_chat_session_id,tool_call:call.ai_tool_call_id,status:call.status,output:call.output_json,recording_mode:'real-python-grpc-deterministic'},null,2));
    const link=call.output_json.url; expect(link).toMatch(/^\/tasks\//);
    await expect(proposal).toContainText('Высокий');
    await page.locator(`.ai-entity-link[href="${link}"]`).first().click();
    await expect(page.getByRole('heading',{name:'Подготовить презентацию проекта',exact:true})).toBeVisible();await pause(2500);
  } else {
    const record={workflow:recordWorkflow,chat:recordChat,invoice:recordInvoice}[scenario];
    if(!record) throw new Error('Unknown scenario');
    await record({page,base,project,pause,headers:await csrf(context)});
    await page.screenshot({path:path.join(artifacts,`${scenario}.png`)}); await pause(2500);
  }
  await context.close();
  await fs.copyFile(await page.video().path(),path.join(artifacts,`${scenario}.webm`));
  console.log('Recorded',scenario);
} catch(error){await page.screenshot({path:path.join(artifacts,`${scenario}-error.png`)});console.error(await page.locator('body').innerText());throw error;}
finally {await browser.close();}
