import type { ActionRun } from "../../../lib/api";

export function OfficeUsagePanel({ actions }: { actions: ActionRun[] }) {
  const calls = actions.flatMap((action) => {
    const provider = action.provider_response_json as { usage?: { modelId?: string; inputTokens?: string; outputTokens?: string }; executed_nodes?: string[] } | undefined;
    return provider?.usage ? [{ action, ...provider }] : [];
  });
  if (calls.length === 0) return null;
  const input = calls.reduce((total, call) => total + Number(call.usage?.inputTokens ?? 0), 0);
  const output = calls.reduce((total, call) => total + Number(call.usage?.outputTokens ?? 0), 0);
  return <section className="office-decision" aria-label="Расход модели"><strong>Вызовы ИИ-сотрудника</strong><p>{calls.length} вызова · {input} входных / {output} выходных токенов</p><small>{calls[0].usage?.modelId} · Python / LangGraph через gRPC · данные из ответов провайдера</small></section>;
}
