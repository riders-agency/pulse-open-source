package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	rep "riders.agency/pulse/api/internal/modules/auth/repository"
)

var (
	ErrCannotEndCurrentSession = errors.New("cannot end current session: use logout instead")
	ErrSessionNotFound         = errors.New("session not found or already revoked")
	ErrInvalidRefresh          = errors.New("invalid refresh token")
)

type Auth struct {
	Repo             Repository
	SessionTTL       time.Duration
	RefreshTTL       time.Duration
	BcryptCost       int
	AccessRandBytes  int
	RefreshRandBytes int
	CSRFRandBytes    int
}

type Repository interface {
	CreateUser(ctx context.Context, uid uuid.UUID, email, passHash, name string) error
	UserByEmail(ctx context.Context, email string) (uid uuid.UUID, hash string, err error)
	InsertSession(ctx context.Context, sessID, user uuid.UUID, sessTokHash, refHash string, accessExpiresAt string, expiresAt string, userAgent, ipHost string) error
	RefreshSessionByID(ctx context.Context, sessionID uuid.UUID) (*rep.RefreshSessionRow, error)
	RotateSessionTokens(ctx context.Context, sessionID uuid.UUID, sessionHash, refreshHash, accessExpiresAt, expiresAt string) error
	RevokeSession(ctx context.Context, sid uuid.UUID) error
	RevokeSessionWithReason(ctx context.Context, sid uuid.UUID, reason string) error
	UserProfile(ctx context.Context, userID uuid.UUID) (*rep.UserProfileRow, error)
	WorkspaceContext(ctx context.Context, userID uuid.UUID) (*rep.WorkspaceContextRow, error)
	UpsertWorkspaceContext(ctx context.Context, userID uuid.UUID, companyID, projectID *uuid.UUID) error
	ListActiveSessions(ctx context.Context, userID uuid.UUID) ([]rep.SessionRow, error)
	RevokeSessionIfOwned(ctx context.Context, sessID, ownerID uuid.UUID) (int64, error)
	PatchUserProfile(ctx context.Context, userID uuid.UUID, p rep.ProfilePatch) error
}

func New(r *rep.Pool, sessTTL, refreshTTL time.Duration) *Auth {
	return &Auth{
		Repo:             r,
		SessionTTL:       sessTTL,
		RefreshTTL:       refreshTTL,
		BcryptCost:       bcrypt.DefaultCost,
		AccessRandBytes:  32,
		RefreshRandBytes: 32,
		CSRFRandBytes:    32,
	}
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	CSRFToken        string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

func wireToken(kind string, sess uuid.UUID, secret []byte) string {
	payload := fmt.Sprintf("%s|%s|%s", kind, sess.String(), hex.EncodeToString(secret))
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

func DecodeRefreshToken(wire string) (uuid.UUID, []byte, error) {
	dec, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(wire))
	if err != nil {
		return uuid.Nil, nil, err
	}
	parts := strings.SplitN(string(dec), "|", 3)
	if len(parts) != 3 || parts[0] != "pulse-refresh" {
		return uuid.Nil, nil, ErrInvalidRefresh
	}
	sid, err := uuid.Parse(parts[1])
	if err != nil {
		return uuid.Nil, nil, err
	}
	sec, err := hex.DecodeString(parts[2])
	if err != nil {
		return uuid.Nil, nil, err
	}
	return sid, sec, nil
}

func (a *Auth) Register(ctx context.Context, email, password, name string, userAgent, clientIP string) (*TokenPair, error) {
	ph, err := bcrypt.GenerateFromPassword([]byte(password), a.BcryptCost)
	if err != nil {
		return nil, err
	}
	uid := uuid.New()
	if err := a.Repo.CreateUser(ctx, uid, email, string(ph), name); err != nil {
		return nil, err
	}
	return a.issueSession(ctx, uid, userAgent, clientIP)
}

func (a *Auth) Login(ctx context.Context, email, password string, userAgent, clientIP string) (*TokenPair, error) {
	uid, hash, err := a.Repo.UserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, err
	}
	return a.issueSession(ctx, uid, userAgent, clientIP)
}

func (a *Auth) IssueSessionForUser(ctx context.Context, userID uuid.UUID, userAgent, clientIP string) (*TokenPair, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user id is required")
	}
	return a.issueSession(ctx, userID, userAgent, clientIP)
}

func (a *Auth) Refresh(ctx context.Context, wire string) (*TokenPair, error) {
	sessionID, refreshSecret, err := DecodeRefreshToken(wire)
	if err != nil {
		return nil, ErrInvalidRefresh
	}
	row, err := a.Repo.RefreshSessionByID(ctx, sessionID)
	if err != nil {
		return nil, ErrInvalidRefresh
	}
	if bcrypt.CompareHashAndPassword([]byte(row.Hash), refreshSecret) != nil {
		_ = a.Repo.RevokeSessionWithReason(ctx, sessionID, "refresh_reuse_detected")
		return nil, ErrInvalidRefresh
	}
	pair, err := a.rotateSession(ctx, sessionID)
	if err != nil {
		return nil, ErrInvalidRefresh
	}
	return pair, nil
}

func (a *Auth) RefreshCSRFToken(ctx context.Context, wire string) (string, error) {
	if _, err := a.validateRefreshToken(ctx, wire); err != nil {
		return "", err
	}
	return a.newCSRFToken()
}

func (a *Auth) Logout(ctx context.Context, sess uuid.UUID) error {
	return a.Repo.RevokeSession(ctx, sess)
}

func (a *Auth) Me(ctx context.Context, userID uuid.UUID) (*rep.UserProfileRow, error) {
	return a.Repo.UserProfile(ctx, userID)
}

type WorkspaceContext struct {
	UserID          string `json:"user_id"`
	ActiveCompanyID string `json:"active_company_id"`
	ActiveProjectID string `json:"active_project_id"`
}

type WorkspaceContextUpdate struct {
	ActiveCompanyID *uuid.UUID
	ActiveProjectID *uuid.UUID
}

func (a *Auth) WorkspaceContext(ctx context.Context, userID uuid.UUID) (*WorkspaceContext, error) {
	row, err := a.Repo.WorkspaceContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &WorkspaceContext{
		UserID:          row.UserID.String(),
		ActiveCompanyID: row.ActiveCompanyID,
		ActiveProjectID: row.ActiveProjectID,
	}, nil
}

func (a *Auth) PutWorkspaceContext(ctx context.Context, userID uuid.UUID, update WorkspaceContextUpdate) error {
	return a.Repo.UpsertWorkspaceContext(ctx, userID, update.ActiveCompanyID, update.ActiveProjectID)
}

func (a *Auth) ListSessions(ctx context.Context, userID uuid.UUID) ([]rep.SessionRow, error) {
	return a.Repo.ListActiveSessions(ctx, userID)
}

func (a *Auth) EndOtherSession(ctx context.Context, userID, sessionID, currentSessionID uuid.UUID) error {
	if sessionID == currentSessionID {
		return ErrCannotEndCurrentSession
	}
	n, err := a.Repo.RevokeSessionIfOwned(ctx, sessionID, userID)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrSessionNotFound
	}
	return nil
}

type ProfileUpdate struct {
	DisplayName *string
	FirstName   *string
}

func (a *Auth) PatchProfile(ctx context.Context, userID uuid.UUID, u ProfileUpdate) error {
	p := rep.ProfilePatch{}
	if u.DisplayName != nil {
		s := strings.TrimSpace(*u.DisplayName)
		p.DisplayName = &s
	}
	if u.FirstName != nil {
		s := strings.TrimSpace(*u.FirstName)
		p.FirstName = &s
	}
	return a.Repo.PatchUserProfile(ctx, userID, p)
}

func (a *Auth) issueSession(ctx context.Context, uid uuid.UUID, userAgent, clientIP string) (*TokenPair, error) {
	sessID := uuid.New()
	pair, accessHash, refreshHash, err := a.newTokenPair(sessID)
	if err != nil {
		return nil, err
	}
	if err := a.Repo.InsertSession(ctx, sessID, uid, accessHash, refreshHash,
		pair.AccessExpiresAt.UTC().Format(time.RFC3339Nano),
		pair.RefreshExpiresAt.UTC().Format(time.RFC3339Nano),
		userAgent, clientIP); err != nil {
		return nil, err
	}
	return pair, nil
}

func (a *Auth) rotateSession(ctx context.Context, sessID uuid.UUID) (*TokenPair, error) {
	pair, accessHash, refreshHash, err := a.newTokenPair(sessID)
	if err != nil {
		return nil, err
	}
	if err := a.Repo.RotateSessionTokens(ctx, sessID, accessHash, refreshHash,
		pair.AccessExpiresAt.UTC().Format(time.RFC3339Nano),
		pair.RefreshExpiresAt.UTC().Format(time.RFC3339Nano)); err != nil {
		return nil, err
	}
	return pair, nil
}

func (a *Auth) newTokenPair(sessID uuid.UUID) (*TokenPair, string, string, error) {
	accessSecret := make([]byte, a.AccessRandBytes)
	refreshSecret := make([]byte, a.RefreshRandBytes)
	if _, err := rand.Read(accessSecret); err != nil {
		return nil, "", "", err
	}
	if _, err := rand.Read(refreshSecret); err != nil {
		return nil, "", "", err
	}
	accessHash, err := bcrypt.GenerateFromPassword(accessSecret, a.BcryptCost)
	if err != nil {
		return nil, "", "", err
	}
	refreshHash, err := bcrypt.GenerateFromPassword(refreshSecret, a.BcryptCost)
	if err != nil {
		return nil, "", "", err
	}
	csrfToken, err := a.newCSRFToken()
	if err != nil {
		return nil, "", "", err
	}
	now := time.Now().UTC()
	pair := &TokenPair{
		AccessToken:      wireToken("pulse", sessID, accessSecret),
		RefreshToken:     wireToken("pulse-refresh", sessID, refreshSecret),
		CSRFToken:        csrfToken,
		AccessExpiresAt:  now.Add(a.SessionTTL),
		RefreshExpiresAt: now.Add(a.RefreshTTL),
	}
	return pair, string(accessHash), string(refreshHash), nil
}

func (a *Auth) validateRefreshToken(ctx context.Context, wire string) (*rep.RefreshSessionRow, error) {
	sessionID, refreshSecret, err := DecodeRefreshToken(wire)
	if err != nil {
		return nil, ErrInvalidRefresh
	}
	row, err := a.Repo.RefreshSessionByID(ctx, sessionID)
	if err != nil {
		return nil, ErrInvalidRefresh
	}
	if bcrypt.CompareHashAndPassword([]byte(row.Hash), refreshSecret) != nil {
		return nil, ErrInvalidRefresh
	}
	return row, nil
}

func (a *Auth) newCSRFToken() (string, error) {
	size := a.CSRFRandBytes
	if size < 16 {
		size = 32
	}
	csrfSecret := make([]byte, size)
	if _, err := rand.Read(csrfSecret); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(csrfSecret), nil
}
