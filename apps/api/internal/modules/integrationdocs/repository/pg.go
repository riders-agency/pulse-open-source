package repository

import (
	"context"
	"encoding/json"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/errs"
)

type Docs struct {
	P *pgxpool.Pool
}

type SearchParams struct {
	Query         string
	ProviderKey   string
	ActionTypeKey string
	Limit         int
	ProjectID     *uuid.UUID
	UserID        *uuid.UUID
}

type SearchResult struct {
	DocKey             string          `json:"doc_key"`
	PublicSlug         string          `json:"public_slug"`
	Title              string          `json:"title"`
	Summary            string          `json:"summary"`
	ProviderKey        string          `json:"provider_key"`
	ActionTypeKeys     []string        `json:"action_type_keys"`
	DependencyTypeKeys []string        `json:"dependency_type_keys"`
	Tags               []string        `json:"tags"`
	Intents            []string        `json:"intents"`
	Score              float64         `json:"score"`
	Reason             string          `json:"reason"`
	Snippet            string          `json:"snippet"`
	PublicURL          string          `json:"public_url"`
	MarkdownURL        string          `json:"markdown_url"`
	Available          json.RawMessage `json:"available_integrations,omitempty"`
}

func (d *Docs) assertProject(ctx context.Context, pid, uid uuid.UUID) error {
	if err := authz.RequireProject(ctx, d.P, pid, authz.PermRead); err != nil {
		return err
	}
	row := d.P.QueryRow(ctx, `
SELECT 1 FROM pulse_read.projects_v1 pr
 JOIN pulse_read.company_memberships_v1 m ON m.company_id = pr.company_id AND m.user_id = $2
 WHERE pr.project_id = $1`, pid.String(), uid.String())
	var one int
	if err := row.Scan(&one); err != nil {
		if err == pgx.ErrNoRows {
			return errs.ErrForbidden
		}
		return err
	}
	return nil
}

func (d *Docs) Search(ctx context.Context, p SearchParams) ([]SearchResult, error) {
	q := strings.TrimSpace(p.Query)
	provider := strings.TrimSpace(p.ProviderKey)
	actionType := strings.TrimSpace(p.ActionTypeKey)
	limit := p.Limit
	if limit <= 0 {
		limit = 8
	}
	if limit > 20 {
		limit = 20
	}

	if p.ProjectID != nil && p.UserID != nil {
		if err := d.assertProject(ctx, *p.ProjectID, *p.UserID); err != nil {
			return nil, err
		}
	}

	projectID := ""
	if p.ProjectID != nil {
		projectID = p.ProjectID.String()
	}

	rows, err := d.P.Query(ctx, searchSQL, q, provider, actionType, limit, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]SearchResult, 0)
	for rows.Next() {
		var row SearchResult
		var available []byte
		if err := rows.Scan(
			&row.DocKey,
			&row.PublicSlug,
			&row.Title,
			&row.Summary,
			&row.ProviderKey,
			&row.ActionTypeKeys,
			&row.DependencyTypeKeys,
			&row.Tags,
			&row.Intents,
			&row.Score,
			&row.Reason,
			&row.Snippet,
			&available,
		); err != nil {
			return nil, err
		}
		row.Score = math.Round(row.Score*10000) / 10000
		row.PublicURL = "https://pulse.riders.agency/docs/integrations/" + row.PublicSlug
		row.MarkdownURL = "https://pulse.riders.agency/api/v1/docs/integrations/" + row.PublicSlug + ".md"
		if len(available) > 0 && string(available) != "null" {
			row.Available = json.RawMessage(available)
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (d *Docs) MarkdownBySlug(ctx context.Context, slug string) (string, error) {
	var markdown string
	err := d.P.QueryRow(ctx, `
SELECT markdown
  FROM integration_docs
 WHERE public_slug = $1 AND status = 'active'`, slug).Scan(&markdown)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", errs.ErrForbidden
		}
		return "", err
	}
	return markdown, nil
}

const searchSQL = `
WITH params AS (
  SELECT
    trim($1::text) AS q,
    lower(unaccent(trim($1::text))) AS raw,
    NULLIF(trim($2::text), '') AS provider_filter,
    NULLIF(trim($3::text), '') AS action_filter,
    NULLIF(trim($5::text), '')::uuid AS project_id,
    CASE
      WHEN trim($1::text) = '' THEN NULL::tsquery
      ELSE websearch_to_tsquery('simple', unaccent(trim($1::text)))
    END AS tsq
),
ranked AS (
 SELECT
    d.*,
    lower(unaccent(d.doc_key || ' ' || d.title || ' ' || d.summary || ' ' || d.provider_key || ' ' ||
      array_to_string(d.action_type_keys || d.dependency_type_keys || d.tags || d.intents, ' '))) AS haystack,
    CASE
      WHEN params.q = '' THEN 0
      WHEN lower(d.doc_key) = params.raw OR lower(d.provider_key) = params.raw THEN 4
      WHEN params.raw = ANY(SELECT lower(item) FROM unnest(d.action_type_keys || d.dependency_type_keys) AS u(item)) THEN 4
      WHEN lower(unaccent(d.doc_key || ' ' || d.title || ' ' || d.summary || ' ' || d.provider_key || ' ' ||
        array_to_string(d.action_type_keys || d.dependency_type_keys || d.tags || d.intents, ' '))) LIKE '%' || params.raw || '%' THEN 2
      ELSE 0
    END AS exact_score,
    CASE
      WHEN params.tsq IS NULL THEN 0
      ELSE ts_rank_cd(d.search_vector, params.tsq, 32)
    END AS fts_score,
    CASE
      WHEN params.q = '' THEN 0
      ELSE greatest(
        similarity(lower(unaccent(d.title)), params.raw),
        similarity(lower(unaccent(d.summary)), params.raw),
        similarity(lower(unaccent(array_to_string(d.tags || d.intents || d.action_type_keys || d.dependency_type_keys, ' '))), params.raw),
        word_similarity(params.raw, lower(unaccent(d.doc_key || ' ' || d.title || ' ' || d.summary || ' ' || d.provider_key || ' ' ||
          array_to_string(d.action_type_keys || d.dependency_type_keys || d.tags || d.intents, ' '))))
      )
    END AS trigram_score,
    CASE
      WHEN params.tsq IS NULL THEN d.summary
      ELSE ts_headline('simple', d.markdown, params.tsq, 'MaxFragments=2, MinWords=4, MaxWords=18, ShortWord=2')
    END AS snippet
  FROM integration_docs d
  CROSS JOIN params
  WHERE d.status = 'active'
    AND (params.provider_filter IS NULL OR d.provider_key = params.provider_filter)
    AND (params.action_filter IS NULL OR params.action_filter = ANY(d.action_type_keys))
    AND (
      params.q = ''
      OR d.search_vector @@ params.tsq
      OR lower(unaccent(d.doc_key || ' ' || d.title || ' ' || d.summary || ' ' || d.provider_key || ' ' ||
        array_to_string(d.action_type_keys || d.dependency_type_keys || d.tags || d.intents, ' '))) LIKE '%' || params.raw || '%'
      OR similarity(lower(unaccent(d.title)), params.raw) > 0.12
      OR similarity(lower(unaccent(d.summary)), params.raw) > 0.12
      OR word_similarity(params.raw, lower(unaccent(d.doc_key || ' ' || d.title || ' ' || d.summary || ' ' || d.provider_key || ' ' ||
        array_to_string(d.action_type_keys || d.dependency_type_keys || d.tags || d.intents, ' ')))) > 0.18
    )
)
SELECT
  ranked.doc_key,
  ranked.public_slug,
  ranked.title,
  ranked.summary,
  ranked.provider_key,
  ranked.action_type_keys,
  ranked.dependency_type_keys,
  ranked.tags,
  ranked.intents,
  (ranked.exact_score + ranked.fts_score + ranked.trigram_score) AS score,
  CASE
    WHEN ranked.exact_score >= 4 THEN 'exact key/provider match'
    WHEN ranked.exact_score >= 2 THEN 'keyword intent match'
    WHEN ranked.fts_score > 0 THEN 'full-text match'
    WHEN ranked.trigram_score > 0 THEN 'fuzzy match'
    ELSE 'catalog default'
  END AS reason,
  ranked.snippet,
  COALESCE(available.items, '[]'::jsonb) AS available_integrations
FROM ranked
LEFT JOIN LATERAL (
  SELECT jsonb_agg(jsonb_build_object(
    'action_dependency_id', ad.action_dependency_id::text,
    'name', ad.name,
    'status', ad.status,
    'dependency_type_key', adt.dependency_type_key,
    'provider_key', adt.provider_key
  ) ORDER BY ad.status <> 'active', ad.name) AS items
  FROM pulse_read.action_dependencies_v1 ad
  JOIN pulse_read.action_dependency_types_v1 adt ON adt.action_dependency_type_id = ad.action_dependency_type_id
  CROSS JOIN params
  WHERE params.project_id IS NOT NULL
    AND ad.project_id = params.project_id
    AND ad.status <> 'deleted'
    AND adt.dependency_type_key = ANY(ranked.dependency_type_keys)
) available ON true
ORDER BY score DESC, ranked.title
LIMIT $4`
