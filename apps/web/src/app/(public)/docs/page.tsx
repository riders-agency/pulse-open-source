import Link from "next/link";

import {
  endpointInventoryURL,
  openAPIJSONURL,
  openAPIYAMLURL,
  publicDocsBase,
  workflowPatchContractMarkdownURL,
} from "../../../features/public-docs/api/publicDocsApi";

export default function DocsLandingPage() {
  const base = publicDocsBase();
  return (
    <main className="docs-page docs-landing">
      <h1 className="docs-landing-title">Документация Riders Pulse</h1>
      <p className="docs-landing-copy">
        Открытые материалы без входа в приложение. Спецификация API обслуживается бэкендом.
      </p>
      <ul className="docs-link-list stack">
        <li>
          <a href={workflowPatchContractMarkdownURL()} rel="noreferrer">
            AI workflow patch contract (Markdown для LLM)
          </a>
        </li>
        <li>
          <Link href="/docs/integrations">Документация интеграций и поиск</Link>
        </li>
        <li>
          <a href={base ? openAPIJSONURL() : "#"} rel="noreferrer">
            OpenAPI (JSON)
          </a>
        </li>
        <li>
          <a href={base ? openAPIYAMLURL() : "#"} rel="noreferrer">
            OpenAPI (YAML)
          </a>
        </li>
        <li>
          <a href={base ? endpointInventoryURL() : "#"} rel="noreferrer">
            Список endpoints
          </a>
        </li>
      </ul>
      {!base ? (
        <p className="form-error">Задайте NEXT_PUBLIC_RP_API_URL, чтобы открывались абсолютные ссылки.</p>
      ) : null}
      <p className="docs-landing-back">
        <Link href="/dashboard">Войти в приложение</Link>
      </p>
    </main>
  );
}
