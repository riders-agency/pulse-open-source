import Link from "next/link";
import { readFile } from "node:fs/promises";
import { join, normalize } from "node:path";

import { ClientRedirectPage } from "../../../../../features/navigation/ui/ClientRedirectPage";
import { integrationDocMarkdownURL } from "../../../../../features/public-docs/api/publicDocsApi";

function safeSlug(value: string) {
  return value.replace(/[^a-z0-9-]/gi, "").toLowerCase();
}

export default async function IntegrationDocPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  if (slug.endsWith(".md")) {
    const mdSlug = safeSlug(slug.replace(/\.md$/i, ""));
    return <ClientRedirectPage label="Переход к Markdown документации" to={integrationDocMarkdownURL(mdSlug)} />;
  }
  const safe = safeSlug(slug);
  const docsRoot = normalize(join(process.cwd(), "../../docs/integrations"));
  const filePath = normalize(join(docsRoot, `${safe}.md`));
  let body = "";
  if (safe && safe === slug && filePath.startsWith(docsRoot)) {
    body = await readFile(filePath, "utf8").catch(() => "");
  }
  if (!body) {
    return (
      <main className="docs-page">
        <h1>Документ не найден</h1>
        <Link href="/docs/integrations">К списку интеграций</Link>
      </main>
    );
  }
  const content = body.replace(/^---[\s\S]*?---\s*/m, "").trim();
  const title = content.match(/^#\s+(.+)$/m)?.[1] ?? slug;
  return (
    <main className="docs-page">
      <div className="docs-topline">
        <Link href="/docs/integrations">Интеграции</Link>
        <a href={integrationDocMarkdownURL(safe)}>Markdown для LLM</a>
      </div>
      <h1>{title}</h1>
      <pre className="docs-markdown">{content}</pre>
    </main>
  );
}
