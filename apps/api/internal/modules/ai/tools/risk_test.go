package tools

import "testing"

func TestNormalizeRiskLevelSupportsCanonicalValues(t *testing.T) {
	tests := map[string]RiskLevel{
		"":         RiskRead,
		"read":     RiskRead,
		"medium":   RiskMedium,
		"high":     RiskHigh,
		"critical": RiskCritical,
	}
	for input, want := range tests {
		if got := NormalizeRiskLevel(input); got != want {
			t.Fatalf("NormalizeRiskLevel(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNormalizeRiskLevelMapsLegacyValues(t *testing.T) {
	tests := map[string]RiskLevel{
		"propose":  RiskMedium,
		"proposal": RiskMedium,
		"write":    RiskHigh,
		"unknown":  RiskMedium,
	}
	for input, want := range tests {
		if got := NormalizeRiskLevel(input); got != want {
			t.Fatalf("NormalizeRiskLevel(%q) = %q, want %q", input, got, want)
		}
	}
}
