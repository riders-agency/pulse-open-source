package http

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	nethttp "net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	cryptoapp "riders.agency/pulse/api/internal/platform/crypto"
	"riders.agency/pulse/api/internal/platform/httputil"
)

const (
	webhookMaxBodyBytes   = 1 << 20
	defaultReplayWindow   = 5 * time.Minute
	webhookRotationWindow = 15 * time.Minute

	headerIdempotencyKey = "Idempotency-Key"
	headerEventID        = "X-Riders-Event-Id"
	headerTimestamp      = "X-Riders-Timestamp"
	headerSignature      = "X-Riders-Signature"
	headerProjectSecret  = "X-Project-Secret"
)

var (
	errWebhookContentType      = errors.New("webhook content-type must be application/json")
	errWebhookBodyTooLarge     = errors.New("webhook body too large")
	errWebhookInvalidJSON      = errors.New("webhook invalid json")
	errWebhookIdempotency      = errors.New("webhook idempotency key required")
	errWebhookTimestamp        = errors.New("webhook timestamp invalid")
	errWebhookCredential       = errors.New("webhook credential invalid")
	errWebhookSignature        = errors.New("webhook signature invalid")
	errWebhookSignatureKey     = errors.New("webhook signature key unavailable")
	errWebhookIdentifierLength = errors.New("webhook identifier too long")
)

type webhookAuthContext struct {
	authType                 string
	projectHookSecretHash    string
	endpointSecretHash       *string
	endpointSecretCiphertext []byte
	previousSecretHash       *string
	previousSecretCiphertext []byte
	previousSecretExpiresAt  *time.Time
	replayWindowSeconds      int
}

type webhookIdentity struct {
	eventID        string
	idempotencyKey string
	timestamp      time.Time
	timestampWire  string
}

type webhookAuthResult struct {
	scheme            string
	signatureVerified bool
}

func readWebhookJSONBody(r *nethttp.Request) ([]byte, any, error) {
	if !webhookContentTypeOK(r.Header.Get("Content-Type")) {
		return nil, nil, errWebhookContentType
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, webhookMaxBodyBytes+1))
	if err != nil {
		return nil, nil, err
	}
	if len(raw) > webhookMaxBodyBytes {
		return nil, nil, errWebhookBodyTooLarge
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var payload any
	if err := dec.Decode(&payload); err != nil {
		return nil, nil, errWebhookInvalidJSON
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return nil, nil, errWebhookInvalidJSON
	}
	return raw, payload, nil
}

func webhookContentTypeOK(value string) bool {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}

func webhookBodySHA256(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func webhookReplayWindow(seconds int) time.Duration {
	if seconds < 30 {
		return defaultReplayWindow
	}
	if seconds > 86400 {
		return 24 * time.Hour
	}
	return time.Duration(seconds) * time.Second
}

func parseWebhookIdentity(r *nethttp.Request, now time.Time, window time.Duration, requireFreshness bool) (webhookIdentity, error) {
	eventID := cleanWebhookIdentifier(r.Header.Get(headerEventID))
	idem := cleanWebhookIdentifier(r.Header.Get(headerIdempotencyKey))
	if idem == "" {
		idem = eventID
	}
	if idem == "" {
		return webhookIdentity{}, errWebhookIdempotency
	}
	if len(idem) > 200 || len(eventID) > 200 {
		return webhookIdentity{}, errWebhookIdentifierLength
	}
	tsWire := strings.TrimSpace(r.Header.Get(headerTimestamp))
	if !requireFreshness {
		return webhookIdentity{eventID: eventID, idempotencyKey: idem, timestampWire: tsWire}, nil
	}
	ts, err := parseWebhookTimestamp(tsWire)
	if err != nil {
		return webhookIdentity{}, errWebhookTimestamp
	}
	if ts.Before(now.Add(-window)) || ts.After(now.Add(window)) {
		return webhookIdentity{}, errWebhookTimestamp
	}
	return webhookIdentity{eventID: eventID, idempotencyKey: idem, timestamp: ts, timestampWire: tsWire}, nil
}

func cleanWebhookIdentifier(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range value {
		if r < 33 || r > 126 {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func parseWebhookTimestamp(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errWebhookTimestamp
	}
	if ts, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return ts, nil
	}
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Time{}, errWebhookTimestamp
	}
	return time.Unix(seconds, 0).UTC(), nil
}

func (ctx webhookAuthContext) verify(r *nethttp.Request, raw []byte, identity webhookIdentity, masterKey string, now time.Time) (webhookAuthResult, error) {
	authType := strings.TrimSpace(ctx.authType)
	if authType == "" {
		authType = "project_secret"
	}
	if authType == "none" {
		return webhookAuthResult{scheme: "none"}, nil
	}
	if authType == "hmac" || strings.TrimSpace(r.Header.Get(headerSignature)) != "" {
		result, err := ctx.verifyHMAC(r, raw, identity, masterKey, now)
		if err == nil {
			return result, nil
		}
		if authType == "hmac" {
			return webhookAuthResult{}, err
		}
	}
	wire := httputil.Bearer(r.Header.Get("Authorization"))
	if wire == "" {
		wire = strings.TrimSpace(r.Header.Get(headerProjectSecret))
	}
	if wire == "" {
		return webhookAuthResult{}, errWebhookCredential
	}
	if ctx.compareCurrentOrPreviousHash(wire, now) {
		return webhookAuthResult{scheme: "secret.v1"}, nil
	}
	return webhookAuthResult{}, errWebhookCredential
}

func (ctx webhookAuthContext) compareCurrentOrPreviousHash(wire string, now time.Time) bool {
	if ctx.endpointSecretHash != nil && strings.TrimSpace(*ctx.endpointSecretHash) != "" {
		if bcrypt.CompareHashAndPassword([]byte(*ctx.endpointSecretHash), []byte(wire)) == nil {
			return true
		}
	} else if strings.TrimSpace(ctx.projectHookSecretHash) != "" {
		if bcrypt.CompareHashAndPassword([]byte(ctx.projectHookSecretHash), []byte(wire)) == nil {
			return true
		}
	}
	if ctx.previousSecretHash == nil || strings.TrimSpace(*ctx.previousSecretHash) == "" {
		return false
	}
	if ctx.previousSecretExpiresAt == nil || !ctx.previousSecretExpiresAt.After(now) {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(*ctx.previousSecretHash), []byte(wire)) == nil
}

func (ctx webhookAuthContext) verifyHMAC(r *nethttp.Request, raw []byte, identity webhookIdentity, masterKey string, now time.Time) (webhookAuthResult, error) {
	sig, err := parseWebhookSignature(r.Header.Get(headerSignature))
	if err != nil {
		return webhookAuthResult{}, err
	}
	secrets := ctx.hmacSecrets(masterKey, now)
	if len(secrets) == 0 {
		return webhookAuthResult{}, errWebhookSignatureKey
	}
	base := webhookSignatureBase(r.Method, r.URL.RequestURI(), identity.timestampWire, identity.eventID, raw)
	for _, secret := range secrets {
		expected := webhookHMAC(secret, base)
		if subtle.ConstantTimeCompare(sig, expected) == 1 {
			return webhookAuthResult{scheme: "hmac-sha256.v1", signatureVerified: true}, nil
		}
	}
	return webhookAuthResult{}, errWebhookSignature
}

func parseWebhookSignature(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "v1=")
	value = strings.TrimPrefix(value, "sha256=")
	if value == "" {
		return nil, errWebhookSignature
	}
	out, err := hex.DecodeString(value)
	if err != nil || len(out) != sha256.Size {
		return nil, errWebhookSignature
	}
	return out, nil
}

func (ctx webhookAuthContext) hmacSecrets(masterKey string, now time.Time) [][]byte {
	out := make([][]byte, 0, 2)
	if len(ctx.endpointSecretCiphertext) > 0 {
		if plain, err := cryptoapp.Open(ctx.endpointSecretCiphertext, masterKey); err == nil && len(plain) > 0 {
			out = append(out, plain)
		}
	}
	if len(ctx.previousSecretCiphertext) > 0 && ctx.previousSecretExpiresAt != nil && ctx.previousSecretExpiresAt.After(now) {
		if plain, err := cryptoapp.Open(ctx.previousSecretCiphertext, masterKey); err == nil && len(plain) > 0 {
			out = append(out, plain)
		}
	}
	return out
}

func webhookSignatureBase(method, requestURI, timestamp, eventID string, raw []byte) []byte {
	var b bytes.Buffer
	b.WriteString("riders-webhook.v1")
	b.WriteByte('\n')
	b.WriteString(strings.ToUpper(strings.TrimSpace(method)))
	b.WriteByte('\n')
	b.WriteString(requestURI)
	b.WriteByte('\n')
	b.WriteString(timestamp)
	b.WriteByte('\n')
	b.WriteString(eventID)
	b.WriteByte('\n')
	b.Write(raw)
	return b.Bytes()
}

func webhookHMAC(secret, base []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(base)
	return mac.Sum(nil)
}
