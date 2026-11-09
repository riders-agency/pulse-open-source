package invoicing

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeDraft(t *testing.T) {
	valid := Input{Kind: "bank_invoice", PayerName: " Test ", Purpose: " Service ", DueDate: "2028-02-29"}
	normalized, err := Normalize(valid)
	if err != nil || normalized.PayerName != "Test" || normalized.Purpose != "Service" {
		t.Fatalf("normalize: %+v %v", normalized, err)
	}
	tests := []Input{valid, valid, valid, valid, valid, valid}
	tests[0].DueDate = "2026-02-29"
	tests[1].Purpose = " "
	tests[2].Kind = "unknown"
	tests[3].Revision = -1
	tests[4].TaxID = strings.Repeat("1", 33)
	tests[5].PayerName = strings.Repeat("я", 241)
	for _, in := range tests {
		if _, err := Normalize(in); !errors.Is(err, ErrInvalid) {
			t.Fatalf("accepted invalid: %+v %v", in, err)
		}
	}
	if providers := NewRegistry().Capabilities(); providers == nil || len(providers) != 0 {
		t.Fatalf("default providers: %+v", providers)
	}
}
