"use client";

import { Table } from "../../components/ui/table";

export function ForwardedKeysTable({
  expose,
  compact,
}: {
  expose: Record<string, unknown> | null;
  /** Скрывает заголовок и подсказку (когда уже есть заголовок карточки). */
  compact?: boolean;
}) {
  if (!expose || Object.keys(expose).length === 0) return null;
  const rows = Object.entries(expose).sort(([a], [b]) => a.localeCompare(b));
  return (
    <div className="stack gap-2">
      {compact ? null : (
        <>
          <div className="text-sm font-medium text-foreground">Выход для следующего шага (expose)</div>
          <p className="muted text-xs">
            Используйте в шаблонах вида{" "}
            <code>{`{{ steps.<ключ_действия>.<поле> }}`}</code> — после поддержки контекста в ранере.
          </p>
        </>
      )}
      <Table>
        <thead>
          <tr>
            <th>Ключ</th>
            <th>Значение</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(([k, v]) => (
            <tr key={k}>
              <td>
                <code className="text-xs">{k}</code>
              </td>
              <td className="break-all font-mono text-xs">{serializeCell(v)}</td>
            </tr>
          ))}
        </tbody>
      </Table>
    </div>
  );
}

function serializeCell(v: unknown): string {
  if (v === null || v === undefined) return "—";
  if (typeof v === "string" || typeof v === "number" || typeof v === "boolean") return String(v);
  try {
    return JSON.stringify(v);
  } catch {
    return String(v);
  }
}
