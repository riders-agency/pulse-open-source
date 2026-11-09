package definitionvalidate

import (
	"strings"
	"testing"
)

func TestValidateConfiguredActionInvariants(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr string
	}{
		{
			name:    "http request requires url",
			raw:     `{"nodes":[{"action_key":"call","action_type_key":"http.request","config":{}}]}`,
			wantErr: "укажите URL",
		},
		{
			name:    "email requires dependency",
			raw:     `{"nodes":[{"action_key":"mail","action_type_key":"email.send","config":{"to":"ops@example.test","subject":"Hello"}}]}`,
			wantErr: "выберите интеграцию SMTP",
		},
		{
			name: "configured actions pass",
			raw: `{"nodes":[
				{"action_key":"call","action_type_key":"http.request","config":{"url":"https://example.test/hook"}},
				{"action_key":"mail","action_type_key":"email.send","config":{"smtp_action_dependency_id":"dep_1","to":"ops@example.test","subject":"Hello"}}
			]}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate([]byte(tc.raw))
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate returned error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Validate error = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}

func FuzzValidateWorkflowDefinition(f *testing.F) {
	for _, seed := range []string{
		`{"nodes":[]}`,
		`{"nodes":[{"action_key":"relay","action_type_key":"telegram.relay.trigger","config":{"action_dependency_id":"dep_1"}}]}`,
		`{"nodes":[{"action_key":"timer","action_type_key":"schedule.trigger","config":{"schedule_key":"daily"}}]}`,
		`{"nodes":[{"action_key":"state","action_type_key":"bot.set_state","config":{"state_key":"awaiting_input"}}]}`,
		`{"nodes":[{"action_key":"call","action_type_key":"http.request","config":{"url":"https://example.test"}}]}`,
		`{"nodes":[{"action_key":"call","action_type_key":"http.request","config":{}}]}`,
		``,
		`not-json`,
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 1<<15 {
			t.Skip("workflow validation fuzz input is intentionally capped")
		}
		_ = Validate([]byte(raw))
	})
}
