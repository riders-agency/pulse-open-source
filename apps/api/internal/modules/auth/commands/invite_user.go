package commands

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func UpsertInviteUser(ctx context.Context, tx pgx.Tx, email, name string) (uuid.UUID, string, bool, error) {
	var existing uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT user_id FROM users WHERE email IS NOT DISTINCT FROM $1::citext AND deleted_at IS NULL`, email).Scan(&existing); err == nil {
		return existing, "", false, nil
	} else if err != pgx.ErrNoRows {
		return uuid.Nil, "", false, err
	}
	password := randomPassword()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, "", false, err
	}
	if name == "" {
		name = email
	}
	userID := uuid.New()
	_, err = tx.Exec(ctx, `
INSERT INTO users (user_id, email, password_hash, first_name, display_name, is_activated)
VALUES ($1,$2,$3,$4,$4,true)`, userID.String(), email, string(hash), name)
	return userID, password, true, err
}

func randomPassword() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "change-me-" + uuid.NewString()[:8]
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
