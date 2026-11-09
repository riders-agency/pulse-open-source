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

func TestTaskArchiveRestorePostgres(t *testing.T) {
	dsn := os.Getenv("PULSE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PULSE_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	ids := seedWorktrackerCommandTestData(t, ctx, pool)
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM companies WHERE company_id=$1`, ids.companyID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE user_id=$1`, ids.userID)
	})
	repo := &workrepo.WorkTracker{P: pool}
	commands := commands.NewCommands(repo)
	if err := commands.SetTaskArchived(ctx, ids.taskID, uuid.New(), true); !errors.Is(err, workrepo.ErrForbidden) {
		t.Fatalf("foreign user: %v", err)
	}
	if _, err := commands.StartTimer(ctx, ids.taskID, ids.userID); err != nil {
		t.Fatal(err)
	}
	if err := commands.SetTaskArchived(ctx, ids.taskID, ids.userID, true); !errors.Is(err, workrepo.ErrTaskTimerActive) {
		t.Fatalf("running timer: %v", err)
	}
	if _, err := commands.StopTimer(ctx, ids.userID, &ids.taskID); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := commands.SetTaskArchived(ctx, ids.taskID, ids.userID, true); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := repo.Task(ctx, ids.taskID, ids.userID); !errors.Is(err, workrepo.ErrNotFound) {
		t.Fatalf("archived detail: %v", err)
	}
	if _, err := commands.StartTimer(ctx, ids.taskID, ids.userID); !errors.Is(err, workrepo.ErrNotFound) {
		t.Fatalf("archived timer: %v", err)
	}
	active, _, err := repo.ListTasks(ctx, workrepo.TaskListInput{ProjectID: ids.projectID}, ids.userID)
	if err != nil || active.Total != 0 {
		t.Fatalf("active: %+v %v", active, err)
	}
	archived, _, err := repo.ListTasks(ctx, workrepo.TaskListInput{ProjectID: ids.projectID, Archived: true}, ids.userID)
	if err != nil || archived.Total != 1 {
		t.Fatalf("archived: %+v %v", archived, err)
	}
	for i := 0; i < 2; i++ {
		if err := commands.SetTaskArchived(ctx, ids.taskID, ids.userID, false); err != nil {
			t.Fatal(err)
		}
	}
	task, _, err := repo.Task(ctx, ids.taskID, ids.userID)
	if err != nil || task.SpentMinutesCached < 1 {
		t.Fatalf("history lost: %+v %v", task, err)
	}
	var events int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM task_activity_events WHERE task_id=$1 AND event_type IN ('task.archived','task.restored')`, ids.taskID).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if events != 2 {
		t.Fatalf("idempotent audit events: %d", events)
	}
}
