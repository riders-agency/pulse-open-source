import Link from "next/link";
import { ArrowUpRight } from "lucide-react";

export function OfficeScenario() {
  return <section id="office-scenario" className="tech-section">
    <div className="tech-section-heading"><span className="tech-eyebrow">Агенты работают вместе</span><h2>Пришло письмо.<br />Отделы уже за работой.</h2><p>Один агент читает почту и распределяет обращения. ИИ-отдел готовит страницу сайта, бухгалтерия — счёт. Вы подтверждаете результат.</p></div>
    <figure className="tech-task-recording"><video controls playsInline preload="none" poster="/tech/recordings/office.webp" aria-label="Запись: почта и работа ИИ-отделов"><source src="/tech/recordings/office.mp4" type="video/mp4" /><track default kind="captions" src="/tech/recordings/office.vtt" srcLang="ru" label="Русский" /><a href="/tech/recordings/office.mp4">Скачать запись</a></video><figcaption><span>Исполненный LangGraph · учебные почта и сайт · счёт без отправки в банк</span><Link href="/tech/office">Разобрать запуск <ArrowUpRight size={14} /></Link></figcaption></figure>
  </section>;
}
