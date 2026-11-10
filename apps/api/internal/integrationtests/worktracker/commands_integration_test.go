package integrationtests

import (
	"context"
	"errors"
	"os"
	"testing"

	commands "riders.agency/pulse/api/internal/modules/worktracker/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
)

func TestCommandsPatchTaskRevisionConflictPostgres(t *testing.T) {
	databaseURL := os.Getenv("PULSE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("PULSE_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(pool.Close)

	ids := seedWorktrackerCommandTestData(t, ctx, pool)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM companies WHERE company_id = $1`, ids.companyID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE user_id = $1`, ids.userID)
	})

	commands := commands.NewCommands(&workrepo.WorkTracker{P: pool})
	title := "Revision one"
	task, err := commands.PatchTask(ctx, ids.userID, workrepo.PatchTaskInput{
		TaskID:           ids.taskID,
		Title:            &title,
		ExpectedRevision: intPtr(0),
	})
	if err != nil {
		t.Fatalf("patch task first revision: %v", err)
	}
	if task.Revision != 1 {
		t.Fatalf("task revision after first patch = %d, want 1", task.Revision)
	}

	staleTitle := "Stale write"
	_, err = commands.PatchTask(ctx, ids.userID, workrepo.PatchTaskInput{
		TaskID:           ids.taskID,
		Title:            &staleTitle,
		ExpectedRevision: intPtr(0),
	})
	if !errors.Is(err, workrepo.ErrConflict) {
		t.Fatalf("stale patch error = %v, want ErrConflict", err)
	}
}

type worktrackerCommandTestIDs struct {
	companyID  uuid.UUID
	userID     uuid.UUID
	projectID  uuid.UUID
	workflowID uuid.UUID
	statusID   uuid.UUID
	roleID     uuid.UUID
	taskID     uuid.UUID
}

func seedWorktrackerCommandTestData(t *testing.T, ctx context.Context, pool *pgxpool.Pool) worktrackerCommandTestIDs {
	t.Helper()
	ids := worktrackerCommandTestIDs{
		companyID:  uuid.New(),
		userID:     uuid.New(),
		projectID:  uuid.New(),
		workflowID: uuid.New(),
		statusID:   uuid.New(),
		roleID:     uuid.New(),
		taskID:     uuid.New(),
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO users (user_id, email, password_hash, first_name, display_name, is_activated)
VALUES ($1, $2, 'test-hash', 'Worktracker Test', 'Worktracker Test', true)`,
		ids.userID, "worktracker-"+ids.userID.String()+"@example.test"); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO companies (company_id, owner_user_id, slug, name, created_by_user_id)
VALUES ($1, $2, $3, 'Worktracker Test', $2)`,
		ids.companyID, ids.userID, "worktracker-"+ids.companyID.String()); err != nil {
		t.Fatalf("insert company: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO company_memberships (company_id, user_id, role, permissions)
VALUES ($1, $2, 'owner', '{}'::jsonb)`,
		ids.companyID, ids.userID); err != nil {
		t.Fatalf("insert company membership: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO projects (project_id, company_id, name, slug, hook_key, hook_secret_hash, created_by_user_id)
VALUES ($1, $2, 'Worktracker Test Project', $3, $4, 'test-secret', $5)`,
		ids.projectID, ids.companyID, "worktracker-"+ids.projectID.String(), "hook-"+ids.projectID.String(), ids.userID); err != nil {
		t.Fatalf("insert project: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO project_members (project_id, user_id, role, permissions, status)
VALUES ($1, $2, 'project_manager', '{"tasks":"admin","time":"approve","reports":"read","processes":"admin"}'::jsonb, 'active')`,
		ids.projectID, ids.userID); err != nil {
		t.Fatalf("insert project member: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO task_workflows (task_workflow_id, company_id, project_id, name, scope, is_default, created_by_user_id)
VALUES ($1, $2, $3, 'Default', 'project', true, $4)`,
		ids.workflowID, ids.companyID, ids.projectID, ids.userID); err != nil {
		t.Fatalf("insert task workflow: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO task_workflow_statuses (task_workflow_status_id, task_workflow_id, name, category, color, position, is_start)
VALUES ($1, $2, 'Backlog', 'backlog', 'slate', 10, true)`,
		ids.statusID, ids.workflowID); err != nil {
		t.Fatalf("insert task status: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO task_process_roles (task_process_role_id, company_id, project_id, task_workflow_id, role_key, name, permissions_json, position, is_system)
VALUES ($1, $2, $3, $4, 'executor', 'Executor', '{"read_tasks":true,"edit_tasks":true,"move_tasks":true,"comment_tasks":true,"log_time":true,"start_timer":true}'::jsonb, 10, true)`,
		ids.roleID, ids.companyID, ids.projectID, ids.workflowID); err != nil {
		t.Fatalf("insert process role: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO task_process_memberships (company_id, project_id, task_workflow_id, user_id, task_process_role_id, status)
VALUES ($1, $2, $3, $4, $5, 'active')`,
		ids.companyID, ids.projectID, ids.workflowID, ids.userID, ids.roleID); err != nil {
		t.Fatalf("insert process membership: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO tasks (
  task_id, company_id, project_id, task_key, title, type,
  task_workflow_status_id, task_workflow_id, priority, severity,
  reporter_user_id, owner_user_id, created_by_user_id
) VALUES ($1,$2,$3,'WT-1','Original','task',$4,$5,'medium','medium',$6,$6,$6)`,
		ids.taskID, ids.companyID, ids.projectID, ids.statusID, ids.workflowID, ids.userID); err != nil {
		t.Fatalf("insert task: %v", err)
	}
	return ids
}

func intPtr(value int) *int {
	return &value
}
