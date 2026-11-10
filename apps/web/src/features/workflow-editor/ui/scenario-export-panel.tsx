import { Copy } from "lucide-react";

import { Button } from "../../../components/ui/button";
import { Field, Textarea } from "../../../components/ui/input";
import type { IntegrationDocSearchResult } from "../../../lib/api";
import { aiWorkflowDocsURL } from "../registry/actions";

export function ScenarioExportPanel({
  jsonText,
  promptText,
  aiTaskText,
  setAiTaskText,
  integrationDocs,
  selectedDocKeys,
  setSelectedDocKeys,
  docsBusy,
  importText,
  setImportText,
  applyImport,
  copyText,
}: {
  jsonText: string;
  promptText: string;
  aiTaskText: string;
  setAiTaskText: (value: string) => void;
  integrationDocs: IntegrationDocSearchResult[];
  selectedDocKeys: string[];
  setSelectedDocKeys: (value: string[] | ((current: string[]) => string[])) => void;
  docsBusy: boolean;
  importText: string;
  setImportText: (value: string) => void;
  applyImport: () => void;
  copyText: (value: string, label: string) => void;
}) {
  return (
    <section className="palette-section scenario-export">
      <h3>Действия patch</h3>
      <div className="scenario-help">
        <strong>Как отдавать ИИ</strong>
        <p>Опишите задачу. Поиск подберет релевантные интеграции, а prompt уже вставит ссылки на документацию и текущий context JSON.</p>
        <a href={aiWorkflowDocsURL} target="_blank" rel="noreferrer">Документация для LLM</a>
      </div>
      <Field label="Задача для ИИ">
        <Textarea value={aiTaskText} onChange={(event) => setAiTaskText(event.target.value)} placeholder="Например: раз в час запросить endpoint и отправить результат в Telegram" />
      </Field>
      <div className="integration-doc-results">
        <div className="between">
          <strong>Документы по намерению</strong>
          {docsBusy ? <small className="muted">поиск...</small> : null}
        </div>
        {integrationDocs.length ? (
          integrationDocs.map((doc) => <IntegrationDocChoice doc={doc} key={doc.doc_key} selectedDocKeys={selectedDocKeys} setSelectedDocKeys={setSelectedDocKeys} />)
        ) : (
          <p className="muted text-xs">Начните писать задачу, чтобы подобрать документацию.</p>
        )}
      </div>
      <div className="scenario-export-actions">
        <Button variant="outline" size="sm" onClick={() => copyText(promptText, "Prompt для ИИ")}><Copy size={14} />Скопировать prompt</Button>
        <Button variant="ghost" size="sm" onClick={() => copyText(jsonText, "Контекст действий")}><Copy size={14} />Только context</Button>
      </div>
      <details className="scenario-export-details"><summary>Контекст для ИИ</summary><pre className="code-block">{jsonText}</pre></details>
      <details className="scenario-export-details">
        <summary>Импорт patch от ИИ</summary>
        <div className="scenario-import">
          <Textarea value={importText} onChange={(event) => setImportText(event.target.value)} placeholder="Вставьте JSON riders-pulse-workflow-patch-v1" />
          <Button variant="secondary" size="sm" disabled={!importText.trim()} onClick={applyImport}>Применить в черновик</Button>
        </div>
      </details>
      <p className="muted text-xs">В запрос всегда добавляется ссылка {aiWorkflowDocsURL}. В ответе нужен только патч для текущего черновика.</p>
    </section>
  );
}

function IntegrationDocChoice({
  doc,
  selectedDocKeys,
  setSelectedDocKeys,
}: {
  doc: IntegrationDocSearchResult;
  selectedDocKeys: string[];
  setSelectedDocKeys: (value: string[] | ((current: string[]) => string[])) => void;
}) {
  const checked = selectedDocKeys.includes(doc.doc_key);
  return (
    <label className="integration-doc-result">
      <input
        type="checkbox"
        checked={checked}
        onChange={(event) =>
          setSelectedDocKeys((current) =>
            event.target.checked ? [...new Set([...current, doc.doc_key])] : current.filter((key) => key !== doc.doc_key),
          )
        }
      />
      <span>
        <span className="integration-doc-title">{doc.title}</span>
        <small>{doc.provider_key} · {doc.reason} · score {doc.score}</small>
        {doc.available_integrations?.length ? <small>Подключено: {doc.available_integrations.map((item) => item.name).join(", ")}</small> : null}
      </span>
    </label>
  );
}
