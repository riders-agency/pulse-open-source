package http

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"riders.agency/pulse/api/internal/platform/problem"
)

func PublicVoiceConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		problem.JSON(w, http.StatusOK, voiceConfigPayload())
	}
}

func OperatorVoiceConfig() http.HandlerFunc {
	return PublicVoiceConfig()
}

func voiceConfigPayload() map[string]any {
	iceServers := configuredICEServers()
	return map[string]any{
		"ice_servers":          iceServers,
		"ice_transport_policy": envDefault("all", "SITECHAT_ICE_TRANSPORT_POLICY", "SITECHAT_WEBRTC_ICE_TRANSPORT_POLICY"),
		"audio": map[string]any{
			"echo_cancellation": true,
			"noise_suppression": true,
			"auto_gain_control": true,
			"channel_count":     1,
		},
	}
}

func configuredICEServers() []map[string]any {
	if raw := strings.TrimSpace(os.Getenv("SITECHAT_ICE_SERVERS_JSON")); raw != "" {
		var out []map[string]any
		if err := json.Unmarshal([]byte(raw), &out); err == nil && len(out) > 0 {
			return normalizeICEServers(out)
		}
	}

	out := []map[string]any{}
	stunURLs := envList("SITECHAT_STUN_URLS")
	if len(stunURLs) == 0 {
		stunURLs = []string{"stun:stun.l.google.com:19302"}
	}
	out = append(out, map[string]any{"urls": stunURLs})

	turnURLs := envList("SITECHAT_TURN_URLS")
	turnUser := envFirst("SITECHAT_TURN_USERNAME", "TURN_USERNAME")
	turnCredential := envFirst("SITECHAT_TURN_CREDENTIAL", "TURN_CREDENTIAL")
	if len(turnURLs) > 0 {
		server := map[string]any{"urls": turnURLs}
		if secret := envFirst("SITECHAT_TURN_AUTH_SECRET", "TURN_AUTH_SECRET"); secret != "" {
			turnUser, turnCredential = turnRESTCredentials(secret)
		}
		if turnUser != "" && turnCredential != "" {
			server["username"] = turnUser
			server["credential"] = turnCredential
		}
		if _, ok := server["credential"]; ok {
			out = append(out, server)
		}
	}
	return out
}

func turnRESTCredentials(secret string) (string, string) {
	ttl := int64(3600)
	if raw := envFirst("SITECHAT_TURN_CREDENTIAL_TTL_SECONDS", "TURN_CREDENTIAL_TTL_SECONDS"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed >= 300 && parsed <= 86400 {
			ttl = parsed
		}
	}
	username := strconv.FormatInt(time.Now().UTC().Add(time.Duration(ttl)*time.Second).Unix(), 10) + ":sitechat"
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write([]byte(username))
	return username, base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func normalizeICEServers(in []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	for _, server := range in {
		urls, ok := server["urls"]
		if !ok {
			continue
		}
		switch value := urls.(type) {
		case string:
			if strings.TrimSpace(value) == "" {
				continue
			}
		case []any:
			if len(value) == 0 {
				continue
			}
		default:
			continue
		}
		out = append(out, server)
	}
	if len(out) == 0 {
		return []map[string]any{{"urls": []string{"stun:stun.l.google.com:19302"}}}
	}
	return out
}

func envList(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func envFirst(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func envDefault(fallback string, keys ...string) string {
	if value := envFirst(keys...); value != "" {
		return value
	}
	return fallback
}
