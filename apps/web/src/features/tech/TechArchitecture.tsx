import { ArrowDown, ArrowRight, ArrowUpRight, Braces, Database, Network, ShieldCheck } from "lucide-react";
import { sourceRoot } from "./scenarios";

const layers = [
  { icon: Braces, title: "Интерфейс", stack: "Next.js · React · TypeScript", text: "Чат, визуальный граф, карточки действий. Пользователь видит предложение и подтверждает изменения.", path: "apps/web/src/features/ai-employee" },
  { icon: ShieldCheck, title: "Прикладной сервер", stack: "Go · OpenAPI · PostgreSQL", text: "Проверка доступа, команды, транзакции и журнал действий. Здесь принимается решение о записи данных.", path: "apps/api/internal/modules/ai" },
  { icon: Network, title: "ИИ-сотрудник", stack: "Python · LangGraph · gRPC", text: "Разбор запроса, поиск контекста, подготовка предложения и возобновление после подтверждения.", path: "apps/ai-agent/src/pulse_agent" },
];

export function TechArchitecture() {
  return (
    <section id="architecture" className="tech-section">
      <div className="tech-section-heading"><span className="tech-eyebrow">Архитектура</span><h2>У каждого слоя своя работа</h2><p>Типизированные контракты соединяют интерфейс, бизнес-логику и ИИ. Ключи провайдеров не передаются в браузер.</p></div>
      <div className="tech-layers">
        {layers.map(({ icon: Icon, ...layer }, index) => (
          <article className="tech-layer" key={layer.title}>
            <span className="tech-layer-index">0{index + 1} / <Icon size={19} /></span>
            <h3>{layer.title}</h3><strong>{layer.stack}</strong><p>{layer.text}</p>
            <a href={`${sourceRoot}/${layer.path}`} target="_blank" rel="noreferrer">Исходный код <ArrowUpRight size={15} /></a>
          </article>
        ))}
      </div>
      <div className="tech-flow" aria-label="Путь команды">
        <span>Запрос в чате</span><ArrowRight size={16} /><span>StartTurn · gRPC</span><ArrowRight size={16} /><span>Предложение</span><ArrowRight size={16} /><span>Подтверждение</span><ArrowRight size={16} /><span><Database size={16} /> Запись в БД</span>
      </div>
      <div className="tech-engineering">
        <div><h3>Проверяемое действие</h3><p>Инструмент описывает вход, права и необходимость подтверждения. Ответ содержит результат и ссылку на созданную сущность.</p><a href={`${sourceRoot}/contracts/proto/ai_agent/v1/pulse_tools.proto`}>Контракт инструментов <ArrowUpRight size={15} /></a></div>
        <div><h3>Продолжение после паузы</h3><p>LangGraph сохраняет состояние в PostgreSQL. Подтверждение возобновляет выполнение; разрыв соединения не означает, что действие не произошло.</p><a href={`${sourceRoot}/apps/ai-agent/src/pulse_agent/checkpointing.py`}>Хранение состояния <ArrowUpRight size={15} /></a></div>
        <div><h3>Надёжное исполнение</h3><p>Worker забирает работу с ограниченной арендой. Журнал эффектов и очередь исходящих событий помогают контролировать повторы.</p><a href={`${sourceRoot}/docs/architecture/effect-ledger.md`}>Журнал эффектов <ArrowUpRight size={15} /></a></div>
      </div>
      <details className="tech-boundaries"><summary>Границы текущей реализации <ArrowDown size={16} /></summary><p>Видеозаписи используют вымышленные данные. В открытом демо подключён APInet: можно получить совет и создать задачу после подтверждения. Данные гостей разделены, обращения к модели ограничены. Запись ИИ-сценария показывает просьбу в чате, подтверждение словами и созданную задачу. Затем раскрывается граф этого выполнения и фактический расход токенов. Банковский интерфейс пока сохраняет черновики: адаптеры банков ещё не реализованы.</p><p>Сервер — модульный монолит с командами владельцев данных и контрактами чтения. Транзакция объединяет действие, результат подтверждения и очередь событий. Проверки границ модулей и конкурентного выполнения входят в репозиторий.</p><a href={`${sourceRoot}/docs/LIMITATIONS.md`}>Ограничения и состояние проекта <ArrowUpRight size={15} /></a></details>
    </section>
  );
}
