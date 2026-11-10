import { useCallback } from "react";

import {
  actionLabel,
  copyConfigForDuplicate,
} from "./editorDraft";
import { normalizeActionNodeLayout } from "./nodeLayout";
import type { ActionNodeType } from "./types";

export function useWorkflowNodeDuplicate({
  nodes,
  selectAction,
  setError,
  setNodes,
  setNotice,
}: {
  nodes: ActionNodeType[];
  selectAction: (nodeId: string) => void;
  setError: (value: string) => void;
  setNodes: (update: (current: ActionNodeType[]) => ActionNodeType[]) => void;
  setNotice: (value: string) => void;
}) {
  return useCallback(
    (nodeId: string) => {
      const source = nodes.find((node) => node.id === nodeId);
      if (!source?.data?.action_type_key) return;
      const suffix = Math.random().toString(36).slice(2, 7);
      const id = `${source.data.action_type_key.replace(/[^a-z0-9]+/g, "_")}_${suffix}`;
      const label = `${source.data.label || actionLabel(source.data.action_type_key)} копия`;
      setNodes((current) =>
        normalizeActionNodeLayout([
          ...current,
          {
            ...source,
            id,
            selected: false,
            dragging: false,
            position: { x: source.position.x + 48, y: source.position.y + 48 },
            data: {
              label,
              action_type_key: source.data.action_type_key,
              config: copyConfigForDuplicate(source.data.action_type_key, source.data.config, suffix),
            },
          },
        ]),
      );
      selectAction(id);
      setNotice("Действие скопировано в черновик.");
      setError("");
    },
    [nodes, selectAction, setError, setNodes, setNotice],
  );
}
