"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";
import { BarChart3, CalendarDays, Columns3, Gauge, ListChecks, Timer, Users, Workflow } from "lucide-react";

import type { WorkAccess } from "../../lib/api";
import { visible } from "./work-utils";

type Item = { href: string; label: string; module: string; icon: ReactNode };

const items: Item[] = [
  { href: "/tasks", label: "Обзор", module: "dashboard", icon: <Gauge size={16} /> },
  { href: "/tasks/all", label: "Все задачи", module: "tasks", icon: <ListChecks size={16} /> },
  { href: "/tasks/boards", label: "Доски", module: "tasks", icon: <Columns3 size={16} /> },
  { href: "/tasks/calendar", label: "Календарь", module: "tasks", icon: <CalendarDays size={16} /> },
  { href: "/tasks/timeline", label: "План сроков", module: "tasks", icon: <CalendarDays size={16} /> },
  { href: "/tasks/workload", label: "Загрузка", module: "employees", icon: <Users size={16} /> },
  { href: "/tasks/time", label: "Таймтрекер", module: "time", icon: <Timer size={16} /> },
  { href: "/tasks/reports", label: "Отчеты", module: "reports", icon: <BarChart3 size={16} /> },
  { href: "/tasks/processes", label: "Процессы", module: "workflows", icon: <Workflow size={16} /> },
];

export function WorkFrame({
  access,
  title,
  description,
  actions,
  children,
}: {
  access?: WorkAccess;
  title: string;
  description: string;
  actions?: ReactNode;
  children: ReactNode;
}) {
  const pathname = usePathname();
  const allowed = items.filter((item) => visible(access, item.module) || (item.href === "/tasks/boards" && visible(access, "tasks")));
  return (
    <main className="work-page" data-feature="work">
      <header className="work-topbar">
        <div className="stack-tight">
          <span className="work-kicker">{access?.project_name || "Трекер задач"}</span>
          <h1>{title}</h1>
          <p>{description}</p>
        </div>
        <div className="work-actions">{actions}</div>
      </header>
      <nav className="work-nav" aria-label="Трекер задач">
        {allowed.map((item) => (
          <Link className={pathname === item.href ? "is-active" : ""} href={item.href} key={item.href}>
            {item.icon}
            <span>{item.label}</span>
          </Link>
        ))}
      </nav>
      {children}
    </main>
  );
}

export function WorkNoProject() {
  return (
    <main className="work-page" data-feature="work">
      <section className="work-empty">
        <h1>Выберите проект</h1>
        <p>Work Tracker работает внутри проекта, чтобы задачи, время и права не смешивались между клиентами.</p>
        <Link className="action-link" href="/projects">
          К проектам
        </Link>
      </section>
    </main>
  );
}
