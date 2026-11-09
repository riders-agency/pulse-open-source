package redact

import (
	"net/http"
	"regexp"
	"strings"
)

const Mask = "***"

var sensitiveKeys = map[string]struct{}{
	"api_key":       {},
	"apikey":        {},
	"authorization": {},
	"cookie":        {},
	"csrf":          {},
	"password":      {},
	"secret":        {},
	"set_cookie":    {},
	"signature":     {},
	"token":         {},
	"vless_url":     {},
	"x_api_key":     {},
}

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]{12,}`),
	regexp.MustCompile(`rp_(?:live|test)_[A-Za-z0-9._~+/=-]{12,}`),
	regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`github_pat_[A-Za-z0-9_]{40,}`),
	regexp.MustCompile(`ghp_[A-Za-z0-9]{20,}`),
}

func KeySensitive(key string) bool {
	normalized := normalizeKey(key)
	if _, ok := sensitiveKeys[normalized]; ok {
		return true
	}
	for item := range sensitiveKeys {
		if strings.Contains(normalized, item) {
			return true
		}
	}
	return false
}

func String(value string) string {
	out := value
	for _, pattern := range secretPatterns {
		out = pattern.ReplaceAllString(out, Mask)
	}
	return out
}

func JSON(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			if KeySensitive(key) {
				out[key] = Mask
				continue
			}
			out[key] = JSON(item)
		}
		return out
	case []any:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, JSON(item))
		}
		return out
	case string:
		return String(v)
	default:
		return value
	}
}

func Header(src http.Header) map[string][]string {
	out := make(map[string][]string, len(src))
	for key, values := range src {
		if KeySensitive(key) {
			out[key] = []string{Mask}
			continue
		}
		clean := make([]string, 0, len(values))
		for _, value := range values {
			clean = append(clean, String(value))
		}
		out[key] = clean
	}
	return out
}

func normalizeKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.NewReplacer("-", "_", ".", "_", " ", "_").Replace(key)
	for strings.Contains(key, "__") {
		key = strings.ReplaceAll(key, "__", "_")
	}
	return key
}
