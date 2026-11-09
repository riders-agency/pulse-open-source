package httptransport

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"riders.agency/pulse/api/internal/platform/demomode"
)

const defaultTimeout = 30 * time.Second

func NewDefaultClient(timeout time.Duration) *http.Client {
	return NewClient(timeout, nil)
}

func NewClient(timeout time.Duration, transport http.RoundTripper) *http.Client {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if transport == nil {
		transport = NewTransport(nil)
	}
	if demomode.Enabled() {
		transport = demoTransport{next: transport}
	}
	return &http.Client{Timeout: timeout, Transport: transport}
}

func NewClientWithRootCAs(timeout time.Duration, rootCAs *x509.CertPool) *http.Client {
	return NewClient(timeout, NewTransport(&tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    rootCAs,
	}))
}

func NewTransport(tlsConfig *tls.Config) *http.Transport {
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.Proxy = http.ProxyFromEnvironment
	base.DialContext = (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext
	base.ForceAttemptHTTP2 = true
	base.MaxIdleConns = 100
	base.MaxIdleConnsPerHost = 20
	base.IdleConnTimeout = 90 * time.Second
	base.TLSHandshakeTimeout = 10 * time.Second
	base.ResponseHeaderTimeout = 30 * time.Second
	base.ExpectContinueTimeout = time.Second
	if tlsConfig != nil {
		base.TLSClientConfig = tlsConfig
	}
	return base
}

func SystemCertPool() *x509.CertPool {
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		return x509.NewCertPool()
	}
	return pool
}

func AppendPEMCerts(pool *x509.CertPool, raw []byte) error {
	if pool == nil {
		return fmt.Errorf("cert pool is nil")
	}
	if ok := pool.AppendCertsFromPEM(raw); !ok {
		return fmt.Errorf("PEM data has no certificates")
	}
	return nil
}

func SplitPathList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ':' || r == ',' || r == ';'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

type demoTransport struct{ next http.RoundTripper }

func (d demoTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" || req.URL.Hostname() != "apinet.cloud" || (req.URL.Port() != "" && req.URL.Port() != "443") {
		return nil, errors.New("external connections are disabled in the public demo")
	}
	return d.next.RoundTrip(req)
}
