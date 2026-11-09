// Обновляет encrypted_value секрета, привязанного к action_dependency (stdin = plaintext ).
// DATABASE_URL и RIDERS_PULSE_MASTER_KEY — из окружения. ACTION_DEPENDENCY_ID — UUID.
//
//	go run ./cmd/set-dep-secret < token.txt

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	cryptoapp "riders.agency/pulse/api/internal/platform/crypto"
)

func main() {
	log.SetFlags(0)
	ctx := context.Background()
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	masterHEX := strings.TrimSpace(os.Getenv("RIDERS_PULSE_MASTER_KEY"))
	depID := strings.TrimSpace(os.Getenv("ACTION_DEPENDENCY_ID"))
	if dsn == "" || depID == "" {
		log.Fatal("need DATABASE_URL and ACTION_DEPENDENCY_ID")
	}

	plaintext, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	pt := strings.TrimSpace(string(plaintext))
	if pt == "" {
		log.Fatal("stdin: empty")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	enc, err := cryptoapp.Seal([]byte(pt), masterHEX)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback(ctx)

	var secretID, projectID, createdByUserID string
	err = tx.QueryRow(ctx, `
SELECT COALESCE(ad.secret_id::text, ''), ad.project_id::text, ad.created_by_user_id::text
  FROM action_dependencies ad
 WHERE ad.action_dependency_id = $1
 FOR UPDATE`, depID).Scan(&secretID, &projectID, &createdByUserID)
	if err != nil {
		log.Fatal(err)
	}
	if secretID == "" {
		secretID = uuid.NewString()
		_, err = tx.Exec(ctx, `
INSERT INTO secrets (secret_id, project_id, name, encrypted_value, created_by_user_id)
VALUES ($1, $2, $3, $4, $5)`, secretID, projectID, "dep:"+depID, enc, createdByUserID)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		tag, updateErr := tx.Exec(ctx, `UPDATE secrets SET encrypted_value = $2 WHERE secret_id = $1`,
			secretID, enc)
		if updateErr != nil {
			log.Fatal(updateErr)
		}
		if tag.RowsAffected() != 1 {
			log.Fatalf("secrets update affected %d rows", tag.RowsAffected())
		}
	}

	_, err = tx.Exec(ctx, `
UPDATE action_dependencies SET secret_id = $2, status = 'active', updated_at = now()
 WHERE action_dependency_id = $1`,
		depID, secretID)
	if err != nil {
		log.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("secret updated ok")
}
