import { ArrowUpRight } from "lucide-react";
import { sourceRoot } from "./scenarios";

const features = [
  ["01", "ИИ-сотрудник на LangGraph", "Просьбы обычными словами, контекст проекта и инструменты. Граф останавливается перед изменениями и продолжает работу после вашего подтверждения.", "apps/ai-agent/src/pulse_agent/graph"],
  ["02", "Задачи и доски", "Статусы, исполнители, приоритеты, сроки и комментарии. Задачу можно удалить и восстановить вместе с историей.", "apps/api/internal/modules/worktracker"],
  ["03", "Процессы и учёт времени", "Шаблоны работы, этапы и таймеры. Видно, кто выполняет задачу и сколько времени на неё потрачено.", "apps/web/src/features/work"],
  ["04", "Визуальные сценарии", "Триггеры, условия и действия в редакторе графа. Черновики, публикация версий и отключение отдельного шага без потери настроек.", "apps/web/src/features/workflow-editor"],
  ["05", "Чат на сайте", "Виджет для посетителя и рабочее место оператора. Переписка, вложения и контекст; при сбое текст остаётся для повторной отправки.", "apps/web/src/features/site-chat-console"],
  ["06", "Клиенты и услуги", "Карточка клиента, его услуги, задачи и документы. Команда работает с общей историей в контексте своего проекта.", "apps/api/internal/modules/clients"],
  ["07", "Оплаты и документы", "График оплат, статусы и банковские черновики со сроком и назначением. Поддержан контракт счёта и платёжной ссылки; выпуск через банк ещё не подключён.", "docs/BANK_INVOICES.md"],
  ["08", "Команда и права", "Компании, проекты и роли. Сервер проверяет доступ к данным и каждому действию, включая инструменты ИИ-сотрудника.", "apps/api/internal/platform/authz"],
  ["09", "Интеграции и вебхуки", "Каталог действий, входящие события, почта, мессенджеры и CRM. Реальные отправки требуют отдельного подключения провайдера.", "apps/api/internal/modules/providers"],
  ["10", "История выполнения", "Шаги сценария, ошибки, результаты и повторы. У ИИ видны пройденный граф, вызовы инструментов и фактический расход токенов.", "apps/api/internal/runtime"],
  ["11", "SDK и фреймворки", "TypeScript, Python, Go, PHP и другие языки. Примеры подключения для React, Next.js, Vue, FastAPI, Django, Laravel и других фреймворков.", "sdk"],
  ["12", "Свой сервер и открытый код", "React и Next.js, Go, Python и PostgreSQL. Исходники, контракты OpenAPI/gRPC и инструкции для самостоятельного запуска.", "README.md"],
];

export function TechFeatures() {
  return (
    <section id="features" className="tech-section">
      <div className="tech-section-heading"><span className="tech-eyebrow">Возможности продукта</span><h2>Работа команды — в одном контексте</h2><p>От обращения посетителя до задачи, сценария и оплаты. У каждого раздела есть интерфейс, серверная логика и границы доступа.</p></div>
      <div className="tech-features">
        {features.map(([number, title, text, path]) => (
          <article key={number}><span className="tech-feature-number">{number}</span><div><h3>{title}</h3><p>{text}</p><a href={`${sourceRoot}/${path}`}>Как устроено <ArrowUpRight size={14} /></a></div></article>
        ))}
      </div>
      <a className="tech-reading-link" href={`${sourceRoot}/docs/FEATURES.md`}>Все возможности и текущая готовность <ArrowUpRight size={16} /></a>
    </section>
  );
}
