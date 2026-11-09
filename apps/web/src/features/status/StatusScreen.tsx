"use client";

import { Activity, Bot, Clock, Monitor, RefreshCw, Server, XCircle } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { getFrontendStatus, getRuntimeReadiness, type ReadinessPayload } from "./api/statusApi";
import { usePulseStore } from "../../store/pulse-store";

function statusLabel(value: unknown) {
  if (value === true || value === "serving" || value === "ok" || value === "ready") return "работает";
  if (value === false || value === "not_serving" || value === "error") return "ошибка";
  return typeof value === "string" && value ? value : "нет данных";
}

function StatusDot({ ok }: { ok: boolean }) {
  return <span className={`status-dot ${ok ? "is-ok" : "is-fail"}`} aria-hidden="true" />;
}

export default function StatusPage() {
  const token = usePulseStore((state) => state.token);
  const theme = usePulseStore((state) => state.theme);
  const [readiness, setReadiness] = useState<ReadinessPayload | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const frontend = useMemo(() => getFrontendStatus(theme), [theme]);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await getRuntimeReadiness(token || undefined);
      setReadiness(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось получить статус системы");
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    void load();
  }, [load]);

  const goStatus = readiness?.go_api ?? {};
  const pythonStatus = readiness?.python_grpc ?? {};
  const goOK = readiness?.ok !== false && statusLabel(goStatus.status || goStatus.ok) !== "ошибка";
  const pythonOK = readiness?.runtime !== "langgraph" || statusLabel(pythonStatus.status || pythonStatus.ok) !== "ошибка";

  return (
    <main className="status-page" data-feature="status">
      <section className="status-header">
        <div>
          <p className="eyebrow">/status</p>
          <h1>Состояние системы</h1>
          <p>Фронт, Go API и Python AI-agent без лишнего шума и без вывода секретов.</p>
        </div>
        <Button type="button" onClick={() => void load()} disabled={loading}>
          <RefreshCw size={16} className={loading ? "spin" : ""} />
          Обновить
        </Button>
      </section>

      {error ? (
        <div className="status-inline-alert">
          <XCircle size={17} />
          <span>{error}</span>
        </div>
      ) : null}

      <section className="status-grid">
        <article className="status-panel">
          <div className="status-panel-title">
            <Monitor size={18} />
            <h2>Интерфейс</h2>
            <Badge>Браузер</Badge>
          </div>
          <div className="status-row"><StatusDot ok /><span>Интерфейс</span><strong>загружен</strong></div>
          <div className="status-kv"><span>API</span><strong>{frontend.api_base}</strong></div>
          <div className="status-kv"><span>Часовой пояс</span><strong>{frontend.timezone || "unknown"}</strong></div>
          <div className="status-kv"><span>Тема</span><strong>{frontend.theme}</strong></div>
        </article>

        <article className="status-panel">
          <div className="status-panel-title">
            <Server size={18} />
            <h2>Сервер</h2>
            <Badge>{readiness?.runtime || "unknown"}</Badge>
          </div>
          <div className="status-row"><StatusDot ok={goOK} /><span>Go API</span><strong>{statusLabel(goStatus.status || goStatus.ok || readiness?.ok)}</strong></div>
          <div className="status-kv"><span>Политика готовности</span><strong>{readiness?.ready_policy || "default"}</strong></div>
          <div className="status-kv"><span>Проверено</span><strong>{readiness?.checked_at || "нет данных"}</strong></div>
        </article>

        <article className="status-panel">
          <div className="status-panel-title">
            <Bot size={18} />
            <h2>ИИ-сотрудник</h2>
            <Badge>Python gRPC</Badge>
          </div>
          <div className="status-row"><StatusDot ok={pythonOK} /><span>Исполнение LangGraph</span><strong>{statusLabel(pythonStatus.status || pythonStatus.ok)}</strong></div>
          <div className="status-kv"><span>Адрес сервиса</span><strong>{String(pythonStatus.endpoint || readiness?.endpoint_note || "internal")}</strong></div>
          <div className="status-kv"><span>Режим</span><strong>{readiness?.runtime || "langgraph"}</strong></div>
        </article>
      </section>

      <section className="status-panel status-wide-panel">
        <div className="status-panel-title">
          <Activity size={18} />
          <h2>Диагностика</h2>
          <Badge>Секреты скрыты</Badge>
        </div>
        <p className="status-muted">Технические данные скрыты по умолчанию. Здесь нет токенов, cookies, provider secrets и auth headers.</p>
        <details className="status-details">
          <summary><Clock size={16} /> Показать redacted payload</summary>
          <pre>{JSON.stringify({ readiness, frontend }, null, 2)}</pre>
        </details>
      </section>
    </main>
  );
}
