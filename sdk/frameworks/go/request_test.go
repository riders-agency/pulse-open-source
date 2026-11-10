package pulse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestForRequestKeepsIdentityAndCancellationSeparate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") == "Bearer alice" {
			_, _ = w.Write([]byte(`{"user_id":"alice"}`))
		} else {
			_, _ = w.Write([]byte(`{"user_id":"bob"}`))
		}
	}))
	defer server.Close()
	for _, token := range []string{"alice", "bob"} {
		request := httptest.NewRequest("GET", "/", nil)
		request.Header.Set("Authorization", "Bearer "+token)
		client, ctx, err := ForRequest(server.URL, request)
		if err != nil {
			t.Fatal(err)
		}
		identity, _, err := client.AuthAPI.GetIdentity(ctx).Execute()
		if err != nil || identity.UserId != token {
			t.Fatalf("identity mismatch for %s", token)
		}
	}
	request := httptest.NewRequest("GET", "/", nil)
	if _, _, err := ForRequest(server.URL, request); err == nil {
		t.Fatal("anonymous request accepted")
	}
	request.Header.Set("Authorization", "Bearer alice")
	ctx, cancel := context.WithCancel(request.Context())
	cancel()
	_, inherited, err := ForRequest(server.URL, request.WithContext(ctx))
	if err != nil || inherited.Err() != context.Canceled {
		t.Fatal("cancellation lost")
	}
}
