"use client";

import dynamic from "next/dynamic";
import { useState } from "react";
import { GitBranch, Hand, ShieldCheck } from "lucide-react";
import { asRecord } from "../aiEmployeePayloads";
import { graphLabels } from "./graphLabels";
import "./graph.css";

const AgentGraph = dynamic(() => import("./AgentGraph"), { ssr: false });
const number = (value: unknown) => typeof value === "number" && Number.isFinite(value) ? Math.max(0, value) : 0;
const text = (value: unknown) => typeof value === "string" ? value : "";
const toolStatus: Record<string, string> = { succeeded: "Выполнен", failed: "Ошибка", pending_approval: "Ожидает подтверждения", needs_approval: "Ожидает подтверждения", approved: "Подтверждён", rejected: "Отклонён", TOOL_CALL: "Вызван", TOOL_RESULT: "Результат получен", TOOL_CALL_STARTED: "Вызван", TOOL_CALL_FINISHED: "Результат получен", CARD_CREATED: "Предложение подготовлено", TOOL_PROPOSED: "Предложение подготовлено" };
const items = (value: unknown) => Array.isArray(value) ? value.map(asRecord) : [];

export function ExecutionTrace({ value }: { value: unknown }) {
  const [expanded, setExpanded] = useState(false);
  const trace = asRecord(value);
  if (trace.schema !== "pulse-execution-v1") return null;
  const waiting = trace.status === "waiting_approval";
  const calls = items(trace.model_calls);
  const nodes = items(trace.nodes);
  const input = number(trace.input_tokens), output = number(trace.output_tokens);
  return <details className="execution-trace" onToggle={(event) => setExpanded(event.currentTarget.open)}>
    <summary><GitBranch size={16} /><span>Как выполнен запрос</span><strong>{calls.length ? `${(input + output).toLocaleString("ru-RU")} токенов` : "Без вызова модели"}</strong></summary>
    <div className="execution-trace-body">
      <p className={waiting ? "trace-waiting" : "trace-completed"}>{waiting ? <Hand size={17} /> : <ShieldCheck size={17} />}{waiting ? "Ожидает вашего решения. Изменение ещё не применено." : "Выполнение завершено. Результат показан в ответе."}</p>
      <dl className="trace-metrics"><div><dt>Вызовы модели</dt><dd>{calls.length}</dd></div><div><dt>Входные токены</dt><dd>{input.toLocaleString("ru-RU")}</dd></div><div><dt>Выходные токены</dt><dd>{output.toLocaleString("ru-RU")}</dd></div></dl>
      <p className="trace-caption">Токены — единицы расхода модели из ответа провайдера. Ключи доступа здесь не показываются.</p>
      {calls.map((call, index) => <div className="trace-model-call" key={`${text(call.node)}-${index}`}><code>{text(call.node)}</code><span>{text(call.provider)} · {text(call.model)}</span><strong>{number(call.input_tokens)} → {number(call.output_tokens)}</strong></div>)}

      {items(trace.tools).length > 0 && <div className="trace-tools"><h4>Инструменты</h4>{items(trace.tools).map((tool, index) => <p key={index}><code>{text(tool.tool)}</code><span>{toolStatus[text(tool.status)] || toolStatus[text(tool.event)] || "Обработан"}</span></p>)}</div>}
      {expanded && <AgentGraph visited={nodes.map((node) => text(node.node))} waiting={waiting} compact />}
      <details className="trace-step-details"><summary>Пройденные шаги · {nodes.length}</summary><ol className="trace-node-list">{nodes.map((node, index) => <li key={`${text(node.node)}-${index}`}><span>{graphLabels[text(node.node)]?.title || text(node.node)}</span><code>{text(node.node)}</code><small>{node.status === "waiting" ? "Ожидание человека" : `${number(node.latency_ms)} мс`}</small></li>)}</ol></details>
    </div>
  </details>;
}
