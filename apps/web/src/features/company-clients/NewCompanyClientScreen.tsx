"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "../../components/ui/button";
import {
  attachClientToProject,
  createCompanyClient,
  findCompanyDaDataParty,
  getCompanyDaDataIntegration,
  saveCompanyDaDataIntegration,
} from "./api/companyClientsApi";
import {
  companyClientFormFromDaData,
  emptyCompanyClientForm,
  type CompanyClientForm,
} from "./model/companyPartyForms";
import { type DaDataIntegration, type DaDataPartySuggestion } from "../../lib/api";
import { usePulseStore } from "../../store/pulse-store";
import { CompanyClientDaDataPanel } from "./ui/CompanyClientDaDataPanel";
import { NewCompanyClientFields } from "./ui/NewCompanyClientFields";

export default function NewClientPage() {
  const params = useParams<{ companyId?: string }>();
  const router = useRouter();
  const token = usePulseStore((s) => s.token);
  const activeCompanyId = usePulseStore((s) => s.activeCompanyId);
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const setActiveCompany = usePulseStore((s) => s.setActiveCompany);
  const setActiveClient = usePulseStore((s) => s.setActiveClient);
  const [form, setForm] = useState<CompanyClientForm>(emptyCompanyClientForm);
  const [dadata, setDadata] = useState<DaDataIntegration | null>(null);
  const [dadataKey, setDadataKey] = useState("");
  const [dadataQuery, setDadataQuery] = useState("");
  const [suggestions, setSuggestions] = useState<DaDataPartySuggestion[]>([]);
  const [busy, setBusy] = useState(false);
  const [dadataBusy, setDadataBusy] = useState(false);
  const [error, setError] = useState("");
  const companyId = params.companyId || activeCompanyId;

  useEffect(() => {
    if (!token) return;
    if (!companyId) return;
    if (params.companyId) setActiveCompany(companyId);
    getCompanyDaDataIntegration(companyId, token)
      .then(setDadata)
      .catch(() => setDadata({ configured: false }));
  }, [companyId, params.companyId, setActiveCompany, token]);

  const saveDaData = async () => {
    if (!companyId) return;
    setError("");
    setDadataBusy(true);
    try {
      const row = await saveCompanyDaDataIntegration(companyId, token, dadataKey);
      setDadata(row);
      setDadataKey("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить DaData");
    } finally {
      setDadataBusy(false);
    }
  };

  const findParty = async () => {
    if (!companyId) return;
    setError("");
    setDadataBusy(true);
    try {
      setSuggestions(await findCompanyDaDataParty(companyId, token, dadataQuery));
    } catch (err) {
      setError(err instanceof Error ? err.message : "DaData не вернула данные");
    } finally {
      setDadataBusy(false);
    }
  };

  const applySuggestion = (item: DaDataPartySuggestion) => {
    setForm(companyClientFormFromDaData(item, form.comment));
    setSuggestions([]);
  };

  const create = async () => {
    if (!companyId) return;
    setError("");
    setBusy(true);
    try {
      const out = await createCompanyClient(companyId, token, form);
      if (activeProjectId) {
        await attachClientToProject(out.client_id, activeProjectId, token);
      }
      setActiveClient(out.client_id);
      router.push(`/clients/${out.client_id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось создать клиента");
    } finally {
      setBusy(false);
    }
  };

  return (
    <main className="page page-narrow">
      <div className="page-title">
        <div>
          <h1>Новый клиент</h1>
          <p className="muted">Карточка создаётся в компании и может быть связана с выбранным проектом.</p>
        </div>
        <Link className="ui-button ui-button--secondary ui-button--md" href="/clients">
          <ArrowLeft size={18} />
          <span>К списку</span>
        </Link>
      </div>

      {!companyId ? <div className="empty-state">Выберите компанию слева, чтобы создать клиента.</div> : null}
      {error ? <div className="form-error">{error}</div> : null}

      <div className="stack">
        <CompanyClientDaDataPanel
          applySuggestion={applySuggestion}
          companyId={companyId}
          dadata={dadata}
          dadataBusy={dadataBusy}
          dadataKey={dadataKey}
          dadataQuery={dadataQuery}
          findParty={findParty}
          saveDaData={saveDaData}
          setDadataKey={setDadataKey}
          setDadataQuery={setDadataQuery}
          suggestions={suggestions}
        />

        <section className="section-panel">
          <header>
            <div>
              <h2>Реквизиты</h2>
              <p>Минимум для сохранения: тип и название.</p>
            </div>
          </header>
          <div className="section-panel-body stack">
            <NewCompanyClientFields form={form} setForm={setForm} />
            <div className="form-actions">
              <Button onClick={create} disabled={busy || !form.name.trim() || !companyId}>
                {busy ? "Создаём..." : "Создать клиента"}
              </Button>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
