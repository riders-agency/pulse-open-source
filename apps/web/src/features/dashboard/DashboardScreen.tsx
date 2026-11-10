"use client";

import { useDashboardModel } from "./model/useDashboardModel";
import { DashboardView } from "./ui/DashboardView";

export function DashboardScreen() {
  return <DashboardView model={useDashboardModel()} />;
}
