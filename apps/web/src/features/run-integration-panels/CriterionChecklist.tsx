"use client";

import { CheckCircle2, XCircle } from "lucide-react";

import { Badge } from "../../components/ui/badge";
import type { CriterionResultWire } from "../../contracts/workflow-step-output";

export function CriterionChecklist({ results }: { results: CriterionResultWire[] }) {
  if (!results.length) return null;
  return (
    <div className="stack gap-2">
      <div className="text-sm font-medium text-foreground">Критерии успеха</div>
      <ul className="stack gap-2">
        {results.map((c, idx) => (
          <li key={`${c.type}-${idx}`} className="flex flex-wrap items-start gap-2 text-sm">
            {c.passed ? <CheckCircle2 className="mt-0.5 shrink-0 text-emerald-600" size={16} /> : <XCircle className="mt-0.5 shrink-0 text-destructive" size={16} />}
            <div className="min-w-0 flex-1">
              <Badge tone={c.passed ? "success" : "danger"}>
                {c.label || c.type}
              </Badge>
              {c.detail ? <div className="muted mt-1 text-xs">{c.detail}</div> : null}
              {(c.expected || c.actual || c.source) ? (
                <div className="muted mt-1 text-xs">
                  {c.expected ? <>Ожидали: <code>{c.expected}</code>. </> : null}
                  {c.actual ? <>Получили: <code>{c.actual}</code>. </> : null}
                  {c.source ? <>Источник: <code>{c.source}</code>.</> : null}
                </div>
              ) : null}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
