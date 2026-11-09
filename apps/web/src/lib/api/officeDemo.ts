export type OfficeNode = { id: string; title: string; detail: string; tool: string };
export type OfficeArtifact = { title?: string; body?: string; path?: string; payer?: string; amount_minor?: number; number?: string; status: string };
export type OfficeEvent = {
  node: string; status: string; text: string; at_ms: number; tool?: string; actor?: string;
  usage?: { provider_key: string; model_id: string; input_tokens: number; output_tokens: number };
  mail?: { sender: string; subject: string; body: string };
  drafts?: Record<string, OfficeArtifact>; results?: Record<string, OfficeArtifact>;
};
export type OfficeState = { status: string; events: OfficeEvent[]; nodes: OfficeNode[]; edges: string[][]; results: Record<string, OfficeArtifact> };

export async function officeDemoRequest(endpoint: string, action = "state", data?: object): Promise<OfficeState> {
  const response = await fetch(`${endpoint}/${action}`, {
    method: data ? "POST" : "GET", mode: "cors", credentials: "omit", cache: "no-store",
    headers: data ? { "Content-Type": "application/json" } : undefined,
    body: data ? JSON.stringify(data) : undefined, signal: AbortSignal.timeout(25000),
  });
  if (!response.ok) throw new Error("Не удалось выполнить действие. Проверьте подключение к демонстрации.");
  return response.json() as Promise<OfficeState>;
}
