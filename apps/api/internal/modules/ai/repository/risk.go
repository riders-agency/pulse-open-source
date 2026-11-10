package repository

import aitools "riders.agency/pulse/api/internal/modules/ai/tools"

func normalizeAIRiskLevel(value string) string {
	return string(aitools.NormalizeRiskLevel(value))
}
