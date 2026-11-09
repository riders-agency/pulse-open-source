"use client";

import { Background, MarkerType, Position, ReactFlow } from "@xyflow/react";
import { useCallback, useMemo } from "react";
import { graphAriaLabels } from "../ai-employee/graph/graphLabels";
import type { OfficeState } from "../../lib/api/officeDemo";
import "@xyflow/react/dist/style.css";

export function OfficeFlow({ state, selected, select }: { state: OfficeState; selected: string; select: (id: string) => void }) {
  const nodes = useMemo(() => state.nodes.map((node, index) => {
    const status = state.events.findLast((event) => event.node === node.id)?.status || "pending";
    return {
      id: node.id, position: { x: 40, y: index * 78 }, sourcePosition: Position.Bottom, targetPosition: Position.Top,
      selected: selected === node.id, className: `office-flow-node is-${status}`,
      data: { label: <><strong>{node.title}</strong><small>{node.tool}</small><em>{status === "completed" ? "✓ Готово" : status === "waiting" ? "Ждёт вас" : status === "running" ? "В работе" : status === "skipped" ? "Пропущен" : ""}</em></> },
    };
  }), [state.nodes, state.events, selected]);
  const edges = useMemo(() => {
    const ids = new Set(state.nodes.map((node) => node.id));
    const visited = state.events.filter((event) => event.status === "completed").map((event) => event.node);
    return state.edges.filter(([from, to]) => ids.has(from) && ids.has(to) && (from !== "approval" || to !== "done" || state.status === "rejected")).map(([from, to]) => ({
    id: `${from}-${to}`, source: from, target: to, type: "smoothstep",
    label: from === "approval" ? to === "apply" ? "Подтверждено" : "Отклонено" : undefined,
    className: visited.some((node, index) => node === from && visited[index + 1] === to) ? "is-traversed" : "",
    markerEnd: { type: MarkerType.ArrowClosed }, ariaLabel: `Переход: ${from} → ${to}`,
    }));
  }, [state.edges, state.events, state.nodes, state.status]);
  const choose = useCallback((_: unknown, node: { id: string }) => select(node.id), [select]);
  return <div className="office-flow"><ReactFlow nodes={nodes} edges={edges} fitView fitViewOptions={{ padding: .08, maxZoom: 1 }} nodesDraggable={false} nodesConnectable={false} panOnDrag={false} zoomOnScroll={false} zoomOnDoubleClick={false} preventScrolling={false} onNodeClick={choose} ariaLabelConfig={graphAriaLabels} proOptions={{ hideAttribution: true }}><Background gap={18} /></ReactFlow></div>;
}
