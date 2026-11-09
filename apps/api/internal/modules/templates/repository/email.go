package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
)

type EmailTemplates struct {
	P *pgxpool.Pool
}

func (t *EmailTemplates) assertProject(ctx context.Context, pid, uid uuid.UUID) error {
	if err := authz.RequireProject(ctx, t.P, pid, authz.PermRead); err != nil {
		return err
	}
	row := t.P.QueryRow(ctx, `
SELECT 1 FROM pulse_read.projects_v1 pr
 JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE pr.project_id = $1`, pid.String(), uid.String())
	var x int
	if err := row.Scan(&x); err != nil {
		if err == pgx.ErrNoRows {
			return errs.ErrForbidden
		}
		return err
	}
	return nil
}

func (t *EmailTemplates) List(ctx context.Context, pid, uid uuid.UUID) ([]map[string]any, error) {
	if err := t.assertProject(ctx, pid, uid); err != nil {
		return nil, err
	}
	rows, err := t.P.Query(ctx, `
SELECT email_template_id::text, name::text, slug::text, subject_template::text,
       body_format::text, created_at
  FROM email_templates WHERE project_id = $1 ORDER BY name`, pid.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, name, slug, subj, bf string
		var ca any
		if err := rows.Scan(&id, &name, &slug, &subj, &bf, &ca); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"email_template_id": id,
			"name":              name,
			"slug":              slug,
			"subject_template":  subj,
			"body_format":       bf,
			"created_at":        ca,
		})
	}
	return out, rows.Err()
}

type EmailCreate struct {
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	SubjectTemplate string `json:"subject_template"`
	BodyTemplate    string `json:"body_template"`
	BodyFormat      string `json:"body_format"`
}

func (t *EmailTemplates) Create(ctx context.Context, pid, uid uuid.UUID, in EmailCreate) (string, error) {
	if err := t.assertProject(ctx, pid, uid); err != nil {
		return "", err
	}
	if in.BodyFormat == "" {
		in.BodyFormat = "html"
	}
	id := uuid.New()
	_, err := t.P.Exec(ctx, `
INSERT INTO email_templates (
  email_template_id, project_id, name, slug, subject_template, body_template, body_format, created_by_user_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id.String(), pid.String(), in.Name, in.Slug, in.SubjectTemplate, in.BodyTemplate, in.BodyFormat, uid.String())
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func (t *EmailTemplates) assertEmailTemplate(ctx context.Context, pid, tplID, uid uuid.UUID) error {
	row := t.P.QueryRow(ctx, `
SELECT 1 FROM email_templates et
 JOIN pulse_read.projects_v1 pr ON pr.project_id = et.project_id
 JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $3
 WHERE et.email_template_id = $1 AND et.project_id = $2`,
		tplID.String(), pid.String(), uid.String())
	var x int
	if err := row.Scan(&x); err != nil {
		if err == pgx.ErrNoRows {
			return errs.ErrForbidden
		}
		return err
	}
	return nil
}

func (t *EmailTemplates) Get(ctx context.Context, pid, tplID, uid uuid.UUID) (map[string]any, error) {
	if err := t.assertEmailTemplate(ctx, pid, tplID, uid); err != nil {
		return nil, err
	}
	row := t.P.QueryRow(ctx, `
SELECT email_template_id::text, name::text, slug::text, subject_template::text,
       body_template::text, body_format::text, created_at
  FROM email_templates WHERE email_template_id = $1`, tplID.String())
	var id, name, slug, subj, body, bf string
	var ca any
	if err := row.Scan(&id, &name, &slug, &subj, &body, &bf, &ca); err != nil {
		return nil, err
	}
	return map[string]any{
		"email_template_id": id,
		"name":              name,
		"slug":              slug,
		"subject_template":  subj,
		"body_template":     body,
		"body_format":       bf,
		"created_at":        ca,
	}, nil
}

type EmailPatch struct {
	Name            *string `json:"name"`
	Slug            *string `json:"slug"`
	SubjectTemplate *string `json:"subject_template"`
	BodyTemplate    *string `json:"body_template"`
	BodyFormat      *string `json:"body_format"`
}

func (t *EmailTemplates) Patch(ctx context.Context, pid, tplID, uid uuid.UUID, p EmailPatch) error {
	if err := t.assertEmailTemplate(ctx, pid, tplID, uid); err != nil {
		return err
	}
	var sets []string
	var args []any
	i := 1
	if p.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", i))
		args = append(args, *p.Name)
		i++
	}
	if p.Slug != nil {
		sets = append(sets, fmt.Sprintf("slug = $%d", i))
		args = append(args, *p.Slug)
		i++
	}
	if p.SubjectTemplate != nil {
		sets = append(sets, fmt.Sprintf("subject_template = $%d", i))
		args = append(args, *p.SubjectTemplate)
		i++
	}
	if p.BodyTemplate != nil {
		sets = append(sets, fmt.Sprintf("body_template = $%d", i))
		args = append(args, *p.BodyTemplate)
		i++
	}
	if p.BodyFormat != nil {
		sets = append(sets, fmt.Sprintf("body_format = $%d", i))
		args = append(args, *p.BodyFormat)
		i++
	}
	if len(sets) == 0 {
		return nil
	}
	args = append(args, tplID.String())
	q := fmt.Sprintf(`UPDATE email_templates SET %s, updated_at = now() WHERE email_template_id = $%d`,
		strings.Join(sets, ", "), i)
	_, err := t.P.Exec(ctx, q, args...)
	return err
}

func (t *EmailTemplates) Delete(ctx context.Context, pid, tplID, uid uuid.UUID) error {
	if err := t.assertEmailTemplate(ctx, pid, tplID, uid); err != nil {
		return err
	}
	_, err := t.P.Exec(ctx, `DELETE FROM email_templates WHERE email_template_id = $1 AND project_id = $2`,
		tplID.String(), pid.String())
	return err
}
