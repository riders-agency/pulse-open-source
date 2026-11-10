package service

import (
	"context"
	"encoding/json"
	"strings"

	projectcommands "riders.agency/pulse/api/internal/modules/projects/commands"

	"github.com/google/uuid"

	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
)

func (c *Commands) CreateWorkflow(ctx context.Context, userID uuid.UUID, in workrepo.CreateWorkflowInput) (*workrepo.Workflow, error) {
	if c == nil || c.WorkTracker == nil || c.WorkTracker.P == nil {
		return nil, workrepo.ErrNotFound
	}
	access, err := c.WorkTracker.AccessForProject(ctx, in.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if err := access.RequireProcessesAdmin(); err != nil {
		return nil, err
	}
	var tpl workrepo.WorkflowTemplate
	useTemplate := false
	if strings.TrimSpace(in.TemplateKey) != "" {
		if found, ok := workrepo.WorkflowTemplateByKey(strings.TrimSpace(in.TemplateKey)); ok {
			tpl = found
			useTemplate = true
		} else {
			return nil, workrepo.ErrValidation
		}
	}
	names := in.StatusNames
	if len(names) == 0 && !useTemplate {
		names = []string{"Backlog", "Planned", "In progress", "Review", "Done"}
	}
	if useTemplate && strings.TrimSpace(in.Name) == "" {
		in.Name = tpl.Name
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, workrepo.ErrValidation
	}

	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	workflowID := uuid.New()
	if _, err := tx.Exec(ctx, `
INSERT INTO task_workflows (task_workflow_id, company_id, project_id, name, description, scope, created_by_user_id)
VALUES ($1,$2,$3,$4,$5,'project',$6)`, workflowID.String(), access.CompanyID.String(), in.ProjectID.String(),
		strings.TrimSpace(in.Name), defaultText(in.Description, tpl.Description), userID.String()); err != nil {
		return nil, err
	}
	statusNames := names
	templateStatuses := tpl.Statuses
	if useTemplate {
		statusNames = make([]string, len(templateStatuses))
		for i, status := range templateStatuses {
			statusNames[i] = status.Name
		}
	}
	statusIDs := make([]uuid.UUID, 0, len(statusNames))
	statusByName := map[string]uuid.UUID{}
	for i, name := range statusNames {
		statusID := uuid.New()
		statusIDs = append(statusIDs, statusID)
		statusByName[strings.ToLower(strings.TrimSpace(name))] = statusID
		category := statusCategory(i, len(statusNames))
		color := statusColor(category)
		var wip *int
		requiresAssignee := false
		requiresEstimate := false
		autoStart := false
		autoStop := false
		if useTemplate {
			category = templateStatuses[i].Category
			color = templateStatuses[i].Color
			wip = templateStatuses[i].WIPLimit
			requiresAssignee = templateStatuses[i].RequiresAssignee
			requiresEstimate = templateStatuses[i].RequiresEstimate
			autoStart = templateStatuses[i].AutoStartTimer
			autoStop = templateStatuses[i].AutoStopTimer
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO task_workflow_statuses (
  task_workflow_status_id, task_workflow_id, name, category, color, position, is_start, is_done,
  wip_limit, requires_assignee, requires_estimate, auto_start_timer, auto_stop_timer
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			statusID.String(), workflowID.String(), strings.TrimSpace(name), category, color,
			(i+1)*10, i == 0, category == "done" || category == "canceled" || i == len(statusNames)-1,
			wip, requiresAssignee, requiresEstimate, autoStart, autoStop); err != nil {
			return nil, err
		}
	}
	if useTemplate {
		for i, move := range tpl.Transitions {
			from := statusByName[strings.ToLower(strings.TrimSpace(move.From))]
			to := statusByName[strings.ToLower(strings.TrimSpace(move.To))]
			if to == uuid.Nil {
				continue
			}
			if _, err := tx.Exec(ctx, `
INSERT INTO task_workflow_transitions (task_workflow_id, from_status_id, to_status_id, name, allowed_roles, required_fields, automation_rules, position)
VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8)`,
				workflowID.String(), nullableUUIDValue(from), to.String(), defaultText(move.Name, "Move to "+move.To),
				move.AllowedRoles, move.RequiredFields, defaultText(move.AutomationJSON, "{}"), (i+1)*10); err != nil {
				return nil, err
			}
		}
	} else {
		for i := 0; i < len(statusIDs)-1; i++ {
			if _, err := tx.Exec(ctx, `
INSERT INTO task_workflow_transitions (task_workflow_id, from_status_id, to_status_id, name, allowed_roles, position)
VALUES ($1,$2,$3,$4,ARRAY['process_owner','process_manager','executor','reviewer'],$5)`,
				workflowID.String(), statusIDs[i].String(), statusIDs[i+1].String(), "Move forward", (i+1)*10); err != nil {
				return nil, err
			}
		}
	}
	if err := insertDefaultProcessRoles(ctx, tx, access.CompanyID, in.ProjectID, workflowID); err != nil {
		return nil, err
	}
	var ownerRoleID uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT task_process_role_id FROM task_process_roles
 WHERE task_workflow_id = $1 AND role_key = 'process_owner'`, workflowID.String()).Scan(&ownerRoleID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO task_process_memberships (
  company_id, project_id, task_workflow_id, user_id, task_process_role_id,
  invited_by_user_id, invited_at, status
) VALUES ($1,$2,$3,$4,$5,$4,now(),'active')
ON CONFLICT (task_workflow_id, user_id) DO UPDATE SET
  task_process_role_id = EXCLUDED.task_process_role_id,
  status = 'active',
  updated_at = now()`,
		access.CompanyID.String(), in.ProjectID.String(), workflowID.String(), userID.String(), ownerRoleID.String()); err != nil {
		return nil, err
	}
	if useTemplate {
		if err := insertCustomFieldSpecs(ctx, tx, access.CompanyID, in.ProjectID, workflowID, tpl.Fields); err != nil {
			return nil, err
		}
		for i, board := range tpl.Boards {
			if err := insertBoardFromTemplate(ctx, tx, access.CompanyID, in.ProjectID, workflowID, userID, board, i == 0); err != nil {
				return nil, err
			}
		}
	}
	if err := projectcommands.SetInitialTaskWorkflow(ctx, tx, in.ProjectID, workflowID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	statuses, err := c.WorkTracker.Statuses(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	roles, _, err := c.WorkTracker.ListProcessRoles(ctx, workflowID, userID)
	if err != nil {
		roles = []workrepo.ProcessRole{}
	}
	return &workrepo.Workflow{
		TaskWorkflowID: workflowID.String(),
		Name:           strings.TrimSpace(in.Name),
		Description:    strings.TrimSpace(in.Description),
		Scope:          "project",
		Statuses:       statuses,
		Roles:          roles,
	}, nil
}

func (c *Commands) CreateBoard(ctx context.Context, userID uuid.UUID, in workrepo.CreateBoardInput) (*workrepo.Board, error) {
	if c == nil || c.WorkTracker == nil || c.WorkTracker.P == nil {
		return nil, workrepo.ErrNotFound
	}
	access, err := c.WorkTracker.AccessForProject(ctx, in.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if err := access.RequireProcessesAdmin(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.TemplateKey) != "" {
		tpl, ok := workrepo.BoardTemplateByKey(strings.TrimSpace(in.TemplateKey))
		if !ok {
			return nil, workrepo.ErrValidation
		}
		if strings.TrimSpace(in.Name) == "" {
			in.Name = tpl.Name
		}
		in.Type = defaultText(in.Type, tpl.Type)
		in.GroupBy = defaultText(in.GroupBy, tpl.GroupBy)
		in.SwimlaneBy = defaultText(in.SwimlaneBy, tpl.SwimlaneBy)
		if len(in.VisibleFields) == 0 {
			in.VisibleFields = tpl.VisibleFields
		}
		if in.CardLayout == nil {
			in.CardLayout = map[string]any{"density": defaultText(tpl.Density, "normal"), "template": tpl.Key}
		}
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, workrepo.ErrValidation
	}
	workflowID := in.WorkflowID
	if workflowID == uuid.Nil {
		if err := c.WorkTracker.P.QueryRow(ctx, `
SELECT COALESCE(p.task_workflow_id, (
  SELECT task_workflow_id FROM task_workflows WHERE project_id = p.project_id ORDER BY is_default DESC, created_at LIMIT 1
))
  FROM pulse_read.projects_v1 p
 WHERE p.project_id = $1`, in.ProjectID.String()).Scan(&workflowID); err != nil {
			return nil, err
		}
	}
	var exists bool
	if err := c.WorkTracker.P.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1 FROM task_workflows WHERE task_workflow_id = $1 AND project_id = $2
)`, workflowID.String(), in.ProjectID.String()).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, workrepo.ErrValidation
	}
	if len(in.VisibleFields) == 0 {
		in.VisibleFields = []string{"client", "project", "assignee", "estimate", "spent", "due_date", "labels"}
	}
	if in.CardLayout == nil {
		in.CardLayout = map[string]any{"density": "normal"}
	}
	if in.Filters == nil {
		in.Filters = map[string]any{}
	}
	if in.Sort == nil {
		in.Sort = map[string]any{"field": "rank", "direction": "asc"}
	}
	boardID := uuid.New()
	fieldsRaw, _ := json.Marshal(in.VisibleFields)
	layoutRaw, _ := json.Marshal(in.CardLayout)
	filterRaw, _ := json.Marshal(in.Filters)
	sortRaw, _ := json.Marshal(in.Sort)
	tx, err := c.WorkTracker.P.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if in.IsDefault {
		if _, err := tx.Exec(ctx, `UPDATE task_boards SET is_default = false WHERE project_id = $1`, in.ProjectID.String()); err != nil {
			return nil, err
		}
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO task_boards (
  task_board_id, company_id, project_id, name, type, task_workflow_id,
  group_by, swimlane_by, filters_json, sort_json, visible_fields_json,
  card_layout_json, created_by_user_id, is_default
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10::jsonb,$11::jsonb,$12::jsonb,$13,$14)`,
		boardID.String(), access.CompanyID.String(), in.ProjectID.String(), strings.TrimSpace(in.Name),
		defaultText(in.Type, "kanban"), workflowID.String(), defaultText(in.GroupBy, "status"),
		defaultText(in.SwimlaneBy, "none"), string(filterRaw), string(sortRaw), string(fieldsRaw),
		string(layoutRaw), userID.String(), in.IsDefault); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	board, _, err := c.WorkTracker.Board(ctx, boardID, userID)
	return board, err
}
