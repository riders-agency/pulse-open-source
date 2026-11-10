package modelgateway

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"path/filepath"
	"testing"
)

func TestDemoModelIsScopedAndBudgetPersistsInProduction(t *testing.T) {
	t.Setenv("PULSE_PUBLIC_DEMO", "false")
	t.Setenv("PULSE_DEMO_BUDGET_PATH", filepath.Join(t.TempDir(), "quota.json"))
	s, err := NewServer(Config{ProviderKey: "fake", DefaultModelID: "real-config"}, nil, WithDemoAccess(func(ctx context.Context, uid, pid string) (bool, error) { return uid == "guest", nil }, Config{ProviderKey: "fake", DefaultModelID: "demo-config"}))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 31; i++ {
		cfg, _, err := s.providerForActor(context.Background(), "guest", "project", "")
		if i < 30 && (err != nil || cfg.DefaultModelID != "demo-config") {
			t.Fatal("demo provider unavailable", err)
		}
		if i == 30 && status.Code(err) != codes.ResourceExhausted {
			t.Fatal("demo budget bypass", err)
		}
	}
	cfg, _, err := s.providerForActor(context.Background(), "real-user", "project", "")
	if err != nil || cfg.DefaultModelID != "real-config" {
		t.Fatal("real provider changed", err)
	}
}
