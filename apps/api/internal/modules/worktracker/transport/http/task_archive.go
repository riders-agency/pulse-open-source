package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"riders.agency/pulse/api/internal/modules/worktracker/service"
)

func archiveTask(wt *service.Commands, archived bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, ok := parseUUID(w, chi.URLParam(r, "task_id"), "task_id")
		if !ok {
			return
		}
		actor, ok := userID(w, r)
		if !ok {
			return
		}
		err := wt.SetTaskArchived(r.Context(), taskID, actor, archived)
		write(w, map[string]any{"archived": archived}, err)
	}
}
