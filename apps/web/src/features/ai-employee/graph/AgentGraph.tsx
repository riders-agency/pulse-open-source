"use client";

import { useMemo, useState } from "react";
import { Background, Controls, MarkerType, Position, ReactFlow } from "@xyflow/react";
import topology from "./agentGraph.generated.json";
import { graphAriaLabels, graphLabels, readPath, taskPath } from "./graphLabels";
import "@xyflow/react/dist/style.css";
import "./graph.css";

type Props = { visited?: string[]; waiting?: boolean; compact?: boolean };

export default function AgentGraph({ visited = [], waiting = false, compact = false }: Props) {
  const [view, setView] = useState(visited.length ? "actual" : "task");
  const [selected, setSelected] = useState(waiting ? "await_approval" : visited.at(-1) || "await_approval");
  const executed = [...(visited[0] === "ingest_turn" ? ["__start__"] : []), ...visited, ...(!waiting && visited.at(-1) === "finalizer" ? ["__end__"] : [])];
  const ordered = view === "actual" ? [...new Set(executed)].filter((id) => topology.nodes.includes(id)) : view === "task" ? taskPath : view === "read" ? readPath : topology.nodes;
  const nodes = useMemo(() => ordered.map((id, index) => ({
    id, position: { x: (Math.floor(index / 4) % 2 ? 3 - index % 4 : index % 4) * 230, y: Math.floor(index / 4) * 125 },
    sourcePosition: index % 4 === 3 ? Position.Bottom : Math.floor(index / 4) % 2 ? Position.Left : Position.Right,
    targetPosition: index % 4 === 0 && index > 0 ? Position.Top : Math.floor(index / 4) % 2 ? Position.Right : Position.Left,
    data: { label: <><strong>{graphLabels[id]?.title || id}</strong><code>{id}</code></> },
    className: `agent-graph-node agent-kind-${graphLabels[id]?.kind || "context"} ${visited.includes(id) ? "is-visited" : ""} ${id === "await_approval" && waiting ? "is-waiting" : ""}`,
    selected: id === selected,
  })), [ordered, selected, visited, waiting]);
  const edges = topology.edges.filter((edge) => ordered.includes(edge.source) && ordered.includes(edge.target) && (view !== "actual" || executed.some((node, index) => node === edge.source && executed[index + 1] === edge.target))).map((edge) => ({
    id: `${edge.source}-${edge.target}`, source: edge.source, target: edge.target,
    ariaLabel: `Переход: ${graphLabels[edge.source]?.title || edge.source} → ${graphLabels[edge.target]?.title || edge.target}`,
    type: "smoothstep", className: edge.conditional ? "agent-edge-conditional" : "",
    markerEnd: { type: MarkerType.ArrowClosed },
  }));
  const detail = graphLabels[selected];
  return <div className={`agent-graph ${compact ? "is-compact" : ""}`}>
    <div className="agent-graph-toolbar"><label>Путь в графе<select value={view} onChange={(event) => { setView(event.target.value); setSelected(event.target.value === "read" ? "read_context" : event.target.value === "actual" && !waiting ? visited.at(-1) || "finalizer" : "await_approval"); }}>{visited.length > 0 && <option value="actual">Путь этого запроса</option>}<option value="task">Создание задачи</option><option value="read">Чтение контекста</option><option value="all">Весь граф · {topology.nodes.length} узлов</option></select></label><span>Сплошная линия — переход · пунктир — условие</span></div>
    <div className="agent-graph-canvas" aria-label="Граф LangGraph из исходного кода">
      <ReactFlow key={view} nodes={nodes} edges={edges} ariaLabelConfig={graphAriaLabels} fitView fitViewOptions={{ minZoom: 0.5, maxZoom: 1 }} nodesDraggable={false} nodesConnectable={false} elementsSelectable onNodeClick={(_, node) => setSelected(node.id)} onSelectionChange={({ nodes: selection }) => { if (selection[0]) setSelected(selection[0].id); }} minZoom={0.25} maxZoom={1.8} preventScrolling={false} proOptions={{ hideAttribution: true }}>
        <Background gap={22} /><Controls showInteractive={false} />
      </ReactFlow>
    </div>
    <div className="agent-node-detail" aria-live="polite"><span>{detail?.kind === "approval" ? "Решение человека · Human-in-the-loop" : "Инструкция узла"}</span><h4>{detail?.title || selected}</h4><code>{selected}</code><p>{detail?.instruction}</p></div>
    <p className="agent-graph-note">{view === "actual" ? "Показаны только зафиксированные шаги и переходы этого запроса. " : "Показан возможный маршрут, а не журнал отдельного запроса. "}Узлы и доступные переходы экспортированы из исполняемого Python-графа.</p>
  </div>;
}
