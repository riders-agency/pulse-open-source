package repository

import (
	"errors"
	"testing"
)

func TestValidateClientIdentifiers(t *testing.T) {
	tests := []struct {
		name       string
		clientType string
		inn        string
		kpp        string
		ogrn       string
		ogrnip     string
		wantErr    bool
	}{
		{name: "legal entity ok", clientType: "legal_entity", inn: "7707083893", kpp: "770701001", ogrn: "1027700132195"},
		{name: "legal entity bad inn", clientType: "legal_entity", inn: "7707", wantErr: true},
		{name: "ie ok", clientType: "individual_entrepreneur", inn: "500100732259", ogrnip: "304500116000157"},
		{name: "ie bad ogrnip", clientType: "individual_entrepreneur", inn: "500100732259", ogrnip: "123", wantErr: true},
		{name: "individual ok without ids", clientType: "individual"},
		{name: "individual bad inn", clientType: "individual", inn: "1234567890", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateClientIdentifiers(tt.clientType, tt.inn, tt.kpp, tt.ogrn, tt.ogrnip)
			if tt.wantErr && !errors.Is(err, ErrInvalidIdentifiers) {
				t.Fatalf("expected ErrInvalidIdentifiers, got %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAdvanceDueDate(t *testing.T) {
	tests := []struct {
		name    string
		current string
		period  string
		want    string
		wantNil bool
	}{
		{name: "weekly", current: "2026-06-02", period: "weekly", want: "2026-06-09"},
		{name: "weekly russian date", current: "02.06.2026", period: "weekly", want: "2026-06-09"},
		{name: "monthly", current: "2026-06-02", period: "monthly", want: "2026-07-02"},
		{name: "quarterly", current: "2026-06-02", period: "quarterly", want: "2026-09-02"},
		{name: "semiannual", current: "2026-06-02", period: "semiannual", want: "2026-12-02"},
		{name: "yearly", current: "2026-06-02", period: "yearly", want: "2027-06-02"},
		{name: "once", current: "2026-06-02", period: "once", wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AdvanceDueDate(tt.current, tt.period)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %q", *got)
				}
				return
			}
			if got == nil || *got != tt.want {
				t.Fatalf("got %v, want %q", got, tt.want)
			}
		})
	}
}

func TestPaymentPeriodEnd(t *testing.T) {
	tests := []struct {
		name   string
		start  string
		period string
		want   string
	}{
		{name: "monthly", start: "2026-05-01", period: "monthly", want: "2026-05-31"},
		{name: "monthly russian date", start: "01.05.2026", period: "monthly", want: "2026-05-31"},
		{name: "weekly", start: "2026-06-04", period: "weekly", want: "2026-06-10"},
		{name: "quarterly", start: "2026-04-01", period: "quarterly", want: "2026-06-30"},
		{name: "once", start: "2026-06-04", period: "once", want: "2026-06-04"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PaymentPeriodEnd(tt.start, tt.period); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParsePaymentDate(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "iso", value: "2026-06-08", want: "2026-06-08"},
		{name: "russian", value: "08.06.2026", want: "2026-06-08"},
		{name: "rfc3339", value: "2026-06-08T15:04:05Z", want: "2026-06-08"},
		{name: "bad", value: "08/06/2026", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePaymentDate(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMapDaDataSuggestions(t *testing.T) {
	raw := map[string]any{
		"suggestions": []any{
			map[string]any{
				"value": "ООО Ромашка",
				"data": map[string]any{
					"type": "LEGAL",
					"inn":  "7707083893",
					"kpp":  "770701001",
					"ogrn": "1027700132195",
					"name": map[string]any{
						"short_with_opf": "ООО Ромашка",
						"full_with_opf":  "Общество с ограниченной ответственностью Ромашка",
					},
					"address": map[string]any{"unrestricted_value": "Москва"},
					"state":   map[string]any{"status": "ACTIVE"},
				},
			},
		},
	}
	got := MapDaDataSuggestions(raw)
	if len(got) != 1 {
		t.Fatalf("got %d suggestions", len(got))
	}
	if got[0]["client_type"] != "legal_entity" || got[0]["inn"] != "7707083893" {
		t.Fatalf("unexpected mapping: %#v", got[0])
	}
}
