package repository

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type jsonMap map[string]any

func (m *jsonMap) Scan(src any) error {
	if src == nil {
		*m = map[string]any{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported json source %T", src)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	*m = out
	return nil
}

func jsonBytes(v any) []byte {
	if v == nil {
		return []byte(`{}`)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{}`)
	}
	return b
}

func optJSON(v *map[string]any) any {
	if v == nil {
		return nil
	}
	return jsonBytes(*v)
}

func optString(v *string) any {
	if v == nil {
		return nil
	}
	return strings.TrimSpace(*v)
}

func nullableInt(v int) any {
	if v <= 0 {
		return nil
	}
	return v
}

func trimLimit(v string, max int) string {
	v = strings.TrimSpace(v)
	if len(v) > max {
		return v[:max]
	}
	return v
}

func hashString(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}

func hashIP(v string) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(v))
	if err == nil {
		v = host
	}
	return hashString(v)
}

func randomToken(n int) string {
	if n <= 0 {
		n = 18
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return strings.ReplaceAll(uuid.NewString(), "-", "")
	}
	return hex.EncodeToString(buf)
}

func embedScript(publicKey string) string {
	return `<script async src="https://pulse.riders.agency/site-chat/widget.js" data-site-chat-key="` + publicKey + `"></script>`
}

func normalizePriority(v string) string {
	switch strings.TrimSpace(v) {
	case "low", "high":
		return v
	default:
		return "normal"
	}
}

func explicitSequenceValue(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), n > 0
	case int:
		return int64(n), n > 0
	case int64:
		return n, n > 0
	case json.Number:
		value, err := n.Int64()
		return value, err == nil && value > 0
	default:
		return 0, false
	}
}

func numberValue(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func numberOrZero(v any) float64 {
	n, _ := numberValue(v)
	return n
}

func stringFromMap(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func uuidOrEmpty(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if err == pgx.ErrNoRows {
		return ErrNotFound
	}
	return err
}
