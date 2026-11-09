package http

import (
	"net/http"
	"path"
	"strings"
)

// parseHookURLPath extracts hook_key, endpoint_key and optional suffix from the request URL path.
func parseHookURLPath(urlPath string) (hk, ek, suffix string, ok bool) {
	u := path.Clean("/" + strings.TrimSpace(urlPath))
	switch {
	case u == "/hooks":
		return "", "", "", false
	case strings.HasPrefix(u, "/hooks/"):
		tail := strings.Trim(strings.TrimPrefix(u, "/hooks/"), "/")
		if tail == "" {
			return "", "", "", false
		}
		parts := strings.Split(tail, "/")
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			return "", "", "", false
		}
		hk, ek = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if hk == "" || ek == "" {
			return "", "", "", false
		}
		if len(parts) > 2 {
			suffixParts := make([]string, 0, len(parts)-2)
			for _, part := range parts[2:] {
				part = strings.TrimSpace(part)
				if part != "" {
					suffixParts = append(suffixParts, part)
				}
			}
			suffix = strings.Join(suffixParts, "/")
		}
		return hk, ek, suffix, true
	default:
		return "", "", "", false
	}
}

// MountHandler handles webhook ingestion and scoped workflow receipts.
func (h *Webhook) MountHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hk, ek, suffix, ok := parseHookURLPath(r.URL.Path)
		if !ok {
			http.Error(w, "bad hooks path", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPost:
			if _, reserved := parseReceiptSuffix(suffix); reserved {
				w.Header().Set("Allow", http.MethodGet)
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			h.ingestKeys(w, r, hk, ek, suffix)
		case http.MethodGet:
			runID, receipt := parseReceiptSuffix(suffix)
			if !receipt {
				w.Header().Set("Allow", http.MethodPost)
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			h.receiptKeys(w, r, hk, ek, runID)
		default:
			w.Header().Set("Allow", http.MethodPost+", "+http.MethodGet)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func parseReceiptSuffix(suffix string) (string, bool) {
	parts := strings.Split(strings.Trim(strings.TrimSpace(suffix), "/"), "/")
	if len(parts) != 3 || parts[0] != "runs" || parts[1] == "" || parts[2] != "receipt" {
		return "", false
	}
	return parts[1], true
}
