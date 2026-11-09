"use client";

import { useEffect, useMemo, useState } from "react";
import { Search, X } from "lucide-react";

import { Field, Input } from "../../components/ui/input";
import type { Client } from "../../lib/api";
import { usePulseStore } from "../../store/pulse-store";
import { listCompanyClients } from "./api/workApi";

export function WorkClientPicker({
  companyId,
  value,
  currentName,
  disabled,
  onChange,
}: {
  companyId?: string | null;
  value?: string;
  currentName?: string;
  disabled?: boolean;
  onChange: (clientId: string) => void;
}) {
  const token = usePulseStore((s) => s.token);
  const [query, setQuery] = useState("");
  const [clients, setClients] = useState<Client[]>([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!token || !companyId || disabled) {
      setClients([]);
      return;
    }
    const handle = window.setTimeout(() => {
      setLoading(true);
      listCompanyClients(companyId, token, query)
        .then((rows) => setClients(rows))
        .catch(() => setClients([]))
        .finally(() => setLoading(false));
    }, query.trim() ? 180 : 0);
    return () => window.clearTimeout(handle);
  }, [companyId, disabled, query, token]);

  const selectedName = useMemo(() => {
    return clients.find((client) => client.client_id === value)?.name || currentName || "";
  }, [clients, currentName, value]);

  return (
    <Field label="Клиент">
      <div className="work-client-picker">
        {value ? (
          <div className="work-client-selected">
            <span>{selectedName || "Клиент выбран"}</span>
            {!disabled ? (
              <button type="button" onClick={() => onChange("")} aria-label="Убрать клиента">
                <X size={14} />
              </button>
            ) : null}
          </div>
        ) : null}
        <label className="work-client-search">
          <Search size={15} />
          <Input
            disabled={disabled}
            value={query}
            onFocus={() => setOpen(true)}
            onChange={(event) => {
              setQuery(event.target.value);
              setOpen(true);
            }}
            placeholder={value ? "Найти другого клиента" : "Найти клиента"}
          />
        </label>
        {open && !disabled ? (
          <div className="work-client-results">
            <button
              type="button"
              className={!value ? "is-active" : ""}
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => {
                onChange("");
                setQuery("");
                setOpen(false);
              }}
            >
              <strong>Без клиента</strong>
              <span>Задача останется внутренней или общей</span>
            </button>
            {clients.slice(0, 8).map((client, index) => (
              <button
                key={client.client_id}
                type="button"
                className={client.client_id === value ? "is-active" : ""}
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => {
                  onChange(client.client_id);
                  setQuery("");
                  setOpen(false);
                }}
              >
                <strong>{client.name}</strong>
                <span>{query.trim() ? "Найден по поиску" : index < 3 ? "В топе клиентов" : "Клиент компании"}</span>
              </button>
            ))}
            {!loading && clients.length === 0 ? <div className="work-client-empty">Клиенты не найдены</div> : null}
            {loading ? <div className="work-client-empty">Ищем клиента...</div> : null}
          </div>
        ) : null}
      </div>
    </Field>
  );
}
