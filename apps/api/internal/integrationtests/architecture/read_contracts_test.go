package architecture

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"riders.agency/pulse/api/internal/platform/dbtest"
)

func TestReadContractsAreNonUpdatableAndExcludeCredentials(t *testing.T) {
	f := dbtest.NewTenant(t)
	rows, err := f.Pool.Query(f.Context, `SELECT table_name,is_updatable FROM information_schema.views WHERE table_schema='pulse_read'`)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for rows.Next() {
		var name, updatable string
		if err := rows.Scan(&name, &updatable); err != nil {
			t.Fatal(err)
		}
		if updatable != "NO" {
			t.Fatalf("updatable projection: %s", name)
		}
		names = append(names, name)
	}
	rows.Close()
	if len(names) != 20 {
		t.Fatalf("got %d contracts", len(names))
	}
	for _, name := range names {
		_, err := f.Pool.Exec(context.Background(), `DELETE FROM pulse_read.`+name+` WHERE false`)
		var pgerr *pgconn.PgError
		if !errors.As(err, &pgerr) || pgerr.Code != "55000" {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	columns, err := f.Pool.Query(f.Context, `SELECT column_name FROM information_schema.columns WHERE table_schema='pulse_read'`)
	if err != nil {
		t.Fatal(err)
	}
	defer columns.Close()
	for columns.Next() {
		var name string
		if err := columns.Scan(&name); err != nil {
			t.Fatal(err)
		}
		for _, secret := range []string{"password", "token", "secret", "encrypted", "country_relays_json"} {
			if strings.Contains(name, secret) {
				t.Fatalf("credential in projection %s", name)
			}
		}
	}
}
