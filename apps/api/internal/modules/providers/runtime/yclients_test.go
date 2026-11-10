package runtime

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseYClientsCredentialsRequiresEncryptedProfile(t *testing.T) {
	credentials, err := parseYClientsCredentials(`{
      "partner_token":"partner",
      "login":"operator@example.test",
      "password":"password",
      "receipt_hmac_secret":"receipt-secret",
      "receipt_basic_username":"stage",
      "receipt_basic_password":"stage-password"
    }`)
	if err != nil || credentials.Login != "operator@example.test" {
		t.Fatalf("credentials=%+v err=%v", credentials, err)
	}
	if _, err := parseYClientsCredentials(`{"partner_token":"partner"}`); err == nil {
		t.Fatal("incomplete profile must be rejected")
	}
	if _, err := parseYClientsCredentials(`{"partner_token":"partner","user_token":"user","receipt_hmac_secret":"receipt","receipt_basic_username":"stage"}`); err == nil {
		t.Fatal("incomplete receipt Basic Auth pair must be rejected")
	}
}

func TestParseServiceIDsSupportsJSONAndRejectsInvalidValues(t *testing.T) {
	values, err := parseServiceIDs(`[30,20]`)
	if err != nil || len(values) != 2 || values[0] != 30 || values[1] != 20 {
		t.Fatalf("values=%v err=%v", values, err)
	}
	if _, err := parseServiceIDs(`30,invalid`); err == nil {
		t.Fatal("invalid service ID must be rejected")
	}
}

func TestMetadataInt64PreservesLargeYClientsIDs(t *testing.T) {
	var metadata map[string]any
	if err := json.Unmarshal([]byte(`{"chain_id":1972020,"location_id":2099931,"storage_id":300}`), &metadata); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]int64{
		"chain_id": 1972020, "location_id": 2099931, "storage_id": 300,
	} {
		if got := metadataInt64(metadata, key); got != want {
			t.Fatalf("metadataInt64(%q) = %d, want %d", key, got, want)
		}
	}
	if got := metadataInt64(map[string]any{"id": json.Number("1972020")}, "id"); got != 1972020 {
		t.Fatalf("json.Number ID = %d", got)
	}
	for _, invalid := range []any{0, -1, 1.5, float64(1<<53) + 2} {
		if got := metadataInt64(map[string]any{"id": invalid}, "id"); got != 0 {
			t.Fatalf("invalid ID %v parsed as %d", invalid, got)
		}
	}
}

func TestSendYClientsReceiptSignsAllowlistedHTTPSRequest(t *testing.T) {
	const secret = "receipt-secret"
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		canonical := strings.Join([]string{
			r.Header.Get("X-Pulse-Timestamp"), r.Header.Get("X-Pulse-Nonce"),
			r.Header.Get("X-Pulse-Audience"), string(body),
		}, "\n")
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write([]byte(canonical))
		want := "v1=" + hex.EncodeToString(mac.Sum(nil))
		if got := r.Header.Get("X-Pulse-Signature"); !hmac.Equal([]byte(got), []byte(want)) {
			t.Fatalf("signature mismatch")
		}
		if r.Header.Get("Idempotency-Key") != "siam:certificate:instance:1" {
			t.Fatalf("idempotency key = %q", r.Header.Get("Idempotency-Key"))
		}
		username, password, ok := r.BasicAuth()
		if !ok || username != "stage" || password != "stage-password" {
			t.Fatalf("receipt Basic Auth mismatch")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	runner := &Runner{HTTPClient: server.Client()}
	dep := &dependencyRecord{Metadata: map[string]any{
		"receipt_url": server.URL, "receipt_audience": "siam-stage",
		"receipt_allowed_hosts": []any{"127.0.0.1"},
	}}
	err := runner.sendYClientsReceipt(context.Background(), dep, yclientsCredentials{
		ReceiptHMACSecret: secret, ReceiptBasicUser: "stage", ReceiptBasicPass: "stage-password",
	}, yclientsReceipt{
		OperationKey: "siam:certificate:instance:1", Outcome: "synced",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSendYClientsReceiptRejectsUnlistedHost(t *testing.T) {
	runner := &Runner{HTTPClient: http.DefaultClient}
	dep := &dependencyRecord{Metadata: map[string]any{
		"receipt_url": "https://example.test/receipt", "receipt_audience": "siam-stage",
		"receipt_allowed_hosts": []any{"stage.shop.example.test"},
	}}
	if err := runner.sendYClientsReceipt(context.Background(), dep, yclientsCredentials{ReceiptHMACSecret: "secret"}, yclientsReceipt{OperationKey: "op"}); err == nil {
		t.Fatal("unlisted receipt host must be rejected")
	}
}
