import { useEffect, useState } from "react";

import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import { rotateTelegramRelayWebhookSecret } from "../api/workflowEditorApi";

export function TelegramRelaySecretCard({
  token,
  projectId,
  actionDependencyId,
  relayKey,
  onSecretChange,
}: {
  token: string;
  projectId?: string;
  actionDependencyId?: string;
  relayKey: string;
  onSecretChange: (secret: string) => void;
}) {
  const [webhookSecret, setWebhookSecret] = useState("");
  const [busy, setBusy] = useState(false);
  const [status, setStatus] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    setWebhookSecret("");
    setStatus("");
    setError("");
    onSecretChange("");
  }, [actionDependencyId, relayKey, onSecretChange]);

  const rotate = async () => {
    setError("");
    setStatus("");
    if (!projectId || !actionDependencyId) {
      setError("Выберите telegram.relay интеграцию.");
      return;
    }
    setBusy(true);
    try {
      const out = await rotateTelegramRelayWebhookSecret(token, projectId, actionDependencyId);
      const nextSecret = out.webhook_secret ?? "";
      setWebhookSecret(nextSecret);
      onSecretChange(nextSecret);
      setStatus("Новый webhook secret сохранен и подставлен в snippets. Скопируйте сейчас: он показывается один раз.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось создать секрет входящего события");
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="rounded-md border border-border p-3">
      <div className="between gap-2">
        <div>
          <div className="text-sm font-medium">Секрет вебхука</div>
          <div className="muted mt-1 text-xs">
            Route key в URL не считается паролем. Secret уходит в Telegram как header и показывается только после генерации.
          </div>
        </div>
        <Button variant="outline" size="sm" disabled={busy || !actionDependencyId} onClick={rotate}>
          {busy ? "Генерируем…" : "Сгенерировать"}
        </Button>
      </div>
      <Input
        readOnly
        value={webhookSecret || "Нажмите «Сгенерировать», чтобы подставить secret в snippets"}
        className="mt-2"
      />
      {status ? <div className="status-success mt-2 text-xs">{status}</div> : null}
      {error ? <div className="status-failed mt-2 text-xs">{error}</div> : null}
    </div>
  );
}
