import type { ActionDiagnosticsWire } from "../../../contracts/workflow-step-output";
import type { ActionRun } from "../../../lib/api";

export type ProviderDiagnosticsContext = {
  action: ActionRun;
  provider: Record<string, unknown> | null;
  statusCode: number;
  text: string;
  failed: boolean;
};

export type ProviderDiagnosticsAnalyzer = (ctx: ProviderDiagnosticsContext) => ActionDiagnosticsWire | null;
