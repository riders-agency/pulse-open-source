package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"riders.agency/pulse/api/internal/modules/runs/contracts"
	"riders.agency/pulse/api/internal/modules/runs/service"
)

func streamRunEvents(runs *service.Commands, w http.ResponseWriter, r *http.Request, runID, userID uuid.UUID) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("stream unsupported")
	}
	page, err := runs.EventPage(r.Context(), runID, userID, contracts.EventCursor{})
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	ticker := time.NewTicker(700 * time.Millisecond)
	defer ticker.Stop()
	for {
		for _, event := range page.Items {
			data, err := json.Marshal(event)
			if err != nil {
				return err
			}
			if _, err = fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				return err
			}
		}
		flusher.Flush()
		if len(page.Items) < 100 {
			select {
			case <-r.Context().Done():
				return nil
			case <-ticker.C:
			}
		}
		page, err = runs.EventPage(r.Context(), runID, userID, page.Next)
		if err != nil {
			return err
		}
	}
}
