"use client";

import { useState } from "react";
import dynamic from "next/dynamic";
import type { OfficeState } from "../../lib/api/officeDemo";

const OfficeFlow = dynamic(() => import("./OfficeFlow").then((module) => module.OfficeFlow), { ssr: false });

export function OfficeGraph({ state }: { state: OfficeState }) {
  const [selected, setSelected] = useState("");
  const current = state.events.at(-1)?.node;
  const node = state.nodes.find((item) => item.id === (selected || current)) ?? state.nodes[0];
  const event = state.events.findLast((item) => item.node === node?.id);
  const calls = state.events.flatMap((item) => item.usage ? [item.usage] : []);
  return <aside className="office-graph" aria-label="Исполняемый граф отделов">
    <header><h2>Как выполняется просьба</h2><span>LangGraph</span></header>
    <OfficeFlow state={state} selected={node?.id || ""} select={setSelected} />
    {node && <div className="office-node-detail"><strong>{node.title}</strong><p>{node.detail}</p><code>{node.tool}</code>{event && <p>{event.text}</p>}</div>}
    {calls.length > 0 && <div className="office-usage"><span>{[...new Set(calls.map((call) => call.model_id))].join(", ")}</span><span>{calls.reduce((sum, call) => sum + call.input_tokens, 0)} входных · {calls.reduce((sum, call) => sum + call.output_tokens, 0)} выходных токенов</span><small>{calls.length} вызова модели · из ответов провайдера</small></div>}
  </aside>;
}
