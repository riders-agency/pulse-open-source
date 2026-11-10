"use client";

import { useState } from "react";
import { Trash2, Undo2 } from "lucide-react";
import { Button } from "../../components/ui/button";
import { Dialog, DialogContent, DialogTitle, DialogDescription } from "../../components/ui/dialog";
import { archiveWorkTask, restoreWorkTask } from "./api/workApi";
import { usePulseStore } from "../../store/pulse-store";

export function WorkTaskArchive({ taskId, title, onChanged }: { taskId: string; title: string; onChanged?: (archived: boolean) => void }) {
  const token = usePulseStore((s) => s.token);
  const [open, setOpen] = useState(false);
  const [archived, setArchived] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const change = async (remove: boolean) => {
    if (!token || busy) return;
    setBusy(true); setError("");
    try {
      await (remove ? archiveWorkTask(taskId, token) : restoreWorkTask(taskId, token));
      setArchived(remove); setOpen(false); onChanged?.(remove);
      window.dispatchEvent(new Event("pulse:ai-entity-changed"));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось изменить задачу");
    } finally { setBusy(false); }
  };
  return <div className="stack">
    {archived ? <div className="notice" role="status">Задача удалена из рабочих списков. <Button variant="ghost" disabled={busy} onClick={() => void change(false)}><Undo2 size={15} />Восстановить</Button></div> : <Button variant="ghost" onClick={() => setOpen(true)}><Trash2 size={15} />Удалить задачу</Button>}
    {!open && error ? <p role="alert" className="form-error">{error}</p> : null}
    <Dialog open={open} onOpenChange={(value) => { if (!busy) setOpen(value); }}>
      <DialogContent>
        <DialogTitle>Удалить задачу «{title}»?</DialogTitle>
        <DialogDescription>Задача переместится в удалённые. Комментарии, файлы и учёт времени сохранятся. Восстановить её можно в разделе «Все задачи → Удалённые». Перед удалением остановите активные таймеры.</DialogDescription>
        {error ? <p role="alert" className="form-error">{error}</p> : null}
        <div className="row"><Button variant="secondary" disabled={busy} onClick={() => setOpen(false)}>Отмена</Button><Button variant="destructive" disabled={busy} onClick={() => void change(true)}>{busy ? "Удаляем…" : "Удалить"}</Button></div>
      </DialogContent>
    </Dialog>
  </div>;
}
