package definitionvalidate

import "testing"

func TestDisabledActionValidation(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		valid bool
	}{
		{`{"nodes":[{"action_key":"mail","action_type_key":"email.send","config":{"disabled":true}}]}`, true},
		{`{"nodes":[{"action_key":"mail","action_type_key":"email.send","config":{"disabled":false}}]}`, false},
		{`{"nodes":[{"action_key":"mail","action_type_key":"email.send","config":{"disabled":"true"}}]}`, false},
		{`{"nodes":[{"action_key":"start","action_type_key":"webhook.trigger","config":{"disabled":true}}]}`, false},
		{`{"nodes":[{"action_key":"start","action_type_key":"task.event_trigger","config":{"disabled":true}}]}`, false},
	} {
		if err := Validate([]byte(tc.raw)); (err == nil) != tc.valid {
			t.Errorf("%s: %v", tc.raw, err)
		}
	}
}
