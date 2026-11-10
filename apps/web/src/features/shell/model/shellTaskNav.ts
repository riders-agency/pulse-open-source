export type ShellTaskNavIcon = "bar" | "calendar" | "columns" | "list" | "timer" | "users" | "workflow";

export type ShellTaskNavItem = {
  href: string;
  label: string;
  icon: ShellTaskNavIcon;
  active: boolean;
};

export function buildShellTaskNavItems({
  canManageProcesses,
  canUse,
  pathname,
}: {
  canManageProcesses: boolean;
  canUse: (module: string) => boolean;
  pathname: string;
}) {
  const items: Array<ShellTaskNavItem & { show: boolean }> = [
    { href: "/tasks/all", label: "Все задачи", icon: "list", show: canUse("tasks"), active: pathname.startsWith("/tasks/all") },
    { href: "/tasks/boards", label: "Доски", icon: "columns", show: canUse("tasks"), active: pathname.startsWith("/tasks/boards") || pathname.startsWith("/work/board") },
    { href: "/tasks/calendar", label: "Календарь", icon: "calendar", show: canUse("tasks"), active: pathname.startsWith("/tasks/calendar") },
    { href: "/tasks/timeline", label: "План сроков", icon: "calendar", show: canUse("tasks"), active: pathname.startsWith("/tasks/timeline") },
    { href: "/tasks/time", label: "Таймтрекер", icon: "timer", show: canUse("time"), active: pathname.startsWith("/tasks/time") || pathname.startsWith("/work/time") },
    { href: "/tasks/reports", label: "Отчеты", icon: "bar", show: canUse("reports"), active: pathname.startsWith("/tasks/reports") || pathname.startsWith("/work/reports") },
    { href: "/tasks/workload", label: "Загрузка", icon: "users", show: canUse("employees"), active: pathname.startsWith("/tasks/workload") },
    { href: "/tasks/processes", label: "Процессы", icon: "workflow", show: canManageProcesses, active: pathname.startsWith("/tasks/processes") || pathname.startsWith("/work/processes") },
    { href: "/tasks/employees", label: "Сотрудники", icon: "users", show: canUse("employees"), active: pathname.startsWith("/tasks/employees") || pathname.startsWith("/work/employees") },
  ];
  return items.filter((item) => item.show).map(({ show: _show, ...item }) => item);
}
