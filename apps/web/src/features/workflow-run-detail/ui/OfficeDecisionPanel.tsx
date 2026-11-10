"use client";
import { useState } from "react";
import { Button } from "../../../components/ui/button";
import { usePulseStore } from "../../../store/pulse-store";
import { decideOfficeRun } from "../../project-workflows/api/officeApi";
import "../../project-workflows/office-workflow.css";

export function OfficeDecisionPanel({ projectId, runId, output }: { projectId: string; runId: string; output: unknown }) {
  const token = usePulseStore((s) => s.token);
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState("");
  const value = output as { result?: { page?: { title?: string; body?: string }; invoice?: { number?: string; amount_minor?: number } } };
  const result = value?.result;
  const decide = async (approved: boolean) => {
    setBusy(true); setError("");
    try { await decideOfficeRun(projectId, token, runId, approved); setDone(true); }
    catch (err) { setError(err instanceof Error ? err.message : "Не удалось сохранить решение"); }
    finally { setBusy(false); }
  };
  return <section className="office-decision" aria-label="Подтверждение изменений"><strong>Ваше решение</strong>
    {result?.page?.title && <><p><b>{result.page.title}</b></p><p>{result.page.body}</p></>}
    {result?.invoice?.number && <p>Черновик счёта {result.invoice.number}: {((result.invoice.amount_minor ?? 0) / 100).toLocaleString("ru-RU")} ₽</p>}
    <p>Страница появится на учебном сайте, счёт сохранится черновиком. Отправки в банк нет.</p>
    {done ? <p role="status">Решение принято. Сценарий продолжает работу.</p> : <div className="row"><Button disabled={busy} onClick={() => void decide(true)}>Подтвердить изменения</Button><Button disabled={busy} variant="outline" onClick={() => void decide(false)}>Отклонить</Button></div>}
    {error && <p role="alert">{error}</p>}
  </section>;
}
