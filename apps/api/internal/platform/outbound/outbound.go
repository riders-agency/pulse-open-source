package outbound

import (
	"errors"
	"os"
	"strings"
)

const EnvKey = "PULSE_OUTBOUND_EFFECTS_MODE"

var ErrDisabled = errors.New("outbound effects are disabled by PULSE_OUTBOUND_EFFECTS_MODE")

func Disabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(EnvKey))) {
	case "disabled", "disable", "off", "blocked", "block", "dry-run", "dryrun":
		return true
	default:
		return false
	}
}
