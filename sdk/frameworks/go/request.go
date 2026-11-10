package pulse

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ForRequest creates a separate client and propagates the caller's cancellation.
func ForRequest(baseURL string, request *http.Request) (*APIClient, context.Context, error) {
	origin, err := url.Parse(baseURL)
	if err != nil || origin.Host == "" || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" || (origin.Path != "" && origin.Path != "/") {
		return nil, nil, errors.New("Pulse URL must be an origin")
	}
	local := origin.Hostname() == "localhost" || origin.Hostname() == "127.0.0.1" || origin.Hostname() == "::1"
	if origin.Scheme != "https" && !(origin.Scheme == "http" && local) {
		return nil, nil, errors.New("Pulse requires HTTPS outside loopback")
	}
	if request == nil {
		return nil, nil, errors.New("a user request is required")
	}
	parts := strings.Fields(request.Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, nil, errors.New("a user Bearer session is required")
	}
	config := NewConfiguration()
	config.Servers = ServerConfigurations{{URL: strings.TrimRight(baseURL, "/")}}
	config.HTTPClient = &http.Client{Timeout: 60 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return NewAPIClient(config), context.WithValue(request.Context(), ContextAccessToken, parts[1]), nil
}
