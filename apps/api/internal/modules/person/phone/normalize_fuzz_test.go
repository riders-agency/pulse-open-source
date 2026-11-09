package phone

import (
	"strings"
	"testing"
	"unicode"
)

func FuzzNormalizePhone(f *testing.F) {
	for _, seed := range []string{
		"89857254855",
		"8 (985) 725-48-55",
		"+44 20 7946 0958",
		"",
		"123",
		"+٧٩٨٥٧٢٥٤٨٥٥",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 512 {
			t.Skip("phone fuzz input is intentionally capped")
		}
		got := Normalize(raw)
		if got.InputPhone != strings.TrimSpace(raw) {
			t.Fatalf("InputPhone = %q, want trimmed input %q", got.InputPhone, strings.TrimSpace(raw))
		}
		if got.IsValid {
			if !strings.HasPrefix(got.PhoneE164, "+") {
				t.Fatalf("valid phone must be E.164-like, got %q", got.PhoneE164)
			}
			digits := strings.TrimPrefix(got.PhoneE164, "+")
			runes := []rune(digits)
			if len(runes) < 10 || len(runes) > 15 {
				t.Fatalf("valid phone digit count = %d, phone=%q", len(runes), got.PhoneE164)
			}
			for _, r := range runes {
				if !unicode.IsDigit(r) {
					t.Fatalf("valid phone contains non-digit rune %q in %q", r, got.PhoneE164)
				}
			}
		}
	})
}
