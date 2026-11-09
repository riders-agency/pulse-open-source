package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	authrepo "riders.agency/pulse/api/internal/modules/auth/repository"
	authsvc "riders.agency/pulse/api/internal/modules/auth/service"
	"riders.agency/pulse/api/internal/platform/authz"
	"riders.agency/pulse/api/internal/platform/config"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

func main() {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(filepath.Join("..", "..", ".env.local"))

	userEmail := flag.String("user-email", os.Getenv("PULSE_OWNER_EMAIL"), "existing user email to mint a scenario session for")
	userIDRaw := flag.String("user-id", "", "existing user id to mint a scenario session for")
	projectIDRaw := flag.String("project-id", os.Getenv("PULSE_PROJECT_ID"), "project id the scenario runner will target")
	ttl := flag.Duration("ttl", 30*time.Minute, "access and refresh token TTL for the temporary scenario session")
	allowProduction := flag.Bool("allow-production", false, "allow minting when RP_ENV=production")
	flag.Parse()

	cfg := config.Load()
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
	if cfg.DatabaseURL == "" {
		fail("DATABASE_URL is required")
	}
	if strings.EqualFold(cfg.Environment, "production") && !*allowProduction {
		fail("refusing to mint a production scenario session without --allow-production")
	}
	if strings.TrimSpace(*projectIDRaw) == "" {
		fail("--project-id or PULSE_PROJECT_ID is required")
	}
	if strings.TrimSpace(*userIDRaw) == "" && strings.TrimSpace(*userEmail) == "" {
		fail("--user-id or --user-email/PULSE_OWNER_EMAIL is required")
	}
	if *ttl < time.Minute || *ttl > 6*time.Hour {
		fail("--ttl must be between 1m and 6h")
	}

	projectID, err := uuid.Parse(strings.TrimSpace(*projectIDRaw))
	if err != nil || projectID == uuid.Nil {
		fail("bad project id")
	}

	ctx := context.Background()
	pool, err := dbpkg.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		fail("db: %v", err)
	}
	defer pool.Close()

	repo := &authrepo.Pool{P: pool}
	userID, err := resolveUserID(ctx, repo, strings.TrimSpace(*userIDRaw), strings.TrimSpace(*userEmail))
	if err != nil {
		fail("resolve user: %v", err)
	}

	claimsCtx := authz.WithSessionClaims(ctx, authz.SessionClaims{UserID: userID, ActorType: authz.ActorSession})
	if err := authz.RequireProject(claimsCtx, pool, projectID, authz.PermWrite); err != nil {
		fail("user %s cannot access project %s: %v", userID, projectID, err)
	}

	auth := authsvc.New(repo, *ttl, *ttl)
	if err := auth.PutWorkspaceContext(claimsCtx, userID, authsvc.WorkspaceContextUpdate{ActiveProjectID: &projectID}); err != nil {
		fail("workspace context: %v", err)
	}
	pair, err := auth.IssueSessionForUser(ctx, userID, "riders-pulse-scenario-runner", "")
	if err != nil {
		fail("issue session: %v", err)
	}

	cookie := "rp_access=" + pair.AccessToken + "; rp_refresh=" + pair.RefreshToken
	fmt.Printf("export PULSE_PROJECT_ID=%s\n", shellQuote(projectID.String()))
	fmt.Printf("export PULSE_AUTH_COOKIE=%s\n", shellQuote(cookie))
	fmt.Fprintf(os.Stderr, "minted scenario session user_id=%s project_id=%s expires_at=%s\n",
		userID,
		projectID,
		pair.AccessExpiresAt.UTC().Format(time.RFC3339),
	)
}

func resolveUserID(ctx context.Context, repo *authrepo.Pool, userIDRaw, email string) (uuid.UUID, error) {
	if userIDRaw != "" {
		userID, err := uuid.Parse(userIDRaw)
		if err != nil || userID == uuid.Nil {
			return uuid.Nil, fmt.Errorf("bad user id")
		}
		if _, err := repo.UserProfile(ctx, userID); err != nil {
			return uuid.Nil, err
		}
		return userID, nil
	}
	userID, _, err := repo.UserByEmail(ctx, email)
	return userID, err
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
