"use client";

import { Trash2 } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import type { CompanyLegalEntity } from "../../../lib/api";
import { legalEntityTypeLabel, statusLabel, statusTone } from "../../../lib/client-labels";

export function CompanyLegalEntityList({
  onRemove,
  rows,
}: {
  onRemove: (id: string) => void;
  rows: CompanyLegalEntity[];
}) {
  if (rows.length === 0) return <div className="empty-state">Юрлица ещё не добавлены</div>;

  return (
    <div className="entity-list">
      {rows.map((item) => (
        <div className="entity-item" key={item.company_legal_entity_id}>
          <div className="between">
            <strong>{item.name}</strong>
            <Badge tone={statusTone(item.status)}>{statusLabel(item.status)}</Badge>
          </div>
          <div className="client-row-meta">
            <span>{legalEntityTypeLabel(item.entity_type)}</span>
            <span className="mono">ИНН {item.inn || "не указан"}</span>
            {item.kpp ? <span className="mono">КПП {item.kpp}</span> : null}
          </div>
          <div className="form-actions">
            <Button variant="ghost" size="sm" onClick={() => onRemove(item.company_legal_entity_id)}>
              <Trash2 size={16} />
              <span>Удалить</span>
            </Button>
          </div>
        </div>
      ))}
    </div>
  );
}
