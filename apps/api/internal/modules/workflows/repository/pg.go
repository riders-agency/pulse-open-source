package repository

import (
	"context"
	"encoding/json"
	"strings"

	webhookcommands "riders.agency/pulse/api/internal/modules/webhooks/commands"
	"riders.agency/pulse/api/internal/modules/workflows/contracts"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	dbpkg "riders.agency/pulse/api/internal/platform/db"

	"riders.agency/pulse/api/internal/modules/workflows/definitionvalidate"
	"riders.agency/pulse/api/internal/modules/workflows/paymentworkflow"
	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
)

type Workflows struct {
	P dbpkg.DB
}

type WorkflowSummary struct {
	ID       string `json:"workflow_id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Status   string `json:"status"`
	RunCount int64  `json:"run_count"`
}

type WorkflowDetail struct {
	WorkflowSummary
	ProjectID              string          `json:"project_id"`
	DraftDefinition        json.RawMessage `json:"draft_definition_json,omitempty"`
	DraftGraph             json.RawMessage `json:"draft_graph_json,omitempty"`
	DraftDSLYAML           *string         `json:"draft_dsl_yaml,omitempty"`
	TestInput              json.RawMessage `json:"test_input_json,omitempty"`
	CurrentWorkflowVersion *string         `json:"current_workflow_version_id,omitempty"`
}

func (w *Workflows) assertWorkflow(ctx context.Context, wfID, uid uuid.UUID) (projectID uuid.UUID, err error) {
	if err := authz.RequireWorkflow(ctx, w.P, wfID, authz.PermRead); err != nil {
		return uuid.Nil, err
	}
	row := w.P.QueryRow(ctx, `
SELECT w.project_id
  FROM workflows w
  JOIN pulse_read.projects_v1 pr ON pr.project_id = w.project_id
  JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE w.workflow_id = $1`, wfID.String(), uid.String())
	err = row.Scan(&projectID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, errs.ErrForbidden
		}
		return uuid.Nil, err
	}
	return projectID, nil
}

func (w *Workflows) assertProject(ctx context.Context, pid, uid uuid.UUID) error {
	if err := authz.RequireProjectProcess(ctx, w.P, pid); err != nil {
		return err
	}
	row := w.P.QueryRow(ctx, `
SELECT 1 FROM pulse_read.projects_v1 pr
 JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE pr.project_id = $1`, pid.String(), uid.String())
	var one int
	return row.Scan(&one)
}

func (w *Workflows) ListByProject(ctx context.Context, projectID, uid uuid.UUID) ([]WorkflowSummary, error) {
	if err := w.assertProject(ctx, projectID, uid); err != nil {
		return nil, errs.ErrForbidden
	}
	all, allowedIDs, err := authz.AllowWorkflowIDs(ctx, w.P, projectID)
	if err != nil {
		return nil, err
	}
	if !all && len(allowedIDs) == 0 {
		return []WorkflowSummary{}, nil
	}
	rows, err := w.P.Query(ctx, `
SELECT w.workflow_id::text,
       w.name::text,
       w.slug::text,
       w.status::text,
       COUNT(wr.workflow_run_id)::bigint AS run_count
  FROM workflows w
  LEFT JOIN pulse_read.workflow_runs_v1 wr ON wr.workflow_id = w.workflow_id
 WHERE w.project_id = $1
   AND ($2::uuid[] IS NULL OR w.workflow_id = ANY($2::uuid[]))
 GROUP BY w.workflow_id, w.name, w.slug, w.status
 ORDER BY w.name`, projectID.String(), nullableUUIDFilter(all, allowedIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]WorkflowSummary, 0)
	for rows.Next() {
		var s WorkflowSummary
		if err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.Status, &s.RunCount); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (w *Workflows) Create(ctx context.Context, projectID, uid uuid.UUID, name, slug string) (*WorkflowSummary, error) {
	if err := authz.RequireProjectProcess(ctx, w.P, projectID); err != nil {
		return nil, errs.ErrForbidden
	}
	if err := w.assertProject(ctx, projectID, uid); err != nil {
		return nil, errs.ErrForbidden
	}
	wid := uuid.New()
	emptyDefinition := `{"nodes":[],"edges":[]}`
	_, err := w.P.Exec(ctx, `
INSERT INTO workflows (
  workflow_id, project_id, name, slug, status,
  draft_definition_json, draft_graph_json, test_input_json, created_by_user_id)
VALUES ($1,$2,$3,$4,'draft',$5::jsonb,$6::jsonb,$7::jsonb,$8)`,
		wid.String(), projectID.String(), name, slug, emptyDefinition, emptyDefinition, `{}`, uid.String())
	if err != nil {
		return nil, err
	}
	return &WorkflowSummary{ID: wid.String(), Name: name, Slug: slug, Status: "draft"}, nil
}

func (w *Workflows) Get(ctx context.Context, wfID, uid uuid.UUID) (*WorkflowDetail, error) {
	_, err := w.assertWorkflow(ctx, wfID, uid)
	if err != nil {
		return nil, err
	}
	row := w.P.QueryRow(ctx, `
SELECT w.workflow_id::text, w.project_id::text, w.name::text, w.slug::text, w.status::text,
       w.draft_definition_json, w.draft_graph_json, w.draft_dsl_yaml,
       w.test_input_json,
       w.current_workflow_version_id::text,
       (SELECT COUNT(*)::bigint FROM pulse_read.workflow_runs_v1 wr WHERE wr.workflow_id = w.workflow_id) AS run_count
  FROM workflows w WHERE w.workflow_id = $1`, wfID.String())
	var d WorkflowDetail
	var cur *string
	var graph, def, testInput []byte
	var dsl *string
	if err := row.Scan(&d.ID, &d.ProjectID, &d.Name, &d.Slug, &d.Status,
		&def, &graph, &dsl, &testInput, &cur, &d.RunCount); err != nil {
		return nil, err
	}
	d.DraftDefinition = def
	if len(graph) > 0 {
		d.DraftGraph = graph
	}
	if len(testInput) > 0 {
		d.TestInput = testInput
	}
	d.DraftDSLYAML = dsl
	if cur != nil && *cur != "" {
		d.CurrentWorkflowVersion = cur
	}
	d.WorkflowSummary = WorkflowSummary{ID: d.ID, Name: d.Name, Slug: d.Slug, Status: d.Status, RunCount: d.RunCount}
	return &d, nil
}

type DraftPatch = contracts.DraftPatch

func (w *Workflows) PutDraft(ctx context.Context, wfID, uid uuid.UUID, patch DraftPatch) error {
	if err := authz.RequireWorkflow(ctx, w.P, wfID, authz.PermWrite); err != nil {
		return err
	}
	_, err := w.assertWorkflow(ctx, wfID, uid)
	if err != nil {
		return err
	}
	def := patch.Definition
	graph := patch.Graph
	if len(def) > 0 || len(graph) > 0 {
		if len(def) == 0 {
			var current []byte
			_ = w.P.QueryRow(ctx, `SELECT draft_definition_json FROM workflows WHERE workflow_id = $1`, wfID.String()).Scan(&current)
			def = current
		}
		if len(graph) == 0 {
			var current []byte
			_ = w.P.QueryRow(ctx, `SELECT draft_graph_json FROM workflows WHERE workflow_id = $1`, wfID.String()).Scan(&current)
			graph = current
		}
		if normDef, normGraph, ok := paymentworkflow.NormalizeDraft(def, graph); ok {
			def = normDef
			graph = normGraph
		}
	}
	_, err = w.P.Exec(ctx, `
UPDATE workflows SET
  draft_definition_json = COALESCE($2::jsonb, draft_definition_json),
  draft_graph_json = COALESCE($3::jsonb, draft_graph_json),
  draft_dsl_yaml = COALESCE($4, draft_dsl_yaml),
  test_input_json = COALESCE($5::jsonb, test_input_json),
  updated_at = now()
 WHERE workflow_id = $1`,
		wfID.String(),
		nullJSON(def),
		nullJSON(graph),
		patch.DSLYAML,
		nullJSON(patch.TestInput))
	return err
}

func nullJSON(b json.RawMessage) any {
	if len(b) == 0 {
		return nil
	}
	return string(b)
}

func nullableUUIDFilter(all bool, ids []uuid.UUID) any {
	if all {
		return nil
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}

func (w *Workflows) PublishTx(ctx context.Context, wfID, uid uuid.UUID, description string) (versionID string, versionNo int, err error) {
	if err := authz.RequireWorkflow(ctx, w.P, wfID, authz.PermWrite); err != nil {
		return "", 0, err
	}
	pid, err := w.assertWorkflow(ctx, wfID, uid)
	if err != nil {
		return "", 0, err
	}
	description = strings.TrimSpace(description)
	if len(description) > 2000 {
		description = description[:2000]
	}
	tx, err := w.transaction()
	if err != nil {
		return "", 0, err
	}

	var def []byte
	var graph []byte
	var dsl *string
	if err := tx.QueryRow(ctx, `
SELECT draft_definition_json, draft_graph_json, draft_dsl_yaml FROM workflows WHERE workflow_id = $1`,
		wfID.String()).Scan(&def, &graph, &dsl); err != nil {
		return "", 0, err
	}
	if normDef, normGraph, ok := paymentworkflow.NormalizeDraft(def, graph); ok {
		def = normDef
		graph = normGraph
		_, _ = tx.Exec(ctx, `
UPDATE workflows SET draft_definition_json = $2::jsonb, draft_graph_json = $3::jsonb, updated_at = now()
 WHERE workflow_id = $1`, wfID.String(), string(def), string(graph))
	}
	if err := definitionvalidate.Validate(def); err != nil {
		return "", 0, err
	}
	var next int
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(MAX(version_no),0)+1 FROM workflow_versions WHERE workflow_id = $1`, wfID.String()).
		Scan(&next); err != nil {
		return "", 0, err
	}
	vid := uuid.New()
	graphb := graph
	if graphb == nil {
		graphb = []byte("null")
	}
	_, err = tx.Exec(ctx, `
INSERT INTO workflow_versions (
  workflow_version_id, project_id, workflow_id, version_no, status,
  definition_json, graph_json, dsl_yaml, created_by_user_id, description)
VALUES ($1,$2,$3,$4,'published',$5::jsonb,$6::jsonb,$7,$8,$9)`,
		vid.String(), pid.String(), wfID.String(), next,
		string(def), string(graphb), dsl, uid.String(), description)
	if err != nil {
		return "", 0, err
	}
	_, err = tx.Exec(ctx, `
UPDATE workflows SET
  status = 'active',
  current_workflow_version_id = $2,
  updated_at = now()
 WHERE workflow_id = $1`, wfID.String(), vid.String())
	if err != nil {
		return "", 0, err
	}
	err = webhookcommands.FollowWorkflowVersion(ctx, tx, wfID, vid.String())
	if err != nil {
		return "", 0, err
	}
	return vid.String(), next, nil
}

func (w *Workflows) ListVersions(ctx context.Context, wfID, uid uuid.UUID) ([]map[string]any, error) {
	_, err := w.assertWorkflow(ctx, wfID, uid)
	if err != nil {
		return nil, err
	}
	rows, err := w.P.Query(ctx, `
SELECT wv.workflow_version_id::text, wv.version_no, wv.status::text, wv.description::text,
       wv.created_at, wv.published_at,
       (w.current_workflow_version_id = wv.workflow_version_id) AS is_current
  FROM workflow_versions wv
  JOIN workflows w ON w.workflow_id = wv.workflow_id
 WHERE wv.workflow_id = $1 AND wv.status <> 'test'
 ORDER BY wv.version_no DESC`, wfID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, status, description string
		var vn int
		var created, pub any
		var isCurrent bool
		if err := rows.Scan(&id, &vn, &status, &description, &created, &pub, &isCurrent); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"workflow_version_id": id,
			"version_no":          vn,
			"status":              status,
			"description":         description,
			"created_at":          created,
			"published_at":        pub,
			"is_current":          isCurrent,
		})
	}
	return out, rows.Err()
}

func (w *Workflows) VersionByID(ctx context.Context, wvID, uid uuid.UUID) (map[string]any, error) {
	row := w.P.QueryRow(ctx, `
SELECT wv.workflow_version_id::text, wv.workflow_id::text, wv.version_no, wv.status::text,
       wv.description::text, wv.definition_json, wv.graph_json, wv.dsl_yaml
  FROM workflow_versions wv
  JOIN workflows w ON w.workflow_id = wv.workflow_id
  JOIN pulse_read.projects_v1 pr ON pr.project_id = w.project_id
  JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE wv.workflow_version_id = $1`, wvID.String(), uid.String())
	var id, wf, st, description string
	var vn int
	var def, graph []byte
	var dsl *string
	if err := row.Scan(&id, &wf, &vn, &st, &description, &def, &graph, &dsl); err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrForbidden
		}
		return nil, err
	}
	return map[string]any{
		"workflow_version_id": id,
		"workflow_id":         wf,
		"version_no":          vn,
		"status":              st,
		"description":         description,
		"definition_json":     json.RawMessage(def),
		"graph_json":          json.RawMessage(graph),
		"dsl_yaml":            dsl,
	}, nil
}

func (w *Workflows) RestoreVersionTx(ctx context.Context, wfID, wvID, uid uuid.UUID) (map[string]any, error) {
	_, err := w.assertWorkflow(ctx, wfID, uid)
	if err != nil {
		return nil, err
	}
	tx, err := w.transaction()
	if err != nil {
		return nil, err
	}

	var id, workflowID, status, description string
	var versionNo int
	var def, graph []byte
	var dsl *string
	if err := tx.QueryRow(ctx, `
SELECT workflow_version_id::text, workflow_id::text, version_no, status::text, description::text,
       definition_json, graph_json, dsl_yaml
  FROM workflow_versions
 WHERE workflow_version_id = $1 AND workflow_id = $2 AND status <> 'test'`,
		wvID.String(), wfID.String()).Scan(&id, &workflowID, &versionNo, &status, &description, &def, &graph, &dsl); err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrForbidden
		}
		return nil, err
	}
	graphb := graph
	if graphb == nil {
		graphb = []byte("null")
	}
	_, err = tx.Exec(ctx, `
UPDATE workflows
   SET status = 'active',
       draft_definition_json = $2::jsonb,
       draft_graph_json = $3::jsonb,
       draft_dsl_yaml = $4,
       current_workflow_version_id = $5,
       updated_at = now()
 WHERE workflow_id = $1`,
		wfID.String(), string(def), string(graphb), dsl, id)
	if err != nil {
		return nil, err
	}
	err = webhookcommands.FollowWorkflowVersion(ctx, tx, wfID, id)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"workflow_version_id": id,
		"workflow_id":         workflowID,
		"version_no":          versionNo,
		"status":              status,
		"description":         description,
		"restored":            true,
	}, nil
}
