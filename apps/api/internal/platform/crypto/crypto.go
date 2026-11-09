package cryptoapp

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"
)

const (
	devPlainMarker = "rp-dev-plain:"
	envelopePrefix = "rp:v1:"
	defaultKeyID   = "default"
)

// Seal encrypts with AES-GCM when RIDERS_PULSE_MASTER_KEY is a 32-byte hex key;
// otherwise it keeps the development plaintext marker. Production startup
// rejects the dev path before this package is used.
func Seal(plaintext []byte, masterKeyHex string) ([]byte, error) {
	key, err := hex.DecodeString(masterKeyHex)
	if err != nil || len(key) != 32 {
		dev := append([]byte(devPlainMarker), plaintext...)
		return dev, nil
	}
	sealed, err := sealRaw(plaintext, key)
	if err != nil {
		return nil, err
	}
	keyID := sanitizeKeyID(os.Getenv("RIDERS_PULSE_MASTER_KEY_ID"))
	out := envelopePrefix + keyID + ":" + base64.RawStdEncoding.EncodeToString(sealed)
	return []byte(out), nil
}

func sealRaw(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return append(nonce, gcm.Seal(nil, nonce, plaintext, nil)...), nil
}

// Open decrypts current envelopes, legacy raw AES-GCM blobs, and the development
// marker. The legacy raw path is kept so old rows can be read and rewrapped.
func Open(blob []byte, masterKeyHex string) ([]byte, error) {
	if len(blob) > len(devPlainMarker) && string(blob[:len(devPlainMarker)]) == devPlainMarker {
		return blob[len(devPlainMarker):], nil
	}
	key, err := hex.DecodeString(masterKeyHex)
	if err != nil || len(key) != 32 {
		return nil, errors.New("encrypted secret requires RP_MASTER hex key")
	}
	raw := blob
	if isEnvelope(blob) {
		_, payload, err := parseEnvelope(blob)
		if err != nil {
			return nil, err
		}
		raw = payload
	}
	return openRaw(raw, key)
}

func openRaw(blob, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ns := gcm.NonceSize()
	if len(blob) < ns {
		return nil, errors.New("short blob")
	}
	return gcm.Open(nil, blob[:ns], blob[ns:], nil)
}

func Rewrap(blob []byte, oldMasterKeyHex string, newMasterKeyHex string, newKeyID string) ([]byte, error) {
	plaintext, err := Open(blob, oldMasterKeyHex)
	if err != nil {
		return nil, err
	}
	if newKeyID == "" {
		return Seal(plaintext, newMasterKeyHex)
	}
	old := os.Getenv("RIDERS_PULSE_MASTER_KEY_ID")
	_ = os.Setenv("RIDERS_PULSE_MASTER_KEY_ID", newKeyID)
	defer func() {
		if old == "" {
			_ = os.Unsetenv("RIDERS_PULSE_MASTER_KEY_ID")
			return
		}
		_ = os.Setenv("RIDERS_PULSE_MASTER_KEY_ID", old)
	}()
	return Seal(plaintext, newMasterKeyHex)
}

func EnvelopeKeyID(blob []byte) (string, bool) {
	if !isEnvelope(blob) {
		return "", false
	}
	keyID, _, err := parseEnvelope(blob)
	if err != nil {
		return "", false
	}
	return keyID, true
}

func isEnvelope(blob []byte) bool {
	return strings.HasPrefix(string(blob), envelopePrefix)
}

func parseEnvelope(blob []byte) (string, []byte, error) {
	text := string(blob)
	rest := strings.TrimPrefix(text, envelopePrefix)
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", nil, errors.New("invalid encrypted envelope")
	}
	payload, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, errors.New("invalid encrypted envelope payload")
	}
	return parts[0], payload, nil
}

func sanitizeKeyID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultKeyID
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-' || r == '.':
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return defaultKeyID
	}
	return out
}
