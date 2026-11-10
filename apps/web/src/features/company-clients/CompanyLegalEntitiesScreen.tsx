"use client";

import { useParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";

import {
  createCompanyLegalEntity,
  deleteCompanyLegalEntity,
  findCompanyDaDataParty,
  getCompanyDaDataIntegration,
  listCompanyLegalEntities,
  saveCompanyDaDataIntegration,
} from "./api/companyClientsApi";
import {
  emptyCompanyLegalEntityForm,
  legalEntityFormFromDaData,
  type CompanyLegalEntityForm,
} from "./model/companyPartyForms";
import { type CompanyLegalEntity, type DaDataIntegration, type DaDataPartySuggestion } from "../../lib/api";
import { usePulseStore } from "../../store/pulse-store";
import { CompanyLegalEntityDaDataPanel } from "./ui/CompanyLegalEntityDaDataPanel";
import { CompanyLegalEntityFields } from "./ui/CompanyLegalEntityFields";
import { CompanyLegalEntityList } from "./ui/CompanyLegalEntityList";

export default function CompanyLegalEntitiesPage() {
  const params = useParams<{ companyId?: string }>();
  const token = usePulseStore((s) => s.token);
  const activeCompanyId = usePulseStore((s) => s.activeCompanyId);
  const setActiveCompany = usePulseStore((s) => s.setActiveCompany);
  const [rows, setRows] = useState<CompanyLegalEntity[]>([]);
  const [form, setForm] = useState<CompanyLegalEntityForm>(emptyCompanyLegalEntityForm);
  const [dadata, setDadata] = useState<DaDataIntegration | null>(null);
  const [dadataKey, setDadataKey] = useState("");
  const [query, setQuery] = useState("");
  const [suggestions, setSuggestions] = useState<DaDataPartySuggestion[]>([]);
  const [busy, setBusy] = useState(false);
  const [dadataBusy, setDadataBusy] = useState(false);
  const [error, setError] = useState("");
  const companyId = params.companyId || activeCompanyId;

  const load = useCallback(() => {
    if (!token) return;
    if (!companyId) return;
    if (params.companyId) setActiveCompany(companyId);
    Promise.all([
      listCompanyLegalEntities(companyId, token),
      getCompanyDaDataIntegration(companyId, token),
    ])
      .then(([entityRows, dadataRow]) => {
        setRows(entityRows);
        setDadata(dadataRow);
        setError("");
      })
      .catch((err) => setError(err.message));
  }, [companyId, params.companyId, setActiveCompany, token]);

  useEffect(load, [load]);

  const saveDaData = async () => {
    if (!companyId) return;
    setDadataBusy(true);
    setError("");
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

  const find = async () => {
    if (!companyId) {
      setError("Выберите компанию, чтобы использовать DaData");
      return;
    }
    setBusy(true);
    setError("");
    try {
      setSuggestions(await findCompanyDaDataParty(companyId, token, query));
    } catch (err) {
      setError(err instanceof Error ? err.message : "DaData не вернула данные");
    } finally {
      setBusy(false);
    }
  };

  const applySuggestion = (item: DaDataPartySuggestion) => {
    setForm(legalEntityFormFromDaData(item, form.comment));
    setSuggestions([]);
  };

  const create = async () => {
    setBusy(true);
    setError("");
    try {
      await createCompanyLegalEntity(companyId, token, form);
      setForm(emptyCompanyLegalEntityForm);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить юрлицо");
    } finally {
      setBusy(false);
    }
  };

  const remove = async (id: string) => {
    if (!window.confirm("Удалить юрлицо из активного списка?")) return;
    setError("");
    try {
      await deleteCompanyLegalEntity(id, token);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось удалить");
    }
  };

  return (
    <main className="page">
      <div className="page-title">
        <div>
          <h1>Наши юрлица</h1>
          <p className="muted">Юридические стороны Riders для договоров, оплат и ответственных.</p>
        </div>
      </div>

      {error ? <div className="form-error">{error}</div> : null}

      <section className="layout-sidebar">
        <div className="section-panel">
          <header>
            <div>
              <h2>Список</h2>
              <p>{rows.length} активных записей</p>
            </div>
          </header>
          <div className="section-panel-body">
            <CompanyLegalEntityList rows={rows} onRemove={remove} />
          </div>
        </div>

        <div className="section-panel">
          <header>
            <div>
              <h2>Новое юрлицо</h2>
              <p>{dadata?.configured ? "Можно заполнить по ИНН" : "DaData подключается на уровне компании"}</p>
            </div>
          </header>
          <div className="section-panel-body stack">
            <CompanyLegalEntityDaDataPanel
              applySuggestion={applySuggestion}
              busy={busy}
              companyId={companyId}
              dadata={dadata}
              dadataBusy={dadataBusy}
              dadataKey={dadataKey}
              find={find}
              query={query}
              saveDaData={saveDaData}
              setDadataKey={setDadataKey}
              setQuery={setQuery}
              suggestions={suggestions}
            />
            <CompanyLegalEntityFields busy={busy} form={form} onCreate={create} setForm={setForm} />
          </div>
        </div>
      </section>
    </main>
  );
}
