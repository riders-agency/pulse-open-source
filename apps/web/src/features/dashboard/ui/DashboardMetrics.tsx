type DashboardMetricsProps = {
  activeCompanyName?: string;
  activeProjectName?: string;
  activeWorkflowCount: number;
};

export function DashboardMetrics({
  activeCompanyName,
  activeProjectName,
  activeWorkflowCount,
}: DashboardMetricsProps) {
  return (
    <section className="grid-3">
      <div className="metric">
        <span>Активная компания</span>
        <strong>{activeCompanyName ?? "Не выбрана"}</strong>
      </div>
      <div className="metric">
        <span>Активный проект</span>
        <strong>{activeProjectName ?? "Не выбран"}</strong>
      </div>
      <div className="metric">
        <span>Активные сценарии</span>
        <strong>{activeWorkflowCount}</strong>
      </div>
    </section>
  );
}
