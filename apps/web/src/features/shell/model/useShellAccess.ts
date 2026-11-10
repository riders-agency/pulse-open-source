"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

import { type WorkAccess } from "../../../lib/api";
import { getShellWorkAccess } from "../api/shellApi";
import { buildShellTaskNavItems } from "./shellTaskNav";

export function useShellAccess({
  activeProjectId,
  pathname,
  token,
}: {
  activeProjectId: string;
  pathname: string;
  token: string | null;
}) {
  const [workAccess, setWorkAccess] = useState<WorkAccess | null>(null);

  useEffect(() => {
    if (!token || !activeProjectId) {
      setWorkAccess(null);
      return;
    }
    let cancelled = false;
    getShellWorkAccess(activeProjectId, token)
      .then((access) => {
        if (!cancelled) setWorkAccess(access);
      })
      .catch(() => {
        if (!cancelled) setWorkAccess(null);
      });
    return () => {
      cancelled = true;
    };
  }, [activeProjectId, token]);

  const processScopedOnly = Boolean(workAccess && !workAccess.all_workflows && !workAccess.can_manage_processes);
  const canUse = useCallback((module: string) => Boolean(workAccess?.visible_modules.includes(module)), [workAccess]);
  const canManageProcesses = Boolean(workAccess?.can_manage_processes);
  const canUseWorkspaceModules = Boolean(workAccess && !processScopedOnly);
  const canUseDocs = canUseWorkspaceModules && canManageProcesses;
  const tasksActive = pathname === "/tasks" || pathname.startsWith("/tasks/") || pathname.startsWith("/work");
  const taskNavItems = useMemo(() => buildShellTaskNavItems({ canManageProcesses, canUse, pathname }), [canManageProcesses, canUse, pathname]);

  return {
    canManageProcesses,
    canUse,
    canUseDocs,
    canUseWorkspaceModules,
    taskNavItems,
    tasksActive,
  };
}
