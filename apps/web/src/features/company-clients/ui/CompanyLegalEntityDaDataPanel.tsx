import { Search } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Field, Input } from "../../../components/ui/input";
import { legalEntityTypeLabel } from "../../../lib/client-labels";
import type { DaDataIntegration, DaDataPartySuggestion } from "../../../lib/api";

export function CompanyLegalEntityDaDataPanel({
  busy,
  companyId,
  dadata,
  dadataBusy,
  dadataKey,
  query,
  suggestions,
  applySuggestion,
  find,
  saveDaData,
  setDadataKey,
  setQuery,
}: {
  busy: boolean;
  companyId?: string;
  dadata: DaDataIntegration | null;
  dadataBusy: boolean;
  dadataKey: string;
  query: string;
  suggestions: DaDataPartySuggestion[];
  applySuggestion: (item: DaDataPartySuggestion) => void;
  find: () => void;
  saveDaData: () => void;
  setDadataKey: (value: string) => void;
  setQuery: (value: string) => void;
}) {
  if (!dadata?.configured) {
    return (
      <div className="grid-2">
        <Field label="API-ключ DaData">
          <Input
            type="password"
            value={dadataKey}
            onChange={(event) => setDadataKey(event.target.value)}
            placeholder="Token"
            autoComplete="off"
          />
        </Field>
        <div className="form-actions">
          <Button onClick={saveDaData} disabled={dadataBusy || !dadataKey.trim() || !companyId}>
            Сохранить ключ
          </Button>
        </div>
      </div>
    );
  }
  return (
    <div className="stack">
      <div className="row">
        <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="ИНН или ОГРН" />
        <Button onClick={find} disabled={busy || !query.trim()}>
          <Search size={18} />
          <span>Найти</span>
        </Button>
      </div>
      {suggestions.length ? (
        <div className="dadata-results">
          {suggestions.map((item) => (
            <button className="dadata-result" type="button" key={`${item.inn}-${item.kpp}`} onClick={() => applySuggestion(item)}>
              <span>{item.name}</span>
              <Badge>{legalEntityTypeLabel(item.client_type)}</Badge>
            </button>
          ))}
        </div>
      ) : null}
    </div>
  );
}
