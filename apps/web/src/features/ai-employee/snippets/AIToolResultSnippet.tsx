import {
  asRecord,
  entityLinkFromPayload,
  recordList,
  textValue,
  toolResultTitle,
} from "../aiEmployeeUtils";

export function AIToolResultSnippet({ payload }: { payload?: Record<string, unknown> }) {
  const entityLink = entityLinkFromPayload(payload);
  const output = asRecord(payload?.output_json || payload?.output);
  const kind = textValue(output.kind) || textValue(payload?.kind);
  const answer = textValue(output.answer) || textValue(payload?.answer);
  if (payload?.kind === "tool_results") {
    const results = recordList(payload.results_json)
      .map((result) => asRecord(result.output_json))
      .filter((item) => textValue(item.kind) !== "task_search_result");
    if (!results.length) return null;
    return (
      <div className="ai-result-card" data-testid="ai-tool-result-card">
        <strong>Результаты чтения</strong>
        {results.map((item, index) => (
          <div className="ai-result-section" key={`${textValue(item.kind) || "result"}-${index}`}>
            <strong>{resultHeading(textValue(item.kind), payload)}</strong>
            {textValue(item.answer) ? <p>{textValue(item.answer)}</p> : null}
            <ResultFields output={item} />
          </div>
        ))}
      </div>
    );
  }
  const shouldRenderGeneric = payload?.kind === "tool_result" && kind !== "task_search_result";
  if (shouldRenderGeneric) {
    return (
      <div className="ai-result-card" data-testid="ai-tool-result-card">
        <strong>{resultHeading(kind, payload)}</strong>
        {answer ? <p>{answer}</p> : null}
        <ResultFields output={output} />
        {entityLink ? <a className="ai-entity-link" href={entityLink.url}>{entityLink.label}</a> : null}
      </div>
    );
  }
  if (!entityLink) return null;
  return (
    <div className="ai-result-card" data-testid="ai-tool-result-card">
      <strong>{toolResultTitle(payload)}</strong>
      <a className="ai-entity-link" href={entityLink.url}>
        {entityLink.label}
      </a>
    </div>
  );
}

function resultHeading(kind: string, payload?: Record<string, unknown>) {
  switch (kind) {
    case "page_context":
      return "Контекст экрана";
    case "client_search_result":
      return "Найденные клиенты";
    case "client_count":
      return "Количество клиентов";
    case "payment_due":
      return "Оплата клиента";
    case "workflow_read_draft":
      return "Сводка сценария";
    case "calculator_result":
      return "Калькулятор";
    default:
      return toolResultTitle(payload);
  }
}

function ResultFields({ output }: { output: Record<string, unknown> }) {
  const kind = textValue(output.kind);
  if (kind === "client_search_result") {
    const matches = recordList(output.matches);
    return matches.length ? (
      <div className="ai-result-list">
        {matches.map((match) => {
          const id = textValue(match.client_id);
          const label = textValue(match.name) || "Клиент";
          return id ? <a className="ai-entity-link" href={`/clients/${id}`} key={id}>{label}</a> : <span key={label}>{label}</span>;
        })}
      </div>
    ) : null;
  }
  if (kind === "client_count") {
    return <div className="ai-result-field"><span>Клиенты</span><strong>{String(output.count ?? 0)}</strong></div>;
  }
  if (kind === "payment_due") {
    return (
      <div className="ai-result-fields">
        {textValue(output.client_name) ? <div className="ai-result-field"><span>Клиент</span><strong>{textValue(output.client_name)}</strong></div> : null}
        {textValue(output.due_date) ? <div className="ai-result-field"><span>Срок</span><strong>{textValue(output.due_date)}</strong></div> : null}
        {output.amount ? <div className="ai-result-field"><span>Сумма</span><strong>{String(output.amount)} {textValue(output.currency)}</strong></div> : null}
      </div>
    );
  }
  if (kind === "workflow_read_draft") {
    return (
      <div className="ai-result-fields">
        <div className="ai-result-field"><span>Действия</span><strong>{String(output.actions_count ?? 0)}</strong></div>
        <div className="ai-result-field"><span>Связи</span><strong>{String(output.edges_count ?? 0)}</strong></div>
        <div className="ai-result-field"><span>Триггеры</span><strong>{String(output.trigger_count ?? 0)}</strong></div>
      </div>
    );
  }
  if (kind === "page_context") {
    return (
      <div className="ai-result-fields">
        <div className="ai-result-field"><span>Тип страницы</span><strong>{textValue(output.page_kind) || "unknown"}</strong></div>
        <div className="ai-result-field"><span>Маршрут</span><strong>{textValue(output.route) || "unknown"}</strong></div>
      </div>
    );
  }
  return null;
}
