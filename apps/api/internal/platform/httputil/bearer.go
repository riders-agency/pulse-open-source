package httputil

import "strings"

func Bearer(hdr string) string {
	const p = "Bearer "
	if len(hdr) > len(p) && strings.EqualFold(hdr[:len(p)], p) {
		return strings.TrimSpace(hdr[len(p):])
	}
	return ""
}
