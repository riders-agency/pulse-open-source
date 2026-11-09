"use client";

import { useState } from "react";
import Link from "next/link";
import { WorkTaskArchive } from "./work-task-archive";
import { Sheet } from "../../components/ui/sheet";
import type { WorkAccess } from "../../lib/api";
import { useWorkTaskDetail } from "./use-work-task-detail";
import { WorkTaskCardTab } from "./work-task-card-tab";
import { WorkTaskHeader } from "./work-task-header";
import {
  WorkTaskAccessTab,
  WorkTaskActivityTab,
  WorkTaskAttachmentsTab,
  WorkTaskCommentsTab,
  WorkTaskTabsNav,
  WorkTaskTimeTab,
} from "./work-task-tabs";

export function WorkTaskScreen({
  taskId,
  access,
  closeHref = "/tasks/boards",
  onClose,
  onChanged,
}: {
  taskId: string;
  access?: WorkAccess;
  closeHref?: string;
  onClose?: () => void;
  onChanged?: () => void;
}) {
  const [archivedTaskId, setArchivedTaskId] = useState("");
  const archived = archivedTaskId === taskId;
  const detail = useWorkTaskDetail({ taskId, access, onChanged });
  const { error, setError, task } = detail;

  if (!task) {
    return (
      <main className="work-task-screen">
        <section className="work-empty">
          <h2>{error ? "Задача недоступна" : "Открываем задачу"}</h2>
          <p>{error ? "Возможно, задача удалена или у вас нет доступа." : "Загружаем карточку задачи…"}</p>
          {error ? <Link href={closeHref}>Вернуться к задачам</Link> : null}
          {error ? <div className="form-error">{error}</div> : null}
        </section>
      </main>
    );
  }

  return (
    <main className="work-task-screen">
      {detail.canWrite ? <WorkTaskArchive key={task.task_id} taskId={task.task_id} title={task.title} onChanged={(removed) => { setArchivedTaskId(removed ? task.task_id : ""); onChanged?.(); }} /> : null}
      {archived ? <Link href={closeHref} onClick={onClose}>Вернуться к задачам</Link> : null}
      <div hidden={archived}>
      <WorkTaskHeader
        activeTaskTimer={detail.activeTaskTimer}
        canUseTime={detail.canUseTime}
        closeHref={closeHref}
        effectiveAccess={detail.effectiveAccess}
        now={detail.now}
        onClose={onClose}
        setError={setError}
        startTimer={detail.startTimer}
        stopTimer={detail.stopTimer}
        task={task}
      />
      {error ? <div className="form-error">{error}</div> : null}
      {detail.timerNotice ? <div className="work-success-note">{detail.timerNotice}</div> : null}
      <WorkTaskTabsNav setTab={detail.setTab} tab={detail.tab} task={task} />
      {detail.tab === "card" ? <WorkTaskCardTab {...detail} task={task} /> : null}
      {detail.tab === "comments" ? <WorkTaskCommentsTab {...detail} task={task} /> : null}
      {detail.tab === "time" ? <WorkTaskTimeTab {...detail} task={task} /> : null}
      {detail.tab === "activity" ? <WorkTaskActivityTab task={task} /> : null}
      {detail.tab === "access" ? <WorkTaskAccessTab {...detail} task={task} /> : null}
      {detail.tab === "attachments" ? <WorkTaskAttachmentsTab {...detail} task={task} /> : null}
      </div>
    </main>
  );
}

export function WorkTaskDrawer({
  taskId,
  access,
  onClose,
  onChanged,
}: {
  taskId: string;
  access?: WorkAccess;
  onClose: () => void;
  onChanged?: () => void;
}) {
  return (
    <Sheet open={Boolean(taskId)} className="work-task-fullscreen-sheet">
      <WorkTaskScreen taskId={taskId} access={access} onClose={onClose} onChanged={onChanged} />
    </Sheet>
  );
}
