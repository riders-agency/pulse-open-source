package http

import (
	"strings"
	"testing"
)

func TestVoiceConfigUsesTurnEnv(t *testing.T) {
	t.Setenv("SITECHAT_STUN_URLS", "stun:one.example:19302, stun:two.example:19302")
	t.Setenv("SITECHAT_TURN_URLS", "turn:pulse.example:3478?transport=udp,turn:pulse.example:3478?transport=tcp")
	t.Setenv("SITECHAT_TURN_USERNAME", "u")
	t.Setenv("SITECHAT_TURN_CREDENTIAL", "p")

	cfg := voiceConfigPayload()
	servers, ok := cfg["ice_servers"].([]map[string]any)
	if !ok || len(servers) != 2 {
		t.Fatalf("unexpected ice servers: %#v", cfg["ice_servers"])
	}
	if servers[1]["username"] != "u" || servers[1]["credential"] != "p" {
		t.Fatalf("turn credentials were not applied: %#v", servers[1])
	}
	urls, ok := servers[1]["urls"].([]string)
	if !ok || len(urls) != 2 || urls[0] != "turn:pulse.example:3478?transport=udp" {
		t.Fatalf("turn urls were not applied: %#v", servers[1]["urls"])
	}
}

func TestVoiceConfigFallsBackToStun(t *testing.T) {
	t.Setenv("SITECHAT_ICE_SERVERS_JSON", "")
	t.Setenv("SITECHAT_STUN_URLS", "")
	t.Setenv("SITECHAT_TURN_URLS", "")
	t.Setenv("SITECHAT_TURN_USERNAME", "")
	t.Setenv("SITECHAT_TURN_CREDENTIAL", "")

	cfg := voiceConfigPayload()
	servers, ok := cfg["ice_servers"].([]map[string]any)
	if !ok || len(servers) != 1 {
		t.Fatalf("unexpected fallback servers: %#v", cfg["ice_servers"])
	}
	urls, ok := servers[0]["urls"].([]string)
	if !ok || len(urls) != 1 || urls[0] != "stun:stun.l.google.com:19302" {
		t.Fatalf("unexpected fallback urls: %#v", servers[0]["urls"])
	}
}

func TestVoiceConfigUsesTurnRESTSecret(t *testing.T) {
	t.Setenv("SITECHAT_TURN_URLS", "turn:pulse.example:3478?transport=tcp")
	t.Setenv("SITECHAT_TURN_AUTH_SECRET", "secret")
	t.Setenv("SITECHAT_TURN_USERNAME", "")
	t.Setenv("SITECHAT_TURN_CREDENTIAL", "")

	cfg := voiceConfigPayload()
	servers, ok := cfg["ice_servers"].([]map[string]any)
	if !ok || len(servers) != 2 {
		t.Fatalf("unexpected ice servers: %#v", cfg["ice_servers"])
	}
	username, ok := servers[1]["username"].(string)
	if !ok || !strings.Contains(username, ":sitechat") {
		t.Fatalf("unexpected turn rest username: %#v", servers[1]["username"])
	}
	credential, ok := servers[1]["credential"].(string)
	if !ok || credential == "" {
		t.Fatalf("expected turn rest credential")
	}
}
