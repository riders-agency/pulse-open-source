package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func defaultText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func textPtr(value *string) any {
	if value == nil {
		return nil
	}
	return strings.TrimSpace(*value)
}

func dateArg(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format("2006-01-02")
}

func nullableUUID(id *uuid.UUID) any {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return id.String()
}

func nullableTextUUID(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func expectedRevision(current int, expected *int) int {
	if expected == nil {
		return current
	}
	return *expected
}

func cleanLabels(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]bool{}
	for _, label := range in {
		label = strings.TrimSpace(label)
		if label == "" || seen[label] {
			continue
		}
		seen[label] = true
		out = append(out, label)
	}
	return out
}

func rankValue(statusID uuid.UUID) float64 {
	if statusID == uuid.Nil {
		return 0
	}
	return float64(time.Now().UnixNano() % 1_000_000_000)
}

func nextTaskKey(ctx context.Context, tx pgx.Tx, projectID uuid.UUID) (string, error) {
	var prefix string
	var next int
	err := tx.QueryRow(ctx, `
SELECT upper(regexp_replace(COALESCE(NULLIF(p.slug, ''), substring(p.name from 1 for 4)), '[^a-zA-Z0-9]+', '', 'g')),
       COALESCE(MAX(NULLIF(regexp_replace(t.task_key, '^[A-Z0-9]+-', ''), '')::int), 0) + 1
  FROM pulse_read.projects_v1 p
  LEFT JOIN tasks t ON t.project_id = p.project_id
 WHERE p.project_id = $1
 GROUP BY p.project_id, p.slug, p.name`, projectID.String()).Scan(&prefix, &next)
	if prefix == "" {
		prefix = "TASK"
	}
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}
	return prefix + "-" + strconv.Itoa(next), err
}

func refreshChecklistCounts(ctx context.Context, tx pgx.Tx, taskID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
UPDATE tasks t
   SET checklist_total_count = COALESCE((SELECT COUNT(*) FROM task_checklist_items i WHERE i.task_id = t.task_id), 0),
       checklist_done_count = COALESCE((SELECT COUNT(*) FROM task_checklist_items i WHERE i.task_id = t.task_id AND i.is_done), 0),
       revision = revision + 1
 WHERE t.task_id = $1`, taskID.String())
	return err
}
