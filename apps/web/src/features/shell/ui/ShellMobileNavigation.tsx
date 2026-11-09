"use client";

import { useState } from "react";
import { Menu, ChevronDown, ChevronRight } from "lucide-react";
import Link from "next/link";
import { Dialog, DialogContent, DialogTitle, DialogDescription, DialogTrigger } from "../../../components/ui/dialog";
import type { useShellWorkspace } from "../model/useShellWorkspace";
import { ShellWorkspaceSwitcher } from "./ShellWorkspaceSwitcher";
import { ShellNavigation } from "./ShellNavigation";
import { ShellFooter } from "./ShellFooter";

type ShellWorkspace = ReturnType<typeof useShellWorkspace>;

export function ShellMobileNavigation({ workspace }: { workspace: ShellWorkspace }) {
  const [open, setOpen] = useState(false);
  return <Dialog open={open} onOpenChange={setOpen}>
    <header className="mobile-app-bar">
      <Link className="brand-lockup" href="/dashboard"><span className="brand-mark">RP</span><strong>Riders Pulse</strong></Link>
      <DialogTrigger className="sidebar-collapse-button" aria-label="Открыть меню"><Menu size={20} /></DialogTrigger>
    </header>
    <DialogContent className="mobile-nav-dialog">
      <DialogTitle>Рабочее пространство</DialogTitle>
      <DialogDescription>Компания, проект и разделы</DialogDescription>
      <ShellWorkspaceSwitcher workspace={workspace} />
      <div onClick={(event) => { if ((event.target as Element).closest("a")) setOpen(false); }}>
        <ShellNavigation chevronDown={<ChevronDown size={15} />} chevronRight={<ChevronRight size={15} />} workspace={workspace} />
        <ShellFooter workspace={workspace} />
      </div>
    </DialogContent>
  </Dialog>;
}
