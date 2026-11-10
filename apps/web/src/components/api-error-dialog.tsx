"use client";

import { AlertTriangle } from "lucide-react";
import { useEffect, useState } from "react";

import { API_ERROR_EVENT, ApiErrorPayload } from "../lib/api";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "./ui/dialog";

export function ApiErrorDialog() {
  const [error, setError] = useState<ApiErrorPayload | null>(null);

  useEffect(() => {
    const onError = (event: Event) => {
      const detail = (event as CustomEvent<ApiErrorPayload>).detail;
      if (detail?.status) setError(detail);
    };
    window.addEventListener(API_ERROR_EVENT, onError);
    return () => window.removeEventListener(API_ERROR_EVENT, onError);
  }, []);

  return (
    <Dialog open={Boolean(error)} onOpenChange={(open) => !open && setError(null)}>
      <DialogContent className="api-error-dialog">
        <DialogHeader>
          <div className="error-dialog-kicker">
            <AlertTriangle size={18} />
            <Badge tone={badgeTone(error?.status)}>{error ? `HTTP ${error.status}` : "Ошибка"}</Badge>
          </div>
          <DialogTitle>{error?.title ?? "Не удалось выполнить действие"}</DialogTitle>
          {error?.detail ? <DialogDescription>{error.detail}</DialogDescription> : null}
        </DialogHeader>

        {error?.action ? (
          <section className="error-dialog-section">
            <h3>Что сделать</h3>
            <p>{error.action}</p>
          </section>
        ) : null}

        <details className="error-dialog-section">
          <summary>Технические сведения</summary>
          <dl className="error-dialog-details">
            {error?.method || error?.path ? (
              <>
                <dt>Запрос</dt>
                <dd>{[error.method, error.path].filter(Boolean).join(" ")}</dd>
              </>
            ) : null}
            {error?.request_id ? (
              <>
                <dt>Код обращения</dt>
                <dd className="mono">{error.request_id}</dd>
              </>
            ) : null}
            {error?.validate && Object.keys(error.validate).length > 0 ? (
              <>
                <dt>Проверка</dt>
                <dd className="mono">{Object.keys(error.validate).join(", ")}</dd>
              </>
            ) : null}
          </dl>
        </details>

        <div className="form-actions">
          <Button onClick={() => setError(null)}>Закрыть</Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function badgeTone(status?: number) {
  if (!status) return "warning" as const;
  if (status >= 500) return "danger" as const;
  if (status === 409 || status === 400) return "warning" as const;
  return "danger" as const;
}
