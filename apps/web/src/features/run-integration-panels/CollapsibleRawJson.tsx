"use client";

import { ChevronDown, ChevronRight } from "lucide-react";
import { useState } from "react";

import { jsonText } from "../../lib/api";

type Props = { title: string; data: unknown; defaultOpen?: boolean };

export function CollapsibleRawJson({ title, data, defaultOpen }: Props) {
  const [open, setOpen] = useState(!!defaultOpen);
  const Icon = open ? ChevronDown : ChevronRight;
  return (
    <div className="rounded-md border border-border">
      <button
        type="button"
        className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm font-medium text-muted-foreground hover:bg-muted/40"
        onClick={() => setOpen((x) => !x)}
      >
        <Icon size={16} />
        {title}
      </button>
      {open ? <pre className="code-block m-0 max-h-80 overflow-auto border-t">{jsonText(data)}</pre> : null}
    </div>
  );
}
