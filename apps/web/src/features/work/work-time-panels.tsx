import { Square } from "lucide-react";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Table } from "../../components/ui/table";
import type { WorkTimeEntry, WorkTimer } from "../../lib/api";
import { minutesLabel, shortDate, timerClock, timeSourceLabel } from "./work-utils";

export function ActiveWorkTimers({
  now,
  onStop,
  timers,
}: {
  now: number;
  onStop: (taskId: string) => void;
  timers: WorkTimer[];
}) {
  if (!timers.length) return null;
  return (
    <section className="work-active-timers">
      {timers.map((row) => (
        <article key={row.timer_session_id}>
          <div>
            <strong>{row.task_title}</strong>
            <span>Таймер идет параллельно с другими активными задачами</span>
          </div>
          <Badge tone="muted">{timerClock(row.started_at, row.elapsed_minutes, now)}</Badge>
          <Button variant="destructive" size="sm" onClick={() => onStop(row.task_id)}><Square size={15} /> Стоп</Button>
        </article>
      ))}
    </section>
  );
}

export function WorkTimeEntriesTable({ entries }: { entries: WorkTimeEntry[] }) {
  return (
    <Table>
      <thead>
        <tr><th>Дата</th><th>Задача</th><th>Сотрудник</th><th>Клиент</th><th>Время</th><th>Тип</th></tr>
      </thead>
      <tbody>
        {entries.map((entry) => (
          <tr key={entry.time_entry_id}>
            <td>{shortDate(entry.entry_date)}</td>
            <td><strong>{entry.task_title || "Задача без названия"}</strong></td>
            <td>{entry.user_name}</td>
            <td>{entry.client_name || "Без клиента"}</td>
            <td>{minutesLabel(entry.duration_minutes)}</td>
            <td><Badge tone={entry.billable ? "success" : "muted"}>{entry.billable ? "Оплачиваемое" : "Неоплачиваемое"} · {timeSourceLabel(entry.source)}</Badge></td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}
