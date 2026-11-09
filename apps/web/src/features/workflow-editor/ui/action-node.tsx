import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Copy } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { startsRunWithoutIncomingEdge } from "../model/actionRules";
import { actionLabel, iconFor, isActionKind } from "../model/editorDraft";
import type { ActionNodeType } from "../model/types";

export function ActionNode({ id, data, selected }: NodeProps<ActionNodeType>) {
  const actionType = data?.action_type_key;
  if (!actionType || !isActionKind(actionType)) {
    return (
      <div className="flow-node">
        <div className="flow-node-title">
          <span>{data?.label ?? id}</span>
        </div>
        <small>Неизвестное действие</small>
      </div>
    );
  }
  const isCondition = actionType === "condition.branch";
  return (
    <div className={`flow-node ${data.config?.disabled ? "flow-node-disabled" : ""} ${selected ? "flow-node-selected" : ""}`}>
      {!startsRunWithoutIncomingEdge(actionType) ? <Handle type="target" position={Position.Left} /> : null}
      <div className="flow-node-head">
        <div className="flow-node-title">
          {iconFor(actionType)}
          <span>{data.label}</span>
        </div>
        <button
          className="flow-node-copy nodrag"
          title="Дублировать действие"
          type="button"
          onClick={(event) => {
            event.stopPropagation();
            data.onDuplicate?.(id);
          }}
        >
          <Copy size={14} />
        </button>
      </div>
      <small>{data.subtitle ?? actionLabel(actionType)}</small>
      <div className="flow-node-badges row">
        {data.config?.disabled ? <Badge tone="warning">Отключено</Badge> : null}
        <Badge tone={data.config?.retry_policy ? "info" : "muted"}>{data.config?.retry_policy?.max_attempts ?? 0}/10 попыток</Badge>
        {actionType.includes("send") ? <Badge>проверки</Badge> : null}
      </div>
      {isCondition ? (
        <>
          <Handle className="flow-handle-true" id="true" type="source" position={Position.Right} />
          <Handle className="flow-handle-false" id="false" type="source" position={Position.Right} />
        </>
      ) : (
        <Handle type="source" position={Position.Right} />
      )}
    </div>
  );
}
