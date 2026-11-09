package repository

import (
	"strings"
	"testing"
	"time"
)

func TestAdvanceDueDateUsesDeterministicBaseDate(t *testing.T) {
	got, err := AdvanceDueDate("2026-06-25", "monthly")
	if err != nil {
		t.Fatalf("AdvanceDueDate returned error: %v", err)
	}
	if got == nil || *got != "2026-07-25" {
		t.Fatalf("AdvanceDueDate = %v, want 2026-07-25", got)
	}
}

func FuzzPaymentDateRules(f *testing.F) {
	for _, seed := range []struct {
		date   string
		period string
	}{
		{"2026-06-25", "weekly"},
		{"25.06.2026", "monthly"},
		{"2026-06-25T18:30:00+03:00", "quarterly"},
		{"2024-02-29", "yearly"},
		{"", "monthly"},
		{"not-a-date", "unknown"},
	} {
		f.Add(seed.date, seed.period)
	}

	f.Fuzz(func(t *testing.T, date, period string) {
		if len(date) > 512 || len(period) > 128 {
			t.Skip("payment date fuzz input is intentionally capped")
		}

		normalized, err := parsePaymentDate(date)
		if err != nil {
			return
		}
		parsed, err := time.Parse("2006-01-02", normalized)
		if err != nil {
			t.Fatalf("parsePaymentDate(%q) returned non-canonical date %q", date, normalized)
		}
		if !parsed.Equal(time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("parsePaymentDate(%q) returned non-UTC date boundary %q", date, normalized)
		}

		end := PaymentPeriodEnd(normalized, period)
		if _, err := time.Parse("2006-01-02", end); err != nil {
			t.Fatalf("PaymentPeriodEnd(%q, %q) = %q, not canonical date", normalized, period, end)
		}
		if strings.TrimSpace(period) == "once" {
			got, err := AdvanceDueDate(normalized, period)
			if err != nil {
				t.Fatalf("AdvanceDueDate once returned error: %v", err)
			}
			if got != nil {
				t.Fatalf("AdvanceDueDate once = %v, want nil", *got)
			}
		}
	})
}
