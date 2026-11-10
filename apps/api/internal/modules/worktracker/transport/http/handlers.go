package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/modules/worktracker/repository"
	"riders.agency/pulse/api/internal/modules/worktracker/service"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	"riders.agency/pulse/api/internal/platform/problem"
)

func Mount(r chi.Router, wt *service.Commands) {
	r.Get("/access", access(wt))
	r.Get("/overview", overview(wt))
	r.Get("/presets", presets())
	r.Get("/task-workflows", workflows(wt))
	r.Post("/task-workflows", createWorkflow(wt))
	r.Get("/task-workflows/{task_workflow_id}/roles", processRoles(wt))
	r.Post("/task-workflows/{task_workflow_id}/roles", upsertProcessRole(wt))
	r.Patch("/task-workflows/{task_workflow_id}/roles/{task_process_role_id}", upsertProcessRole(wt))
	r.Post("/task-workflows/{task_workflow_id}/roles/reset", resetProcessRoles(wt))
	r.Get("/boards", boards(wt))
	r.Post("/boards", createBoard(wt))
	r.Get("/boards/{task_board_id}", board(wt))
	r.Get("/tasks", tasks(wt))
	r.Post("/tasks", createTask(wt))
	r.Get("/tasks/{task_id}", task(wt))
	r.Patch("/tasks/{task_id}", patchTask(wt))
	r.Delete("/tasks/{task_id}", archiveTask(wt, true))
	r.Post("/tasks/{task_id}/restore", archiveTask(wt, false))
	r.Post("/tasks/{task_id}/move", moveTask(wt))
	r.Post("/tasks/{task_id}/comments", addComment(wt))
	r.Get("/tasks/{task_id}/attachments", taskAttachments(wt))
	r.Post("/tasks/{task_id}/attachments", uploadTaskAttachment(wt))
	r.Post("/tasks/{task_id}/checklist", addChecklist(wt))
	r.Patch("/tasks/{task_id}/checklist/{task_checklist_item_id}", patchChecklist(wt))
	r.Put("/tasks/{task_id}/custom-fields/{task_custom_field_id}", upsertCustomField(wt))
	r.Post("/tasks/{task_id}/timer/start", startTimer(wt))
	r.Get("/task-attachments/{task_attachment_id}", downloadTaskAttachment(wt))
	r.Get("/timer/active", activeTimer(wt))
	r.Post("/timer/stop", stopTimer(wt))
	r.Get("/time", timeEntries(wt))
	r.Post("/time", addManualTime(wt))
	r.Get("/employees", employees(wt))
	r.Post("/projects/{project_id}/members", inviteMember(wt))
	r.Get("/reports", report(wt))
}

func presets() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		problem.JSON(w, http.StatusOK, map[string]any{
			"workflow_templates": repository.WorkflowTemplates(),
			"board_templates":    repository.WorkflowTemplates()[0].Boards,
		})
	}
}

func access(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, ok := projectIDFromQuery(w, r)
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, err := wt.AccessForProject(r.Context(), projectID, userID)
		write(w, out, err)
	}
}

func overview(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, ok := projectIDFromQuery(w, r)
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, access, err := wt.Overview(r.Context(), projectID, userID)
		write(w, map[string]any{"access": access, "overview": out}, err)
	}
}

func workflows(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, ok := projectIDFromQuery(w, r)
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		rows, access, err := wt.ListWorkflows(r.Context(), projectID, userID)
		write(w, map[string]any{"access": access, "items": rows}, err)
	}
}

func createWorkflow(wt *service.Commands) http.HandlerFunc {
	type body struct {
		ProjectID           string   `json:"project_id"`
		Name                string   `json:"name"`
		Description         string   `json:"description"`
		StatusNames         []string `json:"status_names"`
		TemplateKey         string   `json:"template_key"`
		CreateDefaultBoards bool     `json:"create_default_boards"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var in body
		if !decode(w, r, &in) {
			return
		}
		projectID, ok := parseUUID(w, in.ProjectID, "project_id")
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, err := wt.CreateWorkflow(r.Context(), userID, repository.CreateWorkflowInput{
			ProjectID: projectID, Name: in.Name, Description: in.Description, StatusNames: in.StatusNames,
			TemplateKey: in.TemplateKey, CreateDefaultBoards: in.CreateDefaultBoards,
		})
		writeStatus(w, http.StatusCreated, out, err)
	}
}

func processRoles(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID, ok := parseUUID(w, chi.URLParam(r, "task_workflow_id"), "task_workflow_id")
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		rows, access, err := wt.ListProcessRoles(r.Context(), workflowID, userID)
		write(w, map[string]any{"access": access, "items": rows}, err)
	}
}

func upsertProcessRole(wt *service.Commands) http.HandlerFunc {
	type body struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Permissions map[string]any `json:"permissions"`
		Position    int            `json:"position"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID, ok := parseUUID(w, chi.URLParam(r, "task_workflow_id"), "task_workflow_id")
		if !ok {
			return
		}
		roleID := uuid.Nil
		if raw := chi.URLParam(r, "task_process_role_id"); strings.TrimSpace(raw) != "" {
			parsed, parsedOK := parseUUID(w, raw, "task_process_role_id")
			if !parsedOK {
				return
			}
			roleID = parsed
		}
		var in body
		if !decode(w, r, &in) {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, err := wt.UpsertProcessRole(r.Context(), userID, repository.UpsertProcessRoleInput{
			WorkflowID: workflowID, RoleID: roleID, Name: in.Name,
			Description: in.Description, Permissions: in.Permissions, Position: in.Position,
		})
		status := http.StatusOK
		if roleID == uuid.Nil {
			status = http.StatusCreated
		}
		writeStatus(w, status, out, err)
	}
}

func resetProcessRoles(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID, ok := parseUUID(w, chi.URLParam(r, "task_workflow_id"), "task_workflow_id")
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		rows, err := wt.ResetProcessRoles(r.Context(), workflowID, userID)
		write(w, map[string]any{"items": rows}, err)
	}
}

func boards(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, ok := projectIDFromQuery(w, r)
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		rows, access, err := wt.ListBoards(r.Context(), projectID, userID)
		write(w, map[string]any{"access": access, "items": rows}, err)
	}
}

func createBoard(wt *service.Commands) http.HandlerFunc {
	type body struct {
		ProjectID     string         `json:"project_id"`
		WorkflowID    string         `json:"task_workflow_id"`
		Name          string         `json:"name"`
		Type          string         `json:"type"`
		GroupBy       string         `json:"group_by"`
		SwimlaneBy    string         `json:"swimlane_by"`
		VisibleFields []string       `json:"visible_fields"`
		CardLayout    map[string]any `json:"card_layout"`
		Filters       map[string]any `json:"filters"`
		Sort          map[string]any `json:"sort"`
		TemplateKey   string         `json:"template_key"`
		IsDefault     bool           `json:"is_default"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var in body
		if !decode(w, r, &in) {
			return
		}
		projectID, ok := parseUUID(w, in.ProjectID, "project_id")
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, err := wt.CreateBoard(r.Context(), userID, repository.CreateBoardInput{
			ProjectID: projectID, WorkflowID: uuidOrNil(in.WorkflowID), Name: in.Name, Type: in.Type,
			GroupBy: in.GroupBy, SwimlaneBy: in.SwimlaneBy, VisibleFields: in.VisibleFields,
			CardLayout: in.CardLayout, Filters: in.Filters, Sort: in.Sort, TemplateKey: in.TemplateKey,
			IsDefault: in.IsDefault,
		})
		writeStatus(w, http.StatusCreated, out, err)
	}
}

func board(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID, ok := parseUUID(w, chi.URLParam(r, "task_board_id"), "task_board_id")
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, access, err := wt.Board(r.Context(), boardID, userID)
		write(w, map[string]any{"access": access, "board": out}, err)
	}
}

func tasks(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, ok := projectIDFromQuery(w, r)
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		query := r.URL.Query()
		in := repository.TaskListInput{
			ProjectID:  projectID,
			Search:     query.Get("search"),
			Archived:   query.Get("archived") == "true",
			WorkflowID: optionalUUID(query.Get("task_workflow_id")),
			StatusID:   optionalUUID(query.Get("status_id")),
			Assignee:   query.Get("assignee"),
			Type:       query.Get("type"),
			Priority:   query.Get("priority"),
			ClientID:   optionalUUID(query.Get("client_id")),
			Due:        query.Get("due"),
			Limit:      intQuery(query.Get("limit")),
			Offset:     intQuery(query.Get("offset")),
		}
		out, access, err := wt.ListTasks(r.Context(), in, userID)
		write(w, map[string]any{"access": access, "tasks": out}, err)
	}
}

func createTask(wt *service.Commands) http.HandlerFunc {
	type body struct {
		ProjectID       string   `json:"project_id"`
		ClientID        string   `json:"client_id"`
		StatusID        string   `json:"task_workflow_status_id"`
		Title           string   `json:"title"`
		Description     string   `json:"description"`
		Type            string   `json:"type"`
		Priority        string   `json:"priority"`
		Severity        string   `json:"severity"`
		AssigneeUserID  string   `json:"assignee_user_id"`
		ReviewerUserID  string   `json:"reviewer_user_id"`
		StartDate       string   `json:"start_date"`
		DueDate         string   `json:"due_date"`
		EstimateMinutes *int     `json:"estimate_minutes"`
		StoryPoints     *float64 `json:"story_points"`
		Labels          []string `json:"labels"`
		ClientVisible   bool     `json:"client_visible"`
		Billable        bool     `json:"billable"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var in body
		if !decode(w, r, &in) {
			return
		}
		projectID, ok := parseUUID(w, in.ProjectID, "project_id")
		if !ok {
			return
		}
		statusID, ok := parseUUID(w, in.StatusID, "task_workflow_status_id")
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, err := wt.CreateTask(r.Context(), userID, repository.CreateTaskInput{
			ProjectID: projectID, ClientID: optionalUUID(in.ClientID), StatusID: statusID,
			Title: in.Title, Description: in.Description, Type: in.Type, Priority: in.Priority,
			Severity: in.Severity, AssigneeUserID: optionalUUID(in.AssigneeUserID),
			ReviewerUserID: optionalUUID(in.ReviewerUserID), StartDate: optionalDate(in.StartDate),
			DueDate: optionalDate(in.DueDate), EstimateMinutes: in.EstimateMinutes,
			StoryPoints: in.StoryPoints, Labels: in.Labels, ClientVisible: in.ClientVisible, Billable: in.Billable,
		})
		writeStatus(w, http.StatusCreated, out, err)
	}
}

func task(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, ok := parseUUID(w, chi.URLParam(r, "task_id"), "task_id")
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, access, err := wt.Task(r.Context(), taskID, userID)
		write(w, map[string]any{"access": access, "task": out}, err)
	}
}

func patchTask(wt *service.Commands) http.HandlerFunc {
	type body struct {
		Title            *string   `json:"title"`
		Description      *string   `json:"description"`
		Type             *string   `json:"type"`
		Priority         *string   `json:"priority"`
		Severity         *string   `json:"severity"`
		AssigneeUserID   *string   `json:"assignee_user_id"`
		ReviewerUserID   *string   `json:"reviewer_user_id"`
		ClientID         *string   `json:"client_id"`
		StartDate        *string   `json:"start_date"`
		DueDate          *string   `json:"due_date"`
		EstimateMinutes  *int      `json:"estimate_minutes"`
		StoryPoints      *float64  `json:"story_points"`
		Labels           *[]string `json:"labels"`
		ClientVisible    *bool     `json:"client_visible"`
		Billable         *bool     `json:"billable"`
		IsBlocked        *bool     `json:"is_blocked"`
		BlockedReason    *string   `json:"blocked_reason"`
		ExpectedRevision *int      `json:"expected_revision"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, ok := parseUUID(w, chi.URLParam(r, "task_id"), "task_id")
		if !ok {
			return
		}
		var in body
		if !decode(w, r, &in) {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, err := wt.PatchTask(r.Context(), userID, repository.PatchTaskInput{
			TaskID: taskID, Title: in.Title, Description: in.Description, Priority: in.Priority,
			Type: in.Type, Severity: in.Severity, AssigneeUserID: optionalUUIDPtr(in.AssigneeUserID),
			ReviewerUserID: optionalUUIDPtr(in.ReviewerUserID), ClientID: optionalUUIDPtr(in.ClientID), ClientIDSet: in.ClientID != nil,
			StartDate: optionalDatePtr(in.StartDate),
			DueDate:   optionalDatePtr(in.DueDate), EstimateMinutes: in.EstimateMinutes, StoryPoints: in.StoryPoints,
			Labels: labelsValue(in.Labels), LabelsSet: in.Labels != nil, ClientVisible: in.ClientVisible,
			Billable: in.Billable, IsBlocked: in.IsBlocked, BlockedReason: in.BlockedReason,
			ExpectedRevision: in.ExpectedRevision,
		})
		write(w, out, err)
	}
}

func addComment(wt *service.Commands) http.HandlerFunc {
	type body struct {
		Body       string `json:"body"`
		Visibility string `json:"visibility"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, ok := parseUUID(w, chi.URLParam(r, "task_id"), "task_id")
		if !ok {
			return
		}
		var in body
		if !decode(w, r, &in) {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, err := wt.AddComment(r.Context(), userID, repository.AddCommentInput{TaskID: taskID, Body: in.Body, Visibility: in.Visibility})
		writeStatus(w, http.StatusCreated, map[string]any{"task": out}, err)
	}
}

func addChecklist(wt *service.Commands) http.HandlerFunc {
	type body struct {
		Title string `json:"title"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, ok := parseUUID(w, chi.URLParam(r, "task_id"), "task_id")
		if !ok {
			return
		}
		var in body
		if !decode(w, r, &in) {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, err := wt.AddChecklistItem(r.Context(), userID, repository.AddChecklistInput{TaskID: taskID, Title: in.Title})
		writeStatus(w, http.StatusCreated, map[string]any{"task": out}, err)
	}
}

func patchChecklist(wt *service.Commands) http.HandlerFunc {
	type body struct {
		Title  *string `json:"title"`
		IsDone *bool   `json:"is_done"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, ok := parseUUID(w, chi.URLParam(r, "task_id"), "task_id")
		if !ok {
			return
		}
		itemID, ok := parseUUID(w, chi.URLParam(r, "task_checklist_item_id"), "task_checklist_item_id")
		if !ok {
			return
		}
		var in body
		if !decode(w, r, &in) {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, err := wt.PatchChecklistItem(r.Context(), userID, repository.PatchChecklistInput{TaskID: taskID, ItemID: itemID, Title: in.Title, IsDone: in.IsDone})
		write(w, map[string]any{"task": out}, err)
	}
}

func upsertCustomField(wt *service.Commands) http.HandlerFunc {
	type body struct {
		Value any `json:"value"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, ok := parseUUID(w, chi.URLParam(r, "task_id"), "task_id")
		if !ok {
			return
		}
		fieldID, ok := parseUUID(w, chi.URLParam(r, "task_custom_field_id"), "task_custom_field_id")
		if !ok {
			return
		}
		var in body
		if !decode(w, r, &in) {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, err := wt.UpsertCustomFieldValue(r.Context(), userID, repository.UpsertCustomFieldValueInput{TaskID: taskID, FieldID: fieldID, Value: in.Value})
		write(w, map[string]any{"task": out}, err)
	}
}

func moveTask(wt *service.Commands) http.HandlerFunc {
	type body struct {
		StatusID         string  `json:"task_workflow_status_id"`
		Rank             float64 `json:"rank"`
		Comment          string  `json:"comment"`
		ExpectedRevision *int    `json:"expected_revision"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, ok := parseUUID(w, chi.URLParam(r, "task_id"), "task_id")
		if !ok {
			return
		}
		var in body
		if !decode(w, r, &in) {
			return
		}
		statusID, ok := parseUUID(w, in.StatusID, "task_workflow_status_id")
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, err := wt.MoveTask(r.Context(), userID, repository.MoveTaskInput{TaskID: taskID, StatusID: statusID, Rank: in.Rank, Comment: in.Comment, ExpectedRevision: in.ExpectedRevision})
		write(w, out, err)
	}
}

func startTimer(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, ok := parseUUID(w, chi.URLParam(r, "task_id"), "task_id")
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, err := wt.StartTimer(r.Context(), taskID, userID)
		writeStatus(w, http.StatusCreated, out, err)
	}
}

func activeTimer(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		timers, err := wt.ActiveTimers(r.Context(), userID)
		if err != nil {
			write(w, nil, err)
			return
		}
		var timer *repository.TimerSession
		if len(timers) > 0 {
			timer = &timers[0]
		}
		write(w, map[string]any{"timer": timer, "timers": timers}, nil)
	}
}

func stopTimer(wt *service.Commands) http.HandlerFunc {
	type body struct {
		TaskID string `json:"task_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		var in body
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&in)
		}
		out, err := wt.StopTimer(r.Context(), userID, optionalUUID(in.TaskID))
		write(w, out, err)
	}
}

func timeEntries(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, ok := projectIDFromQuery(w, r)
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		rows, access, err := wt.TimeEntries(r.Context(), projectID, userID)
		write(w, map[string]any{"access": access, "items": rows}, err)
	}
}

func addManualTime(wt *service.Commands) http.HandlerFunc {
	type body struct {
		TaskID          string `json:"task_id"`
		DurationMinutes int    `json:"duration_minutes"`
		Description     string `json:"description"`
		EntryDate       string `json:"entry_date"`
		Billable        bool   `json:"billable"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var in body
		if !decode(w, r, &in) {
			return
		}
		taskID, ok := parseUUID(w, in.TaskID, "task_id")
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		entryDate := time.Now()
		if parsed := optionalDate(in.EntryDate); parsed != nil {
			entryDate = *parsed
		}
		out, err := wt.AddManualTime(r.Context(), userID, repository.ManualTimeInput{
			TaskID: taskID, DurationMinutes: in.DurationMinutes, Description: in.Description,
			EntryDate: entryDate, Billable: in.Billable,
		})
		writeStatus(w, http.StatusCreated, out, err)
	}
}

func employees(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, ok := projectIDFromQuery(w, r)
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		rows, access, err := wt.Employees(r.Context(), projectID, userID)
		write(w, map[string]any{"access": access, "items": rows}, err)
	}
}

func inviteMember(wt *service.Commands) http.HandlerFunc {
	type body struct {
		Email         string `json:"email"`
		DisplayName   string `json:"display_name"`
		Role          string `json:"role"`
		WorkflowID    string `json:"task_workflow_id"`
		ProcessRoleID string `json:"task_process_role_id"`
		ClientID      string `json:"client_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, ok := parseUUID(w, chi.URLParam(r, "project_id"), "project_id")
		if !ok {
			return
		}
		var in body
		if !decode(w, r, &in) {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		workflowID, ok := parseUUID(w, in.WorkflowID, "task_workflow_id")
		if !ok {
			return
		}
		out, err := wt.InviteMember(r.Context(), userID, repository.InviteMemberInput{
			ProjectID: projectID, WorkflowID: workflowID, RoleID: uuidOrNil(in.ProcessRoleID),
			Email: in.Email, DisplayName: in.DisplayName, Role: in.Role, ClientID: optionalUUID(in.ClientID),
		})
		writeStatus(w, http.StatusCreated, out, err)
	}
}

func report(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, ok := projectIDFromQuery(w, r)
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		out, access, err := wt.Report(r.Context(), projectID, userID)
		write(w, map[string]any{"access": access, "report": out}, err)
	}
}

func userID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	claims, ok := mware.SessionClaimsFrom(r.Context())
	if !ok {
		problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
		return uuid.Nil, false
	}
	return claims.UserID, true
}

func projectIDFromQuery(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	return parseUUID(w, r.URL.Query().Get("project_id"), "project_id")
}

func parseUUID(w http.ResponseWriter, raw string, field string) (uuid.UUID, bool) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil || id == uuid.Nil {
		problem.WriteModule(w, http.StatusBadRequest, "work", "invalid_"+field, "Некорректный идентификатор", "Проверьте "+field+".", "Выберите рабочий контекст и повторите действие.")
		return uuid.Nil, false
	}
	return id, true
}

func optionalUUID(raw string) *uuid.UUID {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return nil
	}
	return &id
}

func uuidOrNil(raw string) uuid.UUID {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil
	}
	return id
}

func optionalUUIDPtr(raw *string) *uuid.UUID {
	if raw == nil {
		return nil
	}
	return optionalUUID(*raw)
}

func optionalDate(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil
	}
	return &t
}

func optionalDatePtr(raw *string) *time.Time {
	if raw == nil {
		return nil
	}
	return optionalDate(*raw)
}

func intQuery(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	value := 0
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return 0
		}
		value = value*10 + int(ch-'0')
	}
	return value
}

func labelsValue(value *[]string) []string {
	if value == nil {
		return nil
	}
	return *value
}

func decode(w http.ResponseWriter, r *http.Request, out any) bool {
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		problem.Write(w, http.StatusBadRequest, "invalid json", "")
		return false
	}
	return true
}

func write(w http.ResponseWriter, value any, err error) {
	writeStatus(w, http.StatusOK, value, err)
}

func writeStatus(w http.ResponseWriter, status int, value any, err error) {
	if err != nil {
		writeError(w, err)
		return
	}
	problem.JSON(w, status, value)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrForbidden):
		problem.WriteModule(w, http.StatusForbidden, "work", "forbidden", "Нет доступа", "Эта роль не может выполнить действие.", "Попросите владельца проекта изменить роль или открыть нужный модуль.")
	case errors.Is(err, repository.ErrTransitionDenied):
		problem.WriteModule(w, http.StatusForbidden, "work", "transition_denied", "Переход запрещен", "Процесс не разрешает этот переход для вашей роли.", "Выберите разрешенный статус или обратитесь к руководителю проекта.")
	case errors.Is(err, repository.ErrValidation):
		problem.WriteModule(w, http.StatusBadRequest, "work", "validation_failed", "Не хватает данных", "Для действия нужны обязательные поля.", "Заполните ответственного, оценку или комментарий и повторите.")
	case errors.Is(err, repository.ErrSecondTimer):
		problem.WriteModule(w, http.StatusConflict, "work", "active_timer_exists", "Таймер уже запущен", "У пользователя может быть только один активный таймер.", "Остановите текущий таймер и запустите новый.")
	case errors.Is(err, repository.ErrTaskTimerActive):
		problem.WriteModule(w, http.StatusConflict, "work", "task_timer_active", "Сначала остановите таймер", "У задачи есть работающий или приостановленный таймер.", "Остановите таймер в карточке задачи и повторите удаление.")
	case errors.Is(err, repository.ErrConflict):
		problem.WriteModule(w, http.StatusConflict, "work", "revision_conflict", "Задача уже изменена", "Данные задачи были обновлены другим действием.", "Обновите карточку и повторите изменение.")
	case errors.Is(err, repository.ErrNotFound):
		problem.WriteModule(w, http.StatusNotFound, "work", "not_found", "Не найдено", "Ресурс не найден в выбранном проекте.", "Обновите список и повторите действие.")
	default:
		slog.Error("work tracker internal error", slog.String("error", err.Error()))
		problem.WriteModule(w, http.StatusInternalServerError, "work", "internal_error", "Ошибка раздела задач", "", "Повторите позже или сообщите код обращения администратору.")
	}
}
