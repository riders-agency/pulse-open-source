package runtime

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (r *Runner) sendYClientsReceipt(ctx context.Context, dep *dependencyRecord, credentials yclientsCredentials, receipt yclientsReceipt) error {
	callbackURL := strings.TrimSpace(stringMetadata(dep.Metadata, "receipt_url"))
	audience := strings.TrimSpace(stringMetadata(dep.Metadata, "receipt_audience"))
	secret := credentials.ReceiptHMACSecret
	if callbackURL == "" || audience == "" || strings.TrimSpace(secret) == "" {
		return fmt.Errorf("yclients receipt configuration is incomplete")
	}
	parsed, err := url.Parse(callbackURL)
	if err != nil || parsed.User != nil || parsed.Hostname() == "" || parsed.Scheme != "https" {
		return fmt.Errorf("yclients receipt URL must be an HTTPS origin without user info")
	}
	if !hostAllowed(parsed.Hostname(), metadataStringSlice(dep.Metadata, "receipt_allowed_hosts")) {
		return fmt.Errorf("yclients receipt host is not allowlisted")
	}
	body, err := json.Marshal(receipt)
	if err != nil {
		return err
	}
	timestamp := fmt.Sprintf("%d", time.Now().UTC().Unix())
	nonce, err := receiptNonce()
	if err != nil {
		return err
	}
	canonical := strings.Join([]string{timestamp, nonce, audience, string(body)}, "\n")
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(canonical))
	signature := hex.EncodeToString(mac.Sum(nil))
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, callbackURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Pulse-Timestamp", timestamp)
	request.Header.Set("X-Pulse-Nonce", nonce)
	request.Header.Set("X-Pulse-Audience", audience)
	request.Header.Set("X-Pulse-Signature", "v1="+signature)
	request.Header.Set("Idempotency-Key", receipt.OperationKey)
	if credentials.ReceiptBasicUser != "" {
		request.SetBasicAuth(credentials.ReceiptBasicUser, credentials.ReceiptBasicPass)
	}
	response, err := r.httpClient().Do(request)
	if err != nil {
		return fmt.Errorf("siam receipt request failed: %w", err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 16<<10))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("siam receipt returned HTTP %d", response.StatusCode)
	}
	return nil
}

func hostAllowed(host string, allowed []string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	for _, candidate := range allowed {
		if host == strings.ToLower(strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}

func receiptNonce() (string, error) {
	raw := make([]byte, 18)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
