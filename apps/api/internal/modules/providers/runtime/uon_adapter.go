package runtime

import (
	"context"

	"riders.agency/pulse/api/internal/runtime/contracts"
)

func executeUONCreateLeadAdapter(ctx context.Context, runner *Runner, req contracts.ActionRequest) contracts.ActionResult {
	return runner.executeProviderAction(ctx, req, runner.executeUonCreateLead)
}
