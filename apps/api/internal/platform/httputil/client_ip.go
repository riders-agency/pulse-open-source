package httputil

import (
	"net"
	"net/http"
	"net/netip"
	"strings"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// TrustedClientIP uses forwarding headers only when the socket peer is trusted.
func TrustedClientIP(config string) func(http.Handler) http.Handler {
	var prefixes []netip.Prefix
	var ranges []string
	for _, value := range strings.Split(config, ",") {
		p, err := netip.ParsePrefix(strings.TrimSpace(value))
		if err == nil && p.Bits() > 0 {
			prefixes = append(prefixes, p)
			ranges = append(ranges, p.String())
		}
	}
	return func(next http.Handler) http.Handler {
		forwarded := chimiddleware.ClientIPFromXFF(ranges...)(next)
		direct := chimiddleware.ClientIPFromRemoteAddr(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}
			peer, err := netip.ParseAddr(host)
			if err == nil {
				for _, prefix := range prefixes {
					if prefix.Contains(peer.Unmap()) {
						forwarded.ServeHTTP(w, r)
						return
					}
				}
			}
			direct.ServeHTTP(w, r)
		})
	}
}
