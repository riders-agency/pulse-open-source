"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { ArrowRight, Plus, Search } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";

import { Badge } from "../../components/ui/badge";
import { Input } from "../../components/ui/input";
import type { Client } from "../../lib/api";
import { clientTypeLabel, statusLabel, statusTone } from "../../lib/client-labels";
import { listCompanyClients } from "./api/companyClientsApi";
import { usePulseStore } from "../../store/pulse-store";

export default function CompanyClientsPage() {
  const params = useParams<{ companyId?: string }>();
  const token = usePulseStore((s) => s.token);
  const activeCompanyId = usePulseStore((s) => s.activeCompanyId);
  const activeClientId = usePulseStore((s) => s.activeClientId);
  const setActiveCompany = usePulseStore((s) => s.setActiveCompany);
  const setActiveClient = usePulseStore((s) => s.setActiveClient);
  const [clients, setClients] = useState<Client[]>([]);
  const [query, setQuery] = useState("");
  const [error, setError] = useState("");

  const filteredClients = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return clients;
    return clients.filter((client) =>
      [client.name, client.short_name, client.full_name, client.inn, client.kpp, client.ogrn, client.ogrnip]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(q)),
    );
  }, [clients, query]);

  const totals = useMemo(
    () => ({
      active: clients.filter((client) => client.status === "active").length,
      projects: clients.reduce((sum, client) => sum + (client.projects_count ?? 0), 0),
      services: clients.reduce((sum, client) => sum + (client.services_count ?? 0), 0),
    }),
    [clients],
  );

  const companyId = params.companyId || activeCompanyId;

  const load = useCallback(() => {
    if (!token) return;
    if (!companyId) return;
    if (params.companyId) setActiveCompany(companyId);
    listCompanyClients(companyId, token)
      .then((rows) => {
        setClients(rows);
        setError("");
      })
      .catch((err) => setError(err.message));
  }, [companyId, params.companyId, setActiveCompany, token]);

  useEffect(load, [load]);

  useEffect(() => {
    const onAIEntityChanged = () => load();
    window.addEventListener("pulse:ai-entity-changed", onAIEntityChanged);
    return () => window.removeEventListener("pulse:ai-entity-changed", onAIEntityChanged);
  }, [load]);

  return (
    <main className="page">
      <div className="clients-toolbar">
        <div>
          <h1>Клиенты</h1>
          <p className="muted">Клиенты компании, услуги, ответственные и связи с проектами.</p>
        </div>
        <Link className={`ui-button ui-button--default ui-button--md ${!companyId ? "is-disabled" : ""}`} href={companyId ? "/clients/new" : "/companies"}>
          <Plus size={18} />
          <span>Новый клиент</span>
        </Link>
      </div>

      {!companyId ? <div className="empty-state">Выберите компанию слева, чтобы открыть клиентов.</div> : null}
      {error ? <Badge tone="danger">{error}</Badge> : null}

      <section className="grid-3">
        <div className="metric"><span>Всего</span><strong>{clients.length}</strong></div>
        <div className="metric"><span>Активные</span><strong>{totals.active}</strong></div>
        <div className="metric"><span>Услуги</span><strong>{totals.services}</strong></div>
      </section>

      <section className="section-panel section-panel-spaced">
        <header>
          <div>
            <h2>Реестр</h2>
            <p>Общий реестр компании с привязками к проектам.</p>
          </div>
          <label className="ui-field client-search">
            <span>Поиск</span>
            <span className="row">
              <Search size={17} />
              <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Название, ИНН, ОГРН" />
            </span>
          </label>
        </header>
        <div className="section-panel-body">
          {filteredClients.length === 0 ? (
            <div className="empty-state">{query ? "Ничего не найдено" : "Клиенты появятся после импорта или создания в компании"}</div>
          ) : (
            <div className="client-card-list">
              {filteredClients.map((client) => (
                <Link
                  className={`client-row-card ${client.client_id === activeClientId ? "is-active" : ""}`}
                  href={`/clients/${client.client_id}`}
                  key={client.client_id}
                  onClick={() => setActiveClient(client.client_id)}
                >
                  <div>
                    <div className="client-row-title">
                      <strong>{client.name}</strong>
                      <Badge>{clientTypeLabel(client.client_type)}</Badge>
                      <Badge tone={statusTone(client.status)}>{statusLabel(client.status)}</Badge>
                    </div>
                    <div className="client-row-meta">
                      <span className="mono">ИНН {client.inn || "не указан"}</span>
                      {client.kpp ? <span className="mono">КПП {client.kpp}</span> : null}
                      <span>{client.projects_count ?? 0} проектов</span>
                      <span>{client.services_count ?? 0} услуг</span>
                    </div>
                  </div>
                  <div className="client-row-actions">
                    <span className="action-link">Открыть карточку <ArrowRight size={15} /></span>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </div>
      </section>
    </main>
  );
}
