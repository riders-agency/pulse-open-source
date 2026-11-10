package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	rep "riders.agency/pulse/api/internal/modules/auth/repository"
)

func TestRefreshReuseRevokesSessionFamily(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepository()
	auth := &Auth{
		Repo:             repo,
		SessionTTL:       15 * time.Minute,
		RefreshTTL:       24 * time.Hour,
		BcryptCost:       bcrypt.MinCost,
		AccessRandBytes:  16,
		RefreshRandBytes: 16,
		CSRFRandBytes:    16,
	}
	userID := uuid.New()

	first, err := auth.IssueSessionForUser(ctx, userID, "test-agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	if first.CSRFToken == "" {
		t.Fatal("issued token pair should include csrf token")
	}

	second, err := auth.Refresh(ctx, first.RefreshToken)
	if err != nil {
		t.Fatalf("refresh with current token: %v", err)
	}
	if second.RefreshToken == first.RefreshToken {
		t.Fatal("refresh token should rotate")
	}
	if second.CSRFToken == "" || second.CSRFToken == first.CSRFToken {
		t.Fatal("csrf token should rotate with refresh")
	}
	if repo.revokedReason != "" {
		t.Fatalf("session unexpectedly revoked with reason %q", repo.revokedReason)
	}

	if _, err := auth.Refresh(ctx, first.RefreshToken); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("old refresh reuse error = %v, want ErrInvalidRefresh", err)
	}
	if repo.revokedReason != "refresh_reuse_detected" {
		t.Fatalf("revoked reason = %q, want refresh_reuse_detected", repo.revokedReason)
	}
	if _, err := auth.Refresh(ctx, second.RefreshToken); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("latest refresh after family revoke error = %v, want ErrInvalidRefresh", err)
	}
}

func TestRefreshCSRFTokenValidatesCurrentRefreshWithoutRotation(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepository()
	auth := &Auth{
		Repo:             repo,
		SessionTTL:       15 * time.Minute,
		RefreshTTL:       24 * time.Hour,
		BcryptCost:       bcrypt.MinCost,
		AccessRandBytes:  16,
		RefreshRandBytes: 16,
		CSRFRandBytes:    16,
	}
	first, err := auth.IssueSessionForUser(ctx, uuid.New(), "test-agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	refreshHash := repo.refreshHash

	csrf, err := auth.RefreshCSRFToken(ctx, first.RefreshToken)
	if err != nil {
		t.Fatalf("refresh csrf token: %v", err)
	}
	if csrf == "" || csrf == first.CSRFToken {
		t.Fatalf("csrf token = %q, want a new token", csrf)
	}
	if repo.refreshHash != refreshHash {
		t.Fatal("refresh csrf token must not rotate refresh token")
	}
	if repo.revokedReason != "" {
		t.Fatalf("session unexpectedly revoked with reason %q", repo.revokedReason)
	}
}

type fakeRepository struct {
	sessionID     uuid.UUID
	userID        uuid.UUID
	refreshHash   string
	revokedReason string
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{}
}

func (f *fakeRepository) CreateUser(context.Context, uuid.UUID, string, string, string) error {
	panic("not implemented")
}

func (f *fakeRepository) UserByEmail(context.Context, string) (uuid.UUID, string, error) {
	panic("not implemented")
}

func (f *fakeRepository) InsertSession(_ context.Context, sessID, user uuid.UUID, _, refHash string, _, _, _, _ string) error {
	f.sessionID = sessID
	f.userID = user
	f.refreshHash = refHash
	return nil
}

func (f *fakeRepository) RefreshSessionByID(_ context.Context, sessionID uuid.UUID) (*rep.RefreshSessionRow, error) {
	if sessionID != f.sessionID || f.revokedReason != "" {
		return nil, errors.New("session not found")
	}
	return &rep.RefreshSessionRow{SessionID: f.sessionID, UserID: f.userID, Hash: f.refreshHash}, nil
}

func (f *fakeRepository) RotateSessionTokens(_ context.Context, sessionID uuid.UUID, _, refreshHash, _, _ string) error {
	if sessionID != f.sessionID || f.revokedReason != "" {
		return errors.New("session not found")
	}
	f.refreshHash = refreshHash
	return nil
}

func (f *fakeRepository) RevokeSession(ctx context.Context, sid uuid.UUID) error {
	return f.RevokeSessionWithReason(ctx, sid, "logout")
}

func (f *fakeRepository) RevokeSessionWithReason(_ context.Context, sid uuid.UUID, reason string) error {
	if sid == f.sessionID {
		f.revokedReason = reason
	}
	return nil
}

func (f *fakeRepository) UserProfile(context.Context, uuid.UUID) (*rep.UserProfileRow, error) {
	panic("not implemented")
}

func (f *fakeRepository) WorkspaceContext(context.Context, uuid.UUID) (*rep.WorkspaceContextRow, error) {
	panic("not implemented")
}

func (f *fakeRepository) UpsertWorkspaceContext(context.Context, uuid.UUID, *uuid.UUID, *uuid.UUID) error {
	panic("not implemented")
}

func (f *fakeRepository) ListActiveSessions(context.Context, uuid.UUID) ([]rep.SessionRow, error) {
	panic("not implemented")
}

func (f *fakeRepository) RevokeSessionIfOwned(context.Context, uuid.UUID, uuid.UUID) (int64, error) {
	panic("not implemented")
}

func (f *fakeRepository) PatchUserProfile(context.Context, uuid.UUID, rep.ProfilePatch) error {
	panic("not implemented")
}
