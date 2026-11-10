import type { ActionAttempt, ActionRun } from "../../lib/api";
import type { ActionDiagnosticsWire, ActionOutputEnvelope, CriterionResultWire } from "../../contracts/workflow-step-output";

export type NormalizedPanelsPayload = {
  criteriaResults: CriterionResultWire[];
  expose: Record<string, unknown> | null;
  criteriaPassed?: boolean;
  provider: Record<string, unknown> | null;
  diagnostics: ActionDiagnosticsWire | null;
};

function parseObject(raw: unknown): Record<string, unknown> | null {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) return null;
  return raw as Record<string, unknown>;
}

export function normalizeActionOutput(raw: unknown): NormalizedPanelsPayload {
  const o = parseObject(raw) as ActionOutputEnvelope | null;
  if (!o)
    return { criteriaResults: [], expose: null, criteriaPassed: undefined, provider: null, diagnostics: null };
  const diagnostics = parseObject(o.diagnostics) as ActionDiagnosticsWire | null;
  return {
    criteriaResults: Array.isArray(o.criteria_results) ? (o.criteria_results as CriterionResultWire[]) : [],
    expose: parseObject(o.expose),
    criteriaPassed: o.criteria_passed,
    provider: null,
    diagnostics,
  };
}

export function normalizeProvider(raw: unknown): Record<string, unknown> | null {
  return parseObject(raw);
}

export function pickLatestAttempt(action: ActionRun): ActionAttempt | undefined {
  const list = action.attempts ?? [];
  if (!list.length) return undefined;
  return [...list].sort((a, b) => (b.attempt_no ?? 0) - (a.attempt_no ?? 0))[0];
}

export function normalizeDiagnostics(raw: unknown): ActionDiagnosticsWire | null {
  return parseObject(raw) as ActionDiagnosticsWire | null;
}
