"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Activity, ArrowRight, Check } from "lucide-react";
import { startDemo } from "./startDemo";
import "./demo.css";

export function DemoWelcome({ enabled, autoStart = false }: { enabled: boolean; autoStart?: boolean }) {
  const router = useRouter();
  const [busy, setBusy] = useState(enabled && autoStart);
  const [error, setError] = useState("");
  const enter = async () => {
    setBusy(true); setError("");
    try { await startDemo(); router.replace("/tasks/all"); }
    catch (err) { setError(err instanceof Error ? err.message : "Не удалось открыть демо. Попробуйте ещё раз."); setBusy(false); }
  };
  useEffect(() => {
    if (!enabled || !autoStart) return;
    let active = true;
    void startDemo().then(() => { if (active) router.replace("/tasks/all"); }).catch((err: unknown) => {
      if (!active) return;
      setError(err instanceof Error ? err.message : "Не удалось открыть демо. Попробуйте ещё раз.");
      setBusy(false);
    });
    return () => { active = false; };
  }, [enabled, autoStart, router]);
  return (
    <main className="demo-page">
      <Link href="/tech" className="demo-logo"><Activity size={24} /> pulse.</Link>
      <div className="demo-welcome"><span className="demo-eyebrow">Платформа для ИИ-агентов</span><h1>Попробуйте Pulse<br />в своём пространстве</h1><p>Задачи уже на доске. Попросите агента составить план или создать задачу — и подтвердите результат прямо в чате.</p>
        <ul><li><Check size={18} /> Без личной почты и пароля</li><li><Check size={18} /> Отдельный проект для каждого гостя</li><li><Check size={18} /> Можно создавать задачи прямо в чате</li></ul>
        {enabled ? <button className="demo-enter" onClick={() => void enter()} disabled={busy}>{busy ? "Готовим ваше пространство…" : "Открыть демо"}<ArrowRight size={18} /></button> : <p>Демо на этой установке отключено. <Link href="/tech">Посмотреть записи сценариев</Link></p>}
        {error && <p className="demo-error" role="alert">{error}</p>}
        <details className="demo-note"><summary>Как устроено демо</summary><p>Используйте вымышленные данные. Запросы ИИ обрабатывает APInet; личные данные и ключи вводить не нужно. Внешние отправки, загрузка файлов и изменение подключений отключены. Доступно до 30 обращений к модели на гостя в сутки, общий лимит демо — 300.</p></details>
      </div>
      <aside className="demo-prompts"><span className="demo-eyebrow">Начните с простого</span><h2>Попросите ИИ-сотрудника</h2><blockquote>«Создай задачу подготовить презентацию проекта на завтра»</blockquote><blockquote>«Какие задачи сейчас в проекте?»</blockquote><blockquote>«Помоги составить план запуска нового продукта»</blockquote><p>Откройте «ИИ сотрудник» в правом нижнем углу. Карточку задачи можно изменить до подтверждения.</p></aside>
    </main>
  );
}
