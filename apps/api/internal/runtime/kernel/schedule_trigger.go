package kernel

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/runtime/contracts"
	"riders.agency/pulse/api/internal/runtime/runqueue"
)

type triggerNotifier interface {
	Notify()
}

type ScheduleTrigger struct {
	P        *pgxpool.Pool
	Notifier triggerNotifier
}

type scheduleActionPlan struct {
	ActionKey     string               `json:"action_key"`
	ActionTypeKey string               `json:"action_type_key"`
	Config        scheduleActionConfig `json:"config"`
}

type scheduleActionConfig struct {
	Enabled         *bool  `json:"enabled"`
	ActiveFrom      string `json:"active_from"`
	ActiveUntil     string `json:"active_until"`
	ScheduleKey     string `json:"schedule_key"`
	IntervalSeconds int    `json:"interval_seconds"`
}

type scheduleDefinition struct {
	Nodes []scheduleActionPlan `json:"nodes"`
}

func (h *ScheduleTrigger) TypeKey() string { return "schedule.trigger" }

func (h *ScheduleTrigger) Sync(ctx context.Context) error {
	h.syncScheduleStates(ctx)
	return nil
}

func (h *ScheduleTrigger) ClaimDue(context.Context, int) ([]contracts.TriggerEvent, error) {
	return nil, nil
}

func (h *ScheduleTrigger) MarkEnqueued(context.Context, contracts.TriggerEvent, string) error {
	return nil
}

func (h *ScheduleTrigger) EnqueueDue(ctx context.Context, limit int) (int, error) {
	if h == nil || h.P == nil {
		return 0, nil
	}
	count := 0
	for limit <= 0 || count < limit {
		if !h.enqueueOneDueSchedule(ctx) {
			break
		}
		count++
	}
	return count, nil
}

func (h *ScheduleTrigger) notify() {
	if h != nil && h.Notifier != nil {
		h.Notifier.Notify()
	}
}

func (h *ScheduleTrigger) syncScheduleStates(ctx context.Context) {
	if h == nil || h.P == nil {
		return
	}
	syncStarted := time.Now().UTC()
	rows, err := h.P.Query(ctx, `
SELECT w.project_id::text, w.workflow_id::text, w.current_workflow_version_id::text, wv.definition_json
  FROM workflows w
  JOIN workflow_versions wv ON wv.workflow_version_id = w.current_workflow_version_id
 WHERE w.current_workflow_version_id IS NOT NULL
   AND w.status = 'active'`)
	if err != nil {
		slog.Warn("schedule sync query failed", slog.Any("err", err))
		return
	}
	defer rows.Close()
	for rows.Next() {
		var projectID, workflowID, versionID string
		var def []byte
		if err := rows.Scan(&projectID, &workflowID, &versionID, &def); err != nil {
			slog.Warn("schedule sync scan failed", slog.Any("err", err))
			continue
		}
		for _, node := range flattenScheduleNodes(def) {
			if node.ActionTypeKey != "schedule.trigger" {
				continue
			}
			scheduleKey := scheduleKeyForNode(node)
			intervalSeconds := clampScheduleInterval(node.Config.IntervalSeconds)
			enabled := scheduleEnabled(node.Config, time.Now().UTC())
			_, err := h.P.Exec(ctx, `
INSERT INTO workflow_schedule_state (
  workflow_id, action_key, workflow_version_id, schedule_key,
  enabled, interval_seconds, next_run_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,clock_timestamp() + ($6::integer * interval '1 second'), clock_timestamp())
ON CONFLICT (workflow_id, action_key) DO UPDATE SET
  workflow_version_id = EXCLUDED.workflow_version_id,
  schedule_key = EXCLUDED.schedule_key,
  enabled = EXCLUDED.enabled,
  interval_seconds = EXCLUDED.interval_seconds,
  next_run_at = CASE
    WHEN workflow_schedule_state.workflow_version_id IS DISTINCT FROM EXCLUDED.workflow_version_id
      OR workflow_schedule_state.interval_seconds IS DISTINCT FROM EXCLUDED.interval_seconds
      OR workflow_schedule_state.enabled = false
    THEN LEAST(workflow_schedule_state.next_run_at, clock_timestamp() + (EXCLUDED.interval_seconds::integer * interval '1 second'))
    ELSE workflow_schedule_state.next_run_at
  END,
  updated_at = clock_timestamp()`,
				workflowID, node.ActionKey, versionID, scheduleKey, enabled, intervalSeconds)
			if err != nil {
				slog.Warn("schedule state upsert failed",
					slog.String("workflow_id", workflowID),
					slog.String("action_key", node.ActionKey),
					slog.Any("err", err),
				)
			}
		}
		_ = projectID
	}
	if err := rows.Err(); err != nil {
		slog.Warn("schedule sync rows failed", slog.Any("err", err))
	}
	_, _ = h.P.Exec(ctx, `
UPDATE workflow_schedule_state
   SET enabled = false, updated_at = clock_timestamp()
 WHERE updated_at < $1`, syncStarted)
}

func (h *ScheduleTrigger) enqueueOneDueSchedule(ctx context.Context) bool {
	tx, err := h.P.Begin(ctx)
	if err != nil {
		return false
	}
	defer tx.Rollback(ctx)
	row := tx.QueryRow(ctx, `
SELECT s.workflow_id::text, s.workflow_version_id::text, s.action_key::text, s.schedule_key::text,
       s.interval_seconds, w.project_id::text
  FROM workflow_schedule_state s
  JOIN workflows w ON w.workflow_id = s.workflow_id
 WHERE s.enabled = true
   AND s.next_run_at <= clock_timestamp()
   AND w.status = 'active'
 ORDER BY s.next_run_at
 FOR UPDATE OF s SKIP LOCKED
 LIMIT 1`)
	var workflowID, versionID, actionKey, scheduleKey, projectID string
	var intervalSeconds int
	if err := row.Scan(&workflowID, &versionID, &actionKey, &scheduleKey, &intervalSeconds, &projectID); err != nil {
		return false
	}
	runID := uuid.New()
	now := time.Now().UTC()
	inputJSON, _ := json.Marshal(scheduleRunInput(scheduleKey, actionKey, versionID, now))
	if _, _, err := runqueue.RequestRunTx(ctx, tx, runqueue.Request{
		WorkflowRunID:     runID,
		ProjectID:         projectID,
		WorkflowID:        workflowID,
		WorkflowVersionID: versionID,
		SourceType:        "schedule",
		InputJSON:         json.RawMessage(inputJSON),
		EventActionKey:    actionKey,
		EventMessage:      "scheduled run queued",
	}); err != nil {
		slog.Warn("schedule run insert failed", slog.String("workflow_id", workflowID), slog.Any("err", err))
		return false
	}
	if _, err := tx.Exec(ctx, `
UPDATE workflow_schedule_state
   SET last_run_at = clock_timestamp(),
       last_workflow_run_id = $3,
       next_run_at = clock_timestamp() + ($4::integer * interval '1 second'),
       updated_at = clock_timestamp()
 WHERE workflow_id = $1 AND action_key = $2`,
		workflowID, actionKey, runID.String(), clampScheduleInterval(intervalSeconds)); err != nil {
		slog.Warn("schedule state update failed", slog.String("workflow_id", workflowID), slog.Any("err", err))
		return false
	}
	if err := tx.Commit(ctx); err != nil {
		return false
	}
	h.notify()
	return true
}

func scheduleRunInput(scheduleKey, actionKey, versionID string, scheduledAt time.Time) map[string]any {
	schedule := map[string]any{
		"key":          scheduleKey,
		"action_key":   actionKey,
		"scheduled_at": scheduledAt.Format(time.RFC3339),
	}
	return map[string]any{
		"_pulse": map[string]any{
			"source_type":         "schedule",
			"schedule_key":        scheduleKey,
			"schedule_action_key": actionKey,
			"workflow_version_id": versionID,
			"scheduled_at":        scheduledAt.Format(time.RFC3339),
		},
		"schedule":  schedule,
		scheduleKey: schedule,
	}
}

func scheduleEnabled(cfg scheduleActionConfig, now time.Time) bool {
	if cfg.Enabled != nil && !*cfg.Enabled {
		return false
	}
	return activePeriodIncludes(cfg.ActiveFrom, cfg.ActiveUntil, now)
}

func activePeriodIncludes(fromRaw, untilRaw string, now time.Time) bool {
	now = now.UTC()
	if from, ok := parseActionBoundary(fromRaw, false); ok && now.Before(from) {
		return false
	}
	if until, ok := parseActionBoundary(untilRaw, true); ok && now.After(until) {
		return false
	}
	return true
}

func parseActionBoundary(raw string, endOfDay bool) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed.UTC(), true
	}
	if parsed, err := time.Parse("2006-01-02", raw); err == nil {
		parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
		if endOfDay {
			parsed = parsed.Add(24*time.Hour - time.Nanosecond)
		}
		return parsed, true
	}
	return time.Time{}, false
}

func scheduleKeyForNode(node scheduleActionPlan) string {
	key := strings.TrimSpace(node.Config.ScheduleKey)
	if key == "" {
		key = strings.TrimSpace(node.ActionKey)
	}
	if key == "" {
		key = "schedule"
	}
	return key
}

func flattenScheduleNodes(def []byte) []scheduleActionPlan {
	var d scheduleDefinition
	_ = json.Unmarshal(def, &d)
	out := make([]scheduleActionPlan, 0, len(d.Nodes))
	for i := range d.Nodes {
		node := &d.Nodes[i]
		node.ActionKey = strings.TrimSpace(node.ActionKey)
		node.ActionTypeKey = strings.TrimSpace(node.ActionTypeKey)
		if node.ActionKey == "" {
			node.ActionKey = node.ActionTypeKey
		}
		if node.ActionTypeKey == "" {
			node.ActionTypeKey = "unknown"
		}
		if node.ActionKey != "" {
			out = append(out, *node)
		}
	}
	return out
}

func clampScheduleInterval(value int) int {
	if value < 10 {
		return 10
	}
	if value > 86400*30 {
		return 86400 * 30
	}
	return value
}
