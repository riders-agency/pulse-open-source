package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	clientsrepo "riders.agency/pulse/api/internal/modules/clients/contracts"
	workrepo "riders.agency/pulse/api/internal/modules/worktracker/contracts"
)

func (a *AI) executeTaskCreate(ctx context.Context, projectID, userID uuid.UUID, input map[string]any) (map[string]any, error) {
	if a.WorkTracker == nil {
		return nil, fmt.Errorf("worktracker service not configured")
	}
	if stringValue(input["client_query"]) != "" && stringValue(input["client_id"]) == "" {
		return nil, fmt.Errorf("client_id required because client was mentioned")
	}
	statusRaw := stringValue(input["status_id"])
	if statusRaw == "" {
		statusRaw = a.ensureDefaultTaskStatus(ctx, projectID, userID)
	}
	statusID, err := uuid.Parse(statusRaw)
	if err != nil {
		return nil, fmt.Errorf("status_id required")
	}
	var clientPtr *uuid.UUID
	if raw := stringValue(input["client_id"]); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			clientPtr = &id
		}
	}
	var duePtr *time.Time
	if raw := stringValue(input["due_date"]); raw != "" {
		if parsed, err := time.Parse("2006-01-02", raw); err == nil {
			duePtr = &parsed
		}
	}
	task, err := a.WorkTracker.CreateTask(ctx, userID, workrepo.CreateTaskInput{
		ProjectID:     projectID,
		ClientID:      clientPtr,
		StatusID:      statusID,
		Title:         stringValue(input["title"]),
		Description:   stringValue(input["description"]),
		Type:          defaultString(stringValue(input["type"]), "task"),
		Priority:      defaultString(stringValue(input["priority"]), "medium"),
		Severity:      "medium",
		DueDate:       duePtr,
		Labels:        stringSlice(input["labels"]),
		ClientVisible: boolValue(input["client_visible"]),
		Billable:      boolValue(input["billable"]),
	})
	if err != nil {
		return nil, err
	}
	output := map[string]any{
		"ok":       true,
		"task_id":  task.TaskID,
		"task_key": task.TaskKey,
		"title":    task.Title,
		"url":      "/tasks/" + task.TaskID,
		"message":  "Задача создана: " + task.Title,
	}
	if boolValue(input["start_timer_after_create"]) {
		taskUUID, parseErr := uuid.Parse(task.TaskID)
		if parseErr != nil {
			output["timer_started"] = false
			output["timer_error"] = "task_id invalid"
			output["message"] = "Задача создана, но таймер не запущен: " + task.Title
		} else if timer, err := a.WorkTracker.StartTimer(ctx, taskUUID, userID); err != nil {
			output["timer_started"] = false
			output["timer_error"] = err.Error()
			output["message"] = "Задача создана, но таймер не запущен: " + task.Title
		} else {
			output["timer_started"] = true
			output["timer_session_id"] = timer.TimerSessionID
			output["message"] = "Задача создана и таймер запущен: " + task.Title
		}
	}
	return output, nil
}

func (a *AI) executeTaskMove(ctx context.Context, projectID, userID uuid.UUID, input map[string]any) (map[string]any, error) {
	if a.WorkTracker == nil {
		return nil, fmt.Errorf("worktracker service not configured")
	}
	taskID, err := uuid.Parse(stringValue(input["task_id"]))
	if err != nil {
		return nil, fmt.Errorf("task_id required")
	}
	statusID, err := uuid.Parse(stringValue(input["to_status_id"]))
	if err != nil {
		return nil, fmt.Errorf("to_status_id required")
	}
	task, err := a.WorkTracker.MoveTask(ctx, userID, workrepo.MoveTaskInput{
		TaskID:   taskID,
		StatusID: statusID,
		Rank:     floatValue(input["rank"]),
		Comment:  stringValue(input["comment"]),
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":          true,
		"task_id":     task.TaskID,
		"task_key":    task.TaskKey,
		"title":       task.Title,
		"status_id":   task.StatusID,
		"status_name": task.StatusName,
		"url":         "/tasks/" + task.TaskID,
		"message":     "Задача стартована: " + task.TaskKey + " · " + task.Title,
	}, nil
}

func (a *AI) executeClientCreate(ctx context.Context, projectID, userID uuid.UUID, input map[string]any) (map[string]any, error) {
	if a.Clients == nil {
		return nil, fmt.Errorf("clients service not configured")
	}
	name := stringValue(input["name"])
	if name == "" {
		return nil, fmt.Errorf("name required")
	}
	clientID, err := a.Clients.CreateForProject(ctx, projectID, userID, clientsrepo.ClientInput{
		ClientType:    defaultString(stringValue(input["client_type"]), "company"),
		Status:        defaultString(stringValue(input["status"]), "active"),
		Name:          name,
		ShortName:     name,
		FullName:      defaultString(stringValue(input["full_name"]), name),
		Comment:       stringValue(input["comment"]),
		DaDataRawJSON: json.RawMessage("{}"),
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":        true,
		"client_id": clientID,
		"name":      name,
		"url":       "/clients/" + clientID,
		"message":   "Клиент создан: " + name,
	}, nil
}

func (a *AI) rememberToolResult(ctx context.Context, sessionID uuid.UUID, toolKey string, output map[string]any) {
	if a == nil || a.P == nil || sessionID == uuid.Nil {
		return
	}
	switch toolKey {
	case "client.create":
		client := map[string]any{
			"client_id": stringValue(output["client_id"]),
			"name":      stringValue(output["name"]),
			"url":       stringValue(output["url"]),
		}
		if stringValue(client["client_id"]) == "" {
			return
		}
		_, _ = a.P.Exec(ctx, `
UPDATE ai_chat_sessions
   SET working_memory_json = jsonb_set(COALESCE(working_memory_json, '{}'::jsonb), '{last_client}', $2::jsonb, true),
       session_revision = session_revision + 1,
       last_activity_at = now()
 WHERE ai_chat_session_id = $1`, sessionID, mustJSON(client))
	case "task.create", "task.move":
		task := map[string]any{
			"task_id":     stringValue(output["task_id"]),
			"task_key":    stringValue(output["task_key"]),
			"title":       stringValue(output["title"]),
			"status_id":   stringValue(output["status_id"]),
			"status_name": stringValue(output["status_name"]),
			"url":         stringValue(output["url"]),
		}
		if stringValue(task["task_id"]) == "" {
			return
		}
		_, _ = a.P.Exec(ctx, `
UPDATE ai_chat_sessions
   SET working_memory_json = jsonb_set(COALESCE(working_memory_json, '{}'::jsonb), '{last_task}', $2::jsonb, true),
       session_revision = session_revision + 1,
       last_activity_at = now()
 WHERE ai_chat_session_id = $1`, sessionID, mustJSON(task))
	}
}

func (a *AI) defaultTaskStatus(ctx context.Context, projectID, userID uuid.UUID) string {
	if a.WorkTracker == nil {
		return ""
	}
	workflows, _, err := a.WorkTracker.ListWorkflows(ctx, projectID, userID)
	if err == nil {
		for _, wf := range workflows {
			for _, st := range wf.Statuses {
				if st.IsStart {
					return st.TaskWorkflowStatusID
				}
			}
			if len(wf.Statuses) > 0 {
				return wf.Statuses[0].TaskWorkflowStatusID
			}
		}
	}
	var id string
	_ = a.P.QueryRow(ctx, `
SELECT s.task_workflow_status_id::text
  FROM pulse_read.task_workflow_statuses_v1 s
  JOIN pulse_read.task_workflows_v1 tw ON tw.task_workflow_id = s.task_workflow_id
 WHERE tw.project_id = $1
 ORDER BY tw.is_default DESC, s.is_start DESC, s.position
 LIMIT 1`, projectID).Scan(&id)
	return id
}

func (a *AI) ensureDefaultTaskStatus(ctx context.Context, projectID, userID uuid.UUID) string {
	if id := a.defaultTaskStatus(ctx, projectID, userID); id != "" {
		return id
	}
	if a.WorkTracker == nil {
		return ""
	}
	workflow, err := a.WorkTracker.CreateWorkflow(ctx, userID, workrepo.CreateWorkflowInput{
		ProjectID:           projectID,
		Name:                "Основной процесс задач",
		Description:         "Создан автоматически при первом подтвержденном создании задачи через AI сотрудника.",
		StatusNames:         []string{"Backlog", "In progress", "Done"},
		CreateDefaultBoards: true,
	})
	if err != nil || workflow == nil {
		return ""
	}
	for _, status := range workflow.Statuses {
		if status.IsStart {
			return status.TaskWorkflowStatusID
		}
	}
	if len(workflow.Statuses) > 0 {
		return workflow.Statuses[0].TaskWorkflowStatusID
	}
	return a.defaultTaskStatus(ctx, projectID, userID)
}
