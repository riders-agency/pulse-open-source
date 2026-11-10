"use client";

import { Save, Trash2 } from "lucide-react";
import type { Dispatch, SetStateAction } from "react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Field, Input } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { ClientService, ClientServiceType, Project } from "../../../lib/api";
import { invoiceScenarioOptions, statusLabel, statusTone } from "../../../lib/client-labels";
import { dateInputProps } from "../../../lib/dates";
import type { ServiceForm } from "../model/clientServiceDetailModel";

export function ClientServiceFormPanel({
  busy,
  form,
  projects,
  remove,
  save,
  service,
  serviceTypes,
  setForm,
}: {
  busy: boolean;
  form: ServiceForm | null;
  projects: Project[];
  remove: () => Promise<void>;
  save: () => Promise<void>;
  service: ClientService | null;
  serviceTypes: ClientServiceType[];
  setForm: Dispatch<SetStateAction<ServiceForm | null>>;
}) {
  return (
    <section className="section-panel section-panel-spaced">
      <header>
        <div>
          <h2>Параметры</h2>
          <p>Эти данные используются очередью оплат и сценариями напоминаний.</p>
        </div>
        <Badge tone={statusTone(service?.status)}>{statusLabel(service?.status)}</Badge>
      </header>
      {form ? (
        <div className="section-panel-body stack">
          <div className="form-grid">
            <Field label="Проект" className="span-2">
              <Select value={form.project_id} onChange={(event) => setForm({ ...form, project_id: event.target.value })}>
                {projects.map((project) => <option key={project.project_id} value={project.project_id}>{project.name}</option>)}
              </Select>
            </Field>
            <Field label="Тип услуги">
              <Select value={form.service_type_key} onChange={(event) => setForm({ ...form, service_type_key: event.target.value })}>
                {serviceTypes.map((type) => <option key={type.service_type_key} value={type.service_type_key}>{type.display_name}</option>)}
              </Select>
            </Field>
            <Field label="Статус">
              <Select value={form.status} onChange={(event) => setForm({ ...form, status: event.target.value })}>
                <option value="active">Активно</option>
                <option value="paused">На паузе</option>
                <option value="archived">В архиве</option>
              </Select>
            </Field>
            <Field label="Название" className="span-2">
              <Input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} />
            </Field>
            <Field label="Сумма">
              <Input inputMode="decimal" value={form.amount} onChange={(event) => setForm({ ...form, amount: event.target.value })} />
            </Field>
            <Field label="Период">
              <Select value={form.billing_period} onChange={(event) => setForm({ ...form, billing_period: event.target.value })}>
                <option value="once">Разово</option>
                <option value="weekly">Еженедельно</option>
                <option value="monthly">Ежемесячно</option>
                <option value="quarterly">Ежеквартально</option>
                <option value="semiannual">Раз в полгода</option>
                <option value="yearly">Ежегодно</option>
              </Select>
            </Field>
            <Field label="Сценарий счета">
              <Select value={form.invoice_scenario_key} onChange={(event) => setForm({ ...form, invoice_scenario_key: event.target.value })}>
                {invoiceScenarioOptions.map((scenario) => <option key={scenario.key} value={scenario.key}>{scenario.label}</option>)}
              </Select>
            </Field>
            <Field label="Следующий период">
              <Input {...dateInputProps} value={form.next_due_date} onChange={(event) => setForm({ ...form, next_due_date: event.target.value })} />
            </Field>
          </div>
          <div className="form-actions">
            <Button variant="destructive" onClick={remove} disabled={busy}><Trash2 size={18} /><span>Удалить</span></Button>
            <Button onClick={save} disabled={busy || !form.name.trim() || !form.project_id}><Save size={18} /><span>{busy ? "Сохраняем..." : "Сохранить"}</span></Button>
          </div>
        </div>
      ) : null}
    </section>
  );
}
