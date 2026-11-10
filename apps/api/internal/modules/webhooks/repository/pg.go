package repository

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	dbpkg "riders.agency/pulse/api/internal/platform/db"

	"riders.agency/pulse/api/internal/modules/workflows/definitionvalidate"
	"riders.agency/pulse/api/internal/platform/authz"
	cryptoapp "riders.agency/pulse/api/internal/platform/crypto"
	"riders.agency/pulse/api/internal/platform/errs"
)

type Webhooks struct {
	P         dbpkg.DB
	MasterHEX string
}

func (w *Webhooks) assertWorkflow(ctx context.Context, wfID, uid uuid.UUID) error {
	if err := authz.RequireWorkflow(ctx, w.P, wfID, authz.PermRead); err != nil {
		return err
	}
	row := w.P.QueryRow(ctx, `
SELECT 1 FROM pulse_read.workflows_v1 wf
 JOIN pulse_read.projects_v1 pr ON pr.project_id = wf.project_id
 JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE wf.workflow_id = $1`, wfID.String(), uid.String())
	var one int
	return row.Scan(&one)
}

func (w *Webhooks) ListByWorkflow(ctx context.Context, wfID, uid uuid.UUID) ([]map[string]any, error) {
	if err := w.assertWorkflow(ctx, wfID, uid); err != nil {
		return nil, errs.ErrForbidden
	}
	rows, err := w.P.Query(ctx, `
SELECT webhook_endpoint_id::text, endpoint_key::text, name::text, status::text, auth_type::text
  FROM webhook_endpoints WHERE workflow_id = $1 ORDER BY name`, wfID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, ek, nm, st, at string
		if err := rows.Scan(&id, &ek, &nm, &st, &at); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"webhook_endpoint_id": id,
			"endpoint_key":        ek,
			"name":                nm,
			"status":              st,
			"auth_type":           at,
		})
	}
	return out, rows.Err()
}

type CreateEndpointInput struct {
	Name              string `json:"name"`
	EndpointKey       string `json:"endpoint_key"`
	WorkflowVersionID string `json:"workflow_version_id"`
}

func (w *Webhooks) Create(ctx context.Context, wfID, uid uuid.UUID, in CreateEndpointInput) (map[string]any, error) {
	if err := w.assertWorkflow(ctx, wfID, uid); err != nil {
		return nil, errs.ErrForbidden
	}
	var pid string
	if err := w.P.QueryRow(ctx, `SELECT project_id::text FROM pulse_read.workflows_v1 WHERE workflow_id = $1`, wfID.String()).
		Scan(&pid); err != nil {
		return nil, err
	}
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	plain := base64.RawURLEncoding.EncodeToString(raw)
	wvid, err := uuid.Parse(in.WorkflowVersionID)
	if err != nil {
		return nil, errs.ErrForbidden
	}
	var wvFound string
	if err := w.P.QueryRow(ctx, `
SELECT workflow_version_id::text FROM pulse_read.workflow_versions_v1
 WHERE workflow_version_id = $1 AND workflow_id = $2 AND project_id::text = $3`,
		wvid.String(), wfID.String(), pid).
		Scan(&wvFound); err != nil {
		return nil, err
	}
	var def []byte
	if err := w.P.QueryRow(ctx, `
SELECT definition_json FROM pulse_read.workflow_versions_v1 WHERE workflow_version_id = $1`,
		wvid.String()).Scan(&def); err != nil {
		return nil, err
	}
	if err := definitionvalidate.Validate(def); err != nil {
		return nil, err
	}
	secHashBytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	secretCiphertext, err := cryptoapp.Seal([]byte(plain), w.MasterHEX)
	if err != nil {
		return nil, err
	}
	eid := uuid.New()
	_, err = w.P.Exec(ctx, `
INSERT INTO webhook_endpoints (
  webhook_endpoint_id, project_id, workflow_id, workflow_version_id,
  name, endpoint_key, status, auth_type, secret_hash, secret_ciphertext, created_by_user_id
) VALUES ($1,$2,$3,$4,$5,$6,'active','bearer',$7,$8,$9)`,
		eid.String(), pid, wfID.String(), wvid.String(),
		in.Name, in.EndpointKey, string(secHashBytes), secretCiphertext, uid.String())
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"webhook_endpoint_id":   eid.String(),
		"endpoint_bearer_plain": plain,
		"endpoint_bearer_note":  "Shown once; Authorization: Bearer … on this webhook if endpoint auth enforced",
	}, nil
}

func (w *Webhooks) Get(ctx context.Context, weID, uid uuid.UUID) (hookKey string, ek string, err error) {
	row := w.P.QueryRow(ctx, `
SELECT pr.hook_key::text, we.endpoint_key::text
  FROM webhook_endpoints we
  JOIN pulse_read.projects_v1 pr ON pr.project_id = we.project_id
  JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE we.webhook_endpoint_id = $1`, weID.String(), uid.String())
	if err := row.Scan(&hookKey, &ek); err != nil {
		if err == pgx.ErrNoRows {
			return "", "", errs.ErrForbidden
		}
		return "", "", err
	}
	return hookKey, ek, nil
}

func PublicAPIBase() string {
	if b := os.Getenv("RP_PUBLIC_API_BASE"); b != "" {
		return strings.TrimRight(b, "/")
	}
	return "https://pulse.riders.agency"
}
