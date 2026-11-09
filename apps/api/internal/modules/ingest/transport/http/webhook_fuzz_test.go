package http

import (
	"strings"
	"testing"
)

func FuzzParseHookURLPath(f *testing.F) {
	for _, seed := range []string{
		"/hooks/kss/form",
		"hooks/kss/order/extra/path",
		"/hooks//kss///form/",
		"/hooks",
		"/api/v1/hooks/kss/form",
		"",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, urlPath string) {
		if len(urlPath) > 2048 {
			t.Skip("hook path fuzz input is intentionally capped")
		}
		hk, ek, suffix, ok := parseHookURLPath(urlPath)
		if !ok {
			if hk != "" || ek != "" || suffix != "" {
				t.Fatalf("invalid path returned keys: hk=%q ek=%q suffix=%q", hk, ek, suffix)
			}
			return
		}
		if strings.TrimSpace(hk) == "" || strings.TrimSpace(ek) == "" {
			t.Fatalf("valid hook path must return non-empty keys: hk=%q ek=%q path=%q", hk, ek, urlPath)
		}
		if strings.Contains(hk, "/") || strings.Contains(ek, "/") {
			t.Fatalf("hook keys must be single path segments: hk=%q ek=%q", hk, ek)
		}
		if strings.HasPrefix(suffix, "/") || strings.HasSuffix(suffix, "/") {
			t.Fatalf("suffix must be trimmed path segment chain: %q", suffix)
		}
	})
}
