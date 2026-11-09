package phone

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		mode    string
		isValid bool
	}{
		{"89857254855", "+79857254855", "ru_11", true},
		{"79857254855", "+79857254855", "ru_11", true},
		{"9857254855", "+79857254855", "ru_10", true},
		{"+79857254855", "+79857254855", "ru_11", true},
		{"8 (985) 725-48-55", "+79857254855", "ru_11", true},
		{"+44 20 7946 0958", "+442079460958", "international", true},
		{"", "", "empty", false},
		{"123", "123", "passthrough", false},
	}
	for _, tc := range tests {
		got := Normalize(tc.in)
		if got.PhoneE164 != tc.want {
			t.Fatalf("Normalize(%q).PhoneE164 = %q, want %q", tc.in, got.PhoneE164, tc.want)
		}
		if got.Mode != tc.mode {
			t.Fatalf("Normalize(%q).Mode = %q, want %q", tc.in, got.Mode, tc.mode)
		}
		if got.IsValid != tc.isValid {
			t.Fatalf("Normalize(%q).IsValid = %v, want %v", tc.in, got.IsValid, tc.isValid)
		}
	}
}
