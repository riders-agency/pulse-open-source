import type { Edge } from "@xyflow/react";
import type { ActionConfig, ActionKind, ActionNodeType } from "./types";
import type { DefinitionNodeJSON, RawGraphNodeJSON } from "./editorDraftTypes";
import { defaultConfig, normalizeEmailNode, palette } from "../registry/actions";
import { actionLabel } from "./actionPresentation";
import { defaultActionNodePosition } from "./nodeLayout";

export function defaultNodePosition(index: number) {
  return defaultActionNodePosition(index);
}

export function normalizeGraphNode(
  node: unknown,
  index: number,
  definitionNodes: DefinitionNodeJSON[] = [],
): ActionNodeType | null {
  if (!isRecord(node)) return null;
  const raw = node as RawGraphNodeJSON;
  const id = String(raw.id ?? raw.action_key ?? "").trim();
  if (!id) return null;

  const defNode = definitionNodes.find((item) => item.action_key === id);
  const actionType =
    raw.data?.action_type_key ??
    raw.integration?.action_type_key ??
    raw.action_type_key ??
    defNode?.action_type_key;
  if (!actionType || !isActionKind(actionType)) return null;

  const label =
    raw.data?.label ??
    raw.integration?.label ??
    raw.label ??
    defNode?.label ??
    actionLabel(actionType);
  const config =
    raw.data?.config ??
    raw.integration?.config ??
    raw.config ??
    defNode?.config ??
    defaultConfig(actionType);
  const position = positionFromUnknown(raw.position) ?? defaultNodePosition(index);

  return {
    id,
    type: "action",
    position,
    data: {
      label,
      action_type_key: actionType,
      config,
    },
  };
}

export function positionByActionKey(graphNodes: unknown[], actionKey: string) {
  for (const raw of graphNodes) {
    const node = raw as { id?: string; position?: { x?: number; y?: number } };
    if (node.id !== actionKey || !node.position) continue;
    return { x: Number(node.position.x ?? 80), y: Number(node.position.y ?? 160) };
  }
  return null;
}

export function filterEdgesToNodes(nodes: ActionNodeType[], edges: Edge[]) {
  const ids = new Set(nodes.map((node) => node.id));
  return edges.filter((edge) => ids.has(edge.source) && ids.has(edge.target));
}

export function cloneConfig(config: ActionConfig): ActionConfig {
  return JSON.parse(JSON.stringify(config)) as ActionConfig;
}

export function copyConfigForDuplicate(type: ActionKind, config: ActionConfig, suffix: string): ActionConfig {
  const next = cloneConfig(config);
  if (type === "webhook.trigger" && next.endpoint_key) {
    next.endpoint_key = `${next.endpoint_key}_${suffix}`;
  }
  if (type === "schedule.trigger" && next.schedule_key) {
    next.schedule_key = `${next.schedule_key}_${suffix}`;
  }
  if (type === "telegram.relay.trigger" && next.relay_key) {
    next.relay_key = `${next.relay_key}_${suffix}`;
  }
  return next;
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === "object" && !Array.isArray(value);
}

export function isActionKind(value: string): value is ActionKind {
  return palette.some((group) => group.items.some((item) => item.type === value));
}

export function positionFromUnknown(value: unknown): { x: number; y: number } | null {
  if (!isRecord(value)) return null;
  const x = Number(value.x);
  const y = Number(value.y);
  if (!Number.isFinite(x) || !Number.isFinite(y)) return null;
  return { x, y };
}

export function persistableNode(node: ActionNodeType): ActionNodeType {
  const normalized = node.data.action_type_key === "email.send" ? normalizeEmailNode(node) : node;
  const { subtitle: _subtitle, onDuplicate: _onDuplicate, ...data } = normalized.data;
  const { selected: _selected, dragging: _dragging, measured: _measured, ...rest } = normalized;
  return { ...rest, data };
}
