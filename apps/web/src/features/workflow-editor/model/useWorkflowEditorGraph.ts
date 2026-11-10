"use client";

import type { DragEvent } from "react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { addEdge, MarkerType, useEdgesState, useNodesState, type Connection, type Edge } from "@xyflow/react";

import {
  actionLabel,
  actionSubtitle,
  replaceSelectedActionInURL,
} from "./editorDraft";
import { normalizeActionNodeLayout } from "./nodeLayout";
import type { ActionConfig, ActionKind, ActionNodeType, RetryPolicy } from "./types";
import { useWorkflowNodeDuplicate } from "./useWorkflowNodeDuplicate";
import { defaultConfig } from "../registry/actions";

export function useWorkflowEditorGraph({
  requestedActionId,
  setError,
  setNotice,
}: {
  requestedActionId: string;
  setError: (value: string) => void;
  setNotice: (value: string) => void;
}) {
  const [nodes, setNodes, onNodesChange] = useNodesState<ActionNodeType>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [selectedNodeId, setSelectedNodeId] = useState("");
  const [graphLoaded, setGraphLoaded] = useState(false);

  const selectedNode = useMemo(() => nodes.find((node) => node.id === selectedNodeId) ?? null, [nodes, selectedNodeId]);
  const selectedSubtitle = useMemo(
    () => (selectedNode ? actionSubtitle(selectedNode) : "Выберите блок на поле сценария"),
    [selectedNode],
  );

  const selectAction = useCallback((nodeId: string) => {
    setSelectedNodeId(nodeId);
    replaceSelectedActionInURL(nodeId);
  }, []);

  const duplicateActionNode = useWorkflowNodeDuplicate({ nodes, selectAction, setError, setNodes, setNotice });

  const flowNodes = useMemo<ActionNodeType[]>(
    () =>
      nodes.map((node) => ({
        ...node,
        selected: node.id === selectedNodeId,
        data: {
          ...node.data,
          subtitle: actionSubtitle(node),
          onDuplicate: duplicateActionNode,
        },
      })),
    [duplicateActionNode, nodes, selectedNodeId],
  );

  useEffect(() => {
    if (!graphLoaded || nodes.length === 0) return;
    if (requestedActionId && nodes.some((node) => node.id === requestedActionId)) {
      if (selectedNodeId !== requestedActionId) setSelectedNodeId(requestedActionId);
      return;
    }
    if (!selectedNodeId || !nodes.some((node) => node.id === selectedNodeId)) {
      const fallback = nodes[0]?.id ?? "";
      setSelectedNodeId(fallback);
      replaceSelectedActionInURL(fallback);
    } else if (!requestedActionId) {
      replaceSelectedActionInURL(selectedNodeId);
    }
  }, [graphLoaded, nodes, requestedActionId, selectedNodeId]);

  const onConnect = useCallback(
    (connection: Connection) =>
      setEdges((current) =>
        addEdge({ ...connection, markerEnd: { type: MarkerType.ArrowClosed }, label: connection.sourceHandle || undefined }, current),
      ),
    [setEdges],
  );

  const addActionNode = useCallback(
    (type: ActionKind, x = 360, y = 180) => {
      const id = `${type.replace(/[^a-z0-9]+/g, "_")}_${Math.random().toString(36).slice(2, 7)}`;
      setNodes((current) =>
        normalizeActionNodeLayout([
          ...current,
          {
            id,
            type: "action",
            position: { x, y },
            data: { label: actionLabel(type), action_type_key: type, config: defaultConfig(type) },
          },
        ]),
      );
      selectAction(id);
    },
    [selectAction, setNodes],
  );

  const onDrop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    const type = event.dataTransfer.getData("application/rp-action") as ActionKind;
    if (type) addActionNode(type, event.nativeEvent.offsetX - 90, event.nativeEvent.offsetY - 30);
  };

  const updateSelectedConfig = (patch: Partial<ActionConfig>) => {
    if (!selectedNode) return;
    setNodes((current) =>
      current.map((node) =>
        node.id === selectedNode.id ? { ...node, data: { ...node.data, config: { ...node.data.config, ...patch } } } : node,
      ),
    );
  };

  const updateSelectedLabel = (label: string) => {
    if (!selectedNode) return;
    setNodes((current) =>
      current.map((node) => (node.id === selectedNode.id ? { ...node, data: { ...node.data, label } } : node)),
    );
  };

  const updateRetry = (patch: Partial<RetryPolicy>) => {
    const current = selectedNode?.data.config.retry_policy ?? defaultConfig("telegram.send_message").retry_policy!;
    updateSelectedConfig({ retry_policy: { ...current, ...patch, max_attempts: Math.min(10, Math.max(0, Number(patch.max_attempts ?? current.max_attempts))) } });
  };

  const deleteSelectedNode = useCallback(() => {
    if (!selectedNode) return;
    const label = selectedNode.data.label || selectedNode.id;
    const ok = window.confirm(`Удалить действие "${label}" из черновика? История выполненных запусков останется без изменений.`);
    if (!ok) return;
    const nextSelectedId = nodes.find((node) => node.id !== selectedNode.id)?.id ?? "";
    setNodes((current) => current.filter((node) => node.id !== selectedNode.id));
    setEdges((current) => current.filter((edge) => edge.source !== selectedNode.id && edge.target !== selectedNode.id));
    selectAction(nextSelectedId);
    setNotice("Действие удалено из черновика. Выполненные запуски сохранены.");
    setError("");
  }, [nodes, selectedNode, selectAction, setEdges, setError, setNodes, setNotice]);

  return {
    addActionNode,
    deleteSelectedNode,
    duplicateActionNode,
    edges,
    flowNodes,
    graphLoaded,
    nodes,
    onConnect,
    onDrop,
    onEdgesChange,
    onNodesChange,
    selectAction,
    selectedNode,
    selectedSubtitle,
    setEdges,
    setGraphLoaded,
    setNodes,
    updateRetry,
    updateSelectedConfig,
    updateSelectedLabel,
  };
}
