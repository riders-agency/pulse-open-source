"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Plus } from "lucide-react";
import { useCallback, useEffect, useState } from "react";

import { Button } from "../../components/ui/button";
import { Field, Input } from "../../components/ui/input";
import { Select } from "../../components/ui/select";
import {
  createClientService,
  getClientServiceCreateContext,
  listClientServiceTypes,
} from "./api/clientServiceDetailApi";
import {
  clientServiceCreatePayload,
  emptyClientServiceCreateForm,
  type ClientServiceCreateForm,
} from "./model/clientServiceDetailModel";
import { type Client, type ClientServiceType, type Project } from "../../lib/api";
import { invoiceScenarioOptions } from "../../lib/client-labels";
import { dateInputProps } from "../../lib/dates";
import { usePulseStore } from "../../store/pulse-store";

export default function NewClientServicePage() {
  const params = useParams<{ clientId?: string }>();
  const router = useRouter();
  const token = usePulseStore((s) => s.token);
  const projects = usePulseStore((s) => s.projects);
  const activeClientId = usePulseStore((s) => s.activeClientId);
  const setActiveCompany = usePulseStore((s) => s.setActiveCompany);
  const setActiveClient = usePulseStore((s) => s.setActiveClient);
  const [client, setClient] = useState<Client | null>(null);
  const [allProjects, setAllProjects] = useState<Project[]>([]);
  const [serviceTypes, setServiceTypes] = useState<ClientServiceType[]>([]);
  const [form, setForm] = useState<ClientServiceCreateForm>(emptyClientServiceCreateForm);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const clientId = params.clientId || activeClientId;

  const load = useCallback(() => {
    if (!token) return;
    if (!clientId) return;
    getClientServiceCreateContext(clientId, token)
      .then(({ client: row, projects: projectList }) => {
        setClient(row);
        setActiveCompany(row.company_id);
        if (params.clientId) setActiveClient(row.client_id);
        setAllProjects(projectList);
        setForm((current) => ({ ...current, project_id: current.project_id || projectList[0]?.project_id || "" }));
      })
      .catch((err) => {
        setError(err.message);
        setAllProjects(projects as Project[]);
      });
  }, [clientId, params.clientId, projects, setActiveClient, setActiveCompany, token]);

  useEffect(load, [load]);

  useEffect(() => {
    if (!token || !form.project_id) return;
    listClientServiceTypes(form.project_id, token)
      .then((rows) => {
        setServiceTypes(rows);
        setForm((current) => {
          if (!rows.length || rows.some((type) => type.service_type_key === current.service_type_key)) {
            return current;
          }
          return { ...current, service_type_key: rows[0].service_type_key };
        });
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить типы услуг"));
  }, [form.project_id, token]);

  const create = async () => {
    setBusy(true);
    setError("");
    try {
      await createClientService(clientId, token, clientServiceCreatePayload(form));
      router.push(`/clients/${clientId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось добавить услугу");
    } finally {
      setBusy(false);
    }
  };

  return (
    <main className="page page-narrow">
      <div className="page-title">
        <div>
          <h1>Новая услуга</h1>
          <p className="muted">{client?.name ?? "Клиент"}</p>
        </div>
        <Link className="ui-button ui-button--secondary ui-button--md" href={clientId ? `/clients/${clientId}` : "/clients"}>
          <ArrowLeft size={18} />
          <span>К карточке</span>
        </Link>
      </div>

      {error ? <div className="form-error">{error}</div> : null}

      <section className="section-panel">
        <header>
          <div>
            <h2>Параметры услуги</h2>
            <p>Очередь оплат будет строиться по проекту и следующей дате.</p>
          </div>
        </header>
        <div className="section-panel-body stack">
          <div className="form-grid">
            <Field label="Проект" className="span-2">
              <Select value={form.project_id} onChange={(event) => setForm({ ...form, project_id: event.target.value })}>
                <option value="">Выберите проект</option>
                {allProjects.map((project) => (
                  <option key={project.project_id} value={project.project_id}>{project.name}</option>
                ))}
              </Select>
            </Field>
            <Field label="Тип услуги">
              <Select value={form.service_type_key} onChange={(event) => setForm({ ...form, service_type_key: event.target.value })}>
                {serviceTypes.map((type) => (
                  <option key={type.service_type_key} value={type.service_type_key}>{type.display_name}</option>
                ))}
              </Select>
            </Field>
            <Field label="Название">
              <Input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} />
            </Field>
            <Field label="Сумма">
              <Input inputMode="decimal" value={form.amount} onChange={(event) => setForm({ ...form, amount: event.target.value })} placeholder="0" />
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
                {invoiceScenarioOptions.map((scenario) => (
                  <option key={scenario.key} value={scenario.key}>{scenario.label}</option>
                ))}
              </Select>
            </Field>
            <Field label="Следующий период">
              <Input {...dateInputProps} value={form.next_due_date} onChange={(event) => setForm({ ...form, next_due_date: event.target.value })} />
            </Field>
          </div>
          <div className="form-actions">
            <Button onClick={create} disabled={busy || !clientId || !form.project_id || !form.name.trim()}>
              <Plus size={18} />
              <span>{busy ? "Добавляем..." : "Добавить услугу"}</span>
            </Button>
          </div>
        </div>
      </section>
    </main>
  );
}
