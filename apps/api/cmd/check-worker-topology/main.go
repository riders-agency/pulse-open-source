package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/platform/app"
	"riders.agency/pulse/api/internal/platform/config"
	"riders.agency/pulse/api/internal/runtime/runqueue"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		fail("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fail("connect: %v", err)
	}
	defer pool.Close()

	fixture, err := createFixture(ctx, pool)
	if err != nil {
		fail("fixture: %v", err)
	}
	defer cleanupFixture(context.Background(), pool, fixture.CompanyID)

	runID := uuid.New()
	input, _ := json.Marshal(map[string]any{"topology": "split-worker"})
	if _, _, err := runqueue.RequestRun(ctx, pool, runqueue.Request{
		WorkflowRunID:     runID,
		ProjectID:         fixture.ProjectID.String(),
		WorkflowID:        fixture.WorkflowID.String(),
		WorkflowVersionID: fixture.VersionID.String(),
		SourceType:        runqueue.SourceTest,
		IdempotencyKey:    "worker-topology:" + runID.String(),
		InputJSON:         input,
		EventMessage:      "worker topology smoke queued",
	}); err != nil {
		fail("request run: %v", err)
	}

	apiOne := buildApplication(ctx, dsn, "topology-api-1")
	defer apiOne.Close()
	apiTwo := buildApplication(ctx, dsn, "topology-api-2")
	defer apiTwo.Close()
	time.Sleep(300 * time.Millisecond)
	status := workflowStatus(ctx, pool, runID)
	if status != "pending" {
		fail("split-worker API executed embedded runtime: status=%s", status)
	}

	workerOne := buildApplication(ctx, dsn, "topology-worker-1")
	defer workerOne.Close()
	workerTwo := buildApplication(ctx, dsn, "topology-worker-2")
	defer workerTwo.Close()

	var wg sync.WaitGroup
	results := make(chan bool, 2)
	for _, application := range []*app.Application{workerOne, workerTwo} {
		wg.Add(1)
		go func(application *app.Application) {
			defer wg.Done()
			if application.RuntimeKernel == nil {
				results <- false
				return
			}
			results <- application.RuntimeKernel.ProcessOne(ctx)
		}(application)
	}
	wg.Wait()
	close(results)
	processed := 0
	for ok := range results {
		if ok {
			processed++
		}
	}
	if processed != 1 {
		fail("expected exactly one worker claim, got %d", processed)
	}

	var actionAttempts, claimCount int
	if err := pool.QueryRow(ctx, `
	SELECT wr.run_claim_count,
	       COALESCE((SELECT SUM(ar.attempt_count)::int FROM action_runs ar WHERE ar.workflow_run_id = wr.workflow_run_id), 0)
	  FROM workflow_runs wr
	 WHERE wr.workflow_run_id = $1`, runID).Scan(&claimCount, &actionAttempts); err != nil {
		fail("verify run: %v", err)
	}
	status = workflowStatus(ctx, pool, runID)
	if status != "succeeded" || claimCount != 1 || actionAttempts != 1 {
		fail("unexpected worker result status=%s claim_count=%d action_attempts=%d", status, claimCount, actionAttempts)
	}
}

func buildApplication(ctx context.Context, dsn, serviceID string) *app.Application {
	cfg := config.Config{
		DatabaseURL:           dsn,
		Environment:           "development",
		RuntimeTopology:       "split-worker",
		SkipMigrations:        true,
		HTTPReadHeaderTimeout: 10 * time.Second,
		HTTPShutdownTimeout:   10 * time.Second,
	}
	_ = os.Setenv("PULSE_INTERNAL_SERVICE_ID", serviceID)
	application, err := app.Build(ctx, cfg)
	if err != nil {
		fail("build %s: %v", serviceID, err)
	}
	return application
}

func workflowStatus(ctx context.Context, pool *pgxpool.Pool, runID uuid.UUID) string {
	var status string
	if err := pool.QueryRow(ctx, `SELECT status FROM workflow_runs WHERE workflow_run_id = $1`, runID).Scan(&status); err != nil {
		fail("workflow status: %v", err)
	}
	return status
}

type fixtureIDs struct {
	UserID     uuid.UUID
	CompanyID  uuid.UUID
	ProjectID  uuid.UUID
	WorkflowID uuid.UUID
	VersionID  uuid.UUID
}

func createFixture(ctx context.Context, pool *pgxpool.Pool) (fixtureIDs, error) {
	f := fixtureIDs{
		UserID:     uuid.New(),
		CompanyID:  uuid.New(),
		ProjectID:  uuid.New(),
		WorkflowID: uuid.New(),
		VersionID:  uuid.New(),
	}
	if _, err := pool.Exec(ctx, `
	INSERT INTO users (user_id, first_name, display_name, is_activated)
	VALUES ($1, 'Worker Topology Smoke', 'Worker Topology Smoke', true)`, f.UserID); err != nil {
		return fixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
	INSERT INTO companies (company_id, owner_user_id, slug, name, created_by_user_id)
	VALUES ($1, $2, $3, 'Worker Topology Smoke', $2)`, f.CompanyID, f.UserID, "worker-topology-"+f.CompanyID.String()); err != nil {
		return fixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
	INSERT INTO projects (project_id, company_id, name, slug, hook_key, hook_secret_hash, created_by_user_id)
	VALUES ($1, $2, 'Worker Topology Smoke', $3, $4, 'hash', $5)`,
		f.ProjectID, f.CompanyID, "worker-topology-"+f.ProjectID.String(), "hook-"+f.ProjectID.String(), f.UserID); err != nil {
		return fixtureIDs{}, err
	}
	def := `{"nodes":[{"action_key":"topology_stop","action_type_key":"stop","config":{}}]}`
	if _, err := pool.Exec(ctx, `
	INSERT INTO workflows (workflow_id, project_id, name, slug, status, draft_definition_json, created_by_user_id)
	VALUES ($1, $2, 'Worker Topology Smoke', $3, 'active', $4::jsonb, $5)`,
		f.WorkflowID, f.ProjectID, "worker-topology-"+f.WorkflowID.String(), def, f.UserID); err != nil {
		return fixtureIDs{}, err
	}
	if _, err := pool.Exec(ctx, `
	INSERT INTO workflow_versions (workflow_version_id, project_id, workflow_id, version_no, status, definition_json, created_by_user_id)
	VALUES ($1, $2, $3, 1, 'published', $4::jsonb, $5)`,
		f.VersionID, f.ProjectID, f.WorkflowID, def, f.UserID); err != nil {
		return fixtureIDs{}, err
	}
	tag, err := pool.Exec(ctx, `UPDATE workflows SET current_workflow_version_id = $2 WHERE workflow_id = $1`, f.WorkflowID, f.VersionID)
	if err != nil {
		return fixtureIDs{}, err
	}
	if tag.RowsAffected() != 1 {
		return fixtureIDs{}, errors.New("workflow version link was not updated")
	}
	return f, nil
}

func cleanupFixture(ctx context.Context, pool *pgxpool.Pool, companyID uuid.UUID) {
	_, _ = pool.Exec(ctx, `DELETE FROM companies WHERE company_id = $1`, companyID)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
