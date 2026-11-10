package http

import (
	"bytes"
	nethttp "net/http"
	"strings"
	"testing"
	"time"

	cryptoapp "riders.agency/pulse/api/internal/platform/crypto"
)

func TestReadWebhookJSONBodyRequiresJSONAndLimit(t *testing.T) {
	req, _ := nethttp.NewRequest(nethttp.MethodPost, "/hooks/h/e", strings.NewReader(`{"ok":true}`))
	req.Header.Set("Content-Type", "text/plain")
	if _, _, err := readWebhookJSONBody(req); err != errWebhookContentType {
		t.Fatalf("expected content-type error, got %v", err)
	}

	req, _ = nethttp.NewRequest(nethttp.MethodPost, "/hooks/h/e", bytes.NewReader(bytes.Repeat([]byte("x"), webhookMaxBodyBytes+1)))
	req.Header.Set("Content-Type", "application/json")
	if _, _, err := readWebhookJSONBody(req); err != errWebhookBodyTooLarge {
		t.Fatalf("expected body-too-large error, got %v", err)
	}

	req, _ = nethttp.NewRequest(nethttp.MethodPost, "/hooks/h/e", strings.NewReader(`{"ok":true}`))
	req.Header.Set("Content-Type", "application/vnd.riders+json")
	raw, payload, err := readWebhookJSONBody(req)
	if err != nil {
		t.Fatalf("read webhook body: %v", err)
	}
	if string(raw) != `{"ok":true}` || payload == nil {
		t.Fatalf("unexpected body parse: raw=%q payload=%#v", raw, payload)
	}
}

func TestParseWebhookIdentityRequiresIdempotencyAndFreshTimestamp(t *testing.T) {
	now := time.Date(2026, 6, 25, 8, 0, 0, 0, time.UTC)
	req, _ := nethttp.NewRequest(nethttp.MethodPost, "/hooks/h/e", nil)
	req.Header.Set(headerTimestamp, now.Format(time.RFC3339Nano))
	if _, err := parseWebhookIdentity(req, now, defaultReplayWindow, true); err != errWebhookIdempotency {
		t.Fatalf("expected idempotency error, got %v", err)
	}

	req.Header.Set(headerEventID, "evt-1")
	req.Header.Set(headerTimestamp, now.Add(-defaultReplayWindow-time.Second).Format(time.RFC3339Nano))
	if _, err := parseWebhookIdentity(req, now, defaultReplayWindow, true); err != errWebhookTimestamp {
		t.Fatalf("expected timestamp error, got %v", err)
	}

	req.Header.Set(headerTimestamp, now.Format(time.RFC3339Nano))
	identity, err := parseWebhookIdentity(req, now, defaultReplayWindow, true)
	if err != nil {
		t.Fatalf("parse identity: %v", err)
	}
	if identity.eventID != "evt-1" || identity.idempotencyKey != "evt-1" {
		t.Fatalf("unexpected identity: %#v", identity)
	}
}

func TestVerifyWebhookHMACUsesExactBody(t *testing.T) {
	now := time.Date(2026, 6, 25, 8, 0, 0, 0, time.UTC)
	secret := []byte("webhook-secret")
	ciphertext, err := cryptoapp.Seal(secret, "")
	if err != nil {
		t.Fatalf("seal secret: %v", err)
	}
	body := []byte(`{"ok":true}`)
	req, _ := nethttp.NewRequest(nethttp.MethodPost, "/hooks/h/e?x=1", bytes.NewReader(body))
	req.Header.Set(headerTimestamp, now.Format(time.RFC3339Nano))
	req.Header.Set(headerEventID, "evt-1")
	identity, err := parseWebhookIdentity(req, now, defaultReplayWindow, true)
	if err != nil {
		t.Fatalf("parse identity: %v", err)
	}
	base := webhookSignatureBase(req.Method, req.URL.RequestURI(), identity.timestampWire, identity.eventID, body)
	req.Header.Set(headerSignature, "v1="+hexString(webhookHMAC(secret, base)))

	ctx := webhookAuthContext{authType: "hmac", endpointSecretCiphertext: ciphertext}
	result, err := ctx.verify(req, body, identity, "", now)
	if err != nil {
		t.Fatalf("verify hmac: %v", err)
	}
	if result.scheme != "hmac-sha256.v1" || !result.signatureVerified {
		t.Fatalf("unexpected auth result: %#v", result)
	}
	if _, err := ctx.verify(req, []byte(`{"ok":false}`), identity, "", now); err != errWebhookSignature {
		t.Fatalf("expected modified body signature failure, got %v", err)
	}
}

func TestVerifyWebhookHMACAcceptsPreviousSecretWithinOverlap(t *testing.T) {
	now := time.Date(2026, 6, 25, 8, 0, 0, 0, time.UTC)
	previous := []byte("previous-secret")
	ciphertext, err := cryptoapp.Seal(previous, "")
	if err != nil {
		t.Fatalf("seal previous secret: %v", err)
	}
	body := []byte(`{"ok":true}`)
	req, _ := nethttp.NewRequest(nethttp.MethodPost, "/hooks/h/e", bytes.NewReader(body))
	req.Header.Set(headerTimestamp, now.Format(time.RFC3339Nano))
	req.Header.Set(headerEventID, "evt-previous")
	identity, err := parseWebhookIdentity(req, now, defaultReplayWindow, true)
	if err != nil {
		t.Fatalf("parse identity: %v", err)
	}
	base := webhookSignatureBase(req.Method, req.URL.RequestURI(), identity.timestampWire, identity.eventID, body)
	req.Header.Set(headerSignature, "v1="+hexString(webhookHMAC(previous, base)))
	expires := now.Add(webhookRotationWindow)
	ctx := webhookAuthContext{
		authType:                 "hmac",
		previousSecretCiphertext: ciphertext,
		previousSecretExpiresAt:  &expires,
	}
	if _, err := ctx.verify(req, body, identity, "", now); err != nil {
		t.Fatalf("expected previous secret accepted during overlap: %v", err)
	}
	expired := now.Add(-time.Second)
	ctx.previousSecretExpiresAt = &expired
	if _, err := ctx.verify(req, body, identity, "", now); err != errWebhookSignatureKey {
		t.Fatalf("expected expired previous secret rejected, got %v", err)
	}
}

func hexString(raw []byte) string {
	const alphabet = "0123456789abcdef"
	out := make([]byte, len(raw)*2)
	for i, b := range raw {
		out[i*2] = alphabet[b>>4]
		out[i*2+1] = alphabet[b&0x0f]
	}
	return string(out)
}
