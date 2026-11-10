import { readFile } from "node:fs/promises";
import { join, normalize } from "node:path";

export const dynamic = "force-dynamic";

function safeSlug(value: string) {
  return value.replace(/[^a-z0-9-]/gi, "").toLowerCase();
}

export async function GET(_: Request, { params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const safe = safeSlug(slug);
  if (!safe || safe !== slug) return new Response("not found", { status: 404 });

  const docsRoot = normalize(join(process.cwd(), "../../docs/integrations"));
  const filePath = normalize(join(docsRoot, `${safe}.md`));
  if (!filePath.startsWith(docsRoot)) return new Response("not found", { status: 404 });

  try {
    const body = await readFile(filePath, "utf8");
    return new Response(body, {
      headers: {
        "Content-Type": "text/markdown; charset=utf-8",
        "Cache-Control": "no-cache, must-revalidate",
      },
    });
  } catch {
    return new Response("not found", { status: 404 });
  }
}
