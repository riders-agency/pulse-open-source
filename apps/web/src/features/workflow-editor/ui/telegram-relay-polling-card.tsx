import { useEffect, useState } from "react";

import { Button } from "../../../components/ui/button";
import { Field, Input } from "../../../components/ui/input";
import { saveTelegramRelayPolling } from "../api/workflowEditorApi";
import {
  defaultTelegramRelayPollingSnippetSettings,
  telegramRelayDocURL,
  type TelegramRelayPollingSnippetSettings,
} from "./telegram-relay-snippets";

export function TelegramRelayPollingCard({
  token,
  projectId,
  actionDependencyId,
  metadata,
  onSnippetSettingsChange,
}: {
  token: string;
  projectId?: string;
  actionDependencyId?: string;
  metadata: Record<string, unknown>;
  onSnippetSettingsChange: (settings: TelegramRelayPollingSnippetSettings) => void;
}) {
  const [enabled, setEnabled] = useState(false);
  const [timeout, setTimeoutValue] = useState(defaultTelegramRelayPollingSnippetSettings.timeout);
  const [limit, setLimit] = useState(defaultTelegramRelayPollingSnippetSettings.limit);
  const [deleteWebhook, setDeleteWebhook] = useState(true);
  const [dropPending, setDropPending] = useState(defaultTelegramRelayPollingSnippetSettings.dropPending);
  const [busy, setBusy] = useState(false);
  const [status, setStatus] = useState("");
  const [error, setError] = useState("");

  const metaEnabled = metadata.polling_enabled === true;
  const metaTimeout = Number(metadata.polling_timeout_seconds);
  const metaLimit = Number(metadata.polling_limit);
  const metaDeleteWebhook = metadata.polling_delete_webhook !== false;
  const metaDropPending = metadata.polling_drop_pending_updates === true;
  const lastOK = typeof metadata.polling_last_ok_at === "string" ? metadata.polling_last_ok_at : "";
  const lastError = typeof metadata.polling_last_error === "string" ? metadata.polling_last_error : "";
  const lastUpdateID = metadata.polling_last_update_id != null ? String(metadata.polling_last_update_id) : "";

  useEffect(() => {
    setEnabled(metaEnabled);
    setTimeoutValue(Number.isFinite(metaTimeout) ? metaTimeout : defaultTelegramRelayPollingSnippetSettings.timeout);
    setLimit(Number.isFinite(metaLimit) ? metaLimit : defaultTelegramRelayPollingSnippetSettings.limit);
    setDeleteWebhook(metaDeleteWebhook);
    setDropPending(metaDropPending);
    setStatus("");
    setError("");
  }, [actionDependencyId, metaDeleteWebhook, metaDropPending, metaEnabled, metaLimit, metaTimeout]);

  useEffect(() => {
    onSnippetSettingsChange({ timeout, limit, dropPending });
  }, [dropPending, limit, onSnippetSettingsChange, timeout]);

  const save = async (nextEnabled = enabled) => {
    setError("");
    setStatus("");
    if (!projectId || !actionDependencyId) {
      setError("Выберите telegram.relay интеграцию.");
      return;
    }
    setBusy(true);
    try {
      const out = await saveTelegramRelayPolling(token, projectId, actionDependencyId, {
        enabled: nextEnabled,
        timeout_seconds: Math.max(1, Math.min(60, Number(timeout) || defaultTelegramRelayPollingSnippetSettings.timeout)),
        limit: Math.max(1, Math.min(100, Number(limit) || defaultTelegramRelayPollingSnippetSettings.limit)),
        delete_webhook: deleteWebhook,
        drop_pending_updates: dropPending,
      });
      setEnabled(out.polling_enabled === true);
      setTimeoutValue(Number(out.polling_timeout_seconds) || defaultTelegramRelayPollingSnippetSettings.timeout);
      setLimit(Number(out.polling_limit) || defaultTelegramRelayPollingSnippetSettings.limit);
      setDeleteWebhook(out.polling_delete_webhook !== false);
      setDropPending(out.polling_drop_pending_updates === true);
      setStatus(nextEnabled ? "Polling включен. Go API будет забирать updates через getUpdates и сохранять их в relay." : "Polling выключен.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить получение сообщений");
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="rounded-md border border-border p-3">
      <div className="between gap-2">
        <div>
          <div className="text-sm font-medium">Polling через getUpdates</div>
          <div className="muted mt-1 text-xs">
            Go API будет сам читать Telegram updates через HTTP long polling. WebSocket не нужен; при включении polling webhook должен быть удален.
          </div>
        </div>
        <Button
          variant={enabled ? "outline" : "default"}
          size="sm"
          disabled={busy || !actionDependencyId}
          onClick={() => void save(!enabled)}
        >
          {busy ? "Сохраняем…" : enabled ? "Выключить" : "Включить"}
        </Button>
      </div>
      <div className="mt-2 grid gap-2 sm:grid-cols-2">
        <Field label="Timeout, sec">
          <Input
            min={1}
            max={60}
            type="number"
            value={timeout}
            onChange={(event) => setTimeoutValue(Number(event.target.value))}
          />
        </Field>
        <Field label="Лимит">
          <Input
            min={1}
            max={100}
            type="number"
            value={limit}
            onChange={(event) => setLimit(Number(event.target.value))}
          />
        </Field>
      </div>
      <label className="row mt-2 text-sm">
        <input
          type="checkbox"
          checked={deleteWebhook}
          onChange={(event) => setDeleteWebhook(event.target.checked)}
        />
        удалять webhook при polling
      </label>
      <label className="row mt-2 text-sm">
        <input
          type="checkbox"
          checked={dropPending}
          onChange={(event) => setDropPending(event.target.checked)}
        />
        удалить pending updates при первом deleteWebhook
      </label>
      <div className="row mt-2">
        <Button variant="outline" size="sm" disabled={busy || !actionDependencyId} onClick={() => void save(enabled)}>
          Сохранить polling
        </Button>
        <a className="muted text-xs" href={telegramRelayDocURL} target="_blank" rel="noreferrer">документация relay</a>
      </div>
      {lastOK || lastUpdateID ? (
        <div className="muted mt-2 text-xs">
          Последний успешный polling: {lastOK || "ещё нет"}{lastUpdateID ? `, update_id ${lastUpdateID}` : ""}
        </div>
      ) : null}
      {lastError ? <div className="status-failed mt-2 text-xs">Последняя ошибка: {lastError}</div> : null}
      {status ? <div className="status-success mt-2 text-xs">{status}</div> : null}
      {error ? <div className="status-failed mt-2 text-xs">{error}</div> : null}
    </div>
  );
}
