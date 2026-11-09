package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIPIgnoresSpoofedHeadersAndKeepsPeer(t *testing.T) {
	for _, tc := range []struct{ peer, xff, want string }{{"203.0.113.8:1234", "198.51.100.1", "203.0.113.8"}, {"127.0.0.1:1234", "198.51.100.1, 203.0.113.8", "203.0.113.8"}, {"127.0.0.1:1234", "invalid", "127.0.0.1"}} {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = tc.peer
		r.Header.Set("X-Forwarded-For", tc.xff)
		r.Header.Set("True-Client-IP", "192.0.2.1")
		handler := TrustedClientIP("127.0.0.1/32")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := ClientIP(r); got != tc.want {
				t.Errorf("%s: got %s want %s", tc.peer, got, tc.want)
			}
			if r.RemoteAddr != tc.peer {
				t.Fatal("socket peer overwritten")
			}
		}))
		handler.ServeHTTP(httptest.NewRecorder(), r)
	}
}
