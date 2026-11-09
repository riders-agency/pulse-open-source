package fio

import (
	"strings"
	"testing"
)

func FuzzNormalizeFIO(f *testing.F) {
	for _, seed := range []string{
		"",
		"Иван",
		"Петров Иван",
		"Петров Иван Сергеевич",
		"River Stone",
		"  Махмудов   Мехрадж  ",
		"one two three four five",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 1024 {
			t.Skip("fio fuzz input is intentionally capped")
		}
		got := Normalize(raw)
		if strings.TrimSpace(raw) == "" {
			if got.IsValid || got.Mode != "empty" || got.Confidence != "none" {
				t.Fatalf("empty FIO result = %#v", got)
			}
			return
		}
		if !got.IsValid {
			t.Fatalf("non-empty FIO should stay usable, got %#v", got)
		}
		if got.InputName != prepareName(raw) {
			t.Fatalf("InputName = %q, want normalized input %q", got.InputName, prepareName(raw))
		}
		if strings.Contains(got.FullName, "  ") {
			t.Fatalf("FullName contains repeated spaces: %#v", got)
		}
	})
}
