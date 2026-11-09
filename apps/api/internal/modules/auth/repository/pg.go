package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	P *pgxpool.Pool
}

type WorkspaceContextRow struct {
	UserID          uuid.UUID
	ActiveCompanyID string
	ActiveProjectID string
}

func (d *Pool) CreateUser(ctx context.Context, uid uuid.UUID, email, passHash, name string) error {
	if name == "" {
		name = email
	}
	_, err := d.P.Exec(ctx, `
INSERT INTO users (user_id, email, password_hash, display_name, first_name, is_activated)
VALUES ($1, $2, $3, $4, $5, true)`, uid.String(), email, passHash, name, name)
	return err
}

func (d *Pool) UserByEmail(ctx context.Context, email string) (uid uuid.UUID, hash string, err error) {
	var id string
	row := d.P.QueryRow(ctx, `
SELECT user_id::text, password_hash::text
  FROM users
 WHERE email IS NOT DISTINCT FROM $1::citext
   AND deleted_at IS NULL`, email)
	err = row.Scan(&id, &hash)
	if err != nil {
		return uuid.Nil, "", err
	}
	uid, err = uuid.Parse(id)
	return uid, hash, err
}

func (d *Pool) InsertSession(ctx context.Context, sessID, user uuid.UUID,
	sessTokHash, refHash string,
	accessExpiresAt string,
	expiresAt string,
	userAgent, ipHost string,
) error {
	_, err := d.P.Exec(ctx, `
INSERT INTO sessions (
  session_id,
  user_id,
  session_token_hash,
  refresh_token_hash,
  client_type,
  access_expires_at,
  expires_at,
  user_agent,
  ip_address)
VALUES ($1, $2, $3, $4, 'web', $5::timestamptz, $6::timestamptz, nullif(trim($7::text), '')::text, cast(nullif(trim($8::text), '') as inet))`,
		sessID.String(), user.String(), sessTokHash, refHash, accessExpiresAt, expiresAt, userAgent, ipHost)
	return err
}

type RefreshSessionRow struct {
	SessionID uuid.UUID
	UserID    uuid.UUID
	Hash      string
}

func (d *Pool) RefreshSessionByID(ctx context.Context, sessionID uuid.UUID) (*RefreshSessionRow, error) {
	var row RefreshSessionRow
	var sessionIDStr, userIDStr string
	err := d.P.QueryRow(ctx, `
SELECT session_id::text, user_id::text, refresh_token_hash::text
  FROM sessions
 WHERE session_id = $1
   AND revoked_at IS NULL
   AND expires_at > now()`, sessionID.String()).Scan(&sessionIDStr, &userIDStr, &row.Hash)
	if err != nil {
		return nil, err
	}
	var parseErr error
	row.SessionID, parseErr = uuid.Parse(sessionIDStr)
	if parseErr != nil {
		return nil, parseErr
	}
	row.UserID, parseErr = uuid.Parse(userIDStr)
	if parseErr != nil {
		return nil, parseErr
	}
	return &row, nil
}

func (d *Pool) RotateSessionTokens(ctx context.Context, sessionID uuid.UUID, sessionHash, refreshHash, accessExpiresAt, expiresAt string) error {
	tag, err := d.P.Exec(ctx, `
UPDATE sessions
   SET session_token_hash = $2,
       refresh_token_hash = $3,
       access_expires_at = $4::timestamptz,
       expires_at = $5::timestamptz,
       last_seen_at = now()
 WHERE session_id = $1
   AND revoked_at IS NULL`,
		sessionID.String(), sessionHash, refreshHash, accessExpiresAt, expiresAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// SessionRow is a non-secret view of a logged-in device session.
type SessionRow struct {
	SessionID uuid.UUID
	CreatedAt time.Time
	ExpiresAt time.Time
	UserAgent *string
	IP        *string
}

func (d *Pool) ListActiveSessions(ctx context.Context, userID uuid.UUID) ([]SessionRow, error) {
	rows, err := d.P.Query(ctx, `
SELECT session_id,
       created_at,
       expires_at,
       user_agent,
       ip_address::text
  FROM sessions
 WHERE user_id = $1
   AND revoked_at IS NULL
   AND expires_at > now()
 ORDER BY created_at DESC`, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SessionRow, 0)
	for rows.Next() {
		var r SessionRow
		var ua, ip *string
		if err := rows.Scan(&r.SessionID, &r.CreatedAt, &r.ExpiresAt, &ua, &ip); err != nil {
			return nil, err
		}
		r.UserAgent = ua
		r.IP = ip
		out = append(out, r)
	}
	return out, rows.Err()
}

// RevokeSessionIfOwned marks a session revoked only if it belongs to userID and is active.
func (d *Pool) RevokeSessionIfOwned(ctx context.Context, sessID, ownerID uuid.UUID) (int64, error) {
	tag, err := d.P.Exec(ctx, `
UPDATE sessions
   SET revoked_at = now(),
       revoked_reason = 'user_revoked'
 WHERE session_id = $1
   AND user_id = $2
   AND revoked_at IS NULL`, sessID.String(), ownerID.String())
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

type UserProfileRow struct {
	UserID      uuid.UUID
	Email       string
	DisplayName string
	FirstName   string
}

func (d *Pool) UserProfile(ctx context.Context, userID uuid.UUID) (*UserProfileRow, error) {
	row := d.P.QueryRow(ctx, `
SELECT user_id::text,
       coalesce(email::text, ''),
       coalesce(display_name::text, ''),
       coalesce(first_name::text, '')
  FROM users
 WHERE deleted_at IS NULL
   AND user_id = $1`, userID.String())
	var uidStr string
	var p UserProfileRow
	err := row.Scan(&uidStr, &p.Email, &p.DisplayName, &p.FirstName)
	if err != nil {
		return nil, err
	}
	p.UserID, err = uuid.Parse(uidStr)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (d *Pool) WorkspaceContext(ctx context.Context, userID uuid.UUID) (*WorkspaceContextRow, error) {
	var row WorkspaceContextRow
	var userIDText string
	err := d.P.QueryRow(ctx, `
SELECT u.user_id::text,
       COALESCE(w.active_company_id::text, u.current_company_id::text, '') AS active_company_id,
       COALESCE(w.active_project_id::text, '') AS active_project_id
  FROM users u
  LEFT JOIN user_workspace_contexts w ON w.user_id = u.user_id
 WHERE u.user_id = $1
   AND u.deleted_at IS NULL`, userID.String()).Scan(&userIDText, &row.ActiveCompanyID, &row.ActiveProjectID)
	if err != nil {
		return nil, err
	}
	parsed, err := uuid.Parse(userIDText)
	if err != nil {
		return nil, err
	}
	row.UserID = parsed
	return &row, nil
}

func (d *Pool) UpsertWorkspaceContext(ctx context.Context, userID uuid.UUID, companyID, projectID *uuid.UUID) error {
	var companyValue any
	if companyID != nil && *companyID != uuid.Nil {
		var one int
		if err := d.P.QueryRow(ctx, `
SELECT 1
  FROM pulse_read.company_memberships_v1
 WHERE user_id = $1
   AND company_id = $2`, userID.String(), companyID.String()).Scan(&one); err != nil {
			if err == pgx.ErrNoRows {
				return pgx.ErrNoRows
			}
			return err
		}
		companyValue = companyID.String()
	}

	var projectValue any
	if projectID != nil && *projectID != uuid.Nil {
		var projectCompanyID string
		if err := d.P.QueryRow(ctx, `
SELECT p.company_id::text
  FROM pulse_read.projects_v1 p
  JOIN pulse_read.company_memberships_v1 m ON m.company_id = p.company_id AND m.user_id = $2
 WHERE p.project_id = $1`, projectID.String(), userID.String()).Scan(&projectCompanyID); err != nil {
			if err == pgx.ErrNoRows {
				return pgx.ErrNoRows
			}
			return err
		}
		if companyID != nil && *companyID != uuid.Nil && projectCompanyID != companyID.String() {
			return pgx.ErrNoRows
		}
		if companyValue == nil {
			companyValue = projectCompanyID
		}
		projectValue = projectID.String()
	}

	_, err := d.P.Exec(ctx, `
INSERT INTO user_workspace_contexts (user_id, active_company_id, active_project_id)
VALUES ($1, $2::uuid, $3::uuid)
ON CONFLICT (user_id) DO UPDATE SET
  active_company_id = EXCLUDED.active_company_id,
  active_project_id = EXCLUDED.active_project_id,
  updated_at = now()`, userID.String(), companyValue, projectValue)
	return err
}

type ProfilePatch struct {
	DisplayName *string
	FirstName   *string
}

func (d *Pool) PatchUserProfile(ctx context.Context, userID uuid.UUID, p ProfilePatch) error {
	switch {
	case p.DisplayName != nil && p.FirstName != nil:
		_, err := d.P.Exec(ctx, `
UPDATE users
   SET display_name = $2,
       first_name = $3
 WHERE user_id = $1
   AND deleted_at IS NULL`, userID.String(), *p.DisplayName, *p.FirstName)
		return err
	case p.DisplayName != nil:
		_, err := d.P.Exec(ctx, `
UPDATE users
   SET display_name = $2
 WHERE user_id = $1
   AND deleted_at IS NULL`, userID.String(), *p.DisplayName)
		return err
	case p.FirstName != nil:
		_, err := d.P.Exec(ctx, `
UPDATE users
   SET first_name = $2
 WHERE user_id = $1
   AND deleted_at IS NULL`, userID.String(), *p.FirstName)
		return err
	default:
		return nil
	}
}

func (d *Pool) RevokeSession(ctx context.Context, sid uuid.UUID) error {
	return d.RevokeSessionWithReason(ctx, sid, "logout")
}

func (d *Pool) RevokeSessionWithReason(ctx context.Context, sid uuid.UUID, reason string) error {
	if reason == "" {
		reason = "revoked"
	}
	tag, err := d.P.Exec(ctx, `
UPDATE sessions
   SET revoked_at = now(),
       revoked_reason = $2
 WHERE session_id = $1
   AND revoked_at IS NULL`, sid.String(), reason)
	if err != nil {
		return err
	}
	_ = tag.RowsAffected()
	return nil
}
