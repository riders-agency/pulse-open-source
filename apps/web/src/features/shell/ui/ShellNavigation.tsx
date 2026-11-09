"use client";

import {
  BriefcaseBusiness,
  Building2,
  CreditCard,
  Home,
  Landmark,
  ListChecks,
  MessageCircle,
  Plug,
  Users,
  Workflow,
} from "lucide-react";
import type { ReactNode } from "react";

import type { useShellWorkspace } from "../model/useShellWorkspace";
import { SidebarLink } from "./SidebarLink";
import { TaskNavIcon } from "./TaskNavIcon";

type ShellWorkspace = ReturnType<typeof useShellWorkspace>;

export function ShellNavigation({
  chevronDown,
  chevronRight,
  workspace,
}: {
  chevronDown: ReactNode;
  chevronRight: ReactNode;
  workspace: ShellWorkspace;
}) {
  const {
    activeCompany,
    activeCompanyId,
    activeProject,
    activeProjectId,
    canManageProcesses,
    canUse,
    canUseWorkspaceModules,
    pathname,
    taskNavItems,
    tasksActive,
    tasksNavOpen,
    toggleTasksNav,
  } = workspace;

  return (
    <nav className="sidebar-nav" aria-label="Основная навигация">
      {canUseWorkspaceModules ? (
        <SidebarLink href="/dashboard" icon={<Home size={17} />} active={pathname === "/dashboard"}>
          Обзор
        </SidebarLink>
      ) : null}
      {canUseWorkspaceModules ? (
        <SidebarLink href="/companies" icon={<Building2 size={17} />} active={pathname.startsWith("/companies") && !activeCompanyId}>
          Компании
        </SidebarLink>
      ) : null}
      {activeCompanyId && canUseWorkspaceModules ? (
        <>
          <div className="sidebar-section">
            <span>{activeCompany?.name ?? "Компания"}</span>
          </div>
          <SidebarLink
            href="/company-legal-entities"
            icon={<Landmark size={17} />}
            active={pathname === "/company-legal-entities" || pathname.includes(`/companies/${activeCompanyId}/legal-entities`)}
          >
            Наши юрлица
          </SidebarLink>
          <SidebarLink
            href="/projects"
            icon={<BriefcaseBusiness size={17} />}
            active={pathname === "/projects" || pathname === "/project" || pathname.includes(`/companies/${activeCompanyId}/projects`)}
          >
            Проекты
          </SidebarLink>
        </>
      ) : null}
      {activeProjectId ? (
        <>
          <div className="sidebar-section">
            <span>{activeProject?.name ?? "Проект"}</span>
          </div>
          {canManageProcesses ? (
            <SidebarLink
              href="/project-workflows"
              icon={<Workflow size={17} />}
              active={pathname.startsWith("/project-workflows") || pathname.includes(`/projects/${activeProjectId}/workflows`) || pathname.startsWith("/workflows/")}
            >
              Сценарии
            </SidebarLink>
          ) : null}
          {taskNavItems.length ? (
            <>
              <button
                type="button"
                className={`sidebar-link sidebar-task-toggle ${tasksActive ? "is-active" : ""}`}
                onClick={toggleTasksNav}
                aria-expanded={tasksNavOpen}
                aria-controls="sidebar-task-nav"
              >
                <ListChecks size={17} />
                <span>Задачи</span>
                {tasksNavOpen ? chevronDown : chevronRight}
              </button>
              {tasksNavOpen ? (
                <div className="sidebar-subnav" id="sidebar-task-nav" aria-label="Задачи">
                  {taskNavItems.map((item) => (
                    <SidebarLink href={item.href} icon={<TaskNavIcon icon={item.icon} />} active={item.active} key={item.href}>
                      {item.label}
                    </SidebarLink>
                  ))}
                </div>
              ) : null}
            </>
          ) : null}
          {canUse("clients") ? (
            <SidebarLink href="/clients" icon={<Users size={17} />} active={pathname === "/clients" || pathname === "/client" || pathname.startsWith("/clients/")}>
              Клиенты
            </SidebarLink>
          ) : null}
          {canUse("reports") ? (
            <SidebarLink
              href="/project-payments"
              icon={<CreditCard size={17} />}
              active={pathname === "/project-payments" || pathname === "/payment-action" || pathname.includes(`/projects/${activeProjectId}/payments`) || pathname.startsWith("/payment-actions/")}
            >
              Оплаты
            </SidebarLink>
          ) : null}
          {canManageProcesses ? (
            <SidebarLink href="/site-chat" icon={<MessageCircle size={17} />} active={pathname === "/site-chat" || pathname.startsWith("/site-chat/")}>
              Чат на сайте
            </SidebarLink>
          ) : null}
          {canManageProcesses ? (
            <SidebarLink
              href="/project-integrations"
              icon={<Plug size={17} />}
              active={pathname.startsWith("/project-integrations") || pathname.includes(`/projects/${activeProjectId}/integrations`)}
            >
              Интеграции
            </SidebarLink>
          ) : null}
        </>
      ) : null}
    </nav>
  );
}
