package httptransport

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublicDemoRejectsOutboundBeforeNetwork(t *testing.T) {
	t.Setenv("PULSE_PUBLIC_DEMO", "true")
	contacted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { contacted = true; w.WriteHeader(204) }))
	defer server.Close()
	client := NewDefaultClient(time.Second)
	for _, url := range []string{server.URL, "https://example.test", "https://apinet.cloud.example.test", "http://apinet.cloud", "https://apinet.cloud:8443"} {
		if _, err := client.Get(url); err == nil {
			t.Fatalf("unexpected allowed URL: %s", url)
		}
	}
	if contacted {
		t.Fatal("request reached denied server")
	}
}
