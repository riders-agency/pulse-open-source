package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"riders.agency/pulse/api/internal/modules/runs/contracts"
)

func (r *Runs) EventPage(ctx context.Context, runID, userID uuid.UUID, cursor contracts.EventCursor) (contracts.EventPage, error) {
	page := contracts.EventPage{Next: cursor, Items: []map[string]any{}}
	if err := r.assertRun(ctx, runID, userID); err != nil {
		return page, err
	}
	rows, err := r.P.Query(ctx, `SELECT run_event_id,event_type,message,COALESCE(action_key,''),payload_json,created_at FROM run_events WHERE workflow_run_id=$1 AND (created_at,run_event_id)>($2::timestamptz,$3::uuid) ORDER BY created_at,run_event_id LIMIT 100`, runID, cursor.At, cursor.ID)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		var kind, message, action string
		var payload []byte
		if err := rows.Scan(&page.Next.ID, &kind, &message, &action, &payload, &page.Next.At); err != nil {
			return page, err
		}
		item := map[string]any{"run_event_id": page.Next.ID.String(), "event_type": kind, "message": message, "created_at": page.Next.At}
		if action != "" {
			item["action_key"] = action
		}
		if len(payload) > 0 {
			item["payload_json"] = json.RawMessage(payload)
		}
		page.Items = append(page.Items, item)
	}
	return page, rows.Err()
}
