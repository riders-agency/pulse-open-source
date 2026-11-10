package cryptoapp

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestSealUsesVersionedEnvelopeWithKeyID(t *testing.T) {
	key := hex.EncodeToString(bytes.Repeat([]byte{7}, 32))
	t.Setenv("RIDERS_PULSE_MASTER_KEY_ID", "key-2026-06")
	sealed, err := Seal([]byte("secret"), key)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	keyID, ok := EnvelopeKeyID(sealed)
	if !ok || keyID != "key-2026-06" {
		t.Fatalf("EnvelopeKeyID = %q,%v", keyID, ok)
	}
	plain, err := Open(sealed, key)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if string(plain) != "secret" {
		t.Fatalf("plain = %q", plain)
	}
}

func TestOpenReadsLegacyRawCiphertextAndRewraps(t *testing.T) {
	oldKey := bytes.Repeat([]byte{1}, 32)
	newKey := hex.EncodeToString(bytes.Repeat([]byte{2}, 32))
	legacy, err := sealRaw([]byte("legacy-secret"), oldKey)
	if err != nil {
		t.Fatalf("legacy seal: %v", err)
	}
	rewrapped, err := Rewrap(legacy, hex.EncodeToString(oldKey), newKey, "key-new")
	if err != nil {
		t.Fatalf("rewrap: %v", err)
	}
	keyID, ok := EnvelopeKeyID(rewrapped)
	if !ok || keyID != "key-new" {
		t.Fatalf("rewrapped key id = %q,%v", keyID, ok)
	}
	plain, err := Open(rewrapped, newKey)
	if err != nil {
		t.Fatalf("open rewrapped: %v", err)
	}
	if string(plain) != "legacy-secret" {
		t.Fatalf("plain = %q", plain)
	}
}

func TestDevMarkerRemainsReadable(t *testing.T) {
	sealed, err := Seal([]byte("dev-secret"), "")
	if err != nil {
		t.Fatalf("dev seal: %v", err)
	}
	if _, ok := EnvelopeKeyID(sealed); ok {
		t.Fatal("dev marker must not look like production envelope")
	}
	plain, err := Open(sealed, "")
	if err != nil {
		t.Fatalf("open dev marker: %v", err)
	}
	if string(plain) != "dev-secret" {
		t.Fatalf("plain = %q", plain)
	}
}
