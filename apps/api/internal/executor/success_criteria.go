package executor

import (
	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
)

// CriterionResult is stored under output_json["criteria_results"].
type CriterionResult = providerdiagnostics.CriterionResult
