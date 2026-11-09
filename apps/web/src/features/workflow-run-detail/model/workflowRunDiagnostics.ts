import type { ActionDiagnosticsWire } from "../../../contracts/workflow-step-output";
import type { ActionRun } from "../../../lib/api";
import { diagnosticsForActionRun } from "../../run-integration-panels/diagnostics/resolve";
import { normalizeActionOutput } from "../../run-integration-panels/parseOutputs";
import { statusTone } from "./workflowRunLabels";

export function diagnosticsForAction(action?: ActionRun | null): ActionDiagnosticsWire | null {
  return diagnosticsForActionRun(action);
}

export function diagnosticCriteria(action?: ActionRun | null) {
  const d = diagnosticsForAction(action);
  const out = normalizeActionOutput(action?.output_json);
  return d?.criteria_results?.length ? d.criteria_results : out.criteriaResults;
}

export function diagnosticTone(status?: string, d?: ActionDiagnosticsWire | null): "success" | "warning" | "danger" | "info" | "muted" {
  if (d?.status === "retryable" || d?.retryable) return "warning";
  return statusTone(status);
}

export function objectSize(value: unknown) {
  if (!value || typeof value !== "object") return 0;
  if (Array.isArray(value)) return value.length;
  return Object.keys(value as Record<string, unknown>).length;
}

export function dataBadge(value: unknown) {
  if (value == null) return "пусто";
  if (Array.isArray(value)) return `${value.length} items`;
  if (typeof value === "object") return `${objectSize(value)} keys`;
  return typeof value;
}
