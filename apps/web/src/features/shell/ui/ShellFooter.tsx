"use client";

import Link from "next/link";
import { Activity, Code2, FileText, LogOut, Moon, Sun, UserCircle } from "lucide-react";

import { Switch } from "../../../components/ui/switch";
import type { useShellWorkspace } from "../model/useShellWorkspace";

type ShellWorkspace = ReturnType<typeof useShellWorkspace>;

export function ShellFooter({ workspace }: { workspace: ShellWorkspace }) {
  const { canUseDocs, logout, theme, toggleTheme } = workspace;

  return (
    <div className="sidebar-footer">
      <span className="sidebar-theme" title={theme === "dark" ? "Темная тема" : "Светлая тема"}>
        <Sun size={16} />
        <Switch checked={theme === "dark"} onClick={toggleTheme} />
        <Moon size={16} />
      </span>
      {canUseDocs ? (
        <Link className="sidebar-link" href="/docs">
          <FileText size={17} />
          <span>Документы</span>
        </Link>
      ) : null}
      <Link className="sidebar-link" href="/account">
        <UserCircle size={17} />
        <span>Аккаунт</span>
      </Link>
      <button type="button" className="sidebar-link sidebar-button" onClick={() => void logout()}>
        <LogOut size={17} />
        <span>Выйти</span>
      </button>
      <Link className="sidebar-status-link" href="/status" title="Статус системы">
        <Activity size={15} />
        <span>Состояние системы</span>
      </Link>
      <Link className="sidebar-status-link" href="/tech"><Code2 size={15} /><span>Как устроен Pulse</span></Link>
    </div>
  );
}
