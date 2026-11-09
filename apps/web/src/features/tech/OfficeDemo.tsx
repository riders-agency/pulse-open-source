"use client";

import Link from "next/link";
import { useEffect, useRef } from "react";
import { Activity, ArrowLeft, Mail, Play, RotateCcw } from "lucide-react";
import { useOfficeDemo } from "./useOfficeDemo";
import { OfficeGraph } from "./OfficeGraph";
import type { OfficeArtifact } from "../../lib/api/officeDemo";
import "./office-demo.css";

function Artifacts({ values, preview = false }: { values: Record<string, OfficeArtifact>; preview?: boolean }) {
  return <div className="office-artifacts" aria-label={preview ? "На подтверждение" : "Сохранённые результаты"}>
    {values.website && <article><small>{preview ? "Будет опубликовано на учебном сайте" : "Опубликовано на учебном сайте"}</small><h3>{values.website.title}</h3><p>{values.website.body}</p><code>{values.website.path}</code></article>}
    {values.invoice && <article><small>Черновик счёта · банк не подключён</small><h3>Счёт {values.invoice.number}</h3><p>{values.invoice.payer}</p><strong>{((values.invoice.amount_minor || 0) / 100).toLocaleString("ru-RU")} ₽</strong></article>}
  </div>;
}

export function OfficeDemo({ endpoint }: { endpoint: string | null }) {
  const { state, busy, error, act, restart } = useOfficeDemo(endpoint);
  const log = useRef<HTMLDivElement>(null);
  useEffect(() => { const element = log.current; if (element) element.scrollTop = element.scrollHeight; }, [state.events.length]);
  const pending = state.status === "waiting_approval";
  const drafts = state.events.findLast((event) => event.drafts)?.drafts;
  return <main className="office-demo">
    <header className="office-header"><Link href="/tech"><ArrowLeft size={17} /> Pulse</Link><span>Платформа для ИИ-агентов</span><small>{endpoint ? "Живой запуск · учебные данные" : "Записанный запуск · учебные данные"}</small></header>
    <div className="office-layout"><section className="office-conversation">
      <header><Activity size={24} /><div><h1>Письмо → работа отделов</h1><p>Агенты готовят. Вы принимаете решение.</p></div></header>
      <div className="office-messages" role="log" aria-label="Работа ИИ-отделов" ref={log}>
        <div className="office-message is-human"><span>Вы</span><p>Следи за входящей почтой. Запросы на сайт передавай ИИ-отделу, счета — бухгалтерии. Перед изменениями спроси меня.</p></div>
        {state.events.map((event, index) => <div key={index} className={`office-message ${event.actor === "human" ? "is-human" : ""}`} data-node={event.node}>
          <span>{event.actor === "human" ? "Вы" : state.nodes.find((node) => node.id === event.node)?.title || "Агент"}</span><p>{event.text}</p>
          {event.mail && <blockquote><strong><Mail size={14} /> {event.mail.sender}</strong><p>{event.mail.body}</p></blockquote>}
          {event.tool && <small>{event.tool}</small>}
        </div>)}
        {pending && drafts && <Artifacts values={drafts} preview />}
        {Object.keys(state.results).length > 0 && <Artifacts values={state.results} />}
      </div>
      <footer>{error && <p role="alert">{error}</p>}
        {endpoint ? <div className="office-actions">
          {state.status === "idle" && <button disabled={busy} onClick={() => void act("start")}><Play size={16} /> Начать наблюдение</button>}
          {state.status === "watching" && <button disabled={busy} onClick={() => void act("deliver")}><Mail size={16} /> Получить учебное письмо</button>}
          {pending && <><button disabled={busy} onClick={() => void act("decision", { approve: true })}>Подтвердить изменения</button><button className="is-secondary" disabled={busy} onClick={() => void act("decision", { approve: false })}>Отклонить</button></>}
          {state.status === "running" && <p role="status">Агенты работают…</p>}
          {state.status === "completed" && <><p role="status">Результаты сохранены. Письмо обработано.</p><a href={`${endpoint}/site`} target="_blank" rel="noreferrer">Открыть учебный сайт ↗</a></>}
        </div> : <button onClick={restart} disabled={state.status === "replaying"}><RotateCcw size={16} /> Воспроизвести журнал</button>}
        <small>Учебный ящик и сайт. Счёт сохраняется как черновик; отправки в банк нет.</small>
      </footer>
    </section><OfficeGraph state={state} /></div>
  </main>;
}
