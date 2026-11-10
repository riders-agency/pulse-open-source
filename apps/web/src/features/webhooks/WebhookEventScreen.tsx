"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { jsonText } from "../../lib/api";
import { getWebhookEvent, replayWebhookEvent } from "./api/webhooksApi";
import { usePulseStore } from "../../store/pulse-store";

export default function WebhookEventPage() {
  const { webhookEventId } = useParams<{ webhookEventId: string }>();
  const token = usePulseStore((s) => s.token);
  const [event, setEvent] = useState<any>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!token) return;
    getWebhookEvent(webhookEventId, token)
      .then(setEvent)
      .catch((err) => setError(err.message));
  }, [token, webhookEventId]);

  const replay = async () => {
    try {
      const out = await replayWebhookEvent(webhookEventId, token);
      window.location.href = `/runs/${out.workflow_run_id}`;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось повторить событие");
    }
  };

  return (
    <main className="page">
      <div className="page-title">
        <div>
          <h1>Входящий вебхук</h1>
          <p className="muted">Детали payload и связанного запуска.</p>
        </div>
        <Button onClick={replay}>Повторить событие</Button>
      </div>
      {error ? <Badge tone="danger">{error}</Badge> : null}
      <section className="grid-2">
        <Card>
          <CardHeader>
            <CardTitle>Событие</CardTitle>
            <CardDescription>{event?.method} {event?.path}</CardDescription>
          </CardHeader>
          <CardContent className="stack">
            <div className="between"><span>Статус</span><Badge>{event?.status ?? "—"}</Badge></div>
            {event?.workflow_run_id ? (
              <Link href={`/runs/${event.workflow_run_id}`}><Button variant="secondary">Открыть запуск</Button></Link>
            ) : null}
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle>Данные события</CardTitle></CardHeader>
          <CardContent>
            <pre className="code-block">{jsonText(event?.payload)}</pre>
          </CardContent>
        </Card>
      </section>
    </main>
  );
}
