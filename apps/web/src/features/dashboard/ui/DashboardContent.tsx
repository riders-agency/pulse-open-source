import type { DashboardModel } from "../model/useDashboardModel";
import { DashboardErrorsCard } from "./DashboardErrorsCard";
import { DashboardMetrics } from "./DashboardMetrics";
import { DashboardQuickActions } from "./DashboardQuickActions";
import { DashboardRunsTable } from "./DashboardRunsTable";

type DashboardContentProps = {
  model: DashboardModel;
};

export function DashboardContent({ model }: DashboardContentProps) {
  return (
    <div className="stack">
      <DashboardMetrics
        activeCompanyName={model.activeCompany?.name}
        activeProjectName={model.activeProject?.name}
        activeWorkflowCount={model.activeWorkflowCount}
      />
      <section className="grid-2">
        <DashboardQuickActions activeCompanyId={model.activeCompanyId} activeProjectId={model.activeProjectId} />
        <DashboardErrorsCard latestErrors={model.latestErrors} />
      </section>
      <DashboardRunsTable runs={model.runs} />
    </div>
  );
}
