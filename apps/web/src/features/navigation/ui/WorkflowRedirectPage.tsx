"use client";

import { ClientRedirectPage } from "./ClientRedirectPage";

export function WorkflowRedirectPage({ workflowId }: { workflowId: string }) {
  return <ClientRedirectPage label="Переход к редактору сценария" to={`/workflows/${workflowId}/edit`} />;
}
