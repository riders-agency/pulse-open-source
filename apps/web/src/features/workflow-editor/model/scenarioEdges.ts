import { MarkerType, type Edge } from "@xyflow/react";
import type { ActionNodeType, ScenarioConnectionJSON } from "./types";

export function branchFromConnection(value: ScenarioConnectionJSON) {
  return value.branch ?? value.label ?? "";
}

export function branchFromEdge(edge: Edge) {
  return String(edge.sourceHandle ?? edge.label ?? "");
}

export function edgeFromScenarioConnection(edge: ScenarioConnectionJSON, index: number, idPrefix = "imported-connection"): Edge | null {
  if (!edge.from || !edge.to) return null;
  const branch = branchFromConnection(edge);
  return {
    id: `${idPrefix}-${index}-${edge.from}-${edge.to}-${branch || "default"}`,
    source: edge.from,
    target: edge.to,
    sourceHandle: branch || undefined,
    label: branch || undefined,
    markerEnd: { type: MarkerType.ArrowClosed },
  };
}

export function validateScenarioEdges(nodes: ActionNodeType[], edges: Edge[]) {
  const actionKeys = new Set(nodes.map((node) => node.id));
  const dangling = edges.find((edge) => !actionKeys.has(edge.source) || !actionKeys.has(edge.target));
  if (dangling) {
    throw new Error(`Связь ${dangling.source} -> ${dangling.target} ссылается на несуществующее действие.`);
  }
}

export function edgesFromScenarioJSON(parsed: unknown): Edge[] {
  const root = parsed as {
    connections?: ScenarioConnectionJSON[];
    draft_definition_json?: { edges?: Array<{ from?: string; to?: string; label?: string }> };
    draft_graph_json?: { edges?: Edge[] };
  };
  if (Array.isArray(root.connections) && root.connections.length > 0) {
    return root.connections
      .map((edge, index) => edgeFromScenarioConnection(edge, index))
      .filter((edge): edge is Edge => !!edge);
  }
  if (Array.isArray(root.draft_graph_json?.edges)) {
    return root.draft_graph_json.edges;
  }
  const defEdges = root.draft_definition_json?.edges ?? [];
  return defEdges
    .filter((edge) => edge.from && edge.to)
    .map((edge, index) => ({
      id: `imported-edge-${index}-${edge.from}-${edge.to}`,
      source: edge.from!,
      target: edge.to!,
      sourceHandle: edge.label || undefined,
      label: edge.label || undefined,
      markerEnd: { type: MarkerType.ArrowClosed },
    }));
}

export function isSameConnection(edge: Edge, from: string, to: string, branch: string | undefined) {
  if (edge.source !== from || edge.target !== to) return false;
  return branch === undefined || branchFromEdge(edge) === branch;
}
