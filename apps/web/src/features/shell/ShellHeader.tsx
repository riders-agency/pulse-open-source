"use client";

import Link from "next/link";
import {
  ChevronDown,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";

import { useShellWorkspace } from "./model/useShellWorkspace";
import { ShellMobileNavigation } from "./ui/ShellMobileNavigation";
import { ShellFooter } from "./ui/ShellFooter";
import { ShellNavigation } from "./ui/ShellNavigation";
import { ShellWorkspaceSwitcher } from "./ui/ShellWorkspaceSwitcher";

export function ShellHeader() {
  const workspace = useShellWorkspace();
  const { sidebarCollapsed, toggleSidebar } = workspace;

  return (
    <>
    <ShellMobileNavigation workspace={workspace} />
    <aside className="app-sidebar">
      <div className="sidebar-top">
        <Link className="brand-lockup" href="/dashboard" aria-label="Riders Pulse">
          <span className="brand-mark">RP</span>
          <span className="brand-title">
            <strong>Riders Pulse</strong>
          </span>
        </Link>
        <button
          type="button"
          className="sidebar-collapse-button"
          onClick={toggleSidebar}
          aria-label={sidebarCollapsed ? "Развернуть меню" : "Свернуть меню"}
          title={sidebarCollapsed ? "Развернуть" : "Свернуть"}
        >
          {sidebarCollapsed ? <ChevronRight size={18} /> : <ChevronLeft size={18} />}
        </button>
      </div>

      <ShellWorkspaceSwitcher workspace={workspace} />
      <ShellNavigation chevronDown={<ChevronDown size={15} />} chevronRight={<ChevronRight size={15} />} workspace={workspace} />
      <ShellFooter workspace={workspace} />
    </aside>
    </>
  );
}
