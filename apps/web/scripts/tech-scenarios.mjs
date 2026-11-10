import { expect } from '@playwright/test';
export async function recordInvoice({page,base,pause}) {
  const action=process.env.PULSE_PAYMENT_ACTION_ID;
  if(!action) throw new Error('Set PULSE_PAYMENT_ACTION_ID');
  await page.goto(`${base}/payment-actions/${action}`);
  await page.getByLabel('Плательщик',{exact:true}).scrollIntoViewIfNeeded();await pause();
  await page.getByLabel('Что подготовить',{exact:true}).selectOption('bank_invoice');
  await page.getByLabel('Плательщик',{exact:true}).fill('ООО «Демо»');
  await page.getByLabel('Назначение платежа',{exact:true}).fill('Разработка интерфейса и интеграций');
  await page.getByLabel('Оплатить до',{exact:true}).fill('30.11.2026');await pause(1800);
  await page.getByRole('button',{name:'Сохранить черновик',exact:true}).click();
  await expect(page.getByRole('status')).toContainText('Черновик сохранён');await pause();
  await page.reload();await page.getByLabel('Плательщик',{exact:true}).scrollIntoViewIfNeeded();
  await expect(page.getByLabel('Назначение платежа',{exact:true})).toHaveValue('Разработка интерфейса и интеграций');
  await expect(page.getByRole('button',{name:'Выпустить в банке',exact:true})).toBeDisabled();
}
export async function recordWorkflow({page,base,project,pause,headers}) {
  await page.goto(`${base}/dashboard`);
  const response=await page.request.post(`${base}/api/v1/projects/${project}/workflows`,{headers,data:{name:'Уведомление после заявки',slug:`tech-${Date.now()}`}});
  expect(response.ok(),await response.text()).toBeTruthy();const workflow=await response.json();
  const nodes=[{action_key:'start',action_type_key:'manual.trigger',label:'Ручной запуск',config:{}},{action_key:'mail',action_type_key:'email.send',label:'Письмо клиенту',config:{}},{action_key:'finish',action_type_key:'stop',label:'Завершение',config:{}}];
  const edges=[{from:'start',to:'mail'},{from:'mail',to:'finish'}];
  const saved=await page.request.put(`${base}/api/v1/workflows/${workflow.workflow_id}/draft`,{headers,data:{draft_definition_json:{nodes,edges},draft_graph_json:{nodes:nodes.map((node,index)=>({id:node.action_key,type:'action',position:{x:50+index*320,y:150},data:{label:node.label,action_type_key:node.action_type_key,config:node.config}})),edges:edges.map((edge,index)=>({id:`edge-${index}`,source:edge.from,target:edge.to}))},draft_dsl_yaml:'',test_input_json:{}}});
  expect(saved.ok(),await saved.text()).toBeTruthy();
  await page.goto(`${base}/workflows/${workflow.workflow_id}/edit?action=mail`);await pause(1700);
  await page.locator('.react-flow__node[data-id="mail"]').click();await pause();
  await page.getByRole('button',{name:'Отключить действие',exact:true}).click();await pause();
  const result=page.waitForResponse(r=>r.url().endsWith(`/workflows/${workflow.workflow_id}/draft`)&&r.request().method()==='PUT');
  await page.getByRole('button',{name:'Сохранить',exact:true}).click();expect((await result).ok()).toBeTruthy();await pause();
  await page.reload();await page.locator('.react-flow__node[data-id="mail"]').click();
  await expect(page.getByRole('button',{name:'Включить действие',exact:true})).toBeVisible();
  await expect(page.locator('.react-flow__node[data-id="mail"]')).toContainText('Отключено');
}
export async function recordChat({page,base,pause}) {
  const ksenia={visit_session_id:'22222222-2222-4222-8222-000000000001',conversation_id:'33333333-3333-4333-8333-000000000001',visitor_id:'demo-ksenia',visitor_name:'Ксения',current_url:'https://example.test/services',current_title:'Услуги',status:'active',started_at:new Date().toISOString(),last_activity_at:new Date().toISOString(),duration_seconds:125,visitor_message_count:1,visitor_state_snapshot:{activity:{connection_status:'online'},signals:{chat_open:true}}};
  await page.route('**/site-visitors/active',r=>r.fulfill({json:[ksenia]}));
  await page.route('**/site-chat-widgets',r=>r.fulfill({json:[]}));
  await page.route('**/site-visit-sessions/*/events?*',r=>r.fulfill({json:[]}));
  const messages=[{message_id:'a',author_kind:'visitor',author_name:'Ксения',body:'Здравствуйте! Подскажите, как подключить чат к сайту?',created_at:new Date().toISOString()}];let fail=true;
  await page.route('**/site-chat-conversations/*/messages',async r=>{
    if(r.request().method()==='POST'){
      if(fail){fail=false;return r.fulfill({status:503,contentType:'application/problem+json',json:{title:'Временная ошибка',detail:'Не удалось отправить сообщение. Повторите попытку.'}});}
      const row={message_id:'sent',author_kind:'operator',author_name:'Денис',body:r.request().postDataJSON().body,created_at:new Date().toISOString()};messages.push(row);return r.fulfill({json:row});
    }return r.fulfill({json:messages});
  });
  await page.goto(`${base}/site-chat`);await pause();
  const input=page.getByRole('textbox',{name:'Сообщение посетителю'});
  await input.pressSequentially('Здравствуйте! Поможем с настройкой чата и интеграцией с вашим сайтом.',{delay:35});await pause(900);
  await page.getByRole('button',{name:'Отправить',exact:true}).click();await expect(input).not.toHaveValue('');await pause(1700);
  await input.press('Enter');await expect(input).toHaveValue('');await expect(page.getByRole('log')).toContainText('Поможем с настройкой');
}
