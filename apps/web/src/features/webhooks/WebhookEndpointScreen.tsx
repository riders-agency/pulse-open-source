"use client";

import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Badge } from "../../components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Table } from "../../components/ui/table";
import { getWebhookEndpointDetail } from "./api/webhooksApi";
import { usePulseStore } from "../../store/pulse-store";

export default function WebhookEndpointPage() {
  const { webhookEndpointId } = useParams<{ webhookEndpointId: string }>();
  const token = usePulseStore((s) => s.token);
  const [detail, setDetail] = useState<any>(null);
  const [snippets, setSnippets] = useState<Record<string, string>>({});
  const [events, setEvents] = useState<any[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!token) return;
    getWebhookEndpointDetail(webhookEndpointId, token)
      .then(({ detail: endpointDetail, snippets: endpointSnippets, events: endpointEvents }) => {
        setDetail(endpointDetail);
        setSnippets(endpointSnippets);
        setEvents(endpointEvents);
      })
      .catch((err) => setError(err.message));
  }, [token, webhookEndpointId]);

  return (
    <main className="page">
      <div className="page-title">
        <div>
          <h1>Вебхук</h1>
          <p className="muted">Настройки endpoint, snippets и события входящих запросов.</p>
        </div>
        <Badge tone={detail?.status === "active" ? "success" : "warning"}>{detail?.status ?? "загрузка"}</Badge>
      </div>
      {error ? <Badge tone="danger">{error}</Badge> : null}
      <section className="layout-sidebar">
        <Card>
          <CardHeader>
            <CardTitle>Примеры вызова</CardTitle>
            <CardDescription>Секреты в браузерный JavaScript не выносить.</CardDescription>
          </CardHeader>
          <CardContent className="stack">
            {Object.entries(snippets).map(([key, value]) => (
              <div key={key}>
                <Badge>{key}</Badge>
                <pre className="code-block">{value}</pre>
              </div>
            ))}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>События</CardTitle>
            <CardDescription>Последние входящие webhook events.</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <thead><tr><th>Статус</th><th>Путь</th></tr></thead>
              <tbody>
                {events.map((event) => (
                  <tr key={event.webhook_event_id}>
                    <td>{event.status}</td>
                    <td className="mono">{event.path}</td>
                  </tr>
                ))}
              </tbody>
            </Table>
          </CardContent>
        </Card>
      </section>
    </main>
  );
}
