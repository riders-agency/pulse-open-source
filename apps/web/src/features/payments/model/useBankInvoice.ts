import { useEffect, useRef, useState } from "react";
import { ApiError } from "../../../lib/api/http";
import { toDateInput, toISODate } from "../../../lib/dates";
import { readBankInvoice, saveBankInvoice, type BankInvoiceInput, type BankInvoiceWorkspace } from "../api/bankInvoicesApi";

const empty: BankInvoiceInput = { revision: 0, kind: "bank_invoice", payer_name: "", tax_id: "", purpose: "", due_date: "" };

export function useBankInvoice(id: string, token: string) {
  const [workspace, setWorkspace] = useState<BankInvoiceWorkspace | null>(null);
  const [input, setInput] = useState(empty);
  const [error, setError] = useState("");
  const [saved, setSaved] = useState(false);
  const [busy, setBusy] = useState(false);
  const [conflict, setConflict] = useState(false);
  const [reload, setReload] = useState(0);
  const request = useRef<{ body: string; key: string } | null>(null);
  const saving = useRef(false);
  const controller = useRef<AbortController | null>(null);

  useEffect(() => {
    const abort = new AbortController();
    controller.current = abort;
    readBankInvoice(id, token, abort.signal).then((data) => {
      if (abort.signal.aborted) return;
      setWorkspace(data);
      setInput(data.draft ? {
        revision: data.draft.revision, kind: data.draft.kind, payer_name: data.draft.payer_name,
        tax_id: data.draft.tax_id, purpose: data.draft.purpose, due_date: toDateInput(data.draft.due_date),
      } : { ...empty, payer_name: data.payer_name });
      setConflict(false); setError(""); request.current = null;
    }).catch((err) => { if (!abort.signal.aborted) setError(err.message); });
    return () => abort.abort();
  }, [id, token, reload]);

  function change<K extends keyof BankInvoiceInput>(key: K, value: BankInvoiceInput[K]) {
    setInput((current) => ({ ...current, [key]: value })); setSaved(false);
  }

  async function save() {
    if (saving.current || !controller.current || conflict) return;
    const payload = { ...input, due_date: toISODate(input.due_date) };
    if (input.due_date && !payload.due_date) { setError("Введите корректную дату в формате дд.мм.гггг"); return; }
    const signal = controller.current.signal;
    saving.current = true; setBusy(true); setError(""); setSaved(false);
    const body = JSON.stringify(payload);
    if (request.current?.body !== body) request.current = { body, key: crypto.randomUUID() };
    try {
      const draft = await saveBankInvoice(id, token, payload, request.current.key, signal);
      if (signal.aborted) return;
      setInput((current) => ({ ...current, revision: draft.revision }));
      setWorkspace((current) => current ? { ...current, draft } : null);
      request.current = null; setSaved(true);
    } catch (err) {
      if (signal.aborted) return;
      setConflict(err instanceof ApiError && err.status === 409);
      setError(err instanceof Error ? err.message : "Не удалось сохранить черновик");
    } finally {
      saving.current = false;
      if (!signal.aborted) setBusy(false);
    }
  }

  return { workspace, input, error, saved, busy, conflict, change, save, refresh: () => { setSaved(false); setReload((v) => v + 1); } };
}
