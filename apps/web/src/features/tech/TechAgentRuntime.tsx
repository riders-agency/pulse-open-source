"use client";

import dynamic from "next/dynamic";
import Image from "next/image";
import { useState } from "react";
import { ArrowUpRight, CheckCircle2, ChevronDown } from "lucide-react";
import { sourceRoot } from "./scenarios";
import recording from "./taskRecording.generated.json";
import "./agent-runtime.css";

const AgentGraph = dynamic(() => import("../ai-employee/graph/AgentGraph"), { ssr: false, loading: () => <p>Загружаем граф запроса…</p> });

export function TechAgentRuntime() {
  const [phase, setPhase] = useState<"proposal" | "completion">("completion");
  const trace = recording[phase];
  return <section id="agent-runtime" className="tech-section tech-recorded-runtime">
    <div className="tech-graph-engine"><a href="https://www.langchain.com/langgraph" target="_blank" rel="noreferrer" aria-label="LangGraph — фреймворк исполнения графа"><Image src="/tech/brands/langgraph.svg" alt="LangGraph" width={189} height={40} /></a><p>Исполняет граф ИИ-агента</p></div>
    <div className="tech-section-heading"><h2>А вот как это выполнено</h2><p>LangGraph исполняет граф в Python: разбирает запрос, сохраняет состояние и ждёт подтверждения человека. Здесь показан путь запроса из записи. Нажмите на шаг, чтобы увидеть его инструкцию.</p></div>
    <div className="tech-run-receipt"><span><CheckCircle2 size={18} /> Задача создана</span><code>{recording.tool.name}</code><span>После «{recording.confirmation}»</span></div>
    <div className="tech-run-phases" role="group" aria-label="Этапы записанного запроса"><button type="button" aria-pressed={phase === "proposal"} onClick={() => setPhase("proposal")}>1. Подготовка и ожидание</button><button type="button" aria-pressed={phase === "completion"} onClick={() => setPhase("completion")}>2. Подтверждение и результат</button></div>
    <AgentGraph key={phase} visited={trace.nodes.map((node) => node.node)} waiting={trace.status === "waiting_approval"} />
    <dl className="tech-run-usage"><div><dt>Модель при разборе просьбы</dt><dd>{recording.proposal.model_calls.map((call) => call.model).join(", ")}</dd></div><div><dt>Входные токены</dt><dd>{recording.proposal.input_tokens.toLocaleString("ru-RU")}</dd></div><div><dt>Выходные токены</dt><dd>{recording.proposal.output_tokens.toLocaleString("ru-RU")}</dd></div></dl>
    <p className="tech-run-caption">Расход из ответа провайдера в этой записи. Граф построен по журналу выполнения Python / LangGraph.</p>
    <details className="tech-run-controls"><summary>Почему сотрудник не может самовольно изменить данные <ChevronDown size={16} /></summary><div><p>Вход определяет пользователя и проект. Модель выбирает действие из разрешённого каталога, но не получает доступ к базе или ключам.</p><p>Перед записью LangGraph сохраняет состояние и ждёт решения человека. После подтверждения Go повторно проверяет права, поля и состояние действия. Только затем создаёт задачу и возвращает её в чат.</p><a href={`${sourceRoot}/docs/AI_CONTROL.md`}>Инструкции, авторизация и контроль в коде <ArrowUpRight size={15} /></a></div></details>
  </section>;
}
