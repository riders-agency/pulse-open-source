"use client";

import { Send } from "lucide-react";

import { CollapsibleRawJson } from "../CollapsibleRawJson";
import type { IntegrationPanelCtx } from "./GenericIntegrationPanel";

function at(root: unknown, path: string) {
  let cur = root;
  for (const part of path.split(".")) {
    if (!cur || typeof cur !== "object" || Array.isArray(cur)) return "";
    cur = (cur as Record<string, unknown>)[part];
  }
  if (cur === undefined || cur === null) return "";
  return typeof cur === "string" ? cur : JSON.stringify(cur);
}

export function TelegramRelayTriggerPanel({ action }: IntegrationPanelCtx) {
  const input = action.input_json as Record<string, unknown> | undefined;
  const command = at(input, "telegram_relay.command");
  const text = at(input, "telegram_relay.message.text") || at(input, "telegram_relay.text");
  const chatID = at(input, "telegram_relay.chat.id") || at(input, "telegram_relay.chat_id");
  const state = at(input, "bot_session.state_key");
  return (
    <div className="stack gap-4">
      <div className="flex items-center gap-2">
        <Send className="text-sky-600" size={20} />
        <div className="text-sm font-medium">Триггер: Telegram relay</div>
      </div>
      <div className="grid gap-2 rounded-md border border-border p-3 text-sm">
        <div className="row"><span className="muted">Команда</span><strong>{command || "—"}</strong></div>
        <div className="row"><span className="muted">Диалог</span><strong>{chatID || "—"}</strong></div>
        <div className="row"><span className="muted">Состояние</span><strong>{state || "—"}</strong></div>
        <div className="muted">{text || "Нет текста сообщения"}</div>
      </div>
      <CollapsibleRawJson title="telegram update / input запуска" data={action.input_json} defaultOpen />
      <CollapsibleRawJson title="output триггера" data={action.output_json} />
    </div>
  );
}
