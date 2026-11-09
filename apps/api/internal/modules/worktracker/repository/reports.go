package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

func (w *WorkTracker) Employees(ctx context.Context, projectID, userID uuid.UUID) ([]Employee, *Access, error) {
	access, err := w.AccessForProject(ctx, projectID, userID)
	if err != nil {
		return nil, nil, err
	}
	if !access.CanUse("employees") && access.Role != "project_manager" {
		return []Employee{}, access, nil
	}
	rows, err := w.pool().Query(ctx, `
WITH task_stats AS (
  SELECT t.assignee_user_id AS user_id,
         COUNT(*) FILTER (WHERE s.category = 'in_progress')::int AS tasks_in_progress,
         COUNT(*) FILTER (WHERE t.due_date < CURRENT_DATE AND t.completed_at IS NULL)::int AS overdue_tasks
    FROM tasks t
    LEFT JOIN task_workflow_statuses s ON s.task_workflow_status_id = t.task_workflow_status_id
   WHERE t.project_id = $1
     AND t.archived_at IS NULL
     AND ($3::boolean OR t.task_workflow_id::text = ANY($4::text[]))
   GROUP BY t.assignee_user_id
),
time_stats AS (
  SELECT te.user_id,
         COALESCE(SUM(te.duration_minutes) FILTER (WHERE te.entry_date >= CURRENT_DATE - 6), 0)::int AS logged_week,
         COALESCE(SUM(te.duration_minutes) FILTER (WHERE te.entry_date >= CURRENT_DATE - 6 AND te.billable), 0)::int AS billable_week
    FROM time_entries te
   WHERE te.project_id = $1
     AND ($3::boolean OR EXISTS (
       SELECT 1 FROM tasks t WHERE t.task_id = te.task_id AND t.task_workflow_id::text = ANY($4::text[])
     ))
   GROUP BY te.user_id
)
SELECT ep.employee_profile_id::text, ep.user_id::text, ep.full_name, COALESCE(ep.email, u.email::text, ''),
       COALESCE(ep.position, ''), COALESCE(ep.department, ''), ep.role, ep.employment_type,
       ep.status, ep.capacity_hours_per_week, ep.skills,
       COALESCE(tsk.tasks_in_progress, 0)::int AS tasks_in_progress,
       COALESCE(tsk.overdue_tasks, 0)::int AS overdue_tasks,
       COALESCE(tim.logged_week, 0)::int AS logged_week,
       COALESCE(tim.billable_week, 0)::int AS billable_week
  FROM employee_profiles ep
  JOIN pulse_read.users_v1 u ON u.user_id = ep.user_id
  JOIN pulse_read.project_members_v1 pm ON pm.user_id = ep.user_id AND pm.project_id = $1 AND pm.status = 'active'
  LEFT JOIN task_stats tsk ON tsk.user_id = ep.user_id
 LEFT JOIN time_stats tim ON tim.user_id = ep.user_id
 WHERE ep.company_id = $2
 ORDER BY ep.status <> 'active', ep.full_name`, projectID.String(), access.CompanyID.String(), access.AllWorkflows, access.WorkflowIDs)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	out := []Employee{}
	for rows.Next() {
		var row Employee
		if err := rows.Scan(
			&row.EmployeeProfileID,
			&row.UserID,
			&row.FullName,
			&row.Email,
			&row.Position,
			&row.Department,
			&row.Role,
			&row.EmploymentType,
			&row.Status,
			&row.CapacityHoursWeek,
			&row.Skills,
			&row.TasksInProgress,
			&row.OverdueTasks,
			&row.LoggedWeekMinutes,
			&row.BillableWeekMinutes,
		); err != nil {
			return nil, nil, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	if err := w.attachProcessAccesses(ctx, projectID, access, out); err != nil {
		return nil, nil, err
	}
	return out, access, nil
}

func (w *WorkTracker) attachProcessAccesses(ctx context.Context, projectID uuid.UUID, access *Access, employees []Employee) error {
	if len(employees) == 0 {
		return nil
	}
	userIDs := make([]string, 0, len(employees))
	index := map[string]int{}
	for i := range employees {
		userIDs = append(userIDs, employees[i].UserID)
		index[employees[i].UserID] = i
	}
	rows, err := w.pool().Query(ctx, `
SELECT pm.user_id::text,
       pm.task_process_membership_id::text,
       pm.task_workflow_id::text,
       tw.name,
       r.task_process_role_id::text,
       r.role_key,
       r.name,
       r.permissions_json,
       COALESCE(pm.client_id::text, '')
  FROM task_process_memberships pm
  JOIN task_process_roles r ON r.task_process_role_id = pm.task_process_role_id
  JOIN task_workflows tw ON tw.task_workflow_id = pm.task_workflow_id
 WHERE pm.project_id = $1
   AND pm.user_id::text = ANY($2::text[])
   AND pm.status = 'active'
   AND ($3::boolean OR pm.task_workflow_id::text = ANY($4::text[]))
 ORDER BY tw.created_at, r.position`, projectID.String(), userIDs, access.AllWorkflows, access.WorkflowIDs)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		var membership ProcessMembership
		var raw []byte
		if err := rows.Scan(
			&userID,
			&membership.TaskProcessMembershipID,
			&membership.TaskWorkflowID,
			&membership.WorkflowName,
			&membership.TaskProcessRoleID,
			&membership.RoleKey,
			&membership.RoleName,
			&raw,
			&membership.ClientID,
		); err != nil {
			return err
		}
		if idx, ok := index[userID]; ok {
			_ = json.Unmarshal(raw, &membership.Permissions)
			if membership.Permissions == nil {
				membership.Permissions = map[string]any{}
			}
			employees[idx].ProcessAccesses = append(employees[idx].ProcessAccesses, membership)
		}
	}
	return rows.Err()
}

func (w *WorkTracker) Report(ctx context.Context, projectID, userID uuid.UUID) (*Report, *Access, error) {
	access, err := w.AccessForProject(ctx, projectID, userID)
	if err != nil {
		return nil, nil, err
	}
	if !access.CanSeeReports {
		return &Report{}, access, nil
	}
	var out Report
	if err := w.pool().QueryRow(ctx, `
SELECT COUNT(*) FILTER (WHERE created_at >= now() - interval '30 days')::int,
       COUNT(*) FILTER (WHERE completed_at >= now() - interval '30 days')::int,
       COUNT(*) FILTER (WHERE due_date < CURRENT_DATE AND completed_at IS NULL)::int,
       COUNT(*) FILTER (WHERE is_blocked)::int,
       COALESCE((SELECT SUM(duration_minutes)::int FROM time_entries te WHERE te.project_id = $1 AND te.entry_date >= CURRENT_DATE - 29 AND ($2::boolean OR EXISTS (SELECT 1 FROM tasks t WHERE t.task_id = te.task_id AND t.task_workflow_id::text = ANY($3::text[])))), 0),
       COALESCE((SELECT SUM(duration_minutes)::int FROM time_entries te WHERE te.project_id = $1 AND te.entry_date >= CURRENT_DATE - 29 AND te.billable AND ($2::boolean OR EXISTS (SELECT 1 FROM tasks t WHERE t.task_id = te.task_id AND t.task_workflow_id::text = ANY($3::text[])))), 0)
  FROM tasks
 WHERE project_id = $1
   AND archived_at IS NULL
   AND ($2::boolean OR task_workflow_id::text = ANY($3::text[]))`, projectID.String(), access.AllWorkflows, access.WorkflowIDs).Scan(
		&out.CreatedTasks,
		&out.CompletedTasks,
		&out.OverdueTasks,
		&out.BlockedTasks,
		&out.LoggedMinutes,
		&out.BillableMinutes,
	); err != nil {
		return nil, nil, err
	}
	var bucketErr error
	out.ByClient, bucketErr = w.reportBuckets(ctx, "client", projectID, access)
	if bucketErr != nil {
		return nil, nil, bucketErr
	}
	out.ByProject, bucketErr = w.reportBuckets(ctx, "project", projectID, access)
	if bucketErr != nil {
		return nil, nil, bucketErr
	}
	out.ByEmployee, bucketErr = w.reportBuckets(ctx, "employee", projectID, access)
	if bucketErr != nil {
		return nil, nil, bucketErr
	}
	out.ByStatus, bucketErr = w.reportBuckets(ctx, "status", projectID, access)
	if bucketErr != nil {
		return nil, nil, bucketErr
	}
	return &out, access, nil
}

func (w *WorkTracker) reportBuckets(ctx context.Context, kind string, projectID uuid.UUID, access *Access) ([]ReportBucket, error) {
	query := map[string]string{
		"client": `
WITH task_counts AS (
  SELECT client_id, COUNT(*)::int AS tasks
    FROM tasks
   WHERE project_id = $1 AND archived_at IS NULL
     AND ($2::boolean OR task_workflow_id::text = ANY($3::text[]))
   GROUP BY client_id
),
time_sums AS (
  SELECT client_id,
         COALESCE(SUM(duration_minutes), 0)::int AS logged_minutes,
         COALESCE(SUM(duration_minutes) FILTER (WHERE billable), 0)::int AS billable_minutes
    FROM time_entries
   WHERE project_id = $1
     AND ($2::boolean OR EXISTS (
       SELECT 1 FROM tasks t WHERE t.task_id = time_entries.task_id AND t.task_workflow_id::text = ANY($3::text[])
     ))
   GROUP BY client_id
)
SELECT COALESCE(c.client_id::text, ''), COALESCE(c.name, 'Без клиента'),
       tc.tasks,
       COALESCE(ts.logged_minutes, 0),
       COALESCE(ts.billable_minutes, 0)
  FROM task_counts tc
  LEFT JOIN time_sums ts ON ts.client_id IS NOT DISTINCT FROM tc.client_id
  LEFT JOIN pulse_read.clients_v1 c ON c.client_id = tc.client_id
 ORDER BY 4 DESC, 2
 LIMIT 8`,
		"project": `
SELECT p.project_id::text, p.name,
       (SELECT COUNT(*)::int FROM tasks WHERE project_id = p.project_id AND archived_at IS NULL AND ($2::boolean OR task_workflow_id::text = ANY($3::text[]))),
       COALESCE((SELECT SUM(duration_minutes)::int FROM time_entries te WHERE te.project_id = p.project_id AND ($2::boolean OR EXISTS (SELECT 1 FROM tasks t WHERE t.task_id = te.task_id AND t.task_workflow_id::text = ANY($3::text[])))), 0),
       COALESCE((SELECT SUM(duration_minutes)::int FROM time_entries te WHERE te.project_id = p.project_id AND te.billable AND ($2::boolean OR EXISTS (SELECT 1 FROM tasks t WHERE t.task_id = te.task_id AND t.task_workflow_id::text = ANY($3::text[])))), 0)
  FROM pulse_read.projects_v1 p
 WHERE p.project_id = $1`,
		"employee": `
WITH task_counts AS (
  SELECT assignee_user_id AS user_id, COUNT(*)::int AS tasks
    FROM tasks
   WHERE project_id = $1 AND archived_at IS NULL
     AND ($2::boolean OR task_workflow_id::text = ANY($3::text[]))
   GROUP BY assignee_user_id
),
time_sums AS (
  SELECT user_id,
         COALESCE(SUM(duration_minutes), 0)::int AS logged_minutes,
         COALESCE(SUM(duration_minutes) FILTER (WHERE billable), 0)::int AS billable_minutes
    FROM time_entries
   WHERE project_id = $1
     AND ($2::boolean OR EXISTS (
       SELECT 1 FROM tasks t WHERE t.task_id = time_entries.task_id AND t.task_workflow_id::text = ANY($3::text[])
     ))
   GROUP BY user_id
),
users_scope AS (
  SELECT user_id FROM task_counts
  UNION
  SELECT user_id FROM time_sums
)
SELECT COALESCE(u.user_id::text, ''), COALESCE(u.display_name, u.email::text, 'Без ответственного'),
       COALESCE(tc.tasks, 0),
       COALESCE(ts.logged_minutes, 0),
       COALESCE(ts.billable_minutes, 0)
  FROM users_scope us
  LEFT JOIN task_counts tc ON tc.user_id IS NOT DISTINCT FROM us.user_id
  LEFT JOIN time_sums ts ON ts.user_id IS NOT DISTINCT FROM us.user_id
  LEFT JOIN pulse_read.users_v1 u ON u.user_id = us.user_id
 ORDER BY 4 DESC, 2
 LIMIT 10`,
		"status": `
WITH task_counts AS (
  SELECT task_workflow_status_id, COUNT(*)::int AS tasks
    FROM tasks
   WHERE project_id = $1 AND archived_at IS NULL
     AND ($2::boolean OR task_workflow_id::text = ANY($3::text[]))
   GROUP BY task_workflow_status_id
),
time_sums AS (
  SELECT task_workflow_status_id,
         COALESCE(SUM(duration_minutes), 0)::int AS logged_minutes,
         COALESCE(SUM(duration_minutes) FILTER (WHERE billable), 0)::int AS billable_minutes
    FROM time_entries
   WHERE project_id = $1
     AND ($2::boolean OR EXISTS (
       SELECT 1 FROM tasks t WHERE t.task_id = time_entries.task_id AND t.task_workflow_id::text = ANY($3::text[])
     ))
   GROUP BY task_workflow_status_id
)
SELECT s.task_workflow_status_id::text, s.name,
       COALESCE(tc.tasks, 0),
       COALESCE(ts.logged_minutes, 0),
       COALESCE(ts.billable_minutes, 0)
 FROM task_workflow_statuses s
  LEFT JOIN task_counts tc ON tc.task_workflow_status_id = s.task_workflow_status_id
  LEFT JOIN time_sums ts ON ts.task_workflow_status_id = s.task_workflow_status_id
 WHERE ($2::boolean OR s.task_workflow_id::text = ANY($3::text[]))
   AND EXISTS (SELECT 1 FROM task_workflows tw WHERE tw.task_workflow_id = s.task_workflow_id AND tw.project_id = $1)
 ORDER BY s.position`,
	}[kind]
	rows, err := w.pool().Query(ctx, query, projectID.String(), access.AllWorkflows, access.WorkflowIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ReportBucket{}
	for rows.Next() {
		var row ReportBucket
		if err := rows.Scan(&row.ID, &row.Name, &row.Tasks, &row.LoggedMinutes, &row.BillableMinutes); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
