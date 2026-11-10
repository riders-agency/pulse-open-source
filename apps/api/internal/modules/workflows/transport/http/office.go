package http

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
	"riders.agency/pulse/api/internal/modules/workflows/office"
	"riders.agency/pulse/api/internal/platform/errs"
	mid "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	"riders.agency/pulse/api/internal/platform/problem"
)

func MountOfficeRoutes(r chi.Router, svc *office.Service) {
	r.Get("/office-demo", officeCommand(svc, "state"))
	r.Post("/office-demo", officeCommand(svc, "ensure"))
	r.Patch("/office-demo", officeCommand(svc, "toggle"))
	r.Post("/office-demo/messages", officeCommand(svc, "deliver"))
	r.Post("/office-demo/decisions/{workflow_run_id}", officeCommand(svc, "decide"))
}

type officeInput struct {
	DeliveryKey uuid.UUID `json:"delivery_key"`
	Enabled     *bool     `json:"enabled"`
	Approved    *bool     `json:"approved"`
}

func officeCommand(svc *office.Service, command string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mid.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, 401, "Требуется вход", "")
			return
		}
		pid, err := uuid.Parse(chi.URLParam(r, "project_id"))
		if err != nil {
			problem.Write(w, 400, "Некорректный проект", "")
			return
		}
		var in officeInput
		if command == "toggle" || command == "deliver" || command == "decide" {
			if json.NewDecoder(r.Body).Decode(&in) != nil {
				problem.Write(w, 400, "Некорректные данные", "")
				return
			}
		}
		var result map[string]any
		switch command {
		case "state":
			result, err = svc.State(r.Context(), pid, claims.UserID)
		case "ensure":
			result, err = svc.Ensure(r.Context(), pid, claims.UserID)
		case "deliver":
			result, err = svc.Deliver(r.Context(), pid, claims.UserID, in.DeliveryKey)
		case "toggle":
			if in.Enabled == nil {
				problem.Write(w, 400, "Укажите enabled", "")
				return
			}
			result, err = svc.SetEnabled(r.Context(), pid, claims.UserID, *in.Enabled)
		case "decide":
			rid, parseErr := uuid.Parse(chi.URLParam(r, "workflow_run_id"))
			if parseErr != nil || in.Approved == nil {
				problem.Write(w, 400, "Укажите решение и запуск", "")
				return
			}
			result, err = svc.Decide(r.Context(), pid, claims.UserID, rid, *in.Approved)
		}
		if errors.Is(err, errs.ErrForbidden) {
			problem.Write(w, 403, "Доступ запрещён", "")
			return
		}
		if errors.Is(err, office.ErrConflict) {
			problem.Write(w, 409, "Действие уже выполнено или очередь заполнена", "Обновите данные сценария.")
			return
		}
		if err != nil {
			problem.Write(w, 500, "Не удалось выполнить действие", "Повторите попытку.")
			return
		}
		problem.JSON(w, 200, result)
	}
}
