"use client";

import { Check, Search } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Field, Input } from "../../../components/ui/input";
import type { DaDataIntegration, DaDataPartySuggestion } from "../../../lib/api";
import { clientTypeLabel } from "../../../lib/client-labels";

export function CompanyClientDaDataPanel({
  applySuggestion,
  companyId,
  dadata,
  dadataBusy,
  dadataKey,
  dadataQuery,
  findParty,
  saveDaData,
  setDadataKey,
  setDadataQuery,
  suggestions,
}: {
  applySuggestion: (item: DaDataPartySuggestion) => void;
  companyId?: string;
  dadata: DaDataIntegration | null;
  dadataBusy: boolean;
  dadataKey: string;
  dadataQuery: string;
  findParty: () => Promise<void>;
  saveDaData: () => Promise<void>;
  setDadataKey: (value: string) => void;
  setDadataQuery: (value: string) => void;
  suggestions: DaDataPartySuggestion[];
}) {
  return (
    <section className="section-panel">
      <header>
        <div>
          <h2>Заполнение по ИНН</h2>
          <p>{dadata?.configured ? "Подключено" : "Сначала добавьте ключ"}</p>
        </div>
        <Badge tone={dadata?.configured ? "success" : "warning"}>{dadata?.configured ? "DaData готова" : "DaData не подключена"}</Badge>
      </header>
      <div className="section-panel-body stack">
        {dadata?.configured ? (
          <>
            <div className="row">
              <Input value={dadataQuery} onChange={(event) => setDadataQuery(event.target.value)} placeholder="ИНН или ОГРН" onKeyDown={(event) => {
                if (event.key === "Enter" && dadataQuery.trim()) void findParty();
              }} />
              <Button onClick={findParty} disabled={dadataBusy || !dadataQuery.trim()}>
                <Search size={18} />
                <span>Найти</span>
              </Button>
            </div>
            {suggestions.length ? (
              <div className="dadata-results">
                {suggestions.map((item) => (
                  <div className="dadata-result" key={`${item.inn}-${item.kpp}-${item.ogrn}-${item.ogrnip}`}>
                    <div>
                      <strong>{item.name}</strong>
                      <div className="client-row-meta">
                        <span>{clientTypeLabel(item.client_type)}</span>
                        <span className="mono">ИНН {item.inn}</span>
                        {item.kpp ? <span className="mono">КПП {item.kpp}</span> : null}
                      </div>
                    </div>
                    <Button variant="secondary" onClick={() => applySuggestion(item)}>
                      <Check size={18} />
                      <span>Заполнить</span>
                    </Button>
                  </div>
                ))}
              </div>
            ) : null}
          </>
        ) : (
          <div className="grid-2">
            <Field label="API-ключ DaData">
              <Input type="password" value={dadataKey} onChange={(event) => setDadataKey(event.target.value)} placeholder="Token" autoComplete="off" />
            </Field>
            <div className="form-actions">
              <Button onClick={saveDaData} disabled={dadataBusy || !dadataKey.trim() || !companyId}>
                Сохранить ключ
              </Button>
            </div>
          </div>
        )}
      </div>
    </section>
  );
}
