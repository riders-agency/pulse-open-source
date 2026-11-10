package tools

import "strings"

func NormalizeRiskLevel(value string) RiskLevel {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(RiskRead):
		return RiskRead
	case "propose", "proposal", string(RiskMedium):
		return RiskMedium
	case "write", string(RiskHigh):
		return RiskHigh
	case string(RiskCritical):
		return RiskCritical
	default:
		return RiskMedium
	}
}

func IsKnownRiskLevel(value string) bool {
	switch NormalizeRiskLevel(value) {
	case RiskRead, RiskMedium, RiskHigh, RiskCritical:
		return true
	default:
		return false
	}
}
