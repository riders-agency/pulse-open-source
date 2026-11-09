"use client";

import { useEffect, useState } from "react";

export function useShellChromeState(pathname: string, theme: string) {
  const [tasksNavOpen, setTasksNavOpen] = useState(true);

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
  }, [theme]);

  useEffect(() => {
    if (pathname === "/tasks" || pathname.startsWith("/tasks/") || pathname.startsWith("/work")) {
      setTasksNavOpen(true);
    }
  }, [pathname]);

  return {
    tasksNavOpen,
    toggleTasksNav: () => setTasksNavOpen((value) => !value),
  };
}
