"use client";
import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { Button } from "../../components/ui/button";
import { Badge } from "../../components/ui/badge";
import { usePulseStore } from "../../store/pulse-store";
import { deliverOfficeMail, ensureOffice, getOffice, setOfficeEnabled, type OfficeState } from "./api/officeApi";
import "./office-workflow.css";

export function OfficeWorkflowPanel({ projectId, onReady }: { projectId: string; onReady?: () => void }) {
  const companyId = usePulseStore((s) => s.activeCompanyId);
  const companies = usePulseStore((s) => s.companies);
  const demoCompany = companies.some((company) => company.company_id === companyId && company.slug?.startsWith("demo-"));
  const token = usePulseStore((s) => s.token);
  const [state, setState] = useState<OfficeState | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const load = useCallback(async () => {
    if (!token || !projectId || (!demoCompany && sessionStorage.getItem("pulse-demo-context") !== projectId)) return;
    let next = await getOffice(projectId, token);
    if (!next.configured) { next = await ensureOffice(projectId, token); onReady?.(); }
    sessionStorage.setItem("pulse-demo-context", projectId);
    setState(next); setError("");
  }, [onReady, projectId, token, demoCompany]);
  useEffect(() => {
    setState(null); setError("");
    void load().catch((err: Error) => setError(err.message));
    const timer = window.setInterval(() => void load().catch((err: Error) => setError(err.message)), 4000);
    return () => window.clearInterval(timer);
  }, [load]);
  const act = async (kind: "mail" | "toggle") => {
    setBusy(true); setError("");
    try {
      const next = kind === "mail" ? await deliverOfficeMail(projectId, token, crypto.randomUUID()) : await setOfficeEnabled(projectId, token, !state?.enabled);
      setState(next);
    } catch (err) { setError(err instanceof Error ? err.message : "Не удалось выполнить действие"); }
    finally { setBusy(false); }
  };
  if (!state?.configured) return error ? <p role="alert">{error}</p> : null;
  return <section className="office-workflow" aria-label="Почта и ИИ-отделы">
    <div className="office-workflow-heading"><div><h2>Почта → работа ИИ-отделов</h2><p>Новое письмо запускает классификацию, подготовку страницы и счёта. Вы подтверждаете изменения.</p></div><Badge tone={state.enabled ? "success" : "muted"}>{state.enabled ? "Опрос каждые 30 секунд" : "Опрос приостановлен"}</Badge></div>
    <div className="office-workflow-actions"><Button disabled={busy || state.pending_mail >= 30} onClick={() => void act("mail")}>Добавить учебное письмо</Button><Button variant="outline" disabled={busy} onClick={() => void act("toggle")}>{state.enabled ? "Приостановить" : "Возобновить"}</Button><Link href={`/workflows/${state.workflow_id}/runs`}>История исполнений</Link></div>
    <p className="muted">Последняя проверка: {state.last_poll_at ? new Date(state.last_poll_at).toLocaleString("ru-RU") : "ожидается"} · Проверок: {state.poll_count} · Новых писем: {state.pending_mail} · Ждут решения: {state.waiting}</p>
    <p className="muted">Храним последние 30 исполнений. Незавершённые не удаляются: при заполнении истории новые письма ждут в очереди.</p>
    {error && <p role="alert" className="form-error">{error}</p>}
    {state.artifacts?.length > 0 && <details><summary>Результаты в демо-компании ({state.artifacts.length})</summary><div className="office-artifacts">{state.artifacts.map((item) => <article key={item.workflow_run_id}>{item.page_json.title && <><strong>{item.page_json.title}</strong><p>{item.page_json.body}</p></>}{item.invoice_json.number && <p>Черновик счёта {item.invoice_json.number} · {((item.invoice_json.amount_minor ?? 0) / 100).toLocaleString("ru-RU")} ₽</p>}<Link href={`/runs/${item.workflow_run_id}`}>Как выполнено</Link></article>)}</div></details>}
    <small>Учебный ящик и сайт. Черновики счетов не отправляются в банк.</small>
  </section>;
}
