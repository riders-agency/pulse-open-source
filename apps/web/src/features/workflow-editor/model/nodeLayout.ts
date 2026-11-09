import type { Edge } from "@xyflow/react";
import type { ActionNodeType } from "./types";

export const ACTION_NODE_LAYOUT = {
  originX: 80,
  originY: 160,
  columnWidth: 300,
  rowHeight: 210,
} as const;

type GridSlot = { column: number; row: number };
type NodeOrder = { node: ActionNodeType; index: number; slot: GridSlot };

export function defaultActionNodePosition(index: number) {
  return slotPosition({ column: index % 3, row: Math.floor(index / 3) });
}

export function normalizeActionNodeLayout(nodes: ActionNodeType[], edges: Edge[] = []): ActionNodeType[] {
  if (nodes.length === 0) return nodes;

  const nodeIDs = new Set(nodes.map((node) => node.id));
  const validEdges = edges.filter((edge) => nodeIDs.has(edge.source) && nodeIDs.has(edge.target));
  if (validEdges.length === 0) return normalizeBySourceSlots(nodes);

  const orders = nodes.map((node, index) => ({ node, index, slot: positionToSlot(node.position, index) }));
  const depths = topologicalDepths(orders, validEdges);
  const rows = allocateRows(orders, validEdges, depths);

  return orders.map(({ node }) => ({
    ...node,
    position: slotPosition({ column: depths.get(node.id) ?? 0, row: rows.get(node.id) ?? 0 }),
  }));
}

function normalizeBySourceSlots(nodes: ActionNodeType[]): ActionNodeType[] {
  const occupied = new Set<string>();
  return nodes.map((node, index) => {
    const desired = positionToSlot(node.position, index);
    const slot = firstFreeSlot(desired, occupied, nodes.length);
    occupied.add(slotKey(slot));
    return { ...node, position: slotPosition(slot) };
  });
}

function topologicalDepths(orders: NodeOrder[], edges: Edge[]) {
  const incoming = new Map<string, number>();
  const outgoing = new Map<string, string[]>();
  const orderByID = new Map(orders.map((item) => [item.node.id, item]));
  for (const { node } of orders) {
    incoming.set(node.id, 0);
    outgoing.set(node.id, []);
  }
  for (const edge of edges) {
    outgoing.get(edge.source)?.push(edge.target);
    incoming.set(edge.target, (incoming.get(edge.target) ?? 0) + 1);
  }

  const depths = new Map<string, number>();
  const queue = orders
    .filter(({ node }) => (incoming.get(node.id) ?? 0) === 0)
    .sort(compareNodeOrder)
    .map(({ node }) => node.id);

  while (queue.length > 0) {
    const id = queue.shift()!;
    const baseDepth = depths.get(id) ?? 0;
    const targets = [...(outgoing.get(id) ?? [])].sort((a, b) => compareNodeOrder(orderByID.get(a), orderByID.get(b)));
    for (const target of targets) {
      depths.set(target, Math.max(depths.get(target) ?? 0, baseDepth + 1));
      incoming.set(target, (incoming.get(target) ?? 0) - 1);
      if ((incoming.get(target) ?? 0) === 0) {
        queue.push(target);
        queue.sort((a, b) => compareNodeOrder(orderByID.get(a), orderByID.get(b)));
      }
    }
  }

  for (const item of orders) {
    if (!depths.has(item.node.id)) {
      depths.set(item.node.id, item.slot.column);
    }
  }

  return depths;
}

function allocateRows(orders: NodeOrder[], edges: Edge[], depths: Map<string, number>) {
  const rows = new Map<string, number>();
  const incomingByTarget = new Map<string, string[]>();
  const byDepth = new Map<number, NodeOrder[]>();
  for (const edge of edges) {
    incomingByTarget.set(edge.target, [...(incomingByTarget.get(edge.target) ?? []), edge.source]);
  }
  for (const item of orders) {
    const depth = depths.get(item.node.id) ?? 0;
    byDepth.set(depth, [...(byDepth.get(depth) ?? []), item]);
  }

  const sortedDepths = [...byDepth.keys()].sort((a, b) => a - b);
  for (const depth of sortedDepths) {
    const occupied = new Set<number>();
    const items = [...(byDepth.get(depth) ?? [])].sort((a, b) => desiredRow(a, incomingByTarget, rows) - desiredRow(b, incomingByTarget, rows) || compareNodeOrder(a, b));
    for (const item of items) {
      const row = firstFreeRow(Math.round(desiredRow(item, incomingByTarget, rows)), occupied);
      occupied.add(row);
      rows.set(item.node.id, row);
    }
  }

  return rows;
}

function desiredRow(item: NodeOrder, incomingByTarget: Map<string, string[]>, rows: Map<string, number>) {
  const parentRows = (incomingByTarget.get(item.node.id) ?? [])
    .map((id) => rows.get(id))
    .filter((value): value is number => typeof value === "number");
  if (parentRows.length === 0) return item.slot.row;
  return parentRows.reduce((sum, row) => sum + row, 0) / parentRows.length;
}

function firstFreeRow(desired: number, occupied: Set<number>) {
  const start = Math.max(0, desired);
  if (!occupied.has(start)) return start;
  for (let offset = 1; offset <= occupied.size + 1; offset += 1) {
    const down = start + offset;
    if (!occupied.has(down)) return down;
    const up = start - offset;
    if (up >= 0 && !occupied.has(up)) return up;
  }
  return occupied.size;
}

function positionToSlot(position: ActionNodeType["position"] | undefined, fallbackIndex: number): GridSlot {
  if (!position || !Number.isFinite(position.x) || !Number.isFinite(position.y)) {
    return { column: fallbackIndex % 3, row: Math.floor(fallbackIndex / 3) };
  }
  return {
    column: Math.max(0, Math.round((position.x - ACTION_NODE_LAYOUT.originX) / ACTION_NODE_LAYOUT.columnWidth)),
    row: Math.max(0, Math.round((position.y - ACTION_NODE_LAYOUT.originY) / ACTION_NODE_LAYOUT.rowHeight)),
  };
}

function firstFreeSlot(desired: GridSlot, occupied: Set<string>, nodeCount: number): GridSlot {
  let column = desired.column;
  let row = desired.row;
  let rowProbeCount = 0;
  const maxRowsBeforeNextColumn = Math.max(8, nodeCount + 2);

  while (occupied.has(slotKey({ column, row }))) {
    row += 1;
    rowProbeCount += 1;
    if (rowProbeCount > maxRowsBeforeNextColumn) {
      column += 1;
      row = 0;
      rowProbeCount = 0;
    }
  }

  return { column, row };
}

function slotPosition(slot: GridSlot) {
  return {
    x: ACTION_NODE_LAYOUT.originX + slot.column * ACTION_NODE_LAYOUT.columnWidth,
    y: ACTION_NODE_LAYOUT.originY + slot.row * ACTION_NODE_LAYOUT.rowHeight,
  };
}

function slotKey(slot: GridSlot) {
  return `${slot.column}:${slot.row}`;
}

function compareNodeOrder(a: NodeOrder | undefined, b: NodeOrder | undefined) {
  if (!a && !b) return 0;
  if (!a) return 1;
  if (!b) return -1;
  return a.slot.row - b.slot.row || a.slot.column - b.slot.column || a.index - b.index || a.node.id.localeCompare(b.node.id);
}
