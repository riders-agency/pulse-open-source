package repository

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (w *WorkTracker) Overview(ctx context.Context, projectID, userID uuid.UUID) (*Overview, *Access, error) {
	access, err := w.AccessForProject(ctx, projectID, userID)
	if err != nil {
		return nil, nil, err
	}
	row := w.pool().QueryRow(ctx, `
WITH visible_tasks AS (
  SELECT t.*
    FROM tasks t
    JOIN task_workflow_statuses s ON s.task_workflow_status_id = t.task_workflow_status_id
   WHERE t.project_id = $1
     AND t.archived_at IS NULL
     AND ($2::text <> 'client_guest' OR (t.client_visible = true AND t.client_id = $3::uuid))
     AND ($2::text <> 'contractor' OR t.assignee_user_id = $4::uuid OR t.reporter_user_id = $4::uuid)
     AND ($5::boolean OR t.task_workflow_id::text = ANY($6::text[]))
)
SELECT
  COUNT(*) FILTER (WHERE due_date = CURRENT_DATE),
  COUNT(*) FILTER (WHERE due_date < CURRENT_DATE AND completed_at IS NULL),
  COUNT(*) FILTER (WHERE assignee_user_id IS NULL),
  COUNT(*) FILTER (WHERE estimate_minutes IS NULL),
  COUNT(*) FILTER (WHERE is_blocked),
  COUNT(*) FILTER (WHERE task_workflow_status_id IN (
    SELECT task_workflow_status_id FROM task_workflow_statuses
     WHERE category = 'review' AND lower(name) NOT LIKE '%client%'
  )),
  COUNT(*) FILTER (WHERE task_workflow_status_id IN (
    SELECT task_workflow_status_id FROM task_workflow_statuses
     WHERE lower(name) LIKE '%client%'
  )),
  COALESCE((SELECT SUM(duration_minutes)::int FROM time_entries WHERE project_id = $1 AND user_id = $4 AND entry_date = CURRENT_DATE), 0),
  COALESCE((SELECT SUM(duration_minutes)::int FROM time_entries WHERE project_id = $1 AND user_id = $4 AND entry_date >= CURRENT_DATE - 6), 0)
FROM visible_tasks`, projectID.String(), access.Role, nullableUUID(access.ClientID), userID.String(), access.AllWorkflows, access.WorkflowIDs)
	var out Overview
	if err := row.Scan(
		&out.TodayTasks,
		&out.OverdueTasks,
		&out.UnassignedTasks,
		&out.UnestimatedTasks,
		&out.BlockedTasks,
		&out.ReviewTasks,
		&out.ClientAcceptance,
		&out.LoggedTodayMinutes,
		&out.LoggedWeekMinutes,
	); err != nil {
		return nil, nil, err
	}
	return &out, access, nil
}

func (w *WorkTracker) ListWorkflows(ctx context.Context, projectID, userID uuid.UUID) ([]Workflow, *Access, error) {
	access, err := w.AccessForProject(ctx, projectID, userID)
	if err != nil {
		return nil, nil, err
	}
	if !access.CanManageProcesses && !access.CanManageMembers && access.Role != "read_only_auditor" {
		return []Workflow{}, access, nil
	}
	rows, err := w.pool().Query(ctx, `
SELECT task_workflow_id::text, name, COALESCE(description, ''), scope, is_default
  FROM task_workflows
 WHERE project_id = $1
   AND ($2::boolean OR task_workflow_id::text = ANY($3::text[]))
 ORDER BY is_default DESC, name`, projectID.String(), access.AllWorkflows, access.WorkflowIDs)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	out := []Workflow{}
	for rows.Next() {
		var row Workflow
		if err := rows.Scan(&row.TaskWorkflowID, &row.Name, &row.Description, &row.Scope, &row.IsDefault); err != nil {
			return nil, nil, err
		}
		statuses, err := w.Statuses(ctx, uuid.MustParse(row.TaskWorkflowID))
		if err != nil {
			return nil, nil, err
		}
		row.Statuses = statuses
		roles, err := processRolesForWorkflow(ctx, w.pool(), uuid.MustParse(row.TaskWorkflowID))
		if err != nil {
			return nil, nil, err
		}
		row.Roles = roles
		out = append(out, row)
	}
	return out, access, rows.Err()
}

func (w *WorkTracker) ListBoards(ctx context.Context, projectID, userID uuid.UUID) ([]Board, *Access, error) {
	access, err := w.AccessForProject(ctx, projectID, userID)
	if err != nil {
		return nil, nil, err
	}
	rows, err := w.pool().Query(ctx, `
SELECT task_board_id::text, name, type, task_workflow_id::text, group_by, swimlane_by, visible_fields_json, card_layout_json
  FROM task_boards
 WHERE project_id = $1
   AND ($2::boolean OR task_workflow_id::text = ANY($3::text[]))
 ORDER BY is_default DESC, name`, projectID.String(), access.AllWorkflows, access.WorkflowIDs)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	out := []Board{}
	for rows.Next() {
		var row Board
		var fields []byte
		var layout []byte
		if err := rows.Scan(&row.TaskBoardID, &row.Name, &row.Type, &row.WorkflowID, &row.GroupBy, &row.SwimlaneBy, &fields, &layout); err != nil {
			return nil, nil, err
		}
		_ = json.Unmarshal(fields, &row.VisibleFields)
		_ = json.Unmarshal(layout, &row.CardLayout)
		row.CanCustomize = access.CanManageProcesses
		out = append(out, row)
	}
	return out, access, rows.Err()
}

func (w *WorkTracker) Board(ctx context.Context, boardID, userID uuid.UUID) (*Board, *Access, error) {
	var projectID uuid.UUID
	if err := w.pool().QueryRow(ctx, `SELECT project_id FROM task_boards WHERE task_board_id = $1`, boardID.String()).Scan(&projectID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}
	access, err := w.AccessForProject(ctx, projectID, userID)
	if err != nil {
		return nil, nil, err
	}
	var board Board
	var workflowID uuid.UUID
	var fields []byte
	var layout []byte
	if err := w.pool().QueryRow(ctx, `
SELECT task_board_id::text, name, type, task_workflow_id, task_workflow_id::text, group_by, swimlane_by, visible_fields_json, card_layout_json
  FROM task_boards
 WHERE task_board_id = $1`, boardID.String()).
		Scan(&board.TaskBoardID, &board.Name, &board.Type, &workflowID, &board.WorkflowID, &board.GroupBy, &board.SwimlaneBy, &fields, &layout); err != nil {
		return nil, nil, err
	}
	_ = json.Unmarshal(fields, &board.VisibleFields)
	_ = json.Unmarshal(layout, &board.CardLayout)
	if !access.CanUseWorkflow(board.WorkflowID) {
		return nil, nil, ErrForbidden
	}
	board.CanCustomize = access.CanManageProcesses
	statuses, err := w.Statuses(ctx, workflowID)
	if err != nil {
		return nil, nil, err
	}
	board.Columns = make([]BoardColumn, 0, len(statuses))
	for _, status := range statuses {
		tasks, err := w.tasksByStatus(ctx, access, uuid.MustParse(status.TaskWorkflowStatusID))
		if err != nil {
			return nil, nil, err
		}
		board.Columns = append(board.Columns, BoardColumn{Status: status, Count: len(tasks), Tasks: tasks})
	}
	board.Summary = map[string]int{"columns": len(board.Columns)}
	return &board, access, nil
}

func (w *WorkTracker) ListTasks(ctx context.Context, in TaskListInput, userID uuid.UUID) (*TaskList, *Access, error) {
	access, err := w.AccessForProject(ctx, in.ProjectID, userID)
	if err != nil {
		return nil, nil, err
	}
	limit := in.Limit
	if limit <= 0 || limit > 200 {
		limit = 80
	}
	offset := in.Offset
	if offset < 0 {
		offset = 0
	}
	where := []string{
		"t.project_id = $1",
		"t.archived_at IS NULL",
		"($2::text <> 'client_guest' OR (t.client_visible = true AND t.client_id = $3::uuid))",
		"($2::text <> 'contractor' OR t.assignee_user_id = $4::uuid OR t.reporter_user_id = $4::uuid)",
	}
	if in.Archived {
		if access.IsClientGuest {
			return nil, nil, ErrForbidden
		}
		where[1] = "t.archived_at IS NOT NULL"
	}
	args := []any{access.ProjectID.String(), access.Role, nullableUUID(access.ClientID), access.UserID.String()}
	next := func(value any) string {
		args = append(args, value)
		return "$" + strconv.Itoa(len(args))
	}
	if !access.AllWorkflows {
		where = append(where, "t.task_workflow_id::text = ANY("+next(access.WorkflowIDs)+"::text[])")
	}
	if in.WorkflowID != nil {
		if !access.CanUseWorkflow(in.WorkflowID.String()) {
			return nil, nil, ErrForbidden
		}
		where = append(where, "t.task_workflow_id = "+next(in.WorkflowID.String())+"::uuid")
	}
	if in.Search = strings.TrimSpace(in.Search); in.Search != "" {
		queryArg := next(in.Search)
		likeArg := next("%" + in.Search + "%")
		where = append(where, `(t.search_vector @@ websearch_to_tsquery('russian'::regconfig, `+queryArg+`) OR t.title ILIKE `+likeArg+` OR t.task_key ILIKE `+likeArg+` OR COALESCE(t.description,'') ILIKE `+likeArg+`)`)
	}
	if in.StatusID != nil {
		where = append(where, "t.task_workflow_status_id = "+next(in.StatusID.String())+"::uuid")
	}
	switch strings.TrimSpace(in.Assignee) {
	case "me":
		where = append(where, "t.assignee_user_id = "+next(access.UserID.String())+"::uuid")
	case "unassigned":
		where = append(where, "t.assignee_user_id IS NULL")
	default:
		if id, err := uuid.Parse(in.Assignee); err == nil && id != uuid.Nil {
			where = append(where, "t.assignee_user_id = "+next(id.String())+"::uuid")
		}
	}
	if in.Type != "" {
		where = append(where, "t.type = "+next(in.Type))
	}
	if in.Priority != "" {
		where = append(where, "t.priority = "+next(in.Priority))
	}
	if in.ClientID != nil {
		where = append(where, "t.client_id = "+next(in.ClientID.String())+"::uuid")
	}
	switch in.Due {
	case "overdue":
		where = append(where, "t.due_date < CURRENT_DATE AND t.completed_at IS NULL")
	case "today":
		where = append(where, "t.due_date = CURRENT_DATE")
	case "week":
		where = append(where, "t.due_date >= CURRENT_DATE AND t.due_date <= CURRENT_DATE + 7")
	case "none":
		where = append(where, "t.due_date IS NULL")
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	countSQL := `SELECT COUNT(*) FROM tasks t JOIN task_workflow_statuses s ON s.task_workflow_status_id = t.task_workflow_status_id WHERE ` + whereSQL
	if err := w.pool().QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, nil, err
	}
	orderSQL := "t.updated_at DESC, t.rank, t.created_at DESC"
	if strings.TrimSpace(in.Search) != "" {
		args = append(args, in.Search)
		orderSQL = `ts_rank_cd(t.search_vector, websearch_to_tsquery('russian'::regconfig, $` + strconv.Itoa(len(args)) + `)) DESC, t.updated_at DESC`
	}
	args = append(args, limit, offset)
	rows, err := w.pool().Query(ctx, taskSelectSQL()+`
 WHERE `+whereSQL+`
 ORDER BY `+orderSQL+`
 LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	items, err := scanTasks(rows)
	if err != nil {
		return nil, nil, err
	}
	return &TaskList{Items: items, Total: total, Limit: limit, Offset: offset}, access, nil
}

func (w *WorkTracker) Task(ctx context.Context, taskID, userID uuid.UUID) (*TaskDetail, *Access, error) {
	access, _, err := w.AccessForTask(ctx, taskID, userID)
	if err != nil {
		return nil, nil, err
	}
	task, err := w.TaskRow(ctx, taskID)
	if err != nil {
		return nil, nil, err
	}
	if !access.CanSeeTask(*task) {
		return nil, nil, ErrForbidden
	}
	comments, err := w.comments(ctx, taskID, access)
	if err != nil {
		return nil, nil, err
	}
	checklist, err := w.checklist(ctx, taskID)
	if err != nil {
		return nil, nil, err
	}
	fields, err := w.customFields(ctx, taskID, *task, access)
	if err != nil {
		return nil, nil, err
	}
	activity, err := w.activity(ctx, taskID)
	if err != nil {
		return nil, nil, err
	}
	timeRows, err := w.timeForTask(ctx, taskID, access)
	if err != nil {
		return nil, nil, err
	}
	attachments, err := w.Attachments(ctx, taskID, userID)
	if err != nil {
		return nil, nil, err
	}
	return &TaskDetail{Task: *task, Comments: comments, Checklist: checklist, CustomFields: fields, Activity: activity, Time: timeRows, Attachments: attachments}, access, nil
}

func (w *WorkTracker) Statuses(ctx context.Context, workflowID uuid.UUID) ([]Status, error) {
	rows, err := w.pool().Query(ctx, `
SELECT task_workflow_status_id::text, name, category, color, position, is_start, is_done,
       wip_limit, requires_assignee, requires_estimate
  FROM task_workflow_statuses
 WHERE task_workflow_id = $1
 ORDER BY position, name`, workflowID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Status{}
	for rows.Next() {
		var row Status
		if err := rows.Scan(
			&row.TaskWorkflowStatusID,
			&row.Name,
			&row.Category,
			&row.Color,
			&row.Position,
			&row.IsStart,
			&row.IsDone,
			&row.WIPLimit,
			&row.RequiresAssignee,
			&row.RequiresEstimate,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (w *WorkTracker) tasksByStatus(ctx context.Context, access *Access, statusID uuid.UUID) ([]Task, error) {
	rows, err := w.pool().Query(ctx, taskSelectSQL()+`
 WHERE t.project_id = $1
   AND t.task_workflow_status_id = $2
   AND t.archived_at IS NULL
   AND ($3::text <> 'client_guest' OR (t.client_visible = true AND t.client_id = $4::uuid))
   AND ($3::text <> 'contractor' OR t.assignee_user_id = $5::uuid OR t.reporter_user_id = $5::uuid)
   AND ($6::boolean OR t.task_workflow_id::text = ANY($7::text[]))
 ORDER BY t.rank, t.created_at
 LIMIT 80`, access.ProjectID.String(), statusID.String(), access.Role, nullableUUID(access.ClientID), access.UserID.String(), access.AllWorkflows, access.WorkflowIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (w *WorkTracker) checklist(ctx context.Context, taskID uuid.UUID) ([]ChecklistItem, error) {
	rows, err := w.pool().Query(ctx, `
SELECT task_checklist_item_id::text, title, is_done, position, completed_at
  FROM task_checklist_items
 WHERE task_id = $1
 ORDER BY position, created_at`, taskID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ChecklistItem{}
	for rows.Next() {
		var item ChecklistItem
		if err := rows.Scan(&item.TaskChecklistItemID, &item.Title, &item.IsDone, &item.Position, &item.CompletedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (w *WorkTracker) customFields(ctx context.Context, taskID uuid.UUID, task Task, access *Access) ([]CustomField, error) {
	rows, err := w.pool().Query(ctx, `
SELECT f.task_custom_field_id::text,
       f.name,
       f.field_key,
       f.type,
       f.required,
       f.options_json,
       v.value_json,
       f.visible_by_default
  FROM task_custom_fields f
  LEFT JOIN task_custom_field_values v
         ON v.task_custom_field_id = f.task_custom_field_id
        AND v.task_id = $1
 WHERE f.company_id = $2
   AND (f.project_id IS NULL OR f.project_id = $3::uuid)
   AND (f.task_workflow_id IS NULL OR f.task_workflow_id = $4::uuid)
   AND (f.applies_to_task_types IS NULL OR cardinality(f.applies_to_task_types) = 0 OR $5 = ANY(f.applies_to_task_types))
 ORDER BY f.position, f.name`, taskID.String(), access.CompanyID.String(), nullableUUIDText(task.ProjectID), nullableUUIDText(task.WorkflowID), task.Type)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CustomField{}
	for rows.Next() {
		var field CustomField
		var optionsRaw []byte
		var valueRaw []byte
		if err := rows.Scan(&field.TaskCustomFieldID, &field.Name, &field.Key, &field.Type, &field.Required, &optionsRaw, &valueRaw, &field.VisibleByDefault); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(optionsRaw, &field.Options)
		if len(valueRaw) > 0 {
			_ = json.Unmarshal(valueRaw, &field.Value)
		}
		out = append(out, field)
	}
	return out, rows.Err()
}

func (w *WorkTracker) TaskRow(ctx context.Context, taskID uuid.UUID) (*Task, error) {
	row := w.pool().QueryRow(ctx, taskSelectSQL()+` WHERE t.task_id = $1`, taskID.String())
	task, err := scanTask(row)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return task, err
}

func taskSelectSQL() string {
	return `
SELECT t.task_id::text, t.task_key, t.title, COALESCE(t.description, ''), t.type,
       t.task_workflow_status_id::text, t.task_workflow_id::text, s.name, s.category, t.priority, t.severity,
       COALESCE(t.assignee_user_id::text, ''), COALESCE(au.display_name, au.email::text, ''),
       COALESCE(t.reporter_user_id::text, ''), COALESCE(t.reviewer_user_id::text, ''),
       COALESCE(t.client_id::text, ''), COALESCE(c.name, ''),
       COALESCE(t.project_id::text, ''), COALESCE(p.name, ''),
       t.start_date, t.due_date, t.completed_at, t.estimate_minutes,
       t.spent_minutes_cached, t.remaining_minutes, t.story_points, t.rank,
       t.is_blocked, COALESCE(t.blocked_reason, ''), t.client_visible, t.billable,
	       t.labels, t.comments_count, t.attachments_count, t.checklist_done_count,
	       t.checklist_total_count, t.revision, t.created_at, t.updated_at
  FROM tasks t
  JOIN task_workflow_statuses s ON s.task_workflow_status_id = t.task_workflow_status_id
  LEFT JOIN pulse_read.users_v1 au ON au.user_id = t.assignee_user_id
  LEFT JOIN pulse_read.clients_v1 c ON c.client_id = t.client_id
  LEFT JOIN pulse_read.projects_v1 p ON p.project_id = t.project_id`
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTasks(rows pgx.Rows) ([]Task, error) {
	out := []Task{}
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *task)
	}
	return out, rows.Err()
}

func scanTask(row rowScanner) (*Task, error) {
	var task Task
	var startDate, dueDate pgtype.Date
	var completedAt pgtype.Timestamptz
	if err := row.Scan(
		&task.TaskID,
		&task.TaskKey,
		&task.Title,
		&task.Description,
		&task.Type,
		&task.StatusID,
		&task.WorkflowID,
		&task.StatusName,
		&task.StatusCategory,
		&task.Priority,
		&task.Severity,
		&task.AssigneeUserID,
		&task.AssigneeName,
		&task.ReporterUserID,
		&task.ReviewerUserID,
		&task.ClientID,
		&task.ClientName,
		&task.ProjectID,
		&task.ProjectName,
		&startDate,
		&dueDate,
		&completedAt,
		&task.EstimateMinutes,
		&task.SpentMinutesCached,
		&task.RemainingMinutes,
		&task.StoryPoints,
		&task.Rank,
		&task.IsBlocked,
		&task.BlockedReason,
		&task.ClientVisible,
		&task.Billable,
		&task.Labels,
		&task.CommentsCount,
		&task.AttachmentsCount,
		&task.ChecklistDoneCount,
		&task.ChecklistTotalCount,
		&task.Revision,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}
	task.StartDate = datePtr(startDate)
	task.DueDate = datePtr(dueDate)
	task.CompletedAt = timestampPtr(completedAt)
	return &task, nil
}

func (w *WorkTracker) comments(ctx context.Context, taskID uuid.UUID, access *Access) ([]TaskComment, error) {
	visibility := "internal"
	if access.IsClientGuest {
		visibility = "client"
	}
	rows, err := w.pool().Query(ctx, `
SELECT tc.task_comment_id::text, COALESCE(u.display_name, u.email::text, 'System'),
       tc.body, tc.visibility, tc.created_at
  FROM task_comments tc
  LEFT JOIN pulse_read.users_v1 u ON u.user_id = tc.author_user_id
 WHERE tc.task_id = $1
   AND ($2::text = 'internal' OR tc.visibility = 'client')
 ORDER BY tc.created_at`, taskID.String(), visibility)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TaskComment{}
	for rows.Next() {
		var row TaskComment
		if err := rows.Scan(&row.TaskCommentID, &row.AuthorName, &row.Body, &row.Visibility, &row.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (w *WorkTracker) activity(ctx context.Context, taskID uuid.UUID) ([]ActivityEvent, error) {
	rows, err := w.pool().Query(ctx, `
SELECT e.event_type, COALESCE(u.display_name, u.email::text, 'System'),
       e.payload_json, e.created_at
  FROM task_activity_events e
  LEFT JOIN pulse_read.users_v1 u ON u.user_id = e.actor_user_id
 WHERE e.task_id = $1
 ORDER BY e.created_at DESC
 LIMIT 40`, taskID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ActivityEvent{}
	for rows.Next() {
		var row ActivityEvent
		if err := rows.Scan(&row.EventType, &row.ActorName, &row.Payload, &row.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func nullableUUID(id *uuid.UUID) any {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return id.String()
}

func nullableUUIDText(value string) any {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return nil
	}
	return id.String()
}

func datePtr(value pgtype.Date) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func timestampPtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}
