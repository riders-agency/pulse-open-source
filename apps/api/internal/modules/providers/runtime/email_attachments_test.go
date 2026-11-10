package runtime

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDownloadEmailAttachmentsAndBuildMultipartMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", `attachment; filename="certificate.pdf"`)
		_, _ = io.WriteString(w, "%PDF-1.7 certificate")
	}))
	defer server.Close()
	host := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")[0]
	runner := &Runner{MaxHTTPClient: server.Client()}
	attachments, err := runner.downloadEmailAttachments(context.Background(), `["`+server.URL+`/certificate.pdf"]`, []string{host})
	if err != nil {
		t.Fatal(err)
	}
	message := string(buildEmailMessageWithAttachments("DemoShop <shop@example.test>", []string{"buyer@example.test"}, "Сертификаты", "Готово", "plain", attachments))
	for _, want := range []string{"multipart/mixed", "application/pdf", "certificate.pdf", "JVBERi0"} {
		if !strings.Contains(message, want) {
			t.Fatalf("message lacks %q: %s", want, message)
		}
	}
}

func TestDownloadEmailAttachmentsRejectsUnknownHostAndCount(t *testing.T) {
	runner := &Runner{}
	if _, err := runner.downloadEmailAttachments(context.Background(), `["https://example.com/file.pdf"]`, nil); err == nil {
		t.Fatal("unknown host must be rejected")
	}
	urls := make([]string, maxEmailDocumentCount+1)
	for index := range urls {
		urls[index] = "https://shop.example.test/file.pdf"
	}
	raw := `["` + strings.Join(urls, `","`) + `"]`
	if _, err := runner.downloadEmailAttachments(context.Background(), raw, nil); err == nil {
		t.Fatal("attachment count limit must be enforced")
	}
}

func TestMaxDocumentMetadataAddsWorkflowAllowlist(t *testing.T) {
	metadata := maxDocumentMetadata(map[string]any{"api_base": "https://platform-api2.max.ru"}, []string{"stage.shop.example.test"})
	hosts := maxDocumentAllowedHosts(metadata)
	if !containsString(hosts, "stage.shop.example.test") {
		t.Fatalf("hosts = %#v", hosts)
	}
}

func TestRequiredDocumentAttachmentsRejectEmptyDelivery(t *testing.T) {
	if err := requireDocumentAttachments("email", true, 0); err == nil {
		t.Fatal("required delivery without PDF must fail")
	}
	if err := requireDocumentAttachments("email", true, 1); err != nil {
		t.Fatalf("ready PDF must pass: %v", err)
	}
	if err := requireDocumentAttachments("email", false, 0); err != nil {
		t.Fatalf("optional delivery must pass: %v", err)
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
