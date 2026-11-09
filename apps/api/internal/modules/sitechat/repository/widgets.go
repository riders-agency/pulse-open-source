package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/platform/authz"
)

func (s *SiteChat) ListWidgets(ctx context.Context, projectID uuid.UUID) ([]WidgetRow, error) {
	if err := authz.RequireProject(ctx, s.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := s.P.Query(ctx, `
SELECT chat_widget_id::text, company_id::text, project_id::text, public_key::text,
       name::text, status::text, title::text, subtitle::text, accent_color::text, position::text,
       heavy_tracking_default, sampling_rate::float8, settings_json, created_at, updated_at
  FROM site_chat_widgets
 WHERE project_id = $1
 ORDER BY created_at DESC`, projectID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []WidgetRow{}
	for rows.Next() {
		item, err := scanWidget(rows)
		if err != nil {
			return nil, err
		}
		item.EmbedScript = embedScript(item.PublicKey)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *SiteChat) CreateWidget(ctx context.Context, projectID, uid uuid.UUID, in WidgetInput) (string, error) {
	if err := authz.RequireProject(ctx, s.P, projectID, authz.PermWrite); err != nil {
		return "", err
	}
	in = normalizeWidgetInput(in)
	companyID, err := s.projectCompany(ctx, projectID)
	if err != nil {
		return "", err
	}
	id := uuid.New()
	key := "scw_" + randomToken(24)
	_, err = s.P.Exec(ctx, `
INSERT INTO site_chat_widgets (
  chat_widget_id, company_id, project_id, public_key, name, title, subtitle, accent_color,
  position, heavy_tracking_default, sampling_rate, settings_json, created_by_user_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::jsonb,$13)`,
		id.String(), companyID.String(), projectID.String(), key, in.Name, in.Title, in.Subtitle,
		in.AccentColor, in.Position, in.HeavyTrackingDefault, in.SamplingRate, jsonBytes(in.Settings), uid.String())
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func (s *SiteChat) PatchWidget(ctx context.Context, widgetID uuid.UUID, patch WidgetPatch) error {
	projectID, err := s.widgetProject(ctx, widgetID)
	if err != nil {
		return err
	}
	if err := authz.RequireProject(ctx, s.P, projectID, authz.PermWrite); err != nil {
		return err
	}
	_, err = s.P.Exec(ctx, `
UPDATE site_chat_widgets
   SET name = COALESCE(NULLIF(btrim($2), ''), name),
       status = COALESCE(NULLIF(btrim($3), ''), status),
       title = COALESCE(NULLIF(btrim($4), ''), title),
       subtitle = COALESCE(NULLIF(btrim($5), ''), subtitle),
       accent_color = COALESCE(NULLIF(btrim($6), ''), accent_color),
       position = COALESCE(NULLIF(btrim($7), ''), position),
       heavy_tracking_default = COALESCE($8, heavy_tracking_default),
       sampling_rate = COALESCE($9, sampling_rate),
       settings_json = COALESCE($10::jsonb, settings_json)
 WHERE chat_widget_id = $1`,
		widgetID.String(), optString(patch.Name), optString(patch.Status), optString(patch.Title),
		optString(patch.Subtitle), optString(patch.AccentColor), optString(patch.Position),
		patch.HeavyTrackingDefault, patch.SamplingRate, optJSON(patch.Settings))
	return err
}

func (s *SiteChat) PublicWidgetConfig(ctx context.Context, publicKey, endpointBase string) (*PublicWidgetConfig, error) {
	row := s.P.QueryRow(ctx, `
SELECT chat_widget_id::text, project_id::text, public_key::text, title::text, subtitle::text,
       accent_color::text, position::text, heavy_tracking_default, sampling_rate::float8, settings_json
  FROM site_chat_widgets
 WHERE public_key = $1 AND status = 'active'`, strings.TrimSpace(publicKey))
	var out PublicWidgetConfig
	if err := row.Scan(&out.ChatWidgetID, &out.ProjectID, &out.PublicKey, &out.Title, &out.Subtitle,
		&out.AccentColor, &out.Position, &out.HeavyTrackingDefault, &out.SamplingRate, (*jsonMap)(&out.Settings)); err != nil {
		return nil, mapErr(err)
	}
	out.EndpointBase = endpointBase
	out.Settings = withDefaultWidgetSettings(out.Settings)
	return &out, nil
}

func scanWidget(rows pgx.Rows) (WidgetRow, error) {
	var item WidgetRow
	if err := rows.Scan(&item.ChatWidgetID, &item.CompanyID, &item.ProjectID, &item.PublicKey,
		&item.Name, &item.Status, &item.Title, &item.Subtitle, &item.AccentColor, &item.Position,
		&item.HeavyTrackingDefault, &item.SamplingRate, (*jsonMap)(&item.Settings), &item.CreatedAt, &item.UpdatedAt); err != nil {
		return item, err
	}
	item.Settings = withDefaultWidgetSettings(item.Settings)
	return item, nil
}

func normalizeWidgetInput(in WidgetInput) WidgetInput {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		in.Name = "Чат на сайте"
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		in.Title = "Чат на сайте"
	}
	in.Subtitle = strings.TrimSpace(in.Subtitle)
	if in.Subtitle == "" {
		in.Subtitle = ""
	}
	in.AccentColor = strings.TrimSpace(in.AccentColor)
	if in.AccentColor == "" {
		in.AccentColor = "#00B2FF"
	}
	if in.Position != "left" {
		in.Position = "right"
	}
	if in.SamplingRate <= 0 || in.SamplingRate > 1 {
		in.SamplingRate = 1
	}
	in.Settings = withDefaultWidgetSettings(in.Settings)
	return in
}

func defaultWidgetSettings() map[string]any {
	return map[string]any{
		"enabled_modes": map[string]any{"chat": true, "call": true, "monitoring": true},
		"launcher": map[string]any{
			"primary_text": "Консультация онлайн",
			"call_text":    "Звонок",
		},
		"privacy_policy_url":         "",
		"call_terms_url":             "",
		"recording_enabled":          false,
		"recording_consent_required": true,
		"call_consent_text":          "Мы обработаем аудио только для соединения с командой. Запись разговора включается только после отдельного согласия.",
		"active_invite": map[string]any{
			"enabled":          true,
			"title":            "Команда на связи",
			"body":             "Помочь на этой странице?",
			"accept_text":      "Ответить",
			"call_text":        "Звонок",
			"decline_text":     "Не сейчас",
			"cooldown_minutes": 10,
		},
		"monitoring": map[string]any{
			"privacy_mode":    true,
			"mask_inputs":     true,
			"mask_text":       false,
			"mask_selectors":  []string{"[data-rp-mask]", ".rp-mask"},
			"block_selectors": []string{"[data-rp-block]", ".rp-block"},
			"capture_iframes": false,
		},
		"theme": map[string]any{
			"accent_color": "#00B2FF",
			"position":     "right",
		},
	}
}

func withDefaultWidgetSettings(settings map[string]any) map[string]any {
	out := defaultWidgetSettings()
	mergeSettings(out, settings)
	return out
}

func mergeSettings(dst, src map[string]any) {
	for key, value := range src {
		if child, ok := value.(map[string]any); ok {
			if baseChild, ok := dst[key].(map[string]any); ok {
				mergeSettings(baseChild, child)
				continue
			}
		}
		dst[key] = value
	}
}
