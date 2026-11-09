package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
)

type MessageTemplates struct {
	P *pgxpool.Pool
}

func (m *MessageTemplates) assertProject(ctx context.Context, pid, uid uuid.UUID) error {
	if err := authz.RequireProject(ctx, m.P, pid, authz.PermRead); err != nil {
		return err
	}
	row := m.P.QueryRow(ctx, `
SELECT 1 FROM pulse_read.projects_v1 pr
 JOIN pulse_read.company_memberships_v1 cm ON cm.company_id = pr.company_id AND cm.user_id = $2
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

func (m *MessageTemplates) List(ctx context.Context, pid, uid uuid.UUID) ([]map[string]any, error) {
	if err := m.assertProject(ctx, pid, uid); err != nil {
		return nil, err
	}
	rows, err := m.P.Query(ctx, `
SELECT message_template_id::text, name::text, slug::text, provider_key::text,
       body_format::text, metadata_json, created_at
  FROM message_templates WHERE project_id = $1 ORDER BY name`, pid.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, name, slug, pk, bf string
		var meta []byte
		var ca any
		if err := rows.Scan(&id, &name, &slug, &pk, &bf, &meta, &ca); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"message_template_id": id,
			"name":                name,
			"slug":                slug,
			"provider_key":        pk,
			"body_format":         bf,
			"metadata_json":       json.RawMessage(meta),
			"created_at":          ca,
		})
	}
	return out, rows.Err()
}

type MessageCreate struct {
	Name         string          `json:"name"`
	Slug         string          `json:"slug"`
	ProviderKey  string          `json:"provider_key"`
	BodyTemplate string          `json:"body_template"`
	BodyFormat   string          `json:"body_format"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
}

func (m *MessageTemplates) Create(ctx context.Context, pid, uid uuid.UUID, in MessageCreate) (string, error) {
	if err := m.assertProject(ctx, pid, uid); err != nil {
		return "", err
	}
	if in.BodyFormat == "" {
		in.BodyFormat = "markdown"
	}
	meta := "{}"
	if len(in.MetadataJSON) > 0 {
		meta = string(in.MetadataJSON)
	}
	id := uuid.New()
	_, err := m.P.Exec(ctx, `
INSERT INTO message_templates (
  message_template_id, project_id, name, slug, provider_key,
  body_template, body_format, metadata_json, created_by_user_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9)`,
		id.String(), pid.String(), in.Name, in.Slug, in.ProviderKey,
		in.BodyTemplate, in.BodyFormat, meta, uid.String())
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func (m *MessageTemplates) assertRow(ctx context.Context, pid, mid, uid uuid.UUID) error {
	row := m.P.QueryRow(ctx, `
SELECT 1 FROM message_templates mt
 JOIN pulse_read.projects_v1 pr ON pr.project_id = mt.project_id
 JOIN pulse_read.company_memberships_v1 cm ON cm.company_id = pr.company_id AND cm.user_id = $3
 WHERE mt.message_template_id = $1 AND mt.project_id = $2`,
		mid.String(), pid.String(), uid.String())
	var x int
	if err := row.Scan(&x); err != nil {
		if err == pgx.ErrNoRows {
			return errs.ErrForbidden
		}
		return err
	}
	return nil
}

func (m *MessageTemplates) Get(ctx context.Context, pid, mid, uid uuid.UUID) (map[string]any, error) {
	if err := m.assertRow(ctx, pid, mid, uid); err != nil {
		return nil, err
	}
	row := m.P.QueryRow(ctx, `
SELECT message_template_id::text, name::text, slug::text, provider_key::text,
       body_template::text, body_format::text, metadata_json, created_at
  FROM message_templates WHERE message_template_id = $1`, mid.String())
	var id, name, slug, pk, body, bf string
	var meta []byte
	var ca any
	if err := row.Scan(&id, &name, &slug, &pk, &body, &bf, &meta, &ca); err != nil {
		return nil, err
	}
	return map[string]any{
		"message_template_id": id,
		"name":                name,
		"slug":                slug,
		"provider_key":        pk,
		"body_template":       body,
		"body_format":         bf,
		"metadata_json":       json.RawMessage(meta),
		"created_at":          ca,
	}, nil
}

type MessagePatch struct {
	Name         *string         `json:"name"`
	Slug         *string         `json:"slug"`
	ProviderKey  *string         `json:"provider_key"`
	BodyTemplate *string         `json:"body_template"`
	BodyFormat   *string         `json:"body_format"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
}

func (m *MessageTemplates) Patch(ctx context.Context, pid, mid, uid uuid.UUID, p MessagePatch) error {
	if err := m.assertRow(ctx, pid, mid, uid); err != nil {
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
	if p.ProviderKey != nil {
		sets = append(sets, fmt.Sprintf("provider_key = $%d", i))
		args = append(args, *p.ProviderKey)
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
	if len(p.MetadataJSON) > 0 {
		sets = append(sets, fmt.Sprintf("metadata_json = $%d::jsonb", i))
		args = append(args, string(p.MetadataJSON))
		i++
	}
	if len(sets) == 0 {
		return nil
	}
	args = append(args, mid.String())
	q := fmt.Sprintf(`UPDATE message_templates SET %s, updated_at = now() WHERE message_template_id = $%d`,
		strings.Join(sets, ", "), i)
	_, err := m.P.Exec(ctx, q, args...)
	return err
}

func (m *MessageTemplates) Delete(ctx context.Context, pid, mid, uid uuid.UUID) error {
	if err := m.assertRow(ctx, pid, mid, uid); err != nil {
		return err
	}
	_, err := m.P.Exec(ctx, `DELETE FROM message_templates WHERE message_template_id = $1 AND project_id = $2`,
		mid.String(), pid.String())
	return err
}
