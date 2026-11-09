package redact

import (
	"net/http"
	"testing"
)

func TestJSONRedactsNestedSecretKeysAndPatterns(t *testing.T) {
	input := map[string]any{
		"name": "safe",
		"headers": map[string]any{
			"Authorization": "Bearer abcdefghijklmnopqrstuvwxyz",
			"X-Trace":       "trace",
		},
		"items": []any{
			map[string]any{"api_key": "secret-value"},
			"prefix " + "sk-" + "abcdefghijklmnopqrstuvwxyz" + " suffix",
		},
	}
	out := JSON(input).(map[string]any)
	headers := out["headers"].(map[string]any)
	if headers["Authorization"] != Mask {
		t.Fatalf("authorization not masked: %#v", headers["Authorization"])
	}
	items := out["items"].([]any)
	if items[0].(map[string]any)["api_key"] != Mask {
		t.Fatalf("api_key not masked: %#v", items[0])
	}
	if items[1].(string) != "prefix *** suffix" {
		t.Fatalf("token pattern not masked: %q", items[1])
	}
	if out["name"] != "safe" || headers["X-Trace"] != "trace" {
		t.Fatalf("safe values changed: %#v", out)
	}
}

func TestHeaderRedaction(t *testing.T) {
	headers := http.Header{}
	headers.Set("Cookie", "rp_access=secret")
	headers.Set("X-Request-ID", "req-1")
	headers.Set("X-Api-Key", "key-value")
	out := Header(headers)
	if got := out["Cookie"][0]; got != Mask {
		t.Fatalf("cookie = %q", got)
	}
	if got := out["X-Api-Key"][0]; got != Mask {
		t.Fatalf("api key = %q", got)
	}
	if got := out["X-Request-Id"][0]; got != "req-1" {
		t.Fatalf("request id = %q", got)
	}
}
