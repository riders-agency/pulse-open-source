"use client";

import { Handle, Position, type Edge, type Node, type NodeProps } from "@xyflow/react";

import { Badge } from "../../../components/ui/badge";
import type { ActionRun } from "../../../lib/api";
import type { WorkflowRunGraph } from "../api/workflowRunApi";
import { diagnosticsForAction } from "../model/workflowRunDiagnostics";
import { actionTypeLabel, statusLabel, statusTone } from "../model/workflowRunLabels";

export type RuntimeNodeData = {
  label: string;
  action_key: string;
  action_type_key: string;
  status: string;
  attempts: string;
  reason?: string;
};

export type RuntimeNodeType = Node<RuntimeNodeData, "runtime">;

function runtimeNodeClassName(status?: string) {
  const tone = statusTone(status);
  return `flow-node flow-node-${tone}`;
}

function RuntimeNode({ data }: NodeProps<RuntimeNodeType>) {
  return (
    <div className={runtimeNodeClassName(data.status)}>
      <Handle type="target" position={Position.Left} />
      <div className="flow-node-title">{data.label}</div>
      <small>{actionTypeLabel(data.action_type_key)}</small>
      {data.reason ? <div className="muted mt-1 line-clamp-2 text-xs">{data.reason}</div> : null}
      <div className="row runtime-node-badges">
        <Badge tone={statusTone(data.status)}>{statusLabel(data.status)}</Badge>
        <Badge>{data.attempts}</Badge>
      </div>
      <Handle type="source" position={Position.Right} />
    </div>
  );
}

export const nodeTypes = { runtime: RuntimeNode };

export function buildRuntimeNodes(actions: ActionRun[], graph: WorkflowRunGraph | null): RuntimeNodeType[] {
  const graphNodes = graph?.graph?.nodes;
  return graphNodes?.length
    ? graphNodes.map((node, index): RuntimeNodeType => {
        const action = actions.find((item) => item.action_key === node.id);
        const diagnostics = diagnosticsForAction(action);
        return {
          id: node.id,
          type: "runtime",
          position: node.position ?? { x: 120 + index * 240, y: 180 },
          data: {
            label: node.data?.label ?? node.id,
            action_key: node.id,
            action_type_key: action?.action_type_key ?? node.data?.action_type_key ?? "unknown",
            status: action?.status ?? "not_started",
            attempts: `${action?.attempt_count ?? 0}/${action?.max_attempts ?? 10}`,
            reason: diagnostics?.summary || diagnostics?.reason_code,
          },
        };
      })
    : actions.map((action, index): RuntimeNodeType => ({
        id: action.action_key || `action-${index}`,
        type: "runtime",
        position: { x: 120 + index * 240, y: 180 },
        data: {
          label: action.action_key || "Действие",
          action_key: action.action_key || "",
          action_type_key: action.action_type_key || "unknown",
          status: action.status || "not_started",
          attempts: `${action.attempt_count ?? 0}/${action.max_attempts ?? 10}`,
        },
      }));
}

export function buildRuntimeEdges(nodes: RuntimeNodeType[], graph: WorkflowRunGraph | null): Edge[] {
  const graphEdges = graph?.graph?.edges as Edge[] | undefined;
  if (graphEdges?.length) return graphEdges;
  return nodes.slice(1).map((node, index) => ({ id: `e-${index}`, source: nodes[index].id, target: node.id }));
}
