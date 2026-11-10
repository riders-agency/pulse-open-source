import type { Node } from "@xyflow/react";

import type { ActionConfig } from "./action-config";

export type { ActionConfig, RetryPolicy } from "./action-config";

export type ActionKind =
  | "mailbox.incoming.trigger"
  | "office.classify"
  | "office.route"
  | "office.website.prepare"
  | "office.invoice.prepare"
  | "office.approve"
  | "office.apply"
  | "webhook.trigger"
  | "telegram.relay.trigger"
  | "schedule.trigger"
  | "manual.trigger"
  | "clients.manual_payment.trigger"
  | "clients.payment_due.trigger"
  | "clients.payment_cycle.collect"
  | "condition.branch"
  | "stop"
  | "delay.wait"
  | "bot.set_state"
  | "http.request"
  | "ai.apinet.normalize_text"
  | "max.send_message"
  | "telegram.send_message"
  | "telegram.edit_message_text"
  | "telegram.answer_callback_query"
  | "telegram.send_document"
  | "person.fio.normalize"
  | "person.phone.normalize"
  | "uon.create_lead"
  | "yclients.certificate.issue"
  | "email.send";

export type JsonField = {
  path: string;
  type: "string" | "number" | "boolean" | "object" | "array";
  required: boolean;
  description: string;
  example: string;
};

export type ActionData = {
  label: string;
  subtitle?: string;
  action_type_key: ActionKind;
  config: ActionConfig;
  onDuplicate?: (nodeId: string) => void;
};

export type ActionNodeType = Node<ActionData, "action">;

export type ScenarioActionJSON = {
  action_key?: string;
  label?: string;
  action_type_key?: string;
  category?: string;
  type_label?: string;
  config?: ActionConfig;
  ui?: { position?: { x: number; y: number } };
};

export type ScenarioConnectionJSON = {
  from?: string;
  to?: string;
  branch?: string;
  label?: string;
  description?: string;
};

export type WorkflowTestMode = "full" | "through_action" | "selected_only";
