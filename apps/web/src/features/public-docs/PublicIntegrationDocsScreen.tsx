"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import {
  integrationDocMarkdownURL,
  searchIntegrationDocs,
  type IntegrationDocSearchResult,
  workflowPatchContractMarkdownURL,
} from "./api/publicDocsApi";

export default function IntegrationDocsPage() {
  const [query, setQuery] = useState("");
  const [items, setItems] = useState<IntegrationDocSearchResult[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    let alive = true;
    searchIntegrationDocs(query)
      .then((rows) => {
        if (!alive) return;
        setItems(rows);
        setError("");
      })
      .catch((err) => {
        if (!alive) return;
        setError(err instanceof Error ? err.message : "Не удалось выполнить поиск");
      });
    return () => {
      alive = false;
    };
  }, [query]);

  return (
    <main className="docs-page">
      <div className="docs-topline">
        <Link href="/docs">Документация</Link>
        <a href={workflowPatchContractMarkdownURL()}>Контракт изменений сценария через ИИ</a>
        <Link href="/docs/integrations/ai-project-employee">ИИ-сотрудник проекта</Link>
        <Link href="/docs/integrations/task-process-patch">Изменение процесса задач</Link>
      </div>
      <h1>Документация интеграций</h1>
      <p className="muted">Публичные sanitized документы для поиска и для LLM. Секреты, токены и приватные значения здесь не хранятся.</p>
      <input
        className="docs-search-input"
        value={query}
        onChange={(event) => setQuery(event.target.value)}
        placeholder="Поиск: телега, письмо, раз в час, запросить endpoint"
      />
      {error ? <p className="form-error">{error}</p> : null}
      <div className="docs-result-list">
        {items.map((item) => (
          <article className="docs-result" key={item.doc_key}>
            <div>
              <h2>{item.title}</h2>
              <p>{item.summary}</p>
              <small>
                {item.provider_key} · {item.action_type_keys.join(", ")} · {item.reason}
              </small>
            </div>
            <div className="docs-result-actions">
              <Link href={`/docs/integrations/${item.public_slug}`}>Открыть</Link>
              <a href={integrationDocMarkdownURL(item.public_slug)}>Markdown</a>
            </div>
          </article>
        ))}
      </div>
    </main>
  );
}
