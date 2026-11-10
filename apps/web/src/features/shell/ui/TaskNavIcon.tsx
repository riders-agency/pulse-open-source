import { BarChart3, CalendarDays, Columns3, ListChecks, Timer, Users, Workflow } from "lucide-react";

import type { ShellTaskNavIcon } from "../model/useShellWorkspace";

export function TaskNavIcon({ icon }: { icon: ShellTaskNavIcon }) {
  switch (icon) {
    case "bar":
      return <BarChart3 size={15} />;
    case "calendar":
      return <CalendarDays size={15} />;
    case "columns":
      return <Columns3 size={15} />;
    case "timer":
      return <Timer size={15} />;
    case "users":
      return <Users size={15} />;
    case "workflow":
      return <Workflow size={15} />;
    case "list":
    default:
      return <ListChecks size={15} />;
  }
}
