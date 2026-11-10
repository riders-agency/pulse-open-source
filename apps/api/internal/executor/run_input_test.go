package executor

import "testing"

func TestEnrichRunInputProjectID(t *testing.T) {
	in := []byte(`{"_pulse":{"source_type":"test","test_mode":"full"}}`)
	out := EnrichRunInputProjectID(in, "11111111-1111-1111-1111-111111111111")
	if projectIDFromRunInput(out) != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected project_id in enriched input, got %s", out)
	}
	unchanged := EnrichRunInputProjectID(out, "22222222-2222-2222-2222-222222222222")
	if projectIDFromRunInput(unchanged) != "11111111-1111-1111-1111-111111111111" {
		t.Fatal("should not overwrite existing project_id")
	}
}
