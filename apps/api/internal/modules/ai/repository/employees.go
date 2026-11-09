package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"riders.agency/pulse/api/internal/modules/ai/modelgateway"
	aitools "riders.agency/pulse/api/internal/modules/ai/tools"
	"riders.agency/pulse/api/internal/platform/authz"
)

func (a *AI) EnsureEmployee(ctx context.Context, projectID, userID uuid.UUID) (map[string]any, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermRead); err != nil {
		return nil, err
	}
	employeeID, companyID, err := a.ensureEmployeeRow(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	employee, err := a.employeePayload(ctx, projectID, employeeID)
	if err != nil {
		return nil, err
	}
	project, err := a.projectPayload(ctx, projectID, companyID)
	if err != nil {
		return nil, err
	}
	deps, err := a.aiDependencies(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"employee":                employee,
		"project":                 project,
		"model_providers":         ModelProviders(),
		"configured_dependencies": deps,
		"setup_required":          !hasActiveAIProvider(deps) && !modelgateway.ServerProviderConfigured(),
	}, nil
}

func (a *AI) PatchEmployee(ctx context.Context, projectID, userID uuid.UUID, in EmployeePatch) (map[string]any, error) {
	if err := authz.RequireProject(ctx, a.P, projectID, authz.PermAdmin); err != nil {
		return nil, err
	}
	if in.ToolPolicy != nil {
		raw, err := json.Marshal(in.ToolPolicy)
		if err != nil {
			return nil, err
		}
		if _, err := aitools.ParseEmployeePolicy(raw); err != nil {
			return nil, err
		}
	}
	employeeID, _, err := a.ensureEmployeeRow(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	status := ""
	if in.Status != nil {
		status = strings.TrimSpace(*in.Status)
		if status != "active" && status != "disabled" && status != "archived" {
			return nil, fmt.Errorf("invalid status")
		}
	}
	displayName := ""
	if in.DisplayName != nil {
		displayName = strings.TrimSpace(*in.DisplayName)
	}
	modelID := ""
	if in.ModelID != nil {
		modelID = strings.TrimSpace(*in.ModelID)
	}
	depID := uuid.Nil
	if in.ActionDependencyID != nil && strings.TrimSpace(*in.ActionDependencyID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*in.ActionDependencyID))
		if err != nil {
			return nil, err
		}
		depID = parsed
	}
	_, err = a.P.Exec(ctx, `
UPDATE project_ai_employees SET
  display_name = CASE WHEN $2 <> '' THEN $2 ELSE display_name END,
  status = CASE WHEN $3 <> '' THEN $3 ELSE status END,
  action_dependency_id = CASE WHEN $4::uuid IS NULL THEN action_dependency_id ELSE $4::uuid END,
  model_id = CASE WHEN $5 <> '' THEN $5 ELSE model_id END,
  profile_json = CASE WHEN $6::jsonb IS NULL THEN profile_json ELSE $6::jsonb END,
  tool_policy_json = CASE WHEN $7::jsonb IS NULL THEN tool_policy_json ELSE $7::jsonb END,
  memory_policy_json = CASE WHEN $8::jsonb IS NULL THEN memory_policy_json ELSE $8::jsonb END,
  ui_settings_json = CASE WHEN $9::jsonb IS NULL THEN ui_settings_json ELSE $9::jsonb END
WHERE project_ai_employee_id = $1`,
		employeeID, displayName, status, nullableUUID(depID), modelID,
		jsonOrNil(in.Profile), jsonOrNil(in.ToolPolicy), jsonOrNil(in.MemoryPolicy), jsonOrNil(in.UISettings))
	if err != nil {
		return nil, err
	}
	return a.EnsureEmployee(ctx, projectID, userID)
}

func (a *AI) ensureEmployeeRow(ctx context.Context, projectID, userID uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	var employeeID, companyID uuid.UUID
	err := a.P.QueryRow(ctx, `
INSERT INTO project_ai_employees (
  project_id, company_id, created_by_user_id, profile_json, tool_policy_json, memory_policy_json, ui_settings_json
)
SELECT project_id, company_id, $2,
       '{"language":"ru","role":"project_ai_employee"}'::jsonb,
       '{"read_without_approval":true,"write_requires_approval":true,"critical_requires_elevated_approval":true}'::jsonb,
       '{"long_term_memory":"approval_only","cross_project_memory":false}'::jsonb,
       '{"widget":"bottom_right","drawer":"right"}'::jsonb
  FROM pulse_read.projects_v1
 WHERE project_id = $1
ON CONFLICT (project_id, employee_key) DO UPDATE SET
  updated_at = project_ai_employees.updated_at
RETURNING project_ai_employee_id, company_id`, projectID, userID).Scan(&employeeID, &companyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, uuid.Nil, ErrNotFound
		}
		return uuid.Nil, uuid.Nil, err
	}
	return employeeID, companyID, nil
}

func (a *AI) employeePayload(ctx context.Context, projectID, employeeID uuid.UUID) (map[string]any, error) {
	var id, eid, displayName, status, modelID, createdAt, updatedAt string
	var depID *string
	var profile, toolPolicy, memoryPolicy, uiSettings []byte
	err := a.P.QueryRow(ctx, `
SELECT project_ai_employee_id::text, employee_key::text, display_name::text, status::text,
       action_dependency_id::text, model_id::text, profile_json, tool_policy_json,
       memory_policy_json, ui_settings_json, created_at::text, updated_at::text
  FROM project_ai_employees
 WHERE project_id = $1
   AND project_ai_employee_id = $2`, projectID, employeeID).Scan(
		&id, &eid, &displayName, &status, &depID, &modelID, &profile, &toolPolicy, &memoryPolicy, &uiSettings, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"project_ai_employee_id": id,
		"employee_key":           eid,
		"display_name":           displayName,
		"status":                 status,
		"action_dependency_id":   depID,
		"model_id":               modelID,
		"profile_json":           json.RawMessage(profile),
		"tool_policy_json":       json.RawMessage(toolPolicy),
		"memory_policy_json":     json.RawMessage(memoryPolicy),
		"ui_settings_json":       json.RawMessage(uiSettings),
		"created_at":             createdAt,
		"updated_at":             updatedAt,
	}, nil
}

func (a *AI) projectPayload(ctx context.Context, projectID, companyID uuid.UUID) (map[string]any, error) {
	var projectName, projectSlug, companyName string
	err := a.P.QueryRow(ctx, `
SELECT p.name::text, p.slug::text, c.name::text
  FROM pulse_read.projects_v1 p
  JOIN pulse_read.companies_v1 c ON c.company_id = p.company_id
 WHERE p.project_id = $1
   AND p.company_id = $2`, projectID, companyID).Scan(&projectName, &projectSlug, &companyName)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"project_id":   projectID.String(),
		"company_id":   companyID.String(),
		"name":         projectName,
		"slug":         projectSlug,
		"company_name": companyName,
	}, nil
}

func (a *AI) aiDependencies(ctx context.Context, projectID uuid.UUID) ([]map[string]any, error) {
	rows, err := a.P.Query(ctx, `
SELECT ad.action_dependency_id::text, ad.name::text, ad.status::text,
       t.dependency_type_key::text, t.provider_key::text, COALESCE(ad.metadata_json, '{}'::jsonb)
  FROM pulse_read.action_dependencies_v1 ad
  JOIN pulse_read.action_dependency_types_v1 t ON t.action_dependency_type_id = ad.action_dependency_type_id
 WHERE ad.project_id = $1
   AND (t.dependency_type_key LIKE 'ai.%' OR t.dependency_type_key = 'network.vless_proxy')
   AND ad.status <> 'deleted'
 ORDER BY t.dependency_type_key, ad.name`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, name, status, depType, provider string
		var meta []byte
		if err := rows.Scan(&id, &name, &status, &depType, &provider, &meta); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"action_dependency_id": id,
			"name":                 name,
			"status":               status,
			"dependency_type_key":  depType,
			"provider_key":         provider,
			"metadata_json":        json.RawMessage(meta),
		})
	}
	return out, rows.Err()
}

func hasActiveAIProvider(deps []map[string]any) bool {
	for _, dep := range deps {
		depType, _ := dep["dependency_type_key"].(string)
		status, _ := dep["status"].(string)
		if strings.HasPrefix(depType, "ai.") && status == "active" {
			return true
		}
	}
	return false
}
