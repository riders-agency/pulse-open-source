import { ArrowDown, ArrowUpRight } from "lucide-react";
import { sourceRoot } from "./scenarios";

const decisions = [
  { area: "React / TypeScript", title: "Интерфейс учитывает сбои сети", text: "Общий HTTP-транспорт управляет отменой, тайм-аутами и повторами чтения. Ответ старого диалога не должен попасть в новый. Черновик остаётся у пользователя после ошибки отправки.", detail: "Команды записи не повторяются транспортом вслепую. Обновление сессии объединяет одновременные запросы. Публичное описание рендерится на сервере; клиентский JavaScript нужен для взаимодействий.", path: "apps/web/src/lib/api/http-transport.ts", code: "AbortSignal · CSRF · single-flight refresh" },
  { area: "Go / PostgreSQL", title: "Правила выдерживают конкурирующие запросы", text: "Удаление задачи проверяет права, блокирует строку и сохраняет изменение вместе с историей. Работающий таймер не позволяет удалить задачу.", detail: "Проверка после SELECT FOR UPDATE закрывает гонку между удалением и запуском таймера. Повтор уже выполненной команды не дублирует событие. Поведение проверяется с настоящим PostgreSQL.", path: "apps/api/internal/modules/worktracker/service/task_archive.go", code: "BEGIN → FOR UPDATE → проверка → COMMIT" },
  { area: "Python / LangGraph / gRPC", title: "Предложение ИИ становится управляемым действием", text: "Граф отделяет разбор запроса, контекст, предложение и ожидание подтверждения. Python обращается к инструментам Go по контракту; права проверяет сервер.", detail: "StartTurn передаёт поток событий, ResumeTurn продолжает диалог после решения пользователя. Checkpoint хранит состояние. Ответ инструмента возвращает результат и ссылку на сущность.", path: "apps/ai-agent/src/pulse_agent/graph/builder.py", code: "StartTurn → предложение → ResumeTurn → результат" },
  { area: "Go worker / интеграции", title: "Потерянный ответ — отдельное состояние", text: "Worker получает работу с ограниченной арендой. Журнал эффектов различает успех, ошибку и неопределённый результат, когда внешний сервис мог уже выполнить действие.", detail: "Lease и fencing token контролируют владение заданием. Для внешнего API нужны идемпотентность и сверка; универсальная гарантия «ровно один раз» не заявляется. Поэтому будущий банковский адаптер предусматривает Lookup.", path: "apps/api/internal/runtime/effects/ledger.go", code: "started / succeeded / failed / unknown" },
];

export function TechCodeTour() {
  return (
    <section id="engineering" className="tech-section">
      <div className="tech-section-heading"><span className="tech-eyebrow">Для технического знакомства</span><h2>Решения, которые можно проверить в коде</h2><p>Четыре примера: состояние React, транзакции Go, граф Python и поведение внешних операций при сбое.</p></div>
      <div className="tech-code-tour">
        {decisions.map((item) => (
          <article key={item.area}><span className="tech-eyebrow">{item.area}</span><h3>{item.title}</h3><p>{item.text}</p><code>{item.code}</code><details><summary>Разобрать решение <ArrowDown size={14} /></summary><p>{item.detail}</p><a href={`${sourceRoot}/${item.path}`}>Открыть реализацию <ArrowUpRight size={14} /></a></details></article>
        ))}
      </div>
      <div className="tech-reading"><div><h3>15 минут с исходниками</h3><p>Маршрут для ревью: файлы, вопросы к реализации, тесты и известные ограничения.</p></div><a className="tech-cta" href={`${sourceRoot}/docs/CODE_TOUR.md`}>Перейти к разбору кода <ArrowUpRight size={16} /></a></div>
      <a className="tech-reading-link" href={`${sourceRoot}/sdk/README.md`}>SDK и интеграции с фреймворками <ArrowUpRight size={16} /></a>
    </section>
  );
}
