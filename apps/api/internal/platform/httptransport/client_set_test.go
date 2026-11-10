package httptransport

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClientWithRootCAsTrustsDedicatedCA(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	pool := x509.NewCertPool()
	if err := AppendPEMCerts(pool, certPEM); err != nil {
		t.Fatalf("append ca bundle: %v", err)
	}
	client := NewClientWithRootCAs(time.Second, pool)
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("client should trust dedicated CA bundle: %v", err)
	}
	resp.Body.Close()
}

func TestAppendPEMCertsRejectsInvalidBundle(t *testing.T) {
	if err := AppendPEMCerts(x509.NewCertPool(), []byte("not a certificate")); err == nil {
		t.Fatal("expected invalid bundle error")
	}
}

func TestSplitPathList(t *testing.T) {
	got := SplitPathList(" /a:/b,/c;/d ")
	want := []string{"/a", "/b", "/c", "/d"}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v want %#v", got, want)
		}
	}
}

func TestSystemPoolFallbackIsUsable(t *testing.T) {
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		t.Skip("system cert pool is unavailable on this host")
	}
	if got := SystemCertPool(); got == nil {
		t.Fatal("system cert pool helper returned nil")
	}
}
