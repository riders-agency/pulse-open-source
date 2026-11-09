package uon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"riders.agency/pulse/api/internal/platform/httptransport"
)

const DefaultAPIBase = "https://api.u-on.ru"

type LeadCreateResult struct {
	HTTP       int
	Raw        string
	ResultCode int
	OK         bool
}

func LeadCreateURL(apiBase, apiKey string) string {
	base := strings.TrimRight(strings.TrimSpace(firstNonEmpty(apiBase, DefaultAPIBase)), "/")
	key := strings.TrimSpace(apiKey)
	return base + "/" + url.PathEscape(key) + "/lead/create.json"
}

func LeadCreate(ctx context.Context, client *http.Client, apiKey, apiBase string, fields map[string]any, timeout time.Duration) (LeadCreateResult, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return LeadCreateResult{}, fmt.Errorf("u-on api_key is required")
	}
	if client == nil {
		client = httptransport.NewDefaultClient(30 * time.Second)
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	post := flattenLeadPost(fields)
	body := encodeLeadPost(post)
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, LeadCreateURL(apiBase, apiKey), strings.NewReader(body))
	if err != nil {
		return LeadCreateResult{}, fmt.Errorf("u-on request build failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := client.Do(req)
	if err != nil {
		return LeadCreateResult{}, fmt.Errorf("u-on request failed: %w", err)
	}
	defer resp.Body.Close()

	rawBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	raw := string(rawBytes)
	out := LeadCreateResult{HTTP: resp.StatusCode, Raw: raw}
	if resp.StatusCode != http.StatusOK {
		return out, fmt.Errorf("u-on returned HTTP %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.Unmarshal(rawBytes, &payload); err != nil {
		return out, fmt.Errorf("u-on response is not valid JSON")
	}
	code := intFromAny(payload["result"])
	out.ResultCode = code
	out.OK = code == http.StatusOK
	if !out.OK {
		return out, fmt.Errorf("u-on result=%d", code)
	}
	return out, nil
}

func BuildSource(prefix, channel string) string {
	prefix = strings.TrimSpace(prefix)
	channel = strings.TrimSpace(channel)
	if prefix == "" {
		prefix = "pulse.riders.agency"
	}
	if channel == "" {
		channel = "обращение"
	}
	return prefix + " / " + channel
}

func HasContact(fields map[string]any) bool {
	for _, key := range []string{"u_phone", "u_phone_mobile", "u_phone_home", "u_email"} {
		if strings.TrimSpace(stringField(fields, key)) != "" {
			if key == "u_email" && !looksLikeEmail(stringField(fields, key)) {
				continue
			}
			return true
		}
	}
	return false
}

func flattenLeadPost(fields map[string]any) map[string]any {
	extended := map[string]string{}
	out := map[string]any{}
	for key, val := range fields {
		sk := strings.TrimSpace(key)
		if sk == "" || val == nil {
			continue
		}
		if sk == "extended_fields" {
			if m, ok := val.(map[string]any); ok {
				for kid, v := range m {
					if s := strings.TrimSpace(fmt.Sprint(v)); s != "" {
						extended[strings.TrimSpace(kid)] = s
					}
				}
			}
			continue
		}
		allowEmpty := sk == "note" || sk == "source" || sk == "requirements_note"
		s := strings.TrimSpace(fmt.Sprint(val))
		if s == "" && !allowEmpty {
			continue
		}
		out[sk] = s
	}
	if len(extended) > 0 {
		out["extended_fields"] = extended
	}
	return out
}

func encodeLeadPost(post map[string]any) string {
	values := url.Values{}
	for key, val := range post {
		switch typed := val.(type) {
		case map[string]string:
			for kid, v := range typed {
				values.Set("extended_fields["+kid+"]", v)
			}
		default:
			values.Set(key, fmt.Sprint(val))
		}
	}
	return values.Encode()
}

func stringField(fields map[string]any, key string) string {
	if fields == nil {
		return ""
	}
	value, ok := fields[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func looksLikeEmail(value string) bool {
	value = strings.TrimSpace(value)
	if !strings.Contains(value, "@") {
		return false
	}
	at := strings.LastIndex(value, "@")
	return at > 0 && at < len(value)-1 && !strings.Contains(value[at+1:], "@")
}

func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
