"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { useCallback, useEffect, useState } from "react";

import type { Client, ClientService, ClientServiceType, Project } from "../../lib/api";
import { billingPeriodLabel, invoiceScenarioLabel, statusLabel } from "../../lib/client-labels";
import {
  deleteClientService,
  getClientServiceDetail,
  listClientServiceTypes,
  updateClientService,
} from "./api/clientServiceDetailApi";
import {
  formatMoney,
  formFromService,
  serviceFormPayload,
  type ServiceForm,
} from "./model/clientServiceDetailModel";
import { usePulseStore } from "../../store/pulse-store";
import { ClientServiceFormPanel } from "./ui/ClientServiceFormPanel";

export default function ClientServicePage() {
  const params = useParams<{ clientServiceId?: string }>();
  const router = useRouter();
  const token = usePulseStore((s) => s.token);
  const activeClientServiceId = usePulseStore((s) => s.activeClientServiceId);
  const setActiveCompany = usePulseStore((s) => s.setActiveCompany);
  const setActiveProject = usePulseStore((s) => s.setActiveProject);
  const setActiveClient = usePulseStore((s) => s.setActiveClient);
  const setActiveClientService = usePulseStore((s) => s.setActiveClientService);
  const [service, setService] = useState<ClientService | null>(null);
  const [client, setClient] = useState<Client | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [serviceTypes, setServiceTypes] = useState<ClientServiceType[]>([]);
  const [form, setForm] = useState<ServiceForm | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const clientServiceId = params.clientServiceId || activeClientServiceId;

  const load = useCallback(() => {
    if (!token) return;
    if (!clientServiceId) {
      setError("Не выбран client_service_id. Откройте услугу из карточки клиента.");
      return;
    }
    getClientServiceDetail(clientServiceId, token)
      .then(({ service: serviceRow, client: clientRow, projects: projectRows, serviceTypes: typeRows }) => {
        setService(serviceRow);
        setForm(formFromService(serviceRow));
        setActiveCompany(serviceRow.company_id);
        setActiveProject(serviceRow.project_id);
        setActiveClient(serviceRow.client_id);
        if (params.clientServiceId) setActiveClientService(serviceRow.client_service_id);
        setClient(clientRow);
        setProjects(projectRows);
        setServiceTypes(typeRows);
        setError("");
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить услугу"));
  }, [clientServiceId, params.clientServiceId, setActiveClient, setActiveClientService, setActiveCompany, setActiveProject, token]);

  useEffect(load, [load]);

  useEffect(() => {
    if (!token || !form?.project_id) return;
    listClientServiceTypes(form.project_id, token)
      .then((rows) => setServiceTypes(rows))
      .catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить типы услуг"));
  }, [form?.project_id, token]);

  const save = async () => {
    if (!form) return;
    if (!clientServiceId) {
      setError("Не выбран client_service_id. Откройте услугу из карточки клиента.");
      return;
    }
    setBusy(true);
    setError("");
    try {
      if (!token) return;
      await updateClientService(clientServiceId, token, serviceFormPayload(form, service?.metadata_json));
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить услугу");
    } finally {
      setBusy(false);
    }
  };

  const remove = async () => {
    if (!service || !window.confirm("Удалить услугу?")) return;
    setBusy(true);
    setError("");
    try {
      if (!clientServiceId || !token) return;
      await deleteClientService(clientServiceId, token);
      setActiveClientService("");
      router.push(service.client_id ? `/clients/${service.client_id}` : "/clients");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось удалить услугу");
      setBusy(false);
    }
  };

  return (
    <main className="page page-narrow">
      <div className="page-title">
        <div>
          <h1>{service?.name ?? "Услуга"}</h1>
          <p className="muted">{client?.name ?? "Клиент"} · {service ? billingPeriodLabel(service.billing_period) : "Период"}</p>
        </div>
        {service ? (
          <Link className="ui-button ui-button--secondary ui-button--md" href={service.client_id ? `/clients/${service.client_id}` : "/clients"}>
            <ArrowLeft size={18} />
            <span>К клиенту</span>
          </Link>
        ) : null}
      </div>

      {error ? <div className="form-error">{error}</div> : null}

      <section className="grid-3">
        <div className="metric"><span>Статус</span><strong>{statusLabel(service?.status)}</strong></div>
        <div className="metric"><span>Сумма</span><strong>{formatMoney(service?.amount ?? 0)}</strong></div>
        <div className="metric"><span>Сценарий счета</span><strong>{invoiceScenarioLabel(form?.invoice_scenario_key)}</strong></div>
      </section>

      <ClientServiceFormPanel
        busy={busy}
        form={form}
        projects={projects}
        remove={remove}
        save={save}
        service={service}
        serviceTypes={serviceTypes}
        setForm={setForm}
      />
    </main>
  );
}
