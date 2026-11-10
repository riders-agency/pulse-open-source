package phone

import (
	"strings"
)

type Result struct {
	PhoneE164  string `json:"phone_e164"`
	PhoneRaw   string `json:"phone_raw"`
	InputPhone string `json:"input_phone"`
	Mode       string `json:"mode"`
	Changed    bool   `json:"changed"`
	IsValid    bool   `json:"is_valid"`
	Warning    string `json:"warning,omitempty"`
}

func digitsOnly(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Normalize formats Russian mobile/landline numbers as +7XXXXXXXXXX (E.164 compact).
func Normalize(raw string) Result {
	input := strings.TrimSpace(raw)
	if input == "" {
		return Result{Mode: "empty", InputPhone: input, IsValid: false, Warning: "empty phone input"}
	}

	digits := digitsOnly(input)
	res := Result{
		PhoneRaw:   input,
		InputPhone: input,
	}

	switch {
	case len(digits) == 11 && (digits[0] == '7' || digits[0] == '8'):
		res.PhoneE164 = "+7" + digits[1:]
		res.Mode = "ru_11"
		res.Changed = res.PhoneE164 != input
		res.IsValid = true
	case len(digits) == 10 && digits[0] == '9':
		res.PhoneE164 = "+7" + digits
		res.Mode = "ru_10"
		res.Changed = true
		res.IsValid = true
	case strings.HasPrefix(input, "+") && len(digits) >= 10 && len(digits) <= 15:
		res.PhoneE164 = "+" + digits
		res.Mode = "international"
		res.Changed = res.PhoneE164 != input
		res.IsValid = true
	default:
		res.PhoneE164 = input
		res.Mode = "passthrough"
		res.Warning = "phone was not normalized"
	}

	return res
}
