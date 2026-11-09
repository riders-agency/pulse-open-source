package max

import (
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewHTTPClientUsesDedicatedCABundle(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	path := filepath.Join(t.TempDir(), "max-ca.pem")
	if err := os.WriteFile(path, certPEM, 0o600); err != nil {
		t.Fatalf("write ca bundle: %v", err)
	}

	client, err := NewHTTPClient(TransportConfig{Timeout: time.Second, CAFiles: []string{path}})
	if err != nil {
		t.Fatalf("new max client: %v", err)
	}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("MAX client should trust dedicated CA bundle: %v", err)
	}
	resp.Body.Close()
}

func TestNewHTTPClientRejectsInvalidBundle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.pem")
	if err := os.WriteFile(path, []byte("not a certificate"), 0o600); err != nil {
		t.Fatalf("write bad bundle: %v", err)
	}
	if _, err := NewHTTPClient(TransportConfig{CAFiles: []string{path}}); err == nil {
		t.Fatal("expected invalid bundle error")
	}
}

func TestCAFilesInDir(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"root.pem", "sub.crt", "note.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	files, err := caFilesInDir(dir)
	if err != nil {
		t.Fatalf("caFilesInDir: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 cert files, got %#v", files)
	}
}
