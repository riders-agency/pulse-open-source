import type { IntegrationDocSearchResult } from "../../../lib/api";
import { aiWorkflowDocsURL } from "../registry/actions";

export { actionCatalogForAI } from "./actionCatalogForAI";

export function compactIntegrationDoc(doc: IntegrationDocSearchResult) {
  return {
    doc_key: doc.doc_key,
    title: doc.title,
    summary: doc.summary,
    provider_key: doc.provider_key,
    action_type_keys: doc.action_type_keys,
    dependency_type_keys: doc.dependency_type_keys,
    reason: doc.reason,
    markdown_url: doc.markdown_url,
    available_integrations: doc.available_integrations ?? [],
  };
}

export function workflowAIPrompt({
  task,
  contextJSON,
  docs,
}: {
  task: string;
  contextJSON: string;
  docs: IntegrationDocSearchResult[];
}) {
  const docLines = docs.length
    ? docs
        .map((doc, index) => {
          const integrations = doc.available_integrations?.length
            ? `\n  Project integrations: ${doc.available_integrations
                .map((item) => `${item.name} (${item.dependency_type_key}, ${item.status}, id=${item.action_dependency_id})`)
                .join("; ")}`
            : "";
          return `${index + 1}. ${doc.title}
  URL: ${doc.markdown_url}
  Provider: ${doc.provider_key}
  Actions: ${doc.action_type_keys.join(", ") || "-"}
  Why selected: ${doc.reason}
  Summary: ${doc.summary}${integrations}`;
        })
        .join("\n\n")
    : "No integration docs were selected. Use only the main contract and existing workflow context.";

  return `Ты редактируешь Riders Pulse workflow.

Главный контракт patch:
${aiWorkflowDocsURL}

Если можешь открывать ссылки, сначала прочитай главный контракт и только релевантные документы интеграций ниже.

Задача пользователя:
${task.trim() || "<пользователь не написал отдельное описание; опирайся на текущий контекст>"}

Как выбирать интеграции:
1. Распознай намерение пользователя.
2. Используй только документы интеграций, которые подходят под намерение.
3. Если нужна проектная интеграция, используй только перечисленные Project integrations.
4. Не придумывай UUID интеграций, токены и секреты.
5. Если нужной интеграции нет в Project integrations, оставь dependency id пустым и добавь это в change_description.

Релевантные документы интеграций:
${docLines}

Текущий context JSON:
\`\`\`json
${contextJSON}
\`\`\`

Верни только валидный JSON riders-pulse-workflow-patch-v1 без markdown.
Заполни change_description кратко; можно пустой строкой.
Patch должен быть минимальным: только operations, которые нужны для задачи.`;
}
