package runtime

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestPrepareMaxFileAttachmentUploadsPDFAndReturnsToken(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/certificate.pdf":
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", `attachment; filename="certificate.pdf"`)
			_, _ = io.WriteString(w, "%PDF-1.7 test")
		case "/uploads":
			if r.URL.Query().Get("type") != "file" {
				t.Fatalf("upload type = %q", r.URL.Query().Get("type"))
			}
			if r.Header.Get("Authorization") != "test-token" {
				t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
			}
			_, _ = fmt.Fprintf(w, `{"url":%q}`, server.URL+"/upload")
		case "/upload":
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Fatalf("parse upload: %v", err)
			}
			file, _, err := r.FormFile("data")
			if err != nil {
				t.Fatalf("data part: %v", err)
			}
			defer file.Close()
			body, _ := io.ReadAll(file)
			if !strings.HasPrefix(string(body), "%PDF-") {
				t.Fatalf("uploaded body = %q", body)
			}
			_, _ = io.WriteString(w, `{"token":"uploaded-file-token"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	host := strings.TrimPrefix(server.URL, "http://")
	runner := &Runner{MaxHTTPClient: server.Client()}
	document, err := runner.prepareMaxFileAttachment(
		context.Background(),
		server.URL,
		"test-token",
		server.URL+"/certificate.pdf",
		map[string]any{"document_allowed_hosts": []any{strings.Split(host, ":")[0]}},
	)
	if err != nil {
		t.Fatalf("prepare attachment: %v", err)
	}
	payload := document.Attachment["payload"].(map[string]any)
	if payload["token"] != "uploaded-file-token" {
		t.Fatalf("token = %#v", payload["token"])
	}
}

func TestValidateMaxDocumentURLRejectsUnknownHost(t *testing.T) {
	if _, err := validateMaxDocumentURL("https://example.com/certificate.pdf", []string{"shop.example.test"}, false); err == nil {
		t.Fatal("expected unknown host rejection")
	}
}

func TestDownloadMaxDocumentUsesHostScopedBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "stage" || password != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = io.WriteString(w, "%PDF-1.7 protected")
	}))
	defer server.Close()
	source, err := url.Parse(server.URL + "/certificate.pdf")
	if err != nil {
		t.Fatal(err)
	}
	runner := &Runner{
		MaxHTTPClient: server.Client(),
		DocumentBasicAuth: map[string]DocumentBasicAuth{
			source.Hostname(): {Username: "stage", Password: "secret"},
		},
	}
	body, _, err := runner.downloadMaxDocument(context.Background(), source)
	if err != nil || !strings.HasPrefix(string(body), "%PDF-") {
		t.Fatalf("body = %q, err = %v", body, err)
	}
}

func TestDownloadMaxDocumentDoesNotForwardBasicAuthOnRedirect(t *testing.T) {
	var redirectedAuthorization string
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectedAuthorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = io.WriteString(w, "%PDF-1.7 redirected")
	}))
	defer target.Close()
	sourceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := r.BasicAuth(); !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, target.URL+"/certificate.pdf", http.StatusFound)
	}))
	defer sourceServer.Close()
	source, err := url.Parse(sourceServer.URL + "/certificate.pdf")
	if err != nil {
		t.Fatal(err)
	}
	runner := &Runner{
		MaxHTTPClient: sourceServer.Client(),
		DocumentBasicAuth: map[string]DocumentBasicAuth{
			source.Hostname(): {Username: "stage", Password: "secret"},
		},
	}
	if _, _, err := runner.downloadMaxDocument(context.Background(), source); err != nil {
		t.Fatal(err)
	}
	if redirectedAuthorization != "" {
		t.Fatal("document Basic Auth must not be forwarded on redirect")
	}
}

func TestParseDocumentBasicAuthIsStrictAndNormalizesHosts(t *testing.T) {
	values, err := ParseDocumentBasicAuth(`{"Stage.shop.example.test.":{"username":" stage ","password":"secret"}}`)
	if err != nil || values["stage.shop.example.test"].Username != "stage" {
		t.Fatalf("values = %#v, err = %v", values, err)
	}
	for _, raw := range []string{
		`not-json`,
		`{"stage.shop.example.test:443":{"username":"stage","password":"secret"}}`,
		`{"stage.shop.example.test":{"username":"stage","password":""}}`,
	} {
		if _, err := ParseDocumentBasicAuth(raw); err == nil {
			t.Fatalf("configuration %q must be rejected", raw)
		}
	}
}

func TestMaxAttachmentNotReadyIsRetryableSignal(t *testing.T) {
	if !maxAttachmentNotReady(map[string]any{"raw_response_json": map[string]any{"code": "attachment.not.ready"}}) {
		t.Fatal("expected attachment.not.ready signal")
	}
}

func TestMaxAttachmentProcessingPollIsBoundedAndUsesSameActionAttempt(t *testing.T) {
	provider := map[string]any{"raw_response_json": map[string]any{"code": "attachment.not.ready"}}
	for attempt := 0; attempt < maxAttachmentProcessingPolls-1; attempt++ {
		if !maxAttachmentProcessingRetry(true, attempt, provider) {
			t.Fatalf("attempt %d must poll", attempt)
		}
	}
	if maxAttachmentProcessingRetry(true, maxAttachmentProcessingPolls-1, provider) {
		t.Fatal("polling must stop at the configured bound")
	}
	if maxAttachmentProcessingRetry(false, 0, provider) {
		t.Fatal("messages without a document must not use attachment polling")
	}
}

func TestParseMaxDocumentURLsSupportsBatchAndLegacySingleURL(t *testing.T) {
	batch, err := parseMaxDocumentURLs("", `["https://shop.example.test/a.pdf","https://shop.example.test/b.pdf"]`)
	if err != nil || len(batch) != 2 {
		t.Fatalf("batch = %#v, err = %v", batch, err)
	}
	legacy, err := parseMaxDocumentURLs("https://shop.example.test/legacy.pdf", "")
	if err != nil || len(legacy) != 1 || legacy[0] != "https://shop.example.test/legacy.pdf" {
		t.Fatalf("legacy = %#v, err = %v", legacy, err)
	}
}

func TestBuildMaxMessageBodiesSendsOneDocumentPerMessage(t *testing.T) {
	base := map[string]any{"text": "Расширенная сводка", "format": "html"}
	attachments := []map[string]any{{"type": "file", "id": "one"}, {"type": "file", "id": "two"}, {"type": "file", "id": "three"}}
	messages := buildMaxMessageBodies(base, attachments)
	if len(messages) != 3 {
		t.Fatalf("messages = %d, want 3", len(messages))
	}
	for index, message := range messages {
		files, ok := message["attachments"].([]map[string]any)
		if !ok || len(files) != 1 || files[0]["id"] != attachments[index]["id"] {
			t.Fatalf("message %d attachments = %#v", index, message["attachments"])
		}
	}
	if messages[0]["text"] != "Расширенная сводка" || messages[1]["text"] != "📎 Вложение 2 из 3" || messages[2]["text"] != "📎 Вложение 3 из 3" {
		t.Fatalf("unexpected batch labels: %#v", messages)
	}
	if _, mutated := base["attachments"]; mutated {
		t.Fatal("base message must remain immutable")
	}
}

func TestSanitizeMaxHTTPResponseExtractsMessageBodyMID(t *testing.T) {
	provider := sanitizeMaxHTTPResponse(http.StatusOK, []byte(`{"message":{"body":{"mid":"mid.test-123"}}}`))
	if provider["message_id"] != "mid.test-123" {
		t.Fatalf("message_id = %#v", provider["message_id"])
	}
	if provider["raw_response_json"] == nil {
		t.Fatal("raw MAX response must remain available to success criteria")
	}
}

func TestParseMaxDocumentURLsRejectsInvalidDuplicateAndOversizedBatch(t *testing.T) {
	for _, raw := range []string{
		`not-json`,
		`["https://shop.example.test/a.pdf","https://shop.example.test/a.pdf"]`,
		`["1","2","3","4","5","6","7","8","9"]`,
	} {
		if _, err := parseMaxDocumentURLs("", raw); err == nil {
			t.Fatalf("raw %q must be rejected", raw)
		}
	}
}
