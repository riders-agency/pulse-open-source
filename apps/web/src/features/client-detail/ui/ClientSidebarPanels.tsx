import Link from "next/link";
import { ArrowRight, CreditCard, Trash2 } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Field, Input, Textarea } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import { dateInputProps } from "../../../lib/dates";
import { responsibleRoleLabel } from "../../../lib/client-labels";
import type { ClientDetailState } from "../model/useClientDetailState";

export function ClientSidebarPanels({ state }: { state: ClientDetailState }) {
  return (
    <aside className="stack">
      <ManualPaymentPanel state={state} />
      <ResponsiblePanel state={state} />
      <ProjectsPanel state={state} />
      <ContactsPanel state={state} />
      <DangerPanel state={state} />
    </aside>
  );
}

function ManualPaymentPanel({ state }: { state: ClientDetailState }) {
  return (
    <section className="section-panel">
      <header><div><h2>Счет оплачен</h2><p>Фиксация поступления денег по услуге. Счета и документы ведутся в очереди оплат.</p></div></header>
      <div className="section-panel-body stack">
        {state.services.length === 0 ? (
          <div className="empty-state">У клиента пока нет услуг. Чтобы внести оплату по услуге, сначала добавьте услугу.</div>
        ) : (
          <>
            <Field label="Услуга">
              <Select
                value={state.paymentForm.client_service_id}
                onChange={(event) => {
                  const service = state.services.find((item) => item.client_service_id === event.target.value);
                  if (service) state.selectPaymentService(service);
                  else state.setPaymentForm({ ...state.paymentForm, client_service_id: event.target.value });
                }}
              >
                <option value="">Выберите услугу</option>
                {state.services.map((item) => <option key={item.client_service_id} value={item.client_service_id}>{item.name}</option>)}
              </Select>
            </Field>
            <Field label="Сумма"><Input inputMode="decimal" value={state.paymentForm.amount} onChange={(event) => state.setPaymentForm({ ...state.paymentForm, amount: event.target.value })} /></Field>
            <Field label="Период с"><Input {...dateInputProps} value={state.paymentForm.period_start} onChange={(event) => state.updatePaymentPeriodStart(event.target.value)} /></Field>
            <Field label="Период по"><Input {...dateInputProps} value={state.paymentForm.period_end} onChange={(event) => state.setPaymentForm({ ...state.paymentForm, period_end: event.target.value })} /></Field>
            <Field label="Следующая дата платежа"><Input {...dateInputProps} value={state.paymentForm.next_due_date} onChange={(event) => state.setPaymentForm({ ...state.paymentForm, next_due_date: event.target.value })} /></Field>
            <Field label="Дата оплаты"><Input {...dateInputProps} value={state.paymentForm.paid_at} onChange={(event) => state.setPaymentForm({ ...state.paymentForm, paid_at: event.target.value })} /></Field>
            <Field label="Способ оплаты">
              <Select value={state.paymentForm.payment_method} onChange={(event) => state.setPaymentForm({ ...state.paymentForm, payment_method: event.target.value })}>
                <option value="legal_invoice">По счету с юрлица</option>
                <option value="cash">Наличные</option>
                <option value="bank_transfer">Банковский перевод</option>
                <option value="card">Карта</option>
                <option value="other">Другое</option>
              </Select>
            </Field>
            <Field label="Комментарий"><Textarea value={state.paymentForm.comment} onChange={(event) => state.setPaymentForm({ ...state.paymentForm, comment: event.target.value })} /></Field>
            <Button onClick={state.recordPayment} disabled={state.paymentBusy || !state.paymentForm.client_service_id || !state.paymentForm.period_start || !state.paymentForm.period_end || !state.paymentForm.amount || !state.paymentForm.payment_method}>
              <CreditCard size={18} />
              <span>{state.paymentBusy ? "Сохраняем..." : "Счет оплачен"}</span>
            </Button>
          </>
        )}
      </div>
    </section>
  );
}

function ResponsiblePanel({ state }: { state: ClientDetailState }) {
  return (
    <section className="section-panel">
      <header><div><h2>Ответственные</h2><p>Со стороны Riders</p></div></header>
      <div className="section-panel-body stack">
        <div className="compact-list">
          {state.responsibles.length === 0 ? <div className="empty-state">Не назначены</div> : null}
          {state.responsibles.map((item) => (
            <div className="compact-item" key={item.client_responsible_id}>
              <div className="between"><strong>{item.display_name}</strong>{item.is_primary ? <Badge tone="success">Главный</Badge> : null}</div>
              <span className="muted">{responsibleRoleLabel(item.role)}{item.legal_entity_name ? ` · ${item.legal_entity_name}` : ""}</span>
              <div className="form-actions"><Button variant="ghost" size="sm" onClick={() => state.deleteResponsible(item.client_responsible_id)}><Trash2 size={16} /></Button></div>
            </div>
          ))}
        </div>
        <Field label="Имя"><Input value={state.responsible.display_name} onChange={(event) => state.setResponsible({ ...state.responsible, display_name: event.target.value })} /></Field>
        <Field label="Роль">
          <Select value={state.responsible.role} onChange={(event) => state.setResponsible({ ...state.responsible, role: event.target.value })}>
            <option value="account_manager">Аккаунт</option><option value="technical_lead">Техника</option><option value="finance">Финансы</option>
            <option value="legal">Документы</option><option value="owner">Владелец</option><option value="other">Другое</option>
          </Select>
        </Field>
        <Field label="Наше юрлицо">
          <Select value={state.responsible.company_legal_entity_id} onChange={(event) => state.setResponsible({ ...state.responsible, company_legal_entity_id: event.target.value })}>
            <option value="">Не выбрано</option>
            {state.legalEntities.map((item) => <option value={item.company_legal_entity_id} key={item.company_legal_entity_id}>{item.name}</option>)}
          </Select>
        </Field>
        <label className="row">
          <input type="checkbox" checked={state.responsible.is_primary} onChange={(event) => state.setResponsible({ ...state.responsible, is_primary: event.target.checked })} />
          <span>Главный ответственный</span>
        </label>
        <Button onClick={state.addResponsible} disabled={!state.responsible.display_name.trim()}>Добавить ответственного</Button>
      </div>
    </section>
  );
}

function ProjectsPanel({ state }: { state: ClientDetailState }) {
  return (
    <section className="section-panel">
      <header><div><h2>Проекты</h2><p>Граница услуг и сценариев</p></div></header>
      <div className="section-panel-body stack">
        <div className="row">
          <Select value={state.projectId} onChange={(event) => state.setProjectId(event.target.value)}>
            <option value="">Выберите проект</option>
            {state.allProjects.map((project) => <option key={project.project_id} value={project.project_id}>{project.name}</option>)}
          </Select>
          <Button onClick={state.linkProject} disabled={!state.projectId}>Связать</Button>
        </div>
        <div className="compact-list">
          {state.links.map((project) => (
            <Link className="compact-item" href={`/projects/${project.project_id}`} key={project.project_id} onClick={() => state.setActiveProject(project.project_id)}>
              <div className="between"><strong>{project.name}</strong><span className="action-link">Открыть проект <ArrowRight size={15} /></span></div>
              <span className="muted">Услуги и сценарии клиента работают внутри этого проекта.</span>
            </Link>
          ))}
        </div>
      </div>
    </section>
  );
}

function ContactsPanel({ state }: { state: ClientDetailState }) {
  return (
    <section className="section-panel">
      <header><div><h2>Контакты</h2><p>Люди клиента</p></div></header>
      <div className="section-panel-body stack">
        <div className="compact-list">
          {state.contacts.length === 0 ? <div className="empty-state">Контактов пока нет</div> : null}
          {state.contacts.map((item) => (
            <div className="compact-item" key={item.client_contact_id}>
              <strong>{[item.last_name, item.first_name, item.middle_name].filter(Boolean).join(" ") || "Контакт"}</strong>
              <span className="muted">{item.email || item.phone || item.telegram || "Канал не указан"}</span>
            </div>
          ))}
        </div>
        <div className="grid-2">
          <Field label="Имя"><Input value={state.contact.first_name} onChange={(event) => state.setContact({ ...state.contact, first_name: event.target.value })} /></Field>
          <Field label="Фамилия"><Input value={state.contact.last_name} onChange={(event) => state.setContact({ ...state.contact, last_name: event.target.value })} /></Field>
        </div>
        <Field label="Электронная почта"><Input value={state.contact.email} onChange={(event) => state.setContact({ ...state.contact, email: event.target.value })} /></Field>
        <Field label="Телефон"><Input value={state.contact.phone} onChange={(event) => state.setContact({ ...state.contact, phone: event.target.value })} /></Field>
        <Field label="Telegram"><Input value={state.contact.telegram} onChange={(event) => state.setContact({ ...state.contact, telegram: event.target.value })} /></Field>
        <Button onClick={state.addContact}>Добавить контакт</Button>
      </div>
    </section>
  );
}

function DangerPanel({ state }: { state: ClientDetailState }) {
  return (
    <section className="section-panel danger-zone">
      <header><div><h2>Удаление</h2><p>Клиент уйдёт из активного списка.</p></div></header>
      <div className="section-panel-body">
        <Button variant="destructive" onClick={state.deleteClient} disabled={state.busy || !state.client}>
          <Trash2 size={18} />
          <span>Удалить клиента</span>
        </Button>
      </div>
    </section>
  );
}
