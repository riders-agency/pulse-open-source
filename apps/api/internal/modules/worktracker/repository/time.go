package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (w *WorkTracker) ActiveTimer(ctx context.Context, userID uuid.UUID) (*TimerSession, error) {
	timers, err := w.ActiveTimers(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(timers) == 0 {
		return nil, ErrNotFound
	}
	return &timers[0], nil
}

func (w *WorkTracker) ActiveTimers(ctx context.Context, userID uuid.UUID) ([]TimerSession, error) {
	rows, err := w.pool().Query(ctx, `
SELECT ts.timer_session_id::text, ts.task_id::text, t.task_key, t.title,
       ts.started_at, ts.state,
       GREATEST(ts.duration_minutes_cached, FLOOR(EXTRACT(EPOCH FROM (now() - ts.started_at)) / 60)::int)
  FROM timer_sessions ts
  JOIN tasks t ON t.task_id = ts.task_id
 WHERE ts.user_id = $1
   AND ts.state IN ('running','paused')
 ORDER BY ts.started_at DESC`, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TimerSession{}
	for rows.Next() {
		var row TimerSession
		if err := rows.Scan(&row.TimerSessionID, &row.TaskID, &row.TaskKey, &row.TaskTitle, &row.StartedAt, &row.State, &row.ElapsedMinutes); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (w *WorkTracker) TimeEntries(ctx context.Context, projectID, userID uuid.UUID) ([]TimeEntry, *Access, error) {
	access, err := w.AccessForProject(ctx, projectID, userID)
	if err != nil {
		return nil, nil, err
	}
	rows, err := w.pool().Query(ctx, timeEntrySelect()+`
 WHERE te.project_id = $1
   AND ($2::text <> 'client_guest')
   AND ($2::text NOT IN ('employee','contractor') OR te.user_id = $3)
   AND ($4::boolean OR te.task_id IS NULL OR EXISTS (
     SELECT 1 FROM tasks t WHERE t.task_id = te.task_id AND t.task_workflow_id::text = ANY($5::text[])
   ))
 ORDER BY te.entry_date DESC, te.created_at DESC
 LIMIT 160`, projectID.String(), access.Role, userID.String(), access.AllWorkflows, access.WorkflowIDs)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	out, err := scanTimeEntries(rows)
	return out, access, err
}

func (w *WorkTracker) TimerSession(ctx context.Context, timerID uuid.UUID) (*TimerSession, error) {
	row := w.pool().QueryRow(ctx, `
SELECT ts.timer_session_id::text, ts.task_id::text, t.task_key, t.title,
       ts.started_at, ts.state,
       GREATEST(ts.duration_minutes_cached, FLOOR(EXTRACT(EPOCH FROM (now() - ts.started_at)) / 60)::int)
  FROM timer_sessions ts
  JOIN tasks t ON t.task_id = ts.task_id
 WHERE ts.timer_session_id = $1`, timerID.String())
	var out TimerSession
	if err := row.Scan(&out.TimerSessionID, &out.TaskID, &out.TaskKey, &out.TaskTitle, &out.StartedAt, &out.State, &out.ElapsedMinutes); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &out, nil
}

func (w *WorkTracker) timeForTask(ctx context.Context, taskID uuid.UUID, access *Access) ([]TimeEntry, error) {
	if access.IsClientGuest {
		return []TimeEntry{}, nil
	}
	rows, err := w.pool().Query(ctx, timeEntrySelect()+`
 WHERE te.task_id = $1
   AND ($2::text NOT IN ('employee','contractor') OR te.user_id = $3)
 ORDER BY te.entry_date DESC, te.created_at DESC
 LIMIT 80`, taskID.String(), access.Role, access.UserID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTimeEntries(rows)
}

func (w *WorkTracker) TimeEntry(ctx context.Context, entryID uuid.UUID) (*TimeEntry, error) {
	row := w.pool().QueryRow(ctx, timeEntrySelect()+` WHERE te.time_entry_id = $1`, entryID.String())
	entry, err := scanTimeEntry(row)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return entry, err
}

func timeEntrySelect() string {
	return `
SELECT te.time_entry_id::text,
       COALESCE(te.task_id::text, ''), COALESCE(t.task_key, ''), COALESCE(t.title, ''),
       COALESCE(te.project_id::text, ''), COALESCE(p.name, ''),
       COALESCE(te.client_id::text, ''), COALESCE(c.name, ''),
       te.user_id::text, COALESCE(u.display_name, u.email::text, ''),
       te.entry_date, te.duration_minutes, COALESCE(te.description, ''),
       te.billable, te.source, te.approved_at
  FROM time_entries te
  LEFT JOIN tasks t ON t.task_id = te.task_id
  LEFT JOIN pulse_read.projects_v1 p ON p.project_id = te.project_id
  LEFT JOIN pulse_read.clients_v1 c ON c.client_id = te.client_id
  LEFT JOIN pulse_read.users_v1 u ON u.user_id = te.user_id`
}

func scanTimeEntries(rows pgx.Rows) ([]TimeEntry, error) {
	out := []TimeEntry{}
	for rows.Next() {
		row, err := scanTimeEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *row)
	}
	return out, rows.Err()
}

func scanTimeEntry(row rowScanner) (*TimeEntry, error) {
	var entry TimeEntry
	var entryDate pgxDate
	var approvedAt pgxTime
	if err := row.Scan(
		&entry.TimeEntryID,
		&entry.TaskID,
		&entry.TaskKey,
		&entry.TaskTitle,
		&entry.ProjectID,
		&entry.ProjectName,
		&entry.ClientID,
		&entry.ClientName,
		&entry.UserID,
		&entry.UserName,
		&entryDate,
		&entry.DurationMinutes,
		&entry.Description,
		&entry.Billable,
		&entry.Source,
		&approvedAt,
	); err != nil {
		return nil, err
	}
	entry.EntryDate = entryDate.Time
	if approvedAt.Valid {
		entry.ApprovedAt = &approvedAt.Time
	}
	return &entry, nil
}

type pgxDate = pgtype.Date
type pgxTime = pgtype.Timestamptz

func nullableTextUUID(value string) any {
	if value == "" {
		return nil
	}
	return value
}
