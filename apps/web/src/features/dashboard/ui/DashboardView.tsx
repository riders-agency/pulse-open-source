import { Badge } from "../../../components/ui/badge";
import { Card, CardContent } from "../../../components/ui/card";
import type { DashboardModel } from "../model/useDashboardModel";
import { DashboardContent } from "./DashboardContent";
import { DashboardHeader } from "./DashboardHeader";
import { DashboardLoginCard } from "./DashboardLoginCard";

type DashboardViewProps = {
  model: DashboardModel;
};

export function DashboardView({ model }: DashboardViewProps) {
  return (
    <main className="page">
      <DashboardHeader activeCompanyId={model.activeCompanyId} activeProjectId={model.activeProjectId} />
      {model.error ? (
        <Card>
          <CardContent>
            <Badge tone="danger">{model.error}</Badge>
          </CardContent>
        </Card>
      ) : null}
      {model.token ? <DashboardContent model={model} /> : <DashboardLoginCard model={model} />}
    </main>
  );
}
