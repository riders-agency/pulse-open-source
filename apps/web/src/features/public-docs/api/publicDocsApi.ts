import { apiBase, type IntegrationDocSearchResult } from "../../../lib/api";

export type { IntegrationDocSearchResult };

export function publicDocsBase() {
  return apiBase || "";
}

export function publicDocAssetURL(path: string) {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  const base = publicDocsBase();
  return base ? `${base}${normalizedPath}` : normalizedPath;
}

export function workflowPatchContractMarkdownURL() {
  return publicDocAssetURL("/api/v1/docs/integrations/workflow-ai-patch-contract.md");
}

export function openAPIJSONURL() {
  return publicDocAssetURL("/api/v1/docs/openapi.json");
}

export function openAPIYAMLURL() {
  return publicDocAssetURL("/api/v1/docs/openapi.yaml");
}

export function endpointInventoryURL() {
  return publicDocAssetURL("/api/v1/docs/endpoints");
}

export function integrationDocMarkdownURL(slug: string) {
  return publicDocAssetURL(`/api/v1/docs/integrations/${slug}.md`);
}

export function integrationDocsSearchURL(query: string, limit = 12) {
  const params = new URLSearchParams();
  params.set("q", query);
  params.set("limit", String(limit));
  return publicDocAssetURL(`/api/v1/docs/integrations/search?${params.toString()}`);
}

export async function searchIntegrationDocs(query: string) {
  const res = await fetch(integrationDocsSearchURL(query));
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const rows = await res.json();
  return Array.isArray(rows) ? (rows as IntegrationDocSearchResult[]) : [];
}
