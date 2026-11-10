import type { ReactNode } from "react";

import { DemoNotice } from "../../demo/DemoNotice";
import { ApiErrorDialog } from "../../../components/api-error-dialog";
import { ProjectAIEmployeeWidget } from "../../ai-employee/ProjectAIEmployeeWidget";
import { ShellHeader } from "../../shell/ShellHeader";

export function AuthenticatedAppShell({ children, sidebarCollapsed }: { children: ReactNode; sidebarCollapsed: boolean }) {
  return (
    <div className={`app-shell ${sidebarCollapsed ? "is-sidebar-collapsed" : ""}`}>
      <ShellHeader />
      <div className="app-main"><DemoNotice />{children}</div>
      <ProjectAIEmployeeWidget />
      <ApiErrorDialog />
    </div>
  );
}
