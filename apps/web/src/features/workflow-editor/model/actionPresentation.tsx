import { Braces, Briefcase, GitBranch, Mail, MessageCircle, MousePointer2, Send, Sparkles, Square, Timer, Webhook } from "lucide-react";

import { palette } from "../registry/actions";
import type { ActionKind, ActionNodeType } from "./types";

export function iconFor(type: ActionKind) {
  const size = 16;
  if (type === "webhook.trigger") return <Webhook size={size} />;
  if (type === "telegram.relay.trigger") return <Send size={size} />;
  if (type === "schedule.trigger") return <Timer size={size} />;
  if (type === "manual.trigger") return <MousePointer2 size={size} />;
  if (type === "clients.manual_payment.trigger") return <MousePointer2 size={size} />;
  if (type === "clients.payment_due.trigger") return <Timer size={size} />;
  if (type === "clients.payment_cycle.collect") return <Braces size={size} />;
  if (type === "condition.branch") return <GitBranch size={size} />;
  if (type === "delay.wait") return <Timer size={size} />;
  if (type === "bot.set_state") return <MessageCircle size={size} />;
  if (type === "http.request") return <Braces size={size} />;
  if (type === "ai.apinet.normalize_text") return <Sparkles size={size} />;
  if (type === "max.send_message") return <MessageCircle size={size} />;
  if (type === "telegram.send_message") return <Send size={size} />;
  if (type === "telegram.edit_message_text") return <Send size={size} />;
  if (type === "telegram.answer_callback_query") return <Send size={size} />;
  if (type === "telegram.send_document") return <Send size={size} />;
  if (type === "person.fio.normalize") return <GitBranch size={size} />;
  if (type === "person.phone.normalize") return <GitBranch size={size} />;
  if (type === "uon.create_lead") return <Briefcase size={size} />;
  if (type === "email.send") return <Mail size={size} />;
  return <Square size={size} />;
}

export function actionLabel(type: ActionKind) {
  return palette.flatMap((group) => group.items).find((item) => item.type === type)?.label ?? type;
}

export function actionTypeMeta(type: ActionKind) {
  for (const group of palette) {
    const item = group.items.find((candidate) => candidate.type === type);
    if (item) return { group: group.group, label: item.label };
  }
  return { group: "Действие", label: type };
}

export function actionSubtitle(node: ActionNodeType) {
  const type = node.data?.action_type_key;
  if (!type || !isPresentationActionKind(type)) return "Действие";
  const meta = actionTypeMeta(type);
  return `${meta.group} · ${meta.label}`;
}

function isPresentationActionKind(value: string): value is ActionKind {
  return palette.some((group) => group.items.some((item) => item.type === value));
}
