package authz

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	dbpkg "riders.agency/pulse/api/internal/platform/db"

	"riders.agency/pulse/api/internal/platform/errs"
)

type ActorType string

const (
	ActorSession ActorType = "session"
	ActorAPIKey  ActorType = "api_key"

	ScopeSystem   = "system"
	ScopeCompany  = "company"
	ScopeProject  = "project"
	ScopeWorkflow = "workflow"
	ScopeClient   = "client"
	ScopeService  = "client_service"
	ScopePayment  = "payment_action"

	PermRead  = "read"
	PermWrite = "write"
	PermAdmin = "admin"
)

type Grant struct {
	ScopeKind   string
	ScopeID     uuid.UUID
	Permissions []string
}

type SessionClaims struct {
	SessionID    uuid.UUID
	UserID       uuid.UUID
	Plaintext    string
	ActorType    ActorType
	APIKeyID     uuid.UUID
	APIKeyPrefix string
	Grants       []Grant
}

type sessKey struct{}

func WithSessionClaims(ctx context.Context, s SessionClaims) context.Context {
	return context.WithValue(ctx, sessKey{}, s)
}

func SessionClaimsFrom(ctx context.Context) (SessionClaims, bool) {
	v, ok := ctx.Value(sessKey{}).(SessionClaims)
	return v, ok
}

func IsAPIKey(ctx context.Context) bool {
	c, ok := SessionClaimsFrom(ctx)
	return ok && c.ActorType == ActorAPIKey
}

func RequireCompany(ctx context.Context, pool dbpkg.Executor, companyID uuid.UUID, permission string) error {
	c, ok := SessionClaimsFrom(ctx)
	if !ok {
		return errs.ErrForbidden
	}
	if c.ActorType != ActorAPIKey {
		return requireSessionCompany(ctx, pool, c.UserID, companyID, permission)
	}
	if hasGrant(c.Grants, ScopeSystem, uuid.Nil, permission) || hasGrant(c.Grants, ScopeCompany, companyID, permission) {
		return nil
	}
	if permission != PermRead {
		return errs.ErrForbidden
	}
	projectIDs, err := projectIDsForCompany(ctx, pool, companyID)
	if err != nil {
		return err
	}
	for _, projectID := range projectIDs {
		if hasGrant(c.Grants, ScopeProject, projectID, permission) {
			return nil
		}
	}
	workflowIDs, err := workflowIDsForCompany(ctx, pool, companyID)
	if err != nil {
		return err
	}
	for _, workflowID := range workflowIDs {
		if hasGrant(c.Grants, ScopeWorkflow, workflowID, permission) {
			return nil
		}
	}
	return errs.ErrForbidden
}

func RequireProject(ctx context.Context, pool dbpkg.Executor, projectID uuid.UUID, permission string) error {
	c, ok := SessionClaimsFrom(ctx)
	if !ok {
		return errs.ErrForbidden
	}
	if c.ActorType != ActorAPIKey {
		return requireSessionProject(ctx, pool, c.UserID, projectID, permission)
	}
	companyID, err := companyIDForProject(ctx, pool, projectID)
	if err != nil {
		return err
	}
	if hasGrant(c.Grants, ScopeSystem, uuid.Nil, permission) ||
		hasGrant(c.Grants, ScopeCompany, companyID, permission) ||
		hasGrant(c.Grants, ScopeProject, projectID, permission) {
		return nil
	}
	if permission != PermRead {
		return errs.ErrForbidden
	}
	workflowIDs, err := workflowIDsForProject(ctx, pool, projectID)
	if err != nil {
		return err
	}
	for _, workflowID := range workflowIDs {
		if hasGrant(c.Grants, ScopeWorkflow, workflowID, permission) {
			return nil
		}
	}
	return errs.ErrForbidden
}

func RequireWorkflow(ctx context.Context, pool dbpkg.Executor, workflowID uuid.UUID, permission string) error {
	c, ok := SessionClaimsFrom(ctx)
	if !ok {
		return errs.ErrForbidden
	}
	if c.ActorType != ActorAPIKey {
		projectID, _, err := projectCompanyForWorkflow(ctx, pool, workflowID)
		if err != nil {
			return err
		}
		return requireSessionProjectProcess(ctx, pool, c.UserID, projectID)
	}
	projectID, companyID, err := projectCompanyForWorkflow(ctx, pool, workflowID)
	if err != nil {
		return err
	}
	if hasGrant(c.Grants, ScopeSystem, uuid.Nil, permission) ||
		hasGrant(c.Grants, ScopeCompany, companyID, permission) ||
		hasGrant(c.Grants, ScopeProject, projectID, permission) ||
		hasGrant(c.Grants, ScopeWorkflow, workflowID, permission) {
		return nil
	}
	return errs.ErrForbidden
}

func RequireProjectProcess(ctx context.Context, pool dbpkg.Executor, projectID uuid.UUID) error {
	c, ok := SessionClaimsFrom(ctx)
	if !ok {
		return errs.ErrForbidden
	}
	if c.ActorType == ActorAPIKey {
		return RequireProject(ctx, pool, projectID, PermWrite)
	}
	return requireSessionProjectProcess(ctx, pool, c.UserID, projectID)
}

func RequireClient(ctx context.Context, pool dbpkg.Executor, clientID uuid.UUID, permission string) error {
	c, ok := SessionClaimsFrom(ctx)
	if !ok {
		return errs.ErrForbidden
	}
	if c.ActorType != ActorAPIKey {
		companyID, err := companyIDForClient(ctx, pool, clientID)
		if err != nil {
			return err
		}
		if err := requireSessionCompany(ctx, pool, c.UserID, companyID, permission); err == nil {
			return nil
		}
		projectIDs, err := projectIDsForClient(ctx, pool, clientID)
		if err != nil {
			return err
		}
		for _, projectID := range projectIDs {
			if err := requireSessionProject(ctx, pool, c.UserID, projectID, permission); err == nil {
				return nil
			}
		}
		return errs.ErrForbidden
	}
	if _, err := companyIDForClient(ctx, pool, clientID); err != nil {
		return err
	}
	if hasGrant(c.Grants, ScopeSystem, uuid.Nil, permission) ||
		hasGrant(c.Grants, ScopeClient, clientID, permission) {
		return nil
	}
	projectIDs, err := projectIDsForClient(ctx, pool, clientID)
	if err != nil {
		return err
	}
	for _, projectID := range projectIDs {
		if hasGrant(c.Grants, ScopeProject, projectID, permission) {
			return nil
		}
	}
	return errs.ErrForbidden
}

func RequireClientService(ctx context.Context, pool dbpkg.Executor, serviceID uuid.UUID, permission string) error {
	c, ok := SessionClaimsFrom(ctx)
	if !ok {
		return errs.ErrForbidden
	}
	if c.ActorType != ActorAPIKey {
		_, projectID, companyID, err := scopeForClientService(ctx, pool, serviceID)
		if err != nil {
			return err
		}
		if err := requireSessionCompany(ctx, pool, c.UserID, companyID, permission); err == nil {
			return nil
		}
		return requireSessionProject(ctx, pool, c.UserID, projectID, permission)
	}
	clientID, projectID, companyID, err := scopeForClientService(ctx, pool, serviceID)
	if err != nil {
		return err
	}
	if hasGrant(c.Grants, ScopeSystem, uuid.Nil, permission) ||
		hasGrant(c.Grants, ScopeCompany, companyID, permission) ||
		hasGrant(c.Grants, ScopeProject, projectID, permission) ||
		hasGrant(c.Grants, ScopeClient, clientID, permission) ||
		hasGrant(c.Grants, ScopeService, serviceID, permission) {
		return nil
	}
	return errs.ErrForbidden
}

func RequirePaymentAction(ctx context.Context, pool dbpkg.Executor, actionID uuid.UUID, permission string) error {
	c, ok := SessionClaimsFrom(ctx)
	if !ok {
		return errs.ErrForbidden
	}
	if c.ActorType != ActorAPIKey {
		_, projectID, _, err := scopeForPaymentAction(ctx, pool, actionID)
		if err != nil {
			return err
		}
		return requireSessionProject(ctx, pool, c.UserID, projectID, permission)
	}
	clientID, projectID, companyID, err := scopeForPaymentAction(ctx, pool, actionID)
	if err != nil {
		return err
	}
	if hasGrant(c.Grants, ScopeSystem, uuid.Nil, permission) ||
		hasGrant(c.Grants, ScopeCompany, companyID, permission) ||
		hasGrant(c.Grants, ScopeProject, projectID, permission) ||
		hasGrant(c.Grants, ScopeClient, clientID, permission) ||
		hasGrant(c.Grants, ScopePayment, actionID, permission) {
		return nil
	}
	return errs.ErrForbidden
}

func RequireRun(ctx context.Context, pool dbpkg.Executor, runID uuid.UUID, permission string) error {
	var workflowID uuid.UUID
	err := pool.QueryRow(ctx, `SELECT workflow_id FROM workflow_runs WHERE workflow_run_id = $1`, runID.String()).Scan(&workflowID)
	if err != nil {
		return err
	}
	return RequireWorkflow(ctx, pool, workflowID, permission)
}

func RequireWebhookEndpoint(ctx context.Context, pool dbpkg.Executor, endpointID uuid.UUID, permission string) error {
	var workflowID uuid.UUID
	err := pool.QueryRow(ctx, `SELECT workflow_id FROM webhook_endpoints WHERE webhook_endpoint_id = $1`, endpointID.String()).Scan(&workflowID)
	if err != nil {
		return err
	}
	return RequireWorkflow(ctx, pool, workflowID, permission)
}

func RequireWebhookEvent(ctx context.Context, pool dbpkg.Executor, eventID uuid.UUID, permission string) error {
	var workflowID uuid.UUID
	err := pool.QueryRow(ctx, `SELECT workflow_id FROM webhook_events WHERE webhook_event_id = $1`, eventID.String()).Scan(&workflowID)
	if err != nil {
		return err
	}
	return RequireWorkflow(ctx, pool, workflowID, permission)
}

func AllowCompanyIDs(ctx context.Context, pool dbpkg.Executor) (all bool, ids []uuid.UUID, err error) {
	c, ok := SessionClaimsFrom(ctx)
	if !ok {
		return false, nil, errs.ErrForbidden
	}
	if c.ActorType != ActorAPIKey {
		ids, err := sessionCompanyIDs(ctx, pool, c.UserID)
		return false, ids, err
	}
	seen := map[uuid.UUID]bool{}
	for _, grant := range c.Grants {
		if !permAllows(grant.Permissions, PermRead) {
			continue
		}
		switch grant.ScopeKind {
		case ScopeSystem:
			return true, nil, nil
		case ScopeCompany:
			seen[grant.ScopeID] = true
		case ScopeProject:
			companyID, e := companyIDForProject(ctx, pool, grant.ScopeID)
			if e != nil {
				if errors.Is(e, pgx.ErrNoRows) {
					continue
				}
				return false, nil, e
			}
			seen[companyID] = true
		case ScopeWorkflow:
			_, companyID, e := projectCompanyForWorkflow(ctx, pool, grant.ScopeID)
			if e != nil {
				if errors.Is(e, pgx.ErrNoRows) {
					continue
				}
				return false, nil, e
			}
			seen[companyID] = true
		case ScopeClient:
			companyID, e := companyIDForClient(ctx, pool, grant.ScopeID)
			if e != nil {
				if errors.Is(e, pgx.ErrNoRows) {
					continue
				}
				return false, nil, e
			}
			seen[companyID] = true
		case ScopeService:
			_, _, companyID, e := scopeForClientService(ctx, pool, grant.ScopeID)
			if e != nil {
				if errors.Is(e, pgx.ErrNoRows) {
					continue
				}
				return false, nil, e
			}
			seen[companyID] = true
		case ScopePayment:
			_, _, companyID, e := scopeForPaymentAction(ctx, pool, grant.ScopeID)
			if e != nil {
				if errors.Is(e, pgx.ErrNoRows) {
					continue
				}
				return false, nil, e
			}
			seen[companyID] = true
		}
	}
	return false, mapKeys(seen), nil
}

func AllowProjectIDs(ctx context.Context, pool dbpkg.Executor, companyID uuid.UUID) (all bool, ids []uuid.UUID, err error) {
	c, ok := SessionClaimsFrom(ctx)
	if !ok {
		return false, nil, errs.ErrForbidden
	}
	if c.ActorType != ActorAPIKey {
		if err := requireSessionCompany(ctx, pool, c.UserID, companyID, PermRead); err != nil {
			return false, nil, err
		}
		ids, err := sessionProjectIDs(ctx, pool, c.UserID, companyID)
		return false, ids, err
	}
	if hasGrant(c.Grants, ScopeSystem, uuid.Nil, PermRead) || hasGrant(c.Grants, ScopeCompany, companyID, PermRead) {
		return true, nil, nil
	}
	seen := map[uuid.UUID]bool{}
	for _, grant := range c.Grants {
		if !permAllows(grant.Permissions, PermRead) {
			continue
		}
		switch grant.ScopeKind {
		case ScopeProject:
			gCompanyID, e := companyIDForProject(ctx, pool, grant.ScopeID)
			if e == nil && gCompanyID == companyID {
				seen[grant.ScopeID] = true
			}
		case ScopeWorkflow:
			projectID, gCompanyID, e := projectCompanyForWorkflow(ctx, pool, grant.ScopeID)
			if e == nil && gCompanyID == companyID {
				seen[projectID] = true
			}
		case ScopeClient:
			projectIDs, e := projectIDsForClient(ctx, pool, grant.ScopeID)
			if e == nil {
				for _, projectID := range projectIDs {
					gCompanyID, e := companyIDForProject(ctx, pool, projectID)
					if e == nil && gCompanyID == companyID {
						seen[projectID] = true
					}
				}
			}
		case ScopeService:
			_, projectID, gCompanyID, e := scopeForClientService(ctx, pool, grant.ScopeID)
			if e == nil && gCompanyID == companyID {
				seen[projectID] = true
			}
		case ScopePayment:
			_, projectID, gCompanyID, e := scopeForPaymentAction(ctx, pool, grant.ScopeID)
			if e == nil && gCompanyID == companyID {
				seen[projectID] = true
			}
		}
	}
	return false, mapKeys(seen), nil
}

func AllowWorkflowIDs(ctx context.Context, pool dbpkg.Executor, projectID uuid.UUID) (all bool, ids []uuid.UUID, err error) {
	c, ok := SessionClaimsFrom(ctx)
	if !ok {
		return false, nil, errs.ErrForbidden
	}
	if c.ActorType != ActorAPIKey {
		if err := requireSessionProjectProcess(ctx, pool, c.UserID, projectID); err != nil {
			return false, nil, err
		}
		return true, nil, nil
	}
	if err := RequireProject(ctx, pool, projectID, PermRead); err == nil {
		if hasGrant(c.Grants, ScopeSystem, uuid.Nil, PermRead) || hasGrant(c.Grants, ScopeProject, projectID, PermRead) {
			return true, nil, nil
		}
		companyID, e := companyIDForProject(ctx, pool, projectID)
		if e != nil {
			return false, nil, e
		}
		if hasGrant(c.Grants, ScopeCompany, companyID, PermRead) {
			return true, nil, nil
		}
	} else {
		return false, nil, err
	}
	seen := map[uuid.UUID]bool{}
	for _, grant := range c.Grants {
		if grant.ScopeKind == ScopeWorkflow && permAllows(grant.Permissions, PermRead) {
			gProjectID, _, e := projectCompanyForWorkflow(ctx, pool, grant.ScopeID)
			if e == nil && gProjectID == projectID {
				seen[grant.ScopeID] = true
			}
		}
	}
	return false, mapKeys(seen), nil
}

func hasGrant(grants []Grant, kind string, id uuid.UUID, permission string) bool {
	for _, grant := range grants {
		if grant.ScopeKind == kind && (kind == ScopeSystem || grant.ScopeID == id) && permAllows(grant.Permissions, permission) {
			return true
		}
	}
	return false
}

func permAllows(perms []string, permission string) bool {
	for _, perm := range perms {
		if perm == PermAdmin || perm == permission || (permission == PermRead && perm == PermWrite) {
			return true
		}
	}
	return false
}

func requireSessionCompany(ctx context.Context, pool dbpkg.Executor, userID, companyID uuid.UUID, permission string) error {
	var role string
	err := pool.QueryRow(ctx, `SELECT role FROM company_memberships WHERE company_id = $1 AND user_id = $2`, companyID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrForbidden
	}
	if err != nil {
		return err
	}
	if !companyRoleAllows(role, permission) {
		return errs.ErrForbidden
	}
	return nil
}

func requireSessionProject(ctx context.Context, pool dbpkg.Executor, userID, projectID uuid.UUID, permission string) error {
	var role string
	err := pool.QueryRow(ctx, `
SELECT pm.role
 FROM project_members pm
 JOIN projects pr ON pr.project_id = pm.project_id
 JOIN company_memberships cm ON cm.company_id = pr.company_id AND cm.user_id = pm.user_id
 WHERE pm.project_id = $1 AND pm.user_id = $2 AND pm.status = 'active'`, projectID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrForbidden
	}
	if err != nil {
		return err
	}
	if !projectRoleAllows(role, permission) {
		return errs.ErrForbidden
	}
	return nil
}

func requireSessionProjectProcess(ctx context.Context, pool dbpkg.Executor, userID, projectID uuid.UUID) error {
	var role string
	err := pool.QueryRow(ctx, `
SELECT pm.role
  FROM project_members pm
  JOIN projects pr ON pr.project_id = pm.project_id
  JOIN company_memberships cm ON cm.company_id = pr.company_id AND cm.user_id = pm.user_id
 WHERE pm.project_id = $1
   AND pm.user_id = $2
   AND pm.status = 'active'`, projectID.String(), userID.String()).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrForbidden
	}
	if err != nil {
		return err
	}
	switch role {
	case "owner", "admin", "operations_manager", "project_manager":
		return nil
	default:
		return errs.ErrForbidden
	}
}

func sessionCompanyIDs(ctx context.Context, pool dbpkg.Executor, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := pool.Query(ctx, `
SELECT company_id
  FROM company_memberships
 WHERE user_id = $1`, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func sessionProjectIDs(ctx context.Context, pool dbpkg.Executor, userID, companyID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := pool.Query(ctx, `
SELECT pm.project_id
  FROM project_members pm
  JOIN projects pr ON pr.project_id = pm.project_id
 WHERE pm.user_id = $1
   AND pr.company_id = $2 AND pm.status = 'active'
 ORDER BY pr.name`, userID.String(), companyID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func companyIDForProject(ctx context.Context, pool dbpkg.Executor, projectID uuid.UUID) (uuid.UUID, error) {
	var companyID uuid.UUID
	err := pool.QueryRow(ctx, `SELECT company_id FROM projects WHERE project_id = $1`, projectID.String()).Scan(&companyID)
	return companyID, err
}

func projectCompanyForWorkflow(ctx context.Context, pool dbpkg.Executor, workflowID uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	var projectID uuid.UUID
	var companyID uuid.UUID
	err := pool.QueryRow(ctx, `
SELECT w.project_id, pr.company_id
  FROM workflows w
  JOIN projects pr ON pr.project_id = w.project_id
 WHERE w.workflow_id = $1`, workflowID.String()).Scan(&projectID, &companyID)
	return projectID, companyID, err
}

func projectIDsForCompany(ctx context.Context, pool dbpkg.Executor, companyID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := pool.Query(ctx, `SELECT project_id FROM projects WHERE company_id = $1`, companyID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func workflowIDsForProject(ctx context.Context, pool dbpkg.Executor, projectID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := pool.Query(ctx, `SELECT workflow_id FROM workflows WHERE project_id = $1`, projectID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func workflowIDsForCompany(ctx context.Context, pool dbpkg.Executor, companyID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := pool.Query(ctx, `
SELECT w.workflow_id
  FROM workflows w
  JOIN projects pr ON pr.project_id = w.project_id
 WHERE pr.company_id = $1`, companyID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func companyIDForClient(ctx context.Context, pool dbpkg.Executor, clientID uuid.UUID) (uuid.UUID, error) {
	var companyID uuid.UUID
	err := pool.QueryRow(ctx, `SELECT company_id FROM clients WHERE client_id = $1`, clientID.String()).Scan(&companyID)
	return companyID, err
}

func projectIDsForClient(ctx context.Context, pool dbpkg.Executor, clientID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := pool.Query(ctx, `SELECT project_id FROM client_projects WHERE client_id = $1`, clientID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func scopeForClientService(ctx context.Context, pool dbpkg.Executor, serviceID uuid.UUID) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	var clientID uuid.UUID
	var projectID uuid.UUID
	var companyID uuid.UUID
	err := pool.QueryRow(ctx, `
SELECT client_id, project_id, company_id
  FROM client_services
 WHERE client_service_id = $1`, serviceID.String()).Scan(&clientID, &projectID, &companyID)
	return clientID, projectID, companyID, err
}

func scopeForPaymentAction(ctx context.Context, pool dbpkg.Executor, actionID uuid.UUID) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	var clientID uuid.UUID
	var projectID uuid.UUID
	var companyID uuid.UUID
	err := pool.QueryRow(ctx, `
SELECT client_id, project_id, company_id
  FROM payment_actions
 WHERE payment_action_id = $1`, actionID.String()).Scan(&clientID, &projectID, &companyID)
	return clientID, projectID, companyID, err
}

func mapKeys(values map[uuid.UUID]bool) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(values))
	for id := range values {
		out = append(out, id)
	}
	return out
}
