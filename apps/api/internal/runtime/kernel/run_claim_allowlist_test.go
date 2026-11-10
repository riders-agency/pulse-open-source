package kernel

import "testing"

func TestWithRunClaimAllowlistNormalizesAndDeduplicatesUUIDs(t *testing.T) {
	const first = "8a8e0da6-a60d-44fd-b3a1-d3a3e21267a3"
	const second = "D52C2A29-1F8B-43AE-B683-10FC8E2A691D"
	kernel := New(nil, nil, WithRunClaimAllowlist(first+", invalid; "+second+" "+first))

	if !kernel.runClaimFilterActive {
		t.Fatal("expected run claim filter to be active")
	}
	if len(kernel.runClaimAllowlist) != 2 {
		t.Fatalf("expected two unique UUIDs, got %v", kernel.runClaimAllowlist)
	}
	if kernel.runClaimAllowlist[0] != first || kernel.runClaimAllowlist[1] != "d52c2a29-1f8b-43ae-b683-10fc8e2a691d" {
		t.Fatalf("unexpected normalized allowlist: %v", kernel.runClaimAllowlist)
	}
}

func TestWithRunClaimAllowlistFailsClosedForInvalidNonEmptyInput(t *testing.T) {
	kernel := New(nil, nil, WithRunClaimAllowlist("not-a-uuid"))

	if !kernel.runClaimFilterActive {
		t.Fatal("expected invalid non-empty allowlist to keep filtering active")
	}
	if len(kernel.runClaimAllowlist) != 0 {
		t.Fatalf("expected no accepted UUIDs, got %v", kernel.runClaimAllowlist)
	}
}

func TestWithRunClaimAllowlistLeavesDefaultWorkerUnfiltered(t *testing.T) {
	kernel := New(nil, nil, WithRunClaimAllowlist("  "))

	if kernel.runClaimFilterActive {
		t.Fatal("expected empty allowlist to preserve default worker behavior")
	}
}
